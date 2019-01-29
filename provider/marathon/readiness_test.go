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
	tsk := task()
	app := application(
		deployments("deploymentId"),
		readinessCheck(0),
		readinessCheckResult(testTaskName, false),
	)

	if ready := rc.Do(tsk, app); !ready {
		t.Error("expected ready = true")
	}
}

func TestEnabledReadinessChecker(t *testing.T) {
	tests := []struct {
		desc          string
		task          marathon.Task
		app           marathon.Application
		rc            readinessChecker
		expectedReady bool
	}{
		{
			desc:          "no deployment running",
			task:          task(),
			app:           application(),
			expectedReady: true,
		},
		{
			desc:          "no readiness checks defined",
			task:          task(),
			app:           application(deployments("deploymentId")),
			expectedReady: true,
		},
		{
			desc: "readiness check result negative",
			task: task(),
			app: application(
				deployments("deploymentId"),
				readinessCheck(0),
				readinessCheckResult("otherTaskID", true),
				readinessCheckResult(testTaskName, false),
			),
			expectedReady: false,
		},
		{
			desc: "readiness check result positive",
			task: task(),
			app: application(
				deployments("deploymentId"),
				readinessCheck(0),
				readinessCheckResult("otherTaskID", false),
				readinessCheckResult(testTaskName, true),
			),
			expectedReady: true,
		},
		{
			desc: "no readiness check result with default timeout",
			task: task(startedAtFromNow(3 * time.Minute)),
			app: application(
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
			task: task(startedAtFromNow(4 * time.Minute)),
			app: application(
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
			task: task(startedAt("invalid")),
			app: application(
				deployments("deploymentId"),
				readinessCheck(0),
			),
			expectedReady: false,
		},
		{
			desc: "task not involved in deployment",
			task: task(startedAtFromNow(1 * time.Hour)),
			app: application(
				deployments("deploymentId"),
				readinessCheck(0),
			),
			rc: readinessChecker{
				checkDefaultTimeout: 10 * time.Second,
			},
			expectedReady: true,
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()
			rc := testReadinessChecker()
			if test.rc.checkDefaultTimeout > 0 {
				rc.checkDefaultTimeout = test.rc.checkDefaultTimeout
			}
			if test.rc.checkSafetyMargin > 0 {
				rc.checkSafetyMargin = test.rc.checkSafetyMargin
			}
			actualReady := test.rc.Do(test.task, test.app)
			if actualReady != test.expectedReady {
				t.Errorf("actual ready = %t, expected ready = %t", actualReady, test.expectedReady)
			}
		})
	}
}
