package rolling

// Aggregator is responsible for compacting a window of time into a single
// value for evaluation.
type Aggregator interface {
	Aggregate() float64
}

// Feeder populates a rolling window of data with input.
type Feeder interface {
	Feed(value float64)
}

// Iterator takes a function and calls it for every point in a window.
type Iterator interface {
	Iterate(func(float64))
}

// Window is a composit of Iterator and Feeder types.
type Window interface {
	Iterator
	Feeder
}
