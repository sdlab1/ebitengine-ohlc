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

// Draw renders volume bars synchronized with the chart's ts_from and ts_to
func (v *Volume) Draw(screen *ebiten.Image, chart *Chart) {
	// Calculate volume area dimensions (20% of chart height)
	volumeHeight := (v.config.Height - v.config.TopMargin - v.config.BottomMargin) * 0.2
	volumeTop := v.config.Height - v.config.BottomMargin - volumeHeight

	// Find max volume for scaling within the displayed range
	maxVolume := 0.0
	for _, d := range chart.Data {
		if d.Time >= chart.ts_from && d.Volume > maxVolume {
			maxVolume = d.Volume
		}
	}
	if maxVolume == 0 {
		return
	}

	// Calculate bar dimensions to match OHLC bars
	totalBarSpace := v.config.BarWidth + v.config.BarSpacing
	volumeBarWidth := (chart.config.BarSpacing * chart.Zoom) - v.config.VolumeSpacing
	if volumeBarWidth < 1 {
		volumeBarWidth = 1 // Ensure minimum width
	}

	// Find the index of the first bar with Time >= ts_from
	startIndex := -1
	for i, d := range chart.Data {
		if d.Time >= chart.ts_from {
			startIndex = i
			break
		}
	}
	if startIndex == -1 {
		return // No bars to display
	}

	// Draw volume bars starting from startIndex until there's no more space
	for i := startIndex; i < len(chart.Data); i++ {
		ohlcv := chart.Data[i]
		x := float32(v.config.LeftMargin + float64(i-startIndex)*totalBarSpace*chart.Zoom)

		// Stop drawing if the bar exceeds the visible area
		if x > float32(v.config.Width-v.config.RightMargin) {
			break
		}

		barHeight := (ohlcv.Volume / maxVolume) * volumeHeight
		y := volumeTop + (volumeHeight - barHeight)

		var barColor color.RGBA
		if ohlcv.Close >= ohlcv.Open {
			barColor = v.config.VolumeUpColor
		} else {
			barColor = v.config.VolumeDownColor
		}

		// Draw volume bar centered under the corresponding OHLC bar
		vector.DrawFilledRect(
			screen,
			x+float32((chart.config.BarSpacing*chart.Zoom-volumeBarWidth)/2),
			float32(y),
			float32(volumeBarWidth),
			float32(barHeight),
			barColor,
			false,
		)
	}
}
