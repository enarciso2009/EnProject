package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func AutenticacaoObrigatoria() gin.HandlerFunc {
	return func(c *gin.Context) {
		cookie, err := c.Cookie("sessao_token")

		// Se não encontrou o cookie ou ele está em branco, barra o acesso
		if err != nil || cookie == "" {
			c.Redirect(http.StatusSeeOther, "/login")
			c.Abort() // Para o processamento da requisição aqui
			return
		}

		c.Next()
	}

}
