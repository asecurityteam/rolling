# rolling #

**A zero/low allocation rolling window library.**

## Window Types ##

The package comes with two, basic window types: time and value based.

### Time Windows ###

```golang
var bucketSize = time.Millisecond
var numberOfBuckets = 1000
var preallocHint = 10000
var w =  rolling.NewTimeWindow(bucketSize, numberOfBuckets, preallocHint)
```

The above example shows a setup for a one second rolling window. The time window
is determined by choosing a bucket size, which is used internally to partition
data, and a number of buckets. The bucket size * number of buckets equals the
total time that represents a window.

The pre-allocation hint is used to generate the initial size of the internal
data structures with the intent of helping to avoid allocations at runtime. This
value should be sufficiently high as to contain all the data points that might
be collected within a given bucket. The data structure will expand as needed to
accommodate any number of data elements so this value does not have to be
strictly correct. It is purely an optimisation.

### Value Windows ##

```golang
var windowSize = 1000
var w = rolling.NewPointWindow(windowSize)
```

The above example shows a setup for a one thousand data point rolling window. As
the number of elements exceeds the window size it will wrap around leaving only
the last `windowSize` elements in the window at any given time.

## Collecting Data ##

Windows can be fed at any time and with any valid float64 value by using the
`Feed(float64)` method attached.

```golang
for _ = range time.Tick(time.Millisecond) {
  w.Feed(1)
}
```

Currently, only `float64` values are valid.

## Aggregating Data ##

The most common use case for a rolling window is to produce some aggregate value
from it. Each window allows raw access to data through the
`Iterate(func(float64))` method. This method will call the given function on
each data point contained within the window.

For ease of use, some common aggregations are included in this package. Namely,

- NewCountAggregator(iterator Iterator)
- NewSumAggregator(iterator Iterator)
- NewMinAggregator(iterator Iterator)
- NewMaxAggregator(iterator Iterator)
- NewAverageAggregator(iterator Iterator)
- NewPercentileAggregator(percentile float64, iterator Iterator, preallocHint int)

The count, sum, min, max, and average each run their respected aggregations on
all data contained within a window. The percentile aggregate calculates the Nth
percentile of values where N is any non-negative float64 value between 0.0 and
100.0. Fractional percentiles, like 99.9, are acceptable.

Sometimes one level of aggregation is enough if the intent is to report on some
rolling metric value. However, there is occasionally the need to convert the
aggregate into some other value for decision making. The most common evaluation
is converting the aggregate into some percentage value for decision making. To
make this easier, this package includes a
`NewPercentageAggregator(aggregator Aggregator, lower float64, upper float64)`
which, when the `Aggregate()` method is called, will take the result of the
inner aggregate and generate a value that represents the percentage between
`lower` and `upper` of that value. If the inner aggregate is less than the lower
then the value is always 0.0. If the value is higher than the upper then the
value is always 1.0.

When evaluating data for decision making, it is also common practice to protect
against sparse data. To help with this practice this package also contains a
`NewLimitedAggregator(limit int, iterator Iterator, aggregator Aggregator)`
which will return 0.0 for all calls to `Aggregate()` when the given window
contains less than `limit` values.

## Examples ##

### Rolling Average Of Values ###

```golang
var windowSize = 100
// rolling 100 point window
var window = rolling.NewPointWindow(windowSize)
// aggregate values with an average
var avg = rolling.NewAverageAggregator(window)

for x := 0; x < 1000; x = x + 1 {
  window.Feed(float64(x))
  log.Printf("Average = %f\n", avg.Aggregate())
}
```

### Limited Ten Second Latency Window Reporting On 99th Percentile ###

```golang
var bucketSize = time.Second
var numberOfBuckets = 10
var preallocHint = 1000
// ten second rolling window with a one second bucket
var window =  rolling.NewTimeWindow(bucketSize, numberOfBuckets, preallocHint)
// aggregate to a 99th percentile
var percentile = rolling.NewPercentileAggregate(99, w, preallocHint)
// start emitting non-zero values after 100ms and emit for all values over 1s
var percentage = rolling.NewPercentageEvaluator(a, .1, 1)
// ensure that there are at least as many points as required to satisfy the percentile
var limited = rolling.NewLimitedEvaluator(100, w, percentage)

for _ = range time.Tick(time.Millisecond) {
  var start = time.Now()
  // do some work here
  w.Feed(time.Since(start).Seconds())
  // roll a die and determine if we should report the latency. reporting will
  // get more frequent as the 99th percentile of latency approaches 1s. all 99th
  // percentiles beyond 1s will be reported.
  var chance = rand.Float64()
  if chance < limited.Evaluate() {
    log.Printf("99th = %f\n", percentile.Aggregate())
  }
}
```

### Dice Roll Percentage From Multiple Metrics ###

```golang
var bucketSize = time.Millisecond
var numberOfBuckets = 1000
var preallocHint = 1000
// one second rolling windows for latency data
var incomingRequests = rolling.NewTimeWindow(bucketSize, numberOfBuckets, preallocHint)
var outgoingrequests = rolling.NewTimeWindow(bucketSize, numberOfBuckets, preallocHint)

go func(w rolling.Window){
  for {
    // Record incoming latency data
  }
}(incomingRequests)
go func(w rolling.Window){
  for {
    // Record outgoing latency data
  }
}(outgoingRequests)

var incomingAvg = rolling.NewAverageAggregator(incomingRequests)
var outgoingAvg = rolling.NewAverageAggregator(outgoingRequests)
var combined = rolling.NewAggregatorIterator(incomingAvg, outgoingAvg)
var maxCombined = rolling.NewMaxAggregator(combined)
var lower = .1
var upper = 1.0
// generate a percentage between 100ms and 1000ms of the highest reported avg
// latency for incoming and outgoing request metrics.
var percCombined = rolling.NewPercentageAggregator(maxCombined, lower, upper)
```

## Contributing ##

### License ###

This project is licensed under Apache 2.0. See LICENSE.txt for details.

### Contributing Agreement ###

Atlassian requires signing a contributor's agreement before we can accept a
patch. If you are an individual you can fill out the
[individual CLA](https://na2.docusign.net/Member/PowerFormSigning.aspx?PowerFormId=3f94fbdc-2fbe-46ac-b14c-5d152700ae5d).
If you are contributing on behalf of your company then please fill out the
[corporate CLA](https://na2.docusign.net/Member/PowerFormSigning.aspx?PowerFormId=e1c17c66-ca4d-4aab-a953-2c231af4a20b).
