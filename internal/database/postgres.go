package database

import (
	"context"
	"fmt"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

var DB *pgxpool.Pool

func ConectarDB() {
	// Se houver uma variavel de ambiente chamada DATABASE_URL, usa ela.
	// Dica: Em produção, mude esta string para ler de variáveis de ambiente (os.Getenv)
	err := godotenv.Load()
	if err != nil {
		fmt.Print("Aviso: Arquivo .env não localizado. Usando variaveis do sistema.")
	}

	connStr := os.Getenv("DATABASE_URL")
	if connStr == "" {
		connStr = "postgres://postgres:Ev323232@@localhost:5432/enproject"
	}

	config, err := pgxpool.ParseConfig(connStr)
	if err != nil {
		fmt.Printf("Erro ao configurar banco: %v\n", err)
		os.Exit(1)
	}

	pool, err := pgxpool.NewWithConfig(context.Background(), config)
	if err != nil {
		fmt.Printf("Erro ao conectar ao Postgres: %v\n", err)
		os.Exit(1)
	}

	DB = pool
	fmt.Println("🐘 Conexão com o PostgreSQL estabelecida com sucesso!")
}
