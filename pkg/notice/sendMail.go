package notice

import (
	"cmdb/config"
	"cmdb/pkg/middleware"
	"crypto/tls"
	"fmt"
	"gopkg.in/gomail.v2"
	"io"
)

// EmailInfo 邮件信息结构体
type EmailInfo struct {
	To          []string          // 收件人邮箱列表
	Cc          []string          // 抄送人邮箱列表
	Subject     string            // 邮件主题
	Body        string            // 邮件正文
	BodyType    string            // 正文类型：text/plain 或 text/html
	Attachments []EmailAttachment // 附件列表
}

// EmailAttachment 邮件附件结构体
type EmailAttachment struct {
	Filename string // 附件文件名
	Content  []byte // 附件内容（字节数组）
	Path     string // 附件文件路径（如果不提供Content，则使用Path）
}

// SendEmail 发送邮件
// 统一的邮件发送函数，支持普通邮件和带附件邮件
func SendEmail(emailInfo *EmailInfo) error {
	if emailInfo == nil {
		return fmt.Errorf("邮件信息不能为空")
	}

	if len(emailInfo.To) == 0 {
		return fmt.Errorf("收件人不能为空")
	}

	// 获取邮件配置
	cfg := config.GetConfig()
	smtpHost := cfg.Email.SMTPHost
	smtpPort := cfg.Email.SMTPPort
	smtpUser := cfg.Email.SMTPUser
	smtpPass := cfg.Email.SMTPPass
	from := cfg.Email.From

	if smtpHost == "" || smtpPort == 0 || smtpUser == "" || smtpPass == "" || from == "" {
		return fmt.Errorf("邮件配置不完整")
	}

	middleware.Logger.Infof("开始发送邮件：SMTP服务器=%s:%d, 发件人=%s, 收件人=%v, 抄送=%v, 主题=%s",
		smtpHost, smtpPort, from, emailInfo.To, emailInfo.Cc, emailInfo.Subject)

	// 创建邮件
	m := gomail.NewMessage()
	m.SetHeader("From", from)
	m.SetHeader("To", emailInfo.To...)

	if len(emailInfo.Cc) > 0 {
		m.SetHeader("Cc", emailInfo.Cc...)
	}

	m.SetHeader("Subject", emailInfo.Subject)

	// 设置邮件正文
	bodyType := emailInfo.BodyType
	if bodyType == "" {
		bodyType = "text/html" // 默认HTML格式
	}
	m.SetBody(bodyType, emailInfo.Body)

	// 添加附件
	for i, attachment := range emailInfo.Attachments {
		if attachment.Content != nil {
			// 使用内存中的附件内容
			middleware.Logger.Infof("添加内存附件[%d]: %s (%d 字节)",
				i+1, attachment.Filename, len(attachment.Content))
			m.Attach(attachment.Filename,
				gomail.SetCopyFunc(func(w io.Writer) error {
					_, err := w.Write(attachment.Content)
					return err
				}))
		} else if attachment.Path != "" {
			// 使用文件路径添加附件
			middleware.Logger.Infof("添加文件附件[%d]: %s", i+1, attachment.Path)
			m.Attach(attachment.Path)
		}
	}

	// 配置SMTP拨号器
	d := gomail.NewDialer(smtpHost, smtpPort, smtpUser, smtpPass)
	d.TLSConfig = &tls.Config{InsecureSkipVerify: true} // 跳过证书验证，生产环境建议移除

	// 发送邮件
	middleware.Logger.Infof("连接SMTP服务器并发送邮件...")
	if err := d.DialAndSend(m); err != nil {
		middleware.Logger.Errorf("邮件发送失败: %v", err)
		return fmt.Errorf("邮件发送失败: %v", err)
	}

	middleware.Logger.Infof("邮件发送成功，收件人: %v, 抄送: %v, 主题: %s",
		emailInfo.To, emailInfo.Cc, emailInfo.Subject)
	return nil
}
