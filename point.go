package rolling

import "sync"

// PointWindow is a rolling window implementation that tracks the last N
// values inserted regardless of insertion time.
type PointWindow struct {
	windowSize int
	window     []float64
	offset     int
	lock       *sync.RWMutex
}

// NewPointWindow generates a RollingWindow that operates on a rolling set of
// input points. The given window size determines the total number of data points
// that are maintained within the window.
func NewPointWindow(windowSize int) *PointWindow {
	return &PointWindow{
		windowSize: windowSize,
		window:     make([]float64, windowSize),
		lock:       &sync.RWMutex{},
	}
}

// Append a value to the window.
func (w *PointWindow) Append(value float64) {
	w.lock.Lock()
	defer w.lock.Unlock()

	w.window[w.offset] = value
	w.offset = (w.offset + 1) % w.windowSize
}

// Iterate over the window values.
func (w *PointWindow) Iterate(f func(float64)) {
	w.lock.RLock()
	defer w.lock.RUnlock()

	for _, point := range w.window {
		f(point)
	}
}
