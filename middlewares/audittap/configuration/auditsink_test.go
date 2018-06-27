package configuration

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseToml(t *testing.T) {

	s := `
	[auditSink.headerMappings]
    	[auditSink.headerMappings.details]
		dfield1 = "header-1"

		dfield2= "header-2"


		[auditSink.headerMappings.tags]

		tfield1 = "header-3"
		#ignore me
	`

	result, _ := parseHeaderMappingsConfiguration(s)
	assert.EqualValues(t, &FieldHeaderMapping{"dfield1": "header-1", "dfield2": "header-2"}, result["details"])
	assert.EqualValues(t, &FieldHeaderMapping{"tfield1": "header-3"}, result["tags"])
}
