package main

import (
	"fmt"
	"time"

	marathon "github.com/gambol99/go-marathon"
)

const marathonURL = "http://127.0.0.1:8080"

func main() {
	config := marathon.NewDefaultConfig()
	config.URL = marathonURL
	client, err := marathon.NewClient(config)
	if err != nil {
		panic(err)
	}

	app := marathon.Application{}
	app.ID = "tasks-test"
	app.Command("sleep 60")
	app.Count(3)
	fmt.Println("Creating app.")
	// Update application will either create or update the app.
	_, err = client.UpdateApplication(&app, false)
	if err != nil {
		panic(err)
	}

	// wait until marathon will launch tasks
	client.WaitOnApplication(app.ID, 10*time.Second)
	fmt.Println("Tasks were deployed.")

	tasks, err := client.Tasks(app.ID)
	if err != nil {
		panic(err)
	}

	host := tasks.Tasks[0].Host
	fmt.Printf("Killing tasks on the host: %s\n", host)

	_, err = client.KillApplicationTasks(app.ID, &marathon.KillApplicationTasksOpts{Scale: true, Host: host})
	if err != nil {
		panic(err)
	}
}
