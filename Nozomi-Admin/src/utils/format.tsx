import { Tag } from 'antd'

export function statusTag(status: string) {
  const color: Record<string, string> = {
    sent: 'processing',
    delivered: 'success',
    bounce: 'error',
    hard_bounce: 'error',
    failed: 'error',
    no_rule: 'warning',
    bad_mapping: 'warning',
    dropped: 'error',
    received: 'default',
  }
  return <Tag color={color[status] || 'default'}>{status || '-'}</Tag>
}

export function splitRecipients(value: unknown) {
  return String(value || '')
    .split(/[,;\s]+/)
    .map((item) => item.trim())
    .filter(Boolean)
}
