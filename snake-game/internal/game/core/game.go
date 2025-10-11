package core

import (
	"fmt"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"log"
	"os"
	"snake-game/internal/game/config"
	"snake-game/internal/game/graphics"
	"snake-game/internal/game/logic"
	"snake-game/internal/game/ui"
	"snake-game/internal/network"
	proto "snake-game/internal/proto/gen"
	"time"
)

type Game struct {
	logic           *logic.GameLogic
	renderer        *graphics.Renderer
	lastUpdate      time.Time
	lastScoreUpdate time.Time
	ui              *ui.ConsoleUI
	games           []*proto.GameAnnouncement
	networkMgr      *network.Manager
	cleanupDone     bool
}

func NewGame() *Game {
	game := &Game{
		lastUpdate: time.Now(),
		ui:         ui.NewConsoleUI(),
	}
	game.networkMgr = network.NewNetworkManager(proto.NodeRole_NORMAL, nil)
	game.networkMgr.SetGameAnnouncementListener(game)
	game.networkMgr.SetGameStateListener(game)
	game.networkMgr.SetGameJoinListener(game)
	game.networkMgr.SetSteerListener(game)
	if err := game.networkMgr.Start(); err != nil {
		log.Printf("Failed to start network manager: %v", err)
	}
	return game
}

func (g *Game) OnGameAnnouncement(games []*proto.GameAnnouncement) {
	g.games = games
}

func (g *Game) OnGameStateReceived(state *proto.GameState) {
	g.logic.SetState(state)
}

func (g *Game) OnGameAddPlayer(player *proto.GamePlayer) {
	g.logic.AddPlayer(player)
}

func (g *Game) OnSteerReceived(playerID int32, direction proto.Direction) error {
	return g.logic.SteerSnake(playerID, direction)
}

func (g *Game) GetLogic() *logic.GameLogic {
	return g.logic
}

func (g *Game) Update() error {
	if ebiten.IsWindowBeingClosed() {
		return g.initiateShutdown()
	}
	g.handleInput()
	if g.networkMgr.GetRole() == proto.NodeRole_MASTER {
		now := time.Now()
		interval := time.Duration(g.logic.Config.StateDelayMs) * time.Millisecond
		if now.Sub(g.lastUpdate) >= interval {
			if err := g.logic.Update(); err != nil {
				return fmt.Errorf("error updating game: %v", err)
			}
			g.lastUpdate = now
			err := g.networkMgr.SendState(g.logic.GetState())
			if err != nil {
				return fmt.Errorf("error updating game: %v", err)
			}
		}
	}
	now := time.Now()
	if now.Sub(g.lastScoreUpdate) >= 250*time.Millisecond {
		ui.ShowScores(g.logic)
		g.lastScoreUpdate = now
	}
	return nil
}

func (g *Game) initiateShutdown() error {
	if g.cleanupDone {
		return nil
	}
	g.cleanupDone = true
	g.cleanup()
	return ebiten.Termination
}

func (g *Game) cleanup() {
	if g.networkMgr != nil {
		g.networkMgr.Close()
	}
}

func (g *Game) handleInput() {
	var newDirection proto.Direction
	keyPressed := false
	if inpututil.IsKeyJustPressed(ebiten.KeyW) || inpututil.IsKeyJustPressed(ebiten.KeyUp) {
		newDirection = proto.Direction_UP
		keyPressed = true
	} else if inpututil.IsKeyJustPressed(ebiten.KeyS) || inpututil.IsKeyJustPressed(ebiten.KeyDown) {
		newDirection = proto.Direction_DOWN
		keyPressed = true
	} else if inpututil.IsKeyJustPressed(ebiten.KeyA) || inpututil.IsKeyJustPressed(ebiten.KeyLeft) {
		newDirection = proto.Direction_LEFT
		keyPressed = true
	} else if inpututil.IsKeyJustPressed(ebiten.KeyD) || inpututil.IsKeyJustPressed(ebiten.KeyRight) {
		newDirection = proto.Direction_RIGHT
		keyPressed = true
	}
	if keyPressed {
		if g.networkMgr.GetRole() == proto.NodeRole_MASTER {
			var masterPlayerID int32
			for _, val := range g.logic.GetPlayers().GetPlayers() {
				if val.GetRole() == proto.NodeRole_MASTER {
					masterPlayerID = val.GetId()
					break
				}
			}
			err := g.OnSteerReceived(masterPlayerID, newDirection)
			if err != nil {
				log.Printf("Error steering master snake: %v", err)
			}
		} else if g.networkMgr.GetRole() == proto.NodeRole_NORMAL {
			g.networkMgr.SendSteer(newDirection)
		} else if g.networkMgr.GetRole() == proto.NodeRole_DEPUTY {
			g.networkMgr.SendSteer(newDirection)
		}
	}
}

