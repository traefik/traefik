// mgo - MongoDB driver for Go
//
// Copyright (c) 2010-2012 - Gustavo Niemeyer <gustavo@niemeyer.net>
//
// All rights reserved.
//
// Redistribution and use in source and binary forms, with or without
// modification, are permitted provided that the following conditions are met:
//
// 1. Redistributions of source code must retain the above copyright notice, this
//    list of conditions and the following disclaimer.
// 2. Redistributions in binary form must reproduce the above copyright notice,
//    this list of conditions and the following disclaimer in the documentation
//    and/or other materials provided with the distribution.
//
// THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS" AND
// ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE IMPLIED
// WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE ARE
// DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT OWNER OR CONTRIBUTORS BE LIABLE FOR
// ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES
// (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES;
// LOSS OF USE, DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND
// ON ANY THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT
// (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE OF THIS
// SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.

package mgo

type queue struct {
	elems               []interface{}
	nelems, popi, pushi int
}

func (q *queue) Len() int {
	return q.nelems
}

func (q *queue) Push(elem interface{}) {
	//debugf("Pushing(pushi=%d popi=%d cap=%d): %#v\n",
	//       q.pushi, q.popi, len(q.elems), elem)
	if q.nelems == len(q.elems) {
		q.expand()
	}
	q.elems[q.pushi] = elem
	q.nelems++
	q.pushi = (q.pushi + 1) % len(q.elems)
	//debugf(" Pushed(pushi=%d popi=%d cap=%d): %#v\n",
	//       q.pushi, q.popi, len(q.elems), elem)
}

func (q *queue) Pop() (elem interface{}) {
	//debugf("Popping(pushi=%d popi=%d cap=%d)\n",
	//       q.pushi, q.popi, len(q.elems))
	if q.nelems == 0 {
		return nil
	}
	elem = q.elems[q.popi]
	q.elems[q.popi] = nil // Help GC.
	q.nelems--
	q.popi = (q.popi + 1) % len(q.elems)
	//debugf(" Popped(pushi=%d popi=%d cap=%d): %#v\n",
	//       q.pushi, q.popi, len(q.elems), elem)
	return elem
}

func (q *queue) expand() {
	curcap := len(q.elems)
	var newcap int
	if curcap == 0 {
		newcap = 8
	} else if curcap < 1024 {
		newcap = curcap * 2
	} else {
		newcap = curcap + (curcap / 4)
	}
	elems := make([]interface{}, newcap)

	if q.popi == 0 {
		copy(elems, q.elems)
		q.pushi = curcap
	} else {
		newpopi := newcap - (curcap - q.popi)
		copy(elems, q.elems[:q.popi])
		copy(elems[newpopi:], q.elems[q.popi:])
		q.popi = newpopi
	}
	for i := range q.elems {
		q.elems[i] = nil // Help GC.
	}
	q.elems = elems
}
