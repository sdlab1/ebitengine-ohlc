// ohlcv.go
package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"math"

	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

type OHLCV struct {
	Open  float64 `json:"open"`
	High  float64 `json:"high"`
	Low   float64 `json:"low"`
	Close float64 `json:"close"`
}

type Chart struct {
	Data    []OHLCV
	OffsetX float64
	Zoom    float64
	prevX   int
	prevY   int
}

func NewChart() *Chart {
	// Load data from JSON file
	file, err := ioutil.ReadFile("ohlcv_json.txt")
	if err != nil {
		log.Fatal("Failed to load OHLCV data:", err)
	}

	var ohlcvData []OHLCV
	if err := json.Unmarshal(file, &ohlcvData); err != nil {
		log.Fatal("Failed to parse OHLCV data:", err)
	}

	// Use first 100 entries
	if len(ohlcvData) > 100 {
		ohlcvData = ohlcvData[:100]
	}

	return &Chart{
		Data: ohlcvData,
		Zoom: 1.0,
	}
}

func (c *Chart) Update() error {
	// Pan with mouse drag
	if ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
		cx, cy := ebiten.CursorPosition()
		if c.prevX != 0 || c.prevY != 0 {
			c.OffsetX += float64(cx - c.prevX)
		}
		c.prevX, c.prevY = cx, cy
	} else {
		c.prevX, c.prevY = 0, 0
	}

	// Zoom with mouse wheel
	_, dy := ebiten.Wheel()
	c.Zoom = math.Max(0.1, c.Zoom*(1+dy*0.1)) // Limit minimum zoom
	return nil
}

func (c *Chart) Draw(screen *ebiten.Image) {
	screen.Fill(color.RGBA{20, 20, 40, 255}) // Dark background

	barWidth := 2.0 // Thin bars (adjust as needed)
	spacing := 1.0  // Minimal spacing between bars

	for i, ohlcv := range c.Data {
		// Calculate x-position with zoom/pan
		x := float64(i)*(barWidth+spacing)*c.Zoom + c.OffsetX

		// Normalize prices to fit screen (adjust 600 to your window height)
		priceScale := 600.0 / (getMaxPrice(c.Data) - getMinPrice(c.Data))

		// Draw the OHLC bar (thin line from Low to High)
		ebitenutil.DrawLine(
			screen,
			x, 600-ohlcv.Low*priceScale,
			x, 600-ohlcv.High*priceScale,
			color.White,
		)

		// Optional: Mark open/close with tiny ticks
		ebitenutil.DrawLine(
			screen,
			x-1, 600-ohlcv.Open*priceScale,
			x+1, 600-ohlcv.Open*priceScale,
			color.RGBA{0, 255, 0, 255}, // Green for open
		)
		ebitenutil.DrawLine(
			screen,
			x-1, 600-ohlcv.Close*priceScale,
			x+1, 600-ohlcv.Close*priceScale,
			color.RGBA{255, 0, 0, 255}, // Red for close
		)
	}
}

// Helper functions
func getMaxPrice(data []OHLCV) float64 {
	max := data[0].High
	for _, d := range data {
		if d.High > max {
			max = d.High
		}
	}
	return max
}

func getMinPrice(data []OHLCV) float64 {
	min := data[0].Low
	for _, d := range data {
		if d.Low < min {
			min = d.Low
		}
	}
	return min
}
