package main

import (
	"encoding/json"
	"log"
	"math"
	"os"

	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
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

// Define common colors
var (
	backgroundColor = color.RGBA{20, 20, 40, 255}
	barColor        = color.White
	openColor       = color.RGBA{0, 255, 0, 255} // Green
	closeColor      = color.RGBA{255, 0, 0, 255} // Red
	axisColor       = color.RGBA{255, 0, 0, 255} // Red for axes
)

func NewChart() *Chart {
	file, err := os.ReadFile("ohlcv_json.txt")
	if err != nil {
		log.Fatal("Failed to load OHLCV data:", err)
	}

	var ohlcvData []OHLCV
	if err := json.Unmarshal(file, &ohlcvData); err != nil {
		log.Fatal("Failed to parse OHLCV data:", err)
	}

	if len(ohlcvData) > 100 {
		ohlcvData = ohlcvData[:100]
	}

	return &Chart{
		Data: ohlcvData,
		Zoom: 1.0,
	}
}

func (c *Chart) Update() error {
	if ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
		cx, cy := ebiten.CursorPosition()
		if c.prevX != 0 || c.prevY != 0 {
			c.OffsetX += float64(cx - c.prevX)
		}
		c.prevX, c.prevY = cx, cy
	} else {
		c.prevX, c.prevY = 0, 0
	}

	_, dy := ebiten.Wheel()
	c.Zoom = math.Max(0.1, c.Zoom*(1+dy*0.1))
	return nil
}

func (c *Chart) Draw(screen *ebiten.Image) {
	screen.Fill(backgroundColor)

	// Draw axes for reference
	vector.StrokeLine(screen, 50, 50, 50, 550, 1, axisColor, false)   // Y-axis
	vector.StrokeLine(screen, 50, 550, 750, 550, 1, axisColor, false) // X-axis

	barWidth := 2.0
	spacing := 1.0
	yScale := 0.0001 // Adjust based on your price range

	for i, ohlcv := range c.Data {
		x := float32(50 + float64(i)*(barWidth+spacing)*c.Zoom + c.OffsetX)

		// Convert prices to screen coordinates
		lowY := float32(550 - (ohlcv.Low*yScale - 700)) // Adjusted for better visibility
		highY := float32(550 - (ohlcv.High*yScale - 700))

		// Draw main OHLC bar
		vector.StrokeLine(
			screen,
			x, lowY,
			x, highY,
			1.0,
			barColor,
			false,
		)

		// Mark open price
		openY := float32(550 - (ohlcv.Open*yScale - 700))
		vector.StrokeLine(
			screen,
			x-2, openY,
			x+2, openY,
			1.0,
			openColor,
			false,
		)

		// Mark close price
		closeY := float32(550 - (ohlcv.Close*yScale - 700))
		vector.StrokeLine(
			screen,
			x-2, closeY,
			x+2, closeY,
			1.0,
			closeColor,
			false,
		)
	}
}
