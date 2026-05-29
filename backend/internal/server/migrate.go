package server

func (a *App) migrate(isNewDB bool) error {
	stmts := []string{
		`create table if not exists settings (key text primary key, value text not null default '')`,
		`create table if not exists upstream_providers (id integer primary key autoincrement, name text not null, type text not null, enabled integer not null default 1, weight integer not null default 100, daily_limit integer not null default 0, created_at text not null, updated_at text not null)`,
		`create table if not exists provider_tencent_config (provider_id integer primary key, secret_id text not null default '', secret_key text not null default '', region text not null default 'ap-guangzhou', from_address text not null default '', reply_to text not null default '', trigger_type text not null default '1', foreign key(provider_id) references upstream_providers(id) on delete cascade)`,
		`create table if not exists provider_smtp_config (provider_id integer primary key, host text not null default '', port integer not null default 25, username text not null default '', password text not null default '', from_address text not null default '', reply_to text not null default '', foreign key(provider_id) references upstream_providers(id) on delete cascade)`,
		`create table if not exists provider_rules (id integer primary key autoincrement, provider_id integer not null, name text not null, enabled integer not null default 1, priority integer not null default 100, script text not null, created_at text not null, updated_at text not null, foreign key(provider_id) references upstream_providers(id) on delete cascade)`,
		`create table if not exists provider_templates (id integer primary key, provider_id integer not null, name text not null, status integer not null, variables text not null default '[]', html text not null default '', text text not null default '', updated_at text not null, foreign key(provider_id) references upstream_providers(id) on delete cascade)`,
		`create table if not exists smtp_accounts (id integer primary key autoincrement, username text not null unique, password text not null, active integer not null default 1, allowed_provider_ids text not null default '[]', created_at text not null, updated_at text not null)`,
		`create table if not exists provider_daily_usage (provider_id integer not null, usage_date text not null, sent_count integer not null default 0, primary key(provider_id, usage_date), foreign key(provider_id) references upstream_providers(id) on delete cascade)`,
		`create table if not exists provider_rr_state (downstream_id integer primary key, cursor integer not null default 0, updated_at text not null)`,
		`create table if not exists messages (id integer primary key autoincrement, downstream_account_id integer, downstream_from text, downstream_to text, subject text, raw text, text_body text, html_body text, provider_id integer, provider_type text, rule_id integer, template_id integer, template_data text, status text not null, error text not null default '', provider_message_id text, callback_event text, callback_reason text, bounce_type text, created_at text not null, updated_at text not null)`,
		`create table if not exists callback_events (id integer primary key autoincrement, message_id integer, provider_message_id text, event text, reason text, bounce_type text, email text, payload text not null, created_at text not null)`,
	}
	for _, stmt := range stmts {
		if _, err := a.db.Exec(stmt); err != nil {
			return err
		}
	}

	if isNewDB {
		if _, err := a.db.Exec(`insert or ignore into settings(key,value) values(?,?)`, settingUpstreamDispatchMode, "queue"); err != nil {
			return err
		}
		if _, err := a.db.Exec(`insert or ignore into upstream_providers(id,name,type,enabled,weight,daily_limit,created_at,updated_at) values(1,'默认腾讯云 SES','tencent',1,100,0,?,?)`, now(), now()); err != nil {
			return err
		}
		if _, err := a.db.Exec(`insert or ignore into provider_tencent_config(provider_id,secret_id,secret_key,region,from_address,reply_to,trigger_type) values(1,'','','ap-guangzhou','','','1')`); err != nil {
			return err
		}
		if _, err := a.db.Exec(`insert or ignore into smtp_accounts(username,password,active,allowed_provider_ids,created_at,updated_at) values('rauthy','change-me',1,'[1]',?,?)`, now(), now()); err != nil {
			return err
		}
		if _, err := a.db.Exec(`insert into provider_rules(provider_id,name,enabled,priority,script,created_at,updated_at)
			select 1, '默认验证码示例', 1, 100, ?, ?, ?
			where not exists (
				select 1 from provider_rules where provider_id = 1 and name = '默认验证码示例'
			)`, defaultRule(), now(), now()); err != nil {
			return err
		}
	}
	return nil
}

func defaultRule() string {
	return `// input: {from,to,subject,text,html,headers,raw}
const body = input.text || input.html || "";
const code = (body.match(/\b\d{6}\b/) || [])[0];
if (!code) null;
else ({
  templateId: 0,
  subject: input.subject || "验证码",
  variables: {
    action: input.subject || "登录",
    code: code
  }
});`
}
