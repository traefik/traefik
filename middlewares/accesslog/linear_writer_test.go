package accesslog

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLinearWriter(t *testing.T) {
	file := logfilePath("")

	w, err := os.Create(file)
	assert.Nil(t, err, "%v", err)
	defer w.Close()

	lw := LinearWriter(w)
	wg := &sync.WaitGroup{}

	n := 10
	wg.Add(n)
	for i := 0; i < n; i++ {
		go func(j int) {
			fmt.Fprintf(lw, "%d\n", j)
			wg.Done()
		}(i)
	}
	wg.Wait()

	err = lw.Close()
	assert.Nil(t, err, "%v", err)

	// read back and verify
	counts := make([]int, n)
	r, err := os.Open(file)
	assert.Nil(t, err, "%v", err)
	defer os.Remove(file)
	defer r.Close()

	scanner := bufio.NewScanner(r)
	for i := 0; i < n; i++ {
		assert.True(t, scanner.Scan())
		line := scanner.Text()
		num, err := strconv.Atoi(line)
		assert.Nil(t, err, "%v", err)
		if 0 <= num && num < n {
			counts[num]++
		} else {
			assert.Fail(t, "Wrong line number", "Got %d", num)
		}
	}

	for i := 0; i < n; i++ {
		assert.Equal(t, 1, counts[i], "Line %d should have count=1 but was %d", i, counts[i])
	}
}
