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
	Zoom      float64
	prevX     int
	prevY     int
	priceMin  float64
	priceMax  float64
	timeStart int64
	timeEnd   int64
	ts_from   int64 // Start timestamp for displayed bars
	ts_to     int64 // End timestamp for displayed bars
	config    ChartConfig
}

func NewChart(config ChartConfig) *Chart {
	return &Chart{
		Zoom:   1.0,
		config: config,
		Data:   make([]OHLCV, 0),
	}
}

// UpdateData updates the chart data and sets the display range
func (c *Chart) UpdateData(newData []OHLCV) {
	if len(newData) == 0 {
		return
	}

	c.Data = newData
	// Recalculate price range for scaling
	c.priceMin, c.priceMax = calculatePriceRange(c.Data)
	c.timeStart = c.Data[0].Time
	c.timeEnd = c.Data[len(c.Data)-1].Time

	// Set ts_from and ts_to to display the last N bars based on visible width
	visibleWidth := c.config.Width - c.config.LeftMargin - c.config.RightMargin
	totalBarSpace := c.config.BarWidth + c.config.BarSpacing
	maxBars := int(visibleWidth / (totalBarSpace * c.Zoom))

	if len(c.Data) > maxBars {
		c.ts_from = c.Data[len(c.Data)-maxBars].Time
	} else {
		c.ts_from = c.Data[0].Time
	}
	c.ts_to = c.Data[len(c.Data)-1].Time // Last bar is set as ts_to
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

// Update handles zooming and panning interactions
func (c *Chart) Update() error {
	// Handle mouse wheel zoom
	_, dy := ebiten.Wheel()
	if dy != 0 {
		cx, _ := ebiten.CursorPosition()
		chartLeft := int(c.config.LeftMargin)
		chartRight := int(c.config.Width - c.config.RightMargin)

		if cx >= chartLeft && cx <= chartRight {
			totalBarSpace := c.config.BarWidth + c.config.BarSpacing
			// Apply zoom
			newZoom := c.Zoom * math.Pow(1.1, dy)
			newZoom = math.Max(0.1, math.Min(newZoom, 10.0))
			c.Zoom = newZoom

			// Recalculate ts_from based on new zoom level
			visibleWidth := c.config.Width - c.config.LeftMargin - c.config.RightMargin
			maxBars := int(visibleWidth / (totalBarSpace * c.Zoom))
			if len(c.Data) > maxBars {
				c.ts_from = c.Data[len(c.Data)-maxBars].Time
			} else {
				c.ts_from = c.Data[0].Time
			}
			c.ts_to = c.Data[len(c.Data)-1].Time
		}
	}

	// Handle mouse drag panning
	if ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
		cx, cy := ebiten.CursorPosition()
		if c.prevX != 0 || c.prevY != 0 {
			dx := float64(cx - c.prevX)
			totalBarSpace := c.config.BarWidth + c.config.BarSpacing
			shiftBars := int(dx / (totalBarSpace * c.Zoom))
			if shiftBars != 0 {
				// Adjust ts_from based on drag
				for i, d := range c.Data {
					if d.Time == c.ts_from {
						newIndex := i - shiftBars
						if newIndex < 0 {
							newIndex = 0
						} else if newIndex >= len(c.Data) {
							newIndex = len(c.Data) - 1
						}
						c.ts_from = c.Data[newIndex].Time
						break
					}
				}
			}
		}
		c.prevX, c.prevY = cx, cy
	} else {
		c.prevX, c.prevY = 0, 0
	}

	return nil
}

// Draw renders the chart starting from ts_from and draws bars until there's no space
func (c *Chart) Draw(screen *ebiten.Image) {
	totalBarSpace := c.config.BarWidth + c.config.BarSpacing
	//visibleWidth := c.config.Width - c.config.LeftMargin - c.config.RightMargin

	// Find the index of the first bar with Time >= ts_from
	startIndex := -1
	for i, d := range c.Data {
		if d.Time >= c.ts_from {
			startIndex = i
			break
		}
	}
	if startIndex == -1 {
		return // No bars to display
	}

	// Draw bars starting from startIndex until there's no more space
	for i := startIndex; i < len(c.Data); i++ {
		ohlcv := c.Data[i]
		x := float32(c.config.LeftMargin + float64(i-startIndex)*totalBarSpace*c.Zoom)

		// Stop drawing if the bar exceeds the visible area
		if x > float32(c.config.Width-c.config.RightMargin) {
			break
		}

		barTop := float32(c.config.Height - c.config.BottomMargin - ((ohlcv.High - c.priceMin) / (c.priceMax - c.priceMin) * float64(c.config.Height-c.config.TopMargin-c.config.BottomMargin)))
		barBottom := float32(c.config.Height - c.config.BottomMargin - ((ohlcv.Low - c.priceMin) / (c.priceMax - c.priceMin) * float64(c.config.Height-c.config.TopMargin-c.config.BottomMargin)))

		// Draw the bar (high to low line)
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

		// Draw open and close ticks
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

func (c *Chart) GetBarPosition(index int) (left, center, right float64) {
	totalBarSpace := (c.config.BarWidth + c.config.BarSpacing) * c.Zoom
	left = c.config.LeftMargin + (float64(index) * totalBarSpace)
	right = left + c.config.BarWidth*c.Zoom
	center = left + (c.config.BarWidth*c.Zoom)/2
	return left, center, right
}
