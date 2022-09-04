package services

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"

	consulapi "github.com/hashicorp/consul/api"
	"github.com/ivanmakarychev/social-network/internal/models"
	"github.com/pkg/errors"
)

type (
	CounterService interface {
		GetUnreadMessagesCounter(ctx context.Context, rq *GetMessagesCounterRq) (*GetCounterRs, error)
	}

	CounterServiceImpl struct {
		serviceName  string
		consulClient *consulapi.Client
	}

	GetMessagesCounterRq struct {
		From models.ProfileID `json:"from"`
		To   models.ProfileID `json:"to"`
	}

	GetCounterRs struct {
		Count int `json:"count"`
	}
)

func NewCounterService(serviceName string, consulClient *consulapi.Client) *CounterServiceImpl {
	return &CounterServiceImpl{
		serviceName:  serviceName,
		consulClient: consulClient,
	}
}

func (c *CounterServiceImpl) GetUnreadMessagesCounter(_ context.Context, rq *GetMessagesCounterRq) (*GetCounterRs, error) {
	url, err := c.formatURL("counter/messages/unread")
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	values := req.URL.Query()
	values.Set("from", rq.From.String())
	values.Set("to", rq.To.String())
	req.URL.RawQuery = values.Encode()
	rs, err := c.do(req)
	if err != nil {
		return nil, err
	}
	defer rs.Body.Close()
	b, err := ioutil.ReadAll(rs.Body)
	if err != nil {
		return nil, err
	}
	if rs.StatusCode != http.StatusOK {
		return nil, errors.New(string(b))
	}
	data := &GetCounterRs{}
	err = json.Unmarshal(b, data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (c *CounterServiceImpl) formatURL(path string) (string, error) {
	serviceEntries, _, err := c.consulClient.Health().Service(c.serviceName, "", true, nil)
	if err != nil {
		return "", errors.Wrapf(err, "failed to get service %q from consul", c.serviceName)
	}
	serviceEntry := serviceEntries[rand.Intn(len(serviceEntries))]
	return fmt.Sprintf("http://%s:%d/%s", serviceEntry.Service.Address, serviceEntry.Service.Port, path), nil
}

func (c *CounterServiceImpl) do(req *http.Request) (*http.Response, error) {
	rs, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Printf(
			"%s %s %s --> %s\n",
			req.Host,
			req.Method,
			req.URL.Path,
			err.Error(),
		)
		return rs, err
	}
	log.Printf(
		"%s %s %s --> %d\n",
		req.Host,
		req.Method,
		req.URL.Path,
		rs.StatusCode,
	)
	return rs, nil
}
