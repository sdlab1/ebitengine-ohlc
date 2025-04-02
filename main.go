package main

import (
	"log"

	"github.com/hajimehoshi/ebiten/v2"
)

type Game struct {
	chart *Chart
	axes  *Axes
}

func main() {
	ebiten.SetWindowSize(1000, 700)
	ebiten.SetWindowTitle("OHLC Chart Viewer")

	game := &Game{
		chart: NewChart(),
		axes:  NewAxes(),
	}

	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}

func (g *Game) Update() error {
	if err := g.chart.Update(); err != nil {
		return err
	}
	g.axes.Update(g.chart)
	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	g.axes.Draw(screen, g.chart) // Pass the chart reference
	g.chart.Draw(screen)
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return 1000, 700
}
