package hydrautil

import (
	"context"
	json "encoding/json"
	"fmt"
	"io"
	ioutil "io/ioutil"
	http "net/http"
)

type userInfoError struct {
	code int
	body []byte
}

func (err userInfoError) Error() string {
	return fmt.Sprintf("[%d] %s", err.code, err.body)
}

type contextKey string

// ContextKeyUserInfo is the context key for the user info
var ContextKeyUserInfo contextKey = "userinfo"

// UserInfo is the user info
type UserInfo interface {
	Subject() string
	GetString(key string) string
	GetInt(key string) int
	GetInt64(key string) int64
}

type userInfo map[string]interface{}

// Subject returns the subject
func (ui userInfo) Subject() string {
	return ui.GetString("sub")
}

// GetString gets an item from the user info map as a string
func (ui userInfo) GetString(key string) string {
	v, ok := ui[key]
	if !ok {
		return ""
	}

	if s, ok := v.(string); ok {
		return s
	}

	return ""
}

// GetInt gets an item from the user info map as a string
func (ui userInfo) GetInt(key string) int {
	v, ok := ui[key]
	if !ok {
		return 0
	}

	if i, ok := v.(int); ok {
		return i
	}

	return 0
}

// GetInt64 gets an item from the user info map as a string
func (ui userInfo) GetInt64(key string) int64 {
	v, ok := ui[key]
	if !ok {
		return 0
	}

	if i, ok := v.(int64); ok {
		return i
	}

	return 0
}

// ErrNoUserInfo is the error returned by UserInfoFromContext when the
// user info is missing from the context
var ErrNoUserInfo = fmt.Errorf("missing user info")

// ErrUserInfoUnauthorized is the error returned by getUserInfo when the
// call to the hydra userinfo endpoint returns a 401
var ErrUserInfoUnauthorized = userInfoError{http.StatusUnauthorized, `{"error": "unauthorized"}`}

// UserInfoFromContext returns the userinfo on the context
func UserInfoFromContext(ctx context.Context) (UserInfo, error) {
	val := ctx.Value(ContextKeyUserInfo)
	if ui, ok := val.(UserInfo); ok {
		return ui, nil
	}
	return userInfo{}, ErrNoUserInfo
}

func getUserInfo(token string) (UserInfo, error) {

	req, err := http.NewRequest(http.MethodGet, "", nil)
	if err != nil {
		return nil, fmt.Errorf("making request: %w", err)
	}

	req.Header.Set("authorization", fmt.Sprintf("Bearer %s", token))

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("sending userinfo request: %w", err)
	}

	switch res.StatusCode {
	case http.StatusUnauthorized:
		return nil, ErrUserInfoUnauthorized
	case http.StatusOK:
		return parseUserInfo(res.Body)
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("reading error response body: %w", err)
	}

	return nil, userInfoError{res.StatusCode, body}

}

func parseUserInfo(body io.ReadCloser) (UserInfo, error) {
	ui := userInfo{}

	dec := json.NewDecoder(body)

	if err := dec.Decode(&ui); err != nil {
		return nil, fmt.Errorf("parsing userinfo response: %w", err)
	}

	return ui, nil
}
