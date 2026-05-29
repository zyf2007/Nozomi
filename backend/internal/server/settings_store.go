package server

import (
	"database/sql"
	"fmt"
)

const settingUpstreamDispatchMode = "upstream_dispatch_mode"

func (a *App) getSetting(key string) (string, error) {
	var value string
	if err := a.db.QueryRow(`select value from settings where key=?`, key).Scan(&value); err != nil {
		if err == sql.ErrNoRows {
			return "", nil
		}
		return "", err
	}
	return value, nil
}

func (a *App) setSetting(key, value string) error {
	_, err := a.db.Exec(`insert into settings(key,value) values(?,?) on conflict(key) do update set value=excluded.value`, key, value)
	return err
}

func (a *App) getUpstreamDispatchMode() (string, error) {
	mode, err := a.getSetting(settingUpstreamDispatchMode)
	if err != nil {
		return "", err
	}
	if mode == "" {
		return "queue", nil
	}
	switch mode {
	case "queue", "round_robin":
		return mode, nil
	default:
		return "", fmt.Errorf("未知的调度模式")
	}
}

func (a *App) setUpstreamDispatchMode(mode string) error {
	switch mode {
	case "queue", "round_robin":
		return a.setSetting(settingUpstreamDispatchMode, mode)
	default:
		return fmt.Errorf("未知的调度模式")
	}
}
