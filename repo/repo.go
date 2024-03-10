package repo

import (
	"context"
	"errors"

	"github.com/alexandrepossebom/rinha-backend-2024/model"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository interface {
	GetCliente(ctx context.Context, id uint32) *model.Cliente
	GetTransacoes(ctx context.Context, id uint32) []model.Transacao
	AddTransacao(ctx context.Context, id uint32, t *model.TransacaoPost) (model.TransacaoReply, error)
}

var ErrLimite = errors.New("saldo insuficiente")

type repository struct {
	db *pgxpool.Pool
}

func NewRepository(connection *pgxpool.Pool) Repository {
	return &repository{db: connection}
}

func (r *repository) GetCliente(ctx context.Context, id uint32) *model.Cliente {
	sqlStatement := `SELECT id, saldo, limite FROM clientes WHERE id = $1`
	rows, err := r.db.Query(ctx, sqlStatement, id)
	if err != nil {
		return nil
	}
	defer rows.Close()

	cliente := &model.Cliente{}
	if rows.Next() {
		err = rows.Scan(&cliente.ID, &cliente.Saldo, &cliente.Limite)
		if err != nil {
			return nil
		}
	} else {
		return nil
	}
	return cliente
}

func (r *repository) GetTransacoes(ctx context.Context, id uint32) []model.Transacao {
	transacoes := make([]model.Transacao, 0, 10)
	sqlStatement := `SELECT valor, tipo, descricao, realizada_em FROM transacoes t WHERE cliente_id = $1 ORDER BY id DESC LIMIT 10`
	rows, err := r.db.Query(ctx, sqlStatement, id)
	if err != nil {
		return transacoes
	}
	defer rows.Close()
	for rows.Next() {
		t := model.Transacao{}
		err = rows.Scan(&t.Valor, &t.Tipo, &t.Descricao, &t.Realizada)
		if err != nil {
			return []model.Transacao{}
		}
		transacoes = append(transacoes, t)
	}
	return transacoes
}

func (r *repository) AddTransacao(ctx context.Context, id uint32, t *model.TransacaoPost) (model.TransacaoReply, error) {
	var tr model.TransacaoReply

	valor := t.Valor
	if t.Tipo == "d" {
		valor = -t.Valor
	}

	var saldo, limite int32
	err := r.db.QueryRow(ctx, `UPDATE clientes SET saldo = saldo + $1 WHERE id = $2 RETURNING saldo, limite`, valor, id).Scan(&saldo, &limite)
	if err != nil {
		return tr, ErrLimite
	}

	_, err = r.db.Exec(ctx, `INSERT INTO transacoes (valor, tipo, descricao, cliente_id) VALUES ($1, $2, $3, $4)`, t.Valor, t.Tipo, t.Descricao, id)
	if err != nil {
		return tr, err
	}

	tr.Saldo = saldo
	tr.Limite = limite

	return tr, nil
}

// func (r *repository) AddTransacao(ctx context.Context, id uint32, t *model.TransacaoPost) (model.TransacaoReply, error) {
// 	var tr model.TransacaoReply

// 	valor := t.Valor
// 	if t.Tipo == "d" {
// 		valor = -t.Valor
// 	}

// 	batch := &pgx.Batch{}
// 	var saldo, limite int32
// 	batch.Queue(`UPDATE clientes SET saldo = saldo + $1 WHERE id = $2 RETURNING saldo, limite`, valor, id)
// 	batch.Queue(`INSERT INTO transacoes (valor, tipo, descricao, cliente_id) VALUES ($1, $2, $3, $4)`, t.Valor, t.Tipo, t.Descricao, id)

// 	br := r.db.SendBatch(ctx, batch)
// 	defer br.Close()

// 	rows, err := br.Query()
// 	if err != nil {
// 		log.Println(err)
// 		return tr, err
// 	}

// 	if !rows.Next() {
// 		return tr, ErrLimite
// 	}

// 	if err := rows.Scan(&saldo, &limite); err != nil {
// 		return tr, err
// 	}

// 	tr.Saldo = saldo
// 	tr.Limite = limite

// 	return tr, nil
// }
