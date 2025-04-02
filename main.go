package main

import (
	"log"

	"github.com/hajimehoshi/ebiten/v2"
)

type Game struct {
	chart       *Chart
	axes        *Axes
	interaction *Interaction
}

func main() {
	ebiten.SetWindowSize(1000, 700)
	ebiten.SetWindowTitle("OHLC Chart Viewer")

	game := &Game{
		chart:       NewChart(),
		axes:        NewAxes(DefaultConfig),
		interaction: NewInteraction(DefaultConfig),
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
	g.interaction.Update(g.chart)
	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	g.axes.Draw(screen, g.chart)
	g.chart.Draw(screen)
	g.interaction.Draw(screen, g.chart)
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return 1000, 700
}
