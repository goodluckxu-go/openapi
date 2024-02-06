package openapi

type validStruct struct {
	valType       string   // 类型
	cutListSign   string   // 列表截取标志
	cutKeyValSign string   // 键值截取标志
	valEnum       []string // 枚举验证
	isUnique      bool     // 是否唯一
	strCutOther   []string // 字符串切割其他值 0-左边界,1-右边界
}

var (
	validDocMap = map[string]*validStruct{
		// info
		"@info.title":          {valType: "string"},
		"@info.description":    {valType: "string"},
		"@info.termsOfService": {valType: "string"},
		"@info.contact.name":   {valType: "string"},
		"@info.contact.url":    {valType: "string"},
		"@info.contact.email":  {valType: "string"},
		"@info.license.name":   {valType: "string"},
		"@info.license.url":    {valType: "string"},
		"@info.version":        {valType: "string"},
		// externalDocs
		"@externalDocs.description": {valType: "string"},
		"@externalDocs.url":         {valType: "string"},
		// servers
		"@servers":               {valType: "array", cutListSign: secondListCutSign, cutKeyValSign: secondKeyValueCutSign},
		"@servers._.url":         {valType: "string"},
		"@servers._.description": {valType: "string"},
		// tags
		"@tags":               {valType: "array", cutListSign: secondListCutSign, cutKeyValSign: secondKeyValueCutSign},
		"@tags._.name":        {valType: "string"},
		"@tags._.description": {valType: "string"},
		// components
		"@components.securitySchemes":                {valType: "array", cutListSign: secondListCutSign, cutKeyValSign: secondKeyValueCutSign},
		"@components.securitySchemes._.field":        {valType: "string", isUnique: true},
		"@components.securitySchemes._.type":         {valType: "string", valEnum: []string{"apiKey", "http", "oauth2"}},
		"@components.securitySchemes._.scheme":       {valType: "string"},
		"@components.securitySchemes._.bearerFormat": {valType: "string"},
		"@components.securitySchemes._.name":         {valType: "string"},
		"@components.securitySchemes._.in":           {valType: "string", valEnum: []string{"query", "header", "cookie"}},
		"@components.securitySchemes._.flows":        {valType: "json"},
	}

	validRoutesMap = map[string]*validStruct{
		"@summary":     {valType: "string"},
		"@description": {valType: "string"},
		"@tags":        {valType: "map", cutListSign: secondListCutSign},
		// param
		"@param":             {valType: "array", cutListSign: secondListCutSign, cutKeyValSign: secondKeyValueCutSign, valEnum: []string{"required"}},
		"@param._.in":        {valType: "string", valEnum: []string{"query", "header", "path", "cookie"}},
		"@param._.name":      {valType: "string"},
		"@param._.type":      {valType: "string", strCutOther: []string{"(", ")"}, valEnum: []string{"integer", "number", "string", "boolean"}},
		"@param._.required":  {valType: "bool"},
		"@param._.desc":      {valType: "string"},
		"@param._.minimum":   {valType: "integer"},
		"@param._.maximum":   {valType: "integer"},
		"@param._.minLength": {valType: "integer"},
		"@param._.maxLength": {valType: "integer"},
		"@param._.example":   {valType: "string"},
		"@param._.default":   {valType: "string"},
		"@param._.enum":      {valType: "list", cutListSign: thirdListCutSign},
		// body
		"@body":           {valType: "map", cutListSign: secondListCutSign, cutKeyValSign: secondKeyValueCutSign},
		"@body._.in":      {valType: "string", valEnum: []string{"application/json", "application/xml", "application/x-www-form-urlencoded"}},
		"@body._.content": {valType: "string"},
		"@body._.desc":    {valType: "string"},
		// res
		"@res":           {valType: "array", cutListSign: secondListCutSign, cutKeyValSign: secondKeyValueCutSign},
		"@res._.status":  {valType: "integer"},
		"@res._.in":      {valType: "string", valEnum: []string{"application/json", "application/xml"}},
		"@res._.content": {valType: "string"},
		"@res._.desc":    {valType: "string"},
		// security
		"@security":   {valType: "map", cutListSign: secondListCutSign, cutKeyValSign: secondKeyValueCutSign},
		"@security._": {valType: "list", cutListSign: thirdListCutSign},
		// @router
		"@router":          {valType: "array", cutListSign: secondListCutSign, cutKeyValSign: secondKeyValueCutSign},
		"@router._.method": {valType: "string", valEnum: []string{"get", "put", "post", "delete", "options", "head", "patch", "trace"}},
		"@router._.path":   {valType: "string"},
	}
)
