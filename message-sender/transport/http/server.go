package http

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	httpSwagger "github.com/swaggo/http-swagger"
	"go.uber.org/zap"

	"message-sender/config"
	"message-sender/model"
	"message-sender/service"
	_ "message-sender/transport/http/docs"
)

type Server struct {
	srv    *http.Server
	router *mux.Router
	logger *zap.Logger
	svc    service.Service
}

func NewServer(cfg *config.Config, logger *zap.Logger, svc service.Service) *Server {
	router := mux.NewRouter()

	server := &Server{
		srv: &http.Server{
			Addr:         fmt.Sprintf(":%d", cfg.Server.Port),
			Handler:      router,
			ReadTimeout:  cfg.Server.ReadTimeout,
			WriteTimeout: cfg.Server.WriteTimeout,
		},
		router: router,
		logger: logger,
		svc:    svc,
	}

	server.registerRoutes()
	return server
}

func (s *Server) Start() error {
	s.logger.Info("Starting HTTP server", zap.String("addr", s.srv.Addr))
	return s.srv.ListenAndServe()
}

func (s *Server) Stop() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	s.logger.Info("Stopping HTTP server")
	return s.srv.Shutdown(ctx)
}

func (s *Server) registerRoutes() {
	api := s.router.PathPrefix("/api").Subrouter()

	api.HandleFunc("/service", s.handleServiceControl).Methods(http.MethodPost)

	api.HandleFunc("/messages/sent", s.handleGetSentMessages).Methods(http.MethodGet)

	s.router.HandleFunc("/health", s.handleHealthCheck).Methods(http.MethodGet)

	s.router.PathPrefix("/swagger/").Handler(httpSwagger.Handler(
		httpSwagger.DeepLinking(true),
		httpSwagger.DocExpansion("none"),
		httpSwagger.DomID("swagger-ui"),
	))
}

func (s *Server) handleServiceControl(w http.ResponseWriter, r *http.Request) {
	var req model.StartStopRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	if !req.Action.IsValid() {
		s.respondWithError(w, http.StatusBadRequest, "Invalid action, must be 'start' or 'stop'")
		return
	}

	ctx := r.Context()
	var err error
	var message string

	switch req.Action {
	case model.ActionStart:
		err = s.svc.StartService(ctx)
		message = "Service started successfully"
	case model.ActionStop:
		err = s.svc.StopService(ctx)
		message = "Service stopped successfully"
	}

	if err != nil {
		s.logger.Error("Failed to control service", zap.Error(err), zap.String("action", string(req.Action)))
		s.respondWithError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to %s service", req.Action))
		return
	}

	s.respondWithJSON(w, http.StatusOK, model.StartStopResponse{
		Status:  "success",
		Message: message,
	})
}

func (s *Server) handleGetSentMessages(w http.ResponseWriter, r *http.Request) {
	pageStr := r.URL.Query().Get("page")
	limitStr := r.URL.Query().Get("limit")

	page := 1
	limit := 10

	if pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}

	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}

	response, err := s.svc.GetSentMessages(r.Context(), page, limit)
	if err != nil {
		s.logger.Error("Failed to get sent messages", zap.Error(err))
		s.respondWithError(w, http.StatusInternalServerError, "Failed to retrieve sent messages")
		return
	}

	s.respondWithJSON(w, http.StatusOK, response)
}

func (s *Server) handleHealthCheck(w http.ResponseWriter, r *http.Request) {
	status, err := s.svc.GetServiceStatus(r.Context())
	if err != nil {
		s.respondWithJSON(w, http.StatusOK, map[string]interface{}{
			"status":  "ok",
			"service": "unknown",
			"time":    time.Now().Format(time.RFC3339),
		})
		return
	}

	s.respondWithJSON(w, http.StatusOK, map[string]interface{}{
		"status":  "ok",
		"service": status,
		"time":    time.Now().Format(time.RFC3339),
	})
}

func (s *Server) respondWithError(w http.ResponseWriter, code int, message string) {
	s.respondWithJSON(w, code, map[string]string{"error": message})
}

func (s *Server) respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	response, err := json.Marshal(payload)
	if err != nil {
		s.logger.Error("Failed to marshal JSON response", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_, err = w.Write(response)
	if err != nil {
		s.logger.Error("Failed to write response", zap.Error(err))
	}
}
