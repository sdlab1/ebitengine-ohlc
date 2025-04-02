// main.go
package main

import (
	"log"

	"github.com/hajimehoshi/ebiten/v2"
)

type Game struct {
	chart *Chart
}

func (g *Game) Update() error {
	return g.chart.Update()
}

func (g *Game) Draw(screen *ebiten.Image) {
	g.chart.Draw(screen)
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return 800, 600
}

func main() {
	ebiten.SetWindowSize(800, 600)
	ebiten.SetWindowTitle("OHLCV Chart")

	game := &Game{
		chart: NewChart(),
	}

	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}
