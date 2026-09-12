package services

import (
	"context"
	"time"

	"EnProject/internal/database"
	"EnProject/internal/models"
)

// CalcularStatusTarefa aplica as regras automáticas com base na data atual
func CalcularStatusTarefa(t *models.Tarefa) {
	hoje := time.Now().Truncate(24 * time.Hour)
	inicio := t.DataInicio.Truncate(24 * time.Hour)
	fim := t.DataFim.Truncate(24 * time.Hour)

	if t.Concluido {
		t.Status = "Finalizado"
		return
	}
	if hoje.Before(inicio) {
		t.Status = "Finalizado"
		return
	}
	if (hoje.After(inicio) || hoje.Equal(inicio)) && (hoje.Before(fim) || hoje.Equal(fim)) {
		t.Status = "Em-Andamento"
		return
	}
	t.Status = "Atrasado"
}

// CalcularStatusProjeto define o status do projeto baseado na pior tarefa
func CalcularStatusProjeto(p *models.Projeto) {
	piorStatus := "Finalizado"
	for i := range p.Tarefas {
		CalcularStatusTarefa(&p.Tarefas[i])

		if p.Tarefas[i].Status == "Atrasado" {
			piorStatus = "Atrasado"
		} else if p.Tarefas[i].Status == "Em-Andamento" && piorStatus != "Em-Andamento" {
			piorStatus = "Em-Andamento"
		}
	}
	p.StatusGeral = piorStatus
}

// SalvarProjetoCompleto grava um novo projeto, suas tarefas e imagens usando o array nativo
func SalvarProjetoCompleto(p *models.Projeto) error {
	ctx := context.Background()
	tx, err := database.DB.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	// 1. Salva o Projeto incluindo o array de imagens diretamente na tabela pai (8 campos)
	queryProj := `INSERT INTO projetos (nome, gerente, resumo, observacoes, introducao, localizacao, encerramento, imagens) VALUES ($1, $2, $3, $4, $5, $6, $7, $8) RETURNING id`
	err = tx.QueryRow(ctx, queryProj, p.Nome, p.Gerente, p.Resumo, p.Observacoes, p.Introducao, p.Localizacao, p.Encerramento, p.Imagens).Scan(&p.ID)
	if err != nil {
		return err
	}

	// 2. Salva as Tarefas vinculadas
	queryTarefa := `INSERT INTO tarefas (projeto_id, nome, responsavel, data_inicio, data_fim, concluido, progresso) VALUES ($1, $2, $3, $4, $5, $6, $7)`
	for _, t := range p.Tarefas {
		_, err = tx.Exec(ctx, queryTarefa, p.ID, t.Nome, t.Responsavel, t.DataInicio, t.DataFim, t.Concluido, t.Progresso)
		if err != nil {
			return err
		}
	}

	// 3. Salva os Materiais vinculados
	queryMat := `INSERT INTO materiais (projeto_id, item, descricao, quantidade) VALUES ($1, $2, $3, $4)`
	for _, m := range p.Materiais {
		_, err = tx.Exec(ctx, queryMat, p.ID, m.Item, m.Descricao, m.Quantidade)
		if err != nil {
			return err
		}
	}
	return tx.Commit(ctx)
}

// AtualizarProjetoCompleto modifica o cabeçalho e reestrutura as tarefas e imagens do projeto no Postgres
func AtualizarProjetoCompleto(p *models.Projeto) error {
	ctx := context.Background()
	tx, err := database.DB.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	// 1. Atualiza dados do cabeçalho incluindo o array de imagens diretamente
	queryProj := `UPDATE projetos SET nome = $1, gerente = $2, resumo = $3, observacoes = $4, introducao = $5, localizacao = $6, encerramento = $7, imagens = $8 WHERE id = $9`
	_, err = tx.Exec(ctx, queryProj, p.Nome, p.Gerente, p.Resumo, p.Observacoes, p.Introducao, p.Localizacao, p.Encerramento, p.Imagens, p.ID)
	if err != nil {
		return err
	}

	// 2. Remove tarefas antigas para atualização limpa
	_, err = tx.Exec(ctx, `DELETE FROM tarefas WHERE projeto_id = $1`, p.ID)
	if err != nil {
		return err
	}

	// 3. Insere a lista nova/atualizada de tarefas com o flag concluido
	queryTarefa := `INSERT INTO tarefas (projeto_id, nome, responsavel, data_inicio, data_fim, concluido, progresso) VALUES ($1, $2, $3, $4, $5, $6, $7)`
	for _, t := range p.Tarefas {
		_, err = tx.Exec(ctx, queryTarefa, p.ID, t.Nome, t.Responsavel, t.DataInicio, t.DataFim, t.Concluido, t.Progresso)
		if err != nil {
			return err
		}
	}

	// 4. Limpa os materiais antigos antes de reinserir (usando o nome correto da tabela: materiais)
	_, err = tx.Exec(ctx, `DELETE FROM materiais WHERE projeto_id = $1`, p.ID)
	if err != nil {
		return err
	}

	// 5. Insere a nova lista de materiais atualizada
	queryMat := `INSERT INTO materiais (projeto_id, item, descricao, quantidade) VALUES ($1, $2, $3, $4)`
	for _, m := range p.Materiais {
		_, err = tx.Exec(ctx, queryMat, p.ID, m.Item, m.Descricao, m.Quantidade)
		if err != nil {
			return err
		}
	}
	return tx.Commit(ctx)
}

