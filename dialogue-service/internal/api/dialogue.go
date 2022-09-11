package api

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/ivanmakarychev/social-network/dialogue-service/internal/models"
)

func (a *API) PostMessage(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		a.postMessage(w, r)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (a *API) GetDialogue(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		a.getDialogue(w, r)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (a *API) getDialogue(w http.ResponseWriter, r *http.Request) {
	ids := r.URL.Query()["who"]
	if len(ids) != 1 {
		http.Error(w, "who", http.StatusBadRequest)
		return
	}
	id, err := strconv.ParseUint(ids[0], 10, 64)
	if err != nil {
		handleError("get dialogue", "parse dialogue id", err, w)
		return
	}
	who := models.ProfileID(id)
	ids = r.URL.Query()["with"]
	if len(ids) != 1 {
		http.Error(w, "with", http.StatusBadRequest)
		return
	}
	id, err = strconv.ParseUint(ids[0], 10, 64)
	if err != nil {
		handleError("get dialogue", "parse dialogue id", err, w)
		return
	}
	with := models.ProfileID(id)
	if with == who {
		http.Error(w, "with==who", http.StatusBadRequest)
		return
	}
	dialogueID := models.DialogueID{
		ProfileID1: who,
		ProfileID2: with,
	}
	messages, err := a.dialogueRepo.GetMessages(
		dialogueID,
	)
	if err != nil {
		handleError("get dialogue", "get dialogue messages", err, w)
		return
	}
	dialogueData := models.DialogueData{
		DialogueID: dialogueID,
		Messages:   messages,
	}

	var unreadMessagesTSs []time.Time
	for _, msg := range messages {
		if msg.Author == with && msg.Status != models.MessageStatusRead {
			unreadMessagesTSs = append(unreadMessagesTSs, msg.TS)
		}
	}
	if len(unreadMessagesTSs) != 0 {
		err = a.saga.UpdateMessages(&models.AlterCounterRequest{
			Key: models.UnreadMessagesCounterKey{
				From: with,
				To:   who,
			},
			Action:            models.ActionRead,
			MessageTimestamps: unreadMessagesTSs,
		})
		if err != nil {
			log.Println("failed to update read messages", err)
		}
	}

	writeJSON("get dialogue", w, dialogueData)
}

func (a *API) postMessage(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		handleError("post message", "read body", err, w)
		return
	}

	data := models.PostMessageRq{}
	err = json.Unmarshal(b, &data)
	if err != nil {
		log.Println(string(b))
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if data.To == data.From {
		http.Error(w, "to==from", http.StatusBadRequest)
		return
	}

	if data.To == 0 || data.From == 0 {
		http.Error(w, "zero to/from", http.StatusBadRequest)
		return
	}

	ts := time.Now()

	err = a.dialogueRepo.SaveMessage(&models.MessageData{
		DialogueID: models.DialogueID{
			ProfileID1: data.From,
			ProfileID2: data.To,
		},
		Message: &models.Message{
			Author: data.From,
			Text:   data.Text,
			TS:     ts,
		},
	})
	if err != nil {
		handleError("post message", "save message", err, w)
		return
	}

	err = a.saga.UpdateMessages(&models.AlterCounterRequest{
		Key: models.UnreadMessagesCounterKey{
			From: data.From,
			To:   data.To,
		},
		Action:            models.ActionCreated,
		MessageTimestamps: []time.Time{ts},
	})
	if err != nil {
		handleError("post message", "saga update messages", err, w)
		return
	}

	w.WriteHeader(http.StatusOK)
}
