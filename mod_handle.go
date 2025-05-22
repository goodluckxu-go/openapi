package openapi

import (
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
)

// 处理mode便于获取引入的struct
type modHandle map[string]string

func (m *modHandle) load(filePath string) (modName string, err error) {
	modFilePath := filepath.Join(filePath, "go.mod")
	vendorFilePath := filepath.Join(filePath, "vendor")
	var modBuf []byte
	modBuf, err = os.ReadFile(modFilePath)
	if err != nil {
		return
	}
	if isDir(vendorFilePath) {
		modName = m.parseMod(string(modBuf), true, vendorFilePath)
	} else {
		modName = m.parseMod(string(modBuf), false, m.modAbsPath())
	}
	return
}

func (m *modHandle) modAbsPath() string {
	userPath := ""
	if runtime.GOOS == "windows" {
		userPath = os.Getenv("USERPROFILE")
	} else {
		userPath = os.Getenv("HOME")
	}
	return filepath.Join(userPath, "go", "pkg", "mod")
}

func (m *modHandle) parseMod(content string, isVendor bool, baseDir string) (modName string) {
	// windows格式转linux格式
	content = strings.ReplaceAll(content, "\r\n", "\n")
	// mac格式转linux格式
	content = strings.ReplaceAll(content, "\r", "\n")
	// 获取当前mod
	list := regexp.MustCompile("module( |\t)+(.*?)\n").FindAllStringSubmatch(content, -1)
	if len(list) == 0 {
		return
	}
	modName = list[0][2]
	projectDir, _ := os.Getwd()
	(*m)[modName] = projectDir
	baseDir, _ = filepath.Abs(baseDir)
	content = handleContentEnter(content)
	content = regexp.MustCompile(`( |\t)+`).ReplaceAllString(content, " ")
	// 解析require
	reg := regexp.MustCompile(`require( |\t)+\(((?s).*?)\n\)`)
	list = reg.FindAllStringSubmatch(content, -1)
	for _, v := range list {
		requireList := strings.Split(v[2], "\n")
		for _, v1 := range requireList {
			v1 = strings.Trim(v1, " ")
			if v1 == "" {
				continue
			}
			v1List := strings.Split(v1, " ")
			if len(v1List) < 2 {
				continue
			}
			if isVendor {
				(*m)[v1List[0]] = filepath.Join(baseDir, v1List[0])
			} else {
				(*m)[v1List[0]] = filepath.Join(baseDir, v1List[0]+"@"+v1List[1])
			}
		}
	}
	return
}
