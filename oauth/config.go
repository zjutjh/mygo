package oauth

const (
	defaultMeZjutURL  = "http://www.me.zjut.edu.cn"
	defaultOAuthURL   = "https://oauth.zjut.edu.cn/cas"
)

var DefaultConfig = Config{
	PersonalCenterURL: defaultMeZjutURL + "/personal-center",
	UserInfoURL:       defaultMeZjutURL + "/api/basic/info",

	LoginURL:          defaultOAuthURL + "/login",
	PublicKeyURL:      defaultOAuthURL + "/v2/getPubKey",

	WrongPasswordMsg:  "用户名或密码错误",
	WrongAccountMsg:   "当前账号无权登录",
	NotActivatedMsg:   "账号未激活，请激活后再登录",

	Resty:             "",
}

type Config struct {
	PersonalCenterURL string `mapstructure:"personal_center_url"`
	UserInfoURL       string `mapstructure:"user_info_url"`

	LoginURL          string `mapstructure:"login_url"`
	PublicKeyURL      string `mapstructure:"public_key_url"`

	WrongPasswordMsg  string `mapstructure:"wrong_password_msg"`
	WrongAccountMsg   string `mapstructure:"wrong_account_msg"`
	NotActivatedMsg   string `mapstructure:"not_activated_msg"`

	// 基础依赖组件实例配置
	Resty string `mapstructure:"resty"`
}
