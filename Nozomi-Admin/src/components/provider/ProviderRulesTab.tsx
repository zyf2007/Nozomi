import { Button, Space, Switch, Table } from 'antd'
import { DeleteOutlined, PlusOutlined } from '@ant-design/icons'
import type { ProviderRule } from '../../types'

type Props = {
  rules: ProviderRule[]
  onEditRule: (rule: ProviderRule | null) => void
  onDeleteRule: (rule: ProviderRule) => Promise<void>
}

export function ProviderRulesTab({ rules, onEditRule, onDeleteRule }: Props) {
  return (
    <>
      <Space className="modal-toolbar" wrap>
        <Button icon={<PlusOutlined />} type="primary" onClick={() => onEditRule(null)}>
          新增规则
        </Button>
      </Space>
      <Table
        rowKey="id"
        dataSource={rules}
        size="small"
        pagination={false}
        columns={[
          { title: '优先级', dataIndex: 'priority', width: 90 },
          { title: '名称', dataIndex: 'name' },
          {
            title: '启用',
            dataIndex: 'enabled',
            width: 90,
            render: (value) => <Switch checked={value} disabled />,
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
  )
}
