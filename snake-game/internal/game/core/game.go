package core

import (
	"github.com/hajimehoshi/ebiten/v2"
	"log"
	"snake-game/internal/game/config"
	"snake-game/internal/game/graphics"
	"snake-game/internal/game/logic"
)

type Game struct {
	Logic    *logic.GameLogic
	Renderer *graphics.Renderer
}

func NewGame(cfg *config.Config) *Game {
	gameLogic := logic.NewGameLogic(cfg)
	renderer := graphics.NewRenderer(gameLogic)

	return &Game{
		Logic:    gameLogic,
		Renderer: renderer,
	}
}

func (g *Game) Update() error {
	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	g.Renderer.Draw(screen)
}

func (g *Game) Layout(width, height int) (int, int) {
	return width, height
}

func (g *Game) Start() {
	ebiten.SetWindowSize(700, 400)
	ebiten.SetWindowTitle("Snake")
	if err := ebiten.RunGame(g); err != nil {
		log.Fatal(err)
	}
}
