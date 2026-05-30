import { useState } from 'react'
import { Form, Modal, Tabs } from 'antd'
import type { FormInstance } from 'antd'
import { requestJson } from '../api/client'
import type { ProviderDetail, ProviderRule, ProviderTemplate } from '../types'
import { ProviderBasicTab } from './provider/ProviderBasicTab'
import { ProviderRulesTab } from './provider/ProviderRulesTab'
import { ProviderTemplateDetailModal } from './provider/ProviderTemplateDetailModal'
import { ProviderTemplatesTab } from './provider/ProviderTemplatesTab'

type Props = {
  provider: ProviderDetail | null
  form: FormInstance
  open: boolean
  onClose: () => void
  onSaved: () => void
  onSyncTemplates: () => Promise<void>
  onEditRule: (rule: ProviderRule | null) => void
  onDeleteRule: (rule: ProviderRule) => Promise<void>
}

export function ProviderModal({ provider, form, open, onClose, onSaved, onSyncTemplates, onEditRule, onDeleteRule }: Props) {
  const providerType = Form.useWatch('type', form) || provider?.type || 'tencent'
  const quotaTimezone = Form.useWatch('quota_timezone', form) || 'Asia/Shanghai'
  const [viewingTemplate, setViewingTemplate] = useState<ProviderTemplate | null>(null)

  return (
    <>
      <Modal
        title={provider ? `编辑上游 · ${provider.name}` : '新增上游'}
        open={open}
        width={1180}
        onCancel={onClose}
        onOk={() => form.submit()}
        okText="保存"
        cancelText="关闭"
        destroyOnClose={false}
      >
        <Form
          form={form}
          layout="vertical"
          onFinish={async (values) => {
            const payload = {
              name: values.name,
              type: values.type,
              enabled: values.enabled ?? true,
              weight: values.weight ?? 100,
              daily_limit: values.daily_limit ?? 0,
              quota_timezone: values.quota_timezone || 'Asia/Shanghai',
              tencent_config: {
                secret_id: values.secret_id || '',
                secret_key: values.secret_key || '',
                region: values.region || 'ap-guangzhou',
                from_address: values.from_address || '',
                reply_to: values.reply_to || '',
                trigger_type: values.trigger_type || '1',
              },
              smtp_config: {
                host: values.host || '',
                port: values.port || 25,
                username: values.username || '',
                password: values.password || '',
                from_address: values.from_address || '',
                reply_to: values.reply_to || '',
              },
              resend_config: {
                api_key: values.api_key || '',
                from_address: values.from_address || '',
                reply_to: values.reply_to || '',
              },
              brevo_config: {
                api_key: values.api_key || '',
                from_address: values.from_address || '',
                from_name: values.from_name || 'Nozomi',
                reply_to: values.reply_to || '',
              },
            }
            if (provider) {
              await requestJson(`/api/providers/${provider.id}`, {
                method: 'PUT',
                body: JSON.stringify({
                  name: payload.name,
                  enabled: payload.enabled,
                  weight: payload.weight,
                  daily_limit: payload.daily_limit,
                  quota_timezone: payload.quota_timezone,
                }),
              })
              if (payload.type === 'tencent') {
                await requestJson(`/api/providers/${provider.id}/tencent`, { method: 'PUT', body: JSON.stringify(payload.tencent_config) })
              } else if (payload.type === 'smtp') {
                await requestJson(`/api/providers/${provider.id}/smtp`, { method: 'PUT', body: JSON.stringify(payload.smtp_config) })
              } else if (payload.type === 'resend') {
                await requestJson(`/api/providers/${provider.id}/resend`, { method: 'PUT', body: JSON.stringify(payload.resend_config) })
              } else {
                await requestJson(`/api/providers/${provider.id}/brevo`, { method: 'PUT', body: JSON.stringify(payload.brevo_config) })
              }
            } else {
              await requestJson('/api/providers', { method: 'POST', body: JSON.stringify(payload) })
            }
            onSaved()
          }}
        >
          <Tabs
            items={[
              {
                key: 'basic',
                label: '基础配置',
                children: <ProviderBasicTab provider={provider} providerType={providerType} quotaTimezone={quotaTimezone} />,
              },
              {
                key: 'rules',
                label: '规则配置',
                disabled: providerType !== 'tencent',
                children: providerType === 'tencent' ? (
                  <ProviderRulesTab rules={provider?.rules || []} onEditRule={onEditRule} onDeleteRule={onDeleteRule} />
                ) : null,
              },
              {
                key: 'templates',
                label: '模板列表',
                disabled: providerType !== 'tencent',
                children: providerType === 'tencent' ? (
                  <ProviderTemplatesTab templates={provider?.templates || []} onSyncTemplates={onSyncTemplates} onViewTemplate={setViewingTemplate} />
                ) : null,
              },
            ]}
          />
        </Form>
      </Modal>
      <ProviderTemplateDetailModal template={viewingTemplate} onClose={() => setViewingTemplate(null)} />
    </>
  )
}