func (g *Game) Draw(screen *ebiten.Image) {
	g.updateCellSize()
	g.renderer.Draw(screen)
}

func (g *Game) updateCellSize() {
	if g.renderer != nil && g.logic != nil {
		field := g.logic.GetField()
		if field != nil {
			width, height := ebiten.WindowSize()
			cellWidth := width / int(field.Width)
			cellHeight := height / int(field.Height)
			cellSize := min(cellWidth, cellHeight)
			g.renderer.SetCellSize(cellSize)
		}
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func (g *Game) Layout(width, height int) (int, int) {
	g.updateCellSize()
	return width, height
}

func (g *Game) Start() {
	ebiten.SetWindowSize(600, 600)
	ebiten.SetWindowTitle("Snake Game")

	for {
		mo := g.ui.ShowMainMenu()
		switch mo {
		case ui.StartNewGame:
			g.startNewGame()
		case ui.JoinGame:
			g.joinGame()
		case ui.ShowGames:
			g.showGames()
		case ui.Exit:
			fmt.Println("Goodbye!")
			os.Exit(0)
		default:
			fmt.Println("Invalid option")
		}
	}
}

func (g *Game) startNewGame() {
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		log.Print("CONFIG_PATH env variable not set")
	}
	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}
	g.logic = logic.NewGameLogic(cfg)
	g.renderer = graphics.NewRenderer(g.logic)
	gameName := g.ui.ReadGameName()
	playerName := g.ui.ReadPlayerName()
	fmt.Printf("Creating game '%s' for player '%s'\n", gameName, playerName)
	g.logic.AddPlayer(g.logic.NewPlayer(playerName, proto.PlayerType_HUMAN, proto.NodeRole_MASTER, logic.GeneratePlayerID()))
	gameAnnounce := &proto.GameAnnouncement{
		Config:   g.logic.Config,
		Players:  g.logic.GetPlayers(),
		GameName: gameName,
		CanJoin:  true,
	}
	g.logic.Init()
	g.networkMgr.ChangeRole(g.logic.GetPlayers().GetPlayers()[0], proto.NodeRole_MASTER)
	g.networkMgr.SetGameAnnouncement(gameAnnounce)
	g.networkMgr.SetActivityManager(gameAnnounce.Config.GetStateDelayMs())
	if err := ebiten.RunGame(g); err != nil {
		log.Fatal(err)
	}
}

func (g *Game) joinGame() {
	gameName, playerRole := g.ui.ReadJoinInfo()
	var targetGame *proto.GameAnnouncement
	for name, gameInfo := range g.networkMgr.AvailableGames {
		if gameName == name && gameInfo.Announcement.CanJoin {
			targetGame = gameInfo.Announcement
			break
		}
	}
	if targetGame == nil {
		fmt.Println("Game not found")
		return
	}
	cfg := targetGame.Config
	g.logic = logic.NewGameLogic(cfg)
	playerName := g.ui.ReadPlayerName()
	g.networkMgr.JoinNotify = make(chan int32, 1)
	err := g.networkMgr.SendJoinRequest(proto.PlayerType_HUMAN, playerName, gameName, playerRole)
	if err != nil {
		fmt.Printf("Failed to send join request: %v\n", err)
		return
	}
	fmt.Printf("Join request sent for game '%s'. Waiting for response...\n", gameName)
	select {
	case playerID := <-g.networkMgr.JoinNotify:
		g.renderer = graphics.NewRenderer(g.logic)
		g.networkMgr.SetGameAnnouncement(targetGame)
		fmt.Printf("Successfully joined as %s! Player ID: %d\n", playerRole, playerID)
		g.networkMgr.SetActivityManager(targetGame.Config.GetStateDelayMs())
		if err := ebiten.RunGame(g); err != nil {
			log.Fatal(err)
		}
	case <-time.After(5 * time.Second):
		fmt.Println("Join timeout: no response from game master")
	}

	g.networkMgr.JoinNotify = nil
}

func (g *Game) showGames() {
	g.ui.ShowGameList(g.games)
}
