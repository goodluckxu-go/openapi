// Package examples
// @info.title:	openapi3文档测试接口
// @info.description: |-
//
//	该接口文档用于测试本程序生成的文档
//
// @info.termsOfService: http://swagger.io/terms/
// @info.contact.email:	807495056@qq.com
// @info.license.name: Apache 2.0
// @info.license.url: http://www.apache.org/licenses/LICENSE-2.0.html
// @info.version: 1.0.0
//
// @externalDocs.description:Find out more about Swagger
// @externalDocs.url: http://swagger.io
//
// @servers: url=/v1;description=版本v1
// @servers: url=/v2;description=版本v2
//
// @tags: name=default
// @tags: name=user;description=用户管理
// @tags: name=admin;description=后台管理
//
// 验证设置
// @components.securitySchemes: field=token;type=apiKey;name=token;in=header
//
// @components.securitySchemes: |-
//
//	field=projectID;type=oauth2;
//	flows={
//		   "implicit": {
//		     "authorizationUrl": "https://example.com/api/oauth/dialog",
//		     "scopes": {
//		       "write:pets": "modify pets in your account",
//		       "read:pets": "read your pets"
//		     }
//		   }
//		 }
//
// @global.res: status=500; in=application/json; content=服务器链接失败; desc=系统内部错误
package examples
