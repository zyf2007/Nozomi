package server

import (
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"time"
)

func decodeB64(raw string) string {
	if raw == "" {
		return ""
	}
	b, err := base64.StdEncoding.DecodeString(raw)
	if err != nil {
		return raw
	}
	return string(b)
}

func findTemplateVars(s string) []string {
	re := regexp.MustCompile(`\{\{\s*([a-zA-Z0-9_.-]+)\s*\}\}`)
	seen := map[string]bool{}
	var out []string
	for _, m := range re.FindAllStringSubmatch(s, -1) {
		if !seen[m[1]] {
			seen[m[1]] = true
			out = append(out, m[1])
		}
	}
	return out
}

func now() string { return time.Now().Format(time.RFC3339) }

func validTimezone(name string) string {
	name = strings.TrimSpace(name)
	if name == "" {
		return "Asia/Shanghai"
	}
	if _, err := time.LoadLocation(name); err != nil {
		return "Asia/Shanghai"
	}
	return name
}

func providerUsageDate(timezone string) string {
	loc, err := time.LoadLocation(validTimezone(timezone))
	if err != nil {
		loc = time.Local
	}
	return time.Now().In(loc).Format("2006-01-02")
}

func boolInt(v bool) int {
	if v {
		return 1
	}
	return 0
}

func ptr(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func ptrStr(s string) *string { return &s }

func nullableInt(v sql.NullInt64) any {
	if v.Valid {
		return v.Int64
	}
	return nil
}

func nullableString(v sql.NullString) any {
	if v.Valid {
		return v.String
	}
	return nil
}

func jsonRawArray(s string) []string {
	var out []string
	if json.Unmarshal([]byte(s), &out) == nil {
		return out
	}
	return []string{}
}

func jsonRawIntArray(s string) []int64 {
	var out []int64
	if json.Unmarshal([]byte(s), &out) == nil {
		return out
	}
	return []int64{}
}

func jsonRawObject(s string) map[string]string {
	var out map[string]string
	if json.Unmarshal([]byte(s), &out) == nil {
		return out
	}
	return map[string]string{}
}

func jsonString(v any) string {
	b, _ := json.Marshal(v)
	return string(b)
}

func strAny(v any) string {
	if v == nil {
		return ""
	}
	return fmt.Sprint(v)
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if strings.TrimSpace(v) != "" {
			return v
		}
	}
	return ""
}
