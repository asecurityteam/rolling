package rolling

import (
	"fmt"
	"testing"
)

// https://gist.github.com/cevaris/bc331cbe970b03816c6b
var epsilon = 0.00000001

func floatEquals(a float64, b float64) bool {
	return (a-b) < epsilon && (b-a) < epsilon
}

var largeEpsilon = 0.001

func floatMostlyEquals(a float64, b float64) bool {
	return (a-b) < largeEpsilon && (b-a) < largeEpsilon
}

func TestPercentileAggregateInterpolateWhenInsufficientData(t *testing.T) {
	var numberOfPoints = 100
	var w = NewPointWindow(numberOfPoints)
	for x := 1; x <= numberOfPoints; x = x + 1 {
		w.Append(float64(x))
	}
	var perc = 99.9
	var a = Percentile(perc)
	var result = a(w)

	// When there are insufficient values to satisfy the precision then the
	// percentile algorithm degenerates to a max function. In this case, we need
	// 1000 values in order to select a 99.9 but only have 100. 100 is also the
	// maximum value and will be selected as k and k+1 in the linear
	// interpolation.
	var expected = 100.0
	if !floatEquals(result, expected) {
		t.Fatalf("%f percentile calculated incorrectly: %f versus %f", perc, expected, result)
	}
}

func TestPercentileAggregateInterpolateWhenSufficientData(t *testing.T) {
	var numberOfPoints = 1000
	var w = NewPointWindow(numberOfPoints)
	for x := 1; x <= numberOfPoints; x = x + 1 {
		w.Append(float64(x))
	}
	var perc = 99.9
	var a = Percentile(perc)
	var result = a(w)
	var expected = 999.5
	if !floatEquals(result, expected) {
		t.Fatalf("%f percentile calculated incorrectly: %f versus %f", perc, expected, result)
	}
}

var aggregateResult float64

type aggregateBench struct {
	inserts       int
	window        iteratable
	aggregate     func(iteratable) float64
	aggregateName string
}

func BenchmarkAggregates(b *testing.B) {
	var baseCases = []*aggregateBench{
		{aggregate: Sum, aggregateName: "sum"},
		{aggregate: Min, aggregateName: "min"},
		{aggregate: Max, aggregateName: "max"},
		{aggregate: Avg, aggregateName: "avg"},
		{aggregate: Count, aggregateName: "count"},
		{aggregate: Percentile(50.0), aggregateName: "p50"},
		{aggregate: Percentile(99.9), aggregateName: "p99.9"},
	}
	var insertions = []int{1, 1000, 10000, 100000}
	var benchCases = make([]*aggregateBench, 0, len(baseCases)*len(insertions))
	for _, baseCase := range baseCases {
		for _, inserts := range insertions {
			var w = NewPointWindow(inserts)
			for x := 1; x <= inserts; x = x + 1 {
				w.Append(float64(x))
			}
			benchCases = append(benchCases, &aggregateBench{
				inserts:       inserts,
				aggregate:     baseCase.aggregate,
				aggregateName: baseCase.aggregateName,
				window:        w,
			})
		}
	}

	for _, benchCase := range benchCases {
		b.Run(fmt.Sprintf("Aggregate:%s-DataPoints:%d", benchCase.aggregateName, benchCase.inserts), func(bt *testing.B) {
			var result float64
			bt.ResetTimer()
			for n := 0; n < bt.N; n = n + 1 {
				result = benchCase.aggregate(benchCase.window)
			}
			aggregateResult = result
		})
	}
}
