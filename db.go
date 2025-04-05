package main

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"image/color"
	"sync"
	"time"

	"github.com/akrylysov/pogreb"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text"
	"golang.org/x/image/font"
	"golang.org/x/image/font/gofont/goregular"
	"golang.org/x/image/font/opentype"
)

type Database struct {
	db             *pogreb.DB
	errorMsg       string // Persistent error message
	fetchStatus    string // Status of ongoing fetch operations
	fetching       bool   // Indicates if fetching is in progress
	fetchMutex     sync.Mutex
	fontFace       font.Face
	fetchStart     time.Time
	totalMinutes   int64 // Total minutes to fetch
	fetchedMinutes int64 // Minutes already fetched
}

func loadFont() font.Face {
	tt, err := opentype.Parse(goregular.TTF)
	if err != nil {
		panic(fmt.Errorf("failed to parse font: %v", err))
	}
	face, err := opentype.NewFace(tt, &opentype.FaceOptions{
		Size:    12,
		DPI:     72,
		Hinting: font.HintingFull,
	})
	if err != nil {
		panic(fmt.Errorf("failed to create font face: %v", err))
	}
	return face
}

func NewDatabase() (*Database, error) {
	db, err := pogreb.Open("BTCUSDT.db", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %v", err)
	}

	d := &Database{
		db:       db,
		fontFace: loadFont(),
	}

	// Start fetching data asynchronously
	go d.ensureLastData()

	return d, nil
}

func (d *Database) Close() error {
	return d.db.Close()
}

func (d *Database) DrawError(screen *ebiten.Image) {
	if d.errorMsg == "" {
		return
	}
	text.Draw(screen, d.errorMsg, d.fontFace, 10, 20, color.RGBA{255, 0, 0, 255})
}

func (d *Database) IsEmpty() (bool, error) {
	it := d.db.Items()
	_, _, err := it.Next()
	if err == pogreb.ErrIterationDone {
		return true, nil
	}
	return false, err
}

func (d *Database) getLatestTimestamp() (int64, error) {
	latestKey := []byte("latest_timestamp")
	value, err := d.db.Get(latestKey)
	if err != nil {
		return 0, fmt.Errorf("failed to get latest timestamp: %v", err)
	}
	if value == nil {
		return 0, fmt.Errorf("no latest timestamp found in database")
	}
	return bytesToInt64(value), nil
}

func (d *Database) setLatestTimestamp(timestamp int64) error {
	latestKey := []byte("latest_timestamp")
	value := int64ToBytes(timestamp)
	return d.db.Put(latestKey, value)
}

func bytesToInt64(b []byte) int64 {
	return int64(binary.BigEndian.Uint64(b))
}

