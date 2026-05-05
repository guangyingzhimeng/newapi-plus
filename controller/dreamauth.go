package controller

import (
	"bytes"
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/i18n"
	"github.com/QuantumNous/new-api/model"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

const (
	defaultDreamAuthBaseURL = "https://guangyingzhimeng.dpdns.org/kite-hub"
	dreamAuthCreatePath     = "/api/open/scan-login/session/create"
	dreamAuthStatusPath     = "/api/open/scan-login/session/status"
	dreamAuthResultPath     = "/api/open/scan-login/session/result"
)

type dreamAuthCreateSessionRequest struct {
	TargetType string `json:"targetType"`
	BizState   string `json:"bizState"`
}

type dreamAuthSessionRequest struct {
	SessionNo string `json:"sessionNo"`
}

type dreamAuthCreatePayload struct {
	BizCode       string `json:"bizCode"`
	BizState      string `json:"bizState"`
	TargetType    string `json:"targetType"`
	ExpireSeconds int    `json:"expireSeconds"`
}

type dreamAuthStatusPayload struct {
	SessionNo string `json:"sessionNo"`
}

type dreamAuthResultPayload struct {
	SessionNo string `json:"sessionNo"`
	Consume   bool   `json:"consume"`
}

type dreamAuthEnvelope struct {
	Code    int             `json:"code"`
	Msg     string          `json:"msg"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data"`
}

type dreamAuthSessionData struct {
	SessionNo string `json:"sessionNo"`
	Scene     string `json:"scene"`
	QRCode    string `json:"qrcode"`
	ExpireAt  string `json:"expireAt"`
	AppCode   string `json:"appCode"`
}

type dreamAuthStatusData struct {
	SessionNo  string `json:"sessionNo"`
	Scene      string `json:"scene"`
	Status     int    `json:"status"`
	MemberRole *int   `json:"memberRole"`
	AuthTime   string `json:"authTime"`
	ExpireAt   string `json:"expireAt"`
}

type dreamAuthResultData struct {
	SessionNo  string `json:"sessionNo"`
	OpenID     string `json:"openid"`
	UnionID    string `json:"unionid"`
	MemberRole *int   `json:"memberRole"`
	AuthTime   string `json:"authTime"`
}

type dreamAuthConfig struct {
	BaseURL   string
	AppCode   string
	AccessKey string
	SecretKey string
}

func getDreamAuthConfig() (dreamAuthConfig, error) {
	cfg := dreamAuthConfig{
		BaseURL:   strings.TrimRight(strings.TrimSpace(os.Getenv("DREAMAUTH_BASE_URL")), "/"),
		AppCode:   strings.TrimSpace(os.Getenv("DREAMAUTH_APP_CODE")),
		AccessKey: strings.TrimSpace(os.Getenv("DREAMAUTH_ACCESS_KEY")),
		SecretKey: strings.TrimSpace(os.Getenv("DREAMAUTH_SECRET_KEY")),
	}
	if cfg.BaseURL == "" {
		cfg.BaseURL = defaultDreamAuthBaseURL
	}
	if cfg.AppCode == "" || cfg.AccessKey == "" || cfg.SecretKey == "" {
		return cfg, errors.New("DreamAuth credentials are not configured")
	}
	return cfg, nil
}

func dreamAuthPayloadJSON(payload any) (string, error) {
	if payload == nil {
		return "", nil
	}
	data, err := common.Marshal(payload)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func dreamAuthHeaders(cfg dreamAuthConfig, method string, path string, payload any) (http.Header, error) {
	payloadJSON, err := dreamAuthPayloadJSON(payload)
	if err != nil {
		return nil, err
	}
	timestamp := strconv.FormatInt(time.Now().Unix(), 10)
	nonce := common.GetRandomString(16)
	source := fmt.Sprintf("%s|%s|%s|%s|%s|%s", strings.ToUpper(method), path, timestamp, nonce, payloadJSON, cfg.SecretKey)
	sum := md5.Sum([]byte(source))

	headers := http.Header{}
	headers.Set("X-Kite-AK", cfg.AccessKey)
	headers.Set("X-Kite-Timestamp", timestamp)
	headers.Set("X-Kite-Nonce", nonce)
	headers.Set("X-Kite-Sign", hex.EncodeToString(sum[:]))
	headers.Set("Content-Type", "application/json")
	return headers, nil
}

func dreamAuthRequest[T any](ctx context.Context, method string, path string, payload any, query url.Values) (T, error) {
	var result T
	cfg, err := getDreamAuthConfig()
	if err != nil {
		return result, err
	}

	var body io.Reader
	if method != http.MethodGet && payload != nil {
		payloadBytes, err := common.Marshal(payload)
		if err != nil {
			return result, err
		}
		body = bytes.NewReader(payloadBytes)
	}

	reqURL := cfg.BaseURL + path
	if len(query) > 0 {
		reqURL += "?" + query.Encode()
	}
	req, err := http.NewRequestWithContext(ctx, method, reqURL, body)
	if err != nil {
		return result, err
	}
	headers, err := dreamAuthHeaders(cfg, method, path, payload)
	if err != nil {
		return result, err
	}
	req.Header = headers

	client := http.Client{Timeout: 20 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return result, err
	}
	defer resp.Body.Close()

	var envelope dreamAuthEnvelope
	if err := common.DecodeJson(resp.Body, &envelope); err != nil {
		return result, errors.New("DreamAuth returned a non-JSON response")
	}
	if resp.StatusCode >= http.StatusBadRequest || envelope.Code != http.StatusOK {
		message := envelope.Msg
		if message == "" {
			message = envelope.Message
		}
		if message == "" {
			message = fmt.Sprintf("DreamAuth HTTP %d", resp.StatusCode)
		}
		return result, errors.New(message)
	}
	if len(envelope.Data) == 0 {
		return result, nil
	}
	if err := common.Unmarshal(envelope.Data, &result); err != nil {
		return result, err
	}
	return result, nil
}

func dreamAuthStatusText(status int, expired bool) string {
	if expired {
		return "二维码已过期"
	}
	switch status {
	case 1:
		return "请使用微信扫描二维码"
	case 2:
		return "已扫码，请在 DreamAuth 中确认授权"
	case 3, 5:
		return "授权成功"
	case 4:
		return "授权已取消"
	case 6:
		return "登录结果已消费"
	case 7:
		return "二维码已过期"
	default:
		return "等待扫码"
	}
}

func isDreamAuthExpired(status dreamAuthStatusData) bool {
	if status.Status == 7 {
		return true
	}
	if status.ExpireAt == "" {
		return false
	}
	expireAt, err := time.Parse("2006-01-02T15:04:05", status.ExpireAt)
	if err != nil {
		expireAt, err = time.Parse(time.RFC3339, status.ExpireAt)
	}
	return err == nil && time.Now().After(expireAt)
}

func CreateDreamAuthSession(c *gin.Context) {
	var req dreamAuthCreateSessionRequest
	if err := common.DecodeJson(c.Request.Body, &req); err != nil && !errors.Is(err, io.EOF) {
		common.ApiErrorI18n(c, i18n.MsgInvalidParams)
		return
	}

	cfg, err := getDreamAuthConfig()
	if err != nil {
		common.ApiErrorMsg(c, err.Error())
		return
	}

	targetType := strings.TrimSpace(req.TargetType)
	if targetType == "" {
		targetType = "user"
	}
	bizState := strings.TrimSpace(req.BizState)
	if bizState == "" {
		bizState = "new-api-" + common.GetRandomString(12)
	}
	payload := dreamAuthCreatePayload{
		BizCode:       "LOGIN",
		BizState:      bizState,
		TargetType:    targetType,
		ExpireSeconds: 300,
	}
	data, err := dreamAuthRequest[dreamAuthSessionData](c.Request.Context(), http.MethodPost, dreamAuthCreatePath, payload, nil)
	if err != nil {
		common.ApiErrorMsg(c, err.Error())
		return
	}
	if data.AppCode == "" {
		data.AppCode = cfg.AppCode
	}
	common.ApiSuccess(c, data)
}

func GetDreamAuthSessionStatus(c *gin.Context) {
	sessionNo := strings.TrimSpace(c.Param("sessionNo"))
	if sessionNo == "" {
		common.ApiErrorI18n(c, i18n.MsgInvalidParams)
		return
	}
	payload := dreamAuthStatusPayload{SessionNo: sessionNo}
	query := url.Values{"sessionNo": []string{sessionNo}}
	data, err := dreamAuthRequest[dreamAuthStatusData](c.Request.Context(), http.MethodGet, dreamAuthStatusPath, payload, query)
	if err != nil {
		common.ApiErrorMsg(c, err.Error())
		return
	}
	expired := isDreamAuthExpired(data)
	common.ApiSuccess(c, gin.H{
		"sessionNo":  data.SessionNo,
		"scene":      data.Scene,
		"status":     data.Status,
		"statusText": dreamAuthStatusText(data.Status, expired),
		"memberRole": data.MemberRole,
		"authTime":   data.AuthTime,
		"expireAt":   data.ExpireAt,
		"loginReady": data.Status == 3 || data.Status == 5,
		"expired":    expired,
	})
}

func CompleteDreamAuthLogin(c *gin.Context) {
	var req dreamAuthSessionRequest
	if err := common.DecodeJson(c.Request.Body, &req); err != nil {
		common.ApiErrorI18n(c, i18n.MsgInvalidParams)
		return
	}
	sessionNo := strings.TrimSpace(req.SessionNo)
	if sessionNo == "" {
		common.ApiErrorI18n(c, i18n.MsgInvalidParams)
		return
	}
	payload := dreamAuthResultPayload{SessionNo: sessionNo, Consume: true}
	data, err := dreamAuthRequest[dreamAuthResultData](c.Request.Context(), http.MethodPost, dreamAuthResultPath, payload, nil)
	if err != nil {
		common.ApiErrorMsg(c, err.Error())
		return
	}
	openID := strings.TrimSpace(data.OpenID)
	if openID == "" {
		common.ApiErrorMsg(c, "DreamAuth result did not include openid")
		return
	}

	user, err := findOrCreateDreamAuthUser(c, openID)
	if err != nil {
		common.ApiErrorMsg(c, err.Error())
		return
	}
	if user.Status != common.UserStatusEnabled {
		common.ApiErrorI18n(c, i18n.MsgOAuthUserBanned)
		return
	}
	setupLogin(user, c)
}

func findOrCreateDreamAuthUser(c *gin.Context, openID string) (*model.User, error) {
	openID = strings.TrimSpace(openID)
	adminOpenID := strings.TrimSpace(os.Getenv("DREAMAUTH_ADMIN_OPENID"))
	isAdmin := adminOpenID != "" && openID == adminOpenID

	common.SysLog(fmt.Sprintf("DreamAuth login: openID=[%s], isAdmin=%v (target admin=[%s])", openID, isAdmin, adminOpenID))

	user := &model.User{WeChatId: openID}
	if model.IsWeChatIdAlreadyTaken(openID) {
		if err := user.FillUserByWeChatId(); err != nil {
			return nil, err
		}
		if user.Id == 0 {
			return nil, errors.New("用户已注销")
		}
		if isAdmin && user.Role != common.RoleRootUser {
			common.SysLog(fmt.Sprintf("Upgrading existing user %d to admin based on openID match", user.Id))
			user.Role = common.RoleRootUser
			model.DB.Model(user).Update("role", common.RoleRootUser)
		}
		return user, nil
	}
	if !common.RegisterEnabled {
		return nil, errors.New("管理员关闭了新用户注册")
	}

	session := sessions.Default(c)
	inviterId := 0
	if affCode, ok := session.Get("aff").(string); ok && affCode != "" {
		inviterId, _ = model.GetUserIdByAffCode(affCode)
	}

	user.Username = "dreamauth_" + strconv.Itoa(model.GetMaxUserId()+1)
	if len(user.Username) > model.UserNameMaxLength {
		user.Username = "da_" + strconv.Itoa(model.GetMaxUserId()+1)
	}
	user.DisplayName = "DreamAuth User"
	if isAdmin {
		user.Role = common.RoleRootUser
		common.SysLog(fmt.Sprintf("Creating new admin user: %s", user.Username))
	} else {
		user.Role = common.RoleCommonUser
	}
	user.Status = common.UserStatusEnabled
	user.WeChatId = openID

	if err := user.Insert(inviterId); err != nil {
		return nil, err
	}
	return user, nil
}
