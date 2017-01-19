package util

import (
	"fmt"
	"io"
	"io/ioutil"
	"strings"
)

func ConvertString2ReadCloser(format string, args ...interface{}) io.ReadCloser {
	return ioutil.NopCloser(strings.NewReader(fmt.Sprintf(format, args)))
}
