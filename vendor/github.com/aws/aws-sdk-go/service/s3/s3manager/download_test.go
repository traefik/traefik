package s3manager_test

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"regexp"
	"strconv"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/awstesting/unit"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

func dlLoggingSvc(data []byte) (*s3.S3, *[]string, *[]string) {
	var m sync.Mutex
	names := []string{}
	ranges := []string{}

	svc := s3.New(unit.Session)
	svc.Handlers.Send.Clear()
	svc.Handlers.Send.PushBack(func(r *request.Request) {
		m.Lock()
		defer m.Unlock()

		names = append(names, r.Operation.Name)
		ranges = append(ranges, *r.Params.(*s3.GetObjectInput).Range)

		rerng := regexp.MustCompile(`bytes=(\d+)-(\d+)`)
		rng := rerng.FindStringSubmatch(r.HTTPRequest.Header.Get("Range"))
		start, _ := strconv.ParseInt(rng[1], 10, 64)
		fin, _ := strconv.ParseInt(rng[2], 10, 64)
		fin++

		if fin > int64(len(data)) {
			fin = int64(len(data))
		}

		bodyBytes := data[start:fin]
		r.HTTPResponse = &http.Response{
			StatusCode: 200,
			Body:       ioutil.NopCloser(bytes.NewReader(bodyBytes)),
			Header:     http.Header{},
		}
		r.HTTPResponse.Header.Set("Content-Range", fmt.Sprintf("bytes %d-%d/%d",
			start, fin-1, len(data)))
		r.HTTPResponse.Header.Set("Content-Length", fmt.Sprintf("%d", len(bodyBytes)))
	})

	return svc, &names, &ranges
}

func dlLoggingSvcNoChunk(data []byte) (*s3.S3, *[]string) {
	var m sync.Mutex
	names := []string{}

	svc := s3.New(unit.Session)
	svc.Handlers.Send.Clear()
	svc.Handlers.Send.PushBack(func(r *request.Request) {
		m.Lock()
		defer m.Unlock()

		names = append(names, r.Operation.Name)

		r.HTTPResponse = &http.Response{
			StatusCode: 200,
			Body:       ioutil.NopCloser(bytes.NewReader(data[:])),
			Header:     http.Header{},
		}
		r.HTTPResponse.Header.Set("Content-Length", fmt.Sprintf("%d", len(data)))
	})

	return svc, &names
}

func dlLoggingSvcNoContentRangeLength(data []byte, states []int) (*s3.S3, *[]string) {
	var m sync.Mutex
	names := []string{}
	var index int = 0

	svc := s3.New(unit.Session)
	svc.Handlers.Send.Clear()
	svc.Handlers.Send.PushBack(func(r *request.Request) {
		m.Lock()
		defer m.Unlock()

		names = append(names, r.Operation.Name)

		r.HTTPResponse = &http.Response{
			StatusCode: states[index],
			Body:       ioutil.NopCloser(bytes.NewReader(data[:])),
			Header:     http.Header{},
		}
		index++
	})

	return svc, &names
}

func dlLoggingSvcContentRangeTotalAny(data []byte, states []int) (*s3.S3, *[]string) {
	var m sync.Mutex
	names := []string{}
	ranges := []string{}
	var index int = 0

	svc := s3.New(unit.Session)
	svc.Handlers.Send.Clear()
	svc.Handlers.Send.PushBack(func(r *request.Request) {
		m.Lock()
		defer m.Unlock()

		names = append(names, r.Operation.Name)
		ranges = append(ranges, *r.Params.(*s3.GetObjectInput).Range)

		rerng := regexp.MustCompile(`bytes=(\d+)-(\d+)`)
		rng := rerng.FindStringSubmatch(r.HTTPRequest.Header.Get("Range"))
		start, _ := strconv.ParseInt(rng[1], 10, 64)
		fin, _ := strconv.ParseInt(rng[2], 10, 64)
		fin++

		if fin >= int64(len(data)) {
			fin = int64(len(data))
		}

		// Setting start and finish to 0 because this state of 1 is suppose to
		// be an error state of 416
		if index == len(states)-1 {
			start = 0
			fin = 0
		}

		bodyBytes := data[start:fin]

		r.HTTPResponse = &http.Response{
			StatusCode: states[index],
			Body:       ioutil.NopCloser(bytes.NewReader(bodyBytes)),
			Header:     http.Header{},
		}
		r.HTTPResponse.Header.Set("Content-Range", fmt.Sprintf("bytes %d-%d/*",
			start, fin-1))
		index++
	})

	return svc, &names
}

