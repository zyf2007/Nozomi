package server

import (
	"encoding/json"
	"strings"

	"github.com/gin-gonic/gin"
	ses "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/ses/v20201002"
)

func (a *App) listProviders(c *gin.Context) {
	rows, err := a.db.Query(`
		select p.id, p.name, p.type, p.enabled, p.weight, p.daily_limit, p.quota_timezone, p.created_at, p.updated_at
		from upstream_providers p
		order by p.weight desc, p.id asc`)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()
	items := []Provider{}
	for rows.Next() {
		var item Provider
		var enabled int
		_ = rows.Scan(&item.ID, &item.Name, &item.Type, &enabled, &item.Weight, &item.DailyLimit, &item.QuotaTimezone, &item.CreatedAt, &item.UpdatedAt)
		item.Enabled = enabled == 1
		item.QuotaTimezone = validTimezone(item.QuotaTimezone)
		_ = a.db.QueryRow(`select coalesce(sent_count, 0) from provider_daily_usage where provider_id=? and usage_date=?`, item.ID, providerUsageDate(item.QuotaTimezone)).Scan(&item.TodaySent)
		items = append(items, item)
	}
	c.JSON(200, items)
}

func (a *App) createProvider(c *gin.Context) {
	var body struct {
		Name          string        `json:"name"`
		Type          string        `json:"type"`
		Enabled       bool          `json:"enabled"`
		Weight        int           `json:"weight"`
		DailyLimit    int           `json:"daily_limit"`
		QuotaTimezone string        `json:"quota_timezone"`
		Tencent       TencentConfig `json:"tencent_config"`
		SMTP          SMTPConfig    `json:"smtp_config"`
		Resend        ResendConfig  `json:"resend_config"`
		Brevo         BrevoConfig   `json:"brevo_config"`
	}
	if err := c.BindJSON(&body); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	if strings.TrimSpace(body.Name) == "" {
		c.JSON(400, gin.H{"error": "名称不能为空"})
		return
	}
	if body.Type != "tencent" && body.Type != "smtp" && body.Type != "resend" && body.Type != "brevo" {
		c.JSON(400, gin.H{"error": "上游类型必须是 tencent、smtp、resend 或 brevo"})
		return
	}
	res, err := a.db.Exec(`insert into upstream_providers(name,type,enabled,weight,daily_limit,quota_timezone,created_at,updated_at) values(?,?,?,?,?,?,?,?)`, body.Name, body.Type, boolInt(body.Enabled), body.Weight, body.DailyLimit, validTimezone(body.QuotaTimezone), now(), now())
	if err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	id, _ := res.LastInsertId()
	if err := a.saveProviderConfig(id, body.Type, body.Tencent, body.SMTP, body.Resend, body.Brevo); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, gin.H{"ok": true, "id": id})
}

func (a *App) getProvider(c *gin.Context) {
	id := c.Param("id")
	detail, err := a.loadProviderDetail(id)
	if err != nil {
		c.JSON(404, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, detail)
}

func (a *App) updateProvider(c *gin.Context) {
	id := c.Param("id")
	var body struct {
		Name          string `json:"name"`
		Enabled       bool   `json:"enabled"`
		Weight        int    `json:"weight"`
		DailyLimit    int    `json:"daily_limit"`
		QuotaTimezone string `json:"quota_timezone"`
	}
	if err := c.BindJSON(&body); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	if _, err := a.db.Exec(`update upstream_providers set name=?, enabled=?, weight=?, daily_limit=?, quota_timezone=?, updated_at=? where id=?`, body.Name, boolInt(body.Enabled), body.Weight, body.DailyLimit, validTimezone(body.QuotaTimezone), now(), id); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, gin.H{"ok": true})
}

func (a *App) deleteProvider(c *gin.Context) {
	if _, err := a.db.Exec(`delete from upstream_providers where id=?`, c.Param("id")); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, gin.H{"ok": true})
}

func (a *App) reorderProviders(c *gin.Context) {
	var body struct {
		Items []struct {
			ID     int64 `json:"id"`
			Weight int   `json:"weight"`
		} `json:"items"`
	}
	if err := c.BindJSON(&body); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	tx, err := a.db.Begin()
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	defer tx.Rollback()
	for _, item := range body.Items {
		if _, err := tx.Exec(`update upstream_providers set weight=?, updated_at=? where id=?`, item.Weight, now(), item.ID); err != nil {
			c.JSON(400, gin.H{"error": err.Error()})
			return
		}
	}
	if err := tx.Commit(); err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, gin.H{"ok": true})
}

func (a *App) getUpstreamDispatchModeAPI(c *gin.Context) {
	mode, err := a.getUpstreamDispatchMode()
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, gin.H{"mode": mode})
}

func (a *App) putUpstreamDispatchModeAPI(c *gin.Context) {
	var body struct {
		Mode string `json:"mode"`
	}
	if err := c.BindJSON(&body); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	if err := a.setUpstreamDispatchMode(body.Mode); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, gin.H{"ok": true})
}

func (a *App) listProviderRules(c *gin.Context) {
	providerID := c.Param("id")
	rows, err := a.db.Query(`select id, provider_id, name, enabled, priority, script, created_at, updated_at from provider_rules where provider_id=? order by priority asc, id asc`, providerID)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()
	items := []ProviderRule{}
	for rows.Next() {
		var item ProviderRule
		var enabled int
		_ = rows.Scan(&item.ID, &item.ProviderID, &item.Name, &enabled, &item.Priority, &item.Script, &item.CreatedAt, &item.UpdatedAt)
		item.Enabled = enabled == 1
		items = append(items, item)
	}
	c.JSON(200, items)
}

