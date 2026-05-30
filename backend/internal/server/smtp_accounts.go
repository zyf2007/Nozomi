package server

import (
	"fmt"
	"net"
	stdsmtp "net/smtp"
	"strings"

	"github.com/gin-gonic/gin"
)

func (a *App) listSMTPAccounts(c *gin.Context) {
	rows, err := a.db.Query(`select id, username, active, allowed_provider_ids, created_at, updated_at from smtp_accounts order by id desc`)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()
	items := []SMTPAccount{}
	for rows.Next() {
		var item SMTPAccount
		var active int
		var allowed string
		_ = rows.Scan(&item.ID, &item.Username, &active, &allowed, &item.CreatedAt, &item.UpdatedAt)
		item.Active = active == 1
		item.AllowedProviderIDs = jsonRawIntArray(allowed)
		items = append(items, item)
	}
	c.JSON(200, items)
}

func (a *App) createSMTPAccount(c *gin.Context) {
	var body struct {
		Username           string  `json:"username"`
		Password           string  `json:"password"`
		Active             bool    `json:"active"`
		AllowedProviderIDs []int64 `json:"allowed_provider_ids"`
	}
	if err := c.BindJSON(&body); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	_, err := a.db.Exec(`insert into smtp_accounts(username,password,active,allowed_provider_ids,created_at,updated_at) values(?,?,?,?,?,?)`, body.Username, body.Password, boolInt(body.Active), jsonString(body.AllowedProviderIDs), now(), now())
	if err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, gin.H{"ok": true})
}

func (a *App) updateSMTPAccount(c *gin.Context) {
	id := c.Param("id")
	var body struct {
		Username           string  `json:"username"`
		Password           string  `json:"password"`
		Active             bool    `json:"active"`
		AllowedProviderIDs []int64 `json:"allowed_provider_ids"`
	}
	if err := c.BindJSON(&body); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	if body.Password == "" {
		_, _ = a.db.Exec(`update smtp_accounts set username=?, active=?, allowed_provider_ids=?, updated_at=? where id=?`, body.Username, boolInt(body.Active), jsonString(body.AllowedProviderIDs), now(), id)
	} else {
		_, _ = a.db.Exec(`update smtp_accounts set username=?, password=?, active=?, allowed_provider_ids=?, updated_at=? where id=?`, body.Username, body.Password, boolInt(body.Active), jsonString(body.AllowedProviderIDs), now(), id)
	}
	c.JSON(200, gin.H{"ok": true})
}

func (a *App) deleteSMTPAccount(c *gin.Context) {
	_, _ = a.db.Exec(`delete from smtp_accounts where id=?`, c.Param("id"))
	c.JSON(200, gin.H{"ok": true})
}

func (a *App) testSMTPAccount(c *gin.Context) {
	var body struct {
		From        string           `json:"from"`
		To          []string         `json:"to"`
		Subject     string           `json:"subject"`
		Text        string           `json:"text"`
		HTML        string           `json:"html"`
		Attachments []MailAttachment `json:"attachments"`
	}
	if err := c.BindJSON(&body); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	if len(body.To) == 0 {
		c.JSON(400, gin.H{"error": "至少填写一个收件人"})
		return
	}
	var username, password string
	var active int
	err := a.db.QueryRow(`select username,password,active from smtp_accounts where id=?`, c.Param("id")).Scan(&username, &password, &active)
	if err != nil {
		c.JSON(404, gin.H{"error": "SMTP 账号不存在"})
		return
	}
	if active != 1 {
		c.JSON(400, gin.H{"error": "SMTP 账号未启用"})
		return
	}
	from := firstNonEmpty(strings.TrimSpace(body.From), "tester@nozomi-relay.local")
	subject := firstNonEmpty(strings.TrimSpace(body.Subject), "Nozomi Relay 测试邮件")
	raw := buildMimeMessage(from, body.To, subject, body.Text, body.HTML, nil, body.Attachments)
	addr, authHost, err := a.localSMTPDialTarget()
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	auth := stdsmtp.PlainAuth("", username, password, authHost)
	if err := stdsmtp.SendMail(addr, auth, from, body.To, []byte(raw)); err != nil {
		c.JSON(500, gin.H{"error": err.Error(), "raw": raw})
		return
	}
	c.JSON(200, gin.H{"ok": true, "raw": raw})
}

func (a *App) localSMTPDialTarget() (string, string, error) {
	host, port, err := net.SplitHostPort(a.settings.SMTPAddr)
	if err != nil {
		return "", "", fmt.Errorf("SMTP 监听地址格式错误: %w", err)
	}
	if host == "" || host == "0.0.0.0" || host == "::" || host == "[::]" {
		host = "127.0.0.1"
	}
	return net.JoinHostPort(host, port), host, nil
}
