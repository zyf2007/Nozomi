import { Col, Form, Input, InputNumber, Row, Select, Switch } from 'antd'
import type { ProviderDetail } from '../../types'

type Props = {
  provider: ProviderDetail | null
  providerType: string
  quotaTimezone: string
}

const quotaTimezoneOptions = [
  { label: 'UTC', value: 'UTC' },
  { label: '北京时间 (UTC+8)', value: 'Asia/Shanghai' },
  { label: '东京时间 (UTC+9)', value: 'Asia/Tokyo' },
  { label: '新加坡时间 (UTC+8)', value: 'Asia/Singapore' },
  { label: '美国东部时间', value: 'America/New_York' },
  { label: '美国太平洋时间', value: 'America/Los_Angeles' },
  { label: '欧洲中部时间', value: 'Europe/Berlin' },
]

function formatQuotaResetTime(timezone: string) {
  try {
    const now = new Date()
    const parts = new Intl.DateTimeFormat('en-US', {
      timeZone: timezone || 'Asia/Shanghai',
      year: 'numeric',
      month: '2-digit',
      day: '2-digit',
    }).formatToParts(now)
    const values = Object.fromEntries(parts.map((part) => [part.type, part.value]))
    const resetDate = `${values.year}-${values.month}-${values.day}`
    const firstPass = new Date(`${resetDate}T00:00:00${timezoneOffsetForDate(timezone, now)}`)
    const resetAt = new Date(`${resetDate}T00:00:00${timezoneOffsetForDate(timezone, firstPass)}`)
    return new Intl.DateTimeFormat('zh-CN', {
      timeZone: Intl.DateTimeFormat().resolvedOptions().timeZone,
      hour: '2-digit',
      minute: '2-digit',
      hour12: false,
      timeZoneName: 'short',
    }).format(resetAt)
  } catch {
    return '00:00'
  }
}

function timezoneOffsetForDate(timezone: string, date: Date) {
  const parts = new Intl.DateTimeFormat('en-US', {
    timeZone: timezone || 'Asia/Shanghai',
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
    second: '2-digit',
    hour12: false,
  }).formatToParts(date)
  const values = Object.fromEntries(parts.map((part) => [part.type, part.value]))
  const localTime = Date.UTC(Number(values.year), Number(values.month) - 1, Number(values.day), Number(values.hour), Number(values.minute), Number(values.second))
  const offsetMinutes = Math.round((localTime - date.getTime()) / 60000)
  const sign = offsetMinutes >= 0 ? '+' : '-'
  const abs = Math.abs(offsetMinutes)
  return `${sign}${String(Math.floor(abs / 60)).padStart(2, '0')}:${String(abs % 60).padStart(2, '0')}`
}

