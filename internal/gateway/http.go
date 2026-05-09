package gateway

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	"example.com/tinder/internal/model"
	"example.com/tinder/internal/service"
)

type ctxKey string

const (
	errDetailKey ctxKey = "err_detail"
	errCodeKey   ctxKey = "err_code"
)

type HTTPGateway struct {
	logger *slog.Logger
	svc    service.MatchingService
	mux    *http.ServeMux
}

type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func NewHTTPGateway(logger *slog.Logger, svc service.MatchingService) *HTTPGateway {
	g := &HTTPGateway{
		logger: logger,
		svc:    svc,
		mux:    http.NewServeMux(),
	}
	g.routes()
	return g
}

func (g *HTTPGateway) Handler() http.Handler {
	return g.loggingMiddleware(g.mux)
}

func (g *HTTPGateway) loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		wrapped := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

		next.ServeHTTP(wrapped, r)

		code, _ := r.Context().Value(errCodeKey).(string)
		detail, _ := r.Context().Value(errDetailKey).(string)

		args := []any{
			"method", r.Method,
			"path", r.URL.Path,
			"status", wrapped.statusCode,
			"duration", time.Since(start).String(),
		}
		if code != "" {
			args = append(args, "error_code", code)
		}
		if detail != "" {
			args = append(args, "error_detail", detail)
		}
		switch {
		case wrapped.statusCode >= 500:
			g.logger.Error("server_error", args...)
		case wrapped.statusCode >= 400:
			g.logger.Warn("client_error", args...)
		default:
			g.logger.Info("request_success", args...)
		}
	})
}

func (g *HTTPGateway) replyError(w http.ResponseWriter, r *http.Request, status int, code string, msg string, err error) {

	ctx := context.WithValue(r.Context(), errCodeKey, code)
	if err != nil {
		ctx = context.WithValue(ctx, errDetailKey, err.Error())
	}
	*r = *r.WithContext(ctx)

	writeJSON(w, status, map[string]string{
		"error_code": code,
		"message":    msg,
	})
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func (g *HTTPGateway) routes() {
	g.mux.HandleFunc("POST /api/v1/people/match", g.handleAddSinglePersonAndMatch)
	g.mux.HandleFunc("DELETE /api/v1/people/{name}", g.handleRemoveSinglePerson)
	g.mux.HandleFunc("GET /api/v1/people", g.handleQuerySinglePeople)
	g.mux.HandleFunc("GET /api/v1/people/{name}", g.handleQuerySinglePerson)
	g.mux.HandleFunc("GET /api/v1/people/{name}/matches", g.handleQueryPersonMatches)
}

func (g *HTTPGateway) handleAddSinglePersonAndMatch(w http.ResponseWriter, r *http.Request) {
	var req model.AddSinglePersonAndMatchRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		g.replyError(w, r, http.StatusBadRequest, "INVALID_JSON", "請求格式錯誤", err)
		return
	}
	//TBD 如何確認為唯一性
	person := model.Person{
		Name:        strings.TrimSpace(req.Name),
		Height:      req.Height,
		Gender:      req.Gender,
		WantedDates: req.WantedDates,
	}

	matches, err := g.svc.AddSinglePersonAndMatch(person)
	if err != nil {
		if strings.Contains(err.Error(), "already exists") {
			g.replyError(w, r, http.StatusConflict, "USER_ALREADY_EXISTS", "該姓名已在系統中", err)
			return
		}
		g.replyError(w, r, http.StatusBadRequest, "ADD_PERSON_FAILED", err.Error(), err)
		return
	}

	writeJSON(w, http.StatusOK, model.AddSinglePersonAndMatchResponse{Matches: matches})
}

func (g *HTTPGateway) handleRemoveSinglePerson(w http.ResponseWriter, r *http.Request) {
	name := strings.TrimSpace(r.PathValue("name"))
	if name == "" {
		g.replyError(w, r, http.StatusBadRequest, "MISSING_NAME", "姓名為必填欄位", nil)
		return
	}

	removed := g.svc.RemoveSinglePerson(name)
	if !removed {
		g.replyError(w, r, http.StatusNotFound, "PERSON_NOT_FOUND", "找不到該人員", nil)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (g *HTTPGateway) handleQuerySinglePeople(w http.ResponseWriter, r *http.Request) {
	people := g.svc.QuerySinglePeople()
	writeJSON(w, http.StatusOK, model.QuerySinglePeopleResponse{People: people})
}

func (g *HTTPGateway) handleQuerySinglePerson(w http.ResponseWriter, r *http.Request) {
	name := strings.TrimSpace(r.PathValue("name"))
	if name == "" {
		g.replyError(w, r, http.StatusBadRequest, "MISSING_NAME", "姓名為必填欄位", nil)
		return
	}
	person, ok := g.svc.QuerySinglePerson(name)
	if !ok {
		g.replyError(w, r, http.StatusNotFound, "PERSON_NOT_FOUND", "找不到該人員", nil)
		return
	}
	writeJSON(w, http.StatusOK, person)
}
func (g *HTTPGateway) handleQueryPersonMatches(w http.ResponseWriter, r *http.Request) {
	name := strings.TrimSpace(r.PathValue("name"))
	if name == "" {
		g.replyError(w, r, http.StatusBadRequest, "MISSING_NAME", "姓名為必填欄位", nil)
		return
	}

	rawTop := r.URL.Query().Get("top")
	if rawTop == "" {
		g.replyError(w, r, http.StatusBadRequest, "MISSING_TOP", "請提供查詢數量(top)", nil)
		return
	}

	top, err := strconv.Atoi(rawTop)
	if err != nil || top <= 0 {
		g.replyError(w, r, http.StatusBadRequest, "INVALID_TOP", "top 必須為正整數", err)
		return
	}

	matches, ok := g.svc.QueryPersonMatches(name, top)
	if !ok {
		g.replyError(w, r, http.StatusNotFound, "PERSON_NOT_FOUND", "找不到該人員或其配對資料", nil)
		return
	}

	writeJSON(w, http.StatusOK, model.QueryPersonMatchesResponse{
		Name:    name,
		Matches: matches,
	})
}
