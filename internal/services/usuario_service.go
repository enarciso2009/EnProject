package services

import (
	"EnProject/internal/database"
	"EnProject/internal/models"

	"context"
	"errors"

	"golang.org/x/crypto/bcrypt"
)

// CadastrarUsuario encripta a senha e salva o novo login no banco

func CadastrarUsuario(email, senhaPura string) error {
	ctx := context.Background()

	// Criptografa a senha antes de salvar
	senhaHash, err := bcrypt.GenerateFromPassword([]byte(senhaPura), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	query := `INSERT INTO usuarios (email, senha) VALUES ($1, $2)`
	_, err = database.DB.Exec(ctx, query, email, string(senhaHash))
	return err
}

// AutenticarUsuario valida as credenciais enviadas no login

func AutenticarUsuario(email, senhaPura string) (*models.Usuario, error) {
	ctx := context.Background()
	var u models.Usuario

	query := `SELECT id, email, senha from usuarios WHERE email = $1`
	err := database.DB.QueryRow(ctx, query, email).Scan(&u.ID, &u.Email, &u.Senha)
	if err != nil {
		return nil, errors.New("usuário não encontrado")
	}

	// Compara o hash do banco com a senha digitada pelo usuário
	err = bcrypt.CompareHashAndPassword([]byte(u.Senha), []byte(senhaPura))
	if err != nil {
		return nil, errors.New("senha incorreta")
	}

	return &u, nil
}
func CriarUsuarioCompleto(u *models.Usuario, projetoIDs []int) error {
	ctx := context.Background()
	tx, err := database.DB.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	// 1. Salva o usuário no banco (com a coluna perfil adicionada)
	queryUser := `INSERT INTO usuarios (email, senha, perfil) VALUES ($1, $2, $3) RETURNING id`
	err = tx.QueryRow(ctx, queryUser, u.Email, u.Senha, u.Perfil).Scan(&u.ID)
	if err != nil {
		return err
	}

	// 2. Se for consultor e houver projetos selecionados, cria o vínculo na tabela pivô
	if u.Perfil == "consultor" && len(projetoIDs) > 0 {
		queryVinculo := `INSERT INTO projeto_usuarios (projeto_id, usuario_id) VALUES ($1, $2)`
		for _, projID := range projetoIDs {
			_, err = tx.Exec(ctx, queryVinculo, projID, u.ID)
			if err != nil {
				return err
			}
		}
	}

	return tx.Commit(ctx)
}

// Garanta que este bloco de código existe e ESTÁ SALVO na pasta internal/services
func BuscarUsuarioPorEmail(email string) (*models.Usuario, error) {
	ctx := context.Background()
	var u models.Usuario

	query := `SELECT id, email, perfil FROM usuarios WHERE email = $1`
	err := database.DB.QueryRow(ctx, query, email).Scan(&u.ID, &u.Email, &u.Perfil)
	if err != nil {
		return nil, err
	}
	return &u, nil
}
