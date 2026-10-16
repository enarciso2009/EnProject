package handler

import (
	"EnProject/internal/database"
	"EnProject/internal/models"
	"EnProject/internal/services"
	"net/http"
	"strconv"
	"strings"

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

// ExibirCadastroUsuarios abre a tela de registro alimentando as caixas de seleção e a tabela de usuários ativos
func ExibirCadastroUsuarios(c *gin.Context) {
	listaProjetos, err := services.ListarTodosProjetos(0, "admin")
	if err != nil {
		listaProjetos = []models.Projeto{}
	}

	// Criar uma estrutura local para carregar os dados junto com o JSON do projeto
	type UsuarioExibicao struct {
		models.Usuario
		ProjetosJSON string
	}

	// 🌟 NOVO: Busca todos os usuários do banco para listar na tabela inferior
	var listaUsuarios []UsuarioExibicao
	rows, errQuery := database.DB.Query(c.Request.Context(), "SELECT id, email, perfil FROM usuarios ORDER BY email ASC")
	if errQuery == nil {
		defer rows.Close()
		for rows.Next() {
			var u models.Usuario
			if errScan := rows.Scan(&u.ID, &u.Email, &u.Perfil); errScan == nil {

				var stringsIDs []string
				pRows, errProjs := database.DB.Query(c.Request.Context(), "SELECT projeto_id FROM projeto_usuarios WHERE usuario_id = $1", u.ID)
				if errProjs == nil {
					for pRows.Next() {
						var projID int
						if errPScan := pRows.Scan(&projID); errPScan == nil {
							stringsIDs = append(stringsIDs, strconv.Itoa(projID))
						}
					}
					pRows.Close()
				}
				jsonProjetos := "[" + strings.Join(stringsIDs, ",") + "]"

				listaUsuarios = append(listaUsuarios, UsuarioExibicao{
					Usuario:      u,
					ProjetosJSON: jsonProjetos,
				})

			}
		}
	}

	c.HTML(http.StatusOK, "cadastro_usuario.html", gin.H{
		"Erro":          c.Query("erro"),
		"Sucesso":       c.Query("sucesso"),
		"Projetos":      listaProjetos,
		"ListaUsuarios": listaUsuarios, // Injeta a lista de usuários no HTML
	})
}

// ProcessarCadastroUsuario gerencia de forma inteligente a CRIAÇÃO ou ATUALIZAÇÃO (senha/perfil) de usuários
func ProcessarCadastroUsuario(c *gin.Context) {
	usuarioIDStr := c.PostForm("usuario_id") // Captura se veio um ID do modo de edição
	email := c.PostForm("email")
	senha := c.PostForm("senha")
	perfil := c.PostForm("perfil")

	ctx := c.Request.Context()

	if email == "" {
		c.Redirect(http.StatusSeeOther, "/usuario/novo?erro=E-mail obrigatório.")
		return
	}

	// 🛠️ MODO EDIÇÃO (UPDATE)
	if usuarioIDStr != "" {
		usuarioID, _ := strconv.Atoi(usuarioIDStr)

		if len(strings.TrimSpace(senha)) >= 4 {
			senhaHash, _ := bcrypt.GenerateFromPassword([]byte(senha), bcrypt.DefaultCost)
			query := `UPDATE usuarios SET email = $1, senha = $2, perfil = $3 WHERE id = $4`
			_, err := database.DB.Exec(ctx, query, email, string(senhaHash), perfil, usuarioID)
			if err != nil {
				c.Redirect(http.StatusSeeOther, "/usuario/novo?erro=Erro ao atualizar dados e senha.")
				return
			}
		} else {
			query := `UPDATE usuarios SET email = $1, perfil = $2 WHERE id = $3`
			_, err := database.DB.Exec(ctx, query, email, perfil, usuarioID)
			if err != nil {
				c.Redirect(http.StatusSeeOther, "/usuario/novo?erro=Erro ao atualizar dados do perfil.")
				return
			}
		}

		_, _ = database.DB.Exec(ctx, "DELETE FROM projeto_usuarios WHERE usuario_id = $1", usuarioID)

		if perfil == "consultor" {
			projetosStrArr := c.PostFormArray("projetos_vinculados[]")
			for _, idStr := range projetosStrArr {
				if projID, err := strconv.Atoi(idStr); err == nil {
					queryVinculo := `INSERT INTO projeto_usuarios (projeto_id, usuario_id) VALUES ($1, $2) ON CONFLICT DO NOTHING`
					_, _ = database.DB.Exec(ctx, queryVinculo, projID, usuarioID)
				}
			}
		}

		c.Redirect(http.StatusSeeOther, "/usuario/novo?sucesso=Usuário atualizado com sucesso!")
		return
	}

	// ➕ MODO NOVO CADASTRO (Seu código existente permanece idêntico aqui para baixo)
	if len(senha) < 4 {
		c.Redirect(http.StatusSeeOther, "/usuario/novo?erro=A senha deve ter no minimo 4 caracteres.")
		return
	}

	senhaHash, _ := bcrypt.GenerateFromPassword([]byte(senha), bcrypt.DefaultCost)
	projetosStrArr := c.PostFormArray("projetos_vinculados[]")
	var projetoIDs []int

	if perfil == "consultor" || perfil == "" {
		perfil = "consultor"
		for _, idStr := range projetosStrArr {
			if id, err := strconv.Atoi(idStr); err == nil {
				projetoIDs = append(projetoIDs, id)
			}
		}
	}

	novoUsuario := models.Usuario{
		Email:  email,
		Senha:  string(senhaHash),
		Perfil: perfil,
	}

	err := services.CriarUsuarioCompleto(&novoUsuario, projetoIDs)
	if err != nil {
		c.Redirect(http.StatusSeeOther, "/usuario/novo?erro=Email ja cadastrado no sistema.")
		return
	}

	c.Redirect(http.StatusSeeOther, "/usuario/novo?sucesso=Usuario cadastrado com sucesso!")
}

// ExcluirUsuario remove o acesso de um colaborador permanentemente limpando as chaves estrangeiras
func ExcluirUsuario(c *gin.Context) {
	idStr := c.Query("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID inválido"})
		return
	}

	// Evita que o administrador logado cometa o erro de excluir a si próprio da plataforma
	cookieEmail, _ := c.Cookie("sessao_token")
	var emailDeletando string

	database.DB.QueryRow(c.Request.Context(), "SELECT email FROM usuarios WHERE id = $1", id).Scan(&emailDeletando)

	if cookieEmail == emailDeletando {
		c.JSON(http.StatusForbidden, gin.H{"error": "Você não pode excluir sua própria conta enquanto estiver logado."})
		return
	}

	// Remove o usuário (A tabela pivô projeto_usuarios limpa automaticamente por conta do ON DELETE CASCADE)
	_, err = database.DB.Exec(c.Request.Context(), "DELETE FROM usuarios WHERE id = $1", id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao remover usuário do banco"})
		return
	}

	c.Status(http.StatusOK)
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
