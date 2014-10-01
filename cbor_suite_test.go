package cbor_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestCbor(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Cbor Suite")
}
