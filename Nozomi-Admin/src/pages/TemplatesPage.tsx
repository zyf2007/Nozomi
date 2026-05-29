import { Button, Card, Space, Table, Tag, Tooltip, message } from 'antd'
import { CloudSyncOutlined, CopyOutlined } from '@ant-design/icons'
import type { Template } from '../types'
import { requestJson } from '../api/client'

type Props = {
  templates: Template[]
  onSynced: () => void
}

export function TemplatesPage({ templates, onSynced }: Props) {
  return (
    <Card
      title="腾讯云模板缓存"
      extra={
        <Button
          icon={<CloudSyncOutlined />}
          onClick={async () => {
            await requestJson('/api/templates/sync', { method: 'POST' })
            message.success('模板已同步')
            onSynced()
          }}
        >
          从腾讯云同步
        </Button>
      }
    >
      <Table
        rowKey="id"
        dataSource={templates}
        columns={[
          {
            title: 'ID',
            dataIndex: 'id',
            width: 160,
            render: (id: number) => (
              <Space size={8}>
                <span>{id}</span>
                <Tooltip title="复制 ID">
                  <Button
                    type="text"
                    size="small"
                    icon={<CopyOutlined />}
                    onClick={async () => {
                      await navigator.clipboard.writeText(String(id))
                      message.success('ID 已复制')
                    }}
                  />
                </Tooltip>
              </Space>
            ),
          },
          { title: '名称', dataIndex: 'name' },
          { title: '状态', dataIndex: 'status', width: 120 },
          {
            title: '变量',
            dataIndex: 'variables',
            render: (vars: string[]) => vars.map((v) => <Tag key={v}>{v}</Tag>),
          },
          { title: '更新时间', dataIndex: 'updated_at', width: 220 },
        ]}
      />
    </Card>
  )
}
