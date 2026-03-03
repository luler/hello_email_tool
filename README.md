# Email Tool

基于 Gin 框架的邮件发送服务

## 安装

```bash
go mod tidy
```

## 配置

复制 `.env.example` 为 `.env` 并修改配置：

```bash
cp .env.example .env
```

主要配置项：

```env
# SMTP 邮件配置
SMTP_HOST=smtp.example.com
SMTP_PORT=465
SMTP_USERNAME=your_email@example.com
SMTP_PASSWORD=your_password
SMTP_FROM=your_email@example.com
SMTP_FROM_NAME=EmailTool

# 邮件接口授权码（支持多个，逗号隔开）
EMAIL_AUTH_CODE=your_auth_code

# 页面访问授权码（单个）
WEB_AUTH_CODE=your_web_auth_code
```

## 运行

```bash
# 直接运行
go run main.go serve

# 热重载运行
gin serve run main.go
```

## API 接口

### 发送邮件

**请求地址：** `/api/email`

**请求方式：** `GET` / `POST`

**请求参数：**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| auth_code | string | 是 | 授权码，需与环境变量 EMAIL_AUTH_CODE 之一匹配（支持多个，逗号隔开） |
| to | string | 是 | 收件人，多个用逗号分隔 |
| cc | string | 否 | 抄送人，多个用逗号分隔 |
| subject | string | 是 | 邮件主题 |
| body | string | 是 | 邮件正文 |
| is_html | bool | 否 | 是否为 HTML 格式，支持 `1`/`true` |
| from_name | string | 否 | 发件人名称，默认使用环境变量 SMTP_FROM_NAME |

**请求示例：**

GET 请求：
```
/api/email?auth_code=xxx&to=test@qq.com&subject=测试邮件&body=邮件内容&is_html=1
```

POST 请求：
```json
{
  "auth_code": "xxx",
  "to": "test@qq.com,test2@qq.com",
  "cc": "cc@qq.com",
  "subject": "测试邮件",
  "body": "<h1>HTML内容</h1>",
  "is_html": true,
  "from_name": "自定义发件人"
}
```

**响应示例：**

成功：
```json
{
  "code": 200,
  "data": [],
  "message": "邮件发送成功"
}
```

失败：
```json
{
  "code": 400,
  "data": [],
  "msg": "授权码错误"
}
```
