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
	err := godotenv.Load()
	if err != nil {
		fmt.Println("Aviso: Arquivo .env não localizado. Usando variaveis do sistema.")
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

	// ➕ CHAMA A CRIAÇÃO AUTOMÁTICA DAS TABELAS AQUI
	CriarTabelas()
}

// CriarTabelas verifica se a estrutura existe e cria apenas o que estiver faltando
func CriarTabelas() {
	ctx := context.Background()

	// Lista de queries individuais
	queries := []string{
		`CREATE TABLE IF NOT EXISTS usuarios (
			id SERIAL PRIMARY KEY,
			email VARCHAR(255) UNIQUE NOT NULL,
			senha VARCHAR(255) NOT NULL,
			perfil VARCHAR(20) NOT NULL DEFAULT 'consultor',
			data_criacao TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
		);`,

		`CREATE TABLE IF NOT EXISTS projetos (
			id SERIAL PRIMARY KEY,
			nome VARCHAR(255) NOT NULL,
			gerente VARCHAR(255),
			status_geral VARCHAR(50),
			resumo TEXT,
			total_horas INT DEFAULT 0,
			imagens TEXT[],
			observacoes TEXT,
			introducao TEXT,
			localizacao TEXT,
			encerramento TEXT
		);`,

		`CREATE TABLE IF NOT EXISTS tarefas (
			id SERIAL PRIMARY KEY,
			projeto_id INT REFERENCES projetos(id) ON DELETE CASCADE,
			nome VARCHAR(255) NOT NULL,
			responsavel VARCHAR(255),
			data_inicio TIMESTAMP WITH TIME ZONE,
			data_fim TIMESTAMP WITH TIME ZONE,
			progresso INT DEFAULT 0,
			concluido BOOLEAN DEFAULT FALSE,
			status VARCHAR(50)
		);`,

		`CREATE TABLE IF NOT EXISTS relatorios_diarios (
			id SERIAL PRIMARY KEY,
			projeto_id INT REFERENCES projetos(id) ON DELETE CASCADE,
			data TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
			descricao TEXT,
			imagens TEXT[],
			participantes TEXT,
			pendencias TEXT
		);`,

		`CREATE TABLE IF NOT EXISTS materiais (
			id SERIAL PRIMARY KEY,
			projeto_id INT REFERENCES projetos(id) ON DELETE CASCADE,
			item INT,
			descricao TEXT,
			quantidade INT
		);`,

		`CREATE TABLE IF NOT EXISTS projeto_usuarios (
			projeto_id INT REFERENCES projetos(id) ON DELETE CASCADE,
			usuario_id INT REFERENCES usuarios(id) ON DELETE CASCADE,
			PRIMARY KEY (projeto_id, usuario_id)
		);`,
	}

	// Executa cada query de forma isolada para evitar erros de sintaxe no driver
	for _, query := range queries {
		_, err := DB.Exec(ctx, query)
		if err != nil {
			fmt.Printf("❌ Erro crítico ao criar tabelas no banco de dados: %v\n", err)
			os.Exit(1)
		}
	}

	fmt.Println("📋 Estrutura de tabelas verificada/criada com sucesso!")
}
