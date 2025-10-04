package graphics

import (
	"github.com/hajimehoshi/ebiten/v2"
	"golang.org/x/image/colornames"
	"image/color"
	"snake-game/internal/game/logic"
)

type Renderer struct {
	logic    *logic.GameLogic
	cellSize int
}

func NewRenderer(logic *logic.GameLogic) *Renderer {
	return &Renderer{
		logic:    logic,
		cellSize: 20,
	}
}

func (r *Renderer) Draw(screen *ebiten.Image) {
	screen.Fill(color.RGBA{R: 0x1a, G: 0x1f, B: 0x2d, A: 0xff})
	r.drawGrid(screen)
	r.drawFood(screen)
	r.drawSnakes(screen)
}

func (r *Renderer) drawGrid(screen *ebiten.Image) {
	field := r.logic.GetField()
	for y := 0; int32(y) < field.Height; y++ {
		for x := 0; int32(x) < field.Width; x++ {
			r.drawCell(screen, x, y, color.RGBA{R: 0x2d, G: 0x32, B: 0x45, A: 0xff})
		}
	}
}

func (r *Renderer) drawFood(screen *ebiten.Image) {
	foods := r.logic.GetFoods()
	for _, food := range foods {
		if food != nil {
			r.drawCell(screen, int(food.X), int(food.Y), colornames.Red)
		}
	}
}

func (r *Renderer) drawSnakes(screen *ebiten.Image) {
	snakes := r.logic.GetSnakes()
	for _, snake := range snakes {
		points := snake.GetPoints()
		for i, point := range points {
			if i == 0 {
				r.drawCell(screen, int(point.X), int(point.Y), colornames.Green)
			} else {
				r.drawCell(screen, int(point.X), int(point.Y), colornames.Hotpink)
			}
		}
	}
}

func (r *Renderer) drawCell(screen *ebiten.Image, x, y int, clr color.Color) {
	rect := ebiten.NewImage(r.cellSize-1, r.cellSize-1)
	rect.Fill(clr)
	opts := &ebiten.DrawImageOptions{}
	opts.GeoM.Translate(float64(x*r.cellSize), float64(y*r.cellSize))
	screen.DrawImage(rect, opts)
}
