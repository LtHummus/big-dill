package server

import (
	"embed"
	"encoding/json"
	"fmt"
	"html/template"
	"io/fs"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/websocket"
	"github.com/rs/zerolog/log"

	"github.com/lthummus/big-dill/internal/socket"
	"github.com/lthummus/big-dill/internal/survey"
)

var token string

//go:embed frontend/*
var frontendFS embed.FS

//go:embed templates/results.gohtml
var resultsPageTemplate string

var resultTemplate *template.Template

func init() {
	log.Info().Msg("initializing template")
	t, err := template.New("").Parse(resultsPageTemplate)
	if err != nil {
		log.Panic().Err(err).Msg("could not parse template")
	}
	resultTemplate = t

	log.Info().Msg("loading token")
	token = os.Getenv("BIG_DILL_AUTH_TOKEN")
	if token == "" {
		log.Fatal().Msg("BIG_DILL_AUTH_TOKEN not set")
	}
}

type Server struct {
	dispatch *socket.Dispatch
	upgrader websocket.Upgrader
	mux      *http.ServeMux
	surveyor *survey.Surveyor
}

func New() *Server {
	surveyor := survey.NewSurveyor()
	s := &Server{
		dispatch: socket.NewDispatch(surveyor),
		upgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
		},
		mux:      http.NewServeMux(),
		surveyor: surveyor,
	}

	s.mux.HandleFunc("/socket", func(w http.ResponseWriter, r *http.Request) {
		conn, err := s.upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Warn().Err(err).Msg("could not upgrade connection")
			return
		}

		socket.NewClient(s.dispatch, conn)
	})

	s.mux.HandleFunc("/dump_votes", func(w http.ResponseWriter, r *http.Request) {
		votes := s.surveyor.GetVotes()

		payload, _ := json.Marshal(votes)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(payload)
	})

	s.mux.HandleFunc("GET /resultpage", func(w http.ResponseWriter, r *http.Request) {
		if !hasValidAuth(r) {
			http.Error(w, "no", http.StatusForbidden)
			return
		}

		v := s.surveyor.GetVotes()
		resultArray := []int{
			v["-3"],
			v["-2"],
			v["-1"],
			v["0"],
			v["1"],
			v["2"],
			v["3"],
		}

		resultTemplate.Execute(w, map[string]any{
			"ResultArray": resultArray,
		})
	})

	s.mux.HandleFunc("POST /vote_open", func(w http.ResponseWriter, r *http.Request) {
		if !hasValidAuth(r) {
			http.Error(w, "no", http.StatusForbidden)
			return
		}
		s.surveyor.OpenVote()
	})

	s.mux.HandleFunc("POST /vote_close", func(w http.ResponseWriter, r *http.Request) {
		if !hasValidAuth(r) {
			http.Error(w, "no", http.StatusForbidden)
			return
		}
		s.surveyor.CloseVote()
	})

	s.mux.HandleFunc("/socket_url", func(w http.ResponseWriter, r *http.Request) {
		protocol := "ws"
		if r.Header.Get("X-Forwarded-Proto") == "https" {
			protocol = "wss"
		}
		socketURL := fmt.Sprintf("%s://%s/socket", protocol, r.Host)

		w.Header().Add("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"socket_url": socketURL,
		})
	})

	fileContents, err := fs.Sub(frontendFS, "frontend")
	if err != nil {
		log.Panic().Err(err).Msg("could not get filesystem")
	}
	s.mux.Handle("/", http.FileServer(http.FS(fileContents)))

	return s
}

func hasValidAuth(r *http.Request) bool {
	if r.Header.Get("X-Token") == token {
		return true
	}

	pwd, err := r.Cookie("pwd")
	if err != nil {
		return false
	}

	return pwd.Value == token
}

func (s *Server) ListenAndServe() {
	srv := &http.Server{
		Addr:         fmt.Sprintf("0.0.0.0:%d", 8899),
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
		IdleTimeout:  15 * time.Second,
		Handler:      s.mux,
	}

	err := srv.ListenAndServe()
	if err != nil {
		panic(err)
	}
}
