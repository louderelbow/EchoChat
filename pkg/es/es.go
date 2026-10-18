package es

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// Conf Elasticsearch 连接配置
// Address 为空表示不启用 ES；Username/Password 仅在 ES 开启认证时需要，故均为可选
type Conf struct {
	Address  string `json:",optional"` // 如 http://127.0.0.1:9200
	Index    string `json:",optional"` // 索引名，默认 chat_message
	Username string `json:",optional"`
	Password string `json:",optional"`
}

// Client 轻量 ES 客户端（直接用 HTTP，避免额外 SDK 依赖）
type Client struct {
	conf Conf
	http *http.Client
}

// New 创建客户端；Address 为空时返回 nil（表示未启用 ES）
func New(conf Conf) *Client {
	if conf.Address == "" {
		return nil
	}
	if conf.Index == "" {
		conf.Index = "chat_message"
	}
	return &Client{
		conf: conf,
		http: &http.Client{Timeout: 5 * time.Second},
	}
}

// Enabled 是否已启用
func (c *Client) Enabled() bool { return c != nil && c.conf.Address != "" }

func (c *Client) do(ctx context.Context, method, path string, body any, out any) error {
	var rdr io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return err
		}
		rdr = bytes.NewReader(b)
	}
	req, err := http.NewRequestWithContext(ctx, method, c.conf.Address+path, rdr)
	if err != nil {
		return err
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if c.conf.Username != "" {
		req.SetBasicAuth(c.conf.Username, c.conf.Password)
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	data, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 400 {
		return fmt.Errorf("es %s %s -> %d: %s", method, path, resp.StatusCode, string(data))
	}
	if out != nil && len(data) > 0 {
		return json.Unmarshal(data, out)
	}
	return nil
}

// EnsureIndex 若索引不存在则按 IK 分词建索引（消息全文检索）
func (c *Client) EnsureIndex(ctx context.Context) error {
	if !c.Enabled() {
		return nil
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodHead, c.conf.Address+"/"+c.conf.Index, nil)
	if err != nil {
		return err
	}
	if c.conf.Username != "" {
		req.SetBasicAuth(c.conf.Username, c.conf.Password)
	}
	resp, err := c.http.Do(req)
	if err == nil {
		_ = resp.Body.Close()
		if resp.StatusCode == http.StatusOK {
			return nil
		}
	}

	mapping := map[string]any{
		"mappings": map[string]any{
			"properties": map[string]any{
				"uuid":       map[string]any{"type": "keyword"},
				"session_id": map[string]any{"type": "keyword"},
				"receive_id": map[string]any{"type": "keyword"},
				"send_id":    map[string]any{"type": "keyword"},
				"send_name":  map[string]any{"type": "keyword"},
				"content": map[string]any{
					"type":            "text",
					"analyzer":        "ik_max_word",
					"search_analyzer": "ik_smart",
				},
				"created_at": map[string]any{
					"type":   "date",
					"format": "yyyy-MM-dd HH:mm:ss||strict_date_optional_time||epoch_millis",
				},
			},
		},
	}
	return c.do(ctx, http.MethodPut, "/"+c.conf.Index, mapping, nil)
}

// IndexMessage 写入/更新一条消息文档（幂等：用 uuid 作为文档 id）
func (c *Client) IndexMessage(ctx context.Context, doc map[string]any) error {
	if !c.Enabled() {
		return nil
	}
	id, _ := doc["uuid"].(string)
	if id == "" {
		return fmt.Errorf("es index: missing uuid")
	}
	return c.do(ctx, http.MethodPut, "/"+c.conf.Index+"/_doc/"+id, doc, nil)
}

// SearchReq 聊天记录检索请求（除 keyword 外均为可选，避免 go-zero 解析把未传字段当必填）
type SearchReq struct {
	Keyword   string `json:"keyword,optional"`
	SessionId string `json:"session_id,optional"` // 兼容字段：按会话 id 过滤
	ReceiveId string `json:"receive_id,optional"` // 兼容字段：按接收方过滤
	SelfId    string `json:"self_id,optional"`    // 当前用户 id（单聊过滤用）
	ContactId string `json:"contact_id,optional"` // 会话对象：群 id 或对方用户 id
	Size      int    `json:"size,optional"`
	From      int    `json:"from,optional"`
}

// escapeWildcard 转义 wildcard 查询里的通配符，避免用户输入 * ? 被当通配符
func escapeWildcard(s string) string {
	s = strings.ReplaceAll(s, "\\", "\\\\")
	s = strings.ReplaceAll(s, "*", "\\*")
	s = strings.ReplaceAll(s, "?", "\\?")
	return s
}

// SearchHit 单条命中
type SearchHit struct {
	Uuid      string `json:"uuid"`
	SendId    string `json:"send_id"`
	SendName  string `json:"send_name"`
	Content   string `json:"content"`
	CreatedAt string `json:"created_at"`
	Highlight string `json:"highlight"`
}

// SearchResp 命中结果
type SearchResp struct {
	Total int         `json:"total"`
	Hits  []SearchHit `json:"hits"`
}

// SearchMessages 按会话/接收方 + 关键词检索聊天记录，返回高亮片段
func (c *Client) SearchMessages(ctx context.Context, req SearchReq) (*SearchResp, error) {
	if !c.Enabled() {
		return &SearchResp{}, nil
	}
	if req.Size <= 0 || req.Size > 50 {
		req.Size = 20
	}
	// 关键词：match（分词）+ wildcard（词内子串）双保险。
	// 中文分词会把「你好」切成单个 token，单纯 match 时搜「你」搜不到「你好」，
	// 加上 wildcard 按 token 子串匹配即可覆盖单字/子串检索。
	must := []any{
		map[string]any{"bool": map[string]any{
			"should": []any{
				map[string]any{"match": map[string]any{"content": req.Keyword}},
				map[string]any{"wildcard": map[string]any{"content": map[string]any{
					"value": "*" + escapeWildcard(req.Keyword) + "*",
				}}},
			},
			"minimum_should_match": 1,
		}},
	}
	filter := []any{}
	if req.ContactId != "" {
		// 按会话过滤：群聊用 receive_id（群 id）；单聊用「双方 send_id/receive_id」双向匹配。
		// 不能按 session_id 过滤——同一会话每个成员各有一条 session 行，只会搜到自己发的消息。
		if strings.HasPrefix(req.ContactId, "G") {
			filter = append(filter, map[string]any{"term": map[string]any{"receive_id": req.ContactId}})
		} else if req.SelfId != "" {
			me, peer := req.SelfId, req.ContactId
			filter = append(filter, map[string]any{"bool": map[string]any{
				"should": []any{
					map[string]any{"bool": map[string]any{"must": []any{
						map[string]any{"term": map[string]any{"send_id": me}},
						map[string]any{"term": map[string]any{"receive_id": peer}},
					}}},
					map[string]any{"bool": map[string]any{"must": []any{
						map[string]any{"term": map[string]any{"send_id": peer}},
						map[string]any{"term": map[string]any{"receive_id": me}},
					}}},
				},
				"minimum_should_match": 1,
			}})
		}
	} else {
		if req.SessionId != "" {
			filter = append(filter, map[string]any{"term": map[string]any{"session_id": req.SessionId}})
		}
		if req.ReceiveId != "" {
			filter = append(filter, map[string]any{"term": map[string]any{"receive_id": req.ReceiveId}})
		}
	}
	body := map[string]any{
		"from": req.From,
		"size": req.Size,
		"query": map[string]any{
			"bool": map[string]any{"must": must, "filter": filter},
		},
		"highlight": map[string]any{
			"pre_tags":  []string{"<em>"},
			"post_tags": []string{"</em>"},
			"fields":    map[string]any{"content": map[string]any{}},
		},
		"sort": []any{map[string]any{"created_at": "desc"}},
	}

	var raw struct {
		Hits struct {
			Total struct {
				Value int `json:"value"`
			} `json:"total"`
			Hits []struct {
				Source    map[string]any `json:"_source"`
				Highlight map[string][]string `json:"highlight"`
			} `json:"hits"`
		} `json:"hits"`
	}
	if err := c.do(ctx, http.MethodPost, "/"+c.conf.Index+"/_search", body, &raw); err != nil {
		return nil, err
	}

	resp := &SearchResp{Total: raw.Hits.Total.Value}
	for _, h := range raw.Hits.Hits {
		hit := SearchHit{
			Uuid:      str(h.Source["uuid"]),
			SendId:    str(h.Source["send_id"]),
			SendName:  str(h.Source["send_name"]),
			Content:   str(h.Source["content"]),
			CreatedAt: str(h.Source["created_at"]),
		}
		if frags := h.Highlight["content"]; len(frags) > 0 {
			hit.Highlight = frags[0]
		}
		if hit.Highlight == "" {
			hit.Highlight = hit.Content
		}
		resp.Hits = append(resp.Hits, hit)
	}
	return resp, nil
}

func str(v any) string {
	if s, ok := v.(string); ok {
		return s
	}
	return ""
}
