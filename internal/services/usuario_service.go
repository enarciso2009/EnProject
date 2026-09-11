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
