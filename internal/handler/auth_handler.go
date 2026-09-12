package handler

import (
	"EnProject/internal/models"
	"EnProject/internal/services"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt" // Importação correta para segurança da senha
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
		// Redireciona de volta com mensagem de erro
		c.Redirect(http.StatusSeeOther, "/login?erro=Email ou senha invalidos")
		return
	}

	// Cria o cookie de sessão seguro (valido por 2 horas)
	c.SetCookie("sessao_token", user.Email, 7200, "/", "", false, true)

	c.Redirect(http.StatusSeeOther, "/")
}

// ExibirCadastroUsuarios abre a tela de registro alimentando as caixas de seleção para o Admin
func ExibirCadastroUsuarios(c *gin.Context) {
	// Passando (0, "admin") para que liste todos os projetos disponíveis no sistema
	listaProjetos, err := services.ListarTodosProjetos(0, "admin")
	if err != nil {
		listaProjetos = []models.Projeto{}
	}

	c.HTML(http.StatusOK, "cadastro_usuario.html", gin.H{
		"Erro":     c.Query("erro"),
		"Sucesso":  c.Query("sucesso"),
		"Projetos": listaProjetos,
	})
}

// ProcessarCadastroUsuario grava o novo usuário (Admin ou Consultor) com criptografia e tabela pivô 🔒
func ProcessarCadastroUsuario(c *gin.Context) {
	email := c.PostForm("email")
	senha := c.PostForm("senha")
	perfil := c.PostForm("perfil") // Captura o perfil ('admin' ou 'consultor')

	if email == "" || len(senha) < 4 {
		c.Redirect(http.StatusSeeOther, "/usuario/novo?erro=Dados invalidos. A senha deve ter no minimo 4 caracteres.")
		return
	}

	if perfil == "" {
		perfil = "consultor" // Padrão de segurança
	}

	// 🔒 GERA O HASH DA SENHA ANTES DE ENVIAR PARA O BANCO (Protege contra vazamentos)
	senhaHash, err := bcrypt.GenerateFromPassword([]byte(senha), bcrypt.DefaultCost)
	if err != nil {
		c.Redirect(http.StatusSeeOther, "/usuario/novo?erro=Erro interno ao processar a segurança da senha.")
		return
	}

	// Captura o array de checkboxes dos projetos vinculados enviados pelo formulário HTML
	projetosStrArr := c.PostFormArray("projetos_vinculados[]")
	var projetoIDs []int

	// Se for consultor, faz a conversão segura de string para inteiro (int)
	if perfil == "consultor" {
		for _, idStr := range projetosStrArr {
			if id, err := strconv.Atoi(idStr); err == nil {
				projetoIDs = append(projetoIDs, id)
			}
		}
	}

	// Estrutura o modelo com a SENHA CRIPTOGRAFADA
	novoUsuario := models.Usuario{
		Email:  email,
		Senha:  string(senhaHash), // Enviando o hash seguro mapeado
		Perfil: perfil,
	}

	// Chama o serviço transacionado para persistir os dados e criar os vínculos
	err = services.CriarUsuarioCompleto(&novoUsuario, projetoIDs)
	if err != nil {
		c.Redirect(http.StatusSeeOther, "/usuario/novo?erro=Email ja cadastrado no sistema ou erro ao vincular projetos.")
		return
	}

	c.Redirect(http.StatusSeeOther, "/usuario/novo?sucesso=Usuario cadastrado com sucesso!")
}

// ExibirFormularioUsuario busca os projetos e abre a tela de criação de usuário alternativa
func ExibirFormularioUsuario(c *gin.Context) {
	projetos, err := services.ListarTodosProjetos(0, "admin")
	if err != nil {
		c.String(http.StatusInternalServerError, "Erro ao carregar lista de projetos: %v", err)
		return
	}
	c.HTML(http.StatusOK, "novo_usuario.html", projetos)
}

// LogOut limpa a sessão do usuario e remove o cookie
func LogOut(c *gin.Context) {
	c.SetCookie("sessao_token", "", -1, "/", "", false, true)
	c.Redirect(http.StatusSeeOther, "/login")
}
