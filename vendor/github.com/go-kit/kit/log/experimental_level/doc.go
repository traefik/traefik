// Package level is an EXPERIMENTAL levelled logging package. The API will
// definitely have breaking changes and may be deleted altogether. Be warned!
//
// To use the level package, create a logger as per normal in your func main,
// and wrap it with level.New.
//
//    var logger log.Logger
//    logger = log.NewLogfmtLogger(os.Stderr)
//    logger = level.New(logger, level.Config{Allowed: level.AllowInfoAndAbove}) // <--
//    logger = log.NewContext(logger).With("ts", log.DefaultTimestampUTC)
//
// Then, at the callsites, use one of the level.Debug, Info, Warn, or Error
// helper methods to emit leveled log events.
//
//    logger.Log("foo", "bar") // as normal, no level
//    level.Debug(logger).Log("request_id", reqID, "trace_data", trace.Get())
//    if value > 100 {
//        level.Error(logger).Log("value", value)
//    }
//
// The leveled logger allows precise control over what should happen if a log
// event is emitted without a level key, or if a squelched level is used. Check
// the Config struct for details. And, you can easily use non-default level
// values: create new string constants for whatever you want to change, pass
// them explicitly to the Config struct, and write your own level.Foo-style
// helper methods.
package level
