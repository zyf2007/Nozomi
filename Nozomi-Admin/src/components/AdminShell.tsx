import { Button, Layout, Menu, Space, Typography } from 'antd'
import {
  BarChartOutlined,
  HistoryOutlined,
  ApiOutlined,
  KeyOutlined,
  LogoutOutlined,
  MenuFoldOutlined,
  MenuUnfoldOutlined,
  ReloadOutlined,
} from '@ant-design/icons'
import type { ReactNode } from 'react'

const { Header, Sider, Content } = Layout

type Props = {
  activeKey: string
  collapsed: boolean
  children: ReactNode
  onActiveKeyChange: (key: string) => void
  onCollapsedChange: (collapsed: boolean) => void
  onLogout: () => Promise<void>
  onRefresh: () => void
}

const tabTitles: Record<string, string> = {
  dashboard: '主页',
  settings: '上游配置',
  smtp: 'SMTP 下游',
  logs: '日志',
}

export function AdminShell({
  activeKey,
  collapsed,
  children,
  onActiveKeyChange,
  onCollapsedChange,
  onLogout,
  onRefresh,
}: Props) {
  return (
    <Layout className="shell">
      <Sider collapsible collapsed={collapsed} collapsedWidth={88} trigger={null} width={232} className="sidebar">
        <div className={`brand${collapsed ? ' brand-collapsed' : ''}`}>{collapsed ? 'Nozomi' : 'Nozomi Relay'}</div>
        <Menu
          mode="inline"
          selectedKeys={[activeKey]}
          onClick={(item) => onActiveKeyChange(item.key)}
          items={[
            { key: 'dashboard', icon: <BarChartOutlined />, label: '主页' },
            { key: 'settings', icon: <ApiOutlined />, label: '上游配置' },
            { key: 'smtp', icon: <KeyOutlined />, label: 'SMTP 下游' },
            { key: 'logs', icon: <HistoryOutlined />, label: '日志' },
          ]}
        />
      </Sider>
      <Layout>
        <Header className="titlebar">
          <Space>
            <Button
              type="text"
              icon={collapsed ? <MenuUnfoldOutlined /> : <MenuFoldOutlined />}
              onClick={() => onCollapsedChange(!collapsed)}
            />
            <Typography.Title level={4}>{tabTitles[activeKey] || '主页'}</Typography.Title>
          </Space>
          <Space>
            <Button icon={<ReloadOutlined />} onClick={onRefresh}>
              刷新
            </Button>
            <Button icon={<LogoutOutlined />} onClick={onLogout}>
              退出
            </Button>
          </Space>
        </Header>
        <Content className="content">{children}</Content>
      </Layout>
    </Layout>
  )
}
