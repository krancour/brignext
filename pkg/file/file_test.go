package file

import (
	"fmt"
	"go/build"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestExists(t *testing.T) {
	gopath := os.Getenv("GOPATH")
	if gopath == "" {
		gopath = build.Default.GOPATH
	}
	file := fmt.Sprintf(
		"%s/src/github.com/krancour/brignext/pkg/file/file_test.go",
		gopath,
	)
	require.True(t, Exists(file))
	file = fmt.Sprintf(
		"%s/src/github.com/krancour/brignext/pkg/file/bogus.go",
		gopath,
	)
	require.False(t, Exists(file))
}
