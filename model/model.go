package model

type RegisterRequest struct {
	Email            string `json:"email"`
	Password         string `json:"password"`
	UserType         int    `json:"user_type"`
	VerificationType int    `json:"verification_type"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type APIResponse struct {
	Code  int         `json:"code"`
	Msg   *string     `json:"msg"`
	Ok    bool        `json:"ok"`
	Error *string     `json:"error"`
	Data  interface{} `json:"data"`
}
