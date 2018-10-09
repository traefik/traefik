// +build !go1.7

package tomb

import (
	"golang.org/x/net/context"
)

// WithContext returns a new tomb that is killed when the provided parent
// context is canceled, and a copy of parent with a replaced Done channel
// that is closed when either the tomb is dying or the parent is canceled.
// The returned context may also be obtained via the tomb's Context method.
func WithContext(parent context.Context) (*Tomb, context.Context) {
	var t Tomb
	t.init()
	if parent.Done() != nil {
		go func() {
			select {
			case <-t.Dying():
			case <-parent.Done():
				t.Kill(parent.Err())
			}
		}()
	}
	t.parent = parent
	child, cancel := context.WithCancel(parent)
	t.addChild(parent, child, cancel)
	return &t, child
}

// Context returns a context that is a copy of the provided parent context with
// a replaced Done channel that is closed when either the tomb is dying or the
// parent is cancelled.
//
// If parent is nil, it defaults to the parent provided via WithContext, or an
// empty background parent if the tomb wasn't created via WithContext.
func (t *Tomb) Context(parent context.Context) context.Context {
	t.init()
	t.m.Lock()
	defer t.m.Unlock()

	if parent == nil {
		if t.parent == nil {
			t.parent = context.Background()
		}
		parent = t.parent.(context.Context)
	}

	if child, ok := t.child[parent]; ok {
		return child.context.(context.Context)
	}

	child, cancel := context.WithCancel(parent)
	t.addChild(parent, child, cancel)
	return child
}

func (t *Tomb) addChild(parent context.Context, child context.Context, cancel func()) {
	if t.reason != ErrStillAlive {
		cancel()
		return
	}
	if t.child == nil {
		t.child = make(map[interface{}]childContext)
	}
	t.child[parent] = childContext{child, cancel, child.Done()}
	for parent, child := range t.child {
		select {
		case <-child.done:
			delete(t.child, parent)
		default:
		}
	}
}
