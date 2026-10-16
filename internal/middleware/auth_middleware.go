package middleware

import (
	"context"
	"net/http"

	"EnProject/internal/database" // Importa apenas o database, que não gera ciclo

	"github.com/gin-gonic/gin"
)

func AutenticacaoObrigatoria() gin.HandlerFunc {
	return func(c *gin.Context) {
		cookieEmail, err := c.Cookie("sessao_token")
		if err != nil || cookieEmail == "" {
			c.Redirect(http.StatusSeeOther, "/login")
			c.Abort()
			return
		}

		// 🔍 Busca os dados reais do usuário (ID e Perfil) no banco usando o e-mail do cookie
		var id int
		var perfil string
		query := `SELECT id, perfil FROM usuarios WHERE email = $1`

		errQuery := database.DB.QueryRow(context.Background(), query, cookieEmail).Scan(&id, &perfil)
		if errQuery != nil {
			c.Redirect(http.StatusSeeOther, "/login")
			c.Abort()
			return
		}

		// 🌟 INJETA OS DADOS NO CONTEXTO DO GIN PARA OS HANDLERS COLETAREM
		c.Set("usuario_id", id)
		c.Set("is_admin", perfil == "admin")

		c.Next()
	}
}

// RequererAdmin corrigido sem dependência do pacote services!
func RequererAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Como AutenticacaoObrigatoria roda antes, pegamos o perfil direto do contexto
		isAdminField, existe := c.Get("is_admin")

		if !existe || !isAdminField.(bool) {
			c.String(http.StatusForbidden, "Acesso Negado: Apenas administradores podem realizar esta ação.")
			c.Abort()
			return

		}

		c.Next()
	}
}
