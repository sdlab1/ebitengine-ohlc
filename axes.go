package main

import (
	"fmt"
	"image/color"
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
}

func NewAxes() *Axes {
	face := basicfont.Face7x13
	metrics := face.Metrics()
	textHeight := metrics.Ascent.Ceil() + metrics.Descent.Ceil()
	return &Axes{
		fontFace:    face,
		lastMonth:   -1,
		labelHeight: textHeight,
	}
}

func (a *Axes) Update(chart *Chart) {
	a.lastMonth = -1
}

func calculateStep(rangeSize float64) float64 {
	exponent := math.Floor(math.Log10(rangeSize))
	power := math.Pow(10, exponent)
	return power * 2
}

func (a *Axes) Draw(screen *ebiten.Image, chart *Chart) {
	// Draw background
	vector.DrawFilledRect(screen, 0, 0, 1000, 700, color.RGBA{40, 40, 60, 255}, false)

	// Chart dimensions
	leftMargin := 80.0
	rightMargin := 50.0
	bottomMargin := 50.0
	topMargin := 30.0
	chartHeight := 600.0 - topMargin - bottomMargin

	// Colors
	axisColor := color.RGBA{100, 100, 100, 255}
	gridColor := color.Black
	labelColor := color.White

	// Draw Y axis
	vector.StrokeLine(screen, float32(leftMargin), float32(topMargin), float32(leftMargin), float32(600-bottomMargin), 1, axisColor, false)

	// Calculate price steps with proper spacing
	priceRange := chart.priceMax - chart.priceMin
	minLabels := 5 // Minimum number of price labels we want
	priceStep := calculateStep(priceRange / float64(minLabels))

	// Ensure we have enough labels with good spacing
	pixelsPerUnit := chartHeight / priceRange
	minSpacing := float64(a.labelHeight * 4)
	for (priceStep * pixelsPerUnit) < minSpacing {
		priceStep *= 2
	}

	// Draw all price labels (no skipping)
	for price := math.Ceil(chart.priceMin/priceStep) * priceStep; price <= chart.priceMax; price += priceStep {
		y := 600 - bottomMargin - ((price - chart.priceMin) / priceRange * chartHeight)

		// Draw grid line
		vector.StrokeLine(screen, float32(leftMargin), float32(y), float32(950-rightMargin), float32(y), 1.2, gridColor, false)

		// Format price label
		priceStr := fmt.Sprintf("%.0f", price)
		if price >= 1000000 {
			priceStr = fmt.Sprintf("%.1fM", price/1000000)
		} else if price >= 1000 {
			priceStr = fmt.Sprintf("%.0fK", price/1000)
		}

		// Measure and draw text
		textWidth := font.MeasureString(a.fontFace, priceStr).Ceil()
		text.Draw(screen, priceStr, a.fontFace, int(leftMargin)-60-textWidth/2, int(y)+a.labelHeight/2, labelColor)
	}

	// Draw X axis and vertical grid lines
	vector.StrokeLine(screen, float32(leftMargin), float32(600-bottomMargin), float32(950-rightMargin), float32(600-bottomMargin), 1, axisColor, false)

	timeRange := chart.timeEnd - chart.timeStart
	timeStep := timeRange / 10

	for t := chart.timeStart; t <= chart.timeEnd; t += timeStep {
		x := leftMargin + (float64(t-chart.timeStart)/float64(timeRange))*(950-leftMargin-rightMargin)
		vector.StrokeLine(screen, float32(x), float32(topMargin), float32(x), float32(600-bottomMargin), 1.2, gridColor, false)

		tm := time.Unix(t/1000, 0)
		currentMonth := int(tm.Month())

		var timeText string
		if currentMonth != a.lastMonth {
			timeText = tm.Format("Jan 2")
			a.lastMonth = currentMonth
		} else {
			timeText = tm.Format("2")
		}

		// Measure text width properly
		textWidth := font.MeasureString(a.fontFace, timeText).Ceil()
		text.Draw(screen, timeText, a.fontFace, int(x)-textWidth/2, 620, labelColor)
	}
}
