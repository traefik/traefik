// Copyright 2016 Google Inc. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package logging_test

import (
	"fmt"
	"os"

	"cloud.google.com/go/logging"
	"golang.org/x/net/context"
)

func ExampleNewClient() {
	ctx := context.Background()
	client, err := logging.NewClient(ctx, "my-project")
	if err != nil {
		// TODO: Handle error.
	}
	// Use client to manage logs, metrics and sinks.
	// Close the client when finished.
	if err := client.Close(); err != nil {
		// TODO: Handle error.
	}
}

func ExampleClient_Ping() {
	ctx := context.Background()
	client, err := logging.NewClient(ctx, "my-project")
	if err != nil {
		// TODO: Handle error.
	}
	if err := client.Ping(ctx); err != nil {
		// TODO: Handle error.
	}
}

func ExampleNewClient_errorFunc() {
	ctx := context.Background()
	client, err := logging.NewClient(ctx, "my-project")
	if err != nil {
		// TODO: Handle error.
	}
	// Print all errors to stdout.
	client.OnError = func(e error) {
		fmt.Fprintf(os.Stdout, "logging: %v", e)
	}
	// Use client to manage logs, metrics and sinks.
	// Close the client when finished.
	if err := client.Close(); err != nil {
		// TODO: Handle error.
	}
}

func ExampleClient_Logger() {
	ctx := context.Background()
	client, err := logging.NewClient(ctx, "my-project")
	if err != nil {
		// TODO: Handle error.
	}
	lg := client.Logger("my-log")
	_ = lg // TODO: use the Logger.
}

func ExampleLogger_LogSync() {
	ctx := context.Background()
	client, err := logging.NewClient(ctx, "my-project")
	if err != nil {
		// TODO: Handle error.
	}
	lg := client.Logger("my-log")
	err = lg.LogSync(ctx, logging.Entry{Payload: "red alert"})
	if err != nil {
		// TODO: Handle error.
	}
}

func ExampleLogger_Log() {
	ctx := context.Background()
	client, err := logging.NewClient(ctx, "my-project")
	if err != nil {
		// TODO: Handle error.
	}
	lg := client.Logger("my-log")
	lg.Log(logging.Entry{Payload: "something happened"})
}

func ExampleLogger_Flush() {
	ctx := context.Background()
	client, err := logging.NewClient(ctx, "my-project")
	if err != nil {
		// TODO: Handle error.
	}
	lg := client.Logger("my-log")
	lg.Log(logging.Entry{Payload: "something happened"})
	lg.Flush()
}

func ExampleLogger_StandardLogger() {
	ctx := context.Background()
	client, err := logging.NewClient(ctx, "my-project")
	if err != nil {
		// TODO: Handle error.
	}
	lg := client.Logger("my-log")
	slg := lg.StandardLogger(logging.Info)
	slg.Println("an informative message")
}

func ExampleParseSeverity() {
	sev := logging.ParseSeverity("ALERT")
	fmt.Println(sev)
	// Output: Alert
}
