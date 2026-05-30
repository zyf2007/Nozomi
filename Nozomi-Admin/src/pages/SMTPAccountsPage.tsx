import { Button, Card, Space, Switch, Table, Tag } from 'antd'
import { DeleteOutlined, MailOutlined, SendOutlined } from '@ant-design/icons'
import type { Provider, SMTPAccount } from '../types'

type Props = {
  accounts: SMTPAccount[]
  providers: Provider[]
  onCreate: () => void
  onEdit: (account: SMTPAccount) => void
  onDelete: (account: SMTPAccount) => void
  onTest: (account: SMTPAccount) => void
}

export function SMTPAccountsPage({ accounts, providers, onCreate, onEdit, onDelete, onTest }: Props) {
  const providerNameMap = new Map(providers.map((item) => [item.id, item.name]))

  return (
    <Card
      title="SMTP 下游"
      extra={
        <Button
          type="primary"
          icon={<MailOutlined />}
          onClick={onCreate}
        >
          新增下游
        </Button>
      }
    >
      <Table
        rowKey="id"
        dataSource={accounts}
        columns={[
          { title: '用户名', dataIndex: 'username' },
          {
            title: '可用上游',
            dataIndex: 'allowed_provider_ids',
            render: (ids: number[]) => (
              <Space wrap>
                {ids.map((id) => (
                  <Tag key={id}>{providerNameMap.get(id) || `#${id}`}</Tag>
                ))}
              </Space>
            ),
          },
          { title: '启用', dataIndex: 'active', width: 100, render: (v) => <Switch checked={v} disabled /> },
          { title: '创建时间', dataIndex: 'created_at', width: 220 },
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
                  icon={<DeleteOutlined />}
                  onClick={() => {
                    if (confirm(`确认删除下游「${record.username}」？`)) onDelete(record)
                  }}
                >
                  删除
                </Button>
                <Button size="small" icon={<SendOutlined />} onClick={() => onTest(record)}>
                  发送邮件
                </Button>
              </Space>
            ),
          },
        ]}
      />
    </Card>
  )
}
