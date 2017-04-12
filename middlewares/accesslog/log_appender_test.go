package accesslog

import (
	"bufio"
	"compress/gzip"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/containous/traefik/types"
	"github.com/stretchr/testify/assert"
)

func TestLogAppenderSimpleFile(t *testing.T) {
	file := logfilePath("")
	defer os.Remove(file)

	settings := &types.AccessLog{File: file}
	la, err := NewLogAppender(settings)
	assert.Nil(t, err)

	n := 1000
	for i := 1; i <= n; i++ {
		err = la.Write(fixtureLogDataTable(uint64(i)))
		assert.Nil(t, err, "%v", err)
	}

	err = la.Close()
	assert.Nil(t, err, "%v", err)

	if logdata, err := os.Open(file); err != nil {
		assert.Fail(t, "Failed to read logfile", "%v", err)
	} else {
		scanner := bufio.NewScanner(logdata)

		for i := 1; i <= n; i++ {
			assert.True(t, scanner.Scan())
			line := scanner.Text()
			tokens := strings.Split(line, " ")
			if assert.Equal(t, 16, len(tokens), line) {
				assert.Equal(t, testRemoteHost, tokens[0], line)
				assert.Equal(t, testUsername, tokens[2], line)
				assert.Equal(t, fmt.Sprintf("\"%s", "GET"), tokens[5], line)
				assert.Equal(t, fmt.Sprintf("%s", "/y/xy/z"), tokens[6], line)
				assert.Equal(t, fmt.Sprintf("%s\"", "HTTP/1.1"), tokens[7], line)
				assert.Equal(t, fmt.Sprintf("%d", 200), tokens[8], line)
				assert.Equal(t, "102", tokens[9], line)
				assert.Equal(t, quoted(testReferrer), tokens[10], line)
				assert.Equal(t, quoted(testUserAgent), tokens[11], line)
				assert.Equal(t, fmt.Sprintf("%d", i), tokens[12], line)
				assert.Equal(t, quoted(testFrontendName), tokens[13], line)
				assert.Equal(t, quoted(testTargetURL), tokens[14], line)
			}
		}

		assert.False(t, scanner.Scan())
	}
}

func TestLogAppenderBufferedFile(t *testing.T) {
	file := logfilePath("")
	defer os.Remove(file)

	settings := &types.AccessLog{File: file, BufferSize: "4KiB"}
	la, err := NewLogAppender(settings)
	assert.Nil(t, err)

	n := 1000
	for i := 1; i <= n; i++ {
		err = la.Write(fixtureLogDataTable(uint64(i)))
		assert.Nil(t, err, "%v", err)
	}

	err = la.Close()
	assert.Nil(t, err, "%v", err)

	if logdata, err := os.Open(file); err != nil {
		assert.Fail(t, "Failed to read logfile", "%v", err)
	} else {
		scanner := bufio.NewScanner(logdata)

		for i := 1; i <= n; i++ {
			assert.True(t, scanner.Scan())
			line := scanner.Text()
			tokens := strings.Split(line, " ")
			if assert.Equal(t, 16, len(tokens), line) {
				assert.Equal(t, testRemoteHost, tokens[0], line)
				assert.Equal(t, testUsername, tokens[2], line)
				assert.Equal(t, fmt.Sprintf("\"%s", "GET"), tokens[5], line)
				assert.Equal(t, fmt.Sprintf("%s", "/y/xy/z"), tokens[6], line)
				assert.Equal(t, fmt.Sprintf("%s\"", "HTTP/1.1"), tokens[7], line)
				assert.Equal(t, fmt.Sprintf("%d", 200), tokens[8], line)
				assert.Equal(t, "102", tokens[9], line)
				assert.Equal(t, quoted(testReferrer), tokens[10], line)
				assert.Equal(t, quoted(testUserAgent), tokens[11], line)
				assert.Equal(t, fmt.Sprintf("%d", i), tokens[12], line)
				assert.Equal(t, quoted(testFrontendName), tokens[13], line)
				assert.Equal(t, quoted(testTargetURL), tokens[14], line)
			}
		}

		assert.False(t, scanner.Scan())
	}
}

