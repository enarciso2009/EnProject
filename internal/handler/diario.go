package handler

import (
	"net/http"
	"strconv"
	"time"

	"EnProject/internal/models"
	"EnProject/internal/services"

	"github.com/gin-gonic/gin"
)

// RegistrarDiario captura o envio assíncrono do diário de bordo via AJAX/Fetch API
func RegistrarDiario(c *gin.Context) {
	form, err := c.MultipartForm()
	if err != nil {
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

	if projetoID == 0 || dataStr == "" || descricao == "" || participantes == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Campos obrigatórios ausentes"})
		return
	}

	dataRelatorio, err := time.Parse("2006-01-02", dataStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Formato de data inválido"})
		return
	}

	var diario models.RelatorioDiario
	diario.ProjetoID = projetoID
	diario.Data = dataRelatorio
	diario.Descricao = descricao
	diario.Participantes = participantes
	diario.Pendencias = pendencias

	// Processa e armazena os arquivos físicos de imagens específicos deste diário
	arquivos := form.File["imagens_diario[]"]
	for _, arquivo := range arquivos {
		nomeArquivo := strconv.FormatInt(time.Now().UnixNano(), 10) + "_" + arquivo.Filename
		caminhoSalvar := "uploads/" + nomeArquivo // Salva no diretório físico raiz

		if err := c.SaveUploadedFile(arquivo, caminhoSalvar); err == nil {
			// Adiciona o caminho web relativo na struct
			diario.Imagens = append(diario.Imagens, "/uploads/"+nomeArquivo)
		}
	}

	// Envia para a camada de serviços persistir no banco de dados
	err = services.SalvarRelatorioDiario(&diario)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao salvar relatório no banco"})
		return
	}

	// Retorna estritamente o formato JSON esperado pelo JavaScript na tela
	c.JSON(http.StatusCreated, gin.H{
		"id":      diario.ID,
		"imagens": diario.Imagens,
	})
}
