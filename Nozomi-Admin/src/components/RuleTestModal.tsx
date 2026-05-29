import { Col, Form, Input, Modal, Row, Typography } from 'antd'
import type { FormInstance } from 'antd'
import type { RuleTestResponse } from '../types'

type Props = {
  providerId: number | null
  form: FormInstance
  open: boolean
  result: RuleTestResponse | null
  testing: boolean
  onClose: () => void
  onSubmit: () => void
}

export function RuleTestModal({ providerId, form, open, result, testing, onClose, onSubmit }: Props) {
  return (
    <Modal
      title="测试当前规则脚本"
      open={open}
      width={1080}
      confirmLoading={testing}
      okText="运行规则"
      cancelText="关闭"
      onCancel={onClose}
      onOk={onSubmit}
    >
      <Row gutter={16}>
        <Col xs={24} md={12}>
          <Form form={form} layout="vertical">
            <Form.Item name="from" label="input.from" rules={[{ required: true }]}>
              <Input />
            </Form.Item>
            <Form.Item name="to" label="input.to" rules={[{ required: true }]}>
              <Input placeholder="多个收件人用逗号、空格或换行分隔" />
            </Form.Item>
            <Form.Item name="subject" label="input.subject">
              <Input />
            </Form.Item>
            <Form.Item name="text" label="input.text">
              <Input.TextArea rows={6} />
            </Form.Item>
            <Form.Item name="html" label="input.html">
              <Input.TextArea rows={6} />
            </Form.Item>
            <Form.Item name="raw" label="input.raw">
              <Input.TextArea rows={4} />
            </Form.Item>
            <Form.Item hidden name="provider_id" initialValue={providerId ?? undefined}>
              <Input />
            </Form.Item>
          </Form>
        </Col>
        <Col xs={24} md={12}>
          <Typography.Text strong>规则输出结构</Typography.Text>
          <pre className="json-block test-output">{JSON.stringify(result || {}, null, 2)}</pre>
        </Col>
      </Row>
    </Modal>
  )
}
