//go:build !integration

package cmd

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestCmdUnit(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Cmd Unit Suite")
}
