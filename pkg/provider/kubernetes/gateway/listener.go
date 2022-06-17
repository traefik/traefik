package gateway

import "sigs.k8s.io/gateway-api/apis/v1alpha2"

var shareableListenerProtocols []v1alpha2.ProtocolType = []v1alpha2.ProtocolType{
	v1alpha2.HTTPSProtocolType,
}

// allocatedPortBlock tracks listeners sharing a common port.
type allocatedPortBlock []v1alpha2.Listener

func (apb allocatedPortBlock) accepts(l v1alpha2.Listener, shareable []v1alpha2.ProtocolType) bool {
	if len(apb) == 0 {
		return true
	}
	// if the protocols do not match, port sharing is not possible
	if apb[0].Protocol != l.Protocol {
		return false
	}
	// we might have excluded the protocol from multi-listener setups
	found := false
	for _, lookup := range shareable {
		if lookup == apb[0].Protocol {
			found = true
		}
	}
	if !found {
		return false
	}
	// ensure the uniqueness of the (protocol, hostname, port) tuple
	// by testing for hostname collision
	collision := false
	for _, existing := range apb {
		if existing.Hostname == nil || l.Hostname == nil {
			if existing.Hostname == l.Hostname {
				collision = true
			}
		}
		if existing.Hostname != nil && l.Hostname != nil {
			if *(existing.Hostname) == *(l.Hostname) {
				collision = true
			}
		}
	}
	return !collision
}

func newAllocatedPortBlock(l v1alpha2.Listener) allocatedPortBlock {
	block := make(allocatedPortBlock, 0)
	block = append(block, l)
	return block
}

// allocatedListeners tracks listeners belonging to the same gateway.
type allocatedListeners map[v1alpha2.PortNumber]allocatedPortBlock

func newAllocatedListeners() allocatedListeners {
	return make(allocatedListeners)
}

func (al allocatedListeners) accepts(l v1alpha2.Listener, shareable []v1alpha2.ProtocolType) bool {
	if block, ok := al[l.Port]; ok {
		return block.accepts(l, shareable)
	}
	return true
}

func (al allocatedListeners) add(l v1alpha2.Listener) {
	var block []v1alpha2.Listener
	var ok bool
	if block, ok = al[l.Port]; ok {
		block = append(block, l)
	} else {
		block = newAllocatedPortBlock(l)
	}
	al[l.Port] = block
}

func (al allocatedListeners) count() int {
	sum := 0
	for _, block := range al {
		sum += len(block)
	}
	return sum
}
