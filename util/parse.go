package util

import (
	"io"
	"io/ioutil"
	"strings"
)

func ConvertString2ReadCloser(str string) io.ReadCloser {
	return ioutil.NopCloser(strings.NewReader(str))
}
