package model

type RegisterRequest struct {
	Username         string `json:"username"`
	Password         string `json:"password"`
	UserType         int    `json:"user_type"`
	VerificationType int    `json:"verification_type"`
}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type APIResponse struct {
	Code  int         `json:"code"`
	Msg   *string     `json:"msg"`
	Ok    bool        `json:"ok"`
	Error *string     `json:"error"`
	Data  interface{} `json:"data"`
}
