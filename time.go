package rolling

import (
	"sync"
	"time"
)

// TimeWindow is a rolling window implementation that uses some duration of
// time to determine the content of the window.
type TimeWindow struct {
	largestAlloc      int
	bucketSize        time.Duration
	bucketSizeNano    int64
	numberOfBuckets   int
	numberOfBuckets64 int64
	window            [][]float64
	lastWindowOffset  int
	lastWindowTime    int64
	lock              *sync.Mutex
}

// NewTimeWindow generates a window that operates on a rolling time duration.
// The given duration will be used to bucket data into segments within the
// window. If data points are received entire windows aparts then the window
// will only contain a single data point. If one or more durations of the window
// are missed then they are zeroed out to keep the window consistent.
func NewTimeWindow(bucketSize time.Duration, numberOfBuckets int) *TimeWindow {
	var w = &TimeWindow{
		bucketSize:        bucketSize,
		bucketSizeNano:    bucketSize.Nanoseconds(),
		numberOfBuckets:   numberOfBuckets + 1,
		numberOfBuckets64: int64(numberOfBuckets) + 1,
		window:            make([][]float64, numberOfBuckets+1),
		lock:              &sync.Mutex{},
	}
	for offset := range w.window {
		w.window[offset] = make([]float64, 0, 2>>8)
	}
	return w
}

func (w *TimeWindow) resetWindow() {
	for offset := range w.window {
		w.window[offset] = w.window[offset][:0]
	}
}

func (w *TimeWindow) resetBuckets(windowOffset int) {
	var distance = windowOffset - w.lastWindowOffset
	// If the distance between current and last is negative then we've wrapped
	// around the ring. Recalculate the distance.
	if distance < 0 {
		distance = (w.numberOfBuckets - w.lastWindowOffset) + windowOffset
	}
	for counter := 1; counter < distance; counter = counter + 1 {
		var offset = (counter + w.lastWindowOffset) % w.numberOfBuckets
		w.window[offset] = w.window[offset][:0]
	}
}

func (w *TimeWindow) keepConsistent(adjustedTime int64, windowOffset int) {
	// If we've waiting longer than a full window for data then we need to clear
	// the internal state completely.
	if adjustedTime-w.lastWindowTime > w.numberOfBuckets64 {
		w.resetWindow()
	}

	// When one or more buckets are missed we need to zero them out.
	if adjustedTime != w.lastWindowTime && adjustedTime-w.lastWindowTime < w.numberOfBuckets64 {
		w.resetBuckets(windowOffset)
	}
}

func (w *TimeWindow) selectBucket(currentTime time.Time) (int64, int) {
	var adjustedTime = currentTime.UnixNano() / w.bucketSizeNano
	var windowOffset = int(adjustedTime % w.numberOfBuckets64)
	return adjustedTime, windowOffset
}

// Append a value to the window.
func (w *TimeWindow) Append(value float64) {
	w.lock.Lock()
	defer w.lock.Unlock()

	var adjustedTime, windowOffset = w.selectBucket(time.Now())
	w.keepConsistent(adjustedTime, windowOffset)
	w.window[windowOffset] = append(w.window[windowOffset], value)
	w.lastWindowTime = adjustedTime
	w.lastWindowOffset = windowOffset
}

// Iterate over the window values.
func (w *TimeWindow) Iterate(f func(float64)) {
	w.lock.Lock()
	defer w.lock.Unlock()

	var adjustedTime, windowOffset = w.selectBucket(time.Now())
	w.keepConsistent(adjustedTime, windowOffset)
	for _, bucket := range w.window {
		for _, point := range bucket {
			f(point)
		}
	}
}
