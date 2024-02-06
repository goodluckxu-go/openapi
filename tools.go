package openapi

import (
	"fmt"
	"github.com/invopop/yaml"
	"os"
	"strconv"
	"strings"
)

func isDir(filePath string) bool {
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return false
	}
	if fileInfo.IsDir() {
		return true
	}
	return false
}

func isFile(filePath string) bool {
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return false
	}
	if !fileInfo.IsDir() {
		return true
	}
	return false
}

func handleContentEnter(content string) string {
	content = strings.ReplaceAll(content, "\r\n", "\n")
	content = strings.ReplaceAll(content, "\r", "\n")
	return content
}

func getIndexFirst(s string, substr string) (first, other string) {
	idx := strings.Index(s, substr)
	if idx == -1 {
		first = s
		return
	}
	first = s[0:idx]
	other = s[idx+1:]
	return
}

func inArray[T comparable](v T, list []T) int {
	for i := range list {
		if list[i] == v {
			return i
		}
	}
	return -1
}

func toString(v any) string {
	if v == nil {
		return ""
	}
	return fmt.Sprintf("%v", v)
}

func toFloat64(v any) float64 {
	rs, _ := strconv.ParseFloat(toString(v), 64)
	return rs
}

func toUint64(v any) uint64 {
	rs, _ := strconv.ParseUint(toString(v), 10, 64)
	return rs
}

func toSliceInterface[t any](list []t) []interface{} {
	var rs []interface{}
	for _, v := range list {
		rs = append(rs, v)
	}
	return rs
}

func toPtr[t any](v t) *t {
	return &v
}

// Remove annotation symbols
func remoteAnnotationSymbols(s string) string {
	s = strings.Trim(s, " \t\n")
	s = strings.TrimPrefix(s, "//")
	s = strings.TrimPrefix(s, "/*")
	s = strings.TrimSuffix(s, "*/")
	s = strings.Trim(s, " \t\n")
	return s
}

func yamlMarshal(in interface{}) (out []byte, err error) {
	return yaml.Marshal(in)
}
