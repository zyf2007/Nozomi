import { useCallback, useEffect, useMemo, useState } from 'react'
import { Form, message } from 'antd'
import { requestJson } from './api/client'
import { AdminShell } from './components/AdminShell'
import { LoginPage } from './components/LoginPage'
import { ProviderModal } from './components/ProviderModal'
import { RuleEditorModal } from './components/RuleEditorModal'
import { RuleTestModal } from './components/RuleTestModal'
import { SMTPAccountModal } from './components/SMTPAccountModal'
import { SMTPTestModal } from './components/SMTPTestModal'
import { DashboardPage } from './pages/DashboardPage'
import { LogsPage } from './pages/LogsPage'
import { SettingsPage } from './pages/SettingsPage'
import { SMTPAccountsPage } from './pages/SMTPAccountsPage'
import type {
  Provider,
  ProviderDispatchMode,
  ProviderDetail,
  ProviderRule,
  RelayMessage,
  RuleTestResponse,
  SMTPAccount,
  Session,
  Stats,
  TemplateOption,
} from './types'
import { splitRecipients } from './utils/format'
import './App.css'

function App() {
  const [session, setSession] = useState<Session>({ authenticated: false, username: '' })
  const [loading, setLoading] = useState(true)
  const [collapsed, setCollapsed] = useState(false)
  const [activeKey, setActiveKey] = useState('dashboard')
  const [stats, setStats] = useState<Stats | null>(null)
  const [providers, setProviders] = useState<Provider[]>([])
  const [dispatchMode, setDispatchMode] = useState<ProviderDispatchMode>('queue')
  const [accounts, setAccounts] = useState<SMTPAccount[]>([])
  const [messages, setMessages] = useState<RelayMessage[]>([])
  const [providerOpen, setProviderOpen] = useState(false)
  const [providerSavingOrder, setProviderSavingOrder] = useState(false)
  const [providerDetail, setProviderDetail] = useState<ProviderDetail | null>(null)
  const [editingAccount, setEditingAccount] = useState<SMTPAccount | null>(null)
  const [accountOpen, setAccountOpen] = useState(false)
  const [editingRule, setEditingRule] = useState<ProviderRule | null>(null)
  const [ruleProviderId, setRuleProviderId] = useState<number | null>(null)
  const [ruleOpen, setRuleOpen] = useState(false)
  const [ruleTestOpen, setRuleTestOpen] = useState(false)
  const [ruleTestProviderId, setRuleTestProviderId] = useState<number | null>(null)
  const [ruleTestResult, setRuleTestResult] = useState<RuleTestResponse | null>(null)
  const [smtpTestOpen, setSmtpTestOpen] = useState(false)
  const [smtpTestAccount, setSmtpTestAccount] = useState<SMTPAccount | null>(null)
  const [smtpTestResult, setSmtpTestResult] = useState<Record<string, unknown> | null>(null)
  const [testing, setTesting] = useState(false)
  const [providerForm] = Form.useForm()
  const [loginForm] = Form.useForm()
  const [ruleForm] = Form.useForm()
  const [accountForm] = Form.useForm()
  const [smtpTestForm] = Form.useForm()
  const [ruleTestForm] = Form.useForm()

  const refresh = useCallback(async () => {
    if (!session.authenticated) return
    const [nextStats, nextProviders, nextAccounts, nextMessages, nextDispatchMode] = await Promise.all([
      requestJson<Stats>('/api/stats'),
      requestJson<Provider[]>('/api/providers'),
      requestJson<SMTPAccount[]>('/api/smtp-accounts'),
      requestJson<RelayMessage[]>('/api/messages'),
      requestJson<{ mode: ProviderDispatchMode }>('/api/providers/dispatch-mode'),
    ])
    setStats(nextStats)
    setProviders(Array.isArray(nextProviders) ? nextProviders : [])
    setAccounts(Array.isArray(nextAccounts) ? nextAccounts : [])
    setMessages(Array.isArray(nextMessages) ? nextMessages : [])
    setDispatchMode(nextDispatchMode?.mode || 'queue')
  }, [session.authenticated])

  const refreshProviderDetail = useCallback(
    async (providerId: number) => {
      const detail = await requestJson<ProviderDetail>(`/api/providers/${providerId}`)
      setProviderDetail(detail)
      const fromAddress =
        detail.tencent_config.from_address ||
        detail.smtp_config.from_address ||
        detail.resend_config.from_address ||
        detail.brevo_config.from_address
      const replyTo =
        detail.tencent_config.reply_to ||
        detail.smtp_config.reply_to ||
        detail.resend_config.reply_to ||
        detail.brevo_config.reply_to
      providerForm.setFieldsValue({
        name: detail.name,
        type: detail.type,
        enabled: detail.enabled,
        weight: detail.weight,
        daily_limit: detail.daily_limit,
        secret_id: detail.tencent_config.secret_id,
        secret_key: detail.tencent_config.secret_key,
        region: detail.tencent_config.region,
        from_address: fromAddress,
        reply_to: replyTo,
        trigger_type: detail.tencent_config.trigger_type,
        host: detail.smtp_config.host,
        port: detail.smtp_config.port || 25,
        username: detail.smtp_config.username,
        password: '',
        api_key: detail.resend_config.api_key || detail.brevo_config.api_key,
        from_name: detail.brevo_config.from_name || 'Nozomi',
      })
    },
    [providerForm],
  )

  useEffect(() => {
    requestJson<Session>('/api/auth/session')
      .then(setSession)
      .finally(() => setLoading(false))
  }, [])

  useEffect(() => {
    void Promise.resolve().then(() => refresh().catch((err) => message.error(err.message)))
  }, [refresh])

  const providerTemplateOptions: TemplateOption[] = useMemo(
    () => (providerDetail?.templates || []).map((item) => ({ label: `${item.id} · ${item.name}`, value: item.id })),
    [providerDetail],
  )

  const login = async (values: { username: string; password: string }) => {
    const next = await requestJson<Session>('/api/auth/login', {
      method: 'POST',
      body: JSON.stringify(values),
    })
    setSession(next)
  }

  const logout = async () => {
    await requestJson('/api/auth/logout', { method: 'POST' })
    setSession({ authenticated: false, username: '' })
  }

  const openProviderCreate = () => {
    setProviderDetail(null)
    providerForm.resetFields()
    providerForm.setFieldsValue({
      type: 'tencent',
      enabled: true,
      weight: 100,
      daily_limit: 0,
      region: 'ap-guangzhou',
      trigger_type: '1',
      port: 25,
      api_key: '',
      from_address: '',
      reply_to: '',
      from_name: 'Nozomi',
    })
    setProviderOpen(true)
  }

  const openProviderEdit = async (provider: Provider) => {
    setProviderOpen(true)
    await refreshProviderDetail(provider.id)
  }

  const closeProvider = () => {
    setProviderOpen(false)
    setProviderDetail(null)
  }

  const saveProviderConfig = async () => {
    setProviderSavingOrder(true)
    try {
      await Promise.all([
        requestJson('/api/providers/dispatch-mode', {
          method: 'PUT',
          body: JSON.stringify({ mode: dispatchMode }),
        }),
        requestJson('/api/providers/reorder', {
          method: 'POST',
          body: JSON.stringify({ items: providers.map((item) => ({ id: item.id, weight: item.weight })) }),
        }),
      ])
      message.success('排序已保存')
      refresh()
    } finally {
      setProviderSavingOrder(false)
    }
  }

  const deleteProvider = async (provider: Provider) => {
    await requestJson(`/api/providers/${provider.id}`, { method: 'DELETE' })
    message.success('已删除')
    refresh()
  }

  const openAccountCreate = () => {
    setEditingAccount(null)
    accountForm.resetFields()
    accountForm.setFieldsValue({
      active: true,
      allowed_provider_ids: providers.length ? [providers[0].id] : [],
    })
    setAccountOpen(true)
  }

  const openAccountEdit = (account: SMTPAccount) => {
    setEditingAccount(account)
    accountForm.setFieldsValue({ ...account, password: '' })
    setAccountOpen(true)
  }

  const openSMTPTest = (account: SMTPAccount) => {
    setSmtpTestAccount(account)
    setSmtpTestResult(null)
    smtpTestForm.setFieldsValue({
      from: 'tester@nozomi-relay.local',
      to: 'user@example.com',
      subject: '登录验证码',
      text: '你的登录验证码是 123456，5 分钟内有效。',
      html: '',
    })
    setSmtpTestOpen(true)
  }

  const openRuleEdit = (providerId: number, rule: ProviderRule | null) => {
    setRuleProviderId(providerId)
    setEditingRule(rule)
    if (rule) {
      ruleForm.setFieldsValue(rule)
    } else {
      ruleForm.setFieldsValue({ enabled: true, priority: 100, script: '' })
    }
    setRuleOpen(true)
  }

  const openRuleTest = (providerId: number) => {
    setRuleTestProviderId(providerId)
    setRuleTestResult(null)
    ruleTestForm.setFieldsValue({
      from: 'rauthy@example.com',
      to: 'user@example.com',
      subject: '登录验证码',
      text: '你的登录验证码是 123456，5 分钟内有效。',
      html: '',
      raw: '',
    })
    setRuleTestOpen(true)
  }

  const submitSMTPTest = async () => {
    if (!smtpTestAccount) return
    const values = await smtpTestForm.validateFields()
    setTesting(true)
    try {
      const result = await requestJson<Record<string, unknown>>(`/api/smtp-accounts/${smtpTestAccount.id}/test`, {
        method: 'POST',
        body: JSON.stringify({ ...values, to: splitRecipients(values.to) }),
      })
      setSmtpTestResult(result)
      message.success('测试邮件已发送')
      refresh()
    } catch (err) {
      const error = err instanceof Error ? err.message : String(err)
      setSmtpTestResult({ ok: false, error })
      message.error(error)
      refresh()
    } finally {
      setTesting(false)
    }
  }

  const submitRuleTest = async () => {
    if (!ruleTestProviderId) return
    const [ruleValues, inputValues] = await Promise.all([ruleForm.validateFields(['script']), ruleTestForm.validateFields()])
    setTesting(true)
    try {
      const result = await requestJson<RuleTestResponse>(`/api/providers/${ruleTestProviderId}/rules/test`, {
        method: 'POST',
        body: JSON.stringify({
          provider_id: ruleTestProviderId,
          script: ruleValues.script,
          input: { ...inputValues, to: splitRecipients(inputValues.to), headers: {} },
        }),
      })
      setRuleTestResult(result)
    } catch (err) {
      setRuleTestResult({ matched: false, error: err instanceof Error ? err.message : String(err) })
    } finally {
      setTesting(false)
    }
  }

  const renderPage = () => {
    if (activeKey === 'settings') {
      return (
        <SettingsPage
          providers={providers}
          dispatchMode={dispatchMode}
          saving={providerSavingOrder}
          onCreate={openProviderCreate}
          onEdit={openProviderEdit}
          onDelete={deleteProvider}
          onWeightsChange={setProviders}
          onDispatchModeChange={setDispatchMode}
          onSave={saveProviderConfig}
        />
      )
    }
    if (activeKey === 'smtp') {
      return (
        <SMTPAccountsPage
          accounts={accounts}
          providers={providers}
          onCreate={openAccountCreate}
          onEdit={openAccountEdit}
          onDelete={async (account) => {
            await requestJson(`/api/smtp-accounts/${account.id}`, { method: 'DELETE' })
            message.success('已删除')
            refresh()
          }}
          onTest={openSMTPTest}
        />
      )
    }
    if (activeKey === 'logs') return <LogsPage messages={messages} />
    return <DashboardPage stats={stats} />
  }

  if (loading) return null
  if (!session.authenticated) return <LoginPage form={loginForm} onLogin={login} />

  return (
    <AdminShell
      activeKey={activeKey}
      collapsed={collapsed}
      onActiveKeyChange={setActiveKey}
      onCollapsedChange={setCollapsed}
      onLogout={logout}
      onRefresh={refresh}
    >
      {renderPage()}
      <ProviderModal
        provider={providerDetail}
        form={providerForm}
        open={providerOpen}
        onClose={closeProvider}
        onSaved={() => {
          message.success('已保存')
          closeProvider()
          refresh()
        }}
        onSyncTemplates={async () => {
          if (!providerDetail) return
          await requestJson(`/api/providers/${providerDetail.id}/templates/sync`, { method: 'POST' })
          message.success('模板已同步')
          await refreshProviderDetail(providerDetail.id)
          refresh()
        }}
        onEditRule={(rule) => {
          if (!providerDetail) return
          openRuleEdit(providerDetail.id, rule)
        }}
        onDeleteRule={async (rule) => {
          if (!providerDetail) return
          await requestJson(`/api/providers/${providerDetail.id}/rules/${rule.id}`, { method: 'DELETE' })
          message.success('规则已删除')
          await refreshProviderDetail(providerDetail.id)
          refresh()
        }}
      />
      <RuleEditorModal
        editingRule={editingRule}
        providerId={ruleProviderId}
        form={ruleForm}
        open={ruleOpen}
        templateOptions={providerTemplateOptions}
        onClose={() => setRuleOpen(false)}
        onRefresh={async () => {
          if (providerDetail) {
            await refreshProviderDetail(providerDetail.id)
          }
          refresh()
        }}
        onTest={() => {
          if (ruleProviderId) openRuleTest(ruleProviderId)
        }}
      />
      <SMTPAccountModal
        form={accountForm}
        open={accountOpen}
        account={editingAccount}
        providers={providers}
        onClose={() => setAccountOpen(false)}
        onRefresh={refresh}
      />
      <RuleTestModal
        form={ruleTestForm}
        providerId={ruleTestProviderId}
        open={ruleTestOpen}
        result={ruleTestResult}
        testing={testing}
        onClose={() => setRuleTestOpen(false)}
        onSubmit={submitRuleTest}
      />
      <SMTPTestModal
        account={smtpTestAccount}
        form={smtpTestForm}
        open={smtpTestOpen}
        result={smtpTestResult}
        testing={testing}
        onClose={() => setSmtpTestOpen(false)}
        onSubmit={submitSMTPTest}
      />
    </AdminShell>
  )
}

export default App
