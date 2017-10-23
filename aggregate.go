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

type percentileAggregator struct {
	iterator   Iterator
	values     []float64
	percentile float64
	lock       *sync.Mutex
}

func (a *percentileAggregator) Aggregate() float64 {
	a.lock.Lock()
	defer a.lock.Unlock()
	a.values = a.values[:0]
	a.iterator.Iterate(func(p float64) {
		a.values = append(a.values, p)
	})
	if len(a.values) < 1 {
		return 0.0
	}
	if len(a.values) < 2 {
		return a.values[0]
	}
	sort.Float64s(a.values)
	var n = (a.percentile/100)*float64(len(a.values)) - 1
	var k = int(math.Floor(n))
	var plusOne = k + 1
	if plusOne > len(a.values)-1 {
		plusOne = k
	}
	var f = math.Mod(n, 1)
	return ((1 - f) * a.values[k]) + (f * a.values[plusOne])
}

type simpleAggregator struct {
	iterator Iterator
	f        func(Iterator) float64
}

func (a *simpleAggregator) Aggregate() float64 {
	return a.f(a.iterator)
}

// NewCountAggregator returns an Aggregator that computes the total number of
// elements in a window.
func NewCountAggregator(iterator Iterator) Aggregator {
	return &simpleAggregator{iterator: iterator, f: count}
}

// NewSumAggregator returns an Aggregator that computes the sum of all values
// in a window.
func NewSumAggregator(iterator Iterator) Aggregator {
	return &simpleAggregator{iterator: iterator, f: sum}
}

// NewMinAggregator returns an Aggregator that computes the min of all values
// in a window.
func NewMinAggregator(iterator Iterator) Aggregator {
	return &simpleAggregator{iterator: iterator, f: min}
}

// NewMaxAggregator returns an Aggregator that computes the max of all values
// in a window.
func NewMaxAggregator(iterator Iterator) Aggregator {
	return &simpleAggregator{iterator: iterator, f: max}
}

// NewAverageAggregator returns an Aggregator that computes the average of all values
// in a window.
func NewAverageAggregator(iterator Iterator) Aggregator {
	return &simpleAggregator{iterator: iterator, f: avg}
}

// NewPercentileAggregator returns an Aggregator that computes the given
// percentile of the values in a window. The given percentile is evaluated as
// N percentile such that the value 10.0 is considered to be 10.0 percentile.
// Non-whole numbers maybe be given to calculate, for example, the 99.9
// percentile. If the given percentile can be resolved exactly with the given
// data then the exact value is returned. If it cannot be resolved exactly, such
// as cases where there are not enough data to, then the result will be based on
// linear interpolation of the two closest points.
func NewPercentileAggregator(percentile float64, iterator Iterator, preallocHint int) Aggregator {
	return &percentileAggregator{
		iterator:   iterator,
		values:     make([]float64, preallocHint),
		percentile: percentile,
		lock:       &sync.Mutex{},
	}
}

type percentageAggregateor struct {
	aggregator    Aggregator
	lower         float64
	upper         float64
	adjustedUpper float64
}

func (e *percentageAggregateor) Aggregate() float64 {
	var value = e.aggregator.Aggregate() - e.lower
	if value <= 0 {
		return 0
	}
	return value / e.adjustedUpper
}

// NewPercentageAggregator creates an Aggregator that returns the percent between
// lower and upper of the aggregate value. If the aggregate is less than the
// lower then the result is 0.
func NewPercentageAggregator(aggregator Aggregator, lower float64, upper float64) Aggregator {
	return &percentageAggregateor{
		aggregator:    aggregator,
		lower:         lower,
		upper:         upper,
		adjustedUpper: upper - lower,
	}
}

type limitedAggregator struct {
	evaluator Aggregator
	counter   Aggregator
	limit     float64
}

func (e *limitedAggregator) Aggregate() float64 {
	if e.counter.Aggregate() < e.limit {
		return 0
	}
	return e.evaluator.Aggregate()
}

// NewLimitedAggregator creates an Aggregator that returns zero until the given
// iterator contains more than `limit` values are contained. Once the limit is
// passed, the given Aggregator is called to produce the output.
func NewLimitedAggregator(limit int, iterator Iterator, evaluator Aggregator) Aggregator {
	return &limitedAggregator{evaluator: evaluator, counter: NewCountAggregator(iterator), limit: float64(limit)}
}

type aggregatorIterator struct {
	aggregators []Aggregator
}

func (e *aggregatorIterator) Iterate(fn func(float64)) {
	for _, evaluator := range e.aggregators {
		fn(evaluator.Aggregate())
	}
}

// NewAggregatorIterator converts a set of aggregators into an iterator that
// can be fed back into an Aggregator for selecting on multiple windows.
func NewAggregatorIterator(aggregators ...Aggregator) Iterator {
	return &aggregatorIterator{aggregators}
}
