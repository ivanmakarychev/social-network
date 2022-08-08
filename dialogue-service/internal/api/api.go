package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/ivanmakarychev/social-network/dialogue-service/internal/config"
	"github.com/ivanmakarychev/social-network/dialogue-service/internal/repository"
	"github.com/urfave/negroni"
)

type API struct {
	cfg          config.Server
	dialogueRepo repository.DialogueRepository
}

func NewAPI(
	cfg config.Server,
	dialogueRepo repository.DialogueRepository,
) *API {
	return &API{
		cfg:          cfg,
		dialogueRepo: dialogueRepo,
	}
}

func (a *API) Run() error {
	mux := http.NewServeMux()

	mux.HandleFunc("/dialogue", logger(a.GetDialogue))
	mux.HandleFunc("/dialogue/message/send", logger(a.PostMessage))

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%s", a.cfg.Port),
		Handler: mux,
	}

	log.Printf("start server on %s", srv.Addr)
	return srv.ListenAndServe()
}

func handleError(actor, action string, err error, w http.ResponseWriter) {
	log.Println(
		fmt.Sprintf("[error] %s failed to %s: %s",
			actor,
			action,
			err,
		))
	http.Error(w, "something went wrong", http.StatusInternalServerError)
}

func writeJSON(actor string, w http.ResponseWriter, data any) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		handleError(actor, "marshal to json", err, w)
		return
	}
	_, _ = w.Write(jsonData)
	w.Header().Add("Content-Type", "application/json")
}

func logger(handlerFunc http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		lrw := negroni.NewResponseWriter(w)
		handlerFunc(lrw, r)
		log.Printf("%s %s --> %d\n",
			r.Method,
			r.URL.Path,
			lrw.Status(),
		)
	}
}
