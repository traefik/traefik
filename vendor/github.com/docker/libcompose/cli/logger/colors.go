package logger

import "fmt"

var (
	colorPrefix = make(chan string)
)

func generateColors() {
	i := 0
	colorOrder := []string{
		"36",   // cyan
		"33",   // yellow
		"32",   // green
		"35",   // magenta
		"31",   // red
		"34",   // blue
		"36;1", // intense cyan
		"33;1", // intense yellow
		"32;1", // intense green
		"35;1", // intense magenta
		"31;1", // intense red
		"34;1", // intense blue
	}

	for {
		colorPrefix <- fmt.Sprintf("\033[%sm%%s |\033[0m", colorOrder[i])
		i = (i + 1) % len(colorOrder)
	}
}

func init() {
	go generateColors()
}
