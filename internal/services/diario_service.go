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
		INSERT INTO relatorios_diarios (projeto_id, data, descricao, imagens,participantes, pendencias) 
		VALUES ($1, $2, $3, $4, $5, $6) 
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
	).Scan(&diario.ID)

	if err != nil {
		return err
	}

	return nil
}
