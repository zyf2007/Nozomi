export type Session = { authenticated: boolean; username: string }

export type Stats = {
  total: number
  sent: number
  delivered: number
  bounce: number
  failed: number
  delivery_rate: number
  bounce_rate: number
}

export type ProviderType = 'tencent' | 'smtp'
export type ProviderDispatchMode = 'queue' | 'round_robin'

export type TencentConfig = {
  secret_id: string
  secret_key: string
  region: string
  from_address: string
  reply_to: string
  trigger_type: string
}

export type SMTPConfig = {
  host: string
  port: number
  username: string
  password: string
  from_address: string
  reply_to: string
}

export type Provider = {
  id: number
  name: string
  type: ProviderType
  enabled: boolean
  weight: number
  daily_limit: number
  created_at: string
  updated_at: string
  today_sent: number
}

export type ProviderRule = {
  id: number
  provider_id: number
  name: string
  enabled: boolean
  priority: number
  script: string
  created_at: string
  updated_at: string
}

export type ProviderTemplate = {
  id: number
  provider_id: number
  name: string
  status: number
  variables: string[]
  html: string
  text: string
  updated_at: string
}

export type Rule = {
  id: number
  name: string
  enabled: boolean
  priority: number
  script: string
  updated_at: string
}

export type Template = {
  id: number
  name: string
  status: number
  variables: string[]
  updated_at: string
}

export type ProviderDetail = Provider & {
  tencent_config: TencentConfig
  smtp_config: SMTPConfig
  rules: ProviderRule[]
  templates: ProviderTemplate[]
}

export type SMTPAccount = {
  id: number
  username: string
  active: boolean
  allowed_provider_ids: number[]
  created_at: string
  updated_at: string
}

export type RuleTestResponse = {
  matched: boolean
  result?: Record<string, unknown>
  validation_error?: string
  error?: string
}

export type RelayMessage = {
  id: number
  from: string
  to: string
  subject: string
  downstream_account_id: number | null
  provider_id: number | null
  provider_type: string
  rule_id: number | null
  template_id: number | null
  template_data: Record<string, string>
  status: string
  error: string
  provider_message_id: string
  callback_event: string
  callback_reason: string
  bounce_type: string
  created_at: string
  updated_at: string
}

export type TemplateOption = { label: string; value: number }
