# Leadership: Distributed Leader Election for Clustered Environments.

Leadership is a library for a cluster leader election on top of a distributed
Key/Value store.

It is built using the `docker/libkv` library and is designed to work across multiple
storage backends.

You can use `leadership` with `Consul`, `etcd` and `Zookeeper`.

```go
// Create a store using pkg/store.
client, err := store.NewStore("consul", []string{"127.0.0.1:8500"}, &store.Config{})
if err != nil {
	panic(err)
}

underwood := leadership.NewCandidate(client, "service/swarm/leader", "underwood", 15*time.Second)
electedCh, _ := underwood.RunForElection()

for isElected := range electedCh {
	// This loop will run every time there is a change in our leadership
	// status.

	if isElected {
		// We won the election - we are now the leader.
		// Let's do leader stuff, for example, sleep for a while.
		log.Printf("I won the election! I'm now the leader")
		time.Sleep(10 * time.Second)

		// Tired of being a leader? You can resign anytime.
		candidate.Resign()
	} else {
		// We lost the election but are still running for leadership.
		// `elected == false` is the default state and is the first event
		// we'll receive from the channel. After a successful election,
		// this event can get triggered if someone else steals the
		// leadership or if we resign.

		log.Printf("Lost the election, let's try another time")
	}
}
```

It is possible to follow an election in real-time and get notified whenever
there is a change in leadership:
```go
follower := leadership.NewFollower(client, "service/swarm/leader")
leaderCh, _ := follower.FollowElection()
for leader := range leaderCh {
	// Leader is a string containing the value passed to `NewCandidate`.
	log.Printf("%s is now the leader", leader)
}
log.Fatal("Cannot follow the election, store is probably down")
// Recovery code or exit
```

A typical use case for this is to be able to always send requests to the current
leader.

## Fault tolerance

Leadership returns an error channel for Candidates and Followers that you can use
to be resilient to failures. For example, if the watch on the leader key fails
because the store becomes unavailable, you can retry the process later.

```go
func participate() {
    // Create a store using pkg/store.
    client, err := store.NewStore("consul", []string{"127.0.0.1:8500"}, &store.Config{})
    if err != nil {
        panic(err)
    }

    waitTime := 10 * time.Second
    underwood := leadership.NewCandidate(client, "service/swarm/leader", "underwood", 15*time.Second)

    go func() {
        for {
            run(underwood)
            time.Sleep(waitTime)
            // retry
        }
    }()
}

func run(candidate *leadership.Candidate) {
    electedCh, errCh := candidate.RunForElection()
    for {
        select {
        case isElected := <-electedCh:
            if isElected {
                // Do something
            } else {
                // Do something else
            }

        case err := <-errCh:
            log.Error(err)
            return
        }
    }
}
```

## License

leadership is licensed under the Apache License, Version 2.0. See [LICENSE](LICENSE) for the full license text.
