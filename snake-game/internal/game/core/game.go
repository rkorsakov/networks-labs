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
	logic      *logic.GameLogic
	renderer   *graphics.Renderer
	lastUpdate time.Time
	ui         *ui.ConsoleUI
	games      []*proto.GameAnnouncement
	networkMgr *network.Manager
}

func NewGame() *Game {
	game := &Game{
		lastUpdate: time.Now(),
		ui:         ui.NewConsoleUI(),
	}
	game.networkMgr = network.NewNetworkManager(proto.NodeRole_NORMAL, nil)
	game.networkMgr.SetGameAnnouncementListener(game)
	if err := game.networkMgr.Start(); err != nil {
		log.Printf("Failed to start network manager: %v", err)
	}
	return game
}

func (g *Game) OnGameAnnouncement(games []*proto.GameAnnouncement) {
	g.games = games
	log.Printf("Updated game list: %d games available", len(games))
}

func (g *Game) Update() error {
	g.handleInput()
	now := time.Now()
	interval := time.Duration(g.logic.Config.StateDelayMs) * time.Millisecond
	if now.Sub(g.lastUpdate) >= interval {
		if err := g.logic.Update(); err != nil {
			return err
		}
		g.lastUpdate = now
	}
	return nil
}

func (g *Game) handleInput() {
	if g.networkMgr.GetRole() == proto.NodeRole_MASTER {
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
		var masterPlayerID int32
		for _, val := range g.logic.GetPlayers().GetPlayers() {
			if val.GetRole() == proto.NodeRole_MASTER {
				masterPlayerID = val.GetId()
			}
		}
		if keyPressed {
			err := g.logic.SteerSnake(masterPlayerID, newDirection)
			if err != nil {
				log.Fatal(err)
			}
		}
	}
}

func (g *Game) Draw(screen *ebiten.Image) {
	g.renderer.Draw(screen)
}

func (g *Game) Layout(width, height int) (int, int) {
	return width, height
}

func (g *Game) Start() {
	ebiten.SetWindowSize(800, 600)
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
	g.logic.AddPlayer(g.logic.NewPlayer(playerName, proto.PlayerType_HUMAN, proto.NodeRole_MASTER))
	gameAnnounce := &proto.GameAnnouncement{
		Config:   g.logic.Config,
		Players:  g.logic.GetPlayers(),
		GameName: gameName,
		CanJoin:  true,
	}
	g.logic.Init()
	g.networkMgr.ChangeRole(proto.NodeRole_MASTER, gameAnnounce)
	if err := ebiten.RunGame(g); err != nil {
		log.Fatal(err)
	}
}

func (g *Game) joinGame() {
	//TODO
}

func (g *Game) showGames() {
	g.ui.ShowGameList(g.games)
}
