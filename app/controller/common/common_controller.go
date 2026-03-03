package common

import (
	"gin_base/app/helper/email_helper"
	"gin_base/app/helper/exception_helper"
	"gin_base/app/helper/request_helper"
	"gin_base/app/helper/response_helper"
	"github.com/gin-gonic/gin"
	"os"
	"strings"
)

func Test(c *gin.Context) {
	response_helper.Success(c, "访问成功")
}

func Email(c *gin.Context) {
	type Param struct {
		AuthCode string      `json:"auth_code" mapstructure:"auth_code" validate:"required" label:"授权码"`
		To       string      `json:"to" mapstructure:"to" validate:"required" label:"收件人"`
		Cc       string      `json:"cc" mapstructure:"cc" validate:"omitempty" label:"抄送"`
		Subject  string      `json:"subject" mapstructure:"subject" validate:"required" label:"邮件主题"`
		Body     string      `json:"body" mapstructure:"body" validate:"required" label:"邮件正文"`
		IsHTML   interface{} `json:"is_html" mapstructure:"is_html" validate:"omitempty" label:"是否HTML格式"`
		FromName string      `json:"from_name" mapstructure:"from_name" validate:"omitempty" label:"发件人名称"`
	}
	var param Param
	request_helper.InputStruct(c, &param)

	// 验证授权码（支持多个，逗号隔开）
	authCodes := strings.Split(os.Getenv("EMAIL_AUTH_CODE"), ",")
	validCode := false
	for _, code := range authCodes {
		if strings.TrimSpace(code) == param.AuthCode {
			validCode = true
			break
		}
	}
	if !validCode {
		exception_helper.CommonException("授权码错误")
	}

	// 获取请求IP
	requestIP := c.ClientIP()

	// 获取SMTP配置
	config := email_helper.GetDefaultConfig()

	// 如果请求参数中传入了 from_name，则覆盖默认值
	if param.FromName != "" {
		config.FromName = param.FromName
	}

	// 构建邮件消息（逗号分隔转数组）
	var toList, ccList []string
	if param.To != "" {
		toList = strings.Split(param.To, ",")
	}
	if param.Cc != "" {
		ccList = strings.Split(param.Cc, ",")
	}
	// 解析 is_html 参数（兼容字符串、数字、布尔）
	var isHTML bool
	switch v := param.IsHTML.(type) {
	case bool:
		isHTML = v
	case string:
		isHTML = v == "1" || v == "true"
	case float64:
		isHTML = v == 1
	case int:
		isHTML = v == 1
	}

	message := email_helper.EmailMessage{
		To:      toList,
		Cc:      ccList,
		Subject: param.Subject,
		Body:    param.Body,
		IsHTML:  isHTML,
	}

	// 发送邮件
	result := email_helper.SendEmail(config, message)

	// 记录日志
	email_helper.LogEmailRequest(requestIP, message, config, result, param)

	// 返回结果
	if !result.Success {
		exception_helper.CommonException(result.Error)
	}
	response_helper.Success(c, "邮件发送成功")
}
