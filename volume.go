package main

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

type Volume struct {
	config ChartConfig
}

func NewVolume(config ChartConfig) *Volume {
	return &Volume{config: config}
}

func (v *Volume) Draw(screen *ebiten.Image, chart *Chart) {
	// Calculate volume area dimensions (20% of chart height)
	volumeHeight := (v.config.Height - v.config.TopMargin - v.config.BottomMargin) * 0.2
	volumeTop := v.config.Height - v.config.BottomMargin - volumeHeight

	// Find max volume for scaling
	maxVolume := 0.0
	for _, d := range chart.Data {
		if d.Volume > maxVolume {
			maxVolume = d.Volume
		}
	}
	if maxVolume == 0 {
		return
	}

	// Calculate bar dimensions to match OHLC bars
	totalBarSpace := v.config.BarWidth + v.config.BarSpacing
	volumeBarWidth := v.config.BarSpacing - v.config.VolumeSpacing
	if volumeBarWidth < 1 {
		volumeBarWidth = 1 // Ensure minimum width
	}

	// Draw volume bars
	for i, ohlcv := range chart.Data {
		x := v.config.LeftMargin + (float64(i)*totalBarSpace*chart.Zoom + chart.OffsetX)
		if x < v.config.LeftMargin || x > v.config.Width-v.config.RightMargin {
			continue
		}

		barHeight := (ohlcv.Volume / maxVolume) * volumeHeight
		y := volumeTop + (volumeHeight - barHeight)

		var barColor color.RGBA
		if ohlcv.Close >= ohlcv.Open {
			barColor = v.config.VolumeUpColor
		} else {
			barColor = v.config.VolumeDownColor
		}

		// Center the volume bar within the available spacing
		barX := float32(x) + float32((v.config.BarSpacing-volumeBarWidth)/2)

		vector.DrawFilledRect(
			screen,
			barX,
			float32(y),
			float32(volumeBarWidth),
			float32(barHeight),
			barColor,
			false,
		)
	}
}
