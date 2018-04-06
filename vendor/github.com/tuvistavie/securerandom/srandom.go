package securerandom

import (
	"bytes"
	"crypto/rand"
	"encoding/base64"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"strings"
)

func RandomBytes(n int) ([]byte, error) {
	bytes := make([]byte, n)
	_, err := rand.Read(bytes)
	if err != nil {
		return nil, err
	}
	return bytes, nil
}

func Base64(n int, padded bool) (string, error) {
	bytes, err := RandomBytes(n)
	if err != nil {
		return "", err
	}
	result := base64.StdEncoding.EncodeToString(bytes)
	result = strings.Replace(result, "\n", "", -1)
	if !padded {
		result = strings.Replace(result, "=", "", -1)
	}
	return result, nil
}

func UrlSafeBase64(n int, padded bool) (string, error) {
	result, err := Base64(n, padded)
	if err != nil {
		return "", err
	}
	result = strings.Replace(result, "+", "-", -1)
	result = strings.Replace(result, "/", "_", -1)
	return result, nil
}

func Hex(n int) (string, error) {
	bytes, err := RandomBytes(n)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

func Uuid() (string, error) {
	var first, last uint32
	var middle [4]uint16
	randomBytes, err := RandomBytes(16)
	if err != nil {
		return "", err
	}
	buffer := bytes.NewBuffer(randomBytes)
	binary.Read(buffer, binary.BigEndian, &first)
	for i := 0; i < 4; i++ {
		binary.Read(buffer, binary.BigEndian, &middle[i])
	}
	binary.Read(buffer, binary.BigEndian, &last)
	middle[1] = (middle[1] & 0x0fff) | 0x4000
	middle[2] = (middle[2] & 0x3fff) | 0x8000
	return fmt.Sprintf("%08x-%04x-%04x-%04x-%04x%08x",
		first, middle[0], middle[1], middle[2], middle[3], last), nil
}
