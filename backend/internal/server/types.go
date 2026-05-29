package server

import "database/sql"

type Settings struct {
	HTTPAddr      string
	SMTPAddr      string
	DataDir       string
	DBPath        string
	AdminUsername string
	AdminPassword string
	SessionSecret string
}

type App struct {
	db       *sql.DB
	settings Settings
}

type MailInput struct {
	From    string            `json:"from"`
	To      []string          `json:"to"`
	Subject string            `json:"subject"`
	Text    string            `json:"text"`
	HTML    string            `json:"html"`
	Raw     string            `json:"raw"`
	Headers map[string]string `json:"headers"`
}

type RuleResult struct {
	TemplateID uint64            `json:"templateId"`
	Subject    string            `json:"subject"`
	Variables  map[string]string `json:"variables"`
}

type Provider struct {
	ID         int64  `json:"id"`
	Name       string `json:"name"`
	Type       string `json:"type"`
	Enabled    bool   `json:"enabled"`
	Weight     int    `json:"weight"`
	DailyLimit int    `json:"daily_limit"`
	CreatedAt  string `json:"created_at"`
	UpdatedAt  string `json:"updated_at"`
	TodaySent  int64  `json:"today_sent"`
}

type ProviderDetail struct {
	Provider
	TencentConfig TencentConfig      `json:"tencent_config"`
	SMTPConfig    SMTPConfig         `json:"smtp_config"`
	ResendConfig  ResendConfig       `json:"resend_config"`
	BrevoConfig   BrevoConfig        `json:"brevo_config"`
	Rules         []ProviderRule     `json:"rules"`
	Templates     []ProviderTemplate `json:"templates"`
}

type TencentConfig struct {
	SecretID    string `json:"secret_id"`
	SecretKey   string `json:"secret_key"`
	Region      string `json:"region"`
	FromAddress string `json:"from_address"`
	ReplyTo     string `json:"reply_to"`
	TriggerType string `json:"trigger_type"`
}

type SMTPConfig struct {
	Host        string `json:"host"`
	Port        int    `json:"port"`
	Username    string `json:"username"`
	Password    string `json:"password"`
	FromAddress string `json:"from_address"`
	ReplyTo     string `json:"reply_to"`
}

type ResendConfig struct {
	APIKey      string `json:"api_key"`
	FromAddress string `json:"from_address"`
	ReplyTo     string `json:"reply_to"`
}

type BrevoConfig struct {
	APIKey      string `json:"api_key"`
	FromAddress string `json:"from_address"`
	FromName    string `json:"from_name"`
	ReplyTo     string `json:"reply_to"`
}

type ProviderRule struct {
	ID         int64  `json:"id"`
	ProviderID int64  `json:"provider_id"`
	Name       string `json:"name"`
	Enabled    bool   `json:"enabled"`
	Priority   int    `json:"priority"`
	Script     string `json:"script"`
	CreatedAt  string `json:"created_at"`
	UpdatedAt  string `json:"updated_at"`
}

type ProviderTemplate struct {
	ID         int64    `json:"id"`
	ProviderID int64    `json:"provider_id"`
	Name       string   `json:"name"`
	Status     int      `json:"status"`
	Variables  []string `json:"variables"`
	HTML       string   `json:"html"`
	Text       string   `json:"text"`
	UpdatedAt  string   `json:"updated_at"`
}

type SMTPAccount struct {
	ID                 int64   `json:"id"`
	Username           string  `json:"username"`
	Active             bool    `json:"active"`
	AllowedProviderIDs []int64 `json:"allowed_provider_ids"`
	CreatedAt          string  `json:"created_at"`
	UpdatedAt          string  `json:"updated_at"`
}

type RuleTestResponse struct {
	Matched         bool        `json:"matched"`
	Result          *RuleResult `json:"result,omitempty"`
	ValidationError string      `json:"validation_error,omitempty"`
	Error           string      `json:"error,omitempty"`
}

type RelayMessage struct {
	ID                int64             `json:"id"`
	From              string            `json:"from"`
	To                string            `json:"to"`
	Subject           string            `json:"subject"`
	DownstreamID      int64             `json:"downstream_id"`
	ProviderID        int64             `json:"provider_id"`
	ProviderType      string            `json:"provider_type"`
	RuleID            *int64            `json:"rule_id"`
	TemplateID        *int64            `json:"template_id"`
	TemplateData      map[string]string `json:"template_data"`
	Status            string            `json:"status"`
	Error             string            `json:"error"`
	ProviderMessageID string            `json:"provider_message_id"`
	CallbackEvent     string            `json:"callback_event"`
	CallbackReason    string            `json:"callback_reason"`
	BounceType        string            `json:"bounce_type"`
	CreatedAt         string            `json:"created_at"`
	UpdatedAt         string            `json:"updated_at"`
}

type Stats struct {
	Total        int64   `json:"total"`
	Sent         int64   `json:"sent"`
	Delivered    int64   `json:"delivered"`
	Bounce       int64   `json:"bounce"`
	Failed       int64   `json:"failed"`
	DeliveryRate float64 `json:"delivery_rate"`
	BounceRate   float64 `json:"bounce_rate"`
}

type ruleBody struct {
	ProviderID int64  `json:"provider_id"`
	Name       string `json:"name"`
	Enabled    bool   `json:"enabled"`
	Priority   int    `json:"priority"`
	Script     string `json:"script"`
}
