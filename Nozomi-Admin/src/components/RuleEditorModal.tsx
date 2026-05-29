import { Button, Col, Form, Input, InputNumber, Modal, Row, Switch, Tag } from 'antd'
import type { FormInstance } from 'antd'
import { ExperimentOutlined } from '@ant-design/icons'
import type { Rule, TemplateOption } from '../types'
import { requestJson } from '../api/client'

type Props = {
  editingRule: Rule | null
  providerId: number | null
  form: FormInstance
  open: boolean
  templateOptions: TemplateOption[]
  onClose: () => void
  onRefresh: () => void
  onTest: () => void
}

export function RuleEditorModal({ editingRule, providerId, form, open, templateOptions, onClose, onRefresh, onTest }: Props) {
  return (
    <Modal
      title={editingRule ? '编辑规则' : '新增规则'}
      open={open}
      width={900}
      onCancel={onClose}
      onOk={() => form.submit()}
    >
      <Form
        form={form}
        layout="vertical"
        onFinish={async (values) => {
          if (!providerId) return
          const url = editingRule ? `/api/providers/${providerId}/rules/${editingRule.id}` : `/api/providers/${providerId}/rules`
          await requestJson(url, { method: editingRule ? 'PUT' : 'POST', body: JSON.stringify(values) })
          onClose()
          onRefresh()
        }}
      >
        <Row gutter={16}>
          <Col span={12}>
            <Form.Item name="name" label="名称" rules={[{ required: true }]}>
              <Input />
            </Form.Item>
          </Col>
          <Col span={6}>
            <Form.Item name="priority" label="优先级">
              <InputNumber min={1} className="full" />
            </Form.Item>
          </Col>
          <Col span={6}>
            <Form.Item name="enabled" label="启用" valuePropName="checked">
              <Switch />
            </Form.Item>
          </Col>
        </Row>
        <Form.Item label="已缓存模板">{templateOptions.map((item) => <Tag key={item.value}>{item.label}</Tag>)}</Form.Item>
        <Form.Item name="script" label="JavaScript 规则脚本" rules={[{ required: true }]}>
          <Input.TextArea rows={16} className="code-area" />
        </Form.Item>
        <div className="modal-actions-row">
          <Button icon={<ExperimentOutlined />} onClick={onTest}>
            测试规则
          </Button>
        </div>
      </Form>
    </Modal>
  )
}
