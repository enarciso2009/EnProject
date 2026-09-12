package handler

import (
	"net/http"
	"strconv"
	"time"

	"EnProject/internal/models"
	"EnProject/internal/services"

	"github.com/gin-gonic/gin"
)

// ListarProjetos busca os registros autorizados do banco e renderiza a página inicial
func ListarProjetos(c *gin.Context) {
	// 1. Recupera o e-mail do usuário logado através do cookie de sessão
	cookieEmail, err := c.Cookie("sessao_token")
	if err != nil || cookieEmail == "" {
		c.Redirect(http.StatusSeeOther, "/login")
		return
	}

	// 2. Busca o ID e o Perfil do usuário logado
	user, err := services.BuscarUsuarioPorEmail(cookieEmail)
	if err != nil {
		c.Redirect(http.StatusSeeOther, "/login")
		return
	}

	// 3. Puxa os projetos passando as credenciais do usuário para o filtro SQL
	projetos, err := services.ListarTodosProjetos(user.ID, user.Perfil)
	if err != nil {
		c.String(http.StatusInternalServerError, "Erro ao buscar projetos autorizados: %v", err)
		return
	}

	// Envia a lista filtrada para o index.html
	c.HTML(http.StatusOK, "index.html", projetos)
}

// ExibirFormulario agora serve apenas para exibir a tela de cadastro vazia
func ExibirFormulario(c *gin.Context) {
	c.HTML(http.StatusOK, "form.html", nil)
}

func ProcessarFormulario(c *gin.Context) {
	form, err := c.MultipartForm()
	if err != nil {
		c.String(http.StatusBadRequest, "Erro ao processar formulário multipart: %v", err)
		return
	}

	var proj models.Projeto

	getMultipartFieldValue := func(key string) string {
		if val, ok := form.Value[key]; ok && len(val) > 0 {
			return val[0]
		}
		return ""
	}

	proj.Nome = getMultipartFieldValue("nome_projeto")
	proj.Gerente = getMultipartFieldValue("gerente")
	proj.StatusGeral = getMultipartFieldValue("status_geral")
	proj.Resumo = getMultipartFieldValue("resumo")
	proj.Observacoes = getMultipartFieldValue("observacoes")
	proj.Introducao = getMultipartFieldValue("introducao")
	proj.Localizacao = getMultipartFieldValue("localizacao")
	proj.Encerramento = getMultipartFieldValue("encerramento")

	nomes := form.Value["tarefa_nome[]"]
	responsaveis := form.Value["tarefa_resp[]"]
	inicios := form.Value["tarefa_inicio[]"]
	finais := form.Value["tarefa_fim[]"]
	concluidos := form.Value["tarefa_concluido[]"] // ALINHADO: Sem o "a"
	progressoArr := form.Value["tarefa_progresso[]"]

	matItens := form.Value["mat_item[]"]
	matDescs := form.Value["mat_descricao[]"]
	matQtds := form.Value["mat_quantidade[]"]

	for i := 0; i < len(matDescs); i++ {
		if matDescs[i] == "" {
			continue
		}
		itemNum, _ := strconv.Atoi(matItens[i])
		qtdNum, _ := strconv.Atoi(matQtds[i])
		proj.Materiais = append(proj.Materiais, models.Material{
			Item:       itemNum,
			Descricao:  matDescs[i],
			Quantidade: qtdNum,
		})
	}

	for i := 0; i < len(nomes); i++ {
		if nomes[i] == "" {
			continue
		}

		isConcluido := false
		if i < len(concluidos) && concluidos[i] == "true" {
			isConcluido = true
		}

		progressoInt := 0
		if i < len(progressoArr) {
			progressoInt, _ = strconv.Atoi(progressoArr[i])
		}

		dataIni, _ := time.Parse("2006-01-02", inicios[i])
		dataFim, _ := time.Parse("2006-01-02", finais[i])

		proj.Tarefas = append(proj.Tarefas, models.Tarefa{
			Nome:        nomes[i],
			Responsavel: responsaveis[i],
			DataInicio:  dataIni,
			DataFim:     dataFim,
			Progresso:   progressoInt,
			Concluido:   isConcluido,
		})
	}

	arquivos := form.File["imagens_projeto[]"]
	for _, arquivo := range arquivos {
		nomeArquivo := strconv.FormatInt(time.Now().UnixNano(), 10) + "_" + arquivo.Filename
		caminhoSalvar := "uploads/" + nomeArquivo

		if err := c.SaveUploadedFile(arquivo, caminhoSalvar); err == nil {
			proj.Imagens = append(proj.Imagens, "/uploads/"+nomeArquivo)
		}
	}

	err = services.SalvarProjetoCompleto(&proj)
	if err != nil {
		c.String(http.StatusInternalServerError, "Erro ao salvar no banco: %v", err)
		return
	}

	c.Redirect(http.StatusSeeOther, "/sucesso?id="+strconv.Itoa(proj.ID))
}

