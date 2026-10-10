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

// Função auxiliar para transformar o nome do projeto em um nome de pasta seguro
func gerarNomePastaSeguro(nome string) string {
	nome = strings.ToLower(nome)
	nome = strings.ReplaceAll(nome, " ", "_")
	// Remove qualquer caractere que não seja letra, número ou underline
	reg := regexp.MustCompile(`[^a-z0-9_]`)
	return reg.ReplaceAllString(nome, "")
}

/*
// ListarProjetos busca os registros autorizados do banco e renderiza a página inicial

	func ListarProjetos(c *gin.Context) {
		cookieEmail, err := c.Cookie("sessao_token")
		if err != nil || cookieEmail == "" {
			c.Redirect(http.StatusSeeOther, "/login")
			return
		}

		user, err := services.BuscarUsuarioPorEmail(cookieEmail)
		if err != nil {
			c.Redirect(http.StatusSeeOther, "/login")
			return
		}

		projetos, err := services.ListarTodosProjetos(user.ID, user.Perfil)
		if err != nil {
			c.String(http.StatusInternalServerError, "Erro ao buscar projetos autorizados: %v", err)
			return
		}

		c.HTML(http.StatusOK, "index.html", projetos)
	}
*/
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

	fusoLocal, err := time.LoadLocation("America/Sao_Paulo")
	if err != nil {
		fusoLocal = time.Local
	}

	proj.Nome = getMultipartFieldValue("nome_projeto")
	proj.Gerente = getMultipartFieldValue("gerente")
	proj.StatusGeral = getMultipartFieldValue("status_geral")
	proj.Resumo = getMultipartFieldValue("resumo")
	proj.Observacoes = getMultipartFieldValue("observacoes")
	proj.Introducao = getMultipartFieldValue("introducao")
	proj.Localizacao = getMultipartFieldValue("localizacao")
	proj.Encerramento = getMultipartFieldValue("encerramento")

	// --- LOGICA DE CRIAR A PASTA DINÂMICA ---
	nomePasta := gerarNomePastaSeguro(proj.Nome)
	if nomePasta == "" {
		nomePasta = "projeto_sem_nome"
	}

	// Define o caminho: uploads/nome_do_projeto
	diretorioDestino := filepath.Join("uploads", nomePasta)

	// Cria o diretório no sistema (0755 concede permissões de leitura/escrita padrão)
	if err := os.MkdirAll(diretorioDestino, 0755); err != nil {
		c.String(http.StatusInternalServerError, "Erro ao criar diretório do projeto: %v", err)
		return
	}

	arquivos := form.File["imagens_projeto[]"]
	for _, arquivo := range arquivos {
		nomeArquivo := fmt.Sprintf("%d_%s", time.Now().UnixNano(), arquivo.Filename)

		// Salva o arquivo fisicamente na nova pasta criada
		caminhoSalvar := filepath.Join(diretorioDestino, nomeArquivo)

		if err := c.SaveUploadedFile(arquivo, caminhoSalvar); err == nil {
			// Salva a URL amigável no banco de dados (ex: /uploads/meu_projeto/123_foto.jpg)
			urlBanco := fmt.Sprintf("/uploads/%s/%s", nomePasta, nomeArquivo)
			proj.Imagens = append(proj.Imagens, urlBanco)
		}
	}
	// ----------------------------------------

	// (Restante do mapeamento de tarefas e materiais mantido)
	nomes := form.Value["tarefa_nome[]"]
	responsaveis := form.Value["tarefa_resp[]"]
	inicios := form.Value["tarefa_inicio[]"]
	finais := form.Value["tarefa_fim[]"]
	concluidos := form.Value["tarefa_concluido[]"]
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
		dataIni, _ := time.ParseInLocation("2006-01-02", inicios[i], fusoLocal)
		dataFim, _ := time.ParseInLocation("2006-01-02", finais[i], fusoLocal)

		proj.Tarefas = append(proj.Tarefas, models.Tarefa{
			Nome:        nomes[i],
			Responsavel: responsaveis[i],
			DataInicio:  dataIni,
			DataFim:     dataFim,
			Progresso:   progressoInt,
			Concluido:   isConcluido,
		})
	}

	err = services.SalvarProjetoCompleto(&proj)
	if err != nil {
		c.String(http.StatusInternalServerError, "Erro ao salvar no banco: %v", err)
		return
	}

	c.Redirect(http.StatusSeeOther, "/sucesso?id="+strconv.Itoa(proj.ID))
}

