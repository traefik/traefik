/*
Copyright 2017 Google Inc. All Rights Reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package spanner

import (
	"container/heap"
	"math/rand"
	"reflect"
	"sync"
	"testing"
	"time"

	"golang.org/x/net/context"

	"cloud.google.com/go/spanner/internal/testutil"
	sppb "google.golang.org/genproto/googleapis/spanner/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
)

// setup prepares test environment for regular session pool tests.
func setup(t *testing.T, spc SessionPoolConfig) (sp *sessionPool, sc *testutil.MockCloudSpannerClient, cancel func()) {
	sc = testutil.NewMockCloudSpannerClient(t)
	spc.getRPCClient = func() (sppb.SpannerClient, error) {
		return sc, nil
	}
	spc.HealthCheckInterval = 50 * time.Millisecond
	sp, err := newSessionPool("mockdb", spc, nil)
	if err != nil {
		t.Fatalf("cannot create session pool: %v", err)
	}
	cancel = func() {
		sp.close()
	}
	return
}

// TestSessionCreation tests session creation during sessionPool.Take().
func TestSessionCreation(t *testing.T) {
	sp, sc, cancel := setup(t, SessionPoolConfig{})
	defer cancel()
	// Take three sessions from session pool, this should trigger session pool to create three new sessions.
	shs := make([]*sessionHandle, 3)
	// gotDs holds the unique sessions taken from session pool.
	gotDs := map[string]bool{}
	for i := 0; i < len(shs); i++ {
		var err error
		shs[i], err = sp.take(context.Background())
		if err != nil {
			t.Errorf("failed to get session(%v): %v", i, err)
		}
		gotDs[shs[i].getID()] = true
	}
	if len(gotDs) != len(shs) {
		t.Errorf("session pool created %v sessions, want %v", len(gotDs), len(shs))
	}
	if wantDs := sc.DumpSessions(); !reflect.DeepEqual(gotDs, wantDs) {
		t.Errorf("session pool creates sessions %v, want %v", gotDs, wantDs)
	}
	// Verify that created sessions are recorded correctly in session pool.
	sp.mu.Lock()
	if int(sp.numOpened) != len(shs) {
		t.Errorf("session pool reports %v open sessions, want %v", sp.numOpened, len(shs))
	}
	if sp.createReqs != 0 {
		t.Errorf("session pool reports %v session create requests, want 0", int(sp.createReqs))
	}
	sp.mu.Unlock()
	// Verify that created sessions are tracked correctly by healthcheck queue.
	hc := sp.hc
	hc.mu.Lock()
	if hc.queue.Len() != len(shs) {
		t.Errorf("healthcheck queue length = %v, want %v", hc.queue.Len(), len(shs))
	}
	for _, s := range hc.queue.sessions {
		if !gotDs[s.getID()] {
			t.Errorf("session %v is in healthcheck queue, but it is not created by session pool", s.getID())
		}
	}
	hc.mu.Unlock()
}

// TestTakeFromIdleList tests taking sessions from session pool's idle list.
func TestTakeFromIdleList(t *testing.T) {
	sp, sc, cancel := setup(t, SessionPoolConfig{})
	defer cancel()
	// Take ten sessions from session pool and recycle them.
	shs := make([]*sessionHandle, 10)
	for i := 0; i < len(shs); i++ {
		var err error
		shs[i], err = sp.take(context.Background())
		if err != nil {
			t.Errorf("failed to get session(%v): %v", i, err)
		}
	}
	for i := 0; i < len(shs); i++ {
		shs[i].recycle()
	}
	// Further session requests from session pool won't cause mockclient to create more sessions.
	wantSessions := sc.DumpSessions()
	// Take ten sessions from session pool again, this time all sessions should come from idle list.
	gotSessions := map[string]bool{}
	for i := 0; i < len(shs); i++ {
		sh, err := sp.take(context.Background())
		if err != nil {
			t.Errorf("cannot take session from session pool: %v", err)
		}
		gotSessions[sh.getID()] = true
	}
	if len(gotSessions) != 10 {
		t.Errorf("got %v unique sessions, want 10", len(gotSessions))
	}
	if !reflect.DeepEqual(gotSessions, wantSessions) {
		t.Errorf("got sessions: %v, want %v", gotSessions, wantSessions)
	}
}

// TesttakeWriteSessionFromIdleList tests taking write sessions from session pool's idle list.
func TestTakeWriteSessionFromIdleList(t *testing.T) {
	sp, sc, cancel := setup(t, SessionPoolConfig{})
	defer cancel()
	act := testutil.NewAction("Begin", nil)
	acts := make([]testutil.Action, 20)
	for i := 0; i < len(acts); i++ {
		acts[i] = act
	}
	sc.SetActions(acts...)
	// Take ten sessions from session pool and recycle them.
	shs := make([]*sessionHandle, 10)
	for i := 0; i < len(shs); i++ {
		var err error
		shs[i], err = sp.takeWriteSession(context.Background())
		if err != nil {
			t.Errorf("failed to get session(%v): %v", i, err)
		}
	}
	for i := 0; i < len(shs); i++ {
		shs[i].recycle()
	}
	// Further session requests from session pool won't cause mockclient to create more sessions.
	wantSessions := sc.DumpSessions()
	// Take ten sessions from session pool again, this time all sessions should come from idle list.
	gotSessions := map[string]bool{}
	for i := 0; i < len(shs); i++ {
		sh, err := sp.takeWriteSession(context.Background())
		if err != nil {
			t.Errorf("cannot take session from session pool: %v", err)
		}
		gotSessions[sh.getID()] = true
	}
	if len(gotSessions) != 10 {
		t.Errorf("got %v unique sessions, want 10", len(gotSessions))
	}
	if !reflect.DeepEqual(gotSessions, wantSessions) {
		t.Errorf("got sessions: %v, want %v", gotSessions, wantSessions)
	}
}

// TestTakeFromIdleListChecked tests taking sessions from session pool's idle list, but with a extra ping check.
func TestTakeFromIdleListChecked(t *testing.T) {
	if testing.Short() {
		t.SkipNow()
	}
	sp, sc, cancel := setup(t, SessionPoolConfig{})
	defer cancel()
	// Stop healthcheck workers to simulate slow pings.
	sp.hc.close()
	// Create a session and recycle it.
	sh, err := sp.take(context.Background())
	if err != nil {
		t.Errorf("failed to get session: %v", err)
	}
	wantSid := sh.getID()
	sh.recycle()
	<-time.After(time.Second)
	// Two back-to-back session requests, both of them should return the same session created before and
	// none of them should trigger a session ping.
	for i := 0; i < 2; i++ {
		// Take the session from the idle list and recycle it.
		sh, err = sp.take(context.Background())
		if err != nil {
			t.Errorf("%v - failed to get session: %v", i, err)
		}
		if gotSid := sh.getID(); gotSid != wantSid {
			t.Errorf("%v - got session id: %v, want %v", i, gotSid, wantSid)
		}
		// The two back-to-back session requests shouldn't trigger any session pings because sessionPool.Take
		// reschedules the next healthcheck.
		if got, want := sc.DumpPings(), ([]string{wantSid}); !reflect.DeepEqual(got, want) {
			t.Errorf("%v - got ping session requests: %v, want %v", i, got, want)
		}
		sh.recycle()
	}
	// Inject session error to mockclient, and take the session from the session pool, the old session should be destroyed and
	// the session pool will create a new session.
	sc.InjectError("GetSession", grpc.Errorf(codes.NotFound, "Session not found:"))
	// Delay to trigger sessionPool.Take to ping the session.
	<-time.After(time.Second)
	sh, err = sp.take(context.Background())
	if err != nil {
		t.Errorf("failed to get session: %v", err)
	}
	ds := sc.DumpSessions()
	if len(ds) != 1 {
		t.Errorf("dumped sessions from mockclient: %v, want %v", ds, sh.getID())
	}
	if sh.getID() == wantSid {
		t.Errorf("sessionPool.Take still returns the same session %v, want it to create a new one", wantSid)
	}
}

// TestTakeFromIdleWriteListChecked tests taking sessions from session pool's idle list, but with a extra ping check.
func TestTakeFromIdleWriteListChecked(t *testing.T) {
	if testing.Short() {
		t.SkipNow()
	}
	sp, sc, cancel := setup(t, SessionPoolConfig{})
	defer cancel()
	sc.MakeNice()
	// Stop healthcheck workers to simulate slow pings.
	sp.hc.close()
	// Create a session and recycle it.
	sh, err := sp.takeWriteSession(context.Background())
	if err != nil {
		t.Errorf("failed to get session: %v", err)
	}
	wantSid := sh.getID()
	sh.recycle()
	<-time.After(time.Second)
	// Two back-to-back session requests, both of them should return the same session created before and
	// none of them should trigger a session ping.
	for i := 0; i < 2; i++ {
		// Take the session from the idle list and recycle it.
		sh, err = sp.takeWriteSession(context.Background())
		if err != nil {
			t.Errorf("%v - failed to get session: %v", i, err)
		}
		if gotSid := sh.getID(); gotSid != wantSid {
			t.Errorf("%v - got session id: %v, want %v", i, gotSid, wantSid)
		}
		// The two back-to-back session requests shouldn't trigger any session pings because sessionPool.Take
		// reschedules the next healthcheck.
		if got, want := sc.DumpPings(), ([]string{wantSid}); !reflect.DeepEqual(got, want) {
			t.Errorf("%v - got ping session requests: %v, want %v", i, got, want)
		}
		sh.recycle()
	}
	// Inject session error to mockclient, and take the session from the session pool, the old session should be destroyed and
	// the session pool will create a new session.
	sc.InjectError("GetSession", grpc.Errorf(codes.NotFound, "Session not found:"))
	// Delay to trigger sessionPool.Take to ping the session.
	<-time.After(time.Second)
	sh, err = sp.takeWriteSession(context.Background())
	if err != nil {
		t.Errorf("failed to get session: %v", err)
	}
	ds := sc.DumpSessions()
	if len(ds) != 1 {
		t.Errorf("dumped sessions from mockclient: %v, want %v", ds, sh.getID())
	}
	if sh.getID() == wantSid {
		t.Errorf("sessionPool.Take still returns the same session %v, want it to create a new one", wantSid)
	}
}

// TestMaxOpenedSessions tests max open sessions constraint.
func TestMaxOpenedSessions(t *testing.T) {
	if testing.Short() {
		t.SkipNow()
	}
	sp, _, cancel := setup(t, SessionPoolConfig{MaxOpened: 1})
	defer cancel()
	sh1, err := sp.take(context.Background())
	if err != nil {
		t.Errorf("cannot take session from session pool: %v", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	// Session request will timeout due to the max open sessions constraint.
	sh2, gotErr := sp.take(ctx)
	if wantErr := errGetSessionTimeout(); !reflect.DeepEqual(gotErr, wantErr) {
		t.Errorf("the second session retrival returns error %v, want %v", gotErr, wantErr)
	}
	go func() {
		<-time.After(time.Second)
		// destroy the first session to allow the next session request to proceed.
		sh1.destroy()
	}()
	// Now session request can be processed because the first session will be destroyed.
	sh2, err = sp.take(context.Background())
	if err != nil {
		t.Errorf("after the first session is destroyed, session retrival still returns error %v, want nil", err)
	}
	if !sh2.session.isValid() || sh2.getID() == "" {
		t.Errorf("got invalid session: %v", sh2.session)
	}
}

// TestMinOpenedSessions tests min open session constraint.
func TestMinOpenedSessions(t *testing.T) {
	sp, _, cancel := setup(t, SessionPoolConfig{MinOpened: 1})
	defer cancel()
	// Take ten sessions from session pool and recycle them.
	var ss []*session
	var shs []*sessionHandle
	for i := 0; i < 10; i++ {
		sh, err := sp.take(context.Background())
		if err != nil {
			t.Errorf("failed to get session(%v): %v", i, err)
		}
		ss = append(ss, sh.session)
		shs = append(shs, sh)
		sh.recycle()
	}
	for _, sh := range shs {
		sh.recycle()
	}
	// Simulate session expiration.
	for _, s := range ss {
		s.destroy(true)
	}
	sp.mu.Lock()
	defer sp.mu.Unlock()
	// There should be still one session left in idle list due to the min open sessions constraint.
	if sp.idleList.Len() != 1 {
		t.Errorf("got %v sessions in idle list, want 1", sp.idleList.Len())
	}
}

// TestMaxBurst tests max burst constraint.
func TestMaxBurst(t *testing.T) {
	if testing.Short() {
		t.SkipNow()
	}
	sp, sc, cancel := setup(t, SessionPoolConfig{MaxBurst: 1})
	defer cancel()
	// Will cause session creation RPC to be retried forever.
	sc.InjectError("CreateSession", grpc.Errorf(codes.Unavailable, "try later"))
	// This session request will never finish until the injected error is cleared.
	go sp.take(context.Background())
	// Poll for the execution of the first session request.
	for {
		sp.mu.Lock()
		cr := sp.createReqs
		sp.mu.Unlock()
		if cr == 0 {
			<-time.After(time.Second)
			continue
		}
		// The first session request is being executed.
		break
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	sh, gotErr := sp.take(ctx)
	// Since MaxBurst == 1, the second session request should block.
	if wantErr := errGetSessionTimeout(); !reflect.DeepEqual(gotErr, wantErr) {
		t.Errorf("session retrival returns error %v, want %v", gotErr, wantErr)
	}
	// Let the first session request succeed.
	sc.InjectError("CreateSession", nil)
	// Now new session request can proceed because the first session request will eventually succeed.
	sh, err := sp.take(context.Background())
	if err != nil {
		t.Errorf("session retrival returns error %v, want nil", err)
	}
	if !sh.session.isValid() || sh.getID() == "" {
		t.Errorf("got invalid session: %v", sh.session)
	}
}

// TestSessionrecycle tests recycling sessions.
func TestSessionRecycle(t *testing.T) {
	if testing.Short() {
		t.SkipNow()
	}
	sp, _, cancel := setup(t, SessionPoolConfig{MaxSessionAge: 100 * time.Millisecond, MinOpened: 1})
	// Healthcheck is explicitly turned off in this test because it might aggressively expire sessions in idle list.
	sp.hc.close()
	defer cancel()
	var ss []*session
	shs := make([]*sessionHandle, 2)
	for i := 0; i < len(shs); i++ {
		var err error
		shs[i], err = sp.take(context.Background())
		if err != nil {
			t.Errorf("cannot get the session %v: %v", i, err)
		}
		ss = append(ss, shs[i].session)
	}
	// recycle the first session immediately.
	shs[0].recycle()
	// Let the second session expire.
	<-time.After(time.Second)
	// recycle the second session.
	shs[1].recycle()
	// Now the first session should be still valid, but the second session should have been destroyed.
	if !ss[0].isValid() {
		t.Errorf("the first session (%v) is invalid, want it to be valid", ss[0])
	}
	if ss[1].isValid() {
		t.Errorf("the second session (%v) is valid, want it to be invalid", ss[1])
	}
}

// TestSessionDestroy tests destroying sessions.
func TestSessionDestroy(t *testing.T) {
	sp, _, cancel := setup(t, SessionPoolConfig{MinOpened: 1})
	defer cancel()
	sh, err := sp.take(context.Background())
	if err != nil {
		t.Errorf("cannot get session from session pool: %v", err)
	}
	s := sh.session
	sh.recycle()
	if d := s.destroy(true); d || !s.isValid() {
		// Session should be remaining because of min open sessions constraint.
		t.Errorf("session %v was destroyed in expiration mode, want it to stay alive", s)
	}
	if d := s.destroy(false); !d || s.isValid() {
		// Session should be destroyed.
		t.Errorf("failed to destroy session %s", s)
	}
}

// TestHcHeap tests heap operation on top of hcHeap.
func TestHcHeap(t *testing.T) {
	in := []*session{
		&session{nextCheck: time.Unix(10, 0)},
		&session{nextCheck: time.Unix(0, 5)},
		&session{nextCheck: time.Unix(1, 8)},
		&session{nextCheck: time.Unix(11, 7)},
		&session{nextCheck: time.Unix(6, 3)},
	}
	want := []*session{
		&session{nextCheck: time.Unix(1, 8), hcIndex: 0},
		&session{nextCheck: time.Unix(6, 3), hcIndex: 1},
		&session{nextCheck: time.Unix(8, 2), hcIndex: 2},
		&session{nextCheck: time.Unix(10, 0), hcIndex: 3},
		&session{nextCheck: time.Unix(11, 7), hcIndex: 4},
	}
	hh := hcHeap{}
	for _, s := range in {
		heap.Push(&hh, s)
	}
	// Change top of the heap and do a adjustment.
	hh.sessions[0].nextCheck = time.Unix(8, 2)
	heap.Fix(&hh, 0)
	for idx := 0; hh.Len() > 0; idx++ {
		got := heap.Pop(&hh).(*session)
		want[idx].hcIndex = -1
		if !reflect.DeepEqual(got, want[idx]) {
			t.Errorf("%v: heap.Pop returns %v, want %v", idx, got, want[idx])
		}
	}
}

// TestHealthCheckScheduler tests if healthcheck workers can schedule and perform healthchecks properly.
func TestHealthCheckScheduler(t *testing.T) {
	if testing.Short() {
		t.SkipNow()
	}
	sp, sc, cancel := setup(t, SessionPoolConfig{})
	defer cancel()
	// Create 50 sessions.
	ss := []string{}
	for i := 0; i < 50; i++ {
		sh, err := sp.take(context.Background())
		if err != nil {
			t.Errorf("cannot get session from session pool: %v", err)
		}
		ss = append(ss, sh.getID())
	}
	// Sleep for 1s, allowing healthcheck workers to perform some session pings.
	<-time.After(time.Second)
	dp := sc.DumpPings()
	gotPings := map[string]int64{}
	for _, p := range dp {
		gotPings[p]++
	}
	for _, s := range ss {
		// The average ping interval is 50ms.
		want := int64(time.Second) / int64(50*time.Millisecond)
		if got := gotPings[s]; got < want/2 || got > want+want/2 {
			t.Errorf("got %v healthchecks on session %v, want it between (%v, %v)", got, s, want/2, want+want/2)
		}
	}
}

// Tests that a fractions of sessions are prepared for write by health checker.
func TestWriteSessionsPrepared(t *testing.T) {
	if testing.Short() {
		t.SkipNow()
	}
	sp, sc, cancel := setup(t, SessionPoolConfig{WriteSessions: 0.5})
	sc.MakeNice()
	defer cancel()
	shs := make([]*sessionHandle, 10)
	var err error
	for i := 0; i < 10; i++ {
		shs[i], err = sp.take(context.Background())
		if err != nil {
			t.Errorf("cannot get session from session pool: %v", err)
		}
	}
	// Now there are 10 sessions in the pool. Release them.
	for _, sh := range shs {
		sh.recycle()
	}
	// Sleep for 1s, allowing healthcheck workers to invoke begin transaction.
	<-time.After(time.Second)
	wshs := make([]*sessionHandle, 5)
	for i := 0; i < 5; i++ {
		wshs[i], err = sp.takeWriteSession(context.Background())
		if err != nil {
			t.Errorf("cannot get session from session pool: %v", err)
		}
		if wshs[i].getTransactionID() == nil {
			t.Errorf("got nil transaction id from session pool")
		}
	}
	for _, sh := range wshs {
		sh.recycle()
	}
	<-time.After(time.Second)
	// Now force creation of 10 more sessions.
	shs = make([]*sessionHandle, 20)
	for i := 0; i < 20; i++ {
		shs[i], err = sp.take(context.Background())
		if err != nil {
			t.Errorf("cannot get session from session pool: %v", err)
		}
	}
	// Now there are 20 sessions in the pool. Release them.
	for _, sh := range shs {
		sh.recycle()
	}
	<-time.After(time.Second)
	if sp.idleWriteList.Len() != 10 {
		t.Errorf("Expect 10 write prepared session, got: %d", sp.idleWriteList.Len())
	}
}

// TestTakeFromWriteQueue tests that sessionPool.take() returns write prepared sessions as well.
func TestTakeFromWriteQueue(t *testing.T) {
	if testing.Short() {
		t.SkipNow()
	}
	sp, sc, cancel := setup(t, SessionPoolConfig{MaxOpened: 1, WriteSessions: 1.0})
	sc.MakeNice()
	defer cancel()
	sh, err := sp.take(context.Background())
	if err != nil {
		t.Errorf("cannot get session from session pool: %v", err)
	}
	sh.recycle()
	<-time.After(time.Second)
	// The session should now be in write queue but take should also return it.
	if sp.idleWriteList.Len() == 0 {
		t.Errorf("write queue unexpectedly empty")
	}
	if sp.idleList.Len() != 0 {
		t.Errorf("read queue not empty")
	}
	sh, err = sp.take(context.Background())
	if err != nil {
		t.Errorf("cannot get session from session pool: %v", err)
	}
	sh.recycle()
}

// TestSessionHealthCheck tests healthchecking cases.
func TestSessionHealthCheck(t *testing.T) {
	if testing.Short() {
		t.SkipNow()
	}
	sp, sc, cancel := setup(t, SessionPoolConfig{MaxSessionAge: 2 * time.Second})
	defer cancel()
	// Test pinging sessions.
	sh, err := sp.take(context.Background())
	if err != nil {
		t.Errorf("cannot get session from session pool: %v", err)
	}
	<-time.After(time.Second)
	pings := sc.DumpPings()
	if len(pings) == 0 || pings[0] != sh.getID() {
		t.Errorf("healthchecker didn't send any ping to session %v", sh.getID())
	}
	// Test expiring sessions.
	s := sh.session
	sh.recycle()
	// Sleep enough long for session in idle list to expire.
	<-time.After(2 * time.Second)
	if s.isValid() {
		t.Errorf("session(%v) is still alive, want it to expire", s)
	}
	// Test broken session detection.
	sh, err = sp.take(context.Background())
	if err != nil {
		t.Errorf("cannot get session from session pool: %v", err)
	}
	sc.InjectError("GetSession", grpc.Errorf(codes.NotFound, "Session not found:"))
	// Wait for healthcheck workers to find the broken session and tear it down.
	<-time.After(1 * time.Second)
	if sh.session.isValid() {
		t.Errorf("session(%v) is still alive, want it to be dropped by healthcheck workers", s)
	}
	sc.InjectError("GetSession", nil)
	// Test garbage collection.
	sh, err = sp.take(context.Background())
	if err != nil {
		t.Errorf("cannot get session from session pool: %v", err)
	}
	sp.close()
	if sh.session.isValid() {
		t.Errorf("session(%v) is still alive, want it to be garbage collected", s)
	}
	// Test session id refresh.
	// Recreate the session pool with min open sessions constraint.
	sp, err = newSessionPool("mockdb", SessionPoolConfig{
		MaxSessionAge: time.Second,
		MinOpened:     1,
		getRPCClient: func() (sppb.SpannerClient, error) {
			return sc, nil
		},
		HealthCheckInterval: 50 * time.Millisecond,
	}, nil)
	sh, err = sp.take(context.Background())
	if err != nil {
		t.Errorf("cannot get session from session pool: %v", err)
	}
	oid := sh.getID()
	s = sh.session
	sh.recycle()
	<-time.After(2 * time.Second)
	nid := s.getID()
	if nid == "" || nid == oid {
		t.Errorf("healthcheck workers failed to refresh session: oid=%v, nid=%v", oid, nid)
	}
	if gotDs, wantDs := sc.DumpSessions(), (map[string]bool{nid: true}); !reflect.DeepEqual(gotDs, wantDs) {
		t.Errorf("sessions in mockclient: %v, want %v", gotDs, wantDs)
	}
}

// TestStressSessionPool does stress test on session pool by the following concurrent operations:
//	1) Test worker gets a session from the pool.
//	2) Test worker turns a session back into the pool.
//	3) Test worker destroys a session got from the pool.
//	4) Healthcheck retires an old session from the pool's idlelist by refreshing its session id.
//	5) Healthcheck destroys a broken session (because a worker has already destroyed it).
//	6) Test worker closes the session pool.
//
// During the test, it is expected that all sessions that are taken from session pool remains valid and
// when all test workers and healthcheck workers exit, mockclient, session pool and healthchecker should be in consistent state.
func TestStressSessionPool(t *testing.T) {
	// Use concurrent workers to test different session pool built from different configurations.
	if testing.Short() {
		t.SkipNow()
	}
	for ti, cfg := range []SessionPoolConfig{
		SessionPoolConfig{},
		SessionPoolConfig{MaxSessionAge: 20 * time.Millisecond},
		SessionPoolConfig{MinOpened: 10, MaxOpened: 100},
		SessionPoolConfig{MaxBurst: 50},
		SessionPoolConfig{MaxSessionAge: 20 * time.Millisecond, MinOpened: 10, MaxOpened: 200, MaxBurst: 5},
		SessionPoolConfig{MaxSessionAge: 20 * time.Millisecond, MinOpened: 10, MaxOpened: 200, MaxBurst: 5, WriteSessions: 0.2},
	} {
		var wg sync.WaitGroup
		// Create a more aggressive session healthchecker to increase test concurrency.
		cfg.HealthCheckInterval = 50 * time.Millisecond
		cfg.HealthCheckWorkers = 50
		sc := testutil.NewMockCloudSpannerClient(t)
		sc.MakeNice()
		cfg.getRPCClient = func() (sppb.SpannerClient, error) {
			return sc, nil
		}
		sp, _ := newSessionPool("mockdb", cfg, nil)
		for i := 0; i < 100; i++ {
			wg.Add(1)
			// Schedule a test worker.
			go func(idx int, pool *sessionPool, client sppb.SpannerClient) {
				defer wg.Done()
				// Test worker iterates 1K times and tries different session / session pool operations.
				for j := 0; j < 1000; j++ {
					if idx%10 == 0 && j >= 900 {
						// Close the pool in selected set of workers during the middle of the test.
						pool.close()
					}
					// Take a write sessions ~ 20% of the times.
					takeWrite := rand.Intn(5) == 4
					var (
						sh     *sessionHandle
						gotErr error
					)
					if takeWrite {
						sh, gotErr = pool.takeWriteSession(context.Background())
					} else {
						sh, gotErr = pool.take(context.Background())
					}
					if gotErr != nil {
						if pool.isValid() {
							t.Errorf("%v.%v: pool.take returns error when pool is still valid: %v", ti, idx, gotErr)
						}
						if wantErr := errInvalidSessionPool(); !reflect.DeepEqual(gotErr, wantErr) {
							t.Errorf("%v.%v: got error when pool is closed: %v, want %v", ti, idx, gotErr, wantErr)
						}
						continue
					}
					// Verify if session is valid when session pool is valid. Note that if session pool is invalid after sh is taken,
					// then sh might be invalidated by healthcheck workers.
					if (sh.getID() == "" || sh.session == nil || !sh.session.isValid()) && pool.isValid() {
						t.Errorf("%v.%v.%v: pool.take returns invalid session %v", ti, idx, takeWrite, sh.session)
					}
					if takeWrite && sh.getTransactionID() == nil {
						t.Errorf("%v.%v: pool.takeWriteSession returns session %v without transaction", ti, idx, sh.session)
					}
					if int64(cfg.MaxSessionAge) > 0 && rand.Intn(100) < idx {
						// Random sleep before destroying/recycling the session, to give healthcheck worker a chance to step in.
						<-time.After(time.Duration(rand.Int63n(int64(cfg.MaxSessionAge))))
					}
					if rand.Intn(100) < idx {
						// destroy the session.
						sh.destroy()
						continue
					}
					// recycle the session.
					sh.recycle()
				}
			}(i, sp, sc)
		}
		wg.Wait()
		sp.hc.close()
		// Here the states of healthchecker, session pool and mockclient are stable.
		idleSessions := map[string]bool{}
		hcSessions := map[string]bool{}
		mockSessions := sc.DumpSessions()
		// Dump session pool's idle list.
		for sl := sp.idleList.Front(); sl != nil; sl = sl.Next() {
			s := sl.Value.(*session)
			if idleSessions[s.getID()] {
				t.Errorf("%v: found duplicated session in idle list: %v", ti, s.getID())
			}
			idleSessions[s.getID()] = true
		}
		for sl := sp.idleWriteList.Front(); sl != nil; sl = sl.Next() {
			s := sl.Value.(*session)
			if idleSessions[s.getID()] {
				t.Errorf("%v: found duplicated session in idle write list: %v", ti, s.getID())
			}
			idleSessions[s.getID()] = true
		}
		if int(sp.numOpened) != len(idleSessions) {
			t.Errorf("%v: number of opened sessions (%v) != number of idle sessions (%v)", ti, sp.numOpened, len(idleSessions))
		}
		if sp.createReqs != 0 {
			t.Errorf("%v: number of pending session creations = %v, want 0", ti, sp.createReqs)
		}
		// Dump healthcheck queue.
		for _, s := range sp.hc.queue.sessions {
			if hcSessions[s.getID()] {
				t.Errorf("%v: found duplicated session in healthcheck queue: %v", ti, s.getID())
			}
			hcSessions[s.getID()] = true
		}
		// Verify that idleSessions == hcSessions == mockSessions.
		if !reflect.DeepEqual(idleSessions, hcSessions) {
			t.Errorf("%v: sessions in idle list (%v) != sessions in healthcheck queue (%v)", ti, idleSessions, hcSessions)
		}
		if !reflect.DeepEqual(hcSessions, mockSessions) {
			t.Errorf("%v: sessions in healthcheck queue (%v) != sessions in mockclient (%v)", ti, hcSessions, mockSessions)
		}
		sp.close()
		mockSessions = sc.DumpSessions()
		if len(mockSessions) != 0 {
			t.Errorf("Found live sessions: %v", mockSessions)
		}
	}
}
