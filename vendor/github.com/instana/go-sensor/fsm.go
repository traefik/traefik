package instana

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"time"

	f "github.com/looplab/fsm"
)

const (
	eInit     = "init"
	eLookup   = "lookup"
	eAnnounce = "announce"
	eTest     = "test"

	retryPeriod    = 30 * 1000
	maximumRetries = 2
)

type fsmS struct {
	agent   *agentS
	fsm     *f.FSM
	timer   *time.Timer
	retries int
}

func (r *fsmS) init() {

	log.warn("Stan is on the scene.  Starting Instana instrumentation.")
	log.debug("initializing fsm")

	r.fsm = f.NewFSM(
		"none",
		f.Events{
			{Name: eInit, Src: []string{"none", "unannounced", "announced", "ready"}, Dst: "init"},
			{Name: eLookup, Src: []string{"init"}, Dst: "unannounced"},
			{Name: eAnnounce, Src: []string{"unannounced"}, Dst: "announced"},
			{Name: eTest, Src: []string{"announced"}, Dst: "ready"}},
		f.Callbacks{
			"init":              r.lookupAgentHost,
			"enter_unannounced": r.announceSensor,
			"enter_announced":   r.testAgent})

	r.retries = maximumRetries
	r.fsm.Event(eInit)
}

func (r *fsmS) scheduleRetry(e *f.Event, cb func(e *f.Event)) {
	r.timer = time.NewTimer(retryPeriod * time.Millisecond)
	go func() {
		<-r.timer.C
		cb(e)
	}()
}

func (r *fsmS) lookupAgentHost(e *f.Event) {
	cb := func(b bool, host string) {
		if b {
			r.lookupSuccess(host)
		} else {
			gateway := r.getDefaultGateway()
			if gateway != "" {
				go r.checkHost(gateway, func(b bool, host string) {
					if b {
						r.lookupSuccess(host)
					} else {
						log.error("Cannot connect to the agent through localhost or default gateway. Scheduling retry.")
						r.scheduleRetry(e, r.lookupAgentHost)
					}
				})
			} else {
				log.error("Default gateway not available. Scheduling retry")
				r.scheduleRetry(e, r.lookupAgentHost)
			}
		}
	}
	hostNames := []string{
		r.agent.sensor.options.AgentHost,
		os.Getenv("INSTANA_AGENT_HOST"),
		agentDefaultHost,
	}
	for _, name := range hostNames {
		if name != "" {
			go r.checkHost(name, cb)
			return
		}
	}
}

func (r *fsmS) getDefaultGateway() string {
	out, _ := exec.Command("/bin/sh", "-c", "/sbin/ip route | awk '/default/' | cut -d ' ' -f 3 | tr -d '\n'").Output()

	log.debug("checking default gateway", string(out[:]))

	return string(out[:])
}

func (r *fsmS) checkHost(host string, cb func(b bool, host string)) {
	log.debug("checking host", host)

	header, err := r.agent.requestHeader(r.agent.makeHostURL(host, "/"), "GET", "Server")

	cb(err == nil && header == agentHeader, host)
}

func (r *fsmS) lookupSuccess(host string) {
	log.debug("agent lookup success", host)

	r.agent.setHost(host)
	r.retries = maximumRetries
	r.fsm.Event(eLookup)
}

func (r *fsmS) announceSensor(e *f.Event) {
	cb := func(b bool, from *fromS) {
		if b {
			log.info("Host agent available. We're in business. Announced pid:", from.PID)
			r.agent.setFrom(from)
			r.retries = maximumRetries
			r.fsm.Event(eAnnounce)
		} else {
			log.error("Cannot announce sensor. Scheduling retry.")
			r.retries--
			if r.retries > 0 {
				r.scheduleRetry(e, r.announceSensor)
			} else {
				r.fsm.Event(eInit)
			}
		}
	}

	log.debug("announcing sensor to the agent")

	go func(cb func(b bool, from *fromS)) {
		defer func() {
			if r := recover(); r != nil {
				log.debug("Announce recovered:", r)
			}
		}()

		pid := 0
		schedFile := fmt.Sprintf("/proc/%d/sched", os.Getpid())
		if _, err := os.Stat(schedFile); err == nil {
			sf, err := os.Open(schedFile)
			defer sf.Close()
			if err == nil {
				fscanner := bufio.NewScanner(sf)
				fscanner.Scan()
				primaLinea := fscanner.Text()

				r := regexp.MustCompile("\\((\\d+),")
				match := r.FindStringSubmatch(primaLinea)
				i, err := strconv.Atoi(match[1])
				if err == nil {
					pid = i
				}
			}
		}

		if pid == 0 {
			pid = os.Getpid()
		}

		d := &discoveryS{PID: pid}
		d.Name, d.Args = getCommandLine()

		if _, err := os.Stat("/proc"); err == nil {
			if addr, err := net.ResolveTCPAddr("tcp", r.agent.host+":42699"); err == nil {
				if tcpConn, err := net.DialTCP("tcp", nil, addr); err == nil {
					defer tcpConn.Close()

					f, err := tcpConn.File()

					if err != nil {
						log.error(err)
					} else {
						d.Fd = fmt.Sprintf("%v", f.Fd())

						link := fmt.Sprintf("/proc/%d/fd/%d", os.Getpid(), f.Fd())
						if _, err := os.Stat(link); err == nil {
							d.Inode, _ = os.Readlink(link)
						}
					}
				}
			}
		}

		ret := &agentResponse{}
		_, err := r.agent.requestResponse(r.agent.makeURL(agentDiscoveryURL), "PUT", d, ret)
		cb(err == nil,
			&fromS{
				PID:    strconv.Itoa(int(ret.Pid)),
				HostID: ret.HostID})
	}(cb)
}

func (r *fsmS) testAgent(e *f.Event) {
	cb := func(b bool) {
		if b {
			r.retries = maximumRetries
			r.fsm.Event(eTest)
		} else {
			log.debug("Agent is not yet ready. Scheduling retry.")
			r.retries--
			if r.retries > 0 {
				r.scheduleRetry(e, r.testAgent)
			} else {
				r.fsm.Event(eInit)
			}
		}
	}

	log.debug("testing communication with the agent")

	go func(cb func(b bool)) {
		_, err := r.agent.head(r.agent.makeURL(agentDataURL))
		cb(err == nil)
	}(cb)
}

func (r *fsmS) reset() {
	r.retries = maximumRetries
	r.fsm.Event(eInit)
}

func (r *agentS) initFsm() *fsmS {
	ret := new(fsmS)
	ret.agent = r
	ret.init()

	return ret
}

func (r *agentS) canSend() bool {
	return r.fsm.fsm.Current() == "ready"
}
