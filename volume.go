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

	// Use the same bar dimensions as Chart.Draw
	totalBarSpace := v.config.BarWidth + v.config.BarSpacing
	barWidth := float32(v.config.BarWidth * chart.Zoom) // Match Chart's bar width
	if barWidth < 1 {
		barWidth = 1 // Minimum width
	}

	// Draw volume bars
	for i, ohlcv := range chart.Data {
		x := float32(v.config.LeftMargin + (float64(i)*totalBarSpace*chart.Zoom + chart.OffsetX))

		// Center the bar exactly as in Chart.Draw
		x -= barWidth / 2

		// Skip bars outside the visible area, matching Chart.Draw
		if x+barWidth < float32(v.config.LeftMargin) || x > float32(v.config.Width-v.config.RightMargin) {
			continue
		}

		barHeight := (ohlcv.Volume / maxVolume) * volumeHeight
		y := volumeTop + (volumeHeight - barHeight)

		// Set color based on price movement
		var barColor color.RGBA
		if ohlcv.Close >= ohlcv.Open {
			barColor = v.config.VolumeUpColor
		} else {
			barColor = v.config.VolumeDownColor
		}

		// Draw volume bar
		vector.DrawFilledRect(
			screen,
			x,
			float32(y),
			barWidth,
			float32(barHeight),
			barColor,
			false,
		)
	}
}
