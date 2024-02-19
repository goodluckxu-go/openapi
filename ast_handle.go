package openapi

import (
	"encoding/json"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

type astLoadType int

const (
	astLoadTypeDoc astLoadType = 1 << iota
	astLoadTypeRoute
	astLoadTypeStruct
)

type structField struct {
	fieldName string
	fieldType string
	comment   string
	extends   map[string][]string
}

type structInfo struct {
	name    string
	comment string
	list    []structField
}

type astHandle struct {
	fSet           *token.FileSet
	astFile        *ast.File
	structs        map[string]*structInfo            // 所有结构体
	routes         map[string]map[string]interface{} // 所有路由
	docs           map[string]interface{}            // 所有文档
	importMap      map[string]string
	modName        string
	filePath       string
	structPrefix   string
	uniqueFieldMap map[string]bool
	modDir         string
	sameStructs    map[string]string
}

type strCutInfo struct {
	man   string // 主要
	other string // 其他
}

func (a *astHandle) load(filePath string, modName string, loadType astLoadType, modDir ...string) (err error) {
	filePath, err = filepath.Abs(filePath)
	if err != nil {
		return
	}
	if len(modDir) > 0 {
		a.modDir, _ = filepath.Abs(modDir[0])
	}
	a.sameStructs = map[string]string{}
	a.filePath = filePath
	a.modName = modName
	a.fSet = token.NewFileSet()
	a.astFile, err = parser.ParseFile(a.fSet, filePath, nil, parser.ParseComments)
	if err != nil {
		return a.error(err.Error())
	}
	if loadType&astLoadTypeStruct == astLoadTypeStruct {
		// 解析引入
		a.parseImports()
		// 解析结构体
		a.parseStructs()
	}
	a.uniqueFieldMap = map[string]bool{}
	if loadType&astLoadTypeDoc == astLoadTypeDoc {
		if err = a.parseDoc(); err != nil {
			return
		}
	}
	if loadType&astLoadTypeRoute == astLoadTypeRoute {
		if err = a.parseRoutes(); err != nil {
			return
		}
	}
	return
}

func (a *astHandle) parseRoutes() (err error) {
	if a.astFile.Decls == nil {
		return
	}
	a.routes = map[string]map[string]interface{}{}
	var ok bool
	var funcDecl *ast.FuncDecl
	for _, decl := range a.astFile.Decls {
		if funcDecl, ok = decl.(*ast.FuncDecl); ok {
			var rsMap map[string]interface{}
			rsMap, err = a.parseComments(funcDecl.Doc, validRoutesMap)
			if err != nil {
				return
			}
			var routes []map[string]interface{}
			if routes, ok = rsMap["@router"].([]map[string]interface{}); !ok {
				continue
			}
			for _, routeMap := range routes {
				method := toString(routeMap["method"])
				path := toString(routeMap["path"])
				if method == "" || path == "" {
					continue
				}
				a.routes[path+"_"+method] = rsMap
			}
		}
	}
	return
}

func (a *astHandle) parseDoc() (err error) {
	a.docs = map[string]interface{}{}
	for _, comment := range a.astFile.Comments {
		var rsMap map[string]interface{}
		rsMap, err = a.parseComments(comment, validDocMap)
		if err != nil {
			return
		}
		for key, val := range rsMap {
			a.docs[key] = val
		}
	}
	return
}

func (a *astHandle) parseComments(comment *ast.CommentGroup, validMap map[string]*validStruct) (rsMap map[string]interface{}, err error) {
	if comment == nil {
		return
	}
	var key, value string
	var isMull bool
	var pos token.Pos
	rsMap = map[string]interface{}{}
	for _, v := range comment.List {
		v.Text = remoteAnnotationSymbols(v.Text)
		list := strings.Split(v.Text, firstKeyValueCutSign)
		title := a.remoteAnnotationSymbols(list[0])
		validData := validMap[title]
		if validData == nil {
			if isMull {
				if v.Text == "" {
					value += "\n"
				} else {
					value += v.Text + "\n"
				}
				pos = v.Pos()
			}
			continue
		}
		if isMull {
			if err = a.parseCommentLine(v.Pos(), rsMap, key, a.remoteAnnotationSymbols(value), key, validMap); err != nil {
				return
			}
		}
		key = title
		isMull = false
		value = ""
		other := a.remoteAnnotationSymbols(strings.Join(list[1:], firstKeyValueCutSign))
		if other == multiBorderSign {
			isMull = true
			continue
		}
		if err = a.parseCommentLine(v.Pos(), rsMap, key, other, key, validMap); err != nil {
			return
		}
	}
	if isMull {
		if err = a.parseCommentLine(pos, rsMap, key, a.remoteAnnotationSymbols(value), key, validMap); err != nil {
			return
		}
	}
	return
}

func (a *astHandle) parseCommentLine(
	pos token.Pos,
	rsMap map[string]interface{},
	key, value, validKey string,
	validMap map[string]*validStruct,
) (err error) {
	validData := validMap[validKey]
	if validMap == nil {
		return
	}
	if validData.isUnique {
		uniqueKey := fmt.Sprintf("%v_%v", validKey, value)
		if a.uniqueFieldMap[uniqueKey] {
			err = a.errorPos(fmt.Sprintf(errorRepeat, key, value), pos)
			return
		}
		a.uniqueFieldMap[uniqueKey] = true
	}
	switch validData.valType {
	case validTypeString:
		validTitle := value
		var cutInfo *strCutInfo
		if len(validData.strCutOther) == 2 {
			list := strings.Split(value, validData.strCutOther[0])
			if len(list) > 1 && strings.HasSuffix(list[len(list)-1], validData.strCutOther[1]) {
				validTitle = list[0]
				cutInfo = &strCutInfo{
					man:   list[0],
					other: strings.TrimSuffix(strings.Join(list[1:], validData.strCutOther[0]), validData.strCutOther[1]),
				}
			}
		}
		if len(validData.valEnum) > 0 && inArray(validTitle, validData.valEnum) == -1 {
			err = a.errorPos(fmt.Sprintf(errorNotIn, validTitle, strings.Join(validData.valEnum, ",")), pos)
			return
		}
		if cutInfo != nil {
			rsMap[key] = cutInfo
		} else {
			rsMap[key] = value
		}
	case validTypeInteger:
		if len(validData.valEnum) > 0 && inArray(value, validData.valEnum) == -1 {
			err = a.errorPos(fmt.Sprintf(errorNotIn, value, strings.Join(validData.valEnum, ",")), pos)
			return
		}
		if _, err = strconv.Atoi(value); err != nil {
			err = a.errorPos(err.Error(), pos)
			return
		}
		rsMap[key] = value
	case validTypeBool:
		if len(validData.valEnum) > 0 && inArray(value, validData.valEnum) == -1 {
			err = a.errorPos(fmt.Sprintf(errorNotIn, value, strings.Join(validData.valEnum, ",")), pos)
			return
		}
		if value != "true" && value != "false" {
			err = a.errorPos(fmt.Sprintf(errorType, value, "bool"), pos)
			return
		}
		rsMap[key] = value
	case validTypeJson:
		var rs interface{}
		if err = json.Unmarshal([]byte(value), &rs); err != nil {
			return
		}
		rsMap[key] = rs
	case validTypeArray:
		if validData.cutListSign == "" {
			return
		}
		rs := strings.Split(value, validData.cutListSign)
		for k, v := range rs {
			rs[k] = strings.Trim(v, " ")
		}
		rsMap[key] = rs
	case validTypeMapArray, validTypeMap:
		if validData.cutListSign == "" {
			return
		}
		list := strings.Split(value, validData.cutListSign)
		tmpMap := map[string]interface{}{}
		beforeKey := ""
		for _, v := range list {
			newV := a.remoteAnnotationSymbols(v)
			if validData.cutKeyValSign == "" {
				if len(validData.valEnum) > 0 && inArray(newV, validData.valEnum) == -1 {
					err = a.errorPos(fmt.Sprintf(errorNotIn, newV, strings.Join(validData.valEnum, ",")), pos)
					return
				}
				tmpMap[newV] = "true"
				continue
			}
			vList := strings.Split(newV, validData.cutKeyValSign)
			if len(vList) > 1 {
				title := a.remoteAnnotationSymbols(vList[0])
				childValidTitle := key + "._"
				childValidData := validMap[childValidTitle]
				if childValidData != nil {
					// 所有key通过
					if err = a.parseCommentLine(pos, tmpMap, title, a.remoteAnnotationSymbols(strings.Join(vList[1:],
						validData.cutKeyValSign)), childValidTitle, validMap); err != nil {
						return
					}
					continue
				}
				childValidTitle += "." + title
				childValidData = validMap[childValidTitle]
				if childValidData == nil {
					if beforeKey != "" {
						tmpMap[beforeKey] = fmt.Sprintf("%v%v%v", toString(tmpMap[beforeKey]), validData.cutListSign, v)
					}
					continue
				}
				beforeKey = title
				if err = a.parseCommentLine(pos, tmpMap, title, a.remoteAnnotationSymbols(strings.Join(vList[1:],
					validData.cutKeyValSign)), childValidTitle, validMap); err != nil {
					return
				}
			} else {
				if len(validData.valEnum) > 0 && inArray(newV, validData.valEnum) == -1 {
					err = a.errorPos(fmt.Sprintf(errorNotIn, newV, strings.Join(validData.valEnum, ",")), pos)
					return
				}
				tmpMap[newV] = "true"
			}
		}
		if validData.valType == validTypeMap {
			rsMap[key] = tmpMap
		} else {
			rsList, _ := rsMap[key].([]map[string]interface{})
			if len(tmpMap) > 0 {
				rsList = append(rsList, tmpMap)
			}
			rsMap[key] = rsList
		}
	}
	return
}

func (a *astHandle) parseImports() {
	if a.astFile.Decls == nil {
		return
	}
	var ok bool
	var genDecl *ast.GenDecl
	var importSpec *ast.ImportSpec
	importMap := map[string]string{}
	for _, decl := range a.astFile.Decls {
		if genDecl, ok = decl.(*ast.GenDecl); ok && genDecl.Tok.String() == "import" {
			for _, spec := range genDecl.Specs {
				if importSpec, ok = spec.(*ast.ImportSpec); ok {
					importPath := strings.Trim(importSpec.Path.Value, "\"|`")
					importPathList := strings.Split(importPath, "/")
					importName := importPathList[len(importPathList)-1]
					if importSpec.Name != nil {
						importName = importSpec.Name.Name
					}
					importMap[importName] = importPath
				}
			}
		}
	}
	a.importMap = importMap
}

func (a *astHandle) parseStructs() {
	if a.astFile.Decls == nil {
		return
	}
	a.structImport()
	a.structs = map[string]*structInfo{}
	var ok bool
	var genDecl *ast.GenDecl
	var typeSpce *ast.TypeSpec
	for _, decl := range a.astFile.Decls {
		if genDecl, ok = decl.(*ast.GenDecl); ok && genDecl.Tok.String() == "type" {
			if typeSpce, ok = genDecl.Specs[0].(*ast.TypeSpec); ok {
				strInfo := &structInfo{}
				if strInfo, ok = a.parseStruct(typeSpce); !ok {
					continue
				}
				if genDecl.Doc != nil {
					strInfo.comment = genDecl.Doc.Text()
				}
				strName := strInfo.name
				strInfo.name = strings.ReplaceAll(a.structPrefix+strName, "/", ".")
				a.structs[a.structPrefix+strName] = strInfo
			}
		}
	}
	return
}

func (a *astHandle) structImport() {
	if a.modName == "" {
		return
	}
	pwd := a.modDir
	if pwd == "" {
		pwd, _ = os.Getwd()
	}
	a.structPrefix = strings.TrimPrefix(a.filePath, pwd)
	a.structPrefix = filepath.Dir(a.structPrefix)
	a.structPrefix = filepath.Join(a.modName, a.structPrefix)
	a.structPrefix = strings.ReplaceAll(a.structPrefix, "\\", "/") + "."
}

func (a *astHandle) parseStruct(typeSpec *ast.TypeSpec) (strInfo *structInfo, bl bool) {
	var ok bool
	strInfo = &structInfo{}
	if typeSpec.Name == nil {
		return
	}
	strInfo.name = typeSpec.Name.Name
	if typeSpec.Type == nil {
		return
	}
	var structType *ast.StructType
	if structType, ok = typeSpec.Type.(*ast.StructType); !ok {
		sameTypes := a.getCallType(typeSpec.Type)
		if sameTypes == "" || sameTypes == "interface{}" {
			return
		}
		a.sameStructs[a.structPrefix+strInfo.name] = a.getCallType(typeSpec.Type)
		return
	}
	for _, field := range structType.Fields.List {
		fieldInfo := structField{}
		// 获取名称
		fieldName := ""
		if len(field.Names) > 0 {
			fieldName = field.Names[0].Name
		}
		fieldInfo.fieldName = fieldName
		// 获取类型
		fieldInfo.fieldType = a.getCallType(field.Type)
		// 获取标签
		if field.Tag != nil {
			rsMap := a.getCallTags(field.Tag)
			if rsMap["xml"] != nil {
				rsList, _ := rsMap["xml"].([]string)
				fieldInfo.fieldName = rsList[0]
				delete(rsMap, "xml")
			}
			if rsMap["json"] != nil {
				rsList, _ := rsMap["json"].([]string)
				fieldInfo.fieldName = rsList[0]
				delete(rsMap, "json")
			}
			// 扩展extends
			if len(rsMap) > 0 {
				fieldInfo.extends = map[string][]string{}
			}
			if rsMap["openapi"] != nil {
				switch rsVal := rsMap["openapi"].(type) {
				case []string:
					fieldInfo.extends[rsVal[0]] = []string{"true"}
				case map[string][]string:
					for k1, v1 := range rsVal {
						fieldInfo.extends[k1] = v1
					}
				}
				delete(rsMap, "openapi")
			}
			for k1, v1 := range rsMap {
				v1List, _ := v1.([]string)
				if len(v1List) == 0 {
					continue
				}
				fieldInfo.extends[k1] = v1List
			}
		}
		if fieldInfo.fieldName == "-" {
			continue
		}
		// 获取注释
		if field.Comment != nil {
			fieldInfo.comment = a.remoteAnnotationSymbols(field.Comment.List[0].Text)
		}
		strInfo.list = append(strInfo.list, fieldInfo)
	}
	bl = true
	return
}

func (a *astHandle) getCallTags(expr ast.Expr) (rsMap map[string]interface{}) {
	rsMap = make(map[string]interface{})
	switch val := expr.(type) {
	case *ast.BasicLit:
		reg := regexp.MustCompile(`([a-zA-Z_][a-zA-Z0-9_]*)( \t)*:( \t)*"(.*?[^\\])"`)
		list := reg.FindAllStringSubmatch(val.Value, -1)
		for _, v := range list {
			vList := strings.Split(v[4], secondListCutSign)
			// 单个属性
			if len(vList) == 1 {
				valList := strings.Split(v[4], thirdListCutSign)
				for k1, v1 := range valList {
					valList[k1] = strings.Trim(v1, " ")
				}
				rsMap[v[1]] = valList
				continue
			}
			valMap := map[string][]string{}
			for _, v1 := range vList {
				v1 = strings.Trim(v1, " ")
				eqList := strings.Split(v1, secondKeyValueCutSign)
				if len(eqList) == 1 {
					valMap[v1] = []string{"true"}
				} else {
					valList := strings.Split(strings.Trim(strings.Join(eqList[1:], secondKeyValueCutSign), " "), thirdListCutSign)
					for k2, v2 := range valList {
						valList[k2] = strings.Trim(v2, " ")
					}
					valMap[strings.Trim(eqList[0], " ")] = valList
				}
			}
			rsMap[v[1]] = valMap
		}
	}
	return
}

func (a *astHandle) getCallType(expr ast.Expr) string {
	var ok bool
	switch val := expr.(type) {
	case *ast.Ident:
		// 常规类型
		switch val.Name {
		case "int", "int8", "int16", "int32", "int64",
			"uint", "uint8", "uint16", "uint32", "uint64",
			"float32", "float64", "string", "bool":
			return val.Name
		}
		return a.structPrefix + val.Name
	case *ast.ArrayType:
		// 数组类型
		rs := "["
		if val.Len != nil {
			var baseList *ast.BasicLit
			if baseList, ok = val.Len.(*ast.BasicLit); ok {
				rs += baseList.Value
			}
		}
		rs += "]" + a.getCallType(val.Elt)
		return rs
	case *ast.MapType:
		// map类型
		return "map[" + a.getCallType(val.Key) + "]" + a.getCallType(val.Value)
	case *ast.InterfaceType:
		// interface类型
		return "interface{}"
	case *ast.SelectorExpr:
		// 引用类型
		var xTypeExpr *ast.Ident
		if xTypeExpr, ok = val.X.(*ast.Ident); !ok {
			return ""
		}
		if a.importMap[xTypeExpr.Name] != "" {
			xTypeExpr.Name = a.importMap[xTypeExpr.Name]
		}
		return xTypeExpr.Name + "." + val.Sel.Name
	case *ast.StarExpr:
		// 该项目指针类型使用原类型
		return a.getCallType(val.X)
	}
	return ""
}

func (a *astHandle) remoteAnnotationSymbols(s string) string {
	s = strings.Trim(s, " \t\n")
	s = strings.TrimPrefix(s, "//")
	s = strings.TrimPrefix(s, "/*")
	s = strings.TrimSuffix(s, "*/")
	s = strings.Trim(s, " \t\n")
	return s
}

func (a *astHandle) errorPos(err string, pos token.Pos) error {
	return fmt.Errorf("%v: %v", a.fSet.Position(pos), err)
}

func (a *astHandle) error(err string) error {
	return fmt.Errorf("%v", err)
}
