package util

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"strings"
)

func ConvertString2ReadCloser(str string) io.ReadCloser {
	return ioutil.NopCloser(strings.NewReader(str))
}

func JsonMarshal(v interface{}) string {
	buf, _ := json.Marshal(v)
	return string(buf)
}
