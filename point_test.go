package rolling

import (
	"fmt"
	"math"
	"testing"
	"time"
)

func TestPointWindow(t *testing.T) {
	var numberOfPoints = 100
	var w = NewPointWindow(numberOfPoints)
	for x := 0; x < numberOfPoints; x = x + 1 {
		w.Append(1)
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
			b.Run(fmt.Sprintf("Window Size:%d | Insertions:%d", size, insertion), func(bt *testing.B) {
				var w = NewPointWindow(size)
				bt.ResetTimer()
				for n := 0; n < bt.N; n = n + 1 {
					for x := 0; x < insertion; x = x + 1 {
						w.Append(1)
					}
				}
			})
		}
	}
}
