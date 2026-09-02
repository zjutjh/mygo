package oauth

import "errors"

// 统一认证相关错误
var (
	ErrWrongAccount  = errors.New("账号错误")
	ErrWrongPassword = errors.New("密码错误")
	ErrClosed        = errors.New("统一系统当前不可用")
	ErrEditPassword  = errors.New("统一密码需要修改, 请手动登录统一修改")
	ErrNotActivated  = errors.New("账号未激活")
	ErrOther         = errors.New("其他错误")
)
