package storage

import chk "gopkg.in/check.v1"

type StorageSuite struct{}

var _ = chk.Suite(&StorageSuite{})

// This tests use the Table service, but could also use any other service

func (s *StorageSuite) TestGetServiceProperties(c *chk.C) {
	cli := getTableClient(c)
	sp, err := cli.GetServiceProperties()
	c.Assert(err, chk.IsNil)
	c.Assert(sp, chk.NotNil)
}

func (s *StorageSuite) TestSetServiceProperties(c *chk.C) {
	cli := getTableClient(c)

	t := true
	num := 7
	rp := RetentionPolicy{
		Enabled: true,
		Days:    &num,
	}
	m := Metrics{
		Version:         "1.0",
		Enabled:         true,
		IncludeAPIs:     &t,
		RetentionPolicy: &rp,
	}
	spInput := ServiceProperties{
		Logging: &Logging{
			Version:         "1.0",
			Delete:          true,
			Read:            false,
			Write:           true,
			RetentionPolicy: &rp,
		},
		HourMetrics:   &m,
		MinuteMetrics: &m,
		Cors: &Cors{
			CorsRule: []CorsRule{
				{
					AllowedOrigins:  "*",
					AllowedMethods:  "GET,PUT",
					MaxAgeInSeconds: 500,
					ExposedHeaders:  "x-ms-meta-customheader,x-ms-meta-data*",
					AllowedHeaders:  "x-ms-meta-customheader,x-ms-meta-target*",
				},
			},
		},
	}

	err := cli.SetServiceProperties(spInput)
	c.Assert(err, chk.IsNil)

	spOutput, err := cli.GetServiceProperties()
	c.Assert(err, chk.IsNil)
	c.Assert(spOutput, chk.NotNil)
	c.Assert(*spOutput, chk.DeepEquals, spInput)

	// Back to defaults
	defaultRP := RetentionPolicy{
		Enabled: false,
		Days:    nil,
	}
	m.Enabled = false
	m.IncludeAPIs = nil
	m.RetentionPolicy = &defaultRP
	spInput.Logging.Delete = false
	spInput.Logging.Read = false
	spInput.Logging.Write = false
	spInput.Logging.RetentionPolicy = &defaultRP
	spInput.Cors = &Cors{nil}

	err = cli.SetServiceProperties(spInput)
	c.Assert(err, chk.IsNil)

	spOutput, err = cli.GetServiceProperties()
	c.Assert(err, chk.IsNil)
	c.Assert(spOutput, chk.NotNil)
	c.Assert(*spOutput, chk.DeepEquals, spInput)
}
