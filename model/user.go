package model

// User 用户登录类
type User struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Cpu      int    `json:"cpu"`
	Ram      int    `json:"ram"`
	Ip       string `json:"ip"`
}
