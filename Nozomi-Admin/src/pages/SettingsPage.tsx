import { Button, Card, InputNumber, Select, Space, Switch, Table, Tag } from 'antd'
import { PlusOutlined, SaveOutlined } from '@ant-design/icons'
import type { Provider, ProviderDispatchMode } from '../types'

type Props = {
  providers: Provider[]
  dispatchMode: ProviderDispatchMode
  saving: boolean
  onCreate: () => void
  onEdit: (provider: Provider) => void
  onDelete: (provider: Provider) => void
  onWeightsChange: (providers: Provider[]) => void
  onDispatchModeChange: (mode: ProviderDispatchMode) => void
  onSave: () => void
}

export function SettingsPage({
  providers,
  dispatchMode,
  saving,
  onCreate,
  onEdit,
  onDelete,
  onWeightsChange,
  onDispatchModeChange,
  onSave,
}: Props) {
  const sortedProviders = [...providers].sort((a, b) => b.weight - a.weight || a.id - b.id)
  const typeLabelMap: Record<string, string> = {
    tencent: '腾讯云 SES',
    smtp: 'SMTP',
    resend: 'Resend',
    brevo: 'Brevo',
  }

  return (
    <Card
      title="上游配置"
      extra={
        <Space wrap>
          <span>调度模式</span>
          <Select
            value={dispatchMode}
            style={{ width: 180 }}
            options={[
              { label: '队列模式', value: 'queue' },
              { label: '轮询模式', value: 'round_robin' },
            ]}
            onChange={(value) => onDispatchModeChange(value)}
          />
          <Button icon={<SaveOutlined />} type="primary" loading={saving} onClick={onSave}>
            保存配置
          </Button>
          <Button icon={<PlusOutlined />} onClick={onCreate}>
            新增上游
          </Button>
        </Space>
      }
    >
      <Table
        rowKey="id"
        dataSource={sortedProviders}
        pagination={false}
        columns={[
          { title: '名称', dataIndex: 'name' },
          {
            title: '类型',
            dataIndex: 'type',
            width: 110,
            render: (v) => <Tag>{typeLabelMap[v] || v}</Tag>,
          },
          {
            title: '权重',
            dataIndex: 'weight',
            width: 140,
            render: (_, record) => (
              <InputNumber
                min={0}
                value={record.weight}
                className="full"
                onChange={(value) => {
                  const next = providers.map((item) => (item.id === record.id ? { ...item, weight: Number(value || 0) } : item))
                  onWeightsChange(next)
                }}
              />
            ),
          },
          { title: '每日上限', dataIndex: 'daily_limit', width: 120 },
          { title: '今日已发', dataIndex: 'today_sent', width: 100 },
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
            render: (_, record) => (
              <Space>
                <Button size="small" onClick={() => onEdit(record)}>
                  编辑
                </Button>
                <Button
                  size="small"
                  danger
                  onClick={() => {
                    if (confirm(`确认删除上游「${record.name}」？`)) onDelete(record)
                  }}
                >
                  删除
                </Button>
              </Space>
            ),
          },
        ]}
      />
    </Card>
  )
}