func dlLoggingSvcWithErrReader(cases []testErrReader) (*s3.S3, *[]string) {
	var m sync.Mutex
	names := []string{}
	var index int = 0

	svc := s3.New(unit.Session, &aws.Config{
		MaxRetries: aws.Int(len(cases) - 1),
	})
	svc.Handlers.Send.Clear()
	svc.Handlers.Send.PushBack(func(r *request.Request) {
		m.Lock()
		defer m.Unlock()

		names = append(names, r.Operation.Name)

		c := cases[index]

		r.HTTPResponse = &http.Response{
			StatusCode: http.StatusOK,
			Body:       ioutil.NopCloser(&c),
			Header:     http.Header{},
		}
		r.HTTPResponse.Header.Set("Content-Range",
			fmt.Sprintf("bytes %d-%d/%d", 0, c.Len-1, c.Len))
		r.HTTPResponse.Header.Set("Content-Length", fmt.Sprintf("%d", c.Len))
		index++
	})

	return svc, &names
}

func TestDownloadOrder(t *testing.T) {
	s, names, ranges := dlLoggingSvc(buf12MB)

	d := s3manager.NewDownloaderWithClient(s, func(d *s3manager.Downloader) {
		d.Concurrency = 1
	})
	w := &aws.WriteAtBuffer{}
	n, err := d.Download(w, &s3.GetObjectInput{
		Bucket: aws.String("bucket"),
		Key:    aws.String("key"),
	})

	assert.Nil(t, err)
	assert.Equal(t, int64(len(buf12MB)), n)
	assert.Equal(t, []string{"GetObject", "GetObject", "GetObject"}, *names)
	assert.Equal(t, []string{"bytes=0-5242879", "bytes=5242880-10485759", "bytes=10485760-15728639"}, *ranges)

	count := 0
	for _, b := range w.Bytes() {
		count += int(b)
	}
	assert.Equal(t, 0, count)
}

func TestDownloadZero(t *testing.T) {
	s, names, ranges := dlLoggingSvc([]byte{})

	d := s3manager.NewDownloaderWithClient(s)
	w := &aws.WriteAtBuffer{}
	n, err := d.Download(w, &s3.GetObjectInput{
		Bucket: aws.String("bucket"),
		Key:    aws.String("key"),
	})

	assert.Nil(t, err)
	assert.Equal(t, int64(0), n)
	assert.Equal(t, []string{"GetObject"}, *names)
	assert.Equal(t, []string{"bytes=0-5242879"}, *ranges)
}

func TestDownloadSetPartSize(t *testing.T) {
	s, names, ranges := dlLoggingSvc([]byte{1, 2, 3})

	d := s3manager.NewDownloaderWithClient(s, func(d *s3manager.Downloader) {
		d.Concurrency = 1
		d.PartSize = 1
	})
	w := &aws.WriteAtBuffer{}
	n, err := d.Download(w, &s3.GetObjectInput{
		Bucket: aws.String("bucket"),
		Key:    aws.String("key"),
	})

	assert.Nil(t, err)
	assert.Equal(t, int64(3), n)
	assert.Equal(t, []string{"GetObject", "GetObject", "GetObject"}, *names)
	assert.Equal(t, []string{"bytes=0-0", "bytes=1-1", "bytes=2-2"}, *ranges)
	assert.Equal(t, []byte{1, 2, 3}, w.Bytes())
}

func TestDownloadError(t *testing.T) {
	s, names, _ := dlLoggingSvc([]byte{1, 2, 3})

	num := 0
	s.Handlers.Send.PushBack(func(r *request.Request) {
		num++
		if num > 1 {
			r.HTTPResponse.StatusCode = 400
			r.HTTPResponse.Body = ioutil.NopCloser(bytes.NewReader([]byte{}))
		}
	})

	d := s3manager.NewDownloaderWithClient(s, func(d *s3manager.Downloader) {
		d.Concurrency = 1
		d.PartSize = 1
	})
	w := &aws.WriteAtBuffer{}
	n, err := d.Download(w, &s3.GetObjectInput{
		Bucket: aws.String("bucket"),
		Key:    aws.String("key"),
	})

	assert.NotNil(t, err)
	assert.Equal(t, int64(1), n)
	assert.Equal(t, []string{"GetObject", "GetObject"}, *names)
	assert.Equal(t, []byte{1}, w.Bytes())
}

func TestDownloadNonChunk(t *testing.T) {
	s, names := dlLoggingSvcNoChunk(buf2MB)

	d := s3manager.NewDownloaderWithClient(s, func(d *s3manager.Downloader) {
		d.Concurrency = 1
	})
	w := &aws.WriteAtBuffer{}
	n, err := d.Download(w, &s3.GetObjectInput{
		Bucket: aws.String("bucket"),
		Key:    aws.String("key"),
	})

	assert.Nil(t, err)
	assert.Equal(t, int64(len(buf2MB)), n)
	assert.Equal(t, []string{"GetObject"}, *names)

	count := 0
	for _, b := range w.Bytes() {
		count += int(b)
	}
	assert.Equal(t, 0, count)
}

