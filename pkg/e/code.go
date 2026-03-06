package e

const (
	Success       = 20000
	Created       = 20100
	
	InvalidParams = 40000
	UserExists    = 40001
	
	Unauthorized  = 40100
	LoginFailed   = 40101
	
	NotFound      = 40400
	
	Error         = 50000
)

var msgFlags = map[int]string{
	Success:       "success",
	Created:       "创建/注册成功",
	InvalidParams: "请求参数错误",
	UserExists:    "用户名已存在",
	Unauthorized:  "Token缺失或无效",
	LoginFailed:   "用户名或密码错误",
	NotFound:      "资源不存在",
	Error:         "服务器内部错误",
}

// GetMsg 获取错误码对应的提示信息
func GetMsg(code int) string {
	msg, ok := msgFlags[code]
	if ok {
		return msg
	}
	return msgFlags[Error]
}