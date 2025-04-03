package main

import (
	"log"

	"github.com/hajimehoshi/ebiten/v2"
)

type Game struct {
	chart       *Chart
	axes        *Axes
	interaction *Interaction
	volume      *Volume
}

func main() {
	ebiten.SetWindowSize(1000, 700)
	ebiten.SetWindowTitle("OHLC Chart Viewer")

	config := DefaultConfig // Get the default config

	game := &Game{
		chart:       NewChart(config),
		axes:        NewAxes(config),
		interaction: NewInteraction(config),
		volume:      NewVolume(config),
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
	g.axes.Draw(screen, g.chart)        // Draw grid and axes first
	g.volume.Draw(screen, g.chart)      // Then volume bars
	g.chart.Draw(screen)                // Then OHLC price bars
	g.interaction.Draw(screen, g.chart) // Finally crosshair
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return 1000, 700
}
