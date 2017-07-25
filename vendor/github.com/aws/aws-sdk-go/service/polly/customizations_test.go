package polly

import (
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/awstesting/unit"

	"github.com/stretchr/testify/assert"
)

func TestRestGETStrategy(t *testing.T) {
	svc := New(unit.Session, &aws.Config{Region: aws.String("us-west-2")})
	r, _ := svc.SynthesizeSpeechRequest(nil)
	err := restGETPresignStrategy(r)
	assert.NoError(t, err)
	assert.Equal(t, "GET", r.HTTPRequest.Method)
	assert.NotEqual(t, nil, r.Operation.BeforePresignFn)
}

func TestPresign(t *testing.T) {
	svc := New(unit.Session, &aws.Config{Region: aws.String("us-west-2")})
	r, _ := svc.SynthesizeSpeechRequest(&SynthesizeSpeechInput{
		Text:         aws.String("Moo"),
		OutputFormat: aws.String("mp3"),
		VoiceId:      aws.String("Foo"),
	})
	url, err := r.Presign(time.Second)
	assert.NoError(t, err)
	assert.Regexp(t, `^https://polly.us-west-2.amazonaws.com/v1/speech\?.*?OutputFormat=mp3.*?Text=Moo.*?VoiceId=Foo.*`, url)
}