export function ProviderBasicTab({ provider, providerType, quotaTimezone }: Props) {
  return (
    <>
      <Row gutter={16}>
        <Col xs={24} md={12}>
          <Form.Item name="name" label="名称" rules={[{ required: true }]}>
            <Input placeholder="例如：默认腾讯云 SES" />
          </Form.Item>
        </Col>
        <Col xs={12} md={6}>
          <Form.Item name="type" label="类型" rules={[{ required: true }]}>
            <Select
              options={[
                { label: '腾讯云 SES', value: 'tencent' },
                { label: 'SMTP', value: 'smtp' },
                { label: 'Resend', value: 'resend' },
                { label: 'Brevo', value: 'brevo' },
              ]}
              disabled={!!provider}
            />
          </Form.Item>
        </Col>
        <Col xs={12} md={6}>
          <Form.Item name="enabled" label="启用" valuePropName="checked">
            <Switch />
          </Form.Item>
        </Col>
        <Col xs={12} md={8}>
          <Form.Item name="weight" label="权重">
            <InputNumber min={0} className="full" />
          </Form.Item>
        </Col>
        <Col xs={12} md={8}>
          <Form.Item name="daily_limit" label="每日最大发件量">
            <InputNumber min={0} className="full" />
          </Form.Item>
        </Col>
        <Col xs={24} md={8}>
          <Form.Item name="quota_timezone" label="刷新时区">
            <Select options={quotaTimezoneOptions} showSearch optionFilterProp="label" />
          </Form.Item>
        </Col>
        <Col xs={24}>
          <div className="quota-reset-hint">每日额度按所选时区 00:00 刷新，当前本地时间约 {formatQuotaResetTime(quotaTimezone)} 刷新。</div>
        </Col>
      </Row>

      {providerType === 'tencent' ? (
        <Row gutter={16}>
          <Col xs={24} md={12}>
            <Form.Item name="secret_id" label="SecretId" rules={[{ required: true, message: '请输入 SecretId' }]}>
              <Input />
            </Form.Item>
          </Col>
          <Col xs={24} md={12}>
            <Form.Item name="secret_key" label="SecretKey" rules={[{ required: true, message: '请输入 SecretKey' }]}>
              <Input.Password />
            </Form.Item>
          </Col>
          <Col xs={24} md={8}>
            <Form.Item name="region" label="Region">
              <Input placeholder="ap-guangzhou" />
            </Form.Item>
          </Col>
          <Col xs={24} md={8}>
            <Form.Item name="from_address" label="发信地址">
              <Input placeholder="Nozomi <noreply@example.com>" />
            </Form.Item>
          </Col>
          <Col xs={24} md={8}>
            <Form.Item name="trigger_type" label="TriggerType">
              <Input placeholder="1" />
            </Form.Item>
          </Col>
          <Col xs={24}>
            <Form.Item name="reply_to" label="Reply-To">
              <Input />
            </Form.Item>
          </Col>
        </Row>
      ) : providerType === 'smtp' ? (
        <Row gutter={16}>
          <Col xs={24} md={8}>
            <Form.Item name="host" label="SMTP Host" rules={[{ required: true, message: '请输入 Host' }]}>
              <Input />
            </Form.Item>
          </Col>
          <Col xs={24} md={4}>
            <Form.Item name="port" label="Port" rules={[{ required: true, message: '请输入端口' }]}>
              <InputNumber min={1} max={65535} className="full" />
            </Form.Item>
          </Col>
          <Col xs={24} md={6}>
            <Form.Item name="username" label="用户名" rules={[{ required: true, message: '请输入用户名' }]}>
              <Input />
            </Form.Item>
          </Col>
          <Col xs={24} md={6}>
            <Form.Item name="password" label="密码">
              <Input.Password />
            </Form.Item>
          </Col>
          <Col xs={24}>
            <Form.Item name="from_address" label="发信地址">
              <Input />
            </Form.Item>
          </Col>
          <Col xs={24}>
            <Form.Item name="reply_to" label="Reply-To">
              <Input />
            </Form.Item>
          </Col>
        </Row>
      ) : providerType === 'resend' ? (
        <Row gutter={16}>
          <Col xs={24} md={12}>
            <Form.Item name="api_key" label="API Key" rules={[{ required: true, message: '请输入 API Key' }]}>
              <Input.Password />
            </Form.Item>
          </Col>
          <Col xs={24} md={12}>
            <Form.Item name="from_address" label="发信地址" rules={[{ required: true, message: '请输入发信地址' }]}>
              <Input placeholder="Nozomi <noreply@example.com>" />
            </Form.Item>
          </Col>
          <Col xs={24}>
            <Form.Item name="reply_to" label="Reply-To">
              <Input />
            </Form.Item>
          </Col>
        </Row>
      ) : (
        <Row gutter={16}>
          <Col xs={24} md={12}>
            <Form.Item name="api_key" label="API Key" rules={[{ required: true, message: '请输入 API Key' }]}>
              <Input.Password />
            </Form.Item>
          </Col>
          <Col xs={24} md={12}>
            <Form.Item name="from_name" label="发件人名称">
              <Input placeholder="Nozomi" />
            </Form.Item>
          </Col>
          <Col xs={24} md={12}>
            <Form.Item name="from_address" label="发信地址" rules={[{ required: true, message: '请输入发信地址' }]}>
              <Input placeholder="noreply@example.com" />
            </Form.Item>
          </Col>
          <Col xs={24}>
            <Form.Item name="reply_to" label="Reply-To">
              <Input />
            </Form.Item>
          </Col>
        </Row>
      )}
    </>
  )
}
