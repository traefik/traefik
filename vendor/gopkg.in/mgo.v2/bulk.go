package mgo

import (
	"bytes"
	"sort"

	"gopkg.in/mgo.v2/bson"
)

// Bulk represents an operation that can be prepared with several
// orthogonal changes before being delivered to the server.
//
// MongoDB servers older than version 2.6 do not have proper support for bulk
// operations, so the driver attempts to map its API as much as possible into
// the functionality that works. In particular, in those releases updates and
// removals are sent individually, and inserts are sent in bulk but have
// suboptimal error reporting compared to more recent versions of the server.
// See the documentation of BulkErrorCase for details on that.
//
// Relevant documentation:
//
//   http://blog.mongodb.org/post/84922794768/mongodbs-new-bulk-api
//
type Bulk struct {
	c       *Collection
	opcount int
	actions []bulkAction
	ordered bool
}

type bulkOp int

const (
	bulkInsert bulkOp = iota + 1
	bulkUpdate
	bulkUpdateAll
	bulkRemove
)

type bulkAction struct {
	op   bulkOp
	docs []interface{}
	idxs []int
}

type bulkUpdateOp []interface{}
type bulkDeleteOp []interface{}

// BulkResult holds the results for a bulk operation.
type BulkResult struct {
	Matched  int
	Modified int // Available only for MongoDB 2.6+

	// Be conservative while we understand exactly how to report these
	// results in a useful and convenient way, and also how to emulate
	// them with prior servers.
	private bool
}

// BulkError holds an error returned from running a Bulk operation.
// Individual errors may be obtained and inspected via the Cases method.
type BulkError struct {
	ecases []BulkErrorCase
}

func (e *BulkError) Error() string {
	if len(e.ecases) == 0 {
		return "invalid BulkError instance: no errors"
	}
	if len(e.ecases) == 1 {
		return e.ecases[0].Err.Error()
	}
	msgs := make([]string, 0, len(e.ecases))
	seen := make(map[string]bool)
	for _, ecase := range e.ecases {
		msg := ecase.Err.Error()
		if !seen[msg] {
			seen[msg] = true
			msgs = append(msgs, msg)
		}
	}
	if len(msgs) == 1 {
		return msgs[0]
	}
	var buf bytes.Buffer
	buf.WriteString("multiple errors in bulk operation:\n")
	for _, msg := range msgs {
		buf.WriteString("  - ")
		buf.WriteString(msg)
		buf.WriteByte('\n')
	}
	return buf.String()
}

type bulkErrorCases []BulkErrorCase

func (slice bulkErrorCases) Len() int           { return len(slice) }
func (slice bulkErrorCases) Less(i, j int) bool { return slice[i].Index < slice[j].Index }
func (slice bulkErrorCases) Swap(i, j int)      { slice[i], slice[j] = slice[j], slice[i] }

// BulkErrorCase holds an individual error found while attempting a single change
// within a bulk operation, and the position in which it was enqueued.
//
// MongoDB servers older than version 2.6 do not have proper support for bulk
// operations, so the driver attempts to map its API as much as possible into
// the functionality that works. In particular, only the last error is reported
// for bulk inserts and without any positional information, so the Index
// field is set to -1 in these cases.
type BulkErrorCase struct {
	Index int // Position of operation that failed, or -1 if unknown.
	Err   error
}

// Cases returns all individual errors found while attempting the requested changes.
//
// See the documentation of BulkErrorCase for limitations in older MongoDB releases.
func (e *BulkError) Cases() []BulkErrorCase {
	return e.ecases
}

// Bulk returns a value to prepare the execution of a bulk operation.
func (c *Collection) Bulk() *Bulk {
	return &Bulk{c: c, ordered: true}
}

// Unordered puts the bulk operation in unordered mode.
//
// In unordered mode the indvidual operations may be sent
// out of order, which means latter operations may proceed
// even if prior ones have failed.
func (b *Bulk) Unordered() {
	b.ordered = false
}

func (b *Bulk) action(op bulkOp, opcount int) *bulkAction {
	var action *bulkAction
	if len(b.actions) > 0 && b.actions[len(b.actions)-1].op == op {
		action = &b.actions[len(b.actions)-1]
	} else if !b.ordered {
		for i := range b.actions {
			if b.actions[i].op == op {
				action = &b.actions[i]
				break
			}
		}
	}
	if action == nil {
		b.actions = append(b.actions, bulkAction{op: op})
		action = &b.actions[len(b.actions)-1]
	}
	for i := 0; i < opcount; i++ {
		action.idxs = append(action.idxs, b.opcount)
		b.opcount++
	}
	return action
}

// Insert queues up the provided documents for insertion.
func (b *Bulk) Insert(docs ...interface{}) {
	action := b.action(bulkInsert, len(docs))
	action.docs = append(action.docs, docs...)
}

// Remove queues up the provided selectors for removing matching documents.
// Each selector will remove only a single matching document.
func (b *Bulk) Remove(selectors ...interface{}) {
	action := b.action(bulkRemove, len(selectors))
	for _, selector := range selectors {
		if selector == nil {
			selector = bson.D{}
		}
		action.docs = append(action.docs, &deleteOp{
			Collection: b.c.FullName,
			Selector:   selector,
			Flags:      1,
			Limit:      1,
		})
	}
}

