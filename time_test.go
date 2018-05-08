package rolling

import (
	"fmt"
	"math"
	"testing"
	"time"
)

func TestTimeWindow(t *testing.T) {
	var bucketSize = time.Millisecond * 50
	var numberBuckets = 10
	var w = NewTimeWindow(bucketSize, numberBuckets)
	for x := 0; x < numberBuckets; x = x + 1 {
		w.Append(1)
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
	var w = NewTimeWindow(bucketSize, numberBuckets)
	var target = time.Unix(0, 0)
	var adjustedTime, bucket = w.selectBucket(target)
	if bucket != 0 {
		t.Fatalf("expected bucket 0 but got %d %v", bucket, adjustedTime)
	}
	target = time.Unix(0, int64(50*time.Millisecond))
	_, bucket = w.selectBucket(target)
	if bucket != 1 {
		t.Fatalf("expected bucket 1 but got %d %v", bucket, target)
	}
	target = time.Unix(0, int64(50*time.Millisecond)*10)
	_, bucket = w.selectBucket(target)
	if bucket != 10 {
		t.Fatalf("expected bucket 10 but got %d %v", bucket, target)
	}
	target = time.Unix(0, int64(50*time.Millisecond)*11)
	_, bucket = w.selectBucket(target)
	if bucket != 0 {
		t.Fatalf("expected bucket 0 but got %d %v", bucket, target)
	}
}

func TestTimeWindowConsistency(t *testing.T) {
	var bucketSize = time.Millisecond * 50
	var numberBuckets = 10
	var w = NewTimeWindow(bucketSize, numberBuckets)
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
	var w = NewTimeWindow(bucketSize, numberBuckets)
	var stop = make(chan bool)
	go func() {
		for {
			select {
			case <-stop:
				return
			default:
				w.Append(1)
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
	name          string
	bucketSize    time.Duration
	numberBuckets int
	insertions    int
}

func BenchmarkTimeWindow(b *testing.B) {
	var durations = []time.Duration{time.Millisecond}
	var bucketSizes = []int{1, 10, 100, 1000}
	var insertions = []int{1, 1000, 10000}
	var options = make([]timeWindowOptions, 0, len(durations)*len(bucketSizes)*len(insertions))
	for _, d := range durations {
		for _, s := range bucketSizes {
			for _, i := range insertions {
				options = append(
					options,
					timeWindowOptions{
						name:          fmt.Sprintf("Duration:%v | Buckets:%d | Insertions:%d", d, s, i),
						bucketSize:    d,
						numberBuckets: s,
						insertions:    i,
					},
				)
			}
		}
	}
	b.ResetTimer()
	for _, option := range options {
		b.Run(option.name, func(bt *testing.B) {
			var w = NewTimeWindow(option.bucketSize, option.numberBuckets)
			bt.ResetTimer()
			for n := 0; n < bt.N; n = n + 1 {
				for x := 0; x < option.insertions; x = x + 1 {
					w.Append(1)
				}
			}
		})
	}
}
