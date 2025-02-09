package handler

import (
	"backend/service"
	"encoding/json"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
	"net/http"
)

type Handler struct {
	serv service.Service
	log  *logrus.Logger
}

func NewHandler(serv service.Service, logger *logrus.Logger) *Handler {
	return &Handler{serv: serv, log: logger}
}

func (h *Handler) InitRoutes(router *mux.Router) {
	router.HandleFunc("/pings", h.GetPings).Methods(http.MethodGet)
}

func (h *Handler) GetPings(w http.ResponseWriter, r *http.Request) {
	h.log.Info("new request /pings")
	pings, err := h.serv.GetPings(r.Context())
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		h.log.Error(err)
	}
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(pings)
}
