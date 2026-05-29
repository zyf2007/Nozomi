import { Alert, Card, Col, Progress, Row } from 'antd'
import type { Stats } from '../types'

export function DashboardPage({ stats }: { stats: Stats | null }) {
  return (
    <Row gutter={[16, 16]}>
      <Col xs={24} md={6}>
        <Card title="总请求">{stats?.total ?? 0}</Card>
      </Col>
      <Col xs={24} md={6}>
        <Card title="已提交上游">{stats?.sent ?? 0}</Card>
      </Col>
      <Col xs={24} md={6}>
        <Card title="送达率">
          <Progress percent={Math.round((stats?.delivery_rate || 0) * 100)} />
        </Card>
      </Col>
      <Col xs={24} md={6}>
        <Card title="退信率">
          <Progress
            percent={Math.round((stats?.bounce_rate || 0) * 100)}
            status={(stats?.bounce_rate || 0) > 0.05 ? 'exception' : 'normal'}
          />
        </Card>
      </Col>
      <Col span={24}>
        <Alert
          type="info"
          showIcon
          title="解析规则使用受限 JavaScript 脚本"
          description="脚本拿到 input 邮件对象，返回 { templateId, subject, variables }。返回 null 表示不匹配。后端有 500ms 超时，并会在日志中记录无法匹配或变量缺失的原始邮件。"
        />
      </Col>
    </Row>
  )
}
