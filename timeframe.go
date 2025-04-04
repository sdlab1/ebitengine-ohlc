package main

import (
	"fmt"
	"time"

	"github.com/akrylysov/pogreb"
)

type Timeframe struct {
	db *pogreb.DB
}

func NewTimeframe(db *pogreb.DB) *Timeframe {
	return &Timeframe{db: db}
}

func (tf *Timeframe) Get15MinBars() ([]OHLCV, error) {
	now := time.Now().UTC()
	endTimeMs := now.Unix() * 1000
	startTimeMs := endTimeMs - (300 * 15 * 60 * 1000) // 300 bars * 15 minutes

	// Align start time to the nearest 15-minute boundary
	startTimeMs = (startTimeMs / (15 * 60 * 1000)) * (15 * 60 * 1000)

	// Fetch 1-minute data from the database (4500 minutes = 300 * 15)
	var minuteData []OHLCV
	for t := startTimeMs; t < endTimeMs; t += 60 * 1000 {
		key := int64ToBytes(t)
		value, err := tf.db.Get(key)
		if err != nil {
			fmt.Printf("Warning: Failed to read from db at %d: %v\n", t, err)
			continue
		}
		if value == nil {
			fmt.Printf("Warning: Missing data at %d (%s)\n", t, time.Unix(t/1000, 0).UTC())
			continue
		}
		ohlcv, err := deserializeOHLCV(value)
		if err != nil {
			fmt.Printf("Warning: Failed to deserialize OHLCV at %d: %v\n", t, err)
			continue
		}
		minuteData = append(minuteData, ohlcv)
	}

	// If we have no data at all, return early
	if len(minuteData) == 0 {
		return nil, fmt.Errorf("no data available in requested timeframe")
	}

	// Aggregate into 15-minute bars
	var bars []OHLCV
	currentBarEnd := minuteData[0].Time + 15*60*1000
	var currentBar *OHLCV
	foundIncompleteBar := false

	for i := 0; i < len(minuteData); i++ {
		// Check if we need to start a new bar
		if currentBar == nil || minuteData[i].Time >= currentBarEnd {
			if currentBar != nil && foundIncompleteBar {
				// Add the incomplete bar only if it has at least 5 minutes of data
				bars = append(bars, *currentBar)
				foundIncompleteBar = false
			}

			currentBar = &OHLCV{
				Time:   minuteData[i].Time,
				Open:   minuteData[i].Open,
				High:   minuteData[i].High,
				Low:    minuteData[i].Low,
				Close:  minuteData[i].Close,
				Volume: minuteData[i].Volume,
			}
			currentBarEnd = minuteData[i].Time + 15*60*1000
			continue
		}

		// Update existing bar
		if minuteData[i].High > currentBar.High {
			currentBar.High = minuteData[i].High
		}
		if minuteData[i].Low < currentBar.Low {
			currentBar.Low = minuteData[i].Low
		}
		currentBar.Close = minuteData[i].Close
		currentBar.Volume += minuteData[i].Volume
		foundIncompleteBar = true
	}

	// Add the last bar if meaningful
	if currentBar != nil && (foundIncompleteBar || len(bars) == 0) {
		bars = append(bars, *currentBar)
	}

	return bars, nil
}
