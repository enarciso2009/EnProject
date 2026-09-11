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

// SalvarProjetoCompleto grava um novo projeto, suas tarefas e imagens no banco de dados
func SalvarProjetoCompleto(p *models.Projeto) error {
	ctx := context.Background()
	tx, err := database.DB.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	// 1. Salva o Projeto incluindo participantes (5 campos)
	queryProj := `INSERT INTO projetos (nome, gerente, resumo, participantes, observacoes, introducao, localizacao, encerramento) VALUES ($1, $2, $3, $4, $5, $6, $7, $8) RETURNING id`
	err = tx.QueryRow(ctx, queryProj, p.Nome, p.Gerente, p.Resumo, p.Participantes, p.Observacoes, p.Introducao, p.Localizacao, p.Encerramento).Scan(&p.ID)
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

	// 3. Salva as Imagens vinculadas
	queryImg := `INSERT INTO projeto_imagens (projeto_id, imagem_url) VALUES ($1, $2)`
	for _, imgURL := range p.Imagens {
		_, err = tx.Exec(ctx, queryImg, p.ID, imgURL)
		if err != nil {
			return err
		}
	}

	// 4. Salva os Materiais vinculados
	queryMat := `INSERT INTO projeto_materiais (projeto_id, item, descricao, quantidade) VALUES ($1, $2, $3, $4)`
	for _, m := range p.Materiais {
		_, err = tx.Exec(ctx, queryMat, p.ID, m.Item, m.Descricao, m.Quantidade)
		if err != nil {
			return err
		}
	}
	return tx.Commit(ctx)

}

// AtualizarProjetoCompleto modifica o cabeçalho e reestrutura as tarefas e imagens do projeto no Postgres
// CORRIGIDO: Removido o erro de digitação "Altualizar" para "Atualizar" para sincronizar com seu handler
func AtualizarProjetoCompleto(p *models.Projeto) error {
	ctx := context.Background()
	tx, err := database.DB.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	// 1. Atualiza dados do cabeçalho incluindo participantes
	queryProj := `UPDATE projetos SET nome = $1, gerente = $2, resumo = $3, participantes = $4, observacoes = $5, introducao = $6, localizacao = $7, encerramento = $8 WHERE id = $9`
	_, err = tx.Exec(ctx, queryProj, p.Nome, p.Gerente, p.Resumo, p.Participantes, p.Observacoes, p.Introducao, p.Localizacao, p.Encerramento, p.ID)
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

	// 4. Remove as imagens antigas para atualização limpa
	_, err = tx.Exec(ctx, `DELETE FROM projeto_imagens WHERE projeto_id = $1`, p.ID)
	if err != nil {
		return err
	}

	// 5. Insere a nova lista de imagens atualizada
	queryImg := `INSERT INTO projeto_imagens (projeto_id, imagem_url) VALUES ($1, $2)`
	for _, imgURL := range p.Imagens {
		_, err = tx.Exec(ctx, queryImg, p.ID, imgURL)
		if err != nil {
			return err
		}
	}

	// 6. Limpa os materias antigos antes de reinserir
	_, err = tx.Exec(ctx, `DELETE FROM projeto_materiais WHERE projeto_id = $1`, p.ID)
	if err != nil {
		return err
	}

	// 7. Insere a nova lista de materiais atualizada
	queryMat := `INSERT INTO projeto_materiais (projeto_id, item, descricao, quantidade) VALUES ($1, $2, $3, $4)`
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

	// CORRIGIDO: Mapeamento de 6 campos de retorno para 6 alvos de Scan sem erros de contagem
	queryProj := `SELECT id, nome, gerente, resumo, participantes, observacoes, COALESCE(introducao, ''), COALESCE(localizacao, ''), COALESCE(encerramento, '') FROM projetos WHERE id = $1`

	err := database.DB.QueryRow(ctx, queryProj, id).Scan(
		&p.ID,            //1
		&p.Nome,          //2
		&p.Gerente,       //3
		&p.Resumo,        //4
		&p.Participantes, //6
		&p.Observacoes,   //7
		&p.Introducao,    //8
		&p.Localizacao,   //9
		&p.Encerramento,  //10

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
		err := database.DB.QueryRow(ctx, `SELECT id, nome, responsavel, data_inicio, data_fim, concluido, progresso FROM tarefas WHERE id = $1`, t.ID).Scan()
		err = rows.Scan(
			&t.ID,
			&t.Nome,
			&t.Responsavel,
			&t.DataInicio,
			&t.DataFim,
			&t.Concluido,
			&t.Progresso,
		)
		if err != nil {
			return &p, err
		}
		p.Tarefas = append(p.Tarefas, t)
		// p.TotalHoras += t.Horas
	}

	// Busca Imagens Gerais do Projeto
	queryImg := `SELECT imagem_url FROM projeto_imagens WHERE projeto_id = $1`
	imgRows, err := database.DB.Query(ctx, queryImg, id)
	if err == nil {
		defer imgRows.Close()
		for imgRows.Next() {
			var imgURL string
			if err := imgRows.Scan(&imgURL); err == nil {
				p.Imagens = append(p.Imagens, imgURL)
			}
		}
	}

	// ADICIONADO: Busca dinâmica dos relatórios diários para o histórico da tela de edição
	queryDiarios := `SELECT id, data, descricao, imagens FROM relatorios_diarios WHERE projeto_id = $1 ORDER BY data ASC`
	diarioRows, err := database.DB.Query(ctx, queryDiarios, id)
	if err == nil {
		defer diarioRows.Close()
		for diarioRows.Next() {
			var r models.RelatorioDiario
			err := diarioRows.Scan(&r.ID, &r.Data, &r.Descricao, &r.Imagens)
			if err == nil {
				r.ProjetoID = id
				p.Relatorios = append(p.Relatorios, r)
			}
		}
	}

	queryMat := `SELECT item, descricao, quantidade FROM projeto_materiais WHERE projeto_id = $1 ORDER BY item ASC`
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

// ListarTodosProjetos busca todos os projetos na base para exibir na tela inicial
func ListarTodosProjetos() ([]models.Projeto, error) {
	ctx := context.Background()

	queryAll := `SELECT id, nome, gerente, resumo, participantes, observacoes, COALESCE(introducao, ''), COALESCE(localizacao, ''), COALESCE(encerramento, '') FROM projetos ORDER BY id DESC`
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
			&p.Participantes,
			&p.Observacoes,
			&p.Introducao,
			&p.Localizacao,
			&p.Encerramento,
		)
		if err != nil {
			return nil, err
		}

		// Carrega as tarefas, imagens e diários individuais para calcular o status correto no dashboard
		projCompleto, _ := BuscarProjetoPorID(p.ID)
		lista = append(lista, *projCompleto)
	}
	return lista, nil
}
