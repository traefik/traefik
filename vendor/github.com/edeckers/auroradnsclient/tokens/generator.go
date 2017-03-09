package tokens

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"github.com/Sirupsen/logrus"
	"strings"
	"time"
)

// NewToken generates a token for accessing a specific method of the API
func NewToken(userID string, key string, method string, action string, timestamp time.Time) string {
	fmtTime := timestamp.Format("20060102T150405Z")
	logrus.Debugf("Built timestamp: %s", fmtTime)

	message := strings.Join([]string{method, action, fmtTime}, "")
	logrus.Debugf("Built message: %s", message)

	signatureHmac := hmac.New(sha256.New, []byte(key))

	signatureHmac.Write([]byte(message))

	signature := base64.StdEncoding.EncodeToString([]byte(signatureHmac.Sum(nil)))
	logrus.Debugf("Built signature: %s", signature)

	userIDAndSignature := fmt.Sprintf("%s:%s", userID, signature)

	token := base64.StdEncoding.EncodeToString([]byte(userIDAndSignature))
	logrus.Debugf("Built token: %s", token)

	return token
}
