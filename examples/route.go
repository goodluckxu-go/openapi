package examples

type ExampleRes struct {
	ID   int    `json:"id"`   // 主键
	Name string `json:"name"` // 名称
}

type ExampleBody struct {
	UserName string `json:"user_name"` // 用户名称
	Password string `json:"password"`  // 密码
}

type Test struct {
	ExampleRes
}

// GetList openapi
// @summary: 获取用户列表
// @description: 用户列表接口需要授权
// @tags: admin
// @param: in=query; name=name; type=string; desc=用户名称
// @param: in=query; name=sex; type=string; desc=用户性别
// @res: status=200; in=application/json; content=[]examples.ExampleRes; desc=返回信息
// @res: status=404; in=application/json; content=github.com/getkin/kin-openapi/openapi3.T; desc=错误信息
// @res: status=500; in=application/json; content=examples.Test; desc=错误信息
// @security: token;projectID
// @router: method=get;path=/user/list
func GetList() {

}

// Login openapi
// @summary: 登录后台
// @description: 登录后返回token信息
// @tags: admin
// @body: in=application/json; content=examples.ExampleBody; desc=用户信息
// @res: status=200; in=application/json; content=examples.ExampleRes; desc=返回信息
// @res: status=404; in=application/json; content=examples.ExampleRes; desc=错误信息
// @security: token;projectID
// @router: method=post;path=/admin/login
func Login() {

}

// Logout openapi
// @summary: 退出登录
// @description: 清除后端的登录缓存
// @tags: admin
// @res: status=200; in=application/json; content=examples.ExampleRes; desc=返回信息
// @res: status=404; in=application/json; content=examples.ExampleRes; desc=错误信息
// @security: token;projectID
// @router: method=delete;path=/admin/logout
func Logout() {

}

// Info openapi
// @summary: 用户信息
// @description: 根据用户id查找用户信息
// @tags: admin
// @param: in=path; name=id; type=integer(int64); required; desc=主键; minimum=1;maximum=10;default=8
// @res: status=200; in=application/json; content=examples.ExampleRes; desc=返回信息
// @res: status=404; in=application/json; content=examples.ExampleRes; desc=错误信息
// @security: token;projectID
// @router: method=get;path=/admin/{id}
func Info() {

}
