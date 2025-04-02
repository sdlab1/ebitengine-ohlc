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
}

var (
	barColor   = color.White
	openColor  = color.RGBA{0, 255, 0, 255}
	closeColor = color.RGBA{255, 0, 0, 255}
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
	barWidth := 2.0
	spacing := 1.0

	for i, ohlcv := range c.Data {
		x := float32(100 + float64(i)*(barWidth+spacing)*c.Zoom + c.OffsetX)

		// Skip drawing if outside view
		if x < 50 || x > 950 {
			continue
		}

		lowY := float32(600 - ((ohlcv.Low - c.priceMin) / (c.priceMax - c.priceMin) * 500))
		highY := float32(600 - ((ohlcv.High - c.priceMin) / (c.priceMax - c.priceMin) * 500))

		// Draw OHLC bar
		vector.StrokeLine(
			screen,
			x, lowY,
			x, highY,
			1.0,
			barColor,
			false,
		)

		// Draw open/close ticks
		openY := float32(600 - ((ohlcv.Open - c.priceMin) / (c.priceMax - c.priceMin) * 500))
		closeY := float32(600 - ((ohlcv.Close - c.priceMin) / (c.priceMax - c.priceMin) * 500))

		vector.StrokeLine(
			screen,
			x-2, openY,
			x+2, openY,
			1.0,
			openColor,
			false,
		)

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
