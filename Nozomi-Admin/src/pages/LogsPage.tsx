import { Card, Table } from 'antd'
import type { RelayMessage } from '../types'
import { statusTag } from '../utils/format'

export function LogsPage({ messages }: { messages: RelayMessage[] }) {
  return (
    <Card title="请求与回调日志">
      <Table
        rowKey="id"
        dataSource={messages}
        scroll={{ x: 1280 }}
        expandable={{
          expandedRowRender: (record) => (
            <pre className="json-block">
              {JSON.stringify(
                {
                  template_data: record.template_data,
                  provider_message_id: record.provider_message_id,
                  callback_reason: record.callback_reason,
                  error: record.error,
                },
                null,
                2,
              )}
            </pre>
          ),
        }}
        columns={[
          { title: 'ID', dataIndex: 'id', width: 80 },
          { title: '状态', dataIndex: 'status', width: 130, render: statusTag },
          { title: '主题', dataIndex: 'subject', width: 240 },
          { title: '下游 From', dataIndex: 'from', width: 220 },
          { title: '模板', dataIndex: 'template_id', width: 110 },
          { title: '规则', dataIndex: 'rule_id', width: 90 },
          { title: '回调', dataIndex: 'callback_event', width: 110 },
          { title: '退信类型', dataIndex: 'bounce_type', width: 130 },
          { title: '创建时间', dataIndex: 'created_at', width: 220 },
        ]}
      />
    </Card>
  )
}
