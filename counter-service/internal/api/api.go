package api

import (
	"encoding/json"
	"fmt"
	"github.com/ivanmakarychev/social-network/counter-service/internal/models"
	"log"
	"net/http"
	"os"
	"strconv"

	consulapi "github.com/hashicorp/consul/api"
	"github.com/ivanmakarychev/social-network/counter-service/internal/config"
	"github.com/ivanmakarychev/social-network/counter-service/internal/repository"
	"github.com/pkg/errors"
	"github.com/urfave/negroni"
)

type API struct {
	cfg           config.Service
	counterGetter repository.CounterGetter
}

func NewAPI(
	cfg config.Service,
	counterGetter repository.CounterGetter,
) *API {
	return &API{
		cfg:           cfg,
		counterGetter: counterGetter,
	}
}

func (a *API) Run() error {
	err := a.registerInConsul()
	if err != nil {
		return errors.Wrap(err, "failed to register in consul")
	}

	mux := http.NewServeMux()

	mux.HandleFunc("/counter/messages/unread", logger(a.getUnreadMessagesCount))
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

func (a *API) getUnreadMessagesCount(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		from, err := getProfileID("from", r)
		if err != nil {
			http.Error(w, "from", http.StatusBadRequest)
			return
		}
		to, err := getProfileID("to", r)
		if err != nil {
			http.Error(w, "to", http.StatusBadRequest)
			return
		}
		count, err := a.counterGetter.Get(
			r.Context(),
			models.UnreadMessagesCounterKey{
				From: from,
				To:   to,
			},
		)
		if err != nil {
			handleError("getUnreadMessagesCount", "get counter", err, w)
			return
		}
		writeJSON("getUnreadMessagesCount", w, models.Counter{Count: count})
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
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

func getFromQuery[T any](parameterName string, r *http.Request, parser func(string) (T, error)) (T, error) {
	values := r.URL.Query()[parameterName]
	if len(values) == 0 {
		return parser("")
	}
	return parser(values[0])
}

func getProfileID(parameterName string, r *http.Request) (models.ProfileID, error) {
	return getFromQuery(parameterName, r, func(s string) (models.ProfileID, error) {
		id, err := strconv.ParseUint(s, 10, 64)
		return models.ProfileID(id), err
	})
}
