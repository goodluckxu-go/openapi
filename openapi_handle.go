package openapi

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/getkin/kin-openapi/openapi3"
	"log"
	"path/filepath"
	"strconv"
	"strings"
)

type openapiHandle struct {
	t             *openapi3.T
	structs       map[string]*structInfo
	routesFunc    []routeFuncInfo
	schemas       openapi3.Schemas
	importStructs map[string]bool
	sameStructs   map[string]string
	globalRoutes  map[string]interface{}
}

func (o *openapiHandle) load(rootDir, routeDir, docPath string) {
	o.t = &openapi3.T{
		OpenAPI: Version,
	}
	o.structs = map[string]*structInfo{}
	o.schemas = map[string]*openapi3.SchemaRef{}
	o.importStructs = map[string]bool{}
	o.sameStructs = map[string]string{}
	o.globalRoutes = map[string]interface{}{}
	o.generateDoc(docPath)
	o.generateRoute(rootDir, routeDir)
	if err := o.t.Validate(context.Background()); err != nil {
		log.Fatal(err)
	}
}

func (o *openapiHandle) handleRootDirStructs(rootDir string) {
	fileList := fileHandle{}
	fileList.load(rootDir)
	for _, filePath := range fileList {
		asts := new(astHandle)
		err := asts.load(filePath, projectModName, astLoadTypeStruct)
		if err != nil {
			log.Fatal(err)
		}
		for k, v := range asts.structs {
			o.structs[k] = v
		}
		for k, v := range asts.sameStructs {
			o.sameStructs[k] = v
		}
	}
	// 项目中不添加mod名称的引入，结构体的包+结构体名称不能出现重复，否则原样输出
	repeatStructs := map[string][]string{}
	for k, _ := range o.structs {
		if strings.HasPrefix(k, projectModName) {
			repeatStructs[filepath.Base(k)] = append(repeatStructs[filepath.Base(k)], k)
		}
	}
	for k, v := range repeatStructs {
		if len(v) > 1 {
			continue
		}
		o.structs[k] = o.structs[v[0]]
	}
}

func (o *openapiHandle) addImportStruct(v interface{}) {
	vMap, _ := v.(map[string]interface{})
	bodyMap, _ := vMap["@body"].(map[string]interface{})
	content, _ := bodyMap["content"].(string)
	if content != "" {
		o.importStructs[strings.TrimPrefix(content, "[]")] = true
	}
	resList, _ := vMap["@res"].([]map[string]interface{})
	for _, v1Map := range resList {
		content, _ = v1Map["content"].(string)
		if content != "" {
			o.importStructs[strings.TrimPrefix(content, "[]")] = true
		}
	}
}

func (o *openapiHandle) handleImportStruct() {
	// 注释中使用的结构体解析，去掉已经解析的结构体
	for k, _ := range o.importStructs {
		if o.structs[k] != nil {
			delete(o.importStructs, k)
		}
	}
	// 对比mod文件获取引入文件
	fileMap := map[string]bool{}
	for k, _ := range modPathMap {
		for k1, _ := range o.importStructs {
			other := strings.TrimPrefix(k1, k)
			if other == k1 || fileMap[k] {
				continue
			}
			fileMap[k] = true
		}
	}
	fileModList := map[string][]string{}
	for k, _ := range fileMap {
		files := fileHandle{}
		files.load(modPathMap[k])
		fileModList[k] = append(fileModList[k], files...)
	}
	for k, vList := range fileModList {
		for _, v1 := range vList {
			structHandle := new(astHandle)
			_ = structHandle.load(v1, k, astLoadTypeStruct, modPathMap[k])
			for k2, v2 := range structHandle.structs {
				o.structs[k2] = v2
			}
			for k2, v2 := range structHandle.sameStructs {
				o.sameStructs[k2] = v2
			}
		}
	}
}

func (o *openapiHandle) handleNoStructFieldName() {
	inBool := func(v structField, list []structField) bool {
		for i := range list {
			if list[i].fieldName == v.fieldName {
				return true
			}
		}
		return false
	}
	for k, v := range o.structs {
		var fieldNameList []structField
		for _, fieldInfo := range v.list {
			if fieldInfo.fieldName == "" {
				childStruct := o.structs[fieldInfo.fieldType]
				if childStruct != nil {
					for _, v1 := range childStruct.list {
						if !inBool(v1, fieldNameList) {
							fieldNameList = append(fieldNameList, v1)
						}
					}
				}
			} else {
				fieldNameList = append(fieldNameList, fieldInfo)
			}
		}
		o.structs[k].list = fieldNameList
	}
}

