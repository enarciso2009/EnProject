package handler

import (
	"EnProject/internal/services"
	"net/http"

	"github.com/gin-gonic/gin"
)

// ExibirLogin abre a pagina de login

func ExibirLogin(c *gin.Context) {
	c.HTML(http.StatusOK, "login.html", gin.H{"erro": c.Query("erro")})
}

// ProcessarLogin valida as credenciais e define o Cookie de Sessão

func ProcessarLogin(c *gin.Context) {
	email := c.PostForm("email")
	senha := c.PostForm("senha")

	user, err := services.AutenticarUsuario(email, senha)
	if err != nil {
		// Redireciona de vola com mensagem de erro
		c.Redirect(http.StatusSeeOther, "/login?erro=Email ou senha invalidos")
		return
	}

	// Cria o cookie de sessão seguro (valido por 2 horas)
	// Em Produção, mude 'secure' (6° param) para true se usar HTTPS
	c.SetCookie("sessao_token", user.Email, 7200, "/", "", false, true)

	c.Redirect(http.StatusSeeOther, "/")
}

// ExibirCadastroUsuarios abre a tela de registro de novos acessos
func ExibirCadastroUsuarios(c *gin.Context) {
	c.HTML(http.StatusOK, "cadastro_usuario.html", gin.H{
		"erro":    c.Query("erro"),
		"sucesso": c.Query("sucesso"),
	})
}

// ProcessarCadastroUsuario grava o novo usuário administrador no sistema
func ProcessarCadastroUsuario(c *gin.Context) {
	email := c.PostForm("email")
	senha := c.PostForm("senha")

	if email == "" || len(senha) < 4 {
		c.Redirect(http.StatusSeeOther, "/usuario/novo?erro=Dados invalidos. A senha deve ter no minimo 4 caracteres.")
		return
	}

	err := services.CadastrarUsuario(email, senha)
	if err != nil {
		c.Redirect(http.StatusSeeOther, "/usuario/novo?erro=Email ja cadastrado no sistema.")
		return
	}
	c.Redirect(http.StatusSeeOther, "/usuario/novo?sucesso=Usuario cadastrado com sucesso!")
}

// LogOut limpa a sessão do usuario

func LogOut(c *gin.Context) {
	c.SetCookie("sessao_token", "", -1, "/", "", false, true)
	c.Redirect(http.StatusSeeOther, "/login")
}
