package dingtalk

import "fmt"

type ClientConfig struct {
	ClientID     string
	ClientSecret string
	TokenURL     string
	UserInfoURL  string
}

type UserTokenResponse struct {
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
	ExpireIn     int64  `json:"expireIn"`
	CorpID       string `json:"corpId"`
}

type StaffInfo struct {
	UserID   string
	Name     string
	Nickname string
	Email    string
	DeptIDs  []int64
}

type DeptInfo struct {
	DeptID   int64
	Name     string
	ParentID int64
}

type APIError struct {
	Code       string
	Message    string
	HTTPStatus int
}

func (e *APIError) Error() string {
	return fmt.Sprintf("dingtalk: code=%s msg=%s http=%d", e.Code, e.Message, e.HTTPStatus)
}

type apiResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	ErrCode int    `json:"errcode"`
	ErrMsg  string `json:"errmsg"`
}

func parseError(body []byte, status int) *APIError {
	var v apiResponse
	_ = jsonUnmarshal(body, &v)
	code := v.Code
	if code == "" && v.ErrCode != 0 {
		code = fmt.Sprintf("%d", v.ErrCode)
	}
	msg := v.Message
	if msg == "" {
		msg = v.ErrMsg
	}
	return &APIError{Code: code, Message: msg, HTTPStatus: status}
}
