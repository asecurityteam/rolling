package rolling

import (
	"math"
	"sort"
	"sync"
)

type iteratable interface {
	Iterate(func(float64))
}

// Count returns the number of elements in a window.
func Count(it iteratable) float64 {
	var result = 0.0
	it.Iterate(func(v float64) {
		result = result + 1
	})
	return result
}

// Sum the values within the window.
func Sum(it iteratable) float64 {
	var result = 0.0
	it.Iterate(func(v float64) {
		result = result + v
	})
	return result
}

// Avg the values within the window.
func Avg(it iteratable) float64 {
	var result = 0.0
	var count = 0.0
	it.Iterate(func(v float64) {
		result = result + v
		count = count + 1
	})
	return result / count
}

// Min the values within the window.
func Min(it iteratable) float64 {
	var result = 0.0
	var started = false
	it.Iterate(func(v float64) {
		if !started {
			result = v
			return
		}
		if v < result {
			result = v
		}
	})
	return result
}

// Max the values within the window.
func Max(it iteratable) float64 {
	var result = 0.0
	var started = false
	it.Iterate(func(v float64) {
		if !started {
			result = v
			return
		}
		if v > result {
			result = v
		}
	})
	return result
}

// Percentile returns an aggregating function that computes the
// given percentile calculation for a window.
func Percentile(p float64) func(it iteratable) float64 {
	var values []float64
	var lock = &sync.Mutex{}
	return func(it iteratable) float64 {
		lock.Lock()
		defer lock.Unlock()

		values = values[:0]
		it.Iterate(func(v float64) {
			values = append(values, v)
		})
		sort.Float64s(values)
		var position = (float64(len(values))*(p/100) + .5) - 1
		var k = int(math.Floor(position))
		var f = math.Mod(position, 1)
		if f == 0.0 {
			return values[k]
		}
		var plusOne = k + 1
		if plusOne > len(values)-1 {
			plusOne = k
		}
		return ((1 - f) * values[k]) + (f * values[plusOne])
	}
}
