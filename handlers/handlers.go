package handlers

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/alexandrepossebom/rinha-backend-2024/model"
	"github.com/alexandrepossebom/rinha-backend-2024/repo"
)

type handler struct {
	rp repo.Repository
}

func NewHandler(rp repo.Repository) *handler {
	return &handler{rp: rp}
}

func (h *handler) NewTransacaoHandler(w http.ResponseWriter, r *http.Request) {
	t := &model.TransacaoPost{}
	if err := json.NewDecoder(r.Body).Decode(t); err != nil {
		w.WriteHeader(http.StatusUnprocessableEntity)
		return
	}

	idStr := r.PathValue("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if t.Tipo != "c" && t.Tipo != "d" || t.Valor <= 0 || len(t.Descricao) < 1 || len(t.Descricao) > 10 || err != nil {
		w.WriteHeader(http.StatusUnprocessableEntity)
		return
	}

	tr, err := h.rp.AddTransacao(r.Context(), uint32(id), t)
	if err != nil {
		if errors.Is(err, repo.ErrLimite) {
			w.WriteHeader(http.StatusUnprocessableEntity)
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	replyOkJson(w, tr)
}

func (h *handler) NewExtratoHandler(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	cliente := h.rp.GetCliente(r.Context(), uint32(id))

	if cliente == nil {
		http.Error(w, "Cliente não encontrado", http.StatusNotFound)
		return
	}

	transacoes := h.rp.GetTransacoes(r.Context(), uint32(id))
	extrato := model.Extrato{
		Saldo:      model.Saldo{Total: cliente.Saldo, Limite: cliente.Limite, DataExtrato: time.Now()},
		Transacoes: transacoes,
	}
	replyOkJson(w, extrato)
}

func replyOkJson(w http.ResponseWriter, v any) {
	resp, err := json.Marshal(v)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if _, err = w.Write(resp); err != nil {
		log.Println(err)
	}
}