func TestLogAppenderCompressedFile(t *testing.T) {
	file := logfilePath(".gz")
	defer os.Remove(file)

	cases := []int{
		gzip.NoCompression,
		gzip.DefaultCompression,
		gzip.BestCompression,
	}

	for _, level := range cases {
		settings := &types.AccessLog{File: file, GzipLevel: level}
		la, err := NewLogAppender(settings)
		assert.Nil(t, err)

		n := 1000
		for i := 1; i <= n; i++ {
			err = la.Write(fixtureLogDataTable(uint64(i)))
			assert.Nil(t, err, "%v", err)
		}

		err = la.Close()
		assert.Nil(t, err, "%v", err)

		if logdata, err := os.Open(file); err != nil {
			assert.Fail(t, "Failed to read logfile", "%v", err)
		} else {
			gz, err := gzip.NewReader(logdata)
			assert.Nil(t, err, "%v", err)
			scanner := bufio.NewScanner(gz)

			for i := 1; i <= n; i++ {
				assert.True(t, scanner.Scan())
				line := scanner.Text()
				tokens := strings.Split(line, " ")
				if assert.Equal(t, 16, len(tokens), line) {
					assert.Equal(t, testRemoteHost, tokens[0], line)
					assert.Equal(t, testUsername, tokens[2], line)
					assert.Equal(t, fmt.Sprintf("\"%s", "GET"), tokens[5], line)
					assert.Equal(t, fmt.Sprintf("%s", "/y/xy/z"), tokens[6], line)
					assert.Equal(t, fmt.Sprintf("%s\"", "HTTP/1.1"), tokens[7], line)
					assert.Equal(t, fmt.Sprintf("%d", 200), tokens[8], line)
					assert.Equal(t, "102", tokens[9], line)
					assert.Equal(t, quoted(testReferrer), tokens[10], line)
					assert.Equal(t, quoted(testUserAgent), tokens[11], line)
					assert.Equal(t, fmt.Sprintf("%d", i), tokens[12], line)
					assert.Equal(t, quoted(testFrontendName), tokens[13], line)
					assert.Equal(t, quoted(testTargetURL), tokens[14], line)
				}
			}

			assert.False(t, scanner.Scan())
			os.Remove(file)
		}
	}
}

func TestJsonAppenderDefault(t *testing.T) {
	settings := &types.AccessLog{
		Format:     "json",
		TimeFormat: commonLogTimeFormat,
	}

	la, err := NewLogAppender(settings)

	assert.Nil(t, err, "%v", err)
	assert.Equal(t, jsonLogFormatter{timeFormat: commonLogTimeFormat,
		coreMapping: []tuple{
			{"StartUTC", "StartUTC"},
			{"Duration", "Duration"},
			{"FrontendName", "FrontendName"},
			{"BackendName", "BackendName"},
			{"BackendURL", "BackendURL"},
			{"ClientHost", "ClientHost"},
			{"ClientPort", "ClientPort"},
			{"ClientUsername", "ClientUsername"},
			{"RequestHost", "RequestHost"},
			{"RequestPort", "RequestPort"},
			{"RequestMethod", "RequestMethod"},
			{"RequestPath", "RequestPath"},
			{"RequestProtocol", "RequestProtocol"},
			{"RequestContentSize", "RequestContentSize"},
			{"OriginDuration", "OriginDuration"},
			{"OriginContentSize", "OriginContentSize"},
			{"OriginStatus", "OriginStatus"},
			{"DownstreamStatus", "DownstreamStatus"},
			{"DownstreamContentSize", "DownstreamContentSize"},
			{"RequestCount", "RequestCount"},
		}},
		la.formatter.(jsonLogFormatter))
}

func TestJsonAppenderSetup(t *testing.T) {
	settings := &types.AccessLog{
		Format:     "json",
		TimeFormat: commonLogTimeFormat,
		CoreFields: []string{"Duration:took", "StartUTC", "FrontendName:frontend"},
	}
	la, err := NewLogAppender(settings)
	assert.Nil(t, err, "%v", err)
	assert.Equal(t, jsonLogFormatter{timeFormat: commonLogTimeFormat,
		coreMapping: []tuple{
			{"Duration", "took"},
			{"StartUTC", "StartUTC"},
			{"FrontendName", "frontend"},
		}},
		la.formatter.(jsonLogFormatter))
}
