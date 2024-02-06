// Package examples
// @info.title:	Swagger Petstore - OpenAPI 3.0
// @info.description: |-
//
//	This is a sample Pet Store Server based on the OpenAPI 3.0 specification.  You can find out more about
//	Swagger at [https://swagger.io](https://swagger.io). In the third iteration of the pet store, we've switched to the design first approach!
//	You can now help us improve the API whether it's by making changes to the definition itself or to the code.
//	That way, with time, we can improve the API in general, and expose some of the new features in OAS3.
//	@abc=125
//	_If you're looking for the Swagger 2.0/OAS 2.0 version of Petstore, then click [here](https://editor.swagger.io/?url=https://petstore.swagger.io/v2/swagger.yaml). Alternatively, you can load via the `Edit > Load Petstore OAS 2.0` menu option!_
//
//	Some useful links:
//	- [The Pet Store repository](https://github.com/swagger-api/swagger-petstore)
//	- [The source API definition for the Pet Store](https://github.com/swagger-api/swagger-petstore/blob/master/src/main/resources/openapi.yaml)
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
// @tags: name=admin;description=后台管理
//
// 验证设置
// @components.securitySchemes: field=token;type=apiKey;name=api_key;in=header
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
package examples