func (o *openapiHandle) generateRoute(rootDir, routeDir string) {
	fileList := fileHandle{}
	fileList.load(routeDir)
	routes := map[string]map[string]interface{}{}
	for _, filePath := range fileList {
		asts := new(astHandle)
		err := asts.load(filePath, projectModName, astLoadTypeRoute)
		if err != nil {
			log.Fatal(err)
		}
		for k, v := range asts.routes {
			if routes[k] != nil {
				log.Fatal("路由重复")
			}
			routes[k] = v
			// 增加路由引入结构体
			o.addImportStruct(v)
		}
		o.routesFunc = append(o.routesFunc, asts.routesFunc...)
	}
	if len(routes) == 0 {
		return
	}
	o.handleRootDirStructs(rootDir)
	o.handleImportStruct()
	o.handleNoStructFieldName()
	if o.t.Paths == nil {
		o.t.Paths = &openapi3.Paths{}
	}
	for k, vMap := range routes {
		vList := strings.Split(k, "_")
		path := strings.Join(vList[:len(vList)-1], "_")
		method := vList[len(vList)-1]
		pathItem := &openapi3.PathItem{}
		if o.t.Paths.Value(path) != nil {
			pathItem = o.t.Paths.Value(path)
		}
		operation := &openapi3.Operation{}
		switch method {
		case "get":
			if pathItem.Get != nil {
				operation = pathItem.Get
			}
		case "put":
			if pathItem.Put != nil {
				operation = pathItem.Put
			}
		case "post":
			if pathItem.Post != nil {
				operation = pathItem.Post
			}
		case "delete":
			if pathItem.Delete != nil {
				operation = pathItem.Delete
			}
		case "options":
			if pathItem.Options != nil {
				operation = pathItem.Options
			}
		case "head":
			if pathItem.Head != nil {
				operation = pathItem.Head
			}
		case "patch":
			if pathItem.Patch != nil {
				operation = pathItem.Patch
			}
		}
		// 处理通用路由
		for k1, v1 := range o.globalRoutes {
			if vMap[k1] != nil {
				switch setData := vMap[k1].(type) {
				case []map[string]interface{}:
					v1List, _ := v1.([]map[string]interface{})
					setData = append(setData, v1List...)
					vMap[k1] = setData
				case map[string]interface{}:
					v1Map, _ := v1.(map[string]interface{})
					for k2, v2 := range v1Map {
						setData[k2] = v2
					}
					vMap[k1] = setData
				}
			} else {
				vMap[k1] = v1
			}
		}
		o.handleResponse(vMap)
		o.setOpenAPIByRoute(operation, vMap)
		switch method {
		case "get":
			pathItem.Get = operation
		case "put":
			pathItem.Put = operation
		case "post":
			pathItem.Post = operation
		case "delete":
			pathItem.Delete = operation
		case "options":
			pathItem.Options = operation
		case "head":
			pathItem.Head = operation
		case "patch":
			pathItem.Patch = operation
		}
		o.t.Paths.Set(path, pathItem)
	}
	// 设置schemes
	if o.t.Components == nil {
		o.t.Components = &openapi3.Components{}
	}
	o.t.Components.Schemas = o.schemas
}

func (o *openapiHandle) handleResponse(dataMap map[string]interface{}) {
	resList, _ := dataMap["@res"].([]map[string]interface{})
	isStatusOK := false
	for _, resMap := range resList {
		if resMap["status"] == "200" {
			isStatusOK = true
		}
	}
	// 必定存在200状态
	if !isStatusOK {
		resList = append(resList, map[string]interface{}{
			"in":      "text/plain",
			"status":  "200",
			"desc":    "Success",
			"content": "",
		})
		dataMap["@res"] = resList
	}
}

