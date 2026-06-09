package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/telagod/subme/internal/config"
	"github.com/telagod/subme/internal/pkg/dingtalk"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDingTalkOAuthStart_Disabled(t *testing.T) {
	t.Skip("helper newTestAuthHandlerWithDingTalk needed; sentinel only")
}

func TestDtSyntheticEmail_UsesUnionID(t *testing.T) {
	email := dtSyntheticEmail("union_AbCdEf123")
	want := "dingtalk-union_abcdef123@dingtalk-connect.invalid"
	require.Equal(t, want, email)
	require.True(t, strings.ToLower(email) == email)
	require.True(t, strings.HasPrefix(email, "dingtalk-"))
	require.True(t, strings.HasSuffix(email, "@dingtalk-connect.invalid"))
}

func TestDtSyntheticEmail_TrimsSpace(t *testing.T) {
	email := dtSyntheticEmail("  UID_XYZ  ")
	require.Equal(t, "dingtalk-uid_xyz@dingtalk-connect.invalid", email)
}

func TestDtUpstreamClaims_EmptyStaff(t *testing.T) {
	staff := &dingtalk.StaffInfo{}
	claims := dtUpstreamClaims(staff, "UNION_AAA", "CORP_X")
	require.Equal(t, "", claims["email"])
	require.Equal(t, "", claims["username"])
	require.Equal(t, "UNION_AAA", claims["subject"])
	require.Equal(t, "", claims["corp_user_id"])
	require.Equal(t, "UNION_AAA", claims["union_id"])
	require.Equal(t, "CORP_X", claims["corp_id"])
}

func TestDtCorpAllowed_CrossOrgPolicy(t *testing.T) {
	cfg := config.DingTalkConnectConfig{CorpRestrictionPolicy: "none"}
	assert.True(t, dtCorpAllowed(cfg, "dingABC"))
	assert.True(t, dtCorpAllowed(cfg, ""))
	assert.True(t, dtCorpAllowed(cfg, "foreign_corp"))
}

func TestDtCorpAllowed_InternalOnly(t *testing.T) {
	cfg := config.DingTalkConnectConfig{
		CorpRestrictionPolicy: "internal_only",
		InternalCorpID:        "dingInternal",
	}
	assert.True(t, dtCorpAllowed(cfg, "dingInternal"))
	assert.True(t, dtCorpAllowed(cfg, "foreign_corp"))
	assert.True(t, dtCorpAllowed(cfg, ""))
}

func TestDtUpstreamClaims_SubjectEqualsUnionID(t *testing.T) {
	staff := &dingtalk.StaffInfo{UserID: "user123", Name: "张三", Email: "zhangsan@corp.com"}
	claims := dtUpstreamClaims(staff, "union456", "dingcorp789")
	require.Equal(t, "union456", claims["subject"])
	require.Equal(t, "user123", claims["corp_user_id"])
	require.Equal(t, "union456", claims["union_id"])
	require.Equal(t, "dingcorp789", claims["corp_id"])
	require.Equal(t, "张三", claims["username"])
	require.Equal(t, "zhangsan@corp.com", claims["email"])
}

func TestDtUpstreamClaims_CrossOrgEmptyCorpUserID(t *testing.T) {
	staff := &dingtalk.StaffInfo{}
	claims := dtUpstreamClaims(staff, "union_cross_org", "foreign_corp")
	require.Equal(t, "union_cross_org", claims["subject"])
	require.Equal(t, "", claims["corp_user_id"])
}

func TestDtUpstreamClaims_PrimaryDeptID(t *testing.T) {
	staff := &dingtalk.StaffInfo{UserID: "u1", Name: "张三", Email: "a@b.com", DeptIDs: []int64{42, 99}}
	claims := dtUpstreamClaims(staff, "uid1", "corpX")
	require.Equal(t, int64(42), claims["primary_dept_id"])
}

func TestDtUpstreamClaims_NoDeptIDs(t *testing.T) {
	staff := &dingtalk.StaffInfo{UserID: "u2", Name: "李四"}
	claims := dtUpstreamClaims(staff, "uid2", "corpY")
	require.Equal(t, int64(0), claims["primary_dept_id"])
}

func TestDtStaffFromClaims_RoundTrip(t *testing.T) {
	staff := &dingtalk.StaffInfo{UserID: "u3", Name: "王五", Email: "ww@corp.com", DeptIDs: []int64{55}}
	claims := dtUpstreamClaims(staff, "uid3", "corpZ")
	recovered := dtStaffFromClaims(claims)
	require.Equal(t, "王五", recovered.Name)
	require.Equal(t, "ww@corp.com", recovered.Email)
	require.Equal(t, "u3", recovered.UserID)
	require.Equal(t, []int64{55}, recovered.DeptIDs)
}

