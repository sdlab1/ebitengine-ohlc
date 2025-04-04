package main

import (
	"log"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

type Game struct {
	chart       *Chart
	axes        *Axes
	interaction *Interaction
	volume      *Volume
	lastUpdate  time.Time
	needsRedraw bool // Flag to determine if rendering is needed
	prevMouseX  int  // Previous mouse X position
	prevMouseY  int  // Previous mouse Y position
}

func main() {
	// Disable screen clearing every frame to optimize GPU usage
	ebiten.SetScreenClearedEveryFrame(false)

	ebiten.SetWindowSize(1000, 700)
	ebiten.SetWindowTitle("OHLC Chart Viewer")

	config := DefaultConfig
	chart := NewChart(config)

	// Fetch initial data - use UTC time explicitly
	data, err := Fetch(1000, time.Now().UTC().Unix()*1000) // Get 1000 most recent 1-minute candles
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
		needsRedraw: true, // Initial render is required
		prevMouseX:  -1,   // Initialize to invalid position
		prevMouseY:  -1,
	}

	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}

func (g *Game) Update() error {
	// Track if any input or data change requires a redraw
	inputDetected := false

	// Check keyboard input (e.g., arrow keys for panning)
	if inpututil.IsKeyJustPressed(ebiten.KeyArrowLeft) || inpututil.IsKeyJustPressed(ebiten.KeyArrowRight) {
		inputDetected = true
	}

	// Check mouse input (click, drag, or wheel)
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) || ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
		inputDetected = true
	}
	if _, dy := ebiten.Wheel(); dy != 0 {
		inputDetected = true
	}

	// Check mouse movement
	cx, cy := ebiten.CursorPosition()
	if cx != g.prevMouseX || cy != g.prevMouseY {
		inputDetected = true
	}
	g.prevMouseX, g.prevMouseY = cx, cy

	// Check touch input (for Android) using AppendTouchIDs
	var touchIDs []ebiten.TouchID
	touchIDs = ebiten.AppendTouchIDs(touchIDs) // Collect active touch IDs
	for _, id := range touchIDs {
		// Check if the touch was just pressed (duration of 1 frame)
		if inpututil.TouchPressDuration(id) == 1 {
			inputDetected = true
			break
		}
	}

	// Optionally: Auto-refresh data periodically (e.g., every minute)
	now := time.Now()
	if now.Sub(g.lastUpdate) > time.Minute {
		utcNow := now.UTC()
		endTime := utcNow.Unix() * 1000
		data, err := Fetch(1000, endTime)
		if err == nil {
			g.chart.UpdateData(data)
			inputDetected = true // Data change requires redraw
		}
		g.lastUpdate = now
	}

	// Update chart, axes, and interaction
	if err := g.chart.Update(); err != nil {
		return err
	}
	g.axes.Update(g.chart)
	g.interaction.Update(g.chart)

	// Set needsRedraw if there's input or crosshair visibility changes
	if inputDetected || g.interaction.showCrosshair != g.interaction.prevShowCrosshair {
		g.needsRedraw = true
	}

	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	// Skip rendering if no changes or input detected
	if !g.needsRedraw {
		return
	}

	// Clear the screen with the background color to remove artifacts
	screen.Fill(g.chart.config.BackgroundColor)

	// Render all components
	g.axes.Draw(screen, g.chart)
	g.volume.Draw(screen, g.chart)
	g.chart.Draw(screen)
	g.interaction.Draw(screen, g.chart)

	// Reset the flag after rendering
	g.needsRedraw = false
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return 1000, 700
}
