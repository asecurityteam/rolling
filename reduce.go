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

// FastPercentile implements the pSquare percentile estimation
// algorithm for calculating percentiles from streams of data
// using fixed memory allocations.
func FastPercentile(p float64) func(it iteratable) float64 {
	p = p / 100.0
	return func(it iteratable) float64 {
		var initalObservations = make([]float64, 0, 5)
		var q [5]float64
		var n [5]int
		var nPrime [5]float64
		var dnPrime [5]float64
		var observations uint64

		it.Iterate(func(v float64) {
			observations = observations + 1
			// Record first five observations
			if observations < 6 {
				initalObservations = append(initalObservations, v)
				return
			}
			// Before proceeding beyond the first five, process them.
			if observations == 6 {
				bubbleSort(initalObservations)
				for offset := range q {
					q[offset] = initalObservations[offset]
					n[offset] = offset
				}
				nPrime[0] = 0
				nPrime[1] = 2 * p
				nPrime[2] = 4 * p
				nPrime[3] = 2 + 2*p
				nPrime[4] = 4
				dnPrime[0] = 0
				dnPrime[1] = p / 2
				dnPrime[2] = p
				dnPrime[3] = (1 + p) / 2
				dnPrime[4] = 1
			}
			var k int // k is the target cell to increment
			switch {
			case v < q[0]:
				q[0] = v
				k = 0
			case q[0] <= v && v < q[1]:
				k = 0
			case q[1] <= v && v < q[2]:
				k = 1
			case q[2] <= v && v < q[3]:
				k = 2
			case q[3] <= v && v <= q[4]:
				k = 3
			case v > q[4]:
				q[4] = v
				k = 3
			}
			for x := k + 1; x < 5; x = x + 1 {
				n[x] = n[x] + 1
			}
			nPrime[0] = nPrime[0] + dnPrime[0]
			nPrime[1] = nPrime[1] + dnPrime[1]
			nPrime[2] = nPrime[2] + dnPrime[2]
			nPrime[3] = nPrime[3] + dnPrime[3]
			nPrime[4] = nPrime[4] + dnPrime[4]
			for x := 1; x < 4; x = x + 1 {
				var d = nPrime[x] - float64(n[x])
				if (d >= 1 && (n[x+1]-n[x]) > 1) ||
					(d <= -1 && (n[x-1]-n[x]) < -1) {
					var s = sign(d)
					var si = int(s)
					var nx = float64(n[x])
					var nxPlusOne = float64(n[x+1])
					var nxMinusOne = float64(n[x-1])
					var qx = q[x]
					var qxPlusOne = q[x+1]
					var qxMinusOne = q[x-1]
					var parab = q[x] + (s/(nxPlusOne-nxMinusOne))*((nx-nxMinusOne+s)*(qxPlusOne-qx)/(nxPlusOne-nx)+(nxPlusOne-nx-s)*(qx-qxMinusOne)/(nx-nxMinusOne))
					if qxMinusOne < parab && parab < qxPlusOne {
						q[x] = parab
					} else {
						q[x] = q[x] + s*((q[x+si]-q[x])/float64(n[x+si]-n[x]))
					}
					n[x] = n[x] + si
				}
			}
		})
		// If we have less than five values then degenerate into a max function.
		// This is a reasonable value for data sets this small.
		if observations < 5 {
			bubbleSort(initalObservations)
			return initalObservations[len(initalObservations)-1]
		}
		return q[2]
	}
}

func sign(v float64) float64 {
	if v < 0 {
		return -1
	}
	return 1
}

// using bubblesort because we're only working with datasets of 5 or fewer
// elements.
func bubbleSort(s []float64) {
	for range s {
		for x := 0; x < len(s)-1; x = x + 1 {
			if s[x] > s[x+1] {
				s[x], s[x+1] = s[x+1], s[x]
			}
		}
	}
}
