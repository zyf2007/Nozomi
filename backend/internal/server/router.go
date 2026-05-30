package server

import (
	"net/http"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func (a *App) router() *gin.Engine {
	r := gin.Default()
	origins := a.settings.CORSOrigins
	if len(origins) == 0 {
		origins = []string{"http://localhost:5173", "http://127.0.0.1:5173"}
	}
	r.Use(cors.New(cors.Config{
		AllowOrigins:     origins,
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Content-Type"},
		AllowCredentials: true,
	}))
	r.POST("/api/auth/login", a.login)
	r.POST("/api/auth/logout", a.logout)
	r.GET("/api/auth/session", a.session)
	admin := r.Group("/api", a.requireAuth)
	admin.GET("/providers", a.listProviders)
	admin.GET("/providers/dispatch-mode", a.getUpstreamDispatchModeAPI)
	admin.PUT("/providers/dispatch-mode", a.putUpstreamDispatchModeAPI)
	admin.POST("/providers", a.createProvider)
	admin.GET("/providers/:id", a.getProvider)
	admin.PUT("/providers/:id", a.updateProvider)
	admin.DELETE("/providers/:id", a.deleteProvider)
	admin.POST("/providers/reorder", a.reorderProviders)
	admin.GET("/providers/:id/tencent", a.getTencentConfig)
	admin.PUT("/providers/:id/tencent", a.putTencentConfig)
	admin.GET("/providers/:id/smtp", a.getSMTPConfig)
	admin.PUT("/providers/:id/smtp", a.putSMTPConfig)
	admin.GET("/providers/:id/resend", a.getResendConfig)
	admin.PUT("/providers/:id/resend", a.putResendConfig)
	admin.GET("/providers/:id/brevo", a.getBrevoConfig)
	admin.PUT("/providers/:id/brevo", a.putBrevoConfig)
	admin.GET("/providers/:id/rules", a.listProviderRules)
	admin.POST("/providers/:id/rules", a.createProviderRule)
	admin.PUT("/providers/:id/rules/:ruleID", a.updateProviderRule)
	admin.DELETE("/providers/:id/rules/:ruleID", a.deleteProviderRule)
	admin.POST("/providers/:id/rules/test", a.testProviderRule)
	admin.GET("/providers/:id/templates", a.listProviderTemplates)
	admin.POST("/providers/:id/templates/sync", a.syncProviderTemplates)
	admin.GET("/smtp-accounts", a.listSMTPAccounts)
	admin.POST("/smtp-accounts", a.createSMTPAccount)
	admin.PUT("/smtp-accounts/:id", a.updateSMTPAccount)
	admin.DELETE("/smtp-accounts/:id", a.deleteSMTPAccount)
	admin.POST("/smtp-accounts/:id/test", a.testSMTPAccount)
	admin.GET("/messages", a.listMessages)
	admin.GET("/stats", a.stats)
	r.POST("/api/callback/tencent", a.tencentCallback)
	return r
}

func (a *App) login(c *gin.Context) {
	var body struct{ Username, Password string }
	if err := c.BindJSON(&body); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	if body.Username != a.settings.AdminUsername || body.Password != a.settings.AdminPassword {
		c.JSON(401, gin.H{"error": "账号或密码错误"})
		return
	}
	http.SetCookie(c.Writer, &http.Cookie{Name: "nozomi_session", Value: a.settings.SessionSecret, Path: "/", HttpOnly: true, SameSite: http.SameSiteLaxMode})
	c.JSON(200, gin.H{"authenticated": true, "username": body.Username})
}

func (a *App) logout(c *gin.Context) {
	http.SetCookie(c.Writer, &http.Cookie{Name: "nozomi_session", Value: "", Path: "/", MaxAge: -1, HttpOnly: true})
	c.JSON(200, gin.H{"ok": true})
}

func (a *App) session(c *gin.Context) {
	_, ok := a.authenticated(c)
	c.JSON(200, gin.H{"authenticated": ok, "username": a.settings.AdminUsername})
}

func (a *App) requireAuth(c *gin.Context) {
	if _, ok := a.authenticated(c); !ok {
		c.AbortWithStatusJSON(401, gin.H{"error": "未登录"})
		return
	}
	c.Next()
}

func (a *App) authenticated(c *gin.Context) (string, bool) {
	cookie, err := c.Cookie("nozomi_session")
	return a.settings.AdminUsername, err == nil && cookie == a.settings.SessionSecret
}
