package openapi

import (
	"archive/tar"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"github.com/invopop/yaml"
	"io"
	"net/http"
	"os"
	"path/filepath"
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

func IsFile(filePath string) bool {
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

func GetGithubLastDownUrl(projectPath string) (string, error) {
	apiUrl := fmt.Sprintf("https://api.github.com/repos/%v/releases/latest", projectPath)
	resp, err := http.Get(apiUrl)
	if err != nil {
		return "", err
	}
	buf, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	var res struct {
		TarballUrl string `json:"tarball_url"`
	}
	if err = json.Unmarshal(buf, &res); err != nil {
		return "", err
	}
	return res.TarballUrl, nil
}

func Download(downUrl string, outPath string) error {
	outDir, err := filepath.Abs(filepath.Dir(outPath))
	if err != nil {
		return err
	}
	fileInfo, err := os.Stat(outDir)
	if err != nil {
		_ = os.MkdirAll(outDir, 0777)
	} else if !fileInfo.IsDir() {
		return fmt.Errorf("已存在 %v 文件", outDir)
	}
	fileName := filepath.Base(outPath)
	filePath := filepath.Join(outDir, fileName)
	if IsFile(filePath) {
		return nil
	}
	resp, err := http.Get(downUrl)
	if err != nil {
		return err
	}
	buf, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	return os.WriteFile(filePath, buf, 0777)
}

func UnSwaggerTarball(filePath string, outDir string) error {
	f, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer f.Close()
	gzReader, err := gzip.NewReader(f)
	if err != nil {
		return err
	}
	defer gzReader.Close()
	tarReader := tar.NewReader(gzReader)
	// 遍历tar文件中的所有文件
	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			// 读取完毕
			break
		}
		if err != nil {
			return err
		}

		// 计算目标文件的路径
		tempPath := strings.ReplaceAll(header.Name, "\\", "/")
		list := strings.Split(tempPath, "/")
		if len(list) < 2 {
			continue
		}
		if list[1] != "dist" {
			continue
		}
		targetPath := filepath.Join(outDir, filepath.Clean(strings.Join(list[2:], "/")))
		// 根据header.Typeflag决定如何处理文件
		switch header.Typeflag {
		case tar.TypeReg:
			// 创建一个文件
			outFile, err := os.OpenFile(targetPath, os.O_CREATE|os.O_RDWR, os.FileMode(header.Mode))
			if err != nil {
				return err
			}
			// 将tar文件中的内容写入到文件中
			if _, err := io.Copy(outFile, tarReader); err != nil {
				outFile.Close()
				return err
			}
			outFile.Close()
		case tar.TypeDir:
			// 创建一个目录
			if _, err := os.Stat(targetPath); err != nil {
				if err := os.MkdirAll(targetPath, os.FileMode(header.Mode)); err != nil {
					return err
				}
			}
		default:
			// 忽略其他类型的文件
		}
	}
	return nil
}
