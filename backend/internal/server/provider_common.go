package server

import "fmt"

func (a *App) saveProviderConfig(providerID int64, providerType string, tencentCfg TencentConfig, smtpCfg SMTPConfig, resendCfg ResendConfig, brevoCfg BrevoConfig) error {
	switch providerType {
	case "tencent":
		return a.saveTencentConfig(fmt.Sprintf("%d", providerID), tencentCfg)
	case "smtp":
		return a.saveSMTPConfig(fmt.Sprintf("%d", providerID), smtpCfg)
	case "resend":
		return a.saveResendConfig(fmt.Sprintf("%d", providerID), resendCfg)
	case "brevo":
		return a.saveBrevoConfig(fmt.Sprintf("%d", providerID), brevoCfg)
	default:
		return fmt.Errorf("未知的上游类型")
	}
}

func (a *App) loadProviderDetail(providerID string) (ProviderDetail, error) {
	var detail ProviderDetail
	rows, err := a.db.Query(`select id, name, type, enabled, weight, daily_limit, quota_timezone, created_at, updated_at from upstream_providers where id=?`, providerID)
	if err != nil {
		return detail, err
	}
	defer rows.Close()
	if !rows.Next() {
		return detail, fmt.Errorf("上游不存在")
	}
	var enabled int
	if err := rows.Scan(&detail.ID, &detail.Name, &detail.Type, &enabled, &detail.Weight, &detail.DailyLimit, &detail.QuotaTimezone, &detail.CreatedAt, &detail.UpdatedAt); err != nil {
		return detail, err
	}
	detail.Enabled = enabled == 1
	detail.QuotaTimezone = validTimezone(detail.QuotaTimezone)
	if cfg, err := a.loadTencentConfig(providerID); err == nil {
		detail.TencentConfig = cfg
	}
	if cfg, err := a.loadSMTPConfig(providerID); err == nil {
		detail.SMTPConfig = cfg
	}
	if cfg, err := a.loadResendConfig(providerID); err == nil {
		detail.ResendConfig = cfg
	}
	if cfg, err := a.loadBrevoConfig(providerID); err == nil {
		detail.BrevoConfig = cfg
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
