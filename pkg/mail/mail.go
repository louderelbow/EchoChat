package mail

import (
	"crypto/tls"
	"fmt"
	"mime"
	"net"
	"net/smtp"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/zeromicro/go-zero/core/logx"

	"echochat/pkg/constants"
	"echochat/pkg/random"
	"echochat/pkg/redisx"
)

// SMTP 配置全部来自环境变量，缺省值适配 QQ 邮箱：
//
//	ECHO_MAIL_HOST  SMTP 服务器，默认 smtp.qq.com
//	ECHO_MAIL_PORT  端口，默认 465（SSL）
//	ECHO_MAIL_USER  发件人邮箱（QQ 邮箱地址），必填
//	ECHO_MAIL_PASS  QQ 邮箱 SMTP 授权码，必填
//	ECHO_MAIL_FROM  发件人显示地址，默认取 ECHO_MAIL_USER
//	ECHO_MAIL_TO    验证码收件人，默认取 ECHO_MAIL_USER
type smtpConf struct {
	Host string
	Port int
	User string
	Pass string
	From string
	To   string
}

func env(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func loadConf() smtpConf {
	port, err := strconv.Atoi(env("ECHO_MAIL_PORT", "465"))
	if err != nil {
		port = 465
	}
	user := os.Getenv("ECHO_MAIL_USER")
	return smtpConf{
		Host: env("ECHO_MAIL_HOST", "smtp.qq.com"),
		Port: port,
		User: user,
		Pass: os.Getenv("ECHO_MAIL_PASS"),
		From: env("ECHO_MAIL_FROM", user),
		To:   env("ECHO_MAIL_TO", user),
	}
}

// send 使用隐式 TLS（SMTPS，默认 465 端口）投递一封纯文本邮件。
func (c smtpConf) send(to, subject, body string) error {
	addr := net.JoinHostPort(c.Host, strconv.Itoa(c.Port))
	conn, err := tls.Dial("tcp", addr, &tls.Config{ServerName: c.Host})
	if err != nil {
		return err
	}
	client, err := smtp.NewClient(conn, c.Host)
	if err != nil {
		return err
	}
	defer client.Close()

	if err = client.Auth(smtp.PlainAuth("", c.User, c.Pass, c.Host)); err != nil {
		return err
	}
	if err = client.Mail(c.From); err != nil {
		return err
	}
	if err = client.Rcpt(to); err != nil {
		return err
	}
	w, err := client.Data()
	if err != nil {
		return err
	}
	if _, err = w.Write([]byte(buildMessage(c.From, to, subject, body))); err != nil {
		return err
	}
	if err = w.Close(); err != nil {
		return err
	}
	return client.Quit()
}

func buildMessage(from, to, subject, body string) string {
	// 中文主题需用 MIME 编码，避免乱码
	return fmt.Sprintf(
		"From: %s\r\nTo: %s\r\nSubject: %s\r\nMIME-Version: 1.0\r\nContent-Type: text/plain; charset=UTF-8\r\n\r\n%s\r\n",
		from, to, mime.QEncoding.Encode("UTF-8", subject), body,
	)
}

// Send 使用环境变量配置发送一封纯文本邮件；to 为空时默认发往 ECHO_MAIL_TO。
// 供调用方复用与自测，凭据始终由环境变量提供，不会出现在代码或命令行中。
func Send(to, subject, body string) error {
	c := loadConf()
	if to == "" {
		to = c.To
	}
	if c.User == "" || c.Pass == "" || to == "" {
		return fmt.Errorf("邮件服务未配置：请先设置环境变量 ECHO_MAIL_USER / ECHO_MAIL_PASS")
	}
	return c.send(to, subject, body)
}

// VerificationCode 生成验证码并发送到该账号邮箱，返回 (message, code)：
// 0 成功，-1 系统错误，-2 业务错误。account 既是收件邮箱，也是 Redis 存储验证码的键后缀。
func VerificationCode(account string) (string, int) {
	c := loadConf()
	if c.User == "" || c.Pass == "" {
		msg := "邮件服务未配置：请先设置环境变量 ECHO_MAIL_USER / ECHO_MAIL_PASS"
		logx.Error(msg)
		return constants.SYSTEM_ERROR, -1
	}

	// 收件人：账号本身是邮箱时直接发给它；否则（例如手机号账号）回退到 ECHO_MAIL_TO
	to := account
	if !strings.Contains(account, "@") {
		to = c.To
	}
	if to == "" {
		msg := "无法确定验证码收件邮箱：账号不是邮箱，且未设置 ECHO_MAIL_TO"
		logx.Error(msg)
		return constants.SYSTEM_ERROR, -1
	}

	key := "auth_code_" + account
	code, err := redisx.GetKey(key)
	if err != nil {
		logx.Error(err.Error())
		return constants.SYSTEM_ERROR, -1
	}

	if code != "" {
		// 直接返回，验证码还没过期，用户应该去输验证码
		message := "目前还不能发送验证码，请输入已发送的验证码"
		logx.Info(message)
		return message, -2
	}

	// 验证码过期，重新生成
	code = strconv.Itoa(random.GetRandomInt(6))
	if err := redisx.SetKeyEx(key, code, time.Minute); err != nil { // 1分钟有效
		logx.Error(err.Error())
		return constants.SYSTEM_ERROR, -1
	}

	subject := "Echo Chat 验证码"
	body := fmt.Sprintf("您的验证码是：%s\n\n该验证码 1 分钟内有效，请勿泄露给他人。", code)
	if err := c.send(to, subject, body); err != nil {
		logx.Error(err.Error())
		return constants.SYSTEM_ERROR, -1
	}
	return "验证码发送成功，请到邮箱查收", 0
}
