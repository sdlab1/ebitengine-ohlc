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
		config:       config,
	}
}

func (i *Interaction) Update(chart *Chart) {
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
