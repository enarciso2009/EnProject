package middleware

import (
	"context"
	"net/http"

	"EnProject/internal/database" // Importa apenas o database, que não gera ciclo

	"github.com/gin-gonic/gin"
)

func AutenticacaoObrigatoria() gin.HandlerFunc {
	return func(c *gin.Context) {
		cookie, err := c.Cookie("sessao_token")
		if err != nil || cookie == "" {
			c.Redirect(http.StatusSeeOther, "/login")
			c.Abort()
			return
		}
		c.Next()
	}
}

// RequererAdmin corrigido sem dependência do pacote services!
func RequererAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		cookie, err := c.Cookie("sessao_token")
		if err != nil || cookie == "" {
			c.Redirect(http.StatusSeeOther, "/login")
			c.Abort()
			return
		}

		// Busca o perfil diretamente do banco de dados para quebrar o ciclo de imports
		var perfil string
		query := `SELECT perfil FROM usuarios WHERE email = $1`
		err = database.DB.QueryRow(context.Background(), query, cookie).Scan(&perfil)

		if err != nil || perfil != "admin" {
			c.String(http.StatusForbidden, "Acesso Negado: Apenas administradores podem realizar esta ação.")
			c.Abort()
			return
		}

		c.Next()
	}
}
