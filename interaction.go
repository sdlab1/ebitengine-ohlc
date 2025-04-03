package main

import (
	"fmt"
	"math"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"golang.org/x/image/font"
	"golang.org/x/image/font/gofont/goregular"
	"golang.org/x/image/font/opentype"
)

type Interaction struct {
	crosshairX    float64
	crosshairY    float64
	mousePrice    float64
	mouseTime     int64
	showCrosshair bool
	fontFace      font.Face
	labelHeight   int
	labelPadding  int
	config        ChartConfig
	frameTimes    []float64
	lastUpdate    time.Time
	frameTimeMA   float64
}

func NewInteraction(config ChartConfig) *Interaction {
	tt, err := opentype.Parse(goregular.TTF)
	if err != nil {
		panic(err)
	}

	face, err := opentype.NewFace(tt, &opentype.FaceOptions{
		Size:    12,
		DPI:     72,
		Hinting: font.HintingFull,
	})
	if err != nil {
		panic(err)
	}

	metrics := face.Metrics()
	textHeight := metrics.Ascent.Ceil() + metrics.Descent.Ceil()
	return &Interaction{
		fontFace:     face,
		labelHeight:  textHeight,
		labelPadding: 5,
		frameTimes:   make([]float64, 0, 300), // ~5 seconds at 60fps
		lastUpdate:   time.Now(),
		config:       config,
	}
}

func (i *Interaction) Update(chart *Chart) {
	// Tracking performance in
	now := time.Now()
	frameTime := now.Sub(i.lastUpdate).Seconds() * 1000 // Convert to milliseconds
	i.lastUpdate = now

	// Maintain a 5-second window of frame times
	i.frameTimes = append(i.frameTimes, frameTime)
	if len(i.frameTimes) > 300 { // 60fps * 5s = 300
		i.frameTimes = i.frameTimes[1:]
	}

	// Calculate moving average
	sum := 0.0
	for _, ft := range i.frameTimes {
		sum += ft
	}
	i.frameTimeMA = sum / float64(len(i.frameTimes))

	cx, cy := ebiten.CursorPosition()
	i.crosshairX = float64(cx)
	i.crosshairY = float64(cy)
	i.showCrosshair = cx >= int(i.config.LeftMargin) && cx <= int(i.config.Width-i.config.RightMargin) &&
		cy >= int(i.config.TopMargin) && cy <= int(i.config.Height-i.config.BottomMargin)

	if i.showCrosshair {
		chartHeight := i.config.Height - i.config.TopMargin - i.config.BottomMargin
		priceRange := chart.priceMax - chart.priceMin
		i.mousePrice = chart.priceMax - ((i.crosshairY - i.config.TopMargin) / chartHeight * priceRange)

		chartWidth := i.config.Width - i.config.LeftMargin - i.config.RightMargin
		timeRange := float64(chart.timeEnd - chart.timeStart)
		i.mouseTime = chart.timeStart + int64(((i.crosshairX-i.config.LeftMargin)/chartWidth)*timeRange)
	}
}

func (i *Interaction) Draw(screen *ebiten.Image, chart *Chart) {
	// Draw frame time display first (so it's underneath crosshair)
	frametimeText := fmt.Sprintf("%.1f", i.frameTimeMA) // Just digits
	textWidth := font.MeasureString(i.fontFace, frametimeText).Ceil()
	textHeight := i.labelHeight

	// frame time display Background rectangle
	padding := 4
	rectWidth := textWidth + padding*2
	rectHeight := textHeight + padding*2
	rectX := float32(i.config.Width) - float32(rectWidth) - 2 // 2px from right edge
	rectY := float32(2)

	// Check if mouse is over the display
	cx, cy := ebiten.CursorPosition()
	showUnits := cx >= int(rectX) && cx <= int(rectX)+rectWidth &&
		cy >= int(rectY) && cy <= int(rectY)+rectHeight

	// Add units if hovering
	if showUnits {
		frametimeText += " ms"
		textWidth = font.MeasureString(i.fontFace, frametimeText).Ceil()
		rectWidth = textWidth + padding*2

		// Recalculate X position to keep right-aligned
		rectX = float32(i.config.Width) - float32(rectWidth) - 2
	}

	vector.DrawFilledRect(
		screen,
		rectX, // X position
		rectY, // Y position
		float32(rectWidth),
		float32(rectHeight),
		i.config.FrameTimeMABgColor,
		false,
	)

	// frame time display Text
	text.Draw(
		screen,
		frametimeText,
		i.fontFace,
		int(rectX)+padding,
		5+padding+textHeight/2, // Vertically centered
		i.config.FrameTimeMATextColor,
	)

	if !i.showCrosshair {
		return
	}

	// Draw crosshair lines using config color
	vector.StrokeLine(
		screen,
		float32(i.config.LeftMargin), float32(i.crosshairY),
		float32(i.config.Width-i.config.RightMargin), float32(i.crosshairY),
		1,
		i.config.CrosshairColor,
		false,
	)

	vector.StrokeLine(
		screen,
		float32(i.crosshairX), float32(i.config.TopMargin),
		float32(i.crosshairX), float32(i.config.Height-i.config.BottomMargin),
		1,
		i.config.CrosshairColor,
		false,
	)

	// Draw price label
	priceText := fmt.Sprintf("%.2f", i.mousePrice)
	priceTextWidth := font.MeasureString(i.fontFace, priceText).Ceil()
	priceTextX := int(i.config.LeftMargin) - priceTextWidth - i.labelPadding*2
	priceTextY := int(i.crosshairY) + i.labelHeight/2

	vector.DrawFilledRect(
		screen,
		float32(priceTextX-i.labelPadding),
		float32(priceTextY-i.labelHeight),
		float32(priceTextWidth+i.labelPadding*2),
		float32(i.labelHeight+i.labelPadding),
		i.config.CrosshairBgColor,
		false,
	)

	text.Draw(
		screen,
		priceText,
		i.fontFace,
		priceTextX,
		priceTextY,
		i.config.CrosshairTextColor,
	)

	// Draw time label
	timeText := time.Unix(i.mouseTime/1000, 0).Format("2006-01-02 15:04:05")
	timeTextWidth := font.MeasureString(i.fontFace, timeText).Ceil()
	timeTextX := int(math.Max(
		i.config.LeftMargin,
		math.Min(
			i.crosshairX-float64(timeTextWidth/2),
			i.config.Width-i.config.RightMargin-float64(timeTextWidth),
		),
	))
	timeTextY := int(i.config.Height-i.config.BottomMargin) + i.labelHeight + i.labelPadding*2

	vector.DrawFilledRect(
		screen,
		float32(timeTextX-i.labelPadding),
		float32(timeTextY-i.labelHeight-i.labelPadding),
		float32(timeTextWidth+i.labelPadding*2),
		float32(i.labelHeight+i.labelPadding),
		i.config.CrosshairBgColor,
		false,
	)

	text.Draw(
		screen,
		timeText,
		i.fontFace,
		timeTextX,
		timeTextY,
		i.config.CrosshairTextColor,
	)
}
