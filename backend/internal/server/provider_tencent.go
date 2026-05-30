package server

import (
	"encoding/json"
	"fmt"
	"encoding/base64"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
	ses "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/ses/v20201002"
)

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

func (a *App) loadTencentConfig(providerID string) (TencentConfig, error) {
	var cfg TencentConfig
	err := a.db.QueryRow(`select secret_id, secret_key, region, from_address, reply_to, trigger_type from provider_tencent_config where provider_id=?`, providerID).Scan(&cfg.SecretID, &cfg.SecretKey, &cfg.Region, &cfg.FromAddress, &cfg.ReplyTo, &cfg.TriggerType)
	return cfg, err
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

func (a *App) sendTencent(candidate upstreamCandidate, input MailInput, result RuleResult) (string, string, error) {
	client, err := a.tencentClient(candidate.TencentConfig)
	if err != nil {
		return "", "", err
	}
	cfg := candidate.TencentConfig
	from := strings.TrimSpace(cfg.FromAddress)
	if from == "" {
		return "", "", fmt.Errorf("腾讯云发信地址未配置")
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
	if len(input.Attachments) > 0 {
		req.Attachments = make([]*ses.Attachment, 0, len(input.Attachments))
		for _, attachment := range input.Attachments {
			content, err := base64.StdEncoding.DecodeString(strings.TrimSpace(attachment.ContentBase64))
			if err != nil {
				return "", "", fmt.Errorf("附件 %s 编码无效", attachment.Filename)
			}
			filename := firstNonEmpty(strings.TrimSpace(attachment.Filename), "attachment")
			encoded := base64.StdEncoding.EncodeToString(content)
			fileName := filename
			req.Attachments = append(req.Attachments, &ses.Attachment{
				FileName: &fileName,
				Content:  &encoded,
			})
		}
	}
	sentRaw := jsonString(map[string]any{
		"from":          from,
		"to":            input.To,
		"subject":       subject,
		"reply_to":      replyTo,
		"trigger_type":  triggerType,
		"template_id":   result.TemplateID,
		"template_data": result.Variables,
		"provider_type": "tencent",
		"render_source": "腾讯云 SES 模板邮件，原文由腾讯云按模板渲染生成",
		"attachments":   input.Attachments,
	})
	resp, err := client.SendEmail(req)
	if err != nil {
		return "", sentRaw, fmt.Errorf("%s", tencentErr(err))
	}
	if resp.Response != nil && resp.Response.MessageId != nil {
		return *resp.Response.MessageId, sentRaw, nil
	}
	return "", sentRaw, nil
}
