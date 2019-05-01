// Copyright (c) 2016 Uber Technologies, Inc.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package clock

// timers represents a list of sortable timers.
type Timers []*Timer

func (ts Timers) Len() int { return len(ts) }

func (ts Timers) Swap(i, j int) {
	ts[i], ts[j] = ts[j], ts[i]
}

func (ts Timers) Less(i, j int) bool {
	return ts[i].Next().Before(ts[j].Next())
}

func (ts *Timers) Push(t interface{}) {
	*ts = append(*ts, t.(*Timer))
}

func (ts *Timers) Pop() interface{} {
	t := (*ts)[len(*ts)-1]
	*ts = (*ts)[:len(*ts)-1]
	return t
}
