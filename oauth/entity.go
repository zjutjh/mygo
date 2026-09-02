package oauth

// UserInfo 用户信息
type UserInfo struct {
	College      string `json:"bmmc"`         // 学院
	Name         string `json:"username"`     // 姓名
	StudentID    string `json:"yhm"`          // 学号
	UserType     string `json:"jsdm"`         // 用户类型 [id]
	UserTypeDesc string `json:"jsmc"`         // 用户类型 [description]
	Gender       string `json:"xb"`           // 性别
	Avatar       string `json:"headPortrait"` // 头像链接
}