func (a *App) createProviderRule(c *gin.Context) {
	providerID := c.Param("id")
	var body ruleBody
	if err := c.BindJSON(&body); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	if _, err := a.db.Exec(`insert into provider_rules(provider_id,name,enabled,priority,script,created_at,updated_at) values(?,?,?,?,?,?,?)`, providerID, body.Name, boolInt(body.Enabled), body.Priority, body.Script, now(), now()); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, gin.H{"ok": true})
}

func (a *App) updateProviderRule(c *gin.Context) {
	var body ruleBody
	if err := c.BindJSON(&body); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	if _, err := a.db.Exec(`update provider_rules set name=?, enabled=?, priority=?, script=?, updated_at=? where id=? and provider_id=?`, body.Name, boolInt(body.Enabled), body.Priority, body.Script, now(), c.Param("ruleID"), c.Param("id")); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, gin.H{"ok": true})
}

func (a *App) deleteProviderRule(c *gin.Context) {
	if _, err := a.db.Exec(`delete from provider_rules where id=? and provider_id=?`, c.Param("ruleID"), c.Param("id")); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, gin.H{"ok": true})
}

func (a *App) testProviderRule(c *gin.Context) {
	var body struct {
		Script     string    `json:"script"`
		ProviderID int64     `json:"provider_id"`
		Input      MailInput `json:"input"`
	}
	if err := c.BindJSON(&body); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	if strings.TrimSpace(body.Script) == "" {
		c.JSON(400, gin.H{"error": "规则脚本不能为空"})
		return
	}
	if body.Input.Headers == nil {
		body.Input.Headers = map[string]string{}
	}
	if strings.TrimSpace(body.Input.Raw) != "" && body.Input.Text == "" && body.Input.HTML == "" {
		body.Input = parseMail(body.Input.From, body.Input.To, []byte(body.Input.Raw))
	}
	result, matched, err := runRule(body.Script, body.Input)
	if err != nil {
		c.JSON(400, gin.H{"matched": false, "error": err.Error()})
		return
	}
	var validationError string
	if matched && result.TemplateID != 0 {
		if err := a.validateTemplateVars(uint64(body.ProviderID), result.TemplateID, result.Variables); err != nil {
			validationError = err.Error()
		}
	}
	c.JSON(200, gin.H{"matched": matched, "result": result, "validation_error": validationError})
}

func (a *App) listProviderTemplates(c *gin.Context) {
	rows, err := a.db.Query(`select id, provider_id, name, status, variables, html, text, updated_at from provider_templates where provider_id=? order by id`, c.Param("id"))
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()
	items := []ProviderTemplate{}
	for rows.Next() {
		var item ProviderTemplate
		var vars string
		_ = rows.Scan(&item.ID, &item.ProviderID, &item.Name, &item.Status, &vars, &item.HTML, &item.Text, &item.UpdatedAt)
		item.Variables = jsonRawArray(vars)
		items = append(items, item)
	}
	c.JSON(200, items)
}

func (a *App) syncProviderTemplates(c *gin.Context) {
	cfg, err := a.loadTencentConfig(c.Param("id"))
	if err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	client, err := a.tencentClient(cfg)
	if err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	req := ses.NewListEmailTemplatesRequest()
	limit := uint64(100)
	offset := uint64(0)
	req.Limit = &limit
	req.Offset = &offset
	resp, err := client.ListEmailTemplates(req)
	if err != nil {
		c.JSON(500, gin.H{"error": tencentErr(err)})
		return
	}
	count := 0
	for _, meta := range resp.Response.TemplatesMetadata {
		if meta == nil || meta.TemplateID == nil {
			continue
		}
		detailReq := ses.NewGetEmailTemplateRequest()
		detailReq.TemplateID = meta.TemplateID
		detail, err := client.GetEmailTemplate(detailReq)
		if err != nil {
			continue
		}
		html, text := "", ""
		if detail.Response.TemplateContent != nil {
			html = decodeB64(ptr(detail.Response.TemplateContent.Html))
			text = decodeB64(ptr(detail.Response.TemplateContent.Text))
		}
		vars := findTemplateVars(html + "\n" + text)
		varsJSON, _ := json.Marshal(vars)
		status := int64(0)
		if meta.TemplateStatus != nil {
			status = *meta.TemplateStatus
		}
		name := ptr(meta.TemplateName)
		if detail.Response.TemplateName != nil {
			name = *detail.Response.TemplateName
		}
		_, _ = a.db.Exec(`insert into provider_templates(id,provider_id,name,status,variables,html,text,updated_at) values(?,?,?,?,?,?,?,?) on conflict(id) do update set provider_id=excluded.provider_id,name=excluded.name,status=excluded.status,variables=excluded.variables,html=excluded.html,text=excluded.text,updated_at=excluded.updated_at`, *meta.TemplateID, c.Param("id"), name, status, string(varsJSON), html, text, now())
		count++
	}
	c.JSON(200, gin.H{"ok": true, "count": count})
}

func (a *App) providerType(providerID string) (string, error) {
	var typ string
	if err := a.db.QueryRow(`select type from upstream_providers where id=?`, providerID).Scan(&typ); err != nil {
		return "", err
	}
	return typ, nil
}
