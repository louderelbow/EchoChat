package oss

import (
	"errors"
	"fmt"
	"io"

	aliyunoss "github.com/aliyun/aliyun-oss-go-sdk/oss"
)

// Conf 阿里云 OSS 配置。AK/SK 建议通过环境变量注入（见 echo-api.yaml 的 ${ACID}/${ACSC}）。
type Conf struct {
	Endpoint         string
	BucketName       string
	AccessKeyID      string
	AccessKeySecret  string
	AutoCreateBucket bool
}

// Client 对 aliyun-oss-go-sdk 的轻量封装，仅暴露本项目需要的 Put/Get/URL。
type Client struct {
	bucket   *aliyunoss.Bucket
	endpoint string
	name     string
}

// New 创建 OSS 客户端；bucket 不存在且开启 AutoCreateBucket 时会自动创建（公共读）。
func New(c Conf) (*Client, error) {
	if c.Endpoint == "" || c.BucketName == "" || c.AccessKeyID == "" || c.AccessKeySecret == "" {
		return nil, errors.New("OSS 配置不完整（Endpoint/BucketName/AccessKeyID/AccessKeySecret 必填）")
	}

	// 必须显式设置超时：SDK 默认「无超时」，一旦到 OSS 的网络异常（连接卡住/下载卡住）
	// 请求会一直挂起，直到网关自身 30s 超时被中断——前端只会看到「系统错误」。
	// 这里让 OSS 调用快速失败，便于尽快回退本地磁盘或返回 404。
	cli, err := aliyunoss.New(c.Endpoint, c.AccessKeyID, c.AccessKeySecret, aliyunoss.Timeout(5, 15))
	if err != nil {
		return nil, err
	}

	exist, err := cli.IsBucketExist(c.BucketName)
	if err != nil {
		return nil, err
	}
	if !exist {
		if !c.AutoCreateBucket {
			return nil, fmt.Errorf("bucket %s 不存在，且未开启自动创建", c.BucketName)
		}
		if err = cli.CreateBucket(c.BucketName, aliyunoss.ACL(aliyunoss.ACLPublicRead)); err != nil {
			return nil, err
		}
	}

	bucket, err := cli.Bucket(c.BucketName)
	if err != nil {
		return nil, err
	}
	return &Client{bucket: bucket, endpoint: c.Endpoint, name: c.BucketName}, nil
}

// Put 上传一个对象。
func (c *Client) Put(objectKey string, r io.Reader, size int64, contentType string) error {
	opts := []aliyunoss.Option{aliyunoss.ContentLength(size)}
	if contentType != "" {
		opts = append(opts, aliyunoss.ContentType(contentType))
	}
	return c.bucket.PutObject(objectKey, r, opts...)
}

// Get 读取一个对象，返回内容和 Content-Type。
func (c *Client) Get(objectKey string) (io.ReadCloser, string, error) {
	meta, err := c.bucket.GetObjectDetailedMeta(objectKey)
	if err != nil {
		return nil, "", err
	}
	body, err := c.bucket.GetObject(objectKey)
	if err != nil {
		return nil, "", err
	}
	return body, meta.Get("Content-Type"), nil
}

// URL 返回对象的公网访问地址。
func (c *Client) URL(objectKey string) string {
	return "https://" + c.name + "." + c.endpoint + "/" + objectKey
}