// BuscarProjetoPorID traz um projeto específico, suas tarefas, imagens, relatórios diários e calcula seu status dinâmico
func BuscarProjetoPorID(id int) (*models.Projeto, error) {
	ctx := context.Background()
	var p models.Projeto

	// Buscando o array de imagens diretamente da tabela projetos
	queryProj := `SELECT id, nome, gerente, resumo, observacoes, COALESCE(introducao, ''), COALESCE(localizacao, ''), COALESCE(encerramento, ''), imagens FROM projetos WHERE id = $1`

	err := database.DB.QueryRow(ctx, queryProj, id).Scan(
		&p.ID,
		&p.Nome,
		&p.Gerente,
		&p.Resumo,
		&p.Observacoes,
		&p.Introducao,
		&p.Localizacao,
		&p.Encerramento,
		&p.Imagens, // O pgx faz o Scan automático de TEXT[] do Postgres para []string do Go 🐘
	)
	if err != nil {
		return &p, err
	}

	// Busca Tarefas
	queryTar := `SELECT id, nome, responsavel, data_inicio, data_fim, concluido, progresso FROM tarefas WHERE projeto_id = $1`
	rows, err := database.DB.Query(ctx, queryTar, id)
	if err != nil {
		return &p, err
	}
	defer rows.Close()

	for rows.Next() {
		var t models.Tarefa
		err = rows.Scan(&t.ID, &t.Nome, &t.Responsavel, &t.DataInicio, &t.DataFim, &t.Concluido, &t.Progresso)
		if err != nil {
			return &p, err
		}
		p.Tarefas = append(p.Tarefas, t)
	}

	// Busca dinâmica dos relatórios diários para o histórico da tela de edição (Incluindo pendencias e participantes)
	queryDiarios := `SELECT id, data, descricao, imagens, COALESCE(participantes, ''), COALESCE(pendencias, '') FROM relatorios_diarios WHERE projeto_id = $1 ORDER BY data ASC`
	diarioRows, err := database.DB.Query(ctx, queryDiarios, id)
	if err == nil {
		defer diarioRows.Close()
		for diarioRows.Next() {
			var r models.RelatorioDiario
			err := diarioRows.Scan(&r.ID, &r.Data, &r.Descricao, &r.Imagens, &r.Participantes, &r.Pendencias)
			if err == nil {
				r.ProjetoID = id
				p.Relatorios = append(p.Relatorios, r)
			}
		}
	}

	// Busca Materiais usando a tabela correta 'materiais'
	queryMat := `SELECT item, descricao, quantidade FROM materiais WHERE projeto_id = $1 ORDER BY item ASC`
	matRows, err := database.DB.Query(ctx, queryMat, id)
	if err == nil {
		defer matRows.Close()
		for matRows.Next() {
			var m models.Material
			if err := matRows.Scan(&m.Item, &m.Descricao, &m.Quantidade); err == nil {
				p.Materiais = append(p.Materiais, m)
			}
		}
	}

	CalcularStatusProjeto(&p)
	return &p, nil
}

// ListarTodosProjetos otimizado para evitar loops pesados (N+1)
func ListarTodosProjetos() ([]models.Projeto, error) {
	ctx := context.Background()

	// Carrega os dados básicos necessários para renderizar o painel inicial (index.html)
	queryAll := `SELECT id, nome, gerente, resumo, observacoes, COALESCE(introducao, ''), COALESCE(localizacao, ''), COALESCE(encerramento, '') FROM projetos ORDER BY id DESC`
	rows, err := database.DB.Query(ctx, queryAll)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var lista []models.Projeto
	for rows.Next() {
		var p models.Projeto
		err := rows.Scan(
			&p.ID,
			&p.Nome,
			&p.Gerente,
			&p.Resumo,
			&p.Observacoes,
			&p.Introducao,
			&p.Localizacao,
			&p.Encerramento,
		)
		if err != nil {
			return nil, err
		}

		// Para calcular o StatusGeral ("Em-Andamento", "Atrasado") na Home,
		// buscamos apenas as tarefas associadas de forma leve, em vez de carregar o projeto inteiro com imagens pesadas.
		queryTar := `SELECT id, data_inicio, data_fim, concluido FROM tarefas WHERE projeto_id = $1`
		tRows, err := database.DB.Query(ctx, queryTar, p.ID)
		if err == nil {
			for tRows.Next() {
				var t models.Tarefa
				if err := tRows.Scan(&t.ID, &t.DataInicio, &t.DataFim, &t.Concluido); err == nil {
					p.Tarefas = append(p.Tarefas, t)
				}
			}
			tRows.Close()
		}

		// Calcula dinamicamente se o projeto está atrasado ou finalizado para a cor da tag no HTML
		CalcularStatusProjeto(&p)

		lista = append(lista, p)
	}
	return lista, nil
}
