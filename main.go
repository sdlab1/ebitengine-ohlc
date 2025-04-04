package main

import (
	"log"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

type Game struct {
	chart           *Chart
	axes            *Axes
	interaction     *Interaction
	volume          *Volume
	db              *Database
	timeframe       *Timeframe
	lastUpdate      time.Time
	needsRedraw     bool
	prevMouseX      int
	prevMouseY      int
	prevFetchStatus string
	prevErrorMsg    string
}

func main() {
	// Disable screen clearing optimization to ensure initial draw
	ebiten.SetScreenClearedEveryFrame(false)

	ebiten.SetWindowSize(1000, 700)
	ebiten.SetWindowTitle("OHLC Chart Viewer")

	config := DefaultConfig
	chart := NewChart(config)

	db, err := NewDatabase()
	if err != nil {
		log.Fatal("Failed to initialize database:", err)
	}
	defer db.Close()

	timeframe := NewTimeframe(db.db)

	// Load initial data if available, otherwise chart will be empty until data is fetched
	data, err := timeframe.Get15MinBars()
	if err != nil {
		log.Printf("No initial 15-min bars available: %v", err)
		data = []OHLCV{}
	}
	chart.UpdateData(data)

	game := &Game{
		chart:           chart,
		axes:            NewAxes(config),
		interaction:     NewInteraction(config),
		volume:          NewVolume(config),
		db:              db,
		timeframe:       timeframe,
		lastUpdate:      time.Now(),
		needsRedraw:     true, // Ensure initial render
		prevMouseX:      -1,
		prevMouseY:      -1,
		prevFetchStatus: "",
		prevErrorMsg:    "",
	}

	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}

func (g *Game) Update() error {
	inputDetected := false

	// Check keyboard input
	if inpututil.IsKeyJustPressed(ebiten.KeyArrowLeft) || inpututil.IsKeyJustPressed(ebiten.KeyArrowRight) {
		inputDetected = true
	}
	// Check mouse input
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

	// Check touch input
	var touchIDs []ebiten.TouchID
	touchIDs = ebiten.AppendTouchIDs(touchIDs)
	for _, id := range touchIDs {
		if inpututil.TouchPressDuration(id) == 1 {
			inputDetected = true
			break
		}
	}

	// Check for changes in fetch status or error message
	g.db.fetchMutex.Lock()
	currentFetchStatus := g.db.fetchStatus
	currentErrorMsg := g.db.errorMsg
	fetching := g.db.fetching
	g.db.fetchMutex.Unlock()

	// Force redraw if fetching is in progress or status/error changed
	if fetching || currentFetchStatus != g.prevFetchStatus || currentErrorMsg != g.prevErrorMsg {
		g.needsRedraw = true
		g.prevFetchStatus = currentFetchStatus
		g.prevErrorMsg = currentErrorMsg
	}

	// Auto-refresh data periodically
	now := time.Now()
	if now.Sub(g.lastUpdate) > time.Minute {
		data, err := g.timeframe.Get15MinBars()
		if err != nil {
			log.Printf("Failed to refresh 15-min bars: %v", err)
			if err := g.db.ensureLastData(); err != nil {
				log.Printf("Failed to update database: %v", err)
			}
			data, err = g.timeframe.Get15MinBars()
			if err != nil {
				log.Printf("Retry failed: %v", err)
				return nil
			}
		}
		g.chart.UpdateData(data)
		inputDetected = true
		g.lastUpdate = now
	}

	if err := g.chart.Update(); err != nil {
		return err
	}
	g.axes.Update(g.chart)
	g.interaction.Update(g.chart)

	if inputDetected || g.interaction.showCrosshair != g.interaction.prevShowCrosshair {
		g.needsRedraw = true
	}

	// Force Ebiten to redraw if needed
	if g.needsRedraw {
		// Note: ebiten.RequestUpdate() is available in newer versions of Ebiten.
		// If using an older version, this line can be removed, and the redraw will still work
		// due to the simplified needsRedraw logic.
		// ebiten.RequestUpdate()
	}

	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	// Remove the needsRedraw check to ensure drawing happens when requested
	// Alternatively, keep it if using ebiten.RequestUpdate() in newer Ebiten versions
	// if !g.needsRedraw {
	// 	return
	// }

	screen.Fill(g.chart.config.BackgroundColor)
	g.axes.Draw(screen, g.chart)
	g.volume.Draw(screen, g.chart)
	g.chart.Draw(screen)
	g.interaction.Draw(screen, g.chart)
	g.db.DrawError(screen)
	g.needsRedraw = false
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return 1000, 700
}
