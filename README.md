# rolling #

**A rolling/sliding window implementation for Google-golang**

## Usage ##

The package offers two different forms of a rolling window: one based on
the number of data point and another based on time.

### Point Window ###

```golang
var w = rolling.NewPointWindow(5)

for x := 0; x < 5; x = x + 1 {
  w.Append(x)
}
w.Iterate(func(v float64) { fmt.Printf("%d ")}) // 0 1 2 3 4
w.Append(5)
w.Iterate(func(v float64) { fmt.Printf("%d ")}) // 5 1 2 3 4
w.Append(6)
w.Iterate(func(v float64) { fmt.Printf("%d ")}) // 5 6 2 3 4
```

The above creates a window that always contains 5 data points and then fills
it with the values 0 - 4. When the next value is appended it will overwrite
the first value. The window continuously overwrites the oldest value with the
latest to preserve the specified value count. This type of window is useful
for collecting data that have a known interval on which they are capture or
for tracking data where time is not a factor.

### Time Window ###

```golang
var w = rolling.NewTimeWindow(time.Millisecond, 3000)
var start = time.Now()
for range time.Tick(time.Millisecond) {
  if time.Since(start) > 3*time.Second {
    break
  }
  w.Append(1)
}
```

The above creates a time window that contains 3,000 buckets where each bucket
contains, at most, 1ms of recorded data. The subsequent loop populates each
bucket with exactly one measure (the value 1) and stops when the window is full.
As time progresses, the oldest values will be removed such that if the above
code performed a `time.Sleep(3*time.Second)` then the window would be empty
again.

The choice of bucket size depends on the frequency with which data are expected
to be recorded. On each increment of time equal to the given duration the window
will expire one bucket and purge the collected values. The smaller the bucket
duration then the less data are lost when a bucket expires.

This type of bucket is most useful for collecting real-time values such as
request rates, error rates, and latencies of operations.

## Aggregating Windows ##

Each window exposes an `Iterate(func(float64))` method that can be used to
access the data stored within. Most uses of this method are reductions, or
aggregations, of the data. This package includes some common aggregations
as helpers.

```golang
fmt.Println(rolling.Count(w))
fmt.Println(rolling.Avg(w))
fmt.Println(rolling.Min(w))
fmt.Println(rolling.Max(w))
fmt.Println(rolling.Sum(w))
fmt.Println(rolling.Percentile(99.9)(w))
fmt.Println(rolling.FastPercentile(99.9)(w))
```

The `Count`, `Avg`, `Min`, `Max`, and `Sum` aggregators consume anything that
has an `Iterate(func(float64))` method and perform their expected computation.
The `Percentile` aggregator first takes the target percentile and returns
an aggregating function that works identically to the `Sum`, et all.

For cases of very large datasets, the `FastPercentile` can be used as a
replacement for the standard percentile calculation. This alternative version
uses the p-squared algorithm for estimating the percentile by processing
only one value at a time, in any order. The results are quite accurate but can
vary from the *actual* percentile by a small amount. It's a tradeoff of accuracy
for speed when calculating percentiles from large data sets. For more on the
p-squared algorithm see: <http://www.cs.wustl.edu/~jain/papers/ftp/psqr.pdf>.

## Contributors ##

Pull requests, issues and comments welcome. For pull requests:

*   Add tests for new features and bug fixes
*   Follow the existing style
*   Separate unrelated changes into multiple pull requests

See the existing issues for things to start contributing.

For bigger changes, make sure you start a discussion first by creating
an issue and explaining the intended change.

Atlassian requires contributors to sign a Contributor License Agreement,
known as a CLA. This serves as a record stating that the contributor is
entitled to contribute the code/documentation/translation to the project
and is willing to have it used in distributions and derivative works
(or is willing to transfer ownership).

Prior to accepting your contributions we ask that you please follow the appropriate
link below to digitally sign the CLA. The Corporate CLA is for those who are
contributing as a member of an organization and the individual CLA is for
those contributing as an individual.

*   [CLA for corporate contributors](https://na2.docusign.net/Member/PowerFormSigning.aspx?PowerFormId=e1c17c66-ca4d-4aab-a953-2c231af4a20b)
*   [CLA for individuals](https://na2.docusign.net/Member/PowerFormSigning.aspx?PowerFormId=3f94fbdc-2fbe-46ac-b14c-5d152700ae5d)

## License ##

Copyright (c) 2017 Atlassian and others.
Apache 2.0 licensed, see [LICENSE.txt](LICENSE.txt) file.
