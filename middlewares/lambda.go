package middlewares

import (
	"encoding/base64"
	"encoding/json"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/defaults"
	"github.com/aws/aws-sdk-go/aws/ec2metadata"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/lambda"
	"github.com/containous/traefik/middlewares/tracing"
	"io/ioutil"
	"net/http"
)

// Retry is a middleware that retries requests
type Lambda struct {
	next http.Handler
}

// NewRetry returns a new Retry instance
func NewLambda(next http.Handler) *Lambda {
	return &Lambda{
		next: next,
	}
}

func (l *Lambda) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	sess, err := session.NewSession()
	if err != nil {
		return
	}
	ec2meta := ec2metadata.New(sess)
	identity, err := ec2meta.GetInstanceIdentityDocument()

	cfg := &aws.Config{
		Region: &identity.Region,
		Credentials: credentials.NewChainCredentials(
			[]credentials.Provider{
				&credentials.StaticProvider{
					Value: credentials.Value{
						AccessKeyID:     "",
						SecretAccessKey: "",
					},
				},
				&credentials.EnvProvider{},
				&credentials.SharedCredentialsProvider{},
				defaults.RemoteCredProvider(*(defaults.Config()), defaults.Handlers()),
			}),
	}

	svc := lambda.New(sess, cfg)
	body, err := ioutil.ReadAll(r.Body)
	jsonString, _ := json.Marshal(
		map[string]map[string]string{
			"custom": {
				"X-Request-Context":         r.Header.Get("X-Request-Context"),
				"X-User-Context":            r.Header.Get("X-User-Context"),
				"X-Original-Request-Method": r.Method,
				"X-Original-Request-Url":    r.RequestURI,
			},
		},
	)
	userContext := string(base64.StdEncoding.EncodeToString([]byte(jsonString)))
	input := &lambda.InvokeInput{
		FunctionName:   aws.String(r.URL.Host),
		InvocationType: aws.String("RequestResponse"),
		Payload:        []byte(body),
		ClientContext:  &userContext,
	}
	req, resp := svc.InvokeRequest(input)
	err = req.Send()
	if err != nil {

		/*			switch aerr.Code() {
					case lambda.ErrCodeServiceException:
						fmt.Println(lambda.ErrCodeServiceException, aerr.Error())
					case lambda.ErrCodeResourceNotFoundException:
						fmt.Println(lambda.ErrCodeResourceNotFoundException, aerr.Error())
					case lambda.ErrCodeInvalidRequestContentException:
						fmt.Println(lambda.ErrCodeInvalidRequestContentException, aerr.Error())
					case lambda.ErrCodeRequestTooLargeException:
						fmt.Println(lambda.ErrCodeRequestTooLargeException, aerr.Error())
					case lambda.ErrCodeUnsupportedMediaTypeException:
						fmt.Println(lambda.ErrCodeUnsupportedMediaTypeException, aerr.Error())
					case lambda.ErrCodeTooManyRequestsException:
						fmt.Println(lambda.ErrCodeTooManyRequestsException, aerr.Error())
					case lambda.ErrCodeInvalidParameterValueException:
						fmt.Println(lambda.ErrCodeInvalidParameterValueException, aerr.Error())
					case lambda.ErrCodeEC2UnexpectedException:
						fmt.Println(lambda.ErrCodeEC2UnexpectedException, aerr.Error())
					case lambda.ErrCodeSubnetIPAddressLimitReachedException:
						fmt.Println(lambda.ErrCodeSubnetIPAddressLimitReachedException, aerr.Error())
					case lambda.ErrCodeENILimitReachedException:
						fmt.Println(lambda.ErrCodeENILimitReachedException, aerr.Error())
					case lambda.ErrCodeEC2ThrottledException:
						fmt.Println(lambda.ErrCodeEC2ThrottledException, aerr.Error())
					case lambda.ErrCodeEC2AccessDeniedException:
						fmt.Println(lambda.ErrCodeEC2AccessDeniedException, aerr.Error())
					case lambda.ErrCodeInvalidSubnetIDException:
						fmt.Println(lambda.ErrCodeInvalidSubnetIDException, aerr.Error())
					case lambda.ErrCodeInvalidSecurityGroupIDException:
						fmt.Println(lambda.ErrCodeInvalidSecurityGroupIDException, aerr.Error())
					case lambda.ErrCodeInvalidZipFileException:
						fmt.Println(lambda.ErrCodeInvalidZipFileException, aerr.Error())
					case lambda.ErrCodeKMSDisabledException:
						fmt.Println(lambda.ErrCodeKMSDisabledException, aerr.Error())
					case lambda.ErrCodeKMSInvalidStateException:
						fmt.Println(lambda.ErrCodeKMSInvalidStateException, aerr.Error())
					case lambda.ErrCodeKMSAccessDeniedException:
						fmt.Println(lambda.ErrCodeKMSAccessDeniedException, aerr.Error())
					case lambda.ErrCodeKMSNotFoundException:
						fmt.Println(lambda.ErrCodeKMSNotFoundException, aerr.Error())
					case lambda.ErrCodeInvalidRuntimeException:
						fmt.Println(lambda.ErrCodeInvalidRuntimeException, aerr.Error())
					default:
						fmt.Println(aerr.Error())
					}
		*/

		aerr := err.(awserr.Error)
		tracing.LogResponseCode(tracing.GetSpan(r), 400)
		rw.WriteHeader(400)
		rw.Write([]byte(aerr.Code() + aerr.Error()))
		return

	} else {
		tracing.LogResponseCode(tracing.GetSpan(r), 200)
		rw.WriteHeader(200)
		rw.Write(resp.Payload)
		return
	}
}
