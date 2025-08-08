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

type FileMeta struct {
	Path    string `json:"path"`
	Hash    string `json:"hash"`
	ModTime int64  `json:"mod_time"`
}

// 在 model 包中添加以下结构体

type ChunkUploadResponse struct {
	// 根据你的实际API响应结构调整
	PartIndex  int    `json:"part_index"`
	UploadID   string `json:"upload_id,omitempty"`
	ETag       string `json:"etag,omitempty"`
	IsComplete bool   `json:"is_complete,omitempty"`
}
