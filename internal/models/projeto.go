package models

import "time"

type Tarefa struct {
	ID          int       `json:"id"`
	ProjetoID   int       `json:"projeto_id"`
	Nome        string    `json:"nome"`
	Responsavel string    `json:"responsavel"`
	DataInicio  time.Time `json:"data_inicio"`
	DataFim     time.Time `json:"data_fim"`
	Progresso   int       `json:"progresso"`
	Concluido   bool      `json:"concluido"`
	Status      string    `json:"status"`
}

type RelatorioDiario struct {
	ID            int       `json:"id"`
	ProjetoID     int       `json:"projeto_id"`
	Data          time.Time `json:"data"`
	Descricao     string    `json:"descricao"`
	Imagens       []string  `json:"imagens"`
	Participantes string    `json:"participantes"`
	Pendencias    string    `json:"pendencias"`
}

type Projeto struct {
	ID           int               `json:"id"`
	Nome         string            `json:"nome"`
	Gerente      string            `json:"gerente"`
	StatusGeral  string            `json:"status_geral"`
	Resumo       string            `json:"resumo"`
	Tarefas      []Tarefa          `json:"tarefas"`
	TotalHoras   int               `json:"total_horas"`
	Relatorios   []RelatorioDiario `json:"relatorios"`
	Imagens      []string          `json:"imagens"`
	Observacoes  string            `json:"observacoes"`
	Introducao   string            `json:"introducao"`
	Localizacao  string            `json:"localizacao"`
	Encerramento string            `json:"encerramento"`
	Materiais    []Material        `json:"materiais"`
}

type Usuario struct {
	ID          int
	Email       string
	Senha       string
	DataCriacao time.Time
	Perfil      string
}

type Material struct {
	ID         int
	ProjetoID  int
	Item       int
	Descricao  string
	Quantidade int
}
