package core

import (
	"fmt"
	"github.com/hajimehoshi/ebiten/v2"
	"log"
	"os"
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

func NewGame(cfg *proto.GameConfig) *Game {
	gameLogic := logic.NewGameLogic(cfg)
	renderer := graphics.NewRenderer(gameLogic)
	return &Game{
		logic:      gameLogic,
		renderer:   renderer,
		lastUpdate: time.Now(),
		ui:         ui.NewConsoleUI(),
	}
}

func (g *Game) OnGameAnnouncement(games []*proto.GameAnnouncement) {
	g.games = games
	log.Printf("Updated game list: %d games available", len(games))
}

func (g *Game) Update() error {
	//g.handleInput()
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

//func (g *Game) handleInput() {
//	var newDirection proto.Direction
//	keyPressed := false
//	if inpututil.IsKeyJustPressed(ebiten.KeyW) || inpututil.IsKeyJustPressed(ebiten.KeyUp) {
//		newDirection = proto.Direction_UP
//		keyPressed = true
//	} else if inpututil.IsKeyJustPressed(ebiten.KeyS) || inpututil.IsKeyJustPressed(ebiten.KeyDown) {
//		newDirection = proto.Direction_DOWN
//		keyPressed = true
//	} else if inpututil.IsKeyJustPressed(ebiten.KeyA) || inpututil.IsKeyJustPressed(ebiten.KeyLeft) {
//		newDirection = proto.Direction_LEFT
//		keyPressed = true
//	} else if inpututil.IsKeyJustPressed(ebiten.KeyD) || inpututil.IsKeyJustPressed(ebiten.KeyRight) {
//		newDirection = proto.Direction_RIGHT
//		keyPressed = true
//	}
//	if keyPressed {
//		err := g.logic.SteerSnake(g.playerID, newDirection)
//		if err != nil {
//			log.Fatal(err)
//		}
//	}
//}

func (g *Game) Draw(screen *ebiten.Image) {
	g.renderer.Draw(screen)
}

func (g *Game) Layout(width, height int) (int, int) {
	return width, height
}

func (g *Game) Start() {
	ebiten.SetWindowSize(800, 600)
	ebiten.SetWindowTitle("Snake Game")

	go func() {
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
	}()

	if err := ebiten.RunGame(g); err != nil {
		log.Fatal(err)
	}
}

func (g *Game) startNewGame() {
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
	g.networkMgr = network.NewNetworkManager(proto.NodeRole_MASTER, gameAnnounce)
	g.networkMgr.SetGameAnnouncementListener(g)
	err := g.networkMgr.Start()
	if err != nil {
		log.Fatal(err)
	}
}

func (g *Game) joinGame() {
	g.networkMgr = network.NewNetworkManager(proto.NodeRole_NORMAL, nil)
	g.networkMgr.SetGameAnnouncementListener(g)
	err := g.networkMgr.Start()
	if err != nil {
		log.Fatal(err)
	}
	g.showGames()
}

func (g *Game) showGames() {
	if g.networkMgr == nil {
		g.networkMgr = network.NewNetworkManager(proto.NodeRole_NORMAL, nil)
		g.networkMgr.SetGameAnnouncementListener(g)
		g.networkMgr.Start()
	}
	g.ui.ShowGameList(g.games)
}
