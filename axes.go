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

func calculateNiceStep(rangeSize float64, availableSpace float64, minSpacing float64) float64 {
	// Calculate maximum possible labels
	maxPossibleLabels := math.Floor(availableSpace / minSpacing)
	if maxPossibleLabels < 2 {
		maxPossibleLabels = 2
	}

	// Initial step estimate
	rawStep := rangeSize / maxPossibleLabels

	// Round to nice human-readable number
	magnitude := math.Pow(10, math.Floor(math.Log10(rawStep)))
	step := math.Ceil(rawStep/magnitude) * magnitude

	// Ensure step isn't too small
	minStep := minSpacing * (rangeSize / availableSpace)
	if step < minStep {
		step = minStep
	}

	return step
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
	chartWidth := float64(a.config.Width - a.config.LeftMargin - a.config.RightMargin)
	chartHeight := float64(a.config.Height - a.config.TopMargin - a.config.BottomMargin)

	// Draw Y axis
	vector.StrokeLine(
		screen,
		float32(a.config.LeftMargin), float32(a.config.TopMargin),
		float32(a.config.LeftMargin), float32(a.config.Height-a.config.BottomMargin),
		a.config.AxisWidth,
		a.config.AxisColor,
		false,
	)

	// Calculate price steps with guaranteed spacing
	priceRange := chart.priceMax - chart.priceMin
	minSpacing := float64(a.labelHeight) * a.config.MinLabelSpacing

	priceStep := calculateNiceStep(priceRange, chartHeight, minSpacing)

	// Draw price labels and grid lines
	prevY := math.Inf(-1)
	for price := math.Ceil(chart.priceMin/priceStep) * priceStep; price <= chart.priceMax; price += priceStep {
		y := a.config.Height - a.config.BottomMargin - ((price - chart.priceMin) / priceRange * chartHeight)

		// Ensure minimum spacing
		if prevY == math.Inf(-1) || (prevY-y) >= minSpacing {
			// Draw grid line
			vector.StrokeLine(
				screen,
				float32(a.config.LeftMargin), float32(y),
				float32(a.config.Width-a.config.RightMargin), float32(y),
				a.config.GridWidth,
				a.config.GridColor,
				false,
			)

			// Format price label
			priceStr := formatPriceLabel(price)

			// Measure and draw text
			textWidth := font.MeasureString(a.fontFace, priceStr).Ceil()
			text.Draw(
				screen,
				priceStr,
				a.fontFace,
				int(a.config.LeftMargin)-60-textWidth/2,
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
	timeStep := timeRange / 10

	for t := chart.timeStart; t <= chart.timeEnd; t += timeStep {
		x := a.config.LeftMargin + (float64(t-chart.timeStart) / float64(timeRange) * chartWidth)
		vector.StrokeLine(
			screen,
			float32(x), float32(a.config.TopMargin),
			float32(x), float32(a.config.Height-a.config.BottomMargin),
			a.config.GridWidth,
			a.config.GridColor,
			false,
		)

		tm := time.Unix(t/1000, 0)
		currentMonth := int(tm.Month())

		var timeText string
		if currentMonth != a.lastMonth {
			timeText = tm.Format(TimeFormat.MonthlyFormat)
			a.lastMonth = currentMonth
		} else {
			timeText = tm.Format(TimeFormat.DailyFormat)
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
