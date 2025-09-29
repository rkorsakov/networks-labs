package core

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"log"
	"snake-game/internal/game/graphics"
	"snake-game/internal/game/logic"
	"snake-game/internal/game/ui"
	proto "snake-game/internal/proto/gen"
	"time"
)

type Game struct {
	logic      *logic.GameLogic
	renderer   *graphics.Renderer
	lastUpdate time.Time
	playerID   int32
	ui         *ui.ConsoleUI
}

func NewGame(cfg *proto.GameConfig) *Game {
	gameLogic := logic.NewGameLogic(cfg)
	renderer := graphics.NewRenderer(gameLogic)
	var playerID int32
	for id := range gameLogic.GetPlayers() {
		playerID = id
		break
	}
	return &Game{
		logic:      gameLogic,
		renderer:   renderer,
		lastUpdate: time.Now(),
		playerID:   playerID,
		ui:         ui.NewConsoleUI(),
	}
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
		err := g.logic.SteerSnake(g.playerID, newDirection)
		if err != nil {
			log.Fatal(err)
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
	go func() { g.ui.ShowMainMenu() }()
	if err := ebiten.RunGame(g); err != nil {
		log.Fatal(err)
	}
}
