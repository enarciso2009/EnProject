package handler

import (
	"net/http"
	"strconv"

	"EnProject/internal/services"

	"github.com/gin-gonic/gin"
	"github.com/xuri/excelize/v2"
)

// BaixarHTML busca o projeto no Postgres pelo ID da URL e renderiza o template limpo
func BaixarHTML(c *gin.Context) {
	idStr := c.Query("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.String(http.StatusBadRequest, "ID do projeto inválido")
		return
	}

	// Busca os dados reais salvos no banco de dados
	projeto, err := services.BuscarProjetoPorID(id)
	if err != nil {
		c.String(http.StatusNotFound, "Projeto não encontrado no banco: %v", err)
		return
	}

	c.HTML(http.StatusOK, "report.html", projeto)
}

// BaixarExcel monta a planilha Excel em tempo real com os dados do banco
func BaixarExcel(c *gin.Context) {
	idStr := c.Query("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.String(http.StatusBadRequest, "ID do projeto inválido")
		return
	}

	projeto, err := services.BuscarProjetoPorID(id)
	if err != nil {
		c.String(http.StatusNotFound, "Projeto não encontrado: %v", err)
		return
	}

	f := excelize.NewFile()
	defer f.Close()

	sheet := "Acompanhamento"
	f.SetSheetName("Sheet1", sheet)

	// Injeta os dados do cabeçalho do projeto
	f.SetCellValue(sheet, "A1", "PROJETO:")
	f.SetCellValue(sheet, "B1", projeto.Nome)
	f.SetCellValue(sheet, "A2", "GERENTE:")
	f.SetCellValue(sheet, "B2", projeto.Gerente)
	f.SetCellValue(sheet, "A3", "STATUS GERAL:")
	f.SetCellValue(sheet, "B3", projeto.StatusGeral)
	f.SetCellValue(sheet, "A4", "TOTAL HORAS:")
	f.SetCellValue(sheet, "B4", projeto.TotalHoras)

	// Monta o cabeçalho da tabela de tarefas
	f.SetCellValue(sheet, "A6", "Nome da Tarefa")
	f.SetCellValue(sheet, "B6", "Responsável")
	f.SetCellValue(sheet, "C6", "Data Início")
	f.SetCellValue(sheet, "D6", "Data Fim")
	f.SetCellValue(sheet, "E6", "Horas Dedicadas")

	// Preenche as linhas dinamicamente vindo do array de structs
	for idx, t := range projeto.Tarefas {
		row := strconv.Itoa(7 + idx)
		f.SetCellValue(sheet, "A"+row, t.Nome)
		f.SetCellValue(sheet, "B"+row, t.Responsavel)
		f.SetCellValue(sheet, "C"+row, t.DataInicio.Format("02/01/2006"))
		f.SetCellValue(sheet, "D"+row, t.DataFim.Format("02/01/2006"))
		//f.SetCellValue(sheet, "E"+row, t.Horas)
	}

	// Força o navegador a entender a resposta HTTP como um arquivo de download
	c.Header("Content-Disposition", "attachment; filename=status_projeto_"+idStr+".xlsx")
	c.Header("Content-Type", "application/octet-stream")
	f.Write(c.Writer)
}