// RemoveAll queues up the provided selectors for removing all matching documents.
// Each selector will remove all matching documents.
func (b *Bulk) RemoveAll(selectors ...interface{}) {
	action := b.action(bulkRemove, len(selectors))
	for _, selector := range selectors {
		if selector == nil {
			selector = bson.D{}
		}
		action.docs = append(action.docs, &deleteOp{
			Collection: b.c.FullName,
			Selector:   selector,
			Flags:      0,
			Limit:      0,
		})
	}
}

// Update queues up the provided pairs of updating instructions.
// The first element of each pair selects which documents must be
// updated, and the second element defines how to update it.
// Each pair matches exactly one document for updating at most.
func (b *Bulk) Update(pairs ...interface{}) {
	if len(pairs)%2 != 0 {
		panic("Bulk.Update requires an even number of parameters")
	}
	action := b.action(bulkUpdate, len(pairs)/2)
	for i := 0; i < len(pairs); i += 2 {
		selector := pairs[i]
		if selector == nil {
			selector = bson.D{}
		}
		action.docs = append(action.docs, &updateOp{
			Collection: b.c.FullName,
			Selector:   selector,
			Update:     pairs[i+1],
		})
	}
}

// UpdateAll queues up the provided pairs of updating instructions.
// The first element of each pair selects which documents must be
// updated, and the second element defines how to update it.
// Each pair updates all documents matching the selector.
func (b *Bulk) UpdateAll(pairs ...interface{}) {
	if len(pairs)%2 != 0 {
		panic("Bulk.UpdateAll requires an even number of parameters")
	}
	action := b.action(bulkUpdate, len(pairs)/2)
	for i := 0; i < len(pairs); i += 2 {
		selector := pairs[i]
		if selector == nil {
			selector = bson.D{}
		}
		action.docs = append(action.docs, &updateOp{
			Collection: b.c.FullName,
			Selector:   selector,
			Update:     pairs[i+1],
			Flags:      2,
			Multi:      true,
		})
	}
}

// Upsert queues up the provided pairs of upserting instructions.
// The first element of each pair selects which documents must be
// updated, and the second element defines how to update it.
// Each pair matches exactly one document for updating at most.
func (b *Bulk) Upsert(pairs ...interface{}) {
	if len(pairs)%2 != 0 {
		panic("Bulk.Update requires an even number of parameters")
	}
	action := b.action(bulkUpdate, len(pairs)/2)
	for i := 0; i < len(pairs); i += 2 {
		selector := pairs[i]
		if selector == nil {
			selector = bson.D{}
		}
		action.docs = append(action.docs, &updateOp{
			Collection: b.c.FullName,
			Selector:   selector,
			Update:     pairs[i+1],
			Flags:      1,
			Upsert:     true,
		})
	}
}

// Run runs all the operations queued up.
//
// If an error is reported on an unordered bulk operation, the error value may
// be an aggregation of all issues observed. As an exception to that, Insert
// operations running on MongoDB versions prior to 2.6 will report the last
// error only due to a limitation in the wire protocol.
func (b *Bulk) Run() (*BulkResult, error) {
	var result BulkResult
	var berr BulkError
	var failed bool
	for i := range b.actions {
		action := &b.actions[i]
		var ok bool
		switch action.op {
		case bulkInsert:
			ok = b.runInsert(action, &result, &berr)
		case bulkUpdate:
			ok = b.runUpdate(action, &result, &berr)
		case bulkRemove:
			ok = b.runRemove(action, &result, &berr)
		default:
			panic("unknown bulk operation")
		}
		if !ok {
			failed = true
			if b.ordered {
				break
			}
		}
	}
	if failed {
		sort.Sort(bulkErrorCases(berr.ecases))
		return nil, &berr
	}
	return &result, nil
}

func (b *Bulk) runInsert(action *bulkAction, result *BulkResult, berr *BulkError) bool {
	op := &insertOp{b.c.FullName, action.docs, 0}
	if !b.ordered {
		op.flags = 1 // ContinueOnError
	}
	lerr, err := b.c.writeOp(op, b.ordered)
	return b.checkSuccess(action, berr, lerr, err)
}

func (b *Bulk) runUpdate(action *bulkAction, result *BulkResult, berr *BulkError) bool {
	lerr, err := b.c.writeOp(bulkUpdateOp(action.docs), b.ordered)
	if lerr != nil {
		result.Matched += lerr.N
		result.Modified += lerr.modified
	}
	return b.checkSuccess(action, berr, lerr, err)
}

func (b *Bulk) runRemove(action *bulkAction, result *BulkResult, berr *BulkError) bool {
	lerr, err := b.c.writeOp(bulkDeleteOp(action.docs), b.ordered)
	if lerr != nil {
		result.Matched += lerr.N
		result.Modified += lerr.modified
	}
	return b.checkSuccess(action, berr, lerr, err)
}

func (b *Bulk) checkSuccess(action *bulkAction, berr *BulkError, lerr *LastError, err error) bool {
	if lerr != nil && len(lerr.ecases) > 0 {
		for i := 0; i < len(lerr.ecases); i++ {
			// Map back from the local error index into the visible one.
			ecase := lerr.ecases[i]
			idx := ecase.Index
			if idx >= 0 {
				idx = action.idxs[idx]
			}
			berr.ecases = append(berr.ecases, BulkErrorCase{idx, ecase.Err})
		}
		return false
	} else if err != nil {
		for i := 0; i < len(action.idxs); i++ {
			berr.ecases = append(berr.ecases, BulkErrorCase{action.idxs[i], err})
		}
		return false
	}
	return true
}
