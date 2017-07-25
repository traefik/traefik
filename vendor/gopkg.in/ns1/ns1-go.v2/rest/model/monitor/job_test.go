package monitor

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUnmarshalJobs(t *testing.T) {
	data := []byte(`[
  {
    "id": "52a27d4397d5f07003fdbe7b",
    "config": {
      "host": "1.2.3.4"
    },
    "status": {
      "lga": {
        "since": 1389407609,
        "status": "up"
      },
      "global": {
        "since": 1389407609,
        "status": "up"
      },
      "sjc": {
        "since": 1389404014,
        "status": "up"
      }
    },
    "rules": [
      {
        "key": "rtt",
        "value": 100,
        "comparison": "<"
      }
    ],
    "job_type": "ping",
    "regions": [
      "lga",
      "sjc"
    ],
    "active": true,
    "frequency": 60,
    "policy": "quorum",
    "region_scope": "fixed"
  }
]`)
	mjl := []*Job{}
	if err := json.Unmarshal(data, &mjl); err != nil {
		t.Error(err)
	}
	if len(mjl) != 1 {
		fmt.Println(mjl)
		t.Error("Do not have any jobs")
	}
	j := mjl[0]
	if j.ID != "52a27d4397d5f07003fdbe7b" {
		t.Error("Wrong ID")
	}
	conf := j.Config
	if conf["host"] != "1.2.3.4" {
		t.Error("Wrong host")
	}

	if j.Status["global"].Since != 1389407609 {
		t.Error("since has unexpected value")
	}
	if j.Status["global"].Status != "up" {
		t.Error("Status is not up")
	}

	if j.Status["sjc"].Since != 1389404014 {
		t.Error("sjc since has unexpected value")
	}
	if j.Status["sjc"].Status != "up" {
		t.Error("sjc Status is not up")
	}

	r := j.Rules[0]
	assert.Equal(t, r.Key, "rtt", "RTT rule key is wrong")
	assert.Equal(t, r.Value.(float64), float64(100), "RTT rule value is wrong")
	if r.Comparison != "<" {
		t.Error("RTT rule comparison is wrong")
	}
	if j.Type != "ping" {
		t.Error("Jobtype is wrong")
	}
	if j.Regions[0] != "lga" {
		t.Error("First region is not lga")
	}
	if !j.Active {
		t.Error("Job is not active")
	}
	if j.Frequency != 60 {
		t.Error("Job frequency != 60")
	}
	if j.Policy != "quorum" {
		t.Error("Job policy is not quorum")
	}
	if j.RegionScope != "fixed" {
		t.Error("Job region scope is not fixed")
	}
}

func TestUnmarshalStatusLog(t *testing.T) {
	data := []byte(`{
    "status": "down",
    "region": "lga",
    "since": 1488297041,
    "job": "58b364c09825e00001e2af80",
    "until": 1488297042
  }`)
	log := &StatusLog{}
	if err := json.Unmarshal(data, &log); err != nil {
		t.Error(err)
	}
	if log.Job != "58b364c09825e00001e2af80" {
		t.Error("Wrong job")
	}
	if log.Status != "down" {
		t.Error("Wrong status")
	}
	if log.Region != "lga" {
		t.Error("Wrong region")
	}
	if log.Since != 1488297041 {
		t.Error("Wrong since")
	}
	if log.Until != 1488297042 {
		t.Error("Wrong until")
	}
}

func TestUnmarshalStatusLogMostRecent(t *testing.T) {
	data := []byte(`{
    "status": "down",
    "region": "lga",
    "since": 1488297041,
    "job": "58b364c09825e00001e2af80",
    "until": null
  }`)
	log := &StatusLog{}
	if err := json.Unmarshal(data, &log); err != nil {
		t.Error(err)
	}
	if log.Job != "58b364c09825e00001e2af80" {
		t.Error("Wrong job")
	}
	if log.Status != "down" {
		t.Error("Wrong status")
	}
	if log.Region != "lga" {
		t.Error("Wrong region")
	}
	if log.Since != 1488297041 {
		t.Error("Wrong since")
	}
	if log.Until != 0 {
		t.Error("Wrong until")
	}
}

func TestUnmarshalStatusLogs(t *testing.T) {
	data := []byte(`[
  {
    "status": "up",
    "region": "satellite",
    "since": 1488297044,
    "job": "58b364c09825e00001e2af80",
    "until": null
  },
  {
    "status": "up",
    "region": "master",
    "since": 1488297044,
    "job": "58b364c09825e00001e2af80",
    "until": null
  },
  {
    "status": "up",
    "region": "global",
    "since": 1488297044,
    "job": "58b364c09825e00001e2af80",
    "until": null
  },
  {
    "status": "down",
    "region": "satellite",
    "since": 1488297043,
    "job": "58b364c09825e00001e2af80",
    "until": 1488297044
  },
  {
    "status": "down",
    "region": "master",
    "since": 1488297043,
    "job": "58b364c09825e00001e2af80",
    "until": 1488297044
  },
  {
    "status": "down",
    "region": "global",
    "since": 1488297043,
    "job": "58b364c09825e00001e2af80",
    "until": 1488297044
  },
  {
    "status": "up",
    "region": "satellite",
    "since": 1488297042,
    "job": "58b364c09825e00001e2af80",
    "until": 1488297043
  },
  {
    "status": "up",
    "region": "master",
    "since": 1488297042,
    "job": "58b364c09825e00001e2af80",
    "until": 1488297043
  },
  {
    "status": "up",
    "region": "global",
    "since": 1488297042,
    "job": "58b364c09825e00001e2af80",
    "until": 1488297043
  }
]`)

	logs := []*StatusLog{}
	if err := json.Unmarshal(data, &logs); err != nil {
		t.Error(err)
	}
	if len(logs) != 9 {
		t.Errorf("Do not have correct number of status logs in job history. Expected: %d, Actual: %d", 9, len(logs))
	}
}
