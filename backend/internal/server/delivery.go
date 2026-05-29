package server

import (
	"encoding/json"
	"fmt"
	"mime"
	"net"
	"net/smtp"
	"strconv"
	"strings"
	"time"
	ses "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/ses/v20201002"
)

type upstreamCandidate struct {
	Provider
	TencentConfig TencentConfig
	SMTPConfig    SMTPConfig
	Rules         []ProviderRule
	Templates     []ProviderTemplate
}

func (a *App) processMail(input MailInput, downstreamAccountID int64) error {
	rawTo, _ := json.Marshal(input.To)
	res, err := a.db.Exec(`insert into messages(downstream_account_id,downstream_from,downstream_to,subject,raw,text_body,html_body,status,created_at,updated_at) values(?,?,?,?,?,?,?,?,?,?)`, downstreamAccountID, input.From, string(rawTo), input.Subject, input.Raw, input.Text, input.HTML, "received", now(), now())
	if err != nil {
		return err
	}
	msgID, _ := res.LastInsertId()

	providerIDs, err := a.allowedProviderIDsForDownstream(downstreamAccountID)
	if err != nil {
		a.failMessage(msgID, "failed", err.Error())
		return err
	}
	candidates, err := a.loadUpstreamCandidates(providerIDs)
	if err != nil {
		a.failMessage(msgID, "failed", err.Error())
		return err
	}
	if len(candidates) == 0 {
		err = fmt.Errorf("该下游没有启用任何上游")
		a.failMessage(msgID, "failed", err.Error())
		return err
	}
	mode, err := a.getUpstreamDispatchMode()
	if err != nil {
		a.failMessage(msgID, "failed", err.Error())
		return err
	}
	selected := a.pickUpstreamCandidates(downstreamAccountID, candidates, mode)
	var lastErr error
	for _, candidate := range selected {
		ruleID, templateID, providerMessageID, providerType, templateData, err := a.sendThroughProvider(candidate, input)
		if err != nil {
			lastErr = err
			continue
		}
		templateJSON, _ := json.Marshal(templateData)
		var ruleAny any
		var templateAny any
		if ruleID != 0 {
			ruleAny = ruleID
		}
		if templateID != 0 {
			templateAny = templateID
		}
		_, _ = a.db.Exec(`update messages set provider_id=?,provider_type=?,rule_id=?,template_id=?,template_data=?,status='sent',provider_message_id=?,updated_at=? where id=?`, candidate.ID, providerType, ruleAny, templateAny, string(templateJSON), providerMessageID, now(), msgID)
		return nil
	}
	if lastErr == nil {
		lastErr = fmt.Errorf("没有可用的上游")
	}
	a.failMessage(msgID, "failed", lastErr.Error())
	return lastErr
}

