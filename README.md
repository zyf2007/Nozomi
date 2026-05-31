# Nozomi Relay

Nozomi Relay 是一个基于 Go 的轻量级邮件中转服务，用于从下游 SMTP 转发到腾讯云 SES 等多个邮件推送服务 API。  

由于严格的风控政策，目前大量邮件提供商已经不再支持使用 SMTP 发送邮件，而是强制使用自家的 API。但是，常见的下游应用程序（例如 Rauthy, Gitea 等）仍然只能支持 SMTP 发信。本服务支持接收 SMTP 邮件后用 JavaScript 规则提取变量（例如验证码），再调用上游模板 Sendmail 接口。  

![Nozomi](https://imgoss.qijieya.cn/imgoss/59/6a1a6f8365bd2.png)
## 功能

- Go + Gin + SQLite 后端
- 内置 SMTP relay，支持配置多个下游 SMTP 账号
- 腾讯云 SES / Resend / Brevo API 支持
- 同步腾讯云模板列表与模板变量
- JavaScript 规则脚本提取下游邮件内容并映射模板变量
- 支持按时区设置不同邮件提供商的每日额度重置时间，避免超额扣费
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

如果你从局域网 IP 访问前端，比如 `http://192.168.31.221:5173`，需要在后端 `.env` 里把 `NOZOMI_CORS_ORIGINS` 加上对应的完整 origin，例如：

```bash
NOZOMI_CORS_ORIGINS=http://localhost:5173,http://127.0.0.1:5173,http://192.168.31.221:5173
```

开发模式下，前端 `npm run dev` 仍然是通过 Vite 代理把 `/api` 转发到 `http://127.0.0.1:5000`，不是直接从浏览器打 5000。

## 构建模式

项目支持两种构建模式：

### 前后端分离模式

只构建 Go 后端，前端单独部署或用 Vite 开发服务运行：

```bash
./scripts/build.sh separate
./dist/nozomi-relay
```

这是默认的 API-only 二进制，不包含前端静态文件。

### 合并模式

先构建前端，再把 `Nozomi-Admin/dist` 嵌入 Go 二进制，由 Go 同时提供 API 和管理面板静态服务：

```bash
./scripts/build.sh embedded
./dist/nozomi-relay-embedded
```

启动后访问后端 HTTP 地址即可打开管理面板，例如默认的 `http://127.0.0.1:5000`。`/api/*` 仍然走后端 API，其它路径会优先返回静态文件，未命中时回退到 `index.html` 以支持前端路由。

运行时可以用命令行参数控制是否启用内置静态服务：

```bash
./dist/nozomi-relay-embedded -web auto
./dist/nozomi-relay-embedded -web off
```

普通分离模式二进制即使使用 `-web auto` 也不会提供前端静态文件，因为构建时没有嵌入 dist。

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
  templateId: 100001, // 腾讯云模板 ID
  subject: input.subject, // 邮件主题
  variables: { // 模板变量映射
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