func (o *openapiHandle) setOpenAPIByRoute(dist any, dataMap map[string]interface{}) {
	switch val := dist.(type) {
	case *openapi3.Operation:
		for k, v := range dataMap {
			switch k {
			case "@summary":
				val.Summary = toString(v)
			case "@description":
				val.Description = toString(v)
			case "@tags":
				var tags []string
				vMap, _ := v.(map[string]interface{})
				for k1, _ := range vMap {
					tags = append(tags, k1)
				}
				val.Tags = tags
			case "@param":
				params := openapi3.Parameters{}
				if val.Parameters != nil {
					params = val.Parameters
				}
				vList, _ := v.([]map[string]interface{})
				for _, v1Map := range vList {
					param := &openapi3.ParameterRef{}
					o.setOpenAPIByRoute(param, v1Map)
					params = append(params, param)
				}
				val.Parameters = params
			case "@body":
				body := &openapi3.RequestBodyRef{}
				if val.RequestBody != nil {
					body = val.RequestBody
				}
				if body.Value == nil {
					body.Value = &openapi3.RequestBody{}
				}
				vMap, _ := v.(map[string]interface{})
				ins, _ := vMap["in"].([]string)
				if len(ins) > 0 {
					body.Value.Content = map[string]*openapi3.MediaType{}
					for _, in := range ins {
						mediaType := &openapi3.MediaType{
							Schema: &openapi3.SchemaRef{},
						}
						if vMap["content"] != nil {
							o.setType(mediaType.Schema, toString(vMap["content"]), true)
						}
						body.Value.Content[in] = mediaType
					}
				}
				for k1, v1 := range vMap {
					switch k1 {
					case "desc":
						body.Value.Description = toString(v1)
					}
				}
				val.RequestBody = body
			case "@res":
				responses := &openapi3.Responses{}
				if val.Responses != nil {
					responses = val.Responses
				}
				vList, _ := v.([]map[string]interface{})
				for _, v1Map := range vList {
					status := toString(v1Map["status"])
					response := &openapi3.ResponseRef{
						Value: &openapi3.Response{},
					}
					ins, _ := v1Map["in"].([]string)
					if len(ins) > 0 {
						response.Value.Content = map[string]*openapi3.MediaType{}
						for _, in := range ins {
							mediaType := &openapi3.MediaType{
								Schema: &openapi3.SchemaRef{},
							}
							if v1Map["content"] != nil {
								o.setType(mediaType.Schema, toString(v1Map["content"]), true)
							}
							response.Value.Content[in] = mediaType
						}
					}
					for k2, v2 := range v1Map {
						switch k2 {
						case "desc":
							response.Value.Description = toPtr(toString(v2))
						}
					}
					responses.Set(status, response)
				}
				val.Responses = responses
			case "@security":
				vMap, _ := v.(map[string]interface{})
				securitySchemes := o.t.Components.SecuritySchemes
				securitys := openapi3.SecurityRequirements{}
				if val.Security != nil {
					securitys = *val.Security
				}
				sorts, _ := vMap[sortField].([]string)
				for _, k1 := range sorts {
					if securitySchemes[k1] == nil {
						var keys []string
						for k2, _ := range securitySchemes {
							keys = append(keys, k2)
						}
						log.Fatal(fmt.Sprintf("验证"+errorNotIn, k1, strings.Join(keys, ",")))
					}
					v1 := vMap[k1]
					security := openapi3.SecurityRequirement{}
					if v1 == "true" {
						security[k1] = []string{}
					} else {
						v1List, _ := v1.([]string)
						security[k1] = v1List
					}
					securitys = append(securitys, security)
				}
				val.Security = &securitys
			}
		}
	case *openapi3.ParameterRef:
		if val.Value == nil {
			val.Value = &openapi3.Parameter{}
		}
		for k, v := range dataMap {
			switch k {
			case "in":
				val.Value.In = toString(v)
			case "name":
				val.Value.Name = toString(v)
			case "type":
				if val.Value.Schema == nil {
					val.Value.Schema = &openapi3.SchemaRef{
						Value: &openapi3.Schema{},
					}
				}
				valType := toString(v)
				val.Value.Schema.Value.Type = o.getType(valType)
				if val.Value.Schema.Value.Type != valType {
					val.Value.Schema.Value.Format = valType
				}
			case "required":
				if v == "true" {
					val.Value.Required = true
				}
			case "desc":
				val.Value.Description = toString(v)
			case "minimum":
				if val.Value.Schema == nil {
					val.Value.Schema = &openapi3.SchemaRef{
						Value: &openapi3.Schema{},
					}
				}
				val.Value.Schema.Value.Min = toPtr(toFloat64(v))
			case "maximum":
				if val.Value.Schema == nil {
					val.Value.Schema = &openapi3.SchemaRef{
						Value: &openapi3.Schema{},
					}
				}
				val.Value.Schema.Value.Max = toPtr(toFloat64(v))
			case "minLength":
				if val.Value.Schema == nil {
					val.Value.Schema = &openapi3.SchemaRef{
						Value: &openapi3.Schema{},
					}
				}
				val.Value.Schema.Value.MinLength = toUint64(v)
			case "maxLength":
				if val.Value.Schema == nil {
					val.Value.Schema = &openapi3.SchemaRef{
						Value: &openapi3.Schema{},
					}
				}
				val.Value.Schema.Value.MaxLength = toPtr(toUint64(v))
			case "example":
				if val.Value.Schema == nil {
					val.Value.Schema = &openapi3.SchemaRef{
						Value: &openapi3.Schema{},
					}
				}
				val.Value.Schema.Value.Example = v
			case "default":
				if val.Value.Schema == nil {
					val.Value.Schema = &openapi3.SchemaRef{
						Value: &openapi3.Schema{},
					}
				}
				val.Value.Schema.Value.Default = o.getTypeValue(toString(dataMap["type"]), toString(v))
			case "enum":
				if val.Value.Schema == nil {
					val.Value.Schema = &openapi3.SchemaRef{
						Value: &openapi3.Schema{},
					}
				}
				vList, _ := v.([]string)
				val.Value.Schema.Value.Enum = toSliceInterface(vList)
			}
		}
	}
}

