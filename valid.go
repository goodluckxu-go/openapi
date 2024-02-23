package openapi

type validStruct struct {
	valType       int      // 类型
	cutListSign   string   // 列表截取标志
	cutKeyValSign string   // 键值截取标志
	valEnum       []string // 枚举验证
	isUnique      bool     // 是否唯一
	isSort        bool     // 是否map排序
}

var (
	validDocMap = map[string]*validStruct{
		// info
		"@info.title":          {valType: validTypeString},
		"@info.description":    {valType: validTypeString},
		"@info.termsOfService": {valType: validTypeString},
		"@info.contact.name":   {valType: validTypeString},
		"@info.contact.url":    {valType: validTypeString},
		"@info.contact.email":  {valType: validTypeString},
		"@info.license.name":   {valType: validTypeString},
		"@info.license.url":    {valType: validTypeString},
		"@info.version":        {valType: validTypeString},
		// externalDocs
		"@externalDocs.description": {valType: validTypeString},
		"@externalDocs.url":         {valType: validTypeString},
		// servers
		"@servers":               {valType: validTypeMapArray, cutListSign: secondListCutSign, cutKeyValSign: secondKeyValueCutSign},
		"@servers._.url":         {valType: validTypeString},
		"@servers._.description": {valType: validTypeString},
		// tags
		"@tags":               {valType: validTypeMapArray, cutListSign: secondListCutSign, cutKeyValSign: secondKeyValueCutSign},
		"@tags._.name":        {valType: validTypeString},
		"@tags._.description": {valType: validTypeString},
		// components
		"@components.securitySchemes":                {valType: validTypeMapArray, cutListSign: secondListCutSign, cutKeyValSign: secondKeyValueCutSign},
		"@components.securitySchemes._.field":        {valType: validTypeString, isUnique: true},
		"@components.securitySchemes._.type":         {valType: validTypeString, valEnum: []string{"apiKey", "http", "oauth2"}},
		"@components.securitySchemes._.scheme":       {valType: validTypeString},
		"@components.securitySchemes._.bearerFormat": {valType: validTypeString},
		"@components.securitySchemes._.name":         {valType: validTypeString},
		"@components.securitySchemes._.in":           {valType: validTypeString, valEnum: []string{"query", "header", "cookie"}},
		"@components.securitySchemes._.flows":        {valType: validTypeJson},
		// global
		"@global.res":           validRoutesMap["@res"],
		"@global.res._.status":  validRoutesMap["@res._.status"],
		"@global.res._.in":      validRoutesMap["@res._.in"],
		"@global.res._.content": validRoutesMap["@res._.content"],
		"@global.res._.desc":    validRoutesMap["@res._.desc"],
	}

	validRoutesMap = map[string]*validStruct{
		"@summary":     {valType: validTypeString},
		"@description": {valType: validTypeString},
		"@tags":        {valType: validTypeMap, cutListSign: secondListCutSign},
		// param
		"@param":             {valType: validTypeMapArray, cutListSign: secondListCutSign, cutKeyValSign: secondKeyValueCutSign, valEnum: []string{"required"}},
		"@param._.in":        {valType: validTypeString, valEnum: []string{"query", "header", "path", "cookie"}},
		"@param._.name":      {valType: validTypeString},
		"@param._.type":      {valType: validTypeString},
		"@param._.required":  {valType: validTypeBool},
		"@param._.desc":      {valType: validTypeString},
		"@param._.minimum":   {valType: validTypeInteger},
		"@param._.maximum":   {valType: validTypeInteger},
		"@param._.minLength": {valType: validTypeInteger},
		"@param._.maxLength": {valType: validTypeInteger},
		"@param._.example":   {valType: validTypeString},
		"@param._.default":   {valType: validTypeString},
		"@param._.enum":      {valType: validTypeArray, cutListSign: thirdListCutSign},
		// body
		"@body":           {valType: validTypeMap, cutListSign: secondListCutSign, cutKeyValSign: secondKeyValueCutSign},
		"@body._.in":      {valType: validTypeArray, cutListSign: thirdListCutSign, valEnum: []string{"application/json", "application/xml", "application/x-www-form-urlencoded", "multipart/form-data"}},
		"@body._.content": {valType: validTypeString},
		"@body._.desc":    {valType: validTypeString},
		// res
		"@res":           {valType: validTypeMapArray, cutListSign: secondListCutSign, cutKeyValSign: secondKeyValueCutSign},
		"@res._.status":  {valType: validTypeInteger},
		"@res._.in":      {valType: validTypeArray, cutListSign: thirdListCutSign, valEnum: []string{"application/json", "application/xml"}},
		"@res._.content": {valType: validTypeString},
		"@res._.desc":    {valType: validTypeString},
		// security
		"@security":   {valType: validTypeMap, cutListSign: secondListCutSign, cutKeyValSign: secondKeyValueCutSign, isSort: true},
		"@security._": {valType: validTypeArray, cutListSign: thirdListCutSign},
		// @router
		"@router":          {valType: validTypeMapArray, cutListSign: secondListCutSign, cutKeyValSign: secondKeyValueCutSign},
		"@router._.method": {valType: validTypeString, valEnum: []string{"get", "put", "post", "delete", "options", "head", "patch", "trace"}},
		"@router._.path":   {valType: validTypeString},
	}
)
