package server

import (
	"database/sql"
	"encoding/json"
	"io"

	"github.com/gin-gonic/gin"
)

func (a *App) listMessages(c *gin.Context) {
	rows, err := a.db.Query(`select id,downstream_account_id,downstream_from,downstream_to,subject,sent_raw,provider_id,provider_type,rule_id,template_id,template_data,status,error,provider_message_id,callback_event,callback_reason,bounce_type,created_at,updated_at from messages order by id desc limit 200`)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()
	items := []gin.H{}
	for rows.Next() {
		var id int64
		var downstreamAccountID, providerID sql.NullInt64
		var ruleID, templateID sql.NullInt64
		var from, to, subject, sentRaw, data, status, errText, providerType, providerMessageID, event, reason, bounce, createdAt, updatedAt string
		_ = rows.Scan(&id, &downstreamAccountID, &from, &to, &subject, &sentRaw, &providerID, &providerType, &ruleID, &templateID, &data, &status, &errText, &providerMessageID, &event, &reason, &bounce, &createdAt, &updatedAt)
		items = append(items, gin.H{
			"id":                    id,
			"downstream_account_id": nullableInt(downstreamAccountID),
			"from":                  from,
			"to":                    to,
			"subject":               subject,
			"sent_raw":              sentRaw,
			"provider_id":           nullableInt(providerID),
			"provider_type":         providerType,
			"rule_id":               nullableInt(ruleID),
			"template_id":           nullableInt(templateID),
			"template_data":         jsonRawObject(data),
			"status":                status,
			"error":                 errText,
			"provider_message_id":   providerMessageID,
			"callback_event":        event,
			"callback_reason":       reason,
			"bounce_type":           bounce,
			"created_at":            createdAt,
			"updated_at":            updatedAt,
		})
	}
	c.JSON(200, items)
}

func (a *App) stats(c *gin.Context) {
	var total, sent, delivered, bounce, failed int64
	_ = a.db.QueryRow(`select count(*) from messages`).Scan(&total)
	_ = a.db.QueryRow(`select count(*) from messages where status in ('sent','delivered','bounce','dropped','open','click')`).Scan(&sent)
	_ = a.db.QueryRow(`select count(*) from messages where status='delivered'`).Scan(&delivered)
	_ = a.db.QueryRow(`select count(*) from messages where status='bounce'`).Scan(&bounce)
	_ = a.db.QueryRow(`select count(*) from messages where status='failed'`).Scan(&failed)
	rate := func(n int64) float64 {
		if sent == 0 {
			return 0
		}
		return float64(n) / float64(sent)
	}
	c.JSON(200, gin.H{"total": total, "sent": sent, "delivered": delivered, "bounce": bounce, "failed": failed, "delivery_rate": rate(delivered), "bounce_rate": rate(bounce)})
}

func (a *App) tencentCallback(c *gin.Context) {
	body, _ := io.ReadAll(c.Request.Body)
	var payload map[string]any
	_ = json.Unmarshal(body, &payload)
	event := strAny(payload["event"])
	reason := strAny(payload["reason"])
	bounceType := strAny(payload["bounceType"])
	email := strAny(payload["email"])
	providerID := firstNonEmpty(strAny(payload["bulkId"]), strAny(payload["messageId"]), strAny(payload["MessageId"]))
	var messageID sql.NullInt64
	_ = a.db.QueryRow(`select id from messages where provider_message_id=? order by id desc limit 1`, providerID).Scan(&messageID)
	_, _ = a.db.Exec(`insert into callback_events(message_id,provider_message_id,event,reason,bounce_type,email,payload,created_at) values(?,?,?,?,?,?,?,?)`, nullableInt(messageID), providerID, event, reason, bounceType, email, string(body), now())
	if messageID.Valid {
		status := event
		if status == "" {
			status = "callback"
		}
		_, _ = a.db.Exec(`update messages set status=?, callback_event=?, callback_reason=?, bounce_type=?, updated_at=? where id=?`, status, event, reason, bounceType, now(), messageID.Int64)
	}
	c.JSON(200, gin.H{"ok": true})
}
