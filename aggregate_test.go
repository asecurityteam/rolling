package rolling

import (
	"fmt"
	"testing"
	"time"
)

func TestCountAggregate(t *testing.T) {
	var numberOfPoints = 100
	var w = NewPointWindow(numberOfPoints)
	var a = NewCountAggregator(w)
	var result = a.Aggregate()
	if result != float64(numberOfPoints) {
		t.Fatalf("count calculated incorrectly: %f", result)
	}
	w = NewTimeWindow(time.Millisecond, numberOfPoints, numberOfPoints)
	for x := 0; x < numberOfPoints; x = x + 1 {
		w.Feed(0)
	}
	result = a.Aggregate()
	if result != float64(numberOfPoints) {
		t.Fatalf("count calculated incorrectly: %f", result)
	}
}

func TestSumAggregate(t *testing.T) {
	var numberOfPoints = 100
	var w = NewPointWindow(numberOfPoints)
	for x := 0; x < numberOfPoints; x = x + 1 {
		w.Feed(1)
	}
	var a = NewSumAggregator(w)
	var result = a.Aggregate()
	if result != float64(numberOfPoints) {
		t.Fatalf("sum calculated incorrectly: %f", result)
	}
}

func TestAverageAggregate(t *testing.T) {
	var numberOfPoints = 100
	var w = NewPointWindow(numberOfPoints)
	for x := 0; x < numberOfPoints; x = x + 1 {
		w.Feed(1)
	}
	var a = NewAverageAggregator(w)
	var result = a.Aggregate()
	if result != 1 {
		t.Fatalf("avg calculated incorrectly: %f", result)
	}
}

func TestMinAggregate(t *testing.T) {
	var numberOfPoints = 100
	var w = NewPointWindow(numberOfPoints)
	for x := 0; x < numberOfPoints; x = x + 1 {
		w.Feed(float64(x))
	}
	var a = NewMinAggregator(w)
	var result = a.Aggregate()
	if result != 0 {
		t.Fatalf("min calculated incorrectly: %f", result)
	}
}

func TestMaxAggregate(t *testing.T) {
	var numberOfPoints = 100
	var w = NewPointWindow(numberOfPoints)
	for x := 0; x < numberOfPoints; x = x + 1 {
		w.Feed(float64(x))
	}
	var a = NewMaxAggregator(w)
	var result = a.Aggregate()
	if result != float64(numberOfPoints-1) {
		t.Fatalf("max calculated incorrectly: %f", result)
	}
}

func TestPercentileAggregatePrecise(t *testing.T) {
	var numberOfPoints = 100
	var w = NewPointWindow(numberOfPoints)
	for x := 1; x <= numberOfPoints; x = x + 1 {
		w.Feed(float64(x))
	}
	var perc = 20.0
	var a = NewPercentileAggregator(perc, w, numberOfPoints)
	var result = a.Aggregate()
	if result != perc {
		t.Fatalf("%f percentile calculated incorrectly: %f", perc, result)
	}
	perc = 50.0
	a = NewPercentileAggregator(perc, w, numberOfPoints)
	result = a.Aggregate()
	if result != perc {
		t.Fatalf("%f percentile calculated incorrectly: %f", perc, result)
	}
	perc = 99.0
	a = NewPercentileAggregator(perc, w, numberOfPoints)
	result = a.Aggregate()
	if result != perc {
		t.Fatalf("%f percentile calculated incorrectly: %f", perc, result)
	}
}

func TestPercentileAggregateInterpolate(t *testing.T) {
	var numberOfPoints = 100
	var w = NewPointWindow(numberOfPoints)
	for x := 1; x <= numberOfPoints; x = x + 1 {
		w.Feed(float64(x))
	}
	var perc = 99.9
	var a = NewPercentileAggregator(perc, w, numberOfPoints)
	var result = a.Aggregate()
	var expected = ((1 - .9) * 99) + (.9 * 100)
	if result != expected {
		t.Fatalf("%f percentile calculated incorrectly: %f", perc, result)
	}
}

var percentileResult float64

func BenchmarkPercentileAggregate(b *testing.B) {
	var perc = 99.9
	var preallocations = []int{0, 1000, 10000}
	var insertions = []int{1, 1000, 10000}
	for _, prealloc := range preallocations {
		for _, inserts := range insertions {
			var w = NewPointWindow(inserts)
			for x := 1; x <= inserts; x = x + 1 {
				w.Feed(float64(x))
			}
			var a = NewPercentileAggregator(perc, w, prealloc)
			b.Run(fmt.Sprintf("PreAllocation:%d | DataPoints:%d", prealloc, inserts), func(bt *testing.B) {
				var result float64
				for n := 0; n < bt.N; n = n + 1 {
					result = a.Aggregate()
				}
				percentileResult = result
			})
		}
	}
}

func TestPercentageAggregator(t *testing.T) {
	var numberOfPoints = 100
	var w = NewPointWindow(numberOfPoints)
	for x := 0; x < numberOfPoints; x = x + 1 {
		w.Feed(1)
	}
	var a = NewSumAggregator(w)
	var e = NewPercentageAggregator(a, 0, 100)
	var result = e.Aggregate()
	if result != 1 {
		t.Fatalf("did not evaluate correct percentage: %f", result)
	}
	e = NewPercentageAggregator(a, 100, 1000)
	result = e.Aggregate()
	if result != 0 {
		t.Fatalf("did not evaluate correct percentage: %f", result)
	}

	e = NewPercentageAggregator(a, 50, 150)
	result = e.Aggregate()
	if result != .5 {
		t.Fatalf("did not evaluate correct percentage: %f", result)
	}
}

func TestLimitedAggregator(t *testing.T) {
	var numberOfPoints = 100
	var w = NewTimeWindow(time.Millisecond, 100, numberOfPoints)
	var e = NewLimitedAggregator(5, w, NewPercentageAggregator(NewSumAggregator(w), 5, 15))
	if v := e.Aggregate(); v != 0 {
		t.Fatal(v)
	}
	w.Feed(1)
	if v := e.Aggregate(); v != 0 {
		t.Fatal(v)
	}
	w.Feed(1)
	if v := e.Aggregate(); v != 0 {
		t.Fatal(v)
	}
	for x := 2; x < 10; x = x + 1 {
		w.Feed(1)
	}
	if v := e.Aggregate(); v != .5 {
		t.Fatal(v)
	}

}
