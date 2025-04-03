package main

import (
	"log"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
)

type Game struct {
	chart       *Chart
	axes        *Axes
	interaction *Interaction
	volume      *Volume
	lastUpdate  time.Time
}

func main() {
	ebiten.SetWindowSize(1000, 700)
	ebiten.SetWindowTitle("OHLC Chart Viewer")

	config := DefaultConfig
	chart := NewChart(config)

	// Fetch initial data - use UTC time explicitly
	data, err := Fetch(1000, time.Now().UTC().Unix()*1000) // Get 300 most recent 1-minute candles
	if err != nil {
		log.Fatal("Failed to fetch data:", err)
	}
	chart.UpdateData(data)

	game := &Game{
		chart:       chart,
		axes:        NewAxes(config),
		interaction: NewInteraction(config),
		volume:      NewVolume(config),
		lastUpdate:  time.Now(),
	}

	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}

func (g *Game) Update() error {
	// Optionally: Auto-refresh data periodically (e.g., every minute)
	/*now := time.Now()
	if now.Sub(g.lastUpdate) > time.Minute {
		utcNow := now.UTC()
		endTime := utcNow.Unix() * 1000
		data, err := Fetch(1000, endTime)
		if err == nil {
			g.chart.UpdateData(data)
		}
		g.lastUpdate = now
	}*/

	if err := g.chart.Update(); err != nil {
		return err
	}
	g.axes.Update(g.chart)
	g.interaction.Update(g.chart)
	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	g.axes.Draw(screen, g.chart)
	g.volume.Draw(screen, g.chart)
	g.chart.Draw(screen)
	g.interaction.Draw(screen, g.chart)
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return 1000, 700
}
