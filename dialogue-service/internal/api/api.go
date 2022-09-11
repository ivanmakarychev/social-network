package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/ivanmakarychev/social-network/dialogue-service/internal/saga"

	consulapi "github.com/hashicorp/consul/api"
	"github.com/ivanmakarychev/social-network/dialogue-service/internal/config"
	"github.com/ivanmakarychev/social-network/dialogue-service/internal/repository"
	"github.com/pkg/errors"
	"github.com/urfave/negroni"
)

type API struct {
	cfg          config.Server
	dialogueRepo repository.DialogueRepository
	saga         *saga.Saga
}

func NewAPI(
	cfg config.Server,
	dialogueRepo repository.DialogueRepository,
	saga *saga.Saga,
) *API {
	return &API{
		cfg:          cfg,
		dialogueRepo: dialogueRepo,
		saga:         saga,
	}
}

func (a *API) Run() error {
	err := a.registerInConsul()
	if err != nil {
		return errors.Wrap(err, "failed to register in consul")
	}

	mux := http.NewServeMux()

	mux.HandleFunc("/dialogue", logger(a.GetDialogue))
	mux.HandleFunc("/dialogue/message/send", logger(a.PostMessage))

	mux.HandleFunc("/healthcheck", a.healthcheck)

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%s", a.cfg.Port),
		Handler: mux,
	}

	log.Printf("start server on %s", srv.Addr)
	return srv.ListenAndServe()
}

func (a *API) registerInConsul() error {
	consulCfg := consulapi.DefaultConfig()
	consul, err := consulapi.NewClient(consulCfg)
	if err != nil {
		log.Println(err)
	}

	port, err := strconv.Atoi(a.cfg.Port)
	if err != nil {
		return errors.Wrap(err, "port is not int")
	}
	address, err := os.Hostname()
	if err != nil {
		return errors.Wrap(err, "failed to get hostname")
	}
	serviceID := os.Getenv("CONSUL_SERVICE_ID")

	registration := &consulapi.AgentServiceRegistration{
		ID:      serviceID,
		Name:    a.cfg.ServiceName,
		Port:    port,
		Address: address,
		Check: &consulapi.AgentServiceCheck{
			HTTP:     fmt.Sprintf("http://%s:%d/healthcheck", address, port),
			Interval: "10s",
			Timeout:  "30s",
		},
	}

	err = consul.Agent().ServiceRegister(registration)
	if err != nil {
		log.Printf("failed to register service: %s:%v ", address, port)
	} else {
		log.Printf("successfully registered service: %s:%v", address, port)
	}
	return err
}

func (a *API) healthcheck(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
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
