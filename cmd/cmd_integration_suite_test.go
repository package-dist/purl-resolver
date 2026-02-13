//go:build integration

package cmd_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestCmdIntegration(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Cmd Integration Suite")
}
