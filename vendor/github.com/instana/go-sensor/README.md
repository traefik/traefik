![golang banner 2017-07-11](https://disznc.s3.amazonaws.com/Instana-Go-2017-07-11-at-16.01.45.png)

# Instana Go Sensor
golang-sensor requires Go version 1.7 or greater.

The Instana Go sensor consists of two parts:

* metrics sensor
* [OpenTracing](http://opentracing.io) tracer

[![Build Status](https://travis-ci.org/instana/golang-sensor.svg?branch=master)](https://travis-ci.org/instana/golang-sensor)
[![OpenTracing Badge](https://img.shields.io/badge/OpenTracing-enabled-blue.svg)](http://opentracing.io)

## Sensor

To use sensor only without tracing ability, import the `instana` package and run

```Go
instana.InitSensor(opt)
```

in your main function. The init function takes an `Options` object with the following optional fields:

* **Service** - global service name that will be used to identify the program in the Instana backend
* **AgentHost**, **AgentPort** - default to localhost:42699, set the coordinates of the Instana proxy agent
* **LogLevel** - one of Error, Warn, Info or Debug

Once initialized, the sensor will try to connect to the given Instana agent and in case of connection success will send metrics and snapshot information through the agent to the backend.

## OpenTracing

In case you want to use the OpenTracing tracer, it will automatically initialize the sensor and thus also activate the metrics stream. To activate the global tracer, run for example

```Go
ot.InitGlobalTracer(instana.NewTracerWithOptions(&instana.Options{
	Service:  SERVICE,
	LogLevel: instana.DEBUG}))
```

in your main function. The tracer takes the same options that the sensor takes for initialization, described above.

The tracer is able to protocol and piggyback OpenTracing baggage, tags and logs. Only text mapping is implemented yet, binary is not supported. Also, the tracer tries to map the OpenTracing spans to the Instana model based on OpenTracing recommended tags. See `simple` example for details on how recommended tags are used.

The Instana tracer will remap OpenTracing HTTP headers into Instana Headers, so parallel use with some other OpenTracing model is not possible. The Instana tracer is based on the OpenTracing Go basictracer with necessary modifications to map to the Instana tracing model. Also, sampling isn't implemented yet and will be focus of future work.

## Events API

The sensor, be it instantiated explicitly or implicitly through the tracer, provides a simple wrapper API to send events to Instana as described in [its documentation](https://docs.instana.io/quick_start/api/#event-sdk-rest-web-service).

To learn more, see the [Events API](https://github.com/instana/golang-sensor/blob/master/EventAPI.md) document in this repository.

## Examples

Following examples are included in the `examples` folder:

* **ot-simple/simple.go** - Demonstrates basic usage of the tracer
* **webserver/http.go** - Demonstrates how http server and client should be instrumented
* **rpc/rpc.go** - Demonstrates a basic RPC service
* **event/** - Demonstrates usage of the event API
