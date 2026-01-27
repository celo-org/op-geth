package tests

import (
	"os"
	"strings"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
)

var (
	celoSkipTests     map[string]bool
	celoSkipTestsOnce sync.Once
)

func loadCeloSkipTests(t *testing.T) {
	celoSkipTests = make(map[string]bool)
	data, err := os.ReadFile("celo_skip_tests.txt")
	require.NoError(t, err)
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		celoSkipTests[line] = true
	}
}

func skipCeloTests(t *testing.T, name string) {
	celoSkipTestsOnce.Do(func() { loadCeloSkipTests(t) })
	if celoSkipTests[name] {
		t.Skip("skipped by Celo skip list")
	}
}
