import { Descriptions, Modal, Space, Tag, Typography } from 'antd'
import type { ProviderTemplate } from '../../types'

type Props = {
  template: ProviderTemplate | null
  onClose: () => void
}

export function ProviderTemplateDetailModal({ template, onClose }: Props) {
  return (
    <Modal title={template ? `模板详情 · ${template.name}` : '模板详情'} open={!!template} width={1040} footer={null} onCancel={onClose}>
      {template ? (
        <Space direction="vertical" size={16} className="full">
          <Descriptions size="small" bordered column={2}>
            <Descriptions.Item label="模板 ID">{template.id}</Descriptions.Item>
            <Descriptions.Item label="状态">{template.status}</Descriptions.Item>
            <Descriptions.Item label="名称">{template.name}</Descriptions.Item>
            <Descriptions.Item label="更新时间">{template.updated_at}</Descriptions.Item>
            <Descriptions.Item label="变量" span={2}>
              <Space wrap>
                {template.variables.map((item) => (
                  <Tag key={item}>{item}</Tag>
                ))}
              </Space>
            </Descriptions.Item>
          </Descriptions>
          <div>
            <Typography.Text strong>Text 原文</Typography.Text>
            <pre className="json-block template-source-block">{template.text || '(空)'}</pre>
          </div>
          <div>
            <Typography.Text strong>HTML 原文</Typography.Text>
            <pre className="json-block template-source-block">{template.html || '(空)'}</pre>
          </div>
        </Space>
      ) : null}
    </Modal>
  )
}
