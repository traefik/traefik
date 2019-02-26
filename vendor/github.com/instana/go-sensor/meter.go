package instana

import (
	"runtime"
	"strconv"
	"time"
)

const (
	// SnapshotPeriod is the amount of time in seconds between snapshot reports.
	SnapshotPeriod = 600
)

// SnapshotS struct to hold snapshot data.
type SnapshotS struct {
	Name     string `json:"name"`
	Version  string `json:"version"`
	Root     string `json:"goroot"`
	MaxProcs int    `json:"maxprocs"`
	Compiler string `json:"compiler"`
	NumCPU   int    `json:"cpu"`
}

// MemoryS struct to hold snapshot data.
type MemoryS struct {
	Alloc         uint64  `json:"alloc"`
	TotalAlloc    uint64  `json:"total_alloc"`
	Sys           uint64  `json:"sys"`
	Lookups       uint64  `json:"lookups"`
	Mallocs       uint64  `json:"mallocs"`
	Frees         uint64  `json:"frees"`
	HeapAlloc     uint64  `json:"heap_alloc"`
	HeapSys       uint64  `json:"heap_sys"`
	HeapIdle      uint64  `json:"heap_idle"`
	HeapInuse     uint64  `json:"heap_in_use"`
	HeapReleased  uint64  `json:"heap_released"`
	HeapObjects   uint64  `json:"heap_objects"`
	PauseTotalNs  uint64  `json:"pause_total_ns"`
	PauseNs       uint64  `json:"pause_ns"`
	NumGC         uint32  `json:"num_gc"`
	GCCPUFraction float64 `json:"gc_cpu_fraction"`
}

// MetricsS struct to hold snapshot data.
type MetricsS struct {
	CgoCall   int64    `json:"cgo_call"`
	Goroutine int      `json:"goroutine"`
	Memory    *MemoryS `json:"memory"`
}

// EntityData struct to hold snapshot data.
type EntityData struct {
	PID      int        `json:"pid"`
	Snapshot *SnapshotS `json:"snapshot,omitempty"`
	Metrics  *MetricsS  `json:"metrics"`
}

type meterS struct {
	sensor            *sensorS
	numGC             uint32
	ticker            *time.Ticker
	snapshotCountdown int
}

func (r *meterS) init() {
	r.ticker = time.NewTicker(1 * time.Second)
	go func() {
		r.snapshotCountdown = 1
		for range r.ticker.C {
			if r.sensor.agent.canSend() {
				r.snapshotCountdown--
				var s *SnapshotS
				if r.snapshotCountdown == 0 {
					r.snapshotCountdown = SnapshotPeriod
					s = r.collectSnapshot()
					log.debug("collected snapshot")
				} else {
					s = nil
				}

				pid, _ := strconv.Atoi(r.sensor.agent.from.PID)
				d := &EntityData{
					PID:      pid,
					Snapshot: s,
					Metrics:  r.collectMetrics()}

				go r.send(d)
			}
		}
	}()
}

func (r *meterS) send(d *EntityData) {
	_, err := r.sensor.agent.request(r.sensor.agent.makeURL(agentDataURL), "POST", d)

	if err != nil {
		r.sensor.agent.reset()
	}
}

func (r *meterS) collectMemoryMetrics() *MemoryS {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)
	ret := &MemoryS{
		Alloc:         memStats.Alloc,
		TotalAlloc:    memStats.TotalAlloc,
		Sys:           memStats.Sys,
		Lookups:       memStats.Lookups,
		Mallocs:       memStats.Mallocs,
		Frees:         memStats.Frees,
		HeapAlloc:     memStats.HeapAlloc,
		HeapSys:       memStats.HeapSys,
		HeapIdle:      memStats.HeapIdle,
		HeapInuse:     memStats.HeapInuse,
		HeapReleased:  memStats.HeapReleased,
		HeapObjects:   memStats.HeapObjects,
		PauseTotalNs:  memStats.PauseTotalNs,
		NumGC:         memStats.NumGC,
		GCCPUFraction: memStats.GCCPUFraction}

	if r.numGC < memStats.NumGC {
		ret.PauseNs = memStats.PauseNs[(memStats.NumGC+255)%256]
		r.numGC = memStats.NumGC
	} else {
		ret.PauseNs = 0
	}

	return ret
}

func (r *meterS) collectMetrics() *MetricsS {
	return &MetricsS{
		CgoCall:   runtime.NumCgoCall(),
		Goroutine: runtime.NumGoroutine(),
		Memory:    r.collectMemoryMetrics()}
}

func (r *meterS) collectSnapshot() *SnapshotS {
	return &SnapshotS{
		Name:     r.sensor.serviceName,
		Version:  runtime.Version(),
		Root:     runtime.GOROOT(),
		MaxProcs: runtime.GOMAXPROCS(0),
		Compiler: runtime.Compiler,
		NumCPU:   runtime.NumCPU()}
}

func (r *sensorS) initMeter() *meterS {

	log.debug("initializing meter")

	ret := new(meterS)
	ret.sensor = r
	ret.init()

	return ret
}
