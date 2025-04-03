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
		frameTimes:   make([]float64, 0, 300),
		lastUpdate:   time.Now(),
		config:       config,
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
	rawX, rawY := float64(cx), float64(cy)

	i.crosshairX, i.crosshairY = i.getSnappedPosition(rawX, rawY, chart)

	i.showCrosshair = cx >= int(chart.config.LeftMargin) &&
		cx <= int(chart.config.Width-chart.config.RightMargin) &&
		cy >= int(chart.config.TopMargin) &&
		cy <= int(chart.config.Height-chart.config.BottomMargin)

	if i.showCrosshair {
		i.updatePriceAndTimeValues(chart)
	}
}

func (i *Interaction) getSnappedPosition(rawX, rawY float64, chart *Chart) (float64, float64) {
	if rawX < chart.config.LeftMargin || rawX > chart.config.Width-chart.config.RightMargin ||
		rawY < chart.config.TopMargin || rawY > chart.config.Height-chart.config.BottomMargin {
		return rawX, rawY
	}

	totalBarSpace := chart.config.BarWidth + chart.config.BarSpacing
	barCenters := make([]float64, len(chart.Data))
	for idx := range chart.Data {
		barCenters[idx] = chart.config.LeftMargin + (float64(idx)+0.5)*totalBarSpace*chart.Zoom + chart.OffsetX
	}

	snapThreshold := totalBarSpace * chart.Zoom * 0.5
	nearestX := rawX
	minDist := math.MaxFloat64

	for _, center := range barCenters {
		dist := math.Abs(rawX - center)
		if dist < snapThreshold && dist < minDist {
			minDist = dist
			nearestX = center
		}
	}

	if minDist < snapThreshold {
		return nearestX, rawY
	}
	return rawX, rawY
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
