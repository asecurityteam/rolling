package rolling

import (
	"fmt"
	"testing"
	"time"
)

func TestCountAggregate(t *testing.T) {
	var numberOfPoints = 100
	var w = NewPointWindow(numberOfPoints)
	var a = NewCountRollup(w, "")
	var result = a.Aggregate()
	if result.Value != float64(numberOfPoints) {
		t.Fatalf("count calculated incorrectly: %f", result.Value)
	}
	w = NewTimeWindow(time.Millisecond, numberOfPoints, numberOfPoints)
	for x := 0; x < numberOfPoints; x = x + 1 {
		w.Feed(0)
	}
	result = a.Aggregate()
	if result.Value != float64(numberOfPoints) {
		t.Fatalf("count calculated incorrectly: %f", result.Value)
	}
}

func TestSumAggregate(t *testing.T) {
	var numberOfPoints = 100
	var w = NewPointWindow(numberOfPoints)
	for x := 0; x < numberOfPoints; x = x + 1 {
		w.Feed(1)
	}
	var a = NewSumRollup(w, "")
	var result = a.Aggregate()
	if result.Value != float64(numberOfPoints) {
		t.Fatalf("sum calculated incorrectly: %f", result.Value)
	}
}

func TestAverageAggregate(t *testing.T) {
	var numberOfPoints = 100
	var w = NewPointWindow(numberOfPoints)
	for x := 0; x < numberOfPoints; x = x + 1 {
		w.Feed(1)
	}
	var a = NewAverageRollup(w, "")
	var result = a.Aggregate()
	if result.Value != 1 {
		t.Fatalf("avg calculated incorrectly: %f", result.Value)
	}
}

func TestMinAggregate(t *testing.T) {
	var numberOfPoints = 100
	var w = NewPointWindow(numberOfPoints)
	for x := 0; x < numberOfPoints; x = x + 1 {
		w.Feed(float64(x))
	}
	var a = NewMinRollup(w, "")
	var result = a.Aggregate()
	if result.Value != 0 {
		t.Fatalf("min calculated incorrectly: %f", result.Value)
	}
}

func TestMaxAggregate(t *testing.T) {
	var numberOfPoints = 100
	var w = NewPointWindow(numberOfPoints)
	for x := 0; x < numberOfPoints; x = x + 1 {
		w.Feed(float64(x))
	}
	var a = NewMaxRollup(w, "")
	var result = a.Aggregate()
	if result.Value != float64(numberOfPoints-1) {
		t.Fatalf("max calculated incorrectly: %f", result.Value)
	}
}

func TestPercentileAggregatePrecise(t *testing.T) {
	var numberOfPoints = 100
	var w = NewPointWindow(numberOfPoints)
	for x := 1; x <= numberOfPoints; x = x + 1 {
		w.Feed(float64(x))
	}
	var perc = 20.0
	var a = NewPercentileRollup(perc, w, numberOfPoints, "")
	var result = a.Aggregate()
	if result.Value != perc {
		t.Fatalf("%f percentile calculated incorrectly: %f", perc, result.Value)
	}
	perc = 50.0
	a = NewPercentileRollup(perc, w, numberOfPoints, "")
	result = a.Aggregate()
	if result.Value != perc {
		t.Fatalf("%f percentile calculated incorrectly: %f", perc, result.Value)
	}
	perc = 99.0
	a = NewPercentileRollup(perc, w, numberOfPoints, "")
	result = a.Aggregate()
	if result.Value != perc {
		t.Fatalf("%f percentile calculated incorrectly: %f", perc, result.Value)
	}
}

func TestPercentileAggregateInterpolate(t *testing.T) {
	var numberOfPoints = 100
	var w = NewPointWindow(numberOfPoints)
	for x := 1; x <= numberOfPoints; x = x + 1 {
		w.Feed(float64(x))
	}
	var perc = 99.9
	var a = NewPercentileRollup(perc, w, numberOfPoints, "")
	var result = a.Aggregate()
	var expected = ((1 - .9) * 99) + (.9 * 100)
	if result.Value != expected {
		t.Fatalf("%f percentile calculated incorrectly: %f", perc, result.Value)
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
			var a = NewPercentileRollup(perc, w, prealloc, "")
			b.Run(fmt.Sprintf("PreAllocation:%d | DataPoints:%d", prealloc, inserts), func(bt *testing.B) {
				var result float64
				for n := 0; n < bt.N; n = n + 1 {
					result = a.Aggregate().Value
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
	var a = NewSumRollup(w, "")
	var e = NewPercentageRollup(a, 0, 100, "")
	var result = e.Aggregate()
	if result.Value != 1 {
		t.Fatalf("did not evaluate correct percentage: %f", result.Value)
	}
	e = NewPercentageRollup(a, 100, 1000, "")
	result = e.Aggregate()
	if result.Value != 0 {
		t.Fatalf("did not evaluate correct percentage: %f", result.Value)
	}

	e = NewPercentageRollup(a, 50, 150, "")
	result = e.Aggregate()
	if result.Value != .5 {
		t.Fatalf("did not evaluate correct percentage: %f", result.Value)
	}
}

func TestLimitedAggregator(t *testing.T) {
	var numberOfPoints = 100
	var w = NewTimeWindow(time.Millisecond, 100, numberOfPoints)
	var e = NewLimitedRollup(5, w, NewPercentageRollup(NewSumRollup(w, ""), 5, 15, ""))
	if v := e.Aggregate().Value; v != 0 {
		t.Fatal(v)
	}
	w.Feed(1)
	if v := e.Aggregate().Value; v != 0 {
		t.Fatal(v)
	}
	w.Feed(1)
	if v := e.Aggregate().Value; v != 0 {
		t.Fatal(v)
	}
	for x := 2; x < 10; x = x + 1 {
		w.Feed(1)
	}
	if v := e.Aggregate().Value; v != .5 {
		t.Fatal(v)
	}

}

func TestAggregateNames(t *testing.T) {
	var numberOfPoints = 100
	var w = NewPointWindow(numberOfPoints)
	for x := 0; x < numberOfPoints; x = x + 1 {
		w.Feed(1)
	}
	var a = NewSumRollup(w, "sum")
	var e = NewPercentageRollup(a, 0, 100, "percentage")
	var result = e.Aggregate()
	if result.Name != "percentage" {
		t.Fatalf("aggregate did not preserve name: %s", result.Name)
	}
	if result.Source == nil || result.Source.Name != "sum" {
		t.Fatalf("aggregate did not preserve previous aggregate name")
	}
}
