package openapi

import "testing"

func TestGenerateOpenAPI(t *testing.T) {
	GenerateOpenAPI("./", "./examples", "./examples/doc.go", "./examples/docs", "")
}