func (d *Database) ensureLastData() error {
	d.fetchMutex.Lock()
	d.fetching = true
	d.fetchStart = time.Now()
	d.fetchStatus = "Fetching data via API: https://api.binance.com/api/v3/klines..."
	d.fetchMutex.Unlock()

	defer func() {
		d.fetchMutex.Lock()
		d.fetching = false
		d.fetchStatus = ""
		d.fetchMutex.Unlock()
	}()

	// First check if DB is completely empty
	empty, err := d.IsEmpty()
	if err != nil {
		d.setError(fmt.Errorf("failed to check if DB is empty: %v", err))
		return err
	}

	now := time.Now().UTC()
	endTime := now.Unix() * 1000
	var startTime int64

	if empty {
		// If DB is empty, fetch last 7 days
		startTime = now.Add(-7*24*time.Hour).Unix() * 1000
	} else {
		// Otherwise find the latest timestamp in DB
		latest, err := d.getLatestTimestamp()
		if err != nil {
			d.setError(fmt.Errorf("failed to get latest timestamp: %v", err))
			return err
		}
		startTime = latest + 60*1000 // Start from next minute
	}

	// Check if we need to fetch any data
	if startTime > endTime {
		return nil
	}

	// Add delay between requests
	time.Sleep(3 * time.Second)

	// Calculate how many minutes we need to fetch
	d.totalMinutes = (endTime - startTime) / (60 * 1000)
	d.fetchedMinutes = 0
	const maxLimit = 1000
	var allData []OHLCV

	currentEndTime := endTime
	for d.totalMinutes > 0 {
		fetchMinutes := int64(min(int(d.totalMinutes), maxLimit))
		data, err := Fetch(fetchMinutes, currentEndTime)
		if err != nil {
			d.setError(fmt.Errorf("failed to fetch data ending at %d: %v", currentEndTime, err))
			return err
		}

		if len(data) > 0 {
			chunkStartTime := data[0].Time
			chunkEndTime := data[len(data)-1].Time
			if err := d.checkContinuity(data, chunkStartTime, chunkEndTime); err != nil {
				d.setError(fmt.Errorf("continuity check failed for chunk %d to %d: %v", chunkStartTime, chunkEndTime, err))
				return err
			}
			allData = append(data, allData...)
		}

		// Update fetch status
		d.fetchMutex.Lock()
		d.fetchedMinutes += fetchMinutes
		remainingMinutes := d.totalMinutes - d.fetchedMinutes
		elapsed := time.Since(d.fetchStart).Seconds()
		avgTimePerMinute := elapsed / float64(d.fetchedMinutes)
		remainingTime := int64(avgTimePerMinute * float64(remainingMinutes))
		d.fetchStatus = fmt.Sprintf("Fetching data via API: https://api.binance.com/api/v3/klines... %d minutes remaining (~%ds)", remainingMinutes, remainingTime)
		d.fetchMutex.Unlock()

		// Update for the next chunk
		d.totalMinutes -= fetchMinutes
		if len(data) > 0 {
			currentEndTime = data[0].Time - 60*1000
		} else {
			currentEndTime -= fetchMinutes * 60 * 1000
		}
	}

	// Store data in the database, excluding current incomplete minute
	currentMinute := time.Now().UTC().Unix() / 60 * 60 * 1000
	var latestTimestamp int64
	for _, ohlcv := range allData {
		if ohlcv.Time >= currentMinute {
			continue // Skip current incomplete minute
		}
		key := int64ToBytes(ohlcv.Time)
		value, err := serializeOHLCV(ohlcv)
		if err != nil {
			d.setError(fmt.Errorf("failed to serialize OHLCV: %v", err))
			return err
		}
		if err := d.db.Put(key, value); err != nil {
			d.setError(fmt.Errorf("failed to store data: %v", err))
			return err
		}
		if ohlcv.Time > latestTimestamp {
			latestTimestamp = ohlcv.Time
		}
	}

	// Update latest timestamp
	if latestTimestamp != 0 {
		if err := d.setLatestTimestamp(latestTimestamp); err != nil {
			d.setError(fmt.Errorf("failed to update latest timestamp: %v", err))
			return err
		}
	}

	if err := d.db.Sync(); err != nil {
		d.setError(fmt.Errorf("failed to sync database: %v", err))
		return err
	}

	return nil
}

func (d *Database) checkContinuity(data []OHLCV, startTime, endTime int64) error {
	if len(data) == 0 {
		return fmt.Errorf("no data received")
	}

	expectedTime := startTime
	for _, ohlcv := range data {
		if ohlcv.Time != expectedTime {
			return fmt.Errorf("missing data at %s (expected %d, got %d)",
				time.Unix(expectedTime/1000, 0).UTC().Format(time.RFC3339),
				expectedTime, ohlcv.Time)
		}
		expectedTime += 60 * 1000
	}

	return nil
}

func (d *Database) setError(err error) {
	d.fetchMutex.Lock()
	d.errorMsg = fmt.Sprintf("ERROR: %v", err)
	fmt.Println(d.errorMsg)
	d.fetchMutex.Unlock()
}

func int64ToBytes(i int64) []byte {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, uint64(i))
	return b
}

func serializeOHLCV(o OHLCV) ([]byte, error) {
	data, err := json.Marshal(o)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func deserializeOHLCV(data []byte) (OHLCV, error) {
	var ohlcv OHLCV
	err := json.Unmarshal(data, &ohlcv)
	return ohlcv, err
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
