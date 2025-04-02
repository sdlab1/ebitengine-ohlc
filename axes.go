package main

import (
	"fmt"
	"math"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
)

type Axes struct {
	fontFace    font.Face
	lastMonth   int
	labelHeight int
	config      ChartConfig
}

func NewAxes(config ChartConfig) *Axes {
	face := basicfont.Face7x13
	metrics := face.Metrics()
	textHeight := metrics.Ascent.Ceil() + metrics.Descent.Ceil()
	return &Axes{
		fontFace:    face,
		lastMonth:   -1,
		labelHeight: textHeight,
		config:      config,
	}
}

func (a *Axes) Update(chart *Chart) {
	a.lastMonth = -1
}

func (a *Axes) Draw(screen *ebiten.Image, chart *Chart) {
	// Draw background
	vector.DrawFilledRect(
		screen,
		0, 0,
		float32(a.config.Width), float32(a.config.Height),
		a.config.BackgroundColor,
		false,
	)

	// Calculate chart dimensions
	chartWidth := a.config.Width - a.config.LeftMargin - a.config.RightMargin
	chartHeight := a.config.Height - a.config.TopMargin - a.config.BottomMargin

	// Draw Y axis
	vector.StrokeLine(
		screen,
		float32(a.config.LeftMargin), float32(a.config.TopMargin),
		float32(a.config.LeftMargin), float32(a.config.Height-a.config.BottomMargin),
		a.config.AxisWidth,
		a.config.AxisColor,
		false,
	)

	// Draw price labels and horizontal grid lines
	priceRange := chart.priceMax - chart.priceMin
	priceStep := calculateStep(priceRange/float64(a.config.MinPriceLabels), a.config.MinPriceLabels)
	minSpacing := float64(a.labelHeight) * a.config.MinLabelSpacing
	prevY := math.Inf(-1)

	for price := math.Ceil(chart.priceMin/priceStep) * priceStep; price <= chart.priceMax; price += priceStep {
		y := a.config.Height - a.config.BottomMargin - ((price - chart.priceMin) / priceRange * chartHeight)

		// Ensure minimum spacing between labels
		if prevY == math.Inf(-1) || (prevY-y) >= minSpacing {
			// Draw horizontal grid line
			vector.StrokeLine(
				screen,
				float32(a.config.LeftMargin), float32(y),
				float32(a.config.Width-a.config.RightMargin), float32(y),
				0.5,
				a.config.SecondaryGridColor,
				false,
			)

			// Draw price label
			priceStr := formatPriceLabel(price)
			textWidth := font.MeasureString(a.fontFace, priceStr).Ceil()
			text.Draw(
				screen,
				priceStr,
				a.fontFace,
				int(a.config.LeftMargin)-textWidth-10, // 10px padding from axis
				int(y)+a.labelHeight/2,
				a.config.LabelColor,
			)
			prevY = y
		}
	}

	// Draw X axis and vertical grid lines
	vector.StrokeLine(
		screen,
		float32(a.config.LeftMargin), float32(a.config.Height-a.config.BottomMargin),
		float32(a.config.Width-a.config.RightMargin), float32(a.config.Height-a.config.BottomMargin),
		a.config.AxisWidth,
		a.config.AxisColor,
		false,
	)

	timeRange := chart.timeEnd - chart.timeStart
	timeStep := timeRange / 12 // 12 vertical divisions

	for t := chart.timeStart; t <= chart.timeEnd; t += timeStep {
		x := a.config.LeftMargin + (float64(t-chart.timeStart) / float64(timeRange) * chartWidth)

		// Draw vertical grid line
		vector.StrokeLine(
			screen,
			float32(x), float32(a.config.TopMargin),
			float32(x), float32(a.config.Height-a.config.BottomMargin),
			1.2,
			a.config.PrimaryGridColor,
			false,
		)

		// Draw time label
		tm := time.Unix(t/1000, 0)
		currentMonth := int(tm.Month())

		var timeText string
		if currentMonth != a.lastMonth {
			timeText = tm.Format("Jan 2")
			a.lastMonth = currentMonth
		} else {
			timeText = tm.Format("15:04")
		}

		textWidth := font.MeasureString(a.fontFace, timeText).Ceil()
		text.Draw(
			screen,
			timeText,
			a.fontFace,
			int(x)-textWidth/2,
			int(a.config.Height)-20,
			a.config.LabelColor,
		)
	}
}

func calculateStep(rangeSize float64, minLabels int) float64 {
	exponent := math.Floor(math.Log10(rangeSize))
	power := math.Pow(10, exponent)
	return power
}

func formatPriceLabel(price float64) string {
	switch {
	case price >= 1000000:
		return fmt.Sprintf("%.1fM", price/1000000)
	case price >= 1000:
		return fmt.Sprintf("%.0fK", price/1000)
	default:
		return fmt.Sprintf("%.0f", price)
	}
}
