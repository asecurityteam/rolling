package rolling

import (
	"fmt"
	"math"
	"testing"
	"time"
)

func requiresIterator(it Iterator)   {}
func requiresFeeder(f Feeder)        {}
func requiresRollingWindow(r Window) {}

type embeddedFeeder interface {
	Feed(value float64)
}

func requiresEmbeddedFeeder(f embeddedFeeder) {}

func TestTimeWindowInterfaces(t *testing.T) {
	var w = NewTimeWindow(time.Millisecond, 100, 1000)
	requiresIterator(w)
	requiresFeeder(w)
	requiresRollingWindow(w)
	requiresEmbeddedFeeder(w)
}

func TestTimeWindow(t *testing.T) {
	var bucketSize = time.Millisecond * 50
	var numberBuckets = 10
	var hint = 1000
	var w = NewTimeWindow(bucketSize, numberBuckets, hint)
	for x := 0; x < numberBuckets; x = x + 1 {
		w.Feed(1)
		time.Sleep(bucketSize)
	}
	var final float64
	w.Iterate(func(p float64) {
		final = final + p
	})
	if final != float64(numberBuckets) {
		t.Fatal(final)
	}
}

func TestTimeWindowSelectBucket(t *testing.T) {
	var bucketSize = time.Millisecond * 50
	var numberBuckets = 10
	var hint = 1000
	var w = NewTimeWindow(bucketSize, numberBuckets, hint).(*timeWindow)
	var target = time.Unix(1, 0)
	var _, bucket = w.selectBucket(target)
	if bucket != 0 {
		t.Fatalf("expected bucket 0 but got %d %v", bucket, target)
	}
	target = time.Unix(1, int64(50*time.Millisecond))
	_, bucket = w.selectBucket(target)
	if bucket != 1 {
		t.Fatalf("expected bucket 1 but got %d %v", bucket, target)
	}
}

func TestTimeWindowConsistency(t *testing.T) {
	var bucketSize = time.Millisecond * 50
	var numberBuckets = 10
	var hint = 1000
	var w = NewTimeWindow(bucketSize, numberBuckets, hint).(*timeWindow)
	for offset := range w.window {
		w.window[offset] = append(w.window[offset], 1)
	}
	w.lastWindowTime = time.Now().UnixNano()
	w.lastWindowOffset = 0
	var target = time.Unix(1, 0)
	var adjustedTime, bucket = w.selectBucket(target)
	w.keepConsistent(adjustedTime, bucket)
	if len(w.window[0]) != 1 {
		t.Fatal("data loss while adjusting internal state")
	}
	target = time.Unix(1, int64(50*time.Millisecond))
	adjustedTime, bucket = w.selectBucket(target)
	w.keepConsistent(adjustedTime, bucket)
	if len(w.window[0]) != 1 {
		t.Fatal("data loss while adjusting internal state")
	}
	target = time.Unix(1, int64(5*50*time.Millisecond))
	adjustedTime, bucket = w.selectBucket(target)
	w.keepConsistent(adjustedTime, bucket)
	if len(w.window[0]) != 1 {
		t.Fatal("data loss while adjusting internal state")
	}
	for x := 1; x < 5; x = x + 1 {
		if len(w.window[x]) != 0 {
			t.Fatal("internal state not kept consistent during time gap")
		}
	}
}

func TestTimeWindowDataRace(t *testing.T) {
	var bucketSize = time.Millisecond
	var numberBuckets = 1000
	var hint = 1000
	var w = NewTimeWindow(bucketSize, numberBuckets, hint)
	var stop = make(chan bool)
	go func() {
		for {
			select {
			case <-stop:
				return
			default:
				w.Feed(1)
				time.Sleep(time.Millisecond)
			}
		}
	}()
	go func() {
		var v float64
		for {
			select {
			case <-stop:
				return
			default:
				w.Iterate(func(p float64) {
					v = v + p
					v = math.Mod(v, float64(numberBuckets))
				})
			}
		}
	}()
	time.Sleep(time.Second)
	close(stop)
}

type timeWindowOptions struct {
	n string
	d time.Duration
	s int
	a int
	i int
}

func BenchmarkTimeWindow(b *testing.B) {
	var durations = []time.Duration{time.Millisecond}
	var bucketSizes = []int{1, 10, 100, 1000}
	var preallocations = []int{0, 1000, 10000}
	var insertions = []int{1, 1000, 10000}
	var options = make([]timeWindowOptions, 0, len(durations)*len(bucketSizes)*len(preallocations)*len(insertions))
	for _, d := range durations {
		for _, s := range bucketSizes {
			for _, a := range preallocations {
				for _, i := range insertions {
					options = append(options, timeWindowOptions{d: d, s: s, a: a, i: i, n: fmt.Sprintf("Duration:%v | Bucket Size:%d | PreAllocation:%d | Insertions:%d", d, s, a, i)})
				}
			}
		}
	}
	b.ResetTimer()
	for _, option := range options {
		var w = NewTimeWindow(option.d, option.s, option.a)
		b.Run(option.n, func(bt *testing.B) {
			b.ResetTimer()
			for n := 0; n < bt.N; n = n + 1 {
				for x := 0; x < option.i; x = x + 1 {
					w.Feed(1)
				}
			}
		})
	}
}

func TestPointWindowInterfaces(t *testing.T) {
	var w = NewPointWindow(100)
	requiresIterator(w)
	requiresFeeder(w)
	requiresRollingWindow(w)
	requiresEmbeddedFeeder(w)
}

func TestPointWindow(t *testing.T) {
	var numberOfPoints = 100
	var w = NewPointWindow(numberOfPoints)
	for x := 0; x < numberOfPoints; x = x + 1 {
		w.Feed(1)
	}
	var final float64
	w.Iterate(func(p float64) {
		final = final + p
	})
	if final != float64(numberOfPoints) {
		t.Fatal(final)
	}
}

func TestPointWindowDataRace(t *testing.T) {
	var numberOfPoints = 100
	var w = NewPointWindow(numberOfPoints)
	var stop = make(chan bool)
	go func() {
		for {
			select {
			case <-stop:
				return
			default:
				w.Feed(1)
				time.Sleep(time.Millisecond)
			}
		}
	}()
	go func() {
		var v float64
		for {
			select {
			case <-stop:
				return
			default:
				w.Iterate(func(p float64) {
					v = v + p
					v = math.Mod(v, float64(numberOfPoints))
				})
			}
		}
	}()
	time.Sleep(time.Second)
	close(stop)
}

func BenchmarkPointWindow(b *testing.B) {
	var bucketSizes = []int{1, 10, 100, 1000, 10000}
	var insertions = []int{1, 1000, 10000}
	for _, size := range bucketSizes {
		for _, insertion := range insertions {
			var w = NewPointWindow(size)
			b.Run(fmt.Sprintf("Window Size:%d | Insertions:%d", size, insertion), func(bt *testing.B) {
				b.ResetTimer()
				for n := 0; n < bt.N; n = n + 1 {
					for x := 0; x < insertion; x = x + 1 {
						w.Feed(1)
					}
				}
			})
		}
	}
}
