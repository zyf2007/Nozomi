import { Button, Col, Form, Input, InputNumber, Modal, Row, Select, Space, Switch, Tabs, Table, Tag } from 'antd'
import type { FormInstance } from 'antd'
import { CloudSyncOutlined, DeleteOutlined, PlusOutlined } from '@ant-design/icons'
import { requestJson } from '../api/client'
import type { ProviderDetail, ProviderRule } from '../types'

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

  return (
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
          }
          if (provider) {
              await requestJson(`/api/providers/${provider.id}`, {
                method: 'PUT',
                body: JSON.stringify({
                  name: payload.name,
                  enabled: payload.enabled,
                  weight: payload.weight,
                  daily_limit: payload.daily_limit,
                }),
              })
            if (payload.type === 'tencent') {
              await requestJson(`/api/providers/${provider.id}/tencent`, { method: 'PUT', body: JSON.stringify(payload.tencent_config) })
            } else {
              await requestJson(`/api/providers/${provider.id}/smtp`, { method: 'PUT', body: JSON.stringify(payload.smtp_config) })
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
              children: (
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
                    <Col xs={12} md={16}>
                      <Form.Item name="daily_limit" label="每日最大发件量">
                        <InputNumber min={0} className="full" />
                      </Form.Item>
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
                  ) : (
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
                  )}
                </>
              ),
            },
            {
              key: 'rules',
              label: '规则配置',
              disabled: providerType !== 'tencent',
              children: providerType === 'tencent' ? (
                <>
                  <Space className="modal-toolbar" wrap>
                    <Button icon={<PlusOutlined />} type="primary" onClick={() => onEditRule(null)}>
                      新增规则
                    </Button>
                  </Space>
                  <Table
                    rowKey="id"
                    dataSource={provider?.rules || []}
                    size="small"
                    pagination={false}
                    columns={[
                      { title: '优先级', dataIndex: 'priority', width: 90 },
                      { title: '名称', dataIndex: 'name' },
                      {
                        title: '启用',
                        dataIndex: 'enabled',
                        width: 90,
                        render: (v) => <Switch checked={v} disabled />,
                      },
                      { title: '更新时间', dataIndex: 'updated_at', width: 220 },
                      {
                        title: '操作',
                        width: 180,
                        render: (_, record: ProviderRule) => (
                          <Space>
                            <Button size="small" onClick={() => onEditRule(record)}>
                              编辑
                            </Button>
                            <Button size="small" danger icon={<DeleteOutlined />} onClick={() => onDeleteRule(record)}>
                              删除
                            </Button>
                          </Space>
                        ),
                      },
                    ]}
                  />
                </>
              ) : null,
            },
            {
              key: 'templates',
              label: '模板列表',
              disabled: providerType !== 'tencent',
              children: providerType === 'tencent' ? (
                <>
                  <Space className="modal-toolbar" wrap>
                    <Button icon={<CloudSyncOutlined />} onClick={onSyncTemplates}>
                      从腾讯云同步
                    </Button>
                  </Space>
                  <Table
                    rowKey="id"
                    dataSource={provider?.templates || []}
                    size="small"
                    pagination={false}
                    columns={[
                      { title: 'ID', dataIndex: 'id', width: 120 },
                      { title: '名称', dataIndex: 'name' },
                      { title: '状态', dataIndex: 'status', width: 120 },
                      {
                        title: '变量',
                        dataIndex: 'variables',
                        render: (vars: string[]) => (
                          <Space wrap>
                            {vars.map((v) => (
                              <Tag key={v}>{v}</Tag>
                            ))}
                          </Space>
                        ),
                      },
                      { title: '更新时间', dataIndex: 'updated_at', width: 220 },
                    ]}
                  />
                </>
              ) : null,
            },
          ]}
        />
      </Form>
    </Modal>
  )
}
