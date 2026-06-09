package dingtalk

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

var jsonUnmarshal = json.Unmarshal

type Client struct {
	cfg        ClientConfig
	httpClient *http.Client

	mu          sync.Mutex
	appToken    string
	appTokenExp time.Time
}

func NewClient(cfg ClientConfig, httpClient *http.Client) *Client {
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 10 * time.Second}
	}
	return &Client{cfg: cfg, httpClient: httpClient}
}

func (c *Client) ExchangeCode(ctx context.Context, code string) (*UserTokenResponse, error) {
	payload, _ := json.Marshal(map[string]string{
		"clientId":     c.cfg.ClientID,
		"clientSecret": c.cfg.ClientSecret,
		"code":         code,
		"grantType":    "authorization_code",
	})
	body, status, err := c.doJSON(ctx, http.MethodPost, c.cfg.TokenURL, payload, "")
	if err != nil {
		return nil, err
	}
	if status != http.StatusOK {
		return nil, parseError(body, status)
	}
	var out UserTokenResponse
	if err := jsonUnmarshal(body, &out); err != nil {
		return nil, err
	}
	if strings.TrimSpace(out.AccessToken) == "" {
		return nil, parseError(body, status)
	}
	return &out, nil
}

func (c *Client) GetUserByToken(ctx context.Context, userToken string) (unionID, nick string, err error) {
	body, status, err := c.doJSON(ctx, http.MethodGet, c.cfg.UserInfoURL, nil, userToken)
	if err != nil {
		return "", "", err
	}
	if status != http.StatusOK {
		return "", "", parseError(body, status)
	}
	var v struct {
		UnionID string `json:"unionId"`
		Nick    string `json:"nick"`
	}
	if err := jsonUnmarshal(body, &v); err != nil {
		return "", "", err
	}
	if strings.TrimSpace(v.UnionID) == "" {
		return "", "", parseError(body, status)
	}
	return v.UnionID, v.Nick, nil
}

func (c *Client) GetAppToken(ctx context.Context) (string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.appToken != "" && time.Now().Before(c.appTokenExp) {
		return c.appToken, nil
	}

	payload, _ := json.Marshal(map[string]string{
		"appKey":    c.cfg.ClientID,
		"appSecret": c.cfg.ClientSecret,
	})
	tokenURL := c.deriveAppTokenURL()

	body, status, err := c.doJSONUnlocked(ctx, http.MethodPost, tokenURL, payload, "")
	if err != nil {
		return "", err
	}
	if status != http.StatusOK {
		return "", parseError(body, status)
	}
	var v struct {
		AccessToken string `json:"accessToken"`
		ExpireIn    int64  `json:"expireIn"`
	}
	if err := jsonUnmarshal(body, &v); err != nil {
		return "", err
	}
	if v.AccessToken == "" {
		return "", parseError(body, status)
	}

	c.appToken = v.AccessToken
	ttl := v.ExpireIn
	if ttl > 200 {
		ttl -= 200
	}
	c.appTokenExp = time.Now().Add(time.Duration(ttl) * time.Second)
	return c.appToken, nil
}

func (c *Client) GetUserIDByUnionID(ctx context.Context, unionID string) (string, error) {
	appToken, err := c.GetAppToken(ctx)
	if err != nil {
		return "", err
	}

	payload, _ := json.Marshal(map[string]string{"unionid": unionID})
	targetURL := c.oapiURL("/topapi/user/getbyunionid") + "?access_token=" + url.QueryEscape(appToken)

	body, status, err := c.doJSON(ctx, http.MethodPost, targetURL, payload, "")
	if err != nil {
		return "", err
	}
	if status != http.StatusOK {
		return "", parseError(body, status)
	}
	var v struct {
		Result struct {
			UserID string `json:"userid"`
		} `json:"result"`
		ErrCode int    `json:"errcode"`
		ErrMsg  string `json:"errmsg"`
	}
	if err := jsonUnmarshal(body, &v); err != nil {
		return "", err
	}
	if v.ErrCode != 0 {
		return "", parseError(body, status)
	}
	if strings.TrimSpace(v.Result.UserID) == "" {
		return "", parseError(body, status)
	}
	return v.Result.UserID, nil
}

