package rolling

import (
	"math"
	"sort"
	"sync"
)

func count(it Iterator) float64 {
	var result float64
	it.Iterate(func(p float64) {
		result = result + 1
	})
	return result
}

func sum(it Iterator) float64 {
	var result float64
	it.Iterate(func(p float64) {
		result = result + p
	})
	return result
}

func avg(it Iterator) float64 {
	var result float64
	var numberOfPoints float64
	it.Iterate(func(p float64) {
		result = result + p
		numberOfPoints = numberOfPoints + 1
	})
	return result / numberOfPoints
}

func min(it Iterator) float64 {
	var result float64
	var gotOne bool
	it.Iterate(func(p float64) {
		if !gotOne {
			result = p
			gotOne = true
			return
		}
		result = math.Min(result, p)
	})
	return result
}

func max(it Iterator) float64 {
	var result float64
	var gotOne bool
	it.Iterate(func(p float64) {
		if !gotOne {
			result = p
			gotOne = true
			return
		}
		result = math.Max(result, p)
	})
	return result
}

type percentileRollup struct {
	iterator   Iterator
	values     []float64
	percentile float64
	lock       *sync.Mutex
	name       string
}

func (a *percentileRollup) Name() string {
	return a.name
}

func (a *percentileRollup) Aggregate() *Aggregate {
	a.lock.Lock()
	defer a.lock.Unlock()
	a.values = a.values[:0]
	a.iterator.Iterate(func(p float64) {
		a.values = append(a.values, p)
	})
	if len(a.values) < 1 {
		return &Aggregate{
			Source: nil,
			Name:   a.Name(),
			Value:  0.0,
		}
	}
	if len(a.values) < 2 {
		return &Aggregate{
			Source: nil,
			Name:   a.Name(),
			Value:  a.values[0],
		}
	}
	sort.Float64s(a.values)
	var n = (a.percentile/100)*float64(len(a.values)) - 1
	var k = int(math.Floor(n))
	var plusOne = k + 1
	if plusOne > len(a.values)-1 {
		plusOne = k
	}
	var f = math.Mod(n, 1)
	return &Aggregate{
		Source: nil,
		Name:   a.Name(),
		Value:  ((1 - f) * a.values[k]) + (f * a.values[plusOne]),
	}
}

type simpleRollup struct {
	iterator Iterator
	f        func(Iterator) float64
	name     string
}

func (a *simpleRollup) Aggregate() *Aggregate {
	return &Aggregate{
		Source: nil,
		Name:   a.Name(),
		Value:  a.f(a.iterator),
	}
}

func (a *simpleRollup) Name() string {
	return a.name
}

// NewCountRollup returns an Aggregator that computes the total number of
// elements in a window.
func NewCountRollup(iterator Iterator, name string) Rollup {
	return &simpleRollup{iterator: iterator, f: count, name: name}
}

// NewSumRollup returns an Aggregator that computes the sum of all values
// in a window.
func NewSumRollup(iterator Iterator, name string) Rollup {
	return &simpleRollup{iterator: iterator, f: sum, name: name}
}

// NewMinRollup returns an Aggregator that computes the min of all values
// in a window.
func NewMinRollup(iterator Iterator, name string) Rollup {
	return &simpleRollup{iterator: iterator, f: min, name: name}
}

// NewMaxRollup returns an Aggregator that computes the max of all values
// in a window.
func NewMaxRollup(iterator Iterator, name string) Rollup {
	return &simpleRollup{iterator: iterator, f: max, name: name}
}

// NewAverageRollup returns an Aggregator that computes the average of all values
// in a window.
func NewAverageRollup(iterator Iterator, name string) Rollup {
	return &simpleRollup{iterator: iterator, f: avg, name: name}
}

// NewPercentileRollup returns an Aggregator that computes the given
// percentile of the values in a window. The given percentile is evaluated as
// N percentile such that the value 10.0 is considered to be 10.0 percentile.
// Non-whole numbers maybe be given to calculate, for example, the 99.9
// percentile. If the given percentile can be resolved exactly with the given
// data then the exact value is returned. If it cannot be resolved exactly, such
// as cases where there are not enough data to, then the result will be based on
// linear interpolation of the two closest points.
func NewPercentileRollup(percentile float64, iterator Iterator, preallocHint int, name string) Rollup {
	return &percentileRollup{
		iterator:   iterator,
		values:     make([]float64, preallocHint),
		percentile: percentile,
		lock:       &sync.Mutex{},
		name:       name,
	}
}

type percentageRollup struct {
	aggregator    Aggregator
	lower         float64
	upper         float64
	adjustedUpper float64
	name          string
}

func (e *percentageRollup) Name() string {
	return e.name
}

func (e *percentageRollup) Aggregate() *Aggregate {
	var p = e.aggregator.Aggregate()
	var value = p.Value - e.lower
	if value <= 0 {
		return &Aggregate{
			Source: p,
			Name:   e.Name(),
			Value:  0,
		}
	}
	return &Aggregate{
		Source: p,
		Name:   e.Name(),
		Value:  value / e.adjustedUpper,
	}
}

// NewPercentageRollup creates an Aggregator that returns the percent between
// lower and upper of the aggregate value. If the aggregate is less than the
// lower then the result is 0.
func NewPercentageRollup(aggregator Aggregator, lower float64, upper float64, name string) Rollup {
	return &percentageRollup{
		aggregator:    aggregator,
		lower:         lower,
		upper:         upper,
		adjustedUpper: upper - lower,
		name:          name,
	}
}

type limitedRollup struct {
	rollup  Rollup
	counter Aggregator
	limit   float64
}

func (e *limitedRollup) Name() string {
	return e.rollup.Name()
}

func (e *limitedRollup) Aggregate() *Aggregate {
	if e.counter.Aggregate().Value < e.limit {
		return &Aggregate{
			Source: nil,
			Name:   e.Name(),
			Value:  0,
		}
	}
	return e.rollup.Aggregate()
}

// NewLimitedRollup creates an Aggregator that returns zero until the given
// iterator contains more than `limit` values are contained. Once the limit is
// passed, the given Aggregator is called to produce the output.
func NewLimitedRollup(limit int, iterator Iterator, rollup Rollup) Rollup {
	return &limitedRollup{rollup: rollup, counter: NewCountRollup(iterator, ""), limit: float64(limit)}
}
