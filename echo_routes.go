package main

import (
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"echochat/internal/config"
	"echochat/internal/middleware"
	"echochat/internal/svc"
	"echochat/pkg/authtoken"
	"echochat/pkg/constants"
	"echochat/pkg/es"
	"echochat/pkg/oss"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/rest"
	"github.com/zeromicro/go-zero/rest/httpx"
	"github.com/zeromicro/go-zero/rest/pathvar"
)

// registerExtraRoutes 注册 api 脚手架难以表达的路由：WebSocket、文件上传、静态资源
func registerExtraRoutes(server *rest.Server, ctx *svc.ServiceContext, c config.Config) {
	// 初始化阿里云 OSS 客户端；未配置或初始化失败时回退到本地磁盘存储
	ossClient, err := oss.New(oss.Conf{
		Endpoint:         c.Oss.Endpoint,
		BucketName:       c.Oss.BucketName,
		AccessKeyID:      c.Oss.AccessKeyID,
		AccessKeySecret:  c.Oss.AccessKeySecret,
		AutoCreateBucket: c.Oss.AutoCreateBucket,
	})
	if err != nil {
		logx.Errorf("阿里云 OSS 未启用（%v），上传改为写入本地磁盘", err)
	} else {
		logx.Infof("阿里云 OSS 已启用：%s", ossClient.URL(""))
	}

	server.AddRoutes([]rest.Route{
		{Method: http.MethodGet, Path: "/wss", Handler: wsHandler(ctx)},
		{Method: http.MethodPost, Path: "/message/uploadAvatar", Handler: uploadHandler(c.Static.AvatarPath, ossClient, "avatars")},
		{Method: http.MethodPost, Path: "/message/uploadFile", Handler: uploadHandler(c.Static.FilePath, ossClient, "files")},
		{Method: http.MethodGet, Path: "/static/avatars/:file", Handler: staticHandler(c.Static.AvatarPath, ossClient, "avatars")},
		{Method: http.MethodGet, Path: "/static/files/:file", Handler: staticHandler(c.Static.FilePath, ossClient, "files")},
		{Method: http.MethodPost, Path: "/message/search", Handler: searchMessageHandler(ctx)},
		// 指标暴露端点：供观测栈刮取 IM 业务指标（连接数/下行投递等）
		{Method: http.MethodGet, Path: "/metrics", Handler: promhttp.Handler().ServeHTTP},
	})
}

// searchMessageHandler 聊天记录全文检索（走 Elasticsearch）
func searchMessageHandler(ctx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req es.SearchReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}
		if strings.TrimSpace(req.Keyword) == "" {
			httpx.OkJsonCtx(r.Context(), w, map[string]interface{}{"code": 400, "message": "请输入搜索关键词"})
			return
		}
		if !ctx.Es.Enabled() {
			httpx.OkJsonCtx(r.Context(), w, map[string]interface{}{"code": 400, "message": "检索服务未启用"})
			return
		}
		resp, err := ctx.Es.SearchMessages(r.Context(), req)
		if err != nil {
			logx.Errorf("ES 检索失败: %v", err)
			httpx.OkJsonCtx(r.Context(), w, map[string]interface{}{"code": 500, "message": "检索服务异常"})
			return
		}
		httpx.OkJsonCtx(r.Context(), w, map[string]interface{}{"code": 200, "message": "检索成功", "data": resp})
	}
}

func wsHandler(ctx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		clientID := r.URL.Query().Get("client_id")
		if clientID == "" {
			httpx.OkJsonCtx(r.Context(), w, map[string]interface{}{"code": 400, "message": "clientId获取失败"})
			return
		}
		// 开启鉴权时，校验令牌与 client_id 是否一致，避免凭 uuid 冒充他人连接
		if middleware.Enabled() {
			uuid, _, ok := authtoken.Lookup(middleware.TokenFromRequest(r))
			if !ok || uuid != clientID {
				httpx.OkJsonCtx(r.Context(), w, map[string]interface{}{"code": 401, "message": "登录状态已失效，请重新登录"})
				return
			}
		}
		// 升级成功后由 Hub 接管连接，无法再以 JSON 响应
		_ = ctx.Hub.Serve(w, r, clientID)
	}
}

