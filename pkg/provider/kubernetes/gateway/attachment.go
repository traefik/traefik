package gateway

import (
	"slices"
	"time"

	gatev1 "sigs.k8s.io/gateway-api/apis/v1"
)

type routeClaim struct {
	kind      string
	hostname  gatev1.Hostname
	timestamp time.Time
}

// attachedRoutes tracks which route kind owns each hostname on each listener.
type attachedRoutes map[string][]routeClaim // listenerKey → []routeClaim

func (a attachedRoutes) Claim(listenerKey, kind string, hostname gatev1.Hostname, timestamp time.Time) {
	// 1. Bail if the new claim is not older than every intersecting claim.
	for _, claim := range a[listenerKey] {
		if hostnamesIntersect([]gatev1.Hostname{hostname}, []gatev1.Hostname{claim.hostname}) && !timestamp.Before(claim.timestamp) {
			return
		}
	}

	// 2. Remove any intersecting claims.
	a[listenerKey] = slices.DeleteFunc(a[listenerKey], func(claim routeClaim) bool {
		return hostnamesIntersect([]gatev1.Hostname{hostname}, []gatev1.Hostname{claim.hostname})
	})

	// 3. Append the new claim.
	a[listenerKey] = append(a[listenerKey], routeClaim{hostname: hostname, kind: kind, timestamp: timestamp})
}

func (a attachedRoutes) HasConflict(listenerKey, kind string, hostnames []gatev1.Hostname) bool {
	for _, claim := range a[listenerKey] {
		if claim.kind != kind && hostnamesIntersect(hostnames, []gatev1.Hostname{claim.hostname}) {
			return true
		}
	}

	return false
}

type listenerAttachment struct {
	attached bool
	reason   gatev1.RouteConditionReason

	listenerKey string
	hostnames   []gatev1.Hostname
}

// attachmentForListener returns a listenerAttachment for the given listener and route.
// The attachment indicates whether the route is attached to the listener, and if not, the reason why.
func attachmentForListener(match gatewayListenersForParentRef, listener gatewayListener, routeNamespace, routeKind string, routeHostnames []gatev1.Hostname) listenerAttachment {
	hostnames, ok := findMatchingHostnames(listener.Hostname, routeHostnames)

	attachment := listenerAttachment{
		attached:    true,
		listenerKey: match.gatewayNamespace + "/" + match.gatewayName + "/" + listener.Name,
		hostnames:   hostnames,
	}

	switch {
	case !matchListener(listener, match.parentRef):
		attachment.attached = false
	case !allowRoute(listener, routeNamespace, routeKind):
		attachment.reason = gatev1.RouteReasonNotAllowedByListeners
		attachment.attached = false
	case !ok:
		attachment.reason = gatev1.RouteReasonNoMatchingListenerHostname
		attachment.attached = false
	}

	return attachment
}
