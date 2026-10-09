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

	// CAPTURA DO ID DO DIÁRIO (Se vier preenchido pelo JavaScript, significa que é uma EDIÇÃO)
	diarioIDStr := getMultipartFieldValue("diario_id")
	diarioID, _ := strconv.Atoi(diarioIDStr)

	if projetoID == 0 || dataStr == "" || descricao == "" || participantes == "" {
		println("[Erro Diario] Campos obrigatórios ausentes. ID recebido:", projetoID)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Campos obrigatórios ausentes"})
		return
	}

	// 1. Carrega o fuso horário oficial do Brasil
	fusoLocal, err := time.LoadLocation("America/Sao_Paulo")
	if err != nil {
		// Caso o servidor não possua a tabela de fusos, adota o fuso local da máquina
		fusoLocal = time.Local
	}

	// 2. Faz o parse da data travando-a no fuso horário correto (Substitui o time.Parse antigo)
	dataRelatorio, err := time.ParseInLocation("2006-01-02", dataStr, fusoLocal)
	if err != nil {
		println("[Erro Diario] Formato de data inválido:", dataStr)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Formato de data inválido"})
		return
	}

	// Busca o nome do projeto no banco de dados
	projeto, err := services.BuscarProjetoPorID(projetoID)
	if err != nil {
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
	diario.ID = diarioID // Vincula o ID caso seja edição
	diario.ProjetoID = projetoID
	diario.Data = dataRelatorio
	diario.Descricao = descricao
	diario.Participantes = participantes
	diario.Pendencias = pendencias
	diario.Veiculos = veiculos

	// Se for EDIÇÃO, recupera as imagens que já estavam salvas para não perdê-las ao enviar novas
	if diarioID > 0 {
		// Buscamos o projeto completo para achar as imagens do relatório específico
		projAtual, err := services.BuscarProjetoPorID(projetoID)
		if err == nil {
			for _, r := range projAtual.Relatorios {
				if r.ID == diarioID {
					diario.Imagens = r.Imagens
					break
				}
			}
		}
	}

	// Processa e armazena os novos arquivos físicos de imagens (se houver)
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

	// EXECUÇÃO DO BANCO DE DADOS: Escolhe dinamicamente entre INSERT ou UPDATE
	if diario.ID > 0 {
		// --- MODO EDIÇÃO ---
		err = services.AtualizarRelatorioDiario(&diario)
		if err != nil {
			println("[Erro Diario] Erro ao AtualizarRelatorioDiario no Banco:", err.Error())
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao atualizar relatório no banco"})
			return
		}
	} else {
		// --- MODO NOVO CADASTRO ---
		err = services.SalvarRelatorioDiario(&diario)
		if err != nil {
			println("[Erro Diario] Erro ao SalvarRelatorioDiario no Banco:", err.Error())
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao salvar relatório no banco"})
			return
		}
	}

	// Retorna o sucesso esperado
	c.JSON(http.StatusOK, gin.H{
		"id":      diario.ID,
		"imagens": diario.Imagens,
		"status":  "processado",
	})
}

// ExcluirRelatorio remove o relatório diário do banco e limpa seus anexos físicos
func ExcluirRelatorio(c *gin.Context) {
	idStr := c.Query("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID do relatório inválido"})
		return
	}

	// 1. Busca os caminhos das imagens no banco antes de deletar o registro
	imagens, err := services.BuscarImagensRelatorio(id)
	if err == nil {
		// Varre o array de strings contendo as URLs (/uploads/projeto/fotos/imagem.jpg)
		for _, urlImagem := range imagens {
			if urlImagem != "" {
				// Remove a barra inicial '/' para virar um caminho relativo válido no Windows/Linux (uploads/...)
				caminhoFisico := strings.TrimPrefix(urlImagem, "/")

				// Apaga o arquivo físico do disco
				if err := os.Remove(caminhoFisico); err != nil {
					println("[Aviso Disco] Não foi possível apagar a imagem:", caminhoFisico, "Erro:", err.Error())
				} else {
					println("[Disco] Imagem apagada com sucesso:", caminhoFisico)
				}
			}
		}
	} else {
		println("[Aviso Banco] Falha ao listar imagens para exclusão física:", err.Error())
	}

	// 2. Chama a camada de serviço para deletar o registro do banco de dados
	err = services.ExcluirRelatorioDiario(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Erro ao deletar do banco: %v", err)})
		return
	}

	c.Status(http.StatusOK)
}
