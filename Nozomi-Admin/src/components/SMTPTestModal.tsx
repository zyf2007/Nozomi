import { Col, Modal, Row, Typography } from 'antd'
import type { Dispatch, SetStateAction } from 'react'
import type { SMTPAccount, SMTPTestDraft } from '../types'
import { MailComposer } from './MailComposer'

type Props = {
  account: SMTPAccount | null
  draft: SMTPTestDraft
  open: boolean
  result: Record<string, unknown> | null
  testing: boolean
  onClose: () => void
  onSubmit: () => void
  onChange: Dispatch<SetStateAction<SMTPTestDraft>>
}

export function SMTPTestModal({ account, draft, open, result, testing, onClose, onSubmit, onChange }: Props) {
  return (
    <Modal
      title={`测试 SMTP 账号${account ? ` · ${account.username}` : ''}`}
      open={open}
      width={1200}
      style={{ top: 20 }}
      confirmLoading={testing}
      okText="发送测试"
      cancelText="关闭"
      onCancel={onClose}
      onOk={onSubmit}
      destroyOnClose
    >
      <Row gutter={16}>
        <Col xs={24} lg={15}>
          <MailComposer value={draft} onChange={onChange} />
        </Col>
        <Col xs={24} lg={9}>
          <Typography.Text strong>链路结果</Typography.Text>
          <pre className="json-block test-output">{JSON.stringify(result || {}, null, 2)}</pre>
        </Col>
      </Row>
    </Modal>
  )
}
