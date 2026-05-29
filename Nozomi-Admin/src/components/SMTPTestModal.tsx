import { Col, Form, Input, Modal, Row, Typography } from 'antd'
import type { FormInstance } from 'antd'
import type { SMTPAccount } from '../types'

type Props = {
  account: SMTPAccount | null
  form: FormInstance
  open: boolean
  result: Record<string, unknown> | null
  testing: boolean
  onClose: () => void
  onSubmit: () => void
}

export function SMTPTestModal({ account, form, open, result, testing, onClose, onSubmit }: Props) {
  return (
    <Modal
      title={`测试 SMTP 账号${account ? ` · ${account.username}` : ''}`}
      open={open}
      width={920}
      confirmLoading={testing}
      okText="发送测试"
      cancelText="关闭"
      onCancel={onClose}
      onOk={onSubmit}
    >
      <Row gutter={16}>
        <Col xs={24} md={12}>
          <Form form={form} layout="vertical">
            <Form.Item name="from" label="下游 MAIL FROM" rules={[{ required: true }]}>
              <Input />
            </Form.Item>
            <Form.Item name="to" label="下游 RCPT TO" rules={[{ required: true }]}>
              <Input placeholder="多个收件人用逗号、空格或换行分隔" />
            </Form.Item>
            <Form.Item name="subject" label="Subject" rules={[{ required: true }]}>
              <Input />
            </Form.Item>
            <Form.Item name="text" label="Text Body">
              <Input.TextArea rows={6} />
            </Form.Item>
            <Form.Item name="html" label="HTML Body">
              <Input.TextArea rows={6} />
            </Form.Item>
          </Form>
        </Col>
        <Col xs={24} md={12}>
          <Typography.Text strong>链路结果</Typography.Text>
          <pre className="json-block test-output">{JSON.stringify(result || {}, null, 2)}</pre>
        </Col>
      </Row>
    </Modal>
  )
}
