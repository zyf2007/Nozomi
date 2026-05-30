import { Button, Space, Table, Tag, Tooltip, message } from 'antd'
import { CloudSyncOutlined, CopyOutlined } from '@ant-design/icons'
import type { ProviderTemplate } from '../../types'

type Props = {
  templates: ProviderTemplate[]
  onSyncTemplates: () => Promise<void>
  onViewTemplate: (template: ProviderTemplate) => void
}

export function ProviderTemplatesTab({ templates, onSyncTemplates, onViewTemplate }: Props) {
  return (
    <>
      <Space className="modal-toolbar" wrap>
        <Button icon={<CloudSyncOutlined />} onClick={onSyncTemplates}>
          从腾讯云同步
        </Button>
      </Space>
      <Table
        rowKey="id"
        dataSource={templates}
        size="small"
        pagination={false}
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
            render: (vars: string[]) => (
              <Space wrap>
                {vars.map((item) => (
                  <Tag key={item}>{item}</Tag>
                ))}
              </Space>
            ),
          },
          { title: '更新时间', dataIndex: 'updated_at', width: 220 },
          {
            title: '操作',
            width: 100,
            render: (_, record: ProviderTemplate) => (
              <Button size="small" onClick={() => onViewTemplate(record)}>
                详情
              </Button>
            ),
          },
        ]}
      />
    </>
  )
}
