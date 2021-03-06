package test

import (
	"go/build"
	"math/rand"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/uol/gobol/saw"
	"go.uber.org/zap"
)

// Has some useful functions used in the tests.
// @author rnojiri

// GetLogger - creates a test logger
func GetLogger(t *testing.T) *zap.Logger {

	logger, err := saw.New("DEBUG", "QA")
	assert.NoError(t, err)

	return logger
}

// GetScanPathRoot - build a scan path
func GetScanPathRoot(path string) string {

	gopath := os.Getenv("GOPATH")
	if gopath == "" {
		gopath = build.Default.GOPATH
	}

	return gopath + "src/stash.uol.intranet/s3-log-uploader/" + path
}

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

// RandomString - generates random strings
func RandomString(n int) string {
	rand.Seed(int64(time.Now().Nanosecond()))
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return strings.ToLower(string(b))
}
