package examples

import (
	"time"
)

type UserListResponseSuccess struct {
	ID         int       `json:"id" default:"1"`                            // 主键
	Name       string    `json:"name" default:"张三"`                         // 名称
	Desc       string    `json:"desc" default:"张三非常棒"`                      // 简介
	CreateTime time.Time `json:"create_time" default:"2024-02-20 14:21:13"` // 创建时间
}

type ResponseError struct {
	Code int    `json:"code" default:"404"` // 错误码
	Msg  string `json:"msg" default:"返回失败"` // 错误信息
}

type ResponseSuccess struct {
	Code int `json:"code" default:"0"` // 操作成功
}

// LoginRequest 是登录参数
type LoginRequest struct {
	Account  string `json:"account" openapi:"required"`  // 账号
	Password string `json:"password" openapi:"required"` // 密码
	Code     string `json:"code" required:"true"`        // 验证码
}

type User struct {
}

// GetList openapi
// @summary: 获取用户列表
// @description: 用户列表接口需要授权
// @tags: user
// @param: in=query; name=name; type=string; desc=用户名称
// @param: in=query; name=sex; type=string; desc=用户性别
// @res: status=200; in=application/json; content=[]examples.UserListResponseSuccess; desc=获取成功
// @res: status=404; in=application/json; content=examples.ResponseError; desc=获取失败
// @security: token;projectID=write:pets,read:pets
// @router: method=get;path=/user/list
func (u *User) GetList() {

}

// Info openapi
// @summary: 用户信息
// @description: 根据用户id查找用户信息
// @tags: user
// @param: in=path; name=id; type=int64; required; desc=主键; minimum=1;maximum=10;default=1
// @res: status=200; in=application/json; content=examples.UserListResponseSuccess; desc=返回信息
// @res: status=404; in=application/json; content=examples.ResponseError; desc=获取失败
// @security: token;projectID=write:pets,read:pets
// @router: method=get;path=/user/{id}
func (u *User) Info() {

}

type Admin struct {
}

// Login openapi
// @summary: 登录后台
// @description: 登录后返回token信息
// @tags: admin
// @body: in=application/json; content=examples.LoginRequest; desc=登录参数
// @res: status=200; in=application/json; content=examples.ResponseSuccess; desc=登录成功
// @res: status=404; in=application/json; content=examples.ResponseError; desc=退出成功
// @security: token;projectID=write:pets,read:pets
// @router: method=post;path=/admin/login
func (a Admin) Login() {

}

// Logout openapi
// @summary: 退出登录
// @description: 清除后端的登录缓存
// @tags: admin
// @res: status=200; in=application/json; content=examples.ResponseSuccess; desc=退出成功
// @res: status=404; in=application/json; content=examples.ResponseError; desc=退出失败
// @security: token;projectID=write:pets,read:pets
// @router: method=delete;path=/admin/logout
func (a Admin) Logout() {

}

// Index openapi
// @summary: 测试结构体递归注释
// @res: status=200; in=application/json;content=github.com/getkin/kin-openapi/openapi3.T; desc=递归注释
// @router: method=get;path=/index
func Index() {

}

type UploadRequest struct {
	FileBinary string `json:"file_binary" type:"binary"` // 上传文件
	FileBase64 string `json:"file_base64" type:"binary"` // 上传文件
}

// Upload 上传文件
// @summary: 上传文件
// @body: in=multipart/form-data; content=examples.UploadRequest; desc=上传文件
// @router: method=put;path=/upload
func Upload() {

}