func (c *Client) GetStaffByUserID(ctx context.Context, userID string) (*StaffInfo, error) {
	appToken, err := c.GetAppToken(ctx)
	if err != nil {
		return nil, err
	}

	payload, _ := json.Marshal(map[string]string{"userid": userID})
	targetURL := c.oapiURL("/topapi/v2/user/get") + "?access_token=" + url.QueryEscape(appToken)

	body, status, err := c.doJSON(ctx, http.MethodPost, targetURL, payload, "")
	if err != nil {
		return nil, err
	}
	if status != http.StatusOK {
		return nil, parseError(body, status)
	}
	var v struct {
		Result struct {
			UserID    string  `json:"userid"`
			Name      string  `json:"name"`
			Nickname  string  `json:"nickname"`
			Email     string  `json:"email"`
			OrgEmail  string  `json:"org_email"`
			Extension string  `json:"extension"`
			DeptIDs   []int64 `json:"dept_id_list"`
		} `json:"result"`
		ErrCode int    `json:"errcode"`
		ErrMsg  string `json:"errmsg"`
	}
	if err := jsonUnmarshal(body, &v); err != nil {
		return nil, err
	}
	if v.ErrCode != 0 {
		return nil, parseError(body, status)
	}
	if strings.TrimSpace(v.Result.UserID) == "" {
		return nil, parseError(body, status)
	}

	email := resolveStaffEmail(v.Result.OrgEmail, v.Result.Email, v.Result.Extension)

	return &StaffInfo{
		UserID:   v.Result.UserID,
		Name:     v.Result.Name,
		Nickname: v.Result.Nickname,
		Email:    email,
		DeptIDs:  v.Result.DeptIDs,
	}, nil
}

func (c *Client) GetDeptInfo(ctx context.Context, deptID int64) (*DeptInfo, error) {
	appToken, err := c.GetAppToken(ctx)
	if err != nil {
		return nil, err
	}

	payload, _ := json.Marshal(map[string]any{"dept_id": deptID, "language": "zh_CN"})
	targetURL := c.oapiURL("/topapi/v2/department/get") + "?access_token=" + url.QueryEscape(appToken)

	body, status, err := c.doJSON(ctx, http.MethodPost, targetURL, payload, "")
	if err != nil {
		return nil, err
	}
	if status != http.StatusOK {
		return nil, parseError(body, status)
	}
	var v struct {
		Result struct {
			DeptID   int64  `json:"dept_id"`
			Name     string `json:"name"`
			ParentID int64  `json:"parent_id"`
		} `json:"result"`
		ErrCode int    `json:"errcode"`
		ErrMsg  string `json:"errmsg"`
	}
	if err := jsonUnmarshal(body, &v); err != nil {
		return nil, err
	}
	if v.ErrCode != 0 {
		return nil, parseError(body, status)
	}
	return &DeptInfo{
		DeptID:   v.Result.DeptID,
		Name:     v.Result.Name,
		ParentID: v.Result.ParentID,
	}, nil
}

func resolveStaffEmail(orgEmail, email, extension string) string {
	if e := strings.TrimSpace(orgEmail); e != "" {
		return e
	}
	if e := strings.TrimSpace(email); e != "" {
		return e
	}
	if strings.TrimSpace(extension) != "" {
		var ext map[string]string
		if err := json.Unmarshal([]byte(extension), &ext); err == nil {
			if v, ok := ext["企业邮箱"]; ok {
				if e := strings.TrimSpace(v); e != "" {
					return e
				}
			}
		}
	}
	return ""
}

func (c *Client) deriveAppTokenURL() string {
	return strings.Replace(c.cfg.TokenURL, "/oauth2/userAccessToken", "/oauth2/accessToken", 1)
}

func (c *Client) oapiURL(path string) string {
	u, err := url.Parse(c.cfg.UserInfoURL)
	if err != nil || u.Scheme == "" || u.Host == "" {
		return "https://oapi.dingtalk.com" + path
	}
	host := u.Host
	if strings.HasPrefix(host, "api.") {
		host = "oapi." + strings.TrimPrefix(host, "api.")
	}
	return u.Scheme + "://" + host + path
}

func (c *Client) doJSON(ctx context.Context, method, url string, payload []byte, bearerToken string) ([]byte, int, error) {
	var bodyReader io.Reader
	if payload != nil {
		bodyReader = bytes.NewReader(payload)
	}
	req, err := http.NewRequestWithContext(ctx, method, url, bodyReader)
	if err != nil {
		return nil, 0, err
	}
	if payload != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if bearerToken != "" {
		req.Header.Set("x-acs-dingtalk-access-token", bearerToken)
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer func() { _ = resp.Body.Close() }()
	body, _ := io.ReadAll(resp.Body)
	return body, resp.StatusCode, nil
}

func (c *Client) doJSONUnlocked(ctx context.Context, method, url string, payload []byte, bearerToken string) ([]byte, int, error) {
	return c.doJSON(ctx, method, url, payload, bearerToken)
}

func (c *Client) SetAppTokenForTest(token string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.appToken = token
	c.appTokenExp = time.Now().Add(time.Hour)
}
