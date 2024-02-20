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
	sqlStatement := `SELECT valor, tipo, descricao, realizada_em FROM transacoes t WHERE cliente_id = $1 ORDER BY realizada_em DESC LIMIT 10`
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
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return tr, err
	}
	defer tx.Rollback(ctx)

	valor := t.Valor
	if t.Tipo == "d" {
		valor = -t.Valor
	}
	var saldo, limite int32
	sqlStatement := `UPDATE clientes SET saldo = saldo + $1 WHERE id = $2 RETURNING saldo, limite`
	err = tx.QueryRow(ctx, sqlStatement, valor, id).Scan(&saldo, &limite)
	if err != nil {
		return tr, ErrLimite
	}

	sqlStatement = `INSERT INTO transacoes (valor, tipo, descricao, cliente_id) VALUES ($1, $2, $3, $4)`
	_, err = tx.Exec(ctx, sqlStatement, t.Valor, t.Tipo, t.Descricao, id)
	if err != nil {
		return tr, err
	}

	if err = tx.Commit(ctx); err != nil {
		return tr, err
	}

	tr.Saldo = saldo
	tr.Limite = limite

	return tr, nil
}