func (a *App) loadUpstreamCandidates(providerIDs []int64) ([]upstreamCandidate, error) {
	if len(providerIDs) == 0 {
		return nil, nil
	}
	rows, err := a.db.Query(`
		select p.id, p.name, p.type, p.enabled, p.weight, p.daily_limit, p.created_at, p.updated_at
		from upstream_providers p
		where p.enabled=1 and p.id in (`+placeholders(len(providerIDs))+`)
		order by p.weight desc, p.id asc`, intsAny(providerIDs)...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []upstreamCandidate{}
	for rows.Next() {
		var item upstreamCandidate
		var enabled int
		_ = rows.Scan(&item.ID, &item.Name, &item.Type, &enabled, &item.Weight, &item.DailyLimit, &item.CreatedAt, &item.UpdatedAt)
		item.Enabled = enabled == 1
		if item.Type == "tencent" {
			cfg, err := a.loadTencentConfig(strconv.FormatInt(item.ID, 10))
			if err == nil {
				item.TencentConfig = cfg
				item.Rules, _ = a.listRulesForProvider(strconv.FormatInt(item.ID, 10))
				item.Templates, _ = a.listTemplatesForProvider(strconv.FormatInt(item.ID, 10))
			}
		}
		if item.Type == "smtp" {
			cfg, err := a.loadSMTPConfig(strconv.FormatInt(item.ID, 10))
			if err == nil {
				item.SMTPConfig = cfg
			}
		}
		items = append(items, item)
	}
	return items, nil
}

func (a *App) allowedProviderIDsForDownstream(downstreamID int64) ([]int64, error) {
	var raw string
	var active int
	if err := a.db.QueryRow(`select allowed_provider_ids, active from smtp_accounts where id=?`, downstreamID).Scan(&raw, &active); err != nil {
		return nil, err
	}
	if active != 1 {
		return nil, fmt.Errorf("SMTP 下游未启用")
	}
	ids := jsonRawIntArray(raw)
	return ids, nil
}

func (a *App) nextRoundRobinIndex(downstreamID int64, size int) int {
	if size == 0 {
		return 0
	}
	var cursor int
	_ = a.db.QueryRow(`select cursor from provider_rr_state where downstream_id=?`, downstreamID).Scan(&cursor)
	next := (cursor + 1) % size
	_, _ = a.db.Exec(`insert into provider_rr_state(downstream_id,cursor,updated_at) values(?,?,?) on conflict(downstream_id) do update set cursor=excluded.cursor, updated_at=excluded.updated_at`, downstreamID, next, now())
	return cursor % size
}

func (a *App) sendThroughProvider(candidate upstreamCandidate, input MailInput) (int64, int64, string, string, map[string]string, error) {
	switch candidate.Type {
	case "tencent":
		result, ruleID, err := a.matchProviderRule(candidate, input)
		if err != nil {
			return 0, 0, "", "", nil, err
		}
		if result.TemplateID == 0 {
			return 0, 0, "", "", nil, fmt.Errorf("规则返回的 templateId 不能为 0")
		}
		if err := a.validateTemplateVars(uint64(candidate.ID), result.TemplateID, result.Variables); err != nil {
			return 0, 0, "", "", nil, err
		}
		providerMessageID, err := a.sendTencent(candidate, input, result)
		if err != nil {
			return 0, 0, "", "", nil, err
		}
		if err := a.bumpProviderUsage(candidate.ID); err != nil {
			return 0, 0, "", "", nil, err
		}
		return ruleID, int64(result.TemplateID), providerMessageID, "tencent", result.Variables, nil
	case "smtp":
		providerMessageID, err := a.sendSMTPUpstream(candidate, input)
		if err != nil {
			return 0, 0, "", "", nil, err
		}
		if err := a.bumpProviderUsage(candidate.ID); err != nil {
			return 0, 0, "", "", nil, err
		}
		return 0, 0, providerMessageID, "smtp", map[string]string{}, nil
	default:
		return 0, 0, "", "", nil, fmt.Errorf("未知上游类型")
	}
}

func (a *App) matchProviderRule(candidate upstreamCandidate, input MailInput) (RuleResult, int64, error) {
	for _, rule := range candidate.Rules {
		if !rule.Enabled {
			continue
		}
		result, matched, err := runRule(rule.Script, input)
		if err != nil || !matched {
			continue
		}
		return result, rule.ID, nil
	}
	return RuleResult{}, 0, fmt.Errorf("没有规则匹配这封邮件，请在该上游的规则列表中新增规则")
}

func (a *App) sendTencent(candidate upstreamCandidate, input MailInput, result RuleResult) (string, error) {
	client, err := a.tencentClient(candidate.TencentConfig)
	if err != nil {
		return "", err
	}
	cfg := candidate.TencentConfig
	from := strings.TrimSpace(cfg.FromAddress)
	if from == "" {
		return "", fmt.Errorf("腾讯云发信地址未配置")
	}
	triggerType, _ := strconv.ParseUint(firstNonEmpty(cfg.TriggerType, "1"), 10, 64)
	templateData, _ := json.Marshal(result.Variables)
	req := ses.NewSendEmailRequest()
	req.FromEmailAddress = &from
	subject := firstNonEmpty(result.Subject, input.Subject)
	req.Subject = &subject
	for _, rcpt := range input.To {
		v := rcpt
		req.Destination = append(req.Destination, &v)
	}
	replyTo := firstNonEmpty(cfg.ReplyTo, from)
	req.ReplyToAddresses = &replyTo
	req.HeaderFrom = &from
	req.TriggerType = &triggerType
	req.Template = &ses.Template{TemplateID: &result.TemplateID, TemplateData: ptrStr(string(templateData))}
	resp, err := client.SendEmail(req)
	if err != nil {
		return "", fmt.Errorf("%s", tencentErr(err))
	}
	if resp.Response != nil && resp.Response.MessageId != nil {
		return *resp.Response.MessageId, nil
	}
	return "", nil
}

func (a *App) sendSMTPUpstream(candidate upstreamCandidate, input MailInput) (string, error) {
	cfg := candidate.SMTPConfig
	if strings.TrimSpace(cfg.Host) == "" {
		return "", fmt.Errorf("SMTP 上游主机未配置")
	}
	hostPort := net.JoinHostPort(cfg.Host, strconv.Itoa(cfg.Port))
	from := firstNonEmpty(strings.TrimSpace(cfg.FromAddress), input.From)
	raw := buildMailMessage(from, input.To, input.Subject, input.Text, input.HTML, input.Headers)
	var auth smtp.Auth
	if strings.TrimSpace(cfg.Username) != "" {
		auth = smtp.PlainAuth("", cfg.Username, cfg.Password, cfg.Host)
	}
	if err := smtp.SendMail(hostPort, auth, from, input.To, []byte(raw)); err != nil {
		return "", err
	}
	return fmt.Sprintf("smtp-%d-%d", candidate.ID, time.Now().UnixNano()), nil
}

func buildMailMessage(from string, to []string, subject, textBody, htmlBody string, headers map[string]string) string {
	textBody = firstNonEmpty(textBody, "Nozomi Relay 转发邮件")
	lines := []string{
		"From: " + from,
		"To: " + strings.Join(to, ", "),
		"Subject: " + mimeQEncode(subject),
		"MIME-Version: 1.0",
	}
	for k, v := range headers {
		switch strings.ToLower(k) {
		case "from", "to", "subject", "mime-version", "content-type":
			continue
		}
		lines = append(lines, fmt.Sprintf("%s: %s", k, v))
	}
	if strings.TrimSpace(htmlBody) == "" {
		lines = append(lines, `Content-Type: text/plain; charset="UTF-8"`)
		return strings.Join(lines, "\r\n") + "\r\n\r\n" + textBody + "\r\n"
	}
	boundary := fmt.Sprintf("nozomi-%d", time.Now().UnixNano())
	lines = append(lines, `Content-Type: multipart/alternative; boundary="`+boundary+`"`)
	return strings.Join(lines, "\r\n") + "\r\n\r\n" +
		"--" + boundary + "\r\n" +
		`Content-Type: text/plain; charset="UTF-8"` + "\r\n\r\n" +
		textBody + "\r\n" +
		"--" + boundary + "\r\n" +
		`Content-Type: text/html; charset="UTF-8"` + "\r\n\r\n" +
		htmlBody + "\r\n" +
		"--" + boundary + "--\r\n"
}

func mimeQEncode(s string) string {
	if strings.TrimSpace(s) == "" {
		return ""
	}
	return mime.QEncoding.Encode("UTF-8", s)
}

func (a *App) bumpProviderUsage(providerID int64) error {
	usageDate := time.Now().Format("2006-01-02")
	_, err := a.db.Exec(`insert into provider_daily_usage(provider_id,usage_date,sent_count) values(?,?,1) on conflict(provider_id,usage_date) do update set sent_count=sent_count+1`, providerID, usageDate)
	return err
}

func (a *App) providerLimitReached(providerID int64, dailyLimit int) bool {
	if dailyLimit <= 0 {
		return false
	}
	var count int
	_ = a.db.QueryRow(`select sent_count from provider_daily_usage where provider_id=? and usage_date=?`, providerID, time.Now().Format("2006-01-02")).Scan(&count)
	return count >= dailyLimit
}

func placeholders(n int) string {
	if n <= 0 {
		return ""
	}
	out := make([]string, n)
	for i := range out {
		out[i] = "?"
	}
	return strings.Join(out, ",")
}

func intsAny(values []int64) []any {
	out := make([]any, len(values))
	for i, v := range values {
		out[i] = v
	}
	return out
}

func (a *App) failMessage(id int64, status, errText string) {
	_, _ = a.db.Exec(`update messages set status=?, error=?, updated_at=? where id=?`, status, errText, now(), id)
}

func (a *App) pickUpstreamCandidates(downstreamID int64, candidates []upstreamCandidate, mode string) []upstreamCandidate {
	if len(candidates) == 0 {
		return nil
	}
	filtered := make([]upstreamCandidate, 0, len(candidates))
	for _, candidate := range candidates {
		if !a.providerLimitReached(candidate.ID, candidate.DailyLimit) {
			filtered = append(filtered, candidate)
		}
	}
	if len(filtered) == 0 {
		return nil
	}
	if mode == "round_robin" {
		start := a.nextRoundRobinIndex(downstreamID, len(filtered))
		out := make([]upstreamCandidate, 0, len(filtered))
		for i := 0; i < len(filtered); i++ {
			out = append(out, filtered[(start+i)%len(filtered)])
		}
		return out
	}
	return filtered
}
