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

var (
	barColor   = color.White
	openColor  = color.RGBA{0, 255, 0, 255}
	closeColor = color.RGBA{255, 0, 0, 255}
)

func NewChart(config ChartConfig) *Chart {
	file, err := os.ReadFile("ohlcv_json.txt")
	if err != nil {
		log.Fatal("Failed to load OHLCV data:", err)
	}

	var ohlcvData []OHLCV
	if err := json.Unmarshal(file, &ohlcvData); err != nil {
		log.Fatal("Failed to parse OHLCV data:", err)
	}

	if len(ohlcvData) == 0 {
		log.Fatal("No OHLCV data loaded")
	}

	// Calculate price range
	min, max := calculatePriceRange(ohlcvData)

	return &Chart{
		Data:      ohlcvData,
		Zoom:      1.0,
		priceMin:  min,
		priceMax:  max,
		timeStart: ohlcvData[0].Time,
		timeEnd:   ohlcvData[len(ohlcvData)-1].Time,
		config:    config,
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

			// Apply zoom
			newZoom := c.Zoom * math.Pow(1.1, dy)
			newZoom = math.Max(0.1, math.Min(newZoom, 10.0))

			// Adjust offset to zoom toward mouse position
			c.OffsetX = (float64(cx) - c.config.LeftMargin) - mouseDataPos*(totalBarSpace*newZoom)
			c.Zoom = newZoom
		}
	}

	// Handle mouse drag panning (unchanged)
	if ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
		cx, cy := ebiten.CursorPosition()
		if c.prevX != 0 || c.prevY != 0 {
			c.OffsetX += float64(cx - c.prevX)
		}
		c.prevX, c.prevY = cx, cy
	} else {
		c.prevX, c.prevY = 0, 0
	}

	// Constrain offset
	totalBarsWidth := float64(len(c.Data)) * (c.config.BarWidth + c.config.BarSpacing) * c.Zoom
	visibleWidth := c.config.Width - c.config.LeftMargin - c.config.RightMargin

	if totalBarsWidth < visibleWidth {
		c.OffsetX = 0
	} else {
		maxOffset := totalBarsWidth - visibleWidth
		c.OffsetX = math.Max(0, math.Min(c.OffsetX, maxOffset))
	}

	return nil
}

func (c *Chart) Draw(screen *ebiten.Image) {
	totalBarSpace := c.config.BarWidth + c.config.BarSpacing

	for i, ohlcv := range c.Data {
		x := float32(c.config.LeftMargin + (float64(i)*totalBarSpace*c.Zoom + c.OffsetX))

		if x < float32(c.config.LeftMargin) || x > float32(c.config.Width-c.config.RightMargin) {
			continue
		}

		barTop := float32(c.config.Height - c.config.BottomMargin - ((ohlcv.High - c.priceMin) / (c.priceMax - c.priceMin) * float64(c.config.Height-c.config.TopMargin-c.config.BottomMargin)))
		barBottom := float32(c.config.Height - c.config.BottomMargin - ((ohlcv.Low - c.priceMin) / (c.priceMax - c.priceMin) * float64(c.config.Height-c.config.TopMargin-c.config.BottomMargin)))

		// Draw bar using config color
		vector.StrokeLine(
			screen,
			x, barTop,
			x, barBottom,
			float32(c.config.BarWidth),
			c.config.BarColor,
			false,
		)

		openY := float32(c.config.Height - c.config.BottomMargin - ((ohlcv.Open - c.priceMin) / (c.priceMax - c.priceMin) * float64(c.config.Height-c.config.TopMargin-c.config.BottomMargin)))
		closeY := float32(c.config.Height - c.config.BottomMargin - ((ohlcv.Close - c.priceMin) / (c.priceMax - c.priceMin) * float64(c.config.Height-c.config.TopMargin-c.config.BottomMargin)))

		// Draw ticks using config colors
		vector.StrokeLine(
			screen,
			x-2, openY,
			x+2, openY,
			1.0,
			c.config.OpenColor,
			false,
		)

		vector.StrokeLine(
			screen,
			x-2, closeY,
			x+2, closeY,
			1.0,
			c.config.CloseColor,
			false,
		)
	}
}
