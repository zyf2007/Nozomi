package server

import "testing"

func TestRunRuleUsesLastExpressionResult(t *testing.T) {
	script := `const body = input.text || input.html || "";
const actionMatch = body.match(/(你的)?(登录)?验证码/);
const action = actionMatch ? (actionMatch[2] || "验证码") : "验证码";
const code = (body.match(/\b\d{6}\b/) || [])[0];

if (!code) null;
else ({
  templateId: 10001,
  subject: input.subject || "验证码",
  variables: {
    action: action === "登录" ? "登录" : "验证码",
    code: code
  }
	});`
	result, matched, err := runRule(script, MailInput{
		Subject: "登录验证码",
		Text:    "你的登录验证码是 123456，5 分钟内有效。",
	})
	if err != nil {
		t.Fatalf("runRule returned error: %v", err)
	}
	if !matched {
		t.Fatal("expected rule to match")
	}
	if result.TemplateID != 10001 {
		t.Fatalf("template id = %d, want 10001", result.TemplateID)
	}
	if result.Variables["action"] != "登录" || result.Variables["code"] != "123456" {
		t.Fatalf("variables = %#v", result.Variables)
	}
}

func TestRunRuleSupportsExplicitReturn(t *testing.T) {
	result, matched, err := runRule(`return ({ templateId: 10001, variables: { code: "123456" } });`, MailInput{})
	if err != nil {
		t.Fatalf("runRule returned error: %v", err)
	}
	if !matched || result.Variables["code"] != "123456" {
		t.Fatalf("matched=%v result=%#v", matched, result)
	}
}
