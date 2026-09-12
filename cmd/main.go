package main

import (
	"EnProject/internal/database"
	"EnProject/internal/handler"
	"EnProject/internal/middleware"
	"EnProject/internal/services"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	// 1. CARREGA AS VARIÁVEIS DE AMBIENTE DO .ENV PRIMEIRO
	err := godotenv.Load()
	if err != nil {
		log.Println("Aviso: Erro ao carregar o arquivo .env, usando padrões locais.")
	}

	// 2. INICIALIZA O BANCO DE DADOS POSTGRES (Criando as tabelas e colunas novas automaticamente)
	database.ConectarDB()

	// 3. CONFIGURAÇÃO INICIAL DO GIN ENGINE
	r := gin.Default()

	// Carrega os arquivos HTML da sua pasta web
	r.LoadHTMLGlob("web/*")

	// Configura o diretório público de uploads de fotos de forma correta
	r.Static("/uploads", "./uploads")

	// -------------------------------------------------------------------------
	// 4. ROTAS PÚBLICAS (Acessíveis sem qualquer tipo de login)
	// -------------------------------------------------------------------------
	r.GET("/login", handler.ExibirLogin)
	r.POST("/login", handler.ProcessarLogin)
	r.GET("/logout", handler.LogOut) // Limpa o cookie de sessão do navegador

	// -------------------------------------------------------------------------
	// 5. ROTAS PROTEGIDAS - COMUM (Acessíveis por Administradores e Consultores)
	// -------------------------------------------------------------------------
	comum := r.Group("/")
	comum.Use(middleware.AutenticacaoObrigatoria())
	{
		// Lista os projetos na Home (Filtrado automaticamente se o usuário for consultor)
		comum.GET("/", handler.ListarProjetos)

		// Visualização e download do Relatório Executivo consolidado
		comum.GET("/download/html", handler.BaixarHTML)

		// Página de sucesso após operações de POST bem-sucedidas
		comum.GET("/sucesso", func(c *gin.Context) {
			idStr := c.Query("id")
			id, _ := strconv.Atoi(idStr)
			proj, _ := services.BuscarProjetoPorID(id)
			c.HTML(http.StatusOK, "sucesso.html", proj)
		})
	}

	// -------------------------------------------------------------------------
	// 6. ROTAS EXCLUSIVAS - ADMINISTRADORES (Criação, Edição, Diário e Usuários)
	// -------------------------------------------------------------------------
	admin := r.Group("/")
	admin.Use(middleware.AutenticacaoObrigatoria(), middleware.RequererAdmin())
	{
		// Fluxo de criação de novos Projetos
		admin.GET("/novo", handler.ExibirFormulario)
		admin.POST("/gerar", handler.ProcessarFormulario)

		// Fluxo de edição e atualização de Projetos e Materiais
		admin.GET("/editar", handler.EditarProjeto)
		admin.POST("/editar", handler.ProcessarEdicao)

		// Rota assíncrona (AJAX) para registrar um novo relatório no Diário de Bordo
		// Nota: Caso o nome da sua função seja diferente no projeto, altere após o ponto.
		admin.POST("/projetos/diario", handler.ProcessarCadastroUsuario)

		// Cadastro de novos acessos no sistema com caixas de seleção de projetos
		admin.GET("/usuario/novo", handler.ExibirCadastroUsuarios)
		admin.POST("/usuario/novo", handler.ProcessarCadastroUsuario)
	}

	// -------------------------------------------------------------------------
	// 7. ADQUIRE A PORTA E INICIA O SERVIDOR (Bloqueio final de execução)
	// -------------------------------------------------------------------------
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	fmt.Println("🚀 EnProject rodando com sucesso na porta:", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("Erro crítico ao rodar o servidor Gin: %v", err)
	}
}
