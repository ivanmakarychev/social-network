package services

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/ivanmakarychev/social-network/internal/models"
)

type (
	DialogueService struct {
		connStr string
	}

	PostMessageRq struct {
		From models.ProfileID `json:"from"`
		To   models.ProfileID `json:"to"`
		Text string           `json:"text"`
	}

	GetDialogueRs struct {
		models.DialogueID `json:"dialogue_id"`
		Messages          []*models.Message `json:"messages"`
	}
)

func NewDialogueService(connStr string) *DialogueService {
	return &DialogueService{connStr: connStr}
}

func (d *DialogueService) GetMessages(id models.DialogueID) ([]*models.Message, error) {
	req, err := http.NewRequest(http.MethodGet, d.formatURL("dialogue"), nil)
	if err != nil {
		return nil, err
	}
	values := req.URL.Query()
	values.Set("who", id.ProfileID1.String())
	values.Set("with", id.ProfileID2.String())
	req.URL.RawQuery = values.Encode()
	rs, err := d.do(req)
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
	dialogueData := GetDialogueRs{}
	err = json.Unmarshal(b, &dialogueData)
	if err != nil {
		return nil, err
	}
	return dialogueData.Messages, nil
}

func (d *DialogueService) SaveMessage(msg *models.MessageData) error {
	rq := &PostMessageRq{
		From: msg.Message.Author,
		To:   msg.To,
		Text: msg.Message.Text,
	}
	b, err := json.Marshal(rq)
	if err != nil {
		return err
	}
	httpRq, err := http.NewRequest(http.MethodPost, d.formatURL("dialogue/message/send"), bytes.NewReader(b))
	if err != nil {
		return err
	}
	httpRq.Header.Set("Content-Type", "application/json")
	rs, err := d.do(httpRq)
	if err != nil {
		return err
	}
	defer rs.Body.Close()
	if rs.StatusCode != http.StatusOK {
		errBody, _ := ioutil.ReadAll(rs.Body)
		return errors.New(string(errBody))
	}
	return nil
}

func (d *DialogueService) formatURL(path string) string {
	return fmt.Sprintf("%s/%s", d.connStr, path)
}

func (d *DialogueService) do(req *http.Request) (*http.Response, error) {
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