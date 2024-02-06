package openapi

import (
	"encoding/json"
	"fmt"
	"github.com/getkin/kin-openapi/openapi3"
	"log"
	"path/filepath"
	"strings"
)

type openapiHandle struct {
	t       *openapi3.T
	structs map[string]*structInfo
	schemas openapi3.Schemas
}

func (o *openapiHandle) load(routeDir, docPath string) {
	o.t = &openapi3.T{
		OpenAPI: "3.0.0",
	}
	o.structs = map[string]*structInfo{}
	o.schemas = map[string]*openapi3.SchemaRef{}
	o.generateDoc(docPath)
	o.generateRoute(routeDir)
}

func (o *openapiHandle) noRepeatStructs() {
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

func (o *openapiHandle) generateRoute(routeDir string) {
	fileList := fileHandle{}
	fileList.load(routeDir, true)
	routes := map[string]map[string]interface{}{}
	for _, filePath := range fileList {
		asts := new(astHandle)
		err := asts.load(filePath, projectModName, astLoadTypeRoute|astLoadTypeStruct)
		if err != nil {
			log.Fatal(err)
		}
		for k, v := range asts.routes {
			if routes[k] != nil {
				log.Fatal("路由重复")
			}
			routes[k] = v
		}
		for k, v := range asts.structs {
			o.structs[k] = v
		}
	}
	if len(routes) == 0 {
		return
	}
	o.noRepeatStructs()
	if o.t.Paths == nil {
		o.t.Paths = &openapi3.Paths{}
	}
	for k, v := range routes {
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
		case "trace":
			if pathItem.Trace != nil {
				operation = pathItem.Trace
			}
		}
		o.setOpenAPIByRoute(operation, v)
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
		case "trace":
			pathItem.Trace = operation
		}
		o.t.Paths.Set(path, pathItem)
	}
	// 设置schemes
	if o.t.Components == nil {
		o.t.Components = &openapi3.Components{}
	}
	o.t.Components.Schemas = o.schemas
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
				in := toString(vMap["in"])
				if body.Value.Content == nil {
					body.Value.Content = map[string]*openapi3.MediaType{}
				}
				mediaType := &openapi3.MediaType{
					Schema: &openapi3.SchemaRef{},
				}
				for k1, v1 := range vMap {
					switch k1 {
					case "content":
						structTitle := toString(v1)
						strInfo := o.structs[structTitle]
						if strInfo == nil {
							tempTitle := strings.TrimPrefix(structTitle, "[]")
							if tempTitle != structTitle {
								strInfo = o.structs[tempTitle]
							}
							if strInfo == nil {
								if mediaType.Schema.Value == nil {
									mediaType.Schema.Value = &openapi3.Schema{}
								}
								mediaType.Schema.Value.Type = "string"
								mediaType.Schema.Value.Format = structTitle
							} else {
								mediaType.Schema.Value = &openapi3.Schema{
									Type: "array",
									Items: &openapi3.SchemaRef{
										Ref: o.setScheme(strInfo),
									},
								}
							}
						} else {
							mediaType.Schema.Ref = o.setScheme(strInfo)
						}
					case "desc":
						body.Value.Description = toString(v1)
					}
				}
				body.Value.Content[in] = mediaType
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
						Value: &openapi3.Response{
							Content: map[string]*openapi3.MediaType{},
						},
					}
					in := toString(v1Map["in"])
					if response.Value.Content == nil {
						response.Value.Content = map[string]*openapi3.MediaType{}
					}
					mediaType := &openapi3.MediaType{
						Schema: &openapi3.SchemaRef{},
					}
					for k2, v2 := range v1Map {
						switch k2 {
						case "content":
							structTitle := toString(v2)
							strInfo := o.structs[structTitle]
							if strInfo == nil {
								tempTitle := strings.TrimPrefix(structTitle, "[]")
								if tempTitle != structTitle {
									strInfo = o.structs[tempTitle]
								}
								if strInfo == nil {
									if mediaType.Schema.Value == nil {
										mediaType.Schema.Value = &openapi3.Schema{}
									}
									mediaType.Schema.Value.Type = "string"
									mediaType.Schema.Value.Format = structTitle
								} else {
									mediaType.Schema.Value = &openapi3.Schema{
										Type: "array",
										Items: &openapi3.SchemaRef{
											Ref: o.setScheme(strInfo),
										},
									}
								}
							} else {
								mediaType.Schema.Ref = o.setScheme(strInfo)
							}
						case "desc":
							response.Value.Description = toPtr(toString(v2))
						}
					}
					response.Value.Content[in] = mediaType
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

				for k1, v1 := range vMap {
					if securitySchemes[k1] == nil {
						var keys []string
						for k2, _ := range securitySchemes {
							keys = append(keys, k2)
						}
						log.Fatal(fmt.Sprintf("验证"+errorNotIn, k1, strings.Join(keys, ",")))
					}
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
				if cutInfo, ok := v.(*strCutInfo); ok {
					valType := o.getType(cutInfo.man)
					val.Value.Schema.Value.Type = valType
					val.Value.Schema.Value.Format = cutInfo.other
					if cutInfo.other == "" && valType != cutInfo.man {
						val.Value.Schema.Value.Format = cutInfo.man
					}
				} else {
					val.Value.Schema.Value.Type = toString(v)
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
				val.Value.Schema.Value.Default = v
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

func (o *openapiHandle) setScheme(strInfo *structInfo) (refUrl string) {
	refUrl = "#/components/schemas/" + strInfo.name
	if o.schemas[strInfo.name] != nil {
		return
	}
	schemaRef := &openapi3.SchemaRef{
		Value: &openapi3.Schema{
			Type:       "object",
			Properties: map[string]*openapi3.SchemaRef{},
		},
	}
	for _, v2 := range strInfo.list {
		fieldName := v2.fieldName
		filedType := o.getType(v2.fieldType)
		fieldSchemaRef := &openapi3.SchemaRef{
			Value: &openapi3.Schema{
				Type:        filedType,
				Description: v2.comment,
			},
		}
		if filedType != v2.fieldType {
			fieldSchemaRef.Value.Format = v2.fieldType
		}
		var requiredList []string
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
				fieldSchemaRef.Value.Default = v3[0]
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
		schemaRef.Value.Required = requiredList
	}
	o.schemas[strInfo.name] = schemaRef
	return
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
