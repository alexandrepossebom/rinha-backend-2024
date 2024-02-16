package model

import "time"

type TransacaoPost struct {
	Valor     int32  `json:"valor"`
	Tipo      string `json:"tipo"`
	Descricao string `json:"descricao"`
}

type Transacao struct {
	Valor     int32     `json:"valor"`
	Tipo      string    `json:"tipo"`
	Descricao string    `json:"descricao"`
	Realizada time.Time `json:"realizada_em"`
}

type TransacaoReply struct {
	Limite int32 `json:"limite"`
	Saldo  int32 `json:"saldo"`
}

type Cliente struct {
	ID     int32 `json:"id"`
	Saldo  int32 `json:"saldo"`
	Limite int32 `json:"limite"`
}

type Saldo struct {
	Total       int32     `json:"total"`
	Limite      int32     `json:"limite"`
	DataExtrato time.Time `json:"data_extrato"`
}

type Extrato struct {
	Saldo      Saldo       `json:"saldo"`
	Transacoes []Transacao `json:"ultimas_transacoes"`
}