func (o *openapiHandle) setType(schemeRef *openapi3.SchemaRef, types string, isContent bool, alreadyMaps ...map[string]int) {
	alreadyMap := map[string]int{}
	if len(alreadyMaps) > 0 {
		alreadyMap = alreadyMaps[0]
	}
	if schemeRef == nil {
		schemeRef = &openapi3.SchemaRef{}
	}
	if schemeRef.Value == nil {
		schemeRef.Value = &openapi3.Schema{}
	}
	if alreadyMap[types] > 0 {
		// 第一次重复需要设置ref值
		if alreadyMap[types] == 1 {
			// 克隆map去掉影响
			tempAlreadyMap := cloneMap(alreadyMaps[0])
			tempAlreadyMap[types]++
			schemeRef.Ref = o.setScheme(o.structs[types], tempAlreadyMap)
		}
		return
	}
	if o.sameStructs[types] != "" {
		o.setType(schemeRef, o.sameStructs[types], false, alreadyMap)
		return
	}
	tempTypes := ""
	// 判断是否是数组
	tempTypes = strings.TrimPrefix(types, "[]")
	if tempTypes != types {
		types = tempTypes
		schemeRef.Value.Type = "array"
		schemeRef.Value.Items = &openapi3.SchemaRef{}
		o.setType(schemeRef.Value.Items, types, false, alreadyMap)
		return
	}
	// 判断是否是对象
	tempTypes = strings.TrimPrefix(types, "map[")
	if tempTypes != types {
		mapTypes := ""
		mapTypes, types = getIndexFirst(tempTypes, "]")
		mapTypes = strings.ReplaceAll(mapTypes, "/", ".")
		schemeRef.Value.Type = "object"
		schemeRef.Value.Properties = map[string]*openapi3.SchemaRef{
			mapTypes: {},
		}
		o.setType(schemeRef.Value.Properties[mapTypes], types, false, alreadyMap)
		return
	}
	strInfo := o.structs[types]
	if strInfo == nil {
		schemeRef.Value.Type = o.getType(types)
		if isContent {
			schemeRef.Value.Default = types
		} else if schemeRef.Value.Type != types {
			schemeRef.Value.Format = types
		}
	} else {
		alreadyMap[types]++
		schemeRef.Ref = o.setScheme(strInfo, alreadyMap)
	}

}

