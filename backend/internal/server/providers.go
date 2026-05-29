package server

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
	ses "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/ses/v20201002"
)

func (a *App) listProviders(c *gin.Context) {
	rows, err := a.db.Query(`
		select p.id, p.name, p.type, p.enabled, p.weight, p.daily_limit, p.created_at, p.updated_at,
		       coalesce(u.sent_count, 0)
		from upstream_providers p
		left join provider_daily_usage u on u.provider_id = p.id and u.usage_date = date('now')
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
		_ = rows.Scan(&item.ID, &item.Name, &item.Type, &enabled, &item.Weight, &item.DailyLimit, &item.CreatedAt, &item.UpdatedAt, &item.TodaySent)
		item.Enabled = enabled == 1
		items = append(items, item)
	}
	c.JSON(200, items)
}

func (a *App) createProvider(c *gin.Context) {
	var body struct {
		Name       string       `json:"name"`
		Type       string       `json:"type"`
		Enabled    bool         `json:"enabled"`
		Weight     int          `json:"weight"`
		DailyLimit int          `json:"daily_limit"`
		Tencent    TencentConfig `json:"tencent_config"`
		SMTP       SMTPConfig   `json:"smtp_config"`
	}
	if err := c.BindJSON(&body); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	if strings.TrimSpace(body.Name) == "" {
		c.JSON(400, gin.H{"error": "名称不能为空"})
		return
	}
	if body.Type != "tencent" && body.Type != "smtp" {
		c.JSON(400, gin.H{"error": "上游类型必须是 tencent 或 smtp"})
		return
	}
	res, err := a.db.Exec(`insert into upstream_providers(name,type,enabled,weight,daily_limit,created_at,updated_at) values(?,?,?,?,?,?,?)`, body.Name, body.Type, boolInt(body.Enabled), body.Weight, body.DailyLimit, now(), now())
	if err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	id, _ := res.LastInsertId()
	if err := a.saveProviderConfig(id, body.Type, body.Tencent, body.SMTP); err != nil {
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
		Name       string `json:"name"`
		Enabled    bool   `json:"enabled"`
		Weight     int    `json:"weight"`
		DailyLimit int    `json:"daily_limit"`
	}
	if err := c.BindJSON(&body); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	if _, err := a.db.Exec(`update upstream_providers set name=?, enabled=?, weight=?, daily_limit=?, updated_at=? where id=?`, body.Name, boolInt(body.Enabled), body.Weight, body.DailyLimit, now(), id); err != nil {
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

func (a *App) getTencentConfig(c *gin.Context) {
	cfg, err := a.loadTencentConfig(c.Param("id"))
	if err != nil {
		c.JSON(404, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, cfg)
}

func (a *App) putTencentConfig(c *gin.Context) {
	var cfg TencentConfig
	if err := c.BindJSON(&cfg); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	if err := a.saveTencentConfig(c.Param("id"), cfg); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, gin.H{"ok": true})
}

func (a *App) getSMTPConfig(c *gin.Context) {
	cfg, err := a.loadSMTPConfig(c.Param("id"))
	if err != nil {
		c.JSON(404, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, cfg)
}

func (a *App) putSMTPConfig(c *gin.Context) {
	var cfg SMTPConfig
	if err := c.BindJSON(&cfg); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	if err := a.saveSMTPConfig(c.Param("id"), cfg); err != nil {
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

func (a *App) loadProviderDetail(providerID string) (ProviderDetail, error) {
	var detail ProviderDetail
	rows, err := a.db.Query(`select id, name, type, enabled, weight, daily_limit, created_at, updated_at from upstream_providers where id=?`, providerID)
	if err != nil {
		return detail, err
	}
	defer rows.Close()
	if !rows.Next() {
		return detail, fmt.Errorf("上游不存在")
	}
	var enabled int
	if err := rows.Scan(&detail.ID, &detail.Name, &detail.Type, &enabled, &detail.Weight, &detail.DailyLimit, &detail.CreatedAt, &detail.UpdatedAt); err != nil {
		return detail, err
	}
	detail.Enabled = enabled == 1
	if cfg, err := a.loadTencentConfig(providerID); err == nil {
		detail.TencentConfig = cfg
	}
	if cfg, err := a.loadSMTPConfig(providerID); err == nil {
		detail.SMTPConfig = cfg
	}
	rules, _ := a.listRulesForProvider(providerID)
	detail.Rules = rules
	templates, _ := a.listTemplatesForProvider(providerID)
	detail.Templates = templates
	return detail, nil
}

func (a *App) listRulesForProvider(providerID string) ([]ProviderRule, error) {
	rows, err := a.db.Query(`select id, provider_id, name, enabled, priority, script, created_at, updated_at from provider_rules where provider_id=? order by priority asc, id asc`, providerID)
	if err != nil {
		return nil, err
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
	return items, nil
}

func (a *App) listTemplatesForProvider(providerID string) ([]ProviderTemplate, error) {
	rows, err := a.db.Query(`select id, provider_id, name, status, variables, html, text, updated_at from provider_templates where provider_id=? order by id`, providerID)
	if err != nil {
		return nil, err
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
	return items, nil
}

func (a *App) saveProviderConfig(providerID int64, providerType string, tencentCfg TencentConfig, smtpCfg SMTPConfig) error {
	switch providerType {
	case "tencent":
		_, err := a.db.Exec(`insert into provider_tencent_config(provider_id,secret_id,secret_key,region,from_address,reply_to,trigger_type) values(?,?,?,?,?,?,?) on conflict(provider_id) do update set secret_id=excluded.secret_id, secret_key=excluded.secret_key, region=excluded.region, from_address=excluded.from_address, reply_to=excluded.reply_to, trigger_type=excluded.trigger_type`, providerID, tencentCfg.SecretID, tencentCfg.SecretKey, firstNonEmpty(tencentCfg.Region, "ap-guangzhou"), tencentCfg.FromAddress, tencentCfg.ReplyTo, firstNonEmpty(tencentCfg.TriggerType, "1"))
		return err
	case "smtp":
		_, err := a.db.Exec(`insert into provider_smtp_config(provider_id,host,port,username,password,from_address,reply_to) values(?,?,?,?,?,?,?) on conflict(provider_id) do update set host=excluded.host, port=excluded.port, username=excluded.username, password=excluded.password, from_address=excluded.from_address, reply_to=excluded.reply_to`, providerID, smtpCfg.Host, smtpCfg.Port, smtpCfg.Username, smtpCfg.Password, smtpCfg.FromAddress, smtpCfg.ReplyTo)
		return err
	default:
		return fmt.Errorf("未知的上游类型")
	}
}

func (a *App) saveTencentConfig(providerID string, cfg TencentConfig) error {
	typeName, err := a.providerType(providerID)
	if err != nil {
		return err
	}
	if typeName != "tencent" {
		return fmt.Errorf("该上游不是腾讯云 SES")
	}
	_, err = a.db.Exec(`insert into provider_tencent_config(provider_id,secret_id,secret_key,region,from_address,reply_to,trigger_type) values(?,?,?,?,?,?,?) on conflict(provider_id) do update set secret_id=excluded.secret_id, secret_key=excluded.secret_key, region=excluded.region, from_address=excluded.from_address, reply_to=excluded.reply_to, trigger_type=excluded.trigger_type`, providerID, cfg.SecretID, cfg.SecretKey, firstNonEmpty(cfg.Region, "ap-guangzhou"), cfg.FromAddress, cfg.ReplyTo, firstNonEmpty(cfg.TriggerType, "1"))
	return err
}

func (a *App) saveSMTPConfig(providerID string, cfg SMTPConfig) error {
	typeName, err := a.providerType(providerID)
	if err != nil {
		return err
	}
	if typeName != "smtp" {
		return fmt.Errorf("该上游不是 SMTP")
	}
	if strings.TrimSpace(cfg.Password) == "" {
		_, err = a.db.Exec(`insert into provider_smtp_config(provider_id,host,port,username,password,from_address,reply_to) values(?,?,?,?,?,?,?) on conflict(provider_id) do update set host=excluded.host, port=excluded.port, username=excluded.username, from_address=excluded.from_address, reply_to=excluded.reply_to`, providerID, cfg.Host, cfg.Port, cfg.Username, "", cfg.FromAddress, cfg.ReplyTo)
		return err
	}
	_, err = a.db.Exec(`insert into provider_smtp_config(provider_id,host,port,username,password,from_address,reply_to) values(?,?,?,?,?,?,?) on conflict(provider_id) do update set host=excluded.host, port=excluded.port, username=excluded.username, password=excluded.password, from_address=excluded.from_address, reply_to=excluded.reply_to`, providerID, cfg.Host, cfg.Port, cfg.Username, cfg.Password, cfg.FromAddress, cfg.ReplyTo)
	return err
}

func (a *App) loadTencentConfig(providerID string) (TencentConfig, error) {
	var cfg TencentConfig
	err := a.db.QueryRow(`select secret_id, secret_key, region, from_address, reply_to, trigger_type from provider_tencent_config where provider_id=?`, providerID).Scan(&cfg.SecretID, &cfg.SecretKey, &cfg.Region, &cfg.FromAddress, &cfg.ReplyTo, &cfg.TriggerType)
	return cfg, err
}

func (a *App) loadSMTPConfig(providerID string) (SMTPConfig, error) {
	var cfg SMTPConfig
	err := a.db.QueryRow(`select host, port, username, password, from_address, reply_to from provider_smtp_config where provider_id=?`, providerID).Scan(&cfg.Host, &cfg.Port, &cfg.Username, &cfg.Password, &cfg.FromAddress, &cfg.ReplyTo)
	return cfg, err
}

func (a *App) providerType(providerID string) (string, error) {
	var typ string
	if err := a.db.QueryRow(`select type from upstream_providers where id=?`, providerID).Scan(&typ); err != nil {
		return "", err
	}
	return typ, nil
}

func (a *App) validateTemplateVars(providerID uint64, templateID uint64, variables map[string]string) error {
	var raw string
	err := a.db.QueryRow(`select variables from provider_templates where id=? and provider_id=?`, templateID, providerID).Scan(&raw)
	if err != nil {
		return fmt.Errorf("模板 %d 不在缓存中，请先同步模板列表", templateID)
	}
	var required []string
	_ = json.Unmarshal([]byte(raw), &required)
	var missing []string
	for _, key := range required {
		if strings.TrimSpace(variables[key]) == "" {
			missing = append(missing, key)
		}
	}
	if len(missing) > 0 {
		return fmt.Errorf("模板变量未填满: %s", strings.Join(missing, ", "))
	}
	return nil
}

func (a *App) tencentClient(cfg TencentConfig) (*ses.Client, error) {
	secretID := strings.TrimSpace(cfg.SecretID)
	secretKey := strings.TrimSpace(cfg.SecretKey)
	region := firstNonEmpty(strings.TrimSpace(cfg.Region), "ap-guangzhou")
	if secretID == "" || secretKey == "" {
		return nil, fmt.Errorf("腾讯云 SecretId / SecretKey 未配置")
	}
	return newTencentClient(secretID, secretKey, region)
}

func newTencentClient(secretID, secretKey, region string) (*ses.Client, error) {
	cred := common.NewCredential(secretID, secretKey)
	cpf := profile.NewClientProfile()
	cpf.SignMethod = "TC3-HMAC-SHA256"
	cpf.HttpProfile.Endpoint = fmt.Sprintf("ses.%s.tencentcloudapi.com", region)
	return ses.NewClient(cred, region, cpf)
}
