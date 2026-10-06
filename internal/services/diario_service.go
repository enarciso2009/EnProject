package services

import (
	"EnProject/internal/database"
	"EnProject/internal/models"
	"context"
)

// SalvarRelatorioDiario persiste o diário de bordo diretamente no PostgreSQL usando pgx/v5
func SalvarRelatorioDiario(diario *models.RelatorioDiario) error {
	ctx := context.Background()

	query := `
		INSERT INTO relatorios_diarios (projeto_id, data, descricao, imagens, participantes, pendencias, veiculos) 
		VALUES ($1, $2, $3, $4, $5, $6, $7) 
		RETURNING id
	`

	// O pgx resolve o slice de string ([]string) mapeando nativamente para o text[] do PostgreSQL
	err := database.DB.QueryRow(ctx, query,
		diario.ProjetoID,
		diario.Data,
		diario.Descricao,
		diario.Imagens,
		diario.Participantes,
		diario.Pendencias,
		diario.Veiculos,
	).Scan(&diario.ID)

	if err != nil {
		return err
	}

	return nil
}

// ExcluirRelatorioDiario deleta um registro de histórico da tabela correspondente
func ExcluirRelatorioDiario(id int) error {
	ctx := context.Background()

	query := `DELETE FROM relatorios_diarios WHERE id = $1`
	_, err := database.DB.Exec(ctx, query, id)
	return err
}

// AtualizarRelatorioDiario modifica os campos de um histórico existente incluindo novos anexos no array do Postgres
func AtualizarRelatorioDiario(r *models.RelatorioDiario) error {
	ctx := context.Background()

	query := `UPDATE relatorios_diarios 
	          SET data = $1, descricao = $2, participantes = $3, pendencias = $4, veiculos = $5, imagens = $6 
	          WHERE id = $7`

	_, err := database.DB.Exec(ctx, query, r.Data, r.Descricao, r.Participantes, r.Pendencias, r.Veiculos, r.Imagens, r.ID)
	return err
}

func BuscarImagensRelatorio(id int) ([]string, error) {
	ctx := context.Background()
	var imagens []string

	query := `SELECT imagens FROM relatorios_diarios WHERE id = $1`
	err := database.DB.QueryRow(ctx, query, id).Scan(&imagens)
	if err != nil {
		return nil, err
	}
	return imagens, err
}
