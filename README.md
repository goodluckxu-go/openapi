# openapi3 文档生成

使用ast语法解析器解析注解，根据openapi3语法生成

## 引用(quote)
- github.com/getkin/kin-openapi


## 用法(usage)
多行注释可以使用 |- 符号，和@标题一行，不能有其他内容，知道遇到下一个可用标签或者单独一行为-|。例如：
~~~go
// @info.title: |-
//   这是标题
// @info.description: |-
//   这是描述
//  -|
// 这里的注释是无效的，上面已经终止，如果不些终止符号则为上面标签的注释
// @info.version: 这是版本号
~~~

### 安装命令
~~~shell
go install github.com/goodluckxu-go/openapi/cmd/apigen@latest
~~~

### docs.go文档注释说明
~~~go
// @info.title: 标题
// @info.description: 描述
// @info.termsOfService: 服务条款
// @info.contact.name: 联系人
// @info.contact.url: 联系地址
// @info.contact.email: 联系邮箱
// @info.license.name: 许可证名称
// @info.license.url: 许可证地址
// @info.version: 项目版本号
// @externalDocs.description: 扩展文档描述
// @externalDocs.url: 扩展文档地址
// @servers: url=服务地址; description=服务描述
// @tags: name=标签名称; description=标签描述
// @components.securitySchemes: |-
//  field=验证字段，路由注释中使用;
//  type=验证类型，值包括apiKey,http,oauth2;
//  scheme=http类型必传，例如basic,bearer;
//  bearerFormat=scheme为bearer时可传，用于提示客户端所使用的bearer token的格式，例如JWT;
//  in=apiKey时必传，值包括query,header,cookie;
//  name=apiKey时必传，用于 header、 query 或 cookie 的参数名字;
//  flows=json字符串，文档说明详见：https://openapi.apifox.cn/#oauth-flows-%E5%AF%B9%E8%B1%A1
package main
~~~

#### docs.go 文档中定义公共的 route.go 中的属性，和 route.go 的注释一致，在@后面添加global.
- @global.res  公共返回，和路由的一致
- @global.param  公共参数，和路由的一致
~~~go
// @global.res: status=500; in=application/json; content=服务器链接失败; desc=系统内部错误
package main
~~~

### route.go文档注释说明
- @开始的标题，用 ; 分割，分割成对象。每个对象用 = 分割，分割成键，值，如果不存在 = 好则表示值是字符串 true

- 上面分割的对象中的值，用 , 分割成数组
~~~go
package main

// @summary: 路由总结
// @description: 路由描述
// @tags: 标签组，用;分割，例如：user;admin
// @param: 参数，多行则多个参数，详细说明见下面@param说明
// @body: 传递内容，详细说明见下面@body说明
// @res: 输出内容，详解说明见下面@res说明
// @security: |-
//  验证值，使用 @components.securitySchemes 中定义的 field 的值
//  例如：token;projectID=write:pets,read:pets 表示 存在token验证，projectID验证数组是[write:pets,read:pets]
// @router: |-
//  method=get,put ,post,delete,options,head,patch中的值;
//  path=路由地址，例如：/user/{id}。其中{id}表示@param中的in为path时的关联
func Login() {
}
~~~
#### @param说明
实例：@param: in=path; name=id; type=integer(int64); required; desc=主键
- in 表示参数类型，值有query,header,path,cookie
- name 表示参数名称
- type 表示参数类型，值有integer,number,string,boolean
- required 是否必传参数，实例：required 或者 required=true
- desc 参数描述
- minimum type类型是integer时的最小值
- maximum type类型是integer时的最大值
- minLength type类型是string时的最小长度
- maxLength type类型是string时的最大长度
- example 实例值
- default 默认值
- enum 参数枚举，数组，用,分割，例如：enum=user,name
#### @body说明
实例：@body: in=application/json; content=test/project/app/reqs.LoginAdminReq; desc=用户信息
- in 传入类型，值有 application/json, application/xml, application/x-www-form-urlencoded
- content 传入内容，以.分割，前缀为go.mod查找的命名空间名称(支持github等，必须引入)，后缀为结构体名称。前缀可以是结构体package的名称，这种情况必须不能重复
- desc 传入内容描述
#### @res说明
实例：@res: status=200; in=application/json; content=test/project/app/resps.AdminLoginResp; desc=返回信息
- status integer类型，服务器的状态码
- in 返回类型，值有 application/json, application/xml
- content 返回内容，以.分割，前缀为go.mod查找的命名空间名称(支持github等，必须引入)，后缀为结构体名称。前缀可以是结构体package的名称，这种情况必须不能重复
- desc 返回内容描述

### 结构体注释说明
~~~go
package main

// LoginRequest 结构体注释
type LoginRequest struct {
    Account string `json:"account" yaml:"account" required:"true" minimum:"10"` // 账号注释
	Password string `json:"password" yaml:"password" openapi:"required;minimum=10"` // 密码注释
}
~~~

#### 结构体的标签可以使用 minimum:"10" 和 openapi:"minimum=10" 这两种方式，字段有
- minimum type类型是integer时的最小值
- maximum type类型是integer时的最大值
- minLength type类型是string时的最小长度
- maxLength type类型是string时的最大长度
- minItems type类型是slice时的最小长度
- maxItems type类型是slice时的最大长度
- example 实例值
- default 默认值
- enum 限定值
- required 是否必传参数
- type 类型重定义

## 文件上传
只需要将in设置为 multipart/form-data， 类型设置为 base64 或者 binary 即可

## 关于(about)
灵感为 github.com/swaggo/swag 的项目，因为这个项目无法解析 openapi3 的文档，因此自己实现了一套 openapi3 的文档生成