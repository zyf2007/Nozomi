import { Form, Input, Modal, Select, Switch } from 'antd'
import type { FormInstance } from 'antd'
import { requestJson } from '../api/client'
import type { Provider, SMTPAccount } from '../types'

type Props = {
  form: FormInstance
  open: boolean
  account: SMTPAccount | null
  providers: Provider[]
  onClose: () => void
  onRefresh: () => void
}

export function SMTPAccountModal({ form, open, account, providers, onClose, onRefresh }: Props) {
  const providerOptions = providers.map((item) => ({ label: item.name, value: item.id }))

  return (
    <Modal
      title={account ? `编辑 SMTP 下游 · ${account.username}` : '新增 SMTP 下游'}
      open={open}
      onCancel={onClose}
      onOk={() => form.submit()}
      okText="保存"
      cancelText="关闭"
    >
      <Form
        form={form}
        layout="vertical"
        onFinish={async (values) => {
          const payload = {
            username: values.username,
            password: values.password || '',
            active: values.active ?? true,
            allowed_provider_ids: values.allowed_provider_ids || [],
          }
          if (account) {
            await requestJson(`/api/smtp-accounts/${account.id}`, { method: 'PUT', body: JSON.stringify(payload) })
          } else {
            await requestJson('/api/smtp-accounts', { method: 'POST', body: JSON.stringify(payload) })
          }
          onClose()
          onRefresh()
        }}
      >
        <Form.Item name="username" label="用户名" rules={[{ required: true }]}>
          <Input />
        </Form.Item>
        <Form.Item name="password" label="密码" rules={account ? [] : [{ required: true }]}>
          <Input.Password placeholder={account ? '留空表示不修改' : ''} />
        </Form.Item>
        <Form.Item name="allowed_provider_ids" label="可用上游" rules={[{ required: true, message: '至少选择一个上游' }]}>
          <Select mode="multiple" options={providerOptions} placeholder="选择该下游可用的上游" />
        </Form.Item>
        <Form.Item name="active" label="启用" valuePropName="checked">
          <Switch />
        </Form.Item>
      </Form>
    </Modal>
  )
}
