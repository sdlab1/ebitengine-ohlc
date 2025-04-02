package main

import (
	"fmt"
	"image/color"
	"math"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"golang.org/x/image/font/basicfont"
)

type Axes struct {
	fontFace *basicfont.Face
}

func NewAxes() *Axes {
	return &Axes{
		fontFace: basicfont.Face7x13,
	}
}

func (a *Axes) Update(chart *Chart) {
	// No need to store anything as we'll use chart reference directly
}

func calculateStep(rangeSize float64) float64 {
	exponent := math.Floor(math.Log10(rangeSize))
	power := math.Pow(10, exponent)
	frac := rangeSize / power

	var step float64
	switch {
	case frac > 5:
		step = power
	case frac > 2:
		step = power / 2
	default:
		step = power / 5
	}

	return step
}

func calculateTimeStep(timeRange int64) int64 {
	days := timeRange / (24 * 3600 * 1000)

	switch {
	case days > 365:
		return 30 * 24 * 3600 * 1000 // Monthly
	case days > 30:
		return 7 * 24 * 3600 * 1000 // Weekly
	case days > 7:
		return 24 * 3600 * 1000 // Daily
	default:
		return 3600 * 1000 // Hourly
	}
}

func (a *Axes) Draw(screen *ebiten.Image, chart *Chart) {
	// Draw background
	vector.DrawFilledRect(screen, 0, 0, 1000, 700, color.RGBA{20, 20, 40, 255}, false)

	// Draw Y axis (price)
	vector.StrokeLine(screen, 100, 50, 100, 600, 1, color.White, false)

	// Draw X axis (time)
	vector.StrokeLine(screen, 100, 600, 950, 600, 1, color.White, false)

	// Calculate price steps
	priceRange := chart.priceMax - chart.priceMin
	priceStep := calculateStep(priceRange / 10)

	// Draw price ticks and labels
	for price := math.Ceil(chart.priceMin/priceStep) * priceStep; price <= chart.priceMax; price += priceStep {
		y := 600 - ((price - chart.priceMin) / (chart.priceMax - chart.priceMin) * 500)
		vector.StrokeLine(screen, 95, float32(y), 100, float32(y), 1, color.White, false)
		text.Draw(screen, fmt.Sprintf("%.2f", price), a.fontFace, 50, int(y+6), color.White)
	}

	// Calculate time steps
	timeStep := calculateTimeStep(chart.timeEnd - chart.timeStart)

	// Draw time ticks and labels
	for t := chart.timeStart; t <= chart.timeEnd; t += timeStep {
		x := 100 + ((float64(t-chart.timeStart) / float64(chart.timeEnd-chart.timeStart)) * 850)
		vector.StrokeLine(screen, float32(x), 600, float32(x), 605, 1, color.White, false)

		timeText := time.Unix(t/1000, 0).Format("2006-01-02")
		if timeStep < 24*3600*1000 {
			timeText = time.Unix(t/1000, 0).Format("01-02 15:04")
		}
		text.Draw(screen, timeText, a.fontFace, int(x-30), 620, color.White)
	}
}
