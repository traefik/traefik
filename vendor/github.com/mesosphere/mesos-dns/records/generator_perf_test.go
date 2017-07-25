package records

import (
	"math/rand"
	"strconv"
	"testing"

	"github.com/mesosphere/mesos-dns/records/labels"
	"github.com/mesosphere/mesos-dns/records/state"
)

// BenchmarkInsertRR *only* tests insertRR, not the taskRecord funcs.
func BenchmarkInsertRR(b *testing.B) {
	const (
		clusterSize = 1000
		appCount    = 5
	)
	var (
		slaves = make([]string, clusterSize)
		apps   = make([]string, appCount)
		rg     = &RecordGenerator{
			As:   rrs{},
			SRVs: rrs{},
		}
	)
	for i := 0; i < clusterSize; i++ {
		slaves[i] = "slave" + strconv.Itoa(i)
	}
	for i := 0; i < appCount; i++ {
		apps[i] = "app" + strconv.Itoa(i)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var (
			si = rand.Int31n(clusterSize)
			ai = rand.Int31n(appCount)
		)
		rg.insertRR(apps[ai], slaves[si], "A")
	}
}

func BenchmarkTaskRecord_withoutDiscoveryInfo(b *testing.B) {
	const (
		clusterSize = 1000
		taskCount   = 1000
	)
	type params struct {
		task      state.Task
		f         state.Framework
		domain    string
		spec      labels.Func
		ipSources []string
		enumFW    EnumerableFramework
		rg        RecordGenerator
	}
	var (
		initialState = params{
			task: state.Task{
				State: "TASK_RUNNING",
			},
			f:         state.Framework{Name: "foo"},
			domain:    "mesos",
			spec:      labels.RFC1123,
			ipSources: []string{"host"},
			rg: RecordGenerator{
				As:   rrs{},
				SRVs: rrs{},
			},
		}
		slaves = make([]string, clusterSize)
		tasks  = make([]string, taskCount)
	)
	for i := 0; i < clusterSize; i++ {
		slaves[i] = "slave" + strconv.Itoa(i)
	}
	for i := 0; i < taskCount; i++ {
		tasks[i] = "task" + strconv.Itoa(i)
	}
	tt := initialState
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var (
			si = rand.Int31n(clusterSize)
			ti = rand.Int31n(taskCount)
		)
		tt.task.Name = tasks[ti]
		tt.task.SlaveIP = slaves[si]
		tt.task.SlaveID = "ID-" + slaves[si]
		tt.rg.taskRecord(tt.task, tt.f, tt.domain, tt.spec, tt.ipSources, &tt.enumFW)
	}
}