func (o *openapiHandle) setScheme(strInfo *structInfo, alreadyMaps ...map[string]int) (refUrl string) {
	alreadyMap := map[string]int{}
	if len(alreadyMaps) > 0 {
		alreadyMap = alreadyMaps[0]
	}
	refUrl = "#/components/schemas/" + strInfo.name
	if o.schemas[strInfo.name] != nil {
		return
	}
	schemaRef := &openapi3.SchemaRef{
		Value: &openapi3.Schema{
			Type:        "object",
			Description: strInfo.comment,
			Properties:  map[string]*openapi3.SchemaRef{},
			XML: &openapi3.XML{
				Name: strings.TrimPrefix(filepath.Ext(strInfo.name), "."),
			},
		},
	}
	var requiredList []string
	for _, v2 := range strInfo.list {
		fieldName := v2.fieldName
		fieldSchemaRef := &openapi3.SchemaRef{
			Value: &openapi3.Schema{
				Description: v2.comment,
			},
		}
		o.setType(fieldSchemaRef, v2.fieldType, false, alreadyMap)
		for k3, v3 := range v2.extends {
			switch k3 {
			case "minimum":
				// 数字验证，最小值
				fieldSchemaRef.Value.Min = toPtr(toFloat64(v3[0]))
			case "maximum":
				// 数字验证，最大值
				fieldSchemaRef.Value.Max = toPtr(toFloat64(v3[0]))
			case "minLength":
				// 字符串验证，最小长度
				fieldSchemaRef.Value.MinLength = toUint64(v3[0])
			case "maxLength":
				// 字符串验证，最大长度
				fieldSchemaRef.Value.MaxLength = toPtr(toUint64(v3[0]))
			case "minItems":
				// 数组验证，最小长度
				fieldSchemaRef.Value.MinItems = toUint64(v3[0])
			case "maxItems":
				// 数组验证，最大长度
				fieldSchemaRef.Value.MaxItems = toPtr(toUint64(v3[0]))
			case "example":
				// 实例
				fieldSchemaRef.Value.Example = v3[0]
			case "default":
				// 默认值
				fieldSchemaRef.Value.Default = o.getTypeValue(v2.fieldType, v3[0])
			case "enum":
				// 限定值
				fieldSchemaRef.Value.Enum = toSliceInterface(v3)
			case "required":
				if v3[0] == "true" {
					requiredList = append(requiredList, fieldName)
				}
			}
		}
		schemaRef.Value.Properties[fieldName] = fieldSchemaRef
	}
	schemaRef.Value.Required = requiredList
	o.schemas[strInfo.name] = schemaRef
	return
}

func (o *openapiHandle) getTypeValue(types string, value string) (rs interface{}) {
	rs = value
	types = o.getType(types)
	switch types {
	case "integer":
		rs, _ = strconv.ParseInt(value, 10, 64)
	case "number":
		rs, _ = strconv.ParseFloat(value, 64)
	case "boolean":
		rs, _ = strconv.ParseBool(value)
	}
	return rs
}

func (o *openapiHandle) getType(s string) string {
	switch s {
	case "int", "int8", "int16", "int32", "int64", "uint", "uint8", "uint16", "uint32", "uint64":
		return "integer"
	case "float32", "float64":
		return "number"
	case "bool":
		return "boolean"
	case "integer", "number", "string", "boolean":
		return s
	}
	return "string"
}

func (o *openapiHandle) generateDoc(docPath string) {
	asts := new(astHandle)
	err := asts.load(docPath, projectModName, astLoadTypeDoc)
	if err != nil {
		log.Fatal(err)
	}
	o.setOpenAPIByDoc(o.t, asts.docs)
	// 处理通用路由
	o.globalRoutes["@res"] = asts.docs["@global.res"]
	o.globalRoutes["@param"] = asts.docs["@global.param"]
}

