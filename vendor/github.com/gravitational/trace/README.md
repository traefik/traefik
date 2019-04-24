# Trace

[![GoDoc](https://godoc.org/github.com/gravitational/trace?status.png)](https://godoc.org/github.com/gravitational/trace)


Package for error handling and error reporting

Read more here:

http://gravitational.com/blog/golang_error_handling/

### Capture file, line and function

```golang

import (
     "github.com/gravitational/trace"
)

func someFunc() error {
   return trace.Wrap(err)
}


func main() {
  err := someFunc()
  fmt.Println(err.Error()) // prints file, line and function
}
```

### Emit structured logs to Elastic search using udpbeat

**Add trace's document template to your ElasticSearch cluster**

```shell
curl -XPUT 'http://localhost:9200/_template/trace' -d@udbbeat/template.json
```

**Start udpbeat UDP logs collector and emitter**

```shell
go get github.com/gravitational/udpbeat
udpbeat
```

**Hook up logger to UDP collector**

In your code, attach a logrus hook to use udpbeat:

```golang

import (
   "github.com/gravitational/trace"
   log "github.com/sirupsen/logrus"
)

func main() {
   hook, err := trace.NewUDPHook()
   if err != nil {
       log.Fatalf(err)
   }
   log.SetHook(hook)
}
```

Done! You will get structured logs capturing output, log and error message.
You can edit `udpbeat/template.json` to modify emitted fields to whatever makes sense to your app.





