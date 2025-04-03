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
	crosshairX    float64 // Current x position (only changes when snapping)
	crosshairY    float64 // Current y position (updates freely)
	snappedBarIdx int     // Index of bar we're snapped to (-1 if not snapped)
	fontFace      font.Face
	labelHeight   int
	labelPadding  int
	config        ChartConfig
	mousePrice    float64
	mouseTime     int64
	showCrosshair bool
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
		fontFace:      face,
		labelHeight:   textHeight,
		labelPadding:  5,
		frameTimes:    make([]float64, 0, 300),
		lastUpdate:    time.Now(),
		config:        config,
		snappedBarIdx: -1,
	}
}

func (i *Interaction) Update(chart *Chart) {
	i.updateFrameTimes()
	i.updateCrosshairPosition(chart)
}

func (i *Interaction) updateFrameTimes() {
	now := time.Now()
	frameTime := now.Sub(i.lastUpdate).Seconds() * 1000
	i.lastUpdate = now

	i.frameTimes = append(i.frameTimes, frameTime)
	if len(i.frameTimes) > 300 {
		i.frameTimes = i.frameTimes[1:]
	}

	sum := 0.0
	for _, ft := range i.frameTimes {
		sum += ft
	}
	i.frameTimeMA = sum / float64(len(i.frameTimes))
}

func (i *Interaction) updateCrosshairPosition(chart *Chart) {
	cx, cy := ebiten.CursorPosition()
	mouseX, mouseY := float64(cx), float64(cy)

	// Update both positions freely
	i.crosshairX = mouseX
	i.crosshairY = mouseY

	// Calculate chart boundaries
	chartLeft := chart.config.LeftMargin
	chartRight := chart.config.Width - chart.config.RightMargin
	chartTop := chart.config.TopMargin
	chartBottom := chart.config.Height - chart.config.BottomMargin

	// Check if cursor is within chart bounds
	i.showCrosshair = mouseX >= chartLeft && mouseX <= chartRight &&
		mouseY >= chartTop && mouseY <= chartBottom

	if i.showCrosshair {
		i.updatePriceAndTimeValues(chart)
	}
}

func clamp(value, min, max int) int {
	if min > max {
		min, max = max, min
	}
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}

func (i *Interaction) updatePriceAndTimeValues(chart *Chart) {
	chartHeight := chart.config.Height - chart.config.TopMargin - chart.config.BottomMargin
	priceRange := chart.priceMax - chart.priceMin
	i.mousePrice = chart.priceMax - ((i.crosshairY - chart.config.TopMargin) / chartHeight * priceRange)

	chartWidth := chart.config.Width - chart.config.LeftMargin - chart.config.RightMargin
	timeRange := float64(chart.timeEnd - chart.timeStart)
	i.mouseTime = chart.timeStart + int64(((i.crosshairX-chart.config.LeftMargin)/chartWidth)*timeRange)
}

func (i *Interaction) Draw(screen *ebiten.Image, chart *Chart) {
	i.drawFrameTimeDisplay(screen)
	if !i.showCrosshair {
		return
	}
	i.drawCrosshair(screen, chart)
}

func (i *Interaction) drawFrameTimeDisplay(screen *ebiten.Image) {
	frametimeText := fmt.Sprintf("%.1f", i.frameTimeMA)
	textWidth := font.MeasureString(i.fontFace, frametimeText).Ceil()
	textHeight := i.labelHeight

	padding := 4
	rectWidth := textWidth + padding*2
	rectHeight := textHeight + padding*2
	rectX := float32(i.config.Width) - float32(rectWidth) - 2
	rectY := float32(2)

	cx, cy := ebiten.CursorPosition()
	showUnits := cx >= int(rectX) && cx <= int(rectX)+rectWidth &&
		cy >= int(rectY) && cy <= int(rectY)+rectHeight

	if showUnits {
		frametimeText += " ms"
		textWidth = font.MeasureString(i.fontFace, frametimeText).Ceil()
		rectWidth = textWidth + padding*2
		rectX = float32(i.config.Width) - float32(rectWidth) - 2
	}

	vector.DrawFilledRect(
		screen,
		rectX,
		rectY,
		float32(rectWidth),
		float32(rectHeight),
		i.config.FrameTimeMABgColor,
		false,
	)

	text.Draw(
		screen,
		frametimeText,
		i.fontFace,
		int(rectX)+padding,
		5+padding+textHeight/2,
		i.config.FrameTimeMATextColor,
	)
}

func (i *Interaction) drawCrosshair(screen *ebiten.Image, chart *Chart) {
	// Draw perfectly centered vertical line
	vector.StrokeLine(
		screen,
		float32(i.crosshairX),
		float32(i.config.TopMargin),
		float32(i.crosshairX),
		float32(i.config.Height-i.config.BottomMargin),
		1,
		i.config.CrosshairColor,
		false,
	)

	// Draw horizontal line
	vector.StrokeLine(
		screen,
		float32(i.config.LeftMargin),
		float32(i.crosshairY),
		float32(i.config.Width-i.config.RightMargin),
		float32(i.crosshairY),
		1,
		i.config.CrosshairColor,
		false,
	)

	i.drawPriceLabel(screen)
	i.drawTimeLabel(screen)
}

func (i *Interaction) drawPriceLabel(screen *ebiten.Image) {
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
}

func (i *Interaction) drawTimeLabel(screen *ebiten.Image) {
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
