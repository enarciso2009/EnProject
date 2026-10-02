package handler

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"EnProject/internal/models"
	"EnProject/internal/services"

	"github.com/gin-gonic/gin"
)

// Opcão 2: Nome alterado para evitar conflito com o outro arquivo
func gerarNomePastaDiarioSeguro(nome string) string {
	nome = strings.ToLower(nome)
	nome = strings.ReplaceAll(nome, " ", "_")
	reg := regexp.MustCompile(`[^a-z0-9_]`)
	return reg.ReplaceAllString(nome, "")
}

// RegistrarDiario captura o envio assíncrono do diário de bordo via AJAX/Fetch API
// RegistrarDiario captura o envio assíncrono do diário de bordo via AJAX/Fetch API
func RegistrarDiario(c *gin.Context) {
	form, err := c.MultipartForm()
	if err != nil {
		println("[Erro Diario] Falha ao ler MultipartForm:", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": "Erro ao processar formulário multipart"})
		return
	}

	getMultipartFieldValue := func(key string) string {
		if val, ok := form.Value[key]; ok && len(val) > 0 {
			return val[0]
		}
		return ""
	}

	// Captura os dados essenciais
	idStr := getMultipartFieldValue("projeto_id")
	projetoID, _ := strconv.Atoi(idStr)
	dataStr := getMultipartFieldValue("data")
	descricao := getMultipartFieldValue("descricao")
	participantes := getMultipartFieldValue("participantes")
	pendencias := getMultipartFieldValue("pendencias")
	veiculos := getMultipartFieldValue("veiculos")

	if projetoID == 0 || dataStr == "" || descricao == "" || participantes == "" {
		println("[Erro Diario] Campos obrigatórios ausentes. ID recebido:", projetoID)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Campos obrigatórios ausentes"})
		return
	}

	dataRelatorio, err := time.Parse("2006-01-02", dataStr)
	if err != nil {
		println("[Erro Diario] Formato de data inválido:", dataStr)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Formato de data inválido"})
		return
	}

	// Busca o nome do projeto no banco de dados
	projeto, err := services.BuscarProjetoPorID(projetoID)
	if err != nil {
		// SE O ERRO ACONTECER AQUI, VAI APARECER NO SEU TERMINAL AGORA:
		println("[Erro Diario] Erro ao buscar projeto com ID", projetoID, ":", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Projeto relacionado não encontrado no banco"})
		return
	}

	nomePasta := gerarNomePastaDiarioSeguro(projeto.Nome)
	if nomePasta == "" {
		nomePasta = "projeto_sem_nome"
	}

	// Define o caminho físico: uploads/nome_do_projeto/fotos
	diretorioDestino := filepath.Join("uploads", nomePasta, "fotos")

	// Cria a árvore completa de pastas
	if err := os.MkdirAll(diretorioDestino, 0755); err != nil {
		println("[Erro Diario] Erro ao criar diretório:", diretorioDestino, "Erro:", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao criar diretório para o diário"})
		return
	}

	var diario models.RelatorioDiario
	diario.ProjetoID = projetoID
	diario.Data = dataRelatorio
	diario.Descricao = descricao
	diario.Participantes = participantes
	diario.Pendencias = pendencias
	diario.Veiculos = veiculos

	// Processa e armazena os arquivos físicos de imagens
	arquivos := form.File["imagens_diario[]"]
	for _, arquivo := range arquivos {
		nomeArquivo := fmt.Sprintf("%d_%s", time.Now().UnixNano(), arquivo.Filename)
		caminhoSalvar := filepath.Join(diretorioDestino, nomeArquivo)

		if err := c.SaveUploadedFile(arquivo, caminhoSalvar); err != nil {
			println("[Erro Diario] Falha ao salvar arquivo em disco:", err.Error())
		} else {
			urlBanco := fmt.Sprintf("/uploads/%s/fotos/%s", nomePasta, nomeArquivo)
			diario.Imagens = append(diario.Imagens, urlBanco)
		}
	}

	// Envia para a camada de serviços persistir no banco de dados
	err = services.SalvarRelatorioDiario(&diario)
	if err != nil {
		// SE O ERRO FOR NO BANCO DE DADOS, VAI APARECER AQUI:
		println("[Erro Diario] Erro ao SalvarRelatorioDiario no Banco:", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao salvar relatório no banco"})
		return
	}

	// Retorna o sucesso esperado
	c.JSON(http.StatusCreated, gin.H{
		"id":      diario.ID,
		"imagens": diario.Imagens,
	})
}
