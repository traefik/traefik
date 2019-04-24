# gohistogram - Histograms in Go

![build status](https://circleci.com/gh/VividCortex/gohistogram.png?circle-token=d37ec652ea117165cd1b342400a801438f575209)

This package provides [Streaming Approximate Histograms](https://vividcortex.com/blog/2013/07/08/streaming-approximate-histograms/)
for efficient quantile approximations.

The histograms in this package are based on the algorithms found in
Ben-Haim & Yom-Tov's *A Streaming Parallel Decision Tree Algorithm*
([PDF](http://jmlr.org/papers/volume11/ben-haim10a/ben-haim10a.pdf)).
Histogram bins do not have a preset size. As values stream into
the histogram, bins are dynamically added and merged.

Another implementation can be found in the Apache Hive project (see
[NumericHistogram](http://hive.apache.org/docs/r0.11.0/api/org/apache/hadoop/hive/ql/udf/generic/NumericHistogram.html)).

An example:

![histogram](http://i.imgur.com/5OplaRs.png)

The accurate method of calculating quantiles (like percentiles) requires
data to be sorted. Streaming histograms make it possible to approximate
quantiles without sorting (or even individually storing) values.

NumericHistogram is the more basic implementation of a streaming
histogram. WeightedHistogram implements bin values as exponentially-weighted
moving averages.

A maximum bin size is passed as an argument to the constructor methods. A
larger bin size yields more accurate approximations at the cost of increased
memory utilization and performance.

A picture of kittens:

![stack of kittens](http://i.imgur.com/QxRTWAE.jpg)

## Getting started

### Using in your own code

    $ go get github.com/VividCortex/gohistogram
    
```go
import "github.com/VividCortex/gohistogram"
```

### Running tests and making modifications

Get the code into your workspace:

    $ cd $GOPATH
    $ git clone git@github.com:VividCortex/gohistogram.git ./src/github.com/VividCortex/gohistogram

You can run the tests now:

    $ cd src/github.com/VividCortex/gohistogram
    $ go test .

## API Documentation

Full source documentation can be found [here][godoc].

[godoc]: http://godoc.org/github.com/VividCortex/gohistogram

## Contributing

We only accept pull requests for minor fixes or improvements. This includes:

* Small bug fixes
* Typos
* Documentation or comments

Please open issues to discuss new features. Pull requests for new features will be rejected,
so we recommend forking the repository and making changes in your fork for your use case.

## License

Copyright (c) 2013 VividCortex

Released under MIT License. Check `LICENSE` file for details.
