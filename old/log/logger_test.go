package log

import (
	"io/ioutil"
	"os"
	"strings"
	"testing"
	"time"
)

func TestLogRotation(t *testing.T) {
	tempDir, err := ioutil.TempDir("", "traefik_")
	if err != nil {
		t.Fatalf("Error setting up temporary directory: %s", err)
	}

	fileName := tempDir + "traefik.log"
	if err := OpenFile(fileName); err != nil {
		t.Fatalf("Error opening temporary file %s: %s", fileName, err)
	}
	defer CloseFile()

	rotatedFileName := fileName + ".rotated"

	iterations := 20
	halfDone := make(chan bool)
	writeDone := make(chan bool)
	go func() {
		for i := 0; i < iterations; i++ {
			Println("Test log line")
			if i == iterations/2 {
				halfDone <- true
			}
		}
		writeDone <- true
	}()

	<-halfDone
	err = os.Rename(fileName, rotatedFileName)
	if err != nil {
		t.Fatalf("Error renaming file: %s", err)
	}

	err = RotateFile()
	if err != nil {
		t.Fatalf("Error rotating file: %s", err)
	}

	select {
	case <-writeDone:
		gotLineCount := lineCount(t, fileName) + lineCount(t, rotatedFileName)
		if iterations != gotLineCount {
			t.Errorf("Wanted %d written log lines, got %d", iterations, gotLineCount)
		}
	case <-time.After(500 * time.Millisecond):
		t.Fatalf("test timed out")
	}

	close(halfDone)
	close(writeDone)
}

func lineCount(t *testing.T, fileName string) int {
	t.Helper()
	fileContents, err := ioutil.ReadFile(fileName)
	if err != nil {
		t.Fatalf("Error reading from file %s: %s", fileName, err)
	}

	count := 0
	for _, line := range strings.Split(string(fileContents), "\n") {
		if strings.TrimSpace(line) == "" {
			continue
		}
		count++
	}

	return count
}