func TestDownloadNoContentRangeLength(t *testing.T) {
	s, names := dlLoggingSvcNoContentRangeLength(buf2MB, []int{200, 416})

	d := s3manager.NewDownloaderWithClient(s, func(d *s3manager.Downloader) {
		d.Concurrency = 1
	})
	w := &aws.WriteAtBuffer{}
	n, err := d.Download(w, &s3.GetObjectInput{
		Bucket: aws.String("bucket"),
		Key:    aws.String("key"),
	})

	assert.Nil(t, err)
	assert.Equal(t, int64(len(buf2MB)), n)
	assert.Equal(t, []string{"GetObject", "GetObject"}, *names)

	count := 0
	for _, b := range w.Bytes() {
		count += int(b)
	}
	assert.Equal(t, 0, count)
}

func TestDownloadContentRangeTotalAny(t *testing.T) {
	s, names := dlLoggingSvcContentRangeTotalAny(buf2MB, []int{200, 416})

	d := s3manager.NewDownloaderWithClient(s, func(d *s3manager.Downloader) {
		d.Concurrency = 1
	})
	w := &aws.WriteAtBuffer{}
	n, err := d.Download(w, &s3.GetObjectInput{
		Bucket: aws.String("bucket"),
		Key:    aws.String("key"),
	})

	assert.Nil(t, err)
	assert.Equal(t, int64(len(buf2MB)), n)
	assert.Equal(t, []string{"GetObject", "GetObject"}, *names)

	count := 0
	for _, b := range w.Bytes() {
		count += int(b)
	}
	assert.Equal(t, 0, count)
}

func TestDownloadPartBodyRetry_SuccessRetry(t *testing.T) {
	s, names := dlLoggingSvcWithErrReader([]testErrReader{
		{Buf: []byte("ab"), Len: 3, Err: io.ErrUnexpectedEOF},
		{Buf: []byte("123"), Len: 3, Err: io.EOF},
	})

	d := s3manager.NewDownloaderWithClient(s, func(d *s3manager.Downloader) {
		d.Concurrency = 1
	})

	w := &aws.WriteAtBuffer{}
	n, err := d.Download(w, &s3.GetObjectInput{
		Bucket: aws.String("bucket"),
		Key:    aws.String("key"),
	})

	assert.Nil(t, err)
	assert.Equal(t, int64(3), n)
	assert.Equal(t, []string{"GetObject", "GetObject"}, *names)
	assert.Equal(t, []byte("123"), w.Bytes())
}

func TestDownloadPartBodyRetry_SuccessNoRetry(t *testing.T) {
	s, names := dlLoggingSvcWithErrReader([]testErrReader{
		{Buf: []byte("abc"), Len: 3, Err: io.EOF},
	})

	d := s3manager.NewDownloaderWithClient(s, func(d *s3manager.Downloader) {
		d.Concurrency = 1
	})

	w := &aws.WriteAtBuffer{}
	n, err := d.Download(w, &s3.GetObjectInput{
		Bucket: aws.String("bucket"),
		Key:    aws.String("key"),
	})

	assert.Nil(t, err)
	assert.Equal(t, int64(3), n)
	assert.Equal(t, []string{"GetObject"}, *names)
	assert.Equal(t, []byte("abc"), w.Bytes())
}

func TestDownloadPartBodyRetry_FailRetry(t *testing.T) {
	s, names := dlLoggingSvcWithErrReader([]testErrReader{
		{Buf: []byte("ab"), Len: 3, Err: io.ErrUnexpectedEOF},
	})

	d := s3manager.NewDownloaderWithClient(s, func(d *s3manager.Downloader) {
		d.Concurrency = 1
	})

	w := &aws.WriteAtBuffer{}
	n, err := d.Download(w, &s3.GetObjectInput{
		Bucket: aws.String("bucket"),
		Key:    aws.String("key"),
	})

	assert.Error(t, err)
	assert.Equal(t, int64(2), n)
	assert.Equal(t, []string{"GetObject"}, *names)
	assert.Equal(t, []byte("ab"), w.Bytes())
}

type testErrReader struct {
	Buf []byte
	Err error
	Len int64

	off int
}

func (r *testErrReader) Read(p []byte) (int, error) {
	to := len(r.Buf) - r.off

	n := copy(p, r.Buf[r.off:to])
	r.off += n

	if n < len(p) {
		return n, r.Err

	}

	return n, nil
}
