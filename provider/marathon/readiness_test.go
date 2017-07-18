package marathon

import (
	"testing"
	"time"

	"github.com/gambol99/go-marathon"
)

func testReadinessChecker() *readinessChecker {
	return defaultReadinessChecker(false)
}

func TestDisabledReadinessChecker(t *testing.T) {
	var rc *readinessChecker
	task := createTask()
	app := createApplication(
		deployments("deploymentId"),
		readinessCheck(0),
		readinessCheckResult(testTaskName, false),
	)

	if ready := rc.Do(task, app); ready == false {
		t.Error("expected ready = true")
	}
}

func TestEnabledReadinessChecker(t *testing.T) {
	cases := []struct {
		desc          string
		task          marathon.Task
		app           marathon.Application
		rc            readinessChecker
		expectedReady bool
	}{
		{
			desc:          "no deployment running",
			task:          createTask(),
			app:           createApplication(),
			expectedReady: true,
		},
		{
			desc:          "no readiness checks defined",
			task:          createTask(),
			app:           createApplication(deployments("deploymentId")),
			expectedReady: true,
		},
		{
			desc: "readiness check result negative",
			task: createTask(),
			app: createApplication(
				deployments("deploymentId"),
				readinessCheck(0),
				readinessCheckResult("otherTaskID", true),
				readinessCheckResult(testTaskName, false),
			),
			expectedReady: false,
		},
		{
			desc: "readiness check result positive",
			task: createTask(),
			app: createApplication(
				deployments("deploymentId"),
				readinessCheck(0),
				readinessCheckResult("otherTaskID", false),
				readinessCheckResult(testTaskName, true),
			),
			expectedReady: true,
		},
		{
			desc: "no readiness check result with default timeout",
			task: createTask(startedAtFromNow(3 * time.Minute)),
			app: createApplication(
				deployments("deploymentId"),
				readinessCheck(0),
			),
			rc: readinessChecker{
				checkDefaultTimeout: 5 * time.Minute,
			},
			expectedReady: false,
		},
		{
			desc: "no readiness check result with readiness check timeout",
			task: createTask(startedAtFromNow(4 * time.Minute)),
			app: createApplication(
				deployments("deploymentId"),
				readinessCheck(3*time.Minute),
			),
			rc: readinessChecker{
				checkSafetyMargin: 3 * time.Minute,
			},
			expectedReady: false,
		},
		{
			desc: "invalid task start time",
			task: createTask(startedAt("invalid")),
			app: createApplication(
				deployments("deploymentId"),
				readinessCheck(0),
			),
			expectedReady: false,
		},
		{
			desc: "task not involved in deployment",
			task: createTask(startedAtFromNow(1 * time.Hour)),
			app: createApplication(
				deployments("deploymentId"),
				readinessCheck(0),
			),
			rc: readinessChecker{
				checkDefaultTimeout: 10 * time.Second,
			},
			expectedReady: true,
		},
	}

	for _, c := range cases {
		c := c
		t.Run(c.desc, func(t *testing.T) {
			t.Parallel()
			rc := testReadinessChecker()
			if c.rc.checkDefaultTimeout > 0 {
				rc.checkDefaultTimeout = c.rc.checkDefaultTimeout
			}
			if c.rc.checkSafetyMargin > 0 {
				rc.checkSafetyMargin = c.rc.checkSafetyMargin
			}
			actualReady := c.rc.Do(c.task, c.app)
			if actualReady != c.expectedReady {
				t.Errorf("actual ready = %t, expected ready = %t", actualReady, c.expectedReady)
			}
		})
	}
}
