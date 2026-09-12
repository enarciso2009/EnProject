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
		log.Println("Aviso: Erro ao carregar o arquivo .env, usando padrões.")
	}

	// 2. INICIALIZA O BANCO DE DADOS POSTGRES ANTES DAS ROTAS
	database.ConectarDB()

	// 3. CONFIGURAÇÃO INICIAL DO GIN ENGINE
	r := gin.Default()

	// Carrega os arquivos HTML da sua pasta web
	r.LoadHTMLGlob("web/*")

	// Configura o diretório de uploads de fotos de forma correta
	r.Static("/uploads", "./uploads")

	// 4. ROTAS PÚBLICAS (Acessíveis sem login)
	r.GET("/login", handler.ExibirLogin)
	r.POST("/login", handler.ProcessarLogin)

	// 5. ROTAS PROTEGIDAS (Apenas usuários autenticados via middleware)
	protegido := r.Group("/")
	protegido.Use(middleware.AutenticacaoObrigatoria())
	{
		protegido.GET("/", handler.ListarProjetos)
		protegido.GET("/novo", handler.ExibirFormulario)
		protegido.POST("/gerar", handler.ProcessarFormulario)
		protegido.GET("/editar", handler.EditarProjeto)
		protegido.POST("/editar", handler.ProcessarEdicao)

		// Rota assíncrona para registrar o diário de bordo

		protegido.GET("/logout", handler.LogOut)
		protegido.POST("/projetos/diario", handler.RegistrarDiario)

		// Downloads e exportações de relatórios
		protegido.GET("/download/excel", handler.BaixarExcel)
		protegido.GET("/download/html", handler.BaixarHTML)

		// Cadastro de novos logins do sistema
		r.GET("/usuario/novo", handler.ExibirCadastroUsuarios)
		r.POST("/usuario/novo", handler.ProcessarCadastroUsuario)

		// Página de sucesso após operações de POST
		protegido.GET("/sucesso", func(c *gin.Context) {
			idStr := c.Query("id")
			id, _ := strconv.Atoi(idStr)
			proj, _ := services.BuscarProjetoPorID(id)
			c.HTML(http.StatusOK, "sucesso.html", proj)
		})
	}

	// 6. ADQUIRE A PORTA E INICIA O SERVIDOR (O bloqueio final do código)
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	fmt.Println("Servidor rodando com sucesso na porta:", port)
	r.Run(":" + port)
}
