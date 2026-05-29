# Nozomi Relay

SMTP 到腾讯云 SES API 的邮件中转服务。Rauthy 等下游系统仍按 SMTP 发信，本服务接收邮件后用规则提取变量，再调用腾讯云 `SendEmail` 模板接口。

## 功能

- Go + Gin + SQLite 后端
- 内置 SMTP relay，支持配置多个下游 SMTP 账号
- 腾讯云 SES SecretId / SecretKey / Region / 发信地址配置
- 同步腾讯云模板列表与模板变量
- JavaScript 规则脚本提取下游邮件内容并映射模板变量
- 发送历史、错误日志、腾讯云回调事件记录
- Ant Design 管理面板，含统计、配置、模板、规则、账号、日志

## 开发运行

后端：

```bash
cd backend
cp .env.example .env
go run ./cmd/server
```

前端：

```bash
cd Nozomi-Admin
npm run dev
```

默认地址：

- 管理面板：http://127.0.0.1:5173
- 后端 API：http://127.0.0.1:5000
- SMTP relay：127.0.0.1:2525

默认管理员账号来自 `backend/.env`，默认下游 SMTP 账号为 `rauthy / change-me`。

## 规则脚本

规则脚本运行在 Go 内嵌 JavaScript 引擎中，后端提供 `input`：

```js
{
  from,
  to,
  subject,
  text,
  html,
  headers,
  raw
}
```

脚本返回 `null` 表示不匹配；匹配时返回：

```js
({
  templateId: 100001,
  subject: input.subject,
  variables: {
    code: "123456",
    action: "登录"
  }
})
```

后端会检查 `variables` 是否填满腾讯云模板中解析到的 `{{变量名}}`。

## 腾讯云回调

在腾讯云 SES 控制台配置回调地址：

```text
https://你的域名/api/callback/tencent
```

当前版本记录 `delivered`、`bounce`、`dropped`、`open`、`click` 等事件，并根据 `bulkId` / `messageId` 关联发送历史。
