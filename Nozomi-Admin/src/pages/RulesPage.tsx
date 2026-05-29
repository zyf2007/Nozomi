import { Button, Card, Space, Switch, Table } from 'antd'
import type { FormInstance } from 'antd'
import { CodeOutlined } from '@ant-design/icons'
import type { Rule } from '../types'
import { requestJson } from '../api/client'

type Props = {
  rules: Rule[]
  ruleForm: FormInstance
  onEdit: (rule: Rule | null) => void
  onRefresh: () => void
}

export function RulesPage({ rules, ruleForm, onEdit, onRefresh }: Props) {
  return (
    <Card
      title="下游邮件解析规则"
      extra={
        <Button
          type="primary"
          icon={<CodeOutlined />}
          onClick={() => {
            onEdit(null)
            ruleForm.setFieldsValue({ enabled: true, priority: 100, script: '' })
          }}
        >
          新增规则
        </Button>
      }
    >
      <Table
        rowKey="id"
        dataSource={rules}
        columns={[
          { title: '优先级', dataIndex: 'priority', width: 90 },
          { title: '名称', dataIndex: 'name' },
          { title: '启用', dataIndex: 'enabled', width: 90, render: (v) => <Switch checked={v} disabled /> },
          { title: '更新时间', dataIndex: 'updated_at', width: 220 },
          {
            title: '操作',
            width: 160,
            render: (_, record) => (
              <Space>
                <Button
                  size="small"
                  onClick={() => {
                    onEdit(record)
                    ruleForm.setFieldsValue(record)
                  }}
                >
                  编辑
                </Button>
                <Button
                  size="small"
                  danger
                  onClick={async () => {
                    await requestJson(`/api/rules/${record.id}`, { method: 'DELETE' })
                    onRefresh()
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
