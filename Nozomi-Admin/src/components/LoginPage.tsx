import { Button, Card, Form, Input, Typography } from 'antd'
import type { FormInstance } from 'antd'

type Props = {
  form: FormInstance
  onLogin: (values: { username: string; password: string }) => Promise<void>
}

export function LoginPage({ form, onLogin }: Props) {
  return (
    <main className="login-page">
      <Card className="login-card">
        <Typography.Title level={3}>Nozomi Relay Admin</Typography.Title>
        <Typography.Text type="secondary">SMTP 转多平台 API 控制台管理员登录</Typography.Text>
        <Form form={form} layout="vertical" onFinish={onLogin} className="login-form">
          <Form.Item name="username" label="管理员账号" rules={[{ required: true }]}>
            <Input autoComplete="username" />
          </Form.Item>
          <Form.Item name="password" label="管理员密码" rules={[{ required: true }]}>
            <Input.Password autoComplete="current-password" />
          </Form.Item>
          <Button type="primary" htmlType="submit" block>
            登录
          </Button>
        </Form>
      </Card>
    </main>
  )
}