func EditarProjeto(c *gin.Context) {
	idStr := c.Query("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		println("ID inválido ou não fornecido, criando novo projeto.")
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

	fusoLocal, err := time.LoadLocation("America/Sao_Paulo")
	if err != nil {
		fusoLocal = time.Local
	}

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

	// --- REPETE A LÓGICA DE DIRETÓRIO SEGURO NA EDIÇÃO ---
	nomePasta := gerarNomePastaSeguro(proj.Nome)
	if nomePasta == "" {
		nomePasta = "projeto_sem_nome"
	}
	diretorioDestino := filepath.Join("uploads", nomePasta)

	if err := os.MkdirAll(diretorioDestino, 0755); err != nil {
		c.String(http.StatusInternalServerError, "Erro ao criar diretório do projeto: %v", err)
		return
	}

	arquivos := form.File["imagens_projeto[]"]
	for _, arquivo := range arquivos {
		nomeArquivo := fmt.Sprintf("%d_%s", time.Now().UnixNano(), arquivo.Filename)
		caminhoSalvar := filepath.Join(diretorioDestino, nomeArquivo)

		if err := c.SaveUploadedFile(arquivo, caminhoSalvar); err == nil {
			urlBanco := fmt.Sprintf("/uploads/%s/%s", nomePasta, nomeArquivo)
			proj.Imagens = append(proj.Imagens, urlBanco)
		}
	}
	// ----------------------------------------------------

	nomes := form.Value["tarefa_nome[]"]
	responsaveis := form.Value["tarefa_resp[]"]
	inicios := form.Value["tarefa_inicio[]"]
	finais := form.Value["tarefa_fim[]"]
	concluidos := form.Value["tarefa_concluido[]"]
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
		dataIni, _ := time.ParseInLocation("2006-01-02", inicios[i], fusoLocal)
		dataFim, _ := time.ParseInLocation("2006-01-02", finais[i], fusoLocal)

		proj.Tarefas = append(proj.Tarefas, models.Tarefa{
			Nome:        nomes[i],
			Responsavel: responsaveis[i],
			DataInicio:  dataIni,
			DataFim:     dataFim,
			Progresso:   progressoInt,
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

// ExcluirProjeto processa a remoção do projeto do banco e deleta a pasta física
func ExcluirProjeto(c *gin.Context) {
	idStr := c.Query("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID de projeto inválido"})
		return
	}

	// 1. Busca o projeto antes de deletar para saber o nome correto da pasta física
	projeto, err := services.BuscarProjetoPorID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Projeto não localizado no banco"})
		return
	}

	// 2. Remove a pasta física do projeto com todas as imagens salvas
	nomePasta := gerarNomePastaSeguro(projeto.Nome)
	if nomePasta != "" && nomePasta != "projeto_sem_nome" {
		diretorioDestino := filepath.Join("uploads", nomePasta)

		// Remove a pasta e tudo o que estiver dentro dela
		_ = os.RemoveAll(diretorioDestino)
	}

	// 3. Deleta o registro do banco de dados através da camada de serviço
	// Certifique-se de que a função ExcluirProjetoCompleto existe no seu pacote services
	err = services.ExcluirProjetoCompleto(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Erro ao deletar do banco: %v", err)})
		return
	}

	// Retorna sucesso para o JavaScript recarregar a tela
	c.Status(http.StatusOK)
}

type ProjetoComProgresso struct {
	models.Projeto
	ProgressoGeral int
}

func ListarProjetos(c *gin.Context) {
	// 1. Busca todos os projetos do banco de dados (Query blindada com COALESCE)
	projetosOriginais, err := services.ListarTodosOsProjetos()
	if err != nil {
		c.String(http.StatusInternalServerError, "Erro ao listar projetos: %v", err)
		return
	}

	// 2. Calcula a média aritmética do progresso das tarefas de cada projeto
	var projetosExibicao []ProjetoComProgresso
	for _, p := range projetosOriginais {
		somaProgresso := 0
		progressoGeral := 0
		totalTarefas := len(p.Tarefas)

		if totalTarefas > 0 {
			for _, tarefa := range p.Tarefas {
				somaProgresso += tarefa.Progresso
			}
			progressoGeral = somaProgresso / totalTarefas
		}

		projetosExibicao = append(projetosExibicao, ProjetoComProgresso{
			Projeto:        p,
			ProgressoGeral: progressoGeral,
		})
	}

	// 3. Captura o estado de administrador injetado pelo seu middleware de sessão
	// Caso seu middleware salve com outro nome (ex: "role" ou "user"), mude o termo entre aspas abaixo
	val, existe := c.Get("is_admin")
	isAdmin := false
	if existe {
		if b, ok := val.(bool); ok {
			isAdmin = b
		}
	}

	// 4. Envia o pacote completo consolidado para renderização do index.html
	c.HTML(http.StatusOK, "index.html", gin.H{
		"Projetos": projetosExibicao,
		"IsAdmin":  isAdmin,
	})
}
