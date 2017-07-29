package gominlog_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestGominlog(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Gominlog Suite")
}
