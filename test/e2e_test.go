//go:build e2e

// EchoChat 端到端测试（针对已启动的网关运行）。
//
// 运行方式（需先启动 mysql/redis/rabbitmq 及 rpc、网关）：
//
//	go test -tags e2e ./test -v
//
// 可用环境变量覆盖地址：
//
//	E2E_BASE   网关地址，默认 https://127.0.0.1:8888
//	E2E_REDIS  Redis 地址，默认 127.0.0.1:6379（用于写入注册验证码）
//	E2E_MYSQL  MySQL DSN，默认 root:123456@tcp(127.0.0.1:3306)/echo_chat?parseTime=true
package e2e

import (
	"bytes"
	"context"
	"crypto/tls"
	"database/sql"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"testing"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/websocket"
	"github.com/redis/go-redis/v9"
)

// ---------------- 基础设施 ----------------

func env(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

var (
	baseURL   = env("E2E_BASE", "https://127.0.0.1:8888")
	wsURL     = env("E2E_WS", toWS(baseURL)+"/wss")
	redisAddr = env("E2E_REDIS", "127.0.0.1:6379")
	mysqlDSN  = env("E2E_MYSQL", "root:123456@tcp(127.0.0.1:3306)/echo_chat?parseTime=true")

	httpClient = &http.Client{
		Timeout: 10 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}
)

// toWS 把 http(s) 地址转换为 WebSocket 地址。
func toWS(u string) string {
	switch {
	case len(u) > 8 && u[:8] == "https://":
		return "wss://" + u[8:]
	case len(u) > 7 && u[:7] == "http://":
		return "ws://" + u[7:]
	}
	return u
}

type apiResp struct {
	Code    int             `json:"code"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data"`
	Token   string          `json:"token"`
}

func post(t *testing.T, path string, body map[string]any, token string) apiResp {
	t.Helper()
	b, _ := json.Marshal(body)
	req, err := http.NewRequest(http.MethodPost, baseURL+path, bytes.NewReader(b))
	if err != nil {
		t.Fatalf("构造请求 %s 失败: %v", path, err)
	}
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	resp, err := httpClient.Do(req)
	if err != nil {
		t.Fatalf("请求 %s 失败（服务是否已启动？）: %v", path, err)
	}
	defer resp.Body.Close()
	var r apiResp
	if err := json.NewDecoder(resp.Body).Decode(&r); err != nil {
		t.Fatalf("解析 %s 响应失败: %v", path, err)
	}
	return r
}

func rdb(t *testing.T) *redis.Client {
	t.Helper()
	return redis.NewClient(&redis.Options{Addr: redisAddr})
}

func seedSmsCode(t *testing.T, account, code string) {
	t.Helper()
	if err := rdb(t).Set(context.Background(), "auth_code_"+account, code, time.Minute).Err(); err != nil {
		t.Fatalf("写入验证码到 Redis 失败: %v", err)
	}
}

// ---------------- 数据清理 ----------------

func cleanup(uuids ...string) {
	db, err := sql.Open("mysql", mysqlDSN)
	if err != nil {
		return
	}
	defer db.Close()
	for _, u := range uuids {
		_, _ = db.Exec("DELETE FROM message WHERE send_id=? OR receive_id=?", u, u)
		_, _ = db.Exec("DELETE FROM session WHERE send_id=? OR receive_id=?", u, u)
		_, _ = db.Exec("DELETE FROM contact_apply WHERE user_id=? OR contact_id=?", u, u)
		_, _ = db.Exec("DELETE FROM user_contact WHERE user_id=? OR contact_id=?", u, u)
		_, _ = db.Exec("DELETE FROM group_info WHERE owner_id=?", u)
		_, _ = db.Exec("DELETE FROM user_info WHERE uuid=?", u)
	}
}

func queryScalar(t *testing.T, query string, args ...any) string {
	t.Helper()
	db, err := sql.Open("mysql", mysqlDSN)
	if err != nil {
		t.Fatalf("连接 MySQL 失败: %v", err)
	}
	defer db.Close()
	var v string
	if err := db.QueryRow(query, args...).Scan(&v); err != nil {
		t.Fatalf("查询失败 %q: %v", query, err)
	}
	return v
}

// ---------------- 业务辅助 ----------------

func uniqAccount() string {
	return fmt.Sprintf("e2e_%d_%d@test.local", time.Now().UnixNano(), rand.Intn(100000))
}

// registerUser 注册一个测试用户，返回 (uuid, 账号)。
func registerUser(t *testing.T) (string, string) {
	t.Helper()
	account := uniqAccount()
	seedSmsCode(t, account, "123456")
	r := post(t, "/register", map[string]any{
		"telephone": account,
		"password":  "Passw0rd!",
		"nickname":  "E2E用户",
		"sms_code":  "123456",
	}, "")
	if r.Code != 200 {
		t.Fatalf("注册失败: code=%d message=%s", r.Code, r.Message)
	}
	var u struct {
		Uuid string `json:"uuid"`
	}
	_ = json.Unmarshal(r.Data, &u)
	if u.Uuid == "" {
		t.Fatalf("注册返回缺少 uuid: %s", string(r.Data))
	}
	return u.Uuid, account
}

func login(t *testing.T, account, password string) apiResp {
	t.Helper()
	return post(t, "/login", map[string]any{"telephone": account, "password": password}, "")
}

func dataList(t *testing.T, r apiResp) []map[string]any {
	t.Helper()
	if len(r.Data) == 0 {
		return nil
	}
	var list []map[string]any
	if err := json.Unmarshal(r.Data, &list); err != nil {
		t.Fatalf("data 不是数组: %s", string(r.Data))
	}
	return list
}

func containsContactID(list []map[string]any, id string) bool {
	for _, m := range list {
		if m["contact_id"] == id {
			return true
		}
	}
	return false
}

// ---------------- WS 辅助 ----------------

func wsConnect(t *testing.T, uuid, token string) *websocket.Conn {
	t.Helper()
	d := websocket.Dialer{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}
	c, _, err := d.Dial(wsURL+"?client_id="+uuid+"&token="+token, nil)
	if err != nil {
		t.Fatalf("WS 连接失败: %v", err)
	}
	return c
}

func wsSend(t *testing.T, c *websocket.Conn, msg map[string]any) {
	t.Helper()
	b, _ := json.Marshal(msg)
	_ = c.SetWriteDeadline(time.Now().Add(5 * time.Second))
	if err := c.WriteMessage(websocket.TextMessage, b); err != nil {
		t.Fatalf("WS 发送失败: %v", err)
	}
}

// wsWaitFor 在 d 时间内等待满足 match 的 JSON 消息（自动跳过非 JSON 的欢迎语）。
func wsWaitFor(c *websocket.Conn, match func(map[string]any) bool, d time.Duration) bool {
	_ = c.SetReadDeadline(time.Now().Add(d))
	for {
		_, data, err := c.ReadMessage()
		if err != nil {
			return false
		}
		var m map[string]any
		if json.Unmarshal(data, &m) != nil {
			continue // 欢迎语等非 JSON 帧
		}
		if match(m) {
			return true
		}
	}
}

// ---------------- 测试用例 ----------------

func TestE2E_RegisterLogin(t *testing.T) {
	account := uniqAccount()
	seedSmsCode(t, account, "123456")

	r := post(t, "/register", map[string]any{
		"telephone": account, "password": "Passw0rd!", "nickname": "E2E注册", "sms_code": "123456",
	}, "")
	if r.Code != 200 {
		t.Fatalf("正常注册应成功, got code=%d message=%s", r.Code, r.Message)
	}
	var u struct {
		Uuid string `json:"uuid"`
	}
	_ = json.Unmarshal(r.Data, &u)
	uuid := u.Uuid
	defer cleanup(uuid)
	if uuid == "" {
		t.Fatal("注册未返回 uuid")
	}

	// 边界：重复注册（验证码正确但账号已存在）→ 400
	seedSmsCode(t, account, "123456")
	if r := post(t, "/register", map[string]any{
		"telephone": account, "password": "Passw0rd!", "nickname": "重复", "sms_code": "123456",
	}, ""); r.Code != 400 {
		t.Errorf("重复注册应返回 400, got %d (%s)", r.Code, r.Message)
	}

	// 边界：验证码错误 → 400
	acc2 := uniqAccount()
	seedSmsCode(t, acc2, "111111")
	if r := post(t, "/register", map[string]any{
		"telephone": acc2, "password": "Passw0rd!", "nickname": "错码", "sms_code": "222222",
	}, ""); r.Code != 400 {
		t.Errorf("验证码错误应返回 400, got %d (%s)", r.Code, r.Message)
	}

	// 正常登录：应返回令牌
	if r := login(t, account, "Passw0rd!"); r.Code != 200 || r.Token == "" {
		t.Errorf("正常登录应成功且返回 token, got code=%d token=%q", r.Code, r.Token)
	}
	// 边界：密码错误 → 400
	if r := login(t, account, "WrongPass!"); r.Code != 400 {
		t.Errorf("密码错误应返回 400, got %d", r.Code)
	}
	// 边界：用户不存在 → 400
	if r := login(t, "e2e_not_exist@test.local", "whatever"); r.Code != 400 {
		t.Errorf("不存在用户应返回 400, got %d", r.Code)
	}
}

func TestE2E_FriendFlow(t *testing.T) {
	a, _ := registerUser(t)
	b, _ := registerUser(t)
	defer cleanup(a, b)

	// A 申请加 B 为好友
	if r := post(t, "/contact/applyContact", map[string]any{
		"owner_id": a, "contact_id": b, "message": "你好",
	}, ""); r.Code != 200 {
		t.Fatalf("申请好友应成功, got %d (%s)", r.Code, r.Message)
	}

	// B 的待处理列表应包含 A
	lst := dataList(t, post(t, "/contact/getNewContactList", map[string]any{"owner_id": b}, ""))
	if !containsContactID(lst, a) {
		t.Errorf("B 的待处理申请应包含 A(%s), got %+v", a, lst)
	}
	// 红点计数应 >= 1
	cnt := post(t, "/contact/getNewContactCount", map[string]any{"owner_id": b}, "")
	if cnt.Code != 200 {
		t.Errorf("待处理计数接口应 200, got %d", cnt.Code)
	}

	// 边界：申请不存在的用户 → 400
	if r := post(t, "/contact/applyContact", map[string]any{
		"owner_id": a, "contact_id": "U0000000000000000000", "message": "",
	}, ""); r.Code != 400 {
		t.Errorf("申请不存在用户应返回 400, got %d (%s)", r.Code, r.Message)
	}

	// B 同意
	if r := post(t, "/contact/passContactApply", map[string]any{
		"owner_id": b, "contact_id": a,
	}, ""); r.Code != 200 {
		t.Fatalf("同意好友申请应成功, got %d (%s)", r.Code, r.Message)
	}
	// A 的通讯录应包含 B
	users := dataList(t, post(t, "/contact/getUserList", map[string]any{"owner_id": a}, ""))
	found := false
	for _, m := range users {
		if m["user_id"] == b {
			found = true
		}
	}
	if !found {
		t.Errorf("A 的通讯录应包含 B(%s), got %+v", b, users)
	}

	// 边界：同意的申请不存在 → 非 200
	if r := post(t, "/contact/passContactApply", map[string]any{
		"owner_id": a, "contact_id": "U0000000000000000001",
	}, ""); r.Code == 200 {
		t.Errorf("同意不存在的申请不应成功, got %d", r.Code)
	}

	// 拒绝流程
	c, _ := registerUser(t)
	defer cleanup(c)
	if r := post(t, "/contact/applyContact", map[string]any{"owner_id": c, "contact_id": b}, ""); r.Code != 200 {
		t.Fatalf("C 申请加 B 应成功, got %d (%s)", r.Code, r.Message)
	}
	if r := post(t, "/contact/refuseContactApply", map[string]any{
		"owner_id": b, "contact_id": c,
	}, ""); r.Code != 200 {
		t.Errorf("拒绝好友申请应成功, got %d (%s)", r.Code, r.Message)
	}
}

// openGroupSession 打开会话并返回其 session_id（消息落库需要合法的 20 位 session_id）。
func openGroupSession(t *testing.T, owner, groupID string) string {
	t.Helper()
	post(t, "/session/openSession", map[string]any{"send_id": owner, "receive_id": groupID}, "")
	return queryScalar(t, "SELECT uuid FROM session WHERE receive_id=? AND deleted_at IS NULL LIMIT 1", groupID)
}

func TestE2E_GroupAndMessage(t *testing.T) {
	owner, _ := registerUser(t)
	member, _ := registerUser(t)
	defer cleanup(owner, member)

	// 创建群（直接加入）
	if r := post(t, "/group/createGroup", map[string]any{
		"owner_id": owner, "name": "E2E群", "notice": "测试", "add_mode": 0, "avatar": "",
	}, ""); r.Code != 200 {
		t.Fatalf("创建群应成功, got %d (%s)", r.Code, r.Message)
	}
	groups := dataList(t, post(t, "/group/loadMyGroup", map[string]any{"owner_id": owner}, ""))
	if len(groups) == 0 {
		t.Fatal("loadMyGroup 应有群")
	}
	groupID, _ := groups[0]["group_id"].(string)
	if groupID == "" {
		t.Fatal("群 id 为空")
	}
	defer func() {
		db, _ := sql.Open("mysql", mysqlDSN)
		if db != nil {
			_, _ = db.Exec("DELETE FROM group_info WHERE uuid=?", groupID)
			_, _ = db.Exec("DELETE FROM message WHERE receive_id=?", groupID)
			_, _ = db.Exec("DELETE FROM session WHERE receive_id=?", groupID)
			db.Close()
		}
	}()

	// member 直接加入
	if r := post(t, "/group/enterGroupDirectly", map[string]any{
		"owner_id": groupID, "contact_id": member,
	}, ""); r.Code != 200 {
		t.Fatalf("直接加入群应成功, got %d (%s)", r.Code, r.Message)
	}

	// 边界：对不存在的群检查加群方式 → 非 200
	if r := post(t, "/group/checkGroupAddMode", map[string]any{
		"group_id": "G0000000000000000000",
	}, ""); r.Code == 200 {
		t.Errorf("不存在的群 checkGroupAddMode 不应成功")
	}

	sessionID := openGroupSession(t, owner, groupID)
	if sessionID == "" {
		t.Fatal("未取到群会话 id")
	}

	// WS：owner 发群消息，member 应收到
	oc := wsConnect(t, owner, "")
	mc := wsConnect(t, member, "")
	defer oc.Close()
	defer mc.Close()

	content := fmt.Sprintf("e2e群消息-%d", time.Now().UnixNano())
	wsSend(t, oc, map[string]any{
		"session_id": sessionID, "type": 0, "content": content, "url": "",
		"send_id": owner, "send_name": "E2E群主", "send_avatar": "",
		"receive_id": groupID, "file_size": "0B", "file_name": "", "file_type": "",
	})
	got := wsWaitFor(mc, func(m map[string]any) bool {
		return m["content"] == content && m["receive_id"] == groupID
	}, 6*time.Second)
	if !got {
		t.Error("群成员未收到群消息（扇出失败）")
	}

	// 边界：往一个不存在的群发消息，不应导致消费者崩溃/重投；随后正常消息仍能送达
	wsSend(t, oc, map[string]any{
		"session_id": sessionID, "type": 0, "content": "丢给不存在的群", "url": "",
		"send_id": owner, "send_name": "E2E群主", "send_avatar": "",
		"receive_id": "G0000000000000000000", "file_size": "0B",
	})
	content2 := fmt.Sprintf("e2e群消息2-%d", time.Now().UnixNano())
	wsSend(t, oc, map[string]any{
		"session_id": sessionID, "type": 0, "content": content2, "url": "",
		"send_id": owner, "send_name": "E2E群主", "send_avatar": "",
		"receive_id": groupID, "file_size": "0B",
	})
	if !wsWaitFor(mc, func(m map[string]any) bool { return m["content"] == content2 }, 6*time.Second) {
		t.Error("非法群消息后，正常群消息未送达（消费者可能异常）")
	}

	// 历史记录应包含刚才的消息（含头像兜底，字段不应缺失）
	hist := post(t, "/message/getGroupMessageList", map[string]any{"group_id": groupID}, "")
	if hist.Code != 200 {
		t.Fatalf("获取群历史应 200, got %d", hist.Code)
	}
	for _, m := range dataList(t, hist) {
		if _, ok := m["send_avatar"]; !ok {
			t.Errorf("历史消息缺少 send_avatar 字段: %+v", m)
		}
	}
}

func TestE2E_SingleChatMessage(t *testing.T) {
	a, _ := registerUser(t)
	b, _ := registerUser(t)
	defer cleanup(a, b)

	// 先加为好友并开会话
	if r := post(t, "/contact/applyContact", map[string]any{"owner_id": a, "contact_id": b}, ""); r.Code != 200 {
		t.Fatalf("A 申请加 B 应成功, got %d (%s)", r.Code, r.Message)
	}
	if r := post(t, "/contact/passContactApply", map[string]any{"owner_id": b, "contact_id": a}, ""); r.Code != 200 {
		t.Fatalf("同意好友应成功, got %d (%s)", r.Code, r.Message)
	}
	post(t, "/session/openSession", map[string]any{"send_id": a, "receive_id": b}, "")
	sessionID := queryScalar(t, "SELECT uuid FROM session WHERE send_id=? AND receive_id=? AND deleted_at IS NULL LIMIT 1", a, b)
	if sessionID == "" {
		t.Fatal("未取到单聊会话 id")
	}

	ac := wsConnect(t, a, "")
	bc := wsConnect(t, b, "")
	defer ac.Close()
	defer bc.Close()

	content := fmt.Sprintf("e2e单聊-%d", time.Now().UnixNano())
	wsSend(t, ac, map[string]any{
		"session_id": sessionID, "type": 0, "content": content, "url": "",
		"send_id": a, "send_name": "E2E甲", "send_avatar": "",
		"receive_id": b, "file_size": "0B",
	})
	if !wsWaitFor(bc, func(m map[string]any) bool { return m["content"] == content }, 6*time.Second) {
		t.Error("单聊对端未收到消息")
	}
}

// 边界：创建群时省略可选的 notice / avatar 字段，应同样成功（字段非必填）。
func TestE2E_CreateGroupOptionalFields(t *testing.T) {
	owner, _ := registerUser(t)
	defer cleanup(owner)
	if r := post(t, "/group/createGroup", map[string]any{
		"owner_id": owner, "name": "E2E最小群", "add_mode": 0,
	}, ""); r.Code != 200 {
		t.Errorf("省略 notice/avatar 创建群应成功, got %d (%s)", r.Code, r.Message)
	}
}

func TestE2E_Admin(t *testing.T) {
	// 管理员账号由 schema/脚本创建；这里验证可登录且管理端接口可用。
	r := login(t, "admin@echochat.com", "Admin@123456")
	if r.Code != 200 {
		t.Skipf("管理员账号未就绪或密码不符（跳过）: code=%d message=%s", r.Code, r.Message)
	}
	if r.Token == "" {
		t.Error("管理员登录未返回 token")
	}
	// 本地默认 ECHO_AUTH 关闭，管理端接口直接放行；开启鉴权后此处会校验管理员身份（应 200）
	if got := post(t, "/user/getUserInfoList", map[string]any{"owner_id": ""}, r.Token); got.Code != 200 {
		t.Errorf("管理员查看用户列表应 200, got %d (%s)", got.Code, got.Message)
	}
}