// isAllowedUpload 按上传类型校验扩展名：头像只允许常见图片；普通文件拦截可执行/脚本等危险类型。
func isAllowedUpload(prefix, name string) bool {
	ext := strings.ToLower(filepath.Ext(name))
	if prefix == "avatars" {
		switch ext {
		case ".jpg", ".jpeg", ".png", ".gif", ".webp", ".bmp":
			return true
		}
		return false
	}
	switch ext {
	case ".php", ".jsp", ".jspx", ".asp", ".aspx", ".exe", ".sh", ".bat", ".cmd",
		".jar", ".msi", ".dll", ".html", ".htm", ".svg", ".js":
		return false
	}
	return true
}

// saveLocal 把上传内容写入本地 root 目录（OSS 未配置或上传失败时的兜底）。
func saveLocal(root, name string, r io.Reader) error {
	out, err := os.Create(filepath.Join(root, name))
	if err != nil {
		return err
	}
	defer out.Close()
	_, err = io.Copy(out, r)
	return err
}

// uploadHandler 处理文件上传：优先写入 OSS，OSS 不可用或写入失败时回退到本地 root 目录。
func uploadHandler(root string, ossClient *oss.Client, prefix string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fail := func() {
			httpx.OkJsonCtx(r.Context(), w, map[string]interface{}{"code": 500, "message": constants.SYSTEM_ERROR})
		}
		if err := r.ParseMultipartForm(constants.FILE_MAX_SIZE); err != nil {
			logx.Errorf("解析上传表单失败: %v", err)
			fail()
			return
		}
		for key := range r.MultipartForm.File {
			file, header, err := r.FormFile(key)
			if err != nil {
				logx.Errorf("读取上传文件失败: %v", err)
				fail()
				return
			}
			name := filepath.Base(header.Filename)
			if !isAllowedUpload(prefix, name) {
				_ = file.Close()
				logx.Errorf("拒绝不支持的%s上传：%s", prefix, name)
				fail()
				return
			}

			if ossClient != nil {
				err = ossClient.Put(prefix+"/"+name, file, header.Size, header.Header.Get("Content-Type"))
				if err == nil {
					_ = file.Close()
					continue
				}
				logx.Errorf("上传到 OSS 失败（%v），改为写入本地磁盘", err)
				if _, seekErr := file.Seek(0, io.SeekStart); seekErr != nil {
					_ = file.Close()
					fail()
					return
				}
			}

			if err = saveLocal(root, name, file); err != nil {
				_ = file.Close()
				logx.Errorf("写入本地文件失败: %v", err)
				fail()
				return
			}
			_ = file.Close()
		}
		httpx.OkJsonCtx(r.Context(), w, map[string]interface{}{"code": 200, "message": "上传成功"})
	}
}

// staticHandler 读取静态资源：优先从 OSS 拉取并转发，取不到时回退本地磁盘。
func staticHandler(root string, ossClient *oss.Client, prefix string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		name := filepath.Base(pathvar.Vars(r)["file"])
		if name == "." || name == string(filepath.Separator) || strings.Contains(name, "..") {
			http.NotFound(w, r)
			return
		}
		if ossClient != nil {
			body, contentType, err := ossClient.Get(prefix + "/" + name)
			if err == nil {
				defer body.Close()
				if contentType != "" {
					w.Header().Set("Content-Type", contentType)
				}
				_, _ = io.Copy(w, body)
				return
			}
			logx.Errorf("从 OSS 读取对象失败（%v），改为读取本地磁盘", err)
		}
		http.ServeFile(w, r, filepath.Join(root, name))
	}
}
