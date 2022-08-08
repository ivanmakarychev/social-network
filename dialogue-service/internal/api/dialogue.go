package api

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"

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

	err = a.dialogueRepo.SaveMessage(&models.MessageData{
		DialogueID: models.DialogueID{
			ProfileID1: data.From,
			ProfileID2: data.To,
		},
		Message: &models.Message{
			Author: data.From,
			Text:   data.Text,
		},
	})
	if err != nil {
		handleError("post message", "save message", err, w)
		return
	}
	w.WriteHeader(http.StatusOK)
}
