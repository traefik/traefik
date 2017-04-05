package accesslog

import (
	"bufio"
	"fmt"
	"github.com/stretchr/testify/assert"
	"os"
	"strconv"
	"sync"
	"testing"
)

func TestLinearWriter(t *testing.T) {
	file := logfilePath("")

	w, err := os.Create(file)
	assert.Nil(t, err, "%v", err)

	lw := LinearWriter(w)
	wg := &sync.WaitGroup{}

	n := 10
	wg.Add(n)
	for i := 0; i < n; i++ {
		go func(j int) {
			fmt.Fprintf(lw, "Line %d\n", j)
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
	defer r.Close()
	defer os.Remove(file)

	scanner := bufio.NewScanner(r)
	for i := 0; i < n; i++ {
		assert.True(t, scanner.Scan())
		line := scanner.Text()
		num, err := strconv.Atoi(line[5:])
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
