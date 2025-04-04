package main

import (
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

type OHLCV struct {
	Open   float64 `json:"open"`
	High   float64 `json:"high"`
	Low    float64 `json:"low"`
	Close  float64 `json:"close"`
	Time   int64   `json:"time"`
	Volume float64 `json:"volume"`
}

type Chart struct {
	Data      []OHLCV
	OffsetX   float64
	Zoom      float64
	prevX     int
	prevY     int
	priceMin  float64
	priceMax  float64
	timeStart int64
	timeEnd   int64
	config    ChartConfig
}

func NewChart(config ChartConfig) *Chart {
	return &Chart{
		Zoom:   1.0,
		config: config,
		Data:   make([]OHLCV, 0),
	}
}

func (c *Chart) UpdateData(newData []OHLCV) {
	if len(newData) == 0 {
		c.Data = make([]OHLCV, 0)
		c.priceMin, c.priceMax = 0, 0
		c.timeStart, c.timeEnd = 0, 0
		c.OffsetX = 0
		c.Zoom = 1.0
		return
	}

	c.Data = newData
	c.priceMin, c.priceMax = calculatePriceRange(c.Data)
	c.timeStart = c.Data[0].Time
	c.timeEnd = c.Data[len(c.Data)-1].Time

	c.Zoom = 1.0
	totalBarSpace := c.config.BarWidth + c.config.BarSpacing

	// Position last bar at right edge
	if len(c.Data) > 0 {
		lastBarIdx := len(c.Data) - 1
		c.OffsetX = (c.config.Width - c.config.RightMargin - c.config.BarWidth*c.Zoom/2) - c.config.LeftMargin - float64(lastBarIdx)*totalBarSpace*c.Zoom
	}

}

func calculatePriceRange(data []OHLCV) (min, max float64) {
	min = data[0].Low
	max = data[0].High
	for _, d := range data {
		if d.Low < min {
			min = d.Low
		}
		if d.High > max {
			max = d.High
		}
	}
	return min, max
}

func (c *Chart) Update() error {
	// Handle mouse wheel zoom
	_, dy := ebiten.Wheel()
	if dy != 0 {
		cx, _ := ebiten.CursorPosition()
		chartLeft := int(c.config.LeftMargin)
		chartRight := int(c.config.Width - c.config.RightMargin)

		if cx >= chartLeft && cx <= chartRight {
			totalBarSpace := c.config.BarWidth + c.config.BarSpacing
			mouseDataPos := (float64(cx) - c.config.LeftMargin - c.OffsetX) / (totalBarSpace * c.Zoom)

			// Apply zoom with limits
			newZoom := c.Zoom * math.Pow(1.1, dy)
			newZoom = math.Max(0.1, math.Min(newZoom, 10.0))

			// Adjust offset to zoom around mouse position
			c.OffsetX = (float64(cx) - c.config.LeftMargin) - mouseDataPos*(totalBarSpace*newZoom)
			c.Zoom = newZoom
		}
	}

	// Handle mouse drag panning
	if ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
		cx, cy := ebiten.CursorPosition()
		if c.prevX != 0 || c.prevY != 0 {
			c.OffsetX += float64(cx - c.prevX)
		}
		c.prevX, c.prevY = cx, cy
	} else {
		c.prevX, c.prevY = 0, 0
	}

	// Constrain offset to keep bars within bounds
	totalBarsWidth := float64(len(c.Data)) * (c.config.BarWidth + c.config.BarSpacing) * c.Zoom
	visibleWidth := c.config.Width - c.config.LeftMargin - c.config.RightMargin

	if totalBarsWidth <= visibleWidth {
		c.OffsetX = 0 // All bars fit, no offset
	} else {
		maxOffset := totalBarsWidth - visibleWidth
		c.OffsetX = math.Max(0, math.Min(c.OffsetX, maxOffset))
	}

	return nil
}

func (c *Chart) Draw(screen *ebiten.Image) {
	totalBarSpace := c.config.BarWidth + c.config.BarSpacing
	//visibleWidth := c.config.Width - c.config.LeftMargin - c.config.RightMargin

	// Adjust OffsetX to ensure the last bar is at the right edge
	if len(c.Data) > 0 {
		lastBarIdx := len(c.Data) - 1
		lastBarX := c.config.LeftMargin + (float64(lastBarIdx)*totalBarSpace*c.Zoom + c.OffsetX)
		expectedLastBarX := c.config.Width - c.config.RightMargin - (c.config.BarWidth * c.Zoom / 2)
		if lastBarX != expectedLastBarX {
			c.OffsetX += expectedLastBarX - lastBarX
		}
	}

	for i, ohlcv := range c.Data {
		x := float32(c.config.LeftMargin + (float64(i)*totalBarSpace*c.Zoom + c.OffsetX))

		// Center the bar
		barWidth := float32(c.config.BarWidth * c.Zoom)
		x -= barWidth / 2

		// Skip bars outside the visible area
		if x+barWidth < float32(c.config.LeftMargin) || x > float32(c.config.Width-c.config.RightMargin) {
			continue
		}

		barTop := float32(c.config.Height - c.config.BottomMargin - ((ohlcv.High - c.priceMin) / (c.priceMax - c.priceMin) * float64(c.config.Height-c.config.TopMargin-c.config.BottomMargin)))
		barBottom := float32(c.config.Height - c.config.BottomMargin - ((ohlcv.Low - c.priceMin) / (c.priceMax - c.priceMin) * float64(c.config.Height-c.config.TopMargin-c.config.BottomMargin)))

		// Draw the candlestick wick (high to low)
		vector.StrokeLine(
			screen,
			x+barWidth/2, // Center the wick
			barTop,
			x+barWidth/2,
			barBottom,
			1.0, // Thin wick
			c.config.BarColor,
			false,
		)

		// Calculate Open and Close positions
		openY := float32(c.config.Height - c.config.BottomMargin - ((ohlcv.Open - c.priceMin) / (c.priceMax - c.priceMin) * float64(c.config.Height-c.config.TopMargin-c.config.BottomMargin)))
		closeY := float32(c.config.Height - c.config.BottomMargin - ((ohlcv.Close - c.priceMin) / (c.priceMax - c.priceMin) * float64(c.config.Height-c.config.TopMargin-c.config.BottomMargin)))

		// Draw open tick
		vector.StrokeLine(
			screen,
			x, // Left edge of bar
			openY,
			x+barWidth, // Right edge of bar
			openY,
			2.0,
			c.config.OpenColor,
			false,
		)

		// Draw close tick
		vector.StrokeLine(
			screen,
			x, // Left edge of bar
			closeY,
			x+barWidth, // Right edge of bar
			closeY,
			2.0,
			c.config.CloseColor,
			false,
		)
	}
}

func (c *Chart) GetBarPosition(index int) (left, center, right float64) {
	totalBarSpace := (c.config.BarWidth + c.config.BarSpacing) * c.Zoom
	left = c.config.LeftMargin + (float64(index) * totalBarSpace) + c.OffsetX
	right = left + c.config.BarWidth*c.Zoom
	center = left + (c.config.BarWidth*c.Zoom)/2
	return left, center, right
}
