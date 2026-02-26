package email_helper

import (
	"crypto/tls"
	"fmt"
	"mime" // 新增：用于对邮件头部中的中文字符进行 Base64 编码
	"net/smtp"
	"os"
	"strconv"
	"strings"
)

// EmailConfig SMTP 配置
type EmailConfig struct {
	Host     string
	Port     int
	Username string
	Password string
	From     string
	FromName string
}

// EmailMessage 邮件内容
type EmailMessage struct {
	To      []string // 收件人列表
	Cc      []string // 抄送列表
	Subject string   // 邮件主题
	Body    string   // 邮件正文
	IsHTML  bool     // 是否为HTML格式
}

// EmailResult 发送结果
type EmailResult struct {
	Success bool
	Error   string
}

// GetDefaultConfig 从环境变量获取默认配置
func GetDefaultConfig() EmailConfig {
	port, _ := strconv.Atoi(os.Getenv("SMTP_PORT"))
	if port == 0 {
		port = 587
	}
	return EmailConfig{
		Host:     os.Getenv("SMTP_HOST"),
		Port:     port,
		Username: os.Getenv("SMTP_USERNAME"),
		Password: os.Getenv("SMTP_PASSWORD"),
		From:     os.Getenv("SMTP_FROM"),
		FromName: os.Getenv("SMTP_FROM_NAME"),
	}
}

// SendEmail 发送邮件
func SendEmail(config EmailConfig, message EmailMessage) EmailResult {
	if config.Host == "" {
		return EmailResult{Success: false, Error: "SMTP Host 未配置"}
	}
	if len(message.To) == 0 {
		return EmailResult{Success: false, Error: "收件人不能为空"}
	}

	// 构建邮件内容：废弃 map，按严格的 SMTP 标准顺序拼接
	var msgBuilder strings.Builder

	// 1. 发件人 (处理中文昵称乱码和规范问题)
	if config.FromName != "" {
		// 使用 mime.BEncoding 进行 RFC 2047 编码
		encodedFromName := mime.BEncoding.Encode("UTF-8", config.FromName)
		msgBuilder.WriteString(fmt.Sprintf("From: %s <%s>\r\n", encodedFromName, config.From))
	} else {
		msgBuilder.WriteString(fmt.Sprintf("From: %s\r\n", config.From))
	}

	// 2. 收件人
	msgBuilder.WriteString(fmt.Sprintf("To: %s\r\n", strings.Join(message.To, ",")))

	// 3. 抄送人
	if len(message.Cc) > 0 {
		msgBuilder.WriteString(fmt.Sprintf("Cc: %s\r\n", strings.Join(message.Cc, ",")))
	}

	// 4. 邮件主题 (处理中文主题乱码，这是防止解析崩溃的关键)
	encodedSubject := mime.BEncoding.Encode("UTF-8", message.Subject)
	msgBuilder.WriteString(fmt.Sprintf("Subject: %s\r\n", encodedSubject))

	// 5. MIME 协议版本及格式
	msgBuilder.WriteString("MIME-Version: 1.0\r\n")
	if message.IsHTML {
		msgBuilder.WriteString("Content-Type: text/html; charset=UTF-8\r\n")
	} else {
		msgBuilder.WriteString("Content-Type: text/plain; charset=UTF-8\r\n")
	}

	// 6. 邮件头和正文的严格分隔符 (连续的 \r\n\r\n，因为上一行结尾有一个，这里再加一个)
	msgBuilder.WriteString("\r\n")
	
	// 7. 邮件正文
	msgBuilder.WriteString(message.Body)

	// 合并所有收件人
	allRecipients := append(message.To, message.Cc...)

	// 连接 SMTP 服务器
	addr := fmt.Sprintf("%s:%d", config.Host, config.Port)

	// 认证
	auth := smtp.PlainAuth("", config.Username, config.Password, config.Host)

	// 根据端口选择连接方式
	var err error
	if config.Port == 465 {
		// SSL 连接
		err = sendMailWithSSL(addr, auth, config.From, allRecipients, []byte(msgBuilder.String()))
	} else {
		// TLS 或普通连接 (587, 25)
		err = smtp.SendMail(addr, auth, config.From, allRecipients, []byte(msgBuilder.String()))
	}

	if err != nil {
		return EmailResult{Success: false, Error: err.Error()}
	}

	return EmailResult{Success: true, Error: ""}
}

// sendMailWithSSL 使用 SSL 发送邮件 (端口465)
func sendMailWithSSL(addr string, auth smtp.Auth, from string, to []string, msg []byte) error {
	host := strings.Split(addr, ":")[0]

	tlsConfig := &tls.Config{
		InsecureSkipVerify: true,
		ServerName:         host,
	}

	conn, err := tls.Dial("tcp", addr, tlsConfig)
	if err != nil {
		return err
	}
	defer conn.Close()

	client, err := smtp.NewClient(conn, host)
	if err != nil {
		return err
	}
	defer client.Close()

	if auth != nil {
		if err = client.Auth(auth); err != nil {
			return err
		}
	}

	if err = client.Mail(from); err != nil {
		return err
	}

	for _, addr := range to {
		if err = client.Rcpt(addr); err != nil {
			return err
		}
	}

	w, err := client.Data()
	if err != nil {
		return err
	}

	_, err = w.Write(msg)
	if err != nil {
		return err
	}

	err = w.Close()
	if err != nil {
		return err
	}

	return client.Quit()
}

// SendEmailWithDefaultConfig 使用默认配置发送邮件
func SendEmailWithDefaultConfig(message EmailMessage) EmailResult {
	config := GetDefaultConfig()
	return SendEmail(config, message)
}
