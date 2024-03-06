package openapi

import (
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

type ginHandle struct {
	routesFunc     []routeFuncInfo
	importAliasMap map[string]string
	structAliasMap map[string]string
}

func (g *ginHandle) load(routesFunc []routeFuncInfo, routeDir string) {
	if !isDir(routeDir) {
		err := os.MkdirAll(routeDir, 0777)
		if err != nil {
			log.Fatal(err)
		}
	}
	g.routesFunc = routesFunc
	routesPackage := filepath.Base(routeDir)
	g.importAliasMap = map[string]string{}
	g.structAliasMap = map[string]string{}
	content := "package " + routesPackage + "\n\n"
	content += g.generateImport() + "\n\n"
	content += g.generateStructDefine() + "\n\n"
	content += g.generateRoutes() + "\n\n"
	routePath := filepath.Join(routeDir, "commentsRoutes.go")
	err := os.WriteFile(routePath, []byte(content), 0777)
	if err != nil {
		log.Fatal(err)
	}
}

func (g *ginHandle) generateRoutes() string {
	reg := regexp.MustCompile(`\{(.*?)\}`)
	// 处理Any情况
	anyPathMaps := map[string]bool{}
	pathMaps := map[string][]routeFuncInfo{}
	var securityList []string
	securityMap := map[string]bool{}
	for k, v := range g.routesFunc {
		// 处理path
		v.path = reg.ReplaceAllString(v.path, ":$1")
		// 处理method
		method := "Any"
		if inArray(v.method, []string{"get", "put", "post", "delete", "options", "head", "patch"}) != -1 {
			method = strings.ToUpper(v.method)
		}
		if method == "Any" {
			anyPathMaps[v.path] = true
		}
		pathMaps[v.path] = append(pathMaps[v.path], v)
		v.method = method
		g.routesFunc[k] = v
		for _, v1 := range v.security {
			if securityMap[v1] {
				continue
			}
			securityMap[v1] = true
			securityList = append(securityList, underlineToHumpFirstLower(v1)+"Middleware")
		}
	}
	content := "func RegisterRoutes(routes gin.IRoutes"
	if len(securityList) > 0 {
		content += ", "
		content += strings.Join(securityList, ", ")
		content += " " + "gin.HandlerFunc"
	}
	content += ") {\n"
	for _, v := range g.routesFunc {
		// 添加注释
		content += "\t// " + v.summary + "\n"
		content += "\troutes." + v.method + "(\"" + v.path + "\", setHandlers("
		// 添加中间件
		for _, v1 := range v.security {
			middlewareName := underlineToHumpFirstLower(v1) + "Middleware"
			content += middlewareName + ", "
		}
		// 添加路由
		content += g.getStructAlias(v) + "." + v.funcName
		content += ")...)\n"

	}
	content += "}\n\n"
	// 增加设置handlers方法
	content += "func setHandlers(handlers ...gin.HandlerFunc) (rs []gin.HandlerFunc) {\n"
	content += "\tfor _, handler := range handlers {\n"
	content += "\t\tif handler != nil {\n"
	content += "\t\t\trs = append(rs, handler)\n"
	content += "\t\t}\n"
	content += "\t}\n"
	content += "\treturn\n"
	content += "}"
	return content
}

func (g *ginHandle) generateStructDefine() string {
	structsMap := map[string]bool{}
	maxLen := 0
	var structs []map[string]string
	for _, v := range g.routesFunc {
		if v.funcStruct == "" {
			continue
		}
		aliasImport := g.importAliasMap[v.funcImport]
		alias := aliasImport + v.funcStruct
		if structsMap[alias] {
			continue
		}
		structsMap[alias] = true
		structs = append(structs, map[string]string{
			"alias":  alias,
			"struct": aliasImport + "." + v.funcStruct,
		})
		if maxLen < len(alias) {
			maxLen = len(alias)
		}
	}
	content := "var (\n"
	for _, vMap := range structs {
		content += "\t" + fullSpan(vMap["alias"], maxLen+1) + vMap["struct"] + "\n"
	}
	content += ")"
	return content
}

func (g *ginHandle) getStructAlias(info routeFuncInfo) string {
	return g.importAliasMap[info.funcImport] + info.funcStruct
}

func (g *ginHandle) generateImport() string {
	importsMap := map[string]bool{}
	aliasMap := map[string]int{}
	content := "import (\n"
	content += "\t\"" + "github.com/gin-gonic/gin" + "\"\n"
	for _, v := range g.routesFunc {
		if importsMap[v.funcImport] {
			continue
		}
		importsMap[v.funcImport] = true
		alias := filepath.Base(v.funcImport)
		newAlias := alias + toString(aliasMap[alias])
		g.importAliasMap[v.funcImport] = newAlias
		content += "\t" + newAlias + " " + "\"" + v.funcImport + "\"\n"
		aliasMap[alias]++
	}
	content += ")"
	return content
}
