# gominlog [![Build Status](https://travis-ci.org/ArthurHlt/gominlog.svg)](https://travis-ci.org/ArthurHlt/gominlog)
A minimal logger for golang with color in terminal and level.

It will also override the flag `log.Lshortfile`, it will give you the path of the current file which is logged without unnecessary sub folder, but for this you need to set the `packageName` value.
this value should be the same name than your root folder project name.

## Requirement

**Go version**: >= 1.5

## Getting start

```go
import "github.com/ArthurHlt/gominlog"

minlog := gominlog.NewClassicMinLog() // you can also use NewClassicMinLogWithPackageName("your-root-folder-name")

// it will provide a minlog with logger wrting in stdout and with flag set to log.Lshortfile | log.Ldate | log.Ltime
// You will have also color in the terminal for different level
// the log level will, by default, be set to Lall (level all)
// finally the packageName will be took from args[0] (this mean the name of your binary normally)

minlog.Debug("test debuglevel")
minlog.Error("test errorlevel")
minlog.Warning("test warninglevel")
minlog.Severe("test severelevel")
minlog.Info("test %s", "infolevel") // you can also pass values like with fmt.Sprintf

// To remove color:
minlog.WithColor(false)

// To set package Name:
minlog.SetPackageName("your-folder-name")

// To set level of logging:
minlog.SetLevel(gominlog.Lwarning) // you have alos Lall, Loff, Lsevere, Lerror, Linfo and Ldebug.
```

example output:
```bash
2015/11/20 18:58:57 /gominlog/gominlog_test.go:16 DEBUG: test debuglevel
2015/11/20 18:58:57 /gominlog/gominlog_test.go:17 ERROR: test errorlevel
2015/11/20 18:58:57 /gominlog/gominlog_test.go:18 SEVERE: test severelevel
2015/11/20 18:58:57 /gominlog/gominlog_test.go:19 INFO: test infolevel
2015/11/20 18:58:57 /gominlog/gominlog_test.go:21 WARNING: test warninglevel
```