func (o *openapiHandle) setOpenAPIByDoc(dist any, dataMap map[string]interface{}) {
	switch val := dist.(type) {
	case *openapi3.T:
		for k, v := range dataMap {
			title, other := getIndexFirst(k, ".")
			switch title {
			case "@info":
				info := &openapi3.Info{}
				if val.Info != nil {
					info = val.Info
				}
				o.setOpenAPIByDoc(info, map[string]interface{}{
					other: v,
				})
				val.Info = info
			case "@externalDocs":
				externalDocs := &openapi3.ExternalDocs{}
				if val.ExternalDocs != nil {
					externalDocs = val.ExternalDocs
				}
				o.setOpenAPIByDoc(externalDocs, map[string]interface{}{
					other: v,
				})
				val.ExternalDocs = externalDocs
			case "@servers":
				servers := openapi3.Servers{}
				if val.Servers != nil {
					servers = val.Servers
				}
				vList, _ := v.([]map[string]interface{})
				for _, v1Map := range vList {
					server := &openapi3.Server{}
					o.setOpenAPIByDoc(server, v1Map)
					servers = append(servers, server)
				}
				val.Servers = servers
			case "@tags":
				tags := openapi3.Tags{}
				if val.Tags != nil {
					tags = val.Tags
				}
				vList, _ := v.([]map[string]interface{})
				for _, v1Map := range vList {
					tag := &openapi3.Tag{}
					o.setOpenAPIByDoc(tag, v1Map)
					tags = append(tags, tag)
				}
				val.Tags = tags
			case "@components":
				components := &openapi3.Components{}
				if val.Components != nil {
					components = val.Components
				}
				switch other {
				case "securitySchemes":
					if components.SecuritySchemes == nil {
						components.SecuritySchemes = map[string]*openapi3.SecuritySchemeRef{}
					}
					vList, _ := v.([]map[string]interface{})
					for _, vMap := range vList {
						field := toString(vMap["field"])
						securityScheme := &openapi3.SecuritySchemeRef{
							Value: &openapi3.SecurityScheme{},
						}
						for k1, v1 := range vMap {
							switch k1 {
							case "type":
								securityScheme.Value.Type = toString(v1)
							case "scheme":
								securityScheme.Value.Scheme = toString(v1)
							case "bearerFormat":
								securityScheme.Value.BearerFormat = toString(v1)
							case "name":
								securityScheme.Value.Name = toString(v1)
							case "in":
								securityScheme.Value.In = toString(v1)
							case "flows":
								buf, err := json.Marshal(v1)
								if err != nil {
									continue
								}
								flows := openapi3.OAuthFlows{}
								err = json.Unmarshal(buf, &flows)
								if err != nil {
									continue
								}
								securityScheme.Value.Flows = &flows
							}
						}
						components.SecuritySchemes[field] = securityScheme
					}
				}
				val.Components = components
			}
		}
	case *openapi3.Info:
		for k, v := range dataMap {
			title, other := getIndexFirst(k, ".")
			switch title {
			case "title":
				val.Title = toString(v)
			case "description":
				val.Description = toString(v)
			case "termsOfService":
				val.TermsOfService = toString(v)
			case "contact":
				contact := &openapi3.Contact{}
				if val.Contact != nil {
					contact = val.Contact
				}
				o.setOpenAPIByDoc(contact, map[string]interface{}{
					other: v,
				})
				val.Contact = contact
			case "license":
				license := &openapi3.License{}
				if val.License != nil {
					license = val.License
				}
				o.setOpenAPIByDoc(license, map[string]interface{}{
					other: v,
				})
				val.License = license
			case "version":
				val.Version = toString(v)
			}
		}
	case *openapi3.ExternalDocs:
		for k, v := range dataMap {
			switch k {
			case "url":
				val.URL = toString(v)
			case "description":
				val.Description = toString(v)
			}
		}
	case *openapi3.Contact:
		for k, v := range dataMap {
			switch k {
			case "name":
				val.Name = toString(v)
			case "url":
				val.URL = toString(v)
			case "email":
				val.Email = toString(v)
			}
		}
	case *openapi3.License:
		for k, v := range dataMap {
			switch k {
			case "name":
				val.Name = toString(v)
			case "url":
				val.URL = toString(v)
			}
		}
	case *openapi3.Server:
		for k, v := range dataMap {
			switch k {
			case "url":
				val.URL = toString(v)
			case "description":
				val.Description = toString(v)
			}
		}
	case *openapi3.Tag:
		for k, v := range dataMap {
			switch k {
			case "name":
				val.Name = toString(v)
			case "description":
				val.Description = toString(v)
			}
		}
	}
}
