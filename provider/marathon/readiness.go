package marathon

import (
	"time"

	"github.com/containous/traefik/log"
	"github.com/gambol99/go-marathon"
)

const (
	// readinessCheckDefaultTimeout is the default timeout for a readiness
	// check if no check timeout is specified on the application spec. This
	// should really never be the case, but better be safe than sorry.
	readinessCheckDefaultTimeout = 10 * time.Second
	// readinessCheckSafetyMargin is some buffer duration to account for
	// small offsets in readiness check execution.
	readinessCheckSafetyMargin = 5 * time.Second
	readinessLogHeader         = "Marathon readiness check: "
)

type readinessChecker struct {
	checkDefaultTimeout time.Duration
	checkSafetyMargin   time.Duration
	traceLogging        bool
}

func defaultReadinessChecker(isTraceLogging bool) *readinessChecker {
	return &readinessChecker{
		checkDefaultTimeout: readinessCheckDefaultTimeout,
		checkSafetyMargin:   readinessCheckSafetyMargin,
		traceLogging:        isTraceLogging,
	}
}

func (rc *readinessChecker) Do(task marathon.Task, app marathon.Application) bool {
	if rc == nil {
		// Readiness checker disabled.
		return true
	}

	switch {
	case len(app.Deployments) == 0:
		// We only care about readiness during deployments; post-deployment readiness
		// can be covered by a periodic post-deployment probe (i.e., Traefik health checks).
		rc.tracef("task %s app %s: ready = true [no deployment ongoing]", task.ID, app.ID)
		return true

	case app.ReadinessChecks == nil || len(*app.ReadinessChecks) == 0:
		// Applications without configured readiness checks are always considered
		// ready.
		rc.tracef("task %s app %s: ready = true [no readiness checks on app]", task.ID, app.ID)
		return true
	}

	// Loop through all readiness check results and return the results for
	// matching task IDs.
	if app.ReadinessCheckResults != nil {
		for _, readinessCheckResult := range *app.ReadinessCheckResults {
			if readinessCheckResult.TaskID == task.ID {
				rc.tracef("task %s app %s: ready = %t [evaluating readiness check ready state]", task.ID, app.ID, readinessCheckResult.Ready)
				return readinessCheckResult.Ready
			}
		}
	}

	// There's a corner case sometimes hit where the first new task of a
	// deployment goes from TASK_STAGING to TASK_RUNNING without a corresponding
	// readiness check result being included in the API response. This only happens
	// in a very short (yet unlucky) time frame and does not repeat for subsequent
	// tasks of the same deployment.
	// Complicating matters, the situation may occur for both initially deploying
	// applications as well as rolling-upgraded ones where one or more tasks from
	// a previous deployment exist already and are joined by new tasks from a
	// subsequent deployment. We must always make sure that pre-existing tasks
	// maintain their ready state while newly launched tasks must be considered
	// unready until a check result appears.
	// We distinguish the two cases by comparing the current time with the start
	// time of the task: It should take Marathon at most one readiness check timeout
	// interval (plus some safety margin to account for the delayed nature of
	// distributed systems) for readiness check results to be returned along the API
	// response. Once the task turns old enough, we assume it to be part of a
	// pre-existing deployment and mark it as ready. Note that it is okay to err
	// on the side of caution and consider a task unready until the safety time
	// window has elapsed because a newly created task should be readiness-checked
	// and be given a result fairly shortly after its creation (i.e., on the scale
	// of seconds).
	readinessCheckTimeoutSecs := (*app.ReadinessChecks)[0].TimeoutSeconds
	readinessCheckTimeout := time.Duration(readinessCheckTimeoutSecs) * time.Second
	if readinessCheckTimeout == 0 {
		rc.tracef("task %s app %s: readiness check timeout not set, using default value %s", task.ID, app.ID, rc.checkDefaultTimeout)
		readinessCheckTimeout = rc.checkDefaultTimeout
	} else {
		readinessCheckTimeout += rc.checkSafetyMargin
	}

	startTime, err := time.Parse(time.RFC3339, task.StartedAt)
	if err != nil {
		// An unparseable start time should never occur; if it does, we assume the
		// problem should be surfaced as quickly as possible, which is easiest if
		// we shun the task from rotation.
		log.Warnf("Failed to parse start-time %s of task %s from application %s: %s (assuming unready)", task.StartedAt, task.ID, app.ID, err)
		return false
	}

	since := time.Since(startTime)
	if since < readinessCheckTimeout {
		rc.tracef("task %s app %s: ready = false [task with start-time %s not within assumed check timeout window of %s (elapsed time since task start: %s)]", task.ID, app.ID, startTime.Format(time.RFC3339), readinessCheckTimeout, since)
		return false
	}

	// Finally, we can be certain this task is not part of the deployment (i.e.,
	// it's an old task that's going to transition into the TASK_KILLING and/or
	// TASK_KILLED state as new tasks' readiness checks gradually turn green.)
	rc.tracef("task %s app %s: ready = true [task with start-time %s not involved in deployment (elapsed time since task start: %s)]", task.ID, app.ID, startTime.Format(time.RFC3339), since)
	return true
}

func (rc *readinessChecker) tracef(format string, args ...interface{}) {
	if rc.traceLogging {
		log.Debugf(readinessLogHeader+format, args...)
	}
}
