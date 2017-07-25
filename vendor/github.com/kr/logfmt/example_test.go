package logfmt_test

import (
	"bytes"
	"fmt"
	"github.com/kr/logfmt"
	"log"
	"strconv"
)

type Measurement struct {
	Key  string
	Val  float64
	Unit string // (e.g. ms, MB, etc)
}

type Measurements []*Measurement

var measurePrefix = []byte("measure.")

func (mm *Measurements) HandleLogfmt(key, val []byte) error {
	if !bytes.HasPrefix(key, measurePrefix) {
		return nil
	}
	i := bytes.LastIndexFunc(val, isDigit)
	v, err := strconv.ParseFloat(string(val[:i+1]), 10)
	if err != nil {
		return err
	}
	m := &Measurement{
		Key:  string(key[len(measurePrefix):]),
		Val:  v,
		Unit: string(val[i+1:]),
	}
	*mm = append(*mm, m)
	return nil
}

// return true if r is an ASCII digit only, as opposed to unicode.IsDigit.
func isDigit(r rune) bool {
	return '0' <= r && r <= '9'
}

func Example_customHandler() {
	var data = []byte("measure.a=1ms measure.b=10 measure.c=100MB measure.d=1s garbage")

	mm := make(Measurements, 0)
	if err := logfmt.Unmarshal(data, &mm); err != nil {
		log.Fatalf("err=%q", err)
	}
	for _, m := range mm {
		fmt.Printf("%v\n", *m)
	}
	// Output:
	// {a 1 ms}
	// {b 10 }
	// {c 100 MB}
	// {d 1 s}
}