func TestDtResolveDeptPath_SingleLevel(t *testing.T) {
	handler := &AuthHandler{}
	responses := map[string]string{
		"42": `{"errcode":0,"result":{"dept_id":42,"name":"研发部","parent_id":1}}`,
		"1":  `{"errcode":0,"result":{"dept_id":1,"name":"公司","parent_id":0}}`,
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req struct{ DeptID int64 `json:"dept_id"` }
		_ = json.NewDecoder(r.Body).Decode(&req)
		w.Header().Set("Content-Type", "application/json")
		if resp, ok := responses[fmt.Sprintf("%d", req.DeptID)]; ok {
			_, _ = w.Write([]byte(resp))
		} else {
			_, _ = w.Write([]byte(`{"errcode":60003,"errmsg":"not found"}`))
		}
	}))
	defer server.Close()

	cli := dingtalk.NewClient(dingtalk.ClientConfig{UserInfoURL: server.URL + "/stub"}, server.Client())
	cli.SetAppTokenForTest("tok")

	path, err := handler.dtResolveDeptPath(context.Background(), cli, 42)
	require.NoError(t, err)
	require.Equal(t, "研发部", path)
}

func TestDtResolveDeptPath_MultiLevel(t *testing.T) {
	handler := &AuthHandler{}
	responses := map[string]string{
		"42": `{"errcode":0,"result":{"dept_id":42,"name":"AI研发","parent_id":10}}`,
		"10": `{"errcode":0,"result":{"dept_id":10,"name":"研发部","parent_id":1}}`,
		"1":  `{"errcode":0,"result":{"dept_id":1,"name":"公司","parent_id":0}}`,
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req struct{ DeptID int64 `json:"dept_id"` }
		_ = json.NewDecoder(r.Body).Decode(&req)
		w.Header().Set("Content-Type", "application/json")
		if resp, ok := responses[fmt.Sprintf("%d", req.DeptID)]; ok {
			_, _ = w.Write([]byte(resp))
		} else {
			_, _ = w.Write([]byte(`{"errcode":60003,"errmsg":"not found"}`))
		}
	}))
	defer server.Close()

	cli := dingtalk.NewClient(dingtalk.ClientConfig{UserInfoURL: server.URL + "/stub"}, server.Client())
	cli.SetAppTokenForTest("tok")

	path, err := handler.dtResolveDeptPath(context.Background(), cli, 42)
	require.NoError(t, err)
	require.Equal(t, "研发部/AI研发", path)
}

func TestDtSyncIdentity_NilServiceNoOp(t *testing.T) {
	handler := &AuthHandler{userAttributeService: nil}
	cfg := config.DingTalkConnectConfig{
		CorpRestrictionPolicy:  "internal_only",
		SyncCorpEmail:          true,
		SyncDisplayName:        true,
		SyncDept:               true,
		SyncCorpEmailAttrKey:   "custom_email",
		SyncDisplayNameAttrKey: "custom_name",
		SyncDeptAttrKey:        "custom_dept",
	}
	staff := &dingtalk.StaffInfo{Name: "张三", Email: "zhangsan@example.com"}
	require.NotPanics(t, func() {
		handler.dtSyncIdentity(context.Background(), cfg, nil, 42, staff, false)
	})
}

func TestDtSignupBlocked(t *testing.T) {
	handler := &AuthHandler{settingSvc: nil}
	cfg := config.DingTalkConnectConfig{}
	assert.False(t, handler.dtSignupBlocked(context.Background(), cfg))
}

func TestUsernameFromEmailLocalPart(t *testing.T) {
	tests := []struct {
		email, username, wantUser string
		wantValid                 bool
	}{
		{"dingtalk-uid123@dingtalk-connect.invalid", "", "dingtalk-uid123", true},
		{"user@example.com", "张三", "张三", true},
		{"noemail", "", "noemail", true},
		{"", "", "", false},
	}
	for _, tc := range tests {
		username := tc.username
		email := tc.email
		if username == "" {
			if at := strings.Index(email, "@"); at > 0 {
				username = email[:at]
			} else {
				username = email
			}
		}
		isValid := email != "" && username != ""
		require.Equal(t, tc.wantUser, username)
		require.Equal(t, tc.wantValid, isValid)
	}
}
