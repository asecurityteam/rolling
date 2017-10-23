package rolling

import (
	"sync"
	"time"
)

// TimeWindow is a rolling window implementation that uses some duration of
// time to determine the content of the window.
type timeWindow struct {
	prealloc          int
	bucketSize        time.Duration
	bucketSizeNano    int64
	numberOfBuckets   int
	numberOfBuckets64 int64
	window            [][]float64
	lastWindowOffset  int
	lastWindowTime    int64
	lock              *sync.Mutex
}

// NewTimeWindow generates a RollingWindow that operates on a rolling time duration.
// The given duration will be used to bucket data into segments within the window.
// If data points are received entire windows aparts then the window will only
// contain a single data point. If one or more durations of the window are
// missed then they are zeroed out to keep the window consistent.
func NewTimeWindow(bucketSize time.Duration, numberOfBuckets int, preallocHint int) Window {
	var w = &timeWindow{
		prealloc:          preallocHint,
		bucketSize:        bucketSize,
		bucketSizeNano:    bucketSize.Nanoseconds(),
		numberOfBuckets:   numberOfBuckets,
		numberOfBuckets64: int64(numberOfBuckets),
		window:            make([][]float64, numberOfBuckets),
		lock:              &sync.Mutex{},
	}
	for offset := range w.window {
		w.window[offset] = make([]float64, 0, w.prealloc)
	}
	return w
}

func (w *timeWindow) resetWindow() {
	for offset := range w.window {
		w.window[offset] = w.window[offset][:0]
	}
}

func (w *timeWindow) resetBuckets(windowOffset int) {
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

func (w *timeWindow) keepConsistent(adjustedTime int64, windowOffset int) {
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

func (w *timeWindow) selectBucket(currentTime time.Time) (int64, int) {
	var adjustedTime = currentTime.UnixNano() / w.bucketSizeNano
	var windowOffset = int(adjustedTime % w.numberOfBuckets64)
	return adjustedTime, windowOffset
}

func (w *timeWindow) Feed(value float64) {
	w.lock.Lock()
	defer w.lock.Unlock()
	var adjustedTime, windowOffset = w.selectBucket(time.Now())
	w.keepConsistent(adjustedTime, windowOffset)
	w.window[windowOffset] = append(w.window[windowOffset], value)
	w.lastWindowTime = adjustedTime
	w.lastWindowOffset = windowOffset
}

func (w *timeWindow) Iterate(f func(float64)) {
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

type pointWindow struct {
	windowSize       int
	window           [][]float64
	lastWindowOffset int
	lock             *sync.RWMutex
}

// NewPointWindow generates a RollingWindow that operates on a rolling set of
// input points. The given window size determines the total number of data points
// that are maintained within the window.
func NewPointWindow(windowSize int) Window {
	var w = &pointWindow{
		windowSize: windowSize,
		window:     make([][]float64, windowSize),
		lock:       &sync.RWMutex{},
	}
	for offset := range w.window {
		w.window[offset] = make([]float64, 1)
	}
	return w
}

func (w *pointWindow) Feed(value float64) {
	w.lock.Lock()
	defer w.lock.Unlock()
	var windowOffset = (w.lastWindowOffset + 1) % w.windowSize
	w.window[windowOffset][0] = value
	w.lastWindowOffset = windowOffset
}

func (w *pointWindow) Iterate(f func(float64)) {
	w.lock.RLock()
	defer w.lock.RUnlock()
	for _, bucket := range w.window {
		for _, point := range bucket {
			f(point)
		}
	}
}
