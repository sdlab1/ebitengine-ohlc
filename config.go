package main

import (
	"image/color"
)

type ChartConfig struct {
	// Colors
	BackgroundColor color.RGBA
	AxisColor       color.RGBA
	GridColor       color.RGBA
	LabelColor      color.RGBA
	// Grid colors
	PrimaryGridColor   color.RGBA // Vertical grid
	SecondaryGridColor color.RGBA // Horizontal grid

	// Bar colors
	BarColor   color.RGBA
	OpenColor  color.RGBA
	CloseColor color.RGBA

	CrosshairColor     color.RGBA
	CrosshairTextColor color.RGBA
	CrosshairBgColor   color.RGBA

	// Dimensions
	Width        float64
	Height       float64
	LeftMargin   float64
	RightMargin  float64
	TopMargin    float64
	BottomMargin float64

	// Spacing
	MinLabelSpacing float64 // Multiple of text height
	MinPriceLabels  int

	// Bar configuration
	BarWidth   float64
	BarSpacing float64 // New spacing between bars

	// Appearance
	AxisWidth float32
	GridWidth float32
}

var DefaultConfig = ChartConfig{
	BackgroundColor:    color.RGBA{R: 10, G: 10, B: 10, A: 255},
	AxisColor:          color.RGBA{R: 100, G: 100, B: 100, A: 255},
	GridColor:          color.RGBA{R: 50, G: 50, B: 50, A: 255},    // Pure black
	LabelColor:         color.RGBA{R: 255, G: 255, B: 255, A: 255}, // Pure white
	PrimaryGridColor:   color.RGBA{R: 15, G: 15, B: 15, A: 255},    // Darker vertical grid
	SecondaryGridColor: color.RGBA{R: 35, G: 35, B: 35, A: 255},    // Darker horizontal grid
	BarColor:           color.RGBA{R: 255, G: 255, B: 255, A: 255},
	OpenColor:          color.RGBA{R: 255, G: 255, B: 255, A: 255},
	CloseColor:         color.RGBA{R: 255, G: 255, B: 255, A: 255},
	CrosshairColor:     color.RGBA{R: 150, G: 150, B: 150, A: 100},
	CrosshairTextColor: color.RGBA{R: 200, G: 200, B: 200, A: 255},
	CrosshairBgColor:   color.RGBA{R: 40, G: 40, B: 60, A: 200},
	BarWidth:           1.0,
	BarSpacing:         5.0, // Space between bars
	Width:              1000,
	Height:             700,
	LeftMargin:         80,
	RightMargin:        50,
	TopMargin:          30,
	BottomMargin:       50,

	MinLabelSpacing: 4, // 4x text height
	MinPriceLabels:  5,

	AxisWidth: 1.0,
	GridWidth: 1.2,
}

type TimeFormatConfig struct {
	MonthlyFormat string
	DailyFormat   string
	DefaultFormat string
}

var TimeFormat = TimeFormatConfig{
	MonthlyFormat: "Jan 2",
	DailyFormat:   "2",
	DefaultFormat: "Jan 2 15:04",
}
