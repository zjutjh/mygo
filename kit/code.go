package kit

var CodeOK = NewCode(200, "成功")

// 系统错误码
var (
	CodeUnknownError           = NewCode(200101, "未知错误")
	CodeThirdServiceError      = NewCode(200102, "三方服务错误")
	CodeDatabaseError          = NewCode(200103, "数据库错误")
	CodeRedisError             = NewCode(200104, "Redis错误")
	CodeMiddlewareServiceError = NewCode(200105, "中间件服务错误")
)

// 业务通用错误码
var (
	CodeNotLoggedIn        = NewCode(200201, "用户未登录")
	CodeLoginExpired       = NewCode(200202, "登录过期，请重新登录")
	CodePermissionDenied   = NewCode(200203, "用户无权限")
	CodeParameterInvalid   = NewCode(200204, "参数非法")
	CodeDataParseError     = NewCode(200205, "数据解析异常")
	CodeDataNotFound       = NewCode(200206, "数据不存在")
	CodeDataConflict       = NewCode(200207, "数据冲突")
	CodeServiceMaintenance = NewCode(200208, "系统维护中")
	CodeTooFrequently      = NewCode(200209, "操作过于频繁/未获得锁")
)

type Code struct {
	Code    int64
	Message string
}

func NewCode(code int64, message string) Code {
	return Code{
		Code:    code,
		Message: message,
	}
}