// EditarProjeto busca os dados do projeto e abre o formulário pré-preenchido
func EditarProjeto(c *gin.Context) {
	idStr := c.Query("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.String(http.StatusBadRequest, "ID de projeto inválido")
		return
	}

	projeto, err := services.BuscarProjetoPorID(id)
	if err != nil {
		c.String(http.StatusNotFound, "Projeto não localizado no banco")
		return
	}

	c.HTML(http.StatusOK, "editar.html", projeto)
}

// ProcessarEdicao captura o formulário alterado e atualiza o banco de dados
func ProcessarEdicao(c *gin.Context) {
	form, err := c.MultipartForm()
	if err != nil {
		c.String(http.StatusBadRequest, "Erro ao processar formulário multipart: %v", err)
		return
	}

	getMultipartFieldValue := func(key string) string {
		if val, ok := form.Value[key]; ok && len(val) > 0 {
			return val[0]
		}
		return ""
	}

	idStr := getMultipartFieldValue("id")
	id, _ := strconv.Atoi(idStr)

	var proj models.Projeto
	proj.ID = id
	proj.Nome = getMultipartFieldValue("nome_projeto")
	proj.Gerente = getMultipartFieldValue("gerente")
	proj.StatusGeral = getMultipartFieldValue("status_geral")
	proj.Resumo = getMultipartFieldValue("resumo")
	proj.Observacoes = getMultipartFieldValue("observacoes")
	proj.Introducao = getMultipartFieldValue("introducao")
	proj.Localizacao = getMultipartFieldValue("localizacao")
	proj.Encerramento = getMultipartFieldValue("encerramento")

	fotosRestantes := form.Value["fotos_existentes[]"]
	for _, foto := range fotosRestantes {
		if foto != "" {
			proj.Imagens = append(proj.Imagens, foto)
		}
	}

	arquivos := form.File["imagens_projeto[]"]
	for _, arquivo := range arquivos {
		nomeArquivo := strconv.FormatInt(time.Now().UnixNano(), 10) + "_" + arquivo.Filename
		caminhoSalvar := "uploads/" + nomeArquivo

		if err := c.SaveUploadedFile(arquivo, caminhoSalvar); err == nil {
			proj.Imagens = append(proj.Imagens, "/uploads/"+nomeArquivo)
		}
	}

	// 3. Captura os dados das tarefas, INCLUINDO o array de concluídos e de progresso
	nomes := form.Value["tarefa_nome[]"]
	responsaveis := form.Value["tarefa_resp[]"]
	inicios := form.Value["tarefa_inicio[]"]
	finais := form.Value["tarefa_fim[]"]
	concluidos := form.Value["tarefa_concluido[]"]   // CORRIGIDO: Nome igual ao do html ("tarefa_concluido[]")
	progressoArr := form.Value["tarefa_progresso[]"] // ADICIONADO: Captura o progresso na edição
	matItens := form.Value["mat_item[]"]
	matDescs := form.Value["mat_descricao[]"]
	matQtds := form.Value["mat_quantidade[]"]

	for i := 0; i < len(matDescs); i++ {
		if matDescs[i] == "" {
			continue
		}
		itemNum, _ := strconv.Atoi(matItens[i])
		qtdNum, _ := strconv.Atoi(matQtds[i])
		proj.Materiais = append(proj.Materiais, models.Material{
			Item:       itemNum,
			Descricao:  matDescs[i],
			Quantidade: qtdNum,
		})
	}

	for i := 0; i < len(nomes); i++ {
		if nomes[i] == "" {
			continue
		}

		isConcluido := false
		if i < len(concluidos) && concluidos[i] == "true" {
			isConcluido = true
		}

		progressoInt := 0
		if i < len(progressoArr) {
			progressoInt, _ = strconv.Atoi(progressoArr[i])
		}

		dataIni, _ := time.Parse("2006-01-02", inicios[i])
		dataFim, _ := time.Parse("2006-01-02", finais[i])

		proj.Tarefas = append(proj.Tarefas, models.Tarefa{
			Nome:        nomes[i],
			Responsavel: responsaveis[i],
			DataInicio:  dataIni,
			DataFim:     dataFim,
			Progresso:   progressoInt, // CORRIGIDO: Passa o progresso capturado para a struct
			Concluido:   isConcluido,
		})
	}

	err = services.AtualizarProjetoCompleto(&proj)
	if err != nil {
		c.String(http.StatusInternalServerError, "Erro ao atualizar dados: %v", err)
		return
	}

	c.Redirect(http.StatusSeeOther, "/sucesso?id="+strconv.Itoa(proj.ID))
}
