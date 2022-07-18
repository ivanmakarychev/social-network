package presentation

import (
	"fmt"
	"github.com/ivanmakarychev/social-network/internal/models"
	"log"
	"net/http"
	"strconv"
)

func (a *App) SendMessage(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		a.sendMessage(w, r)
	default:
		http.Error(w, "", http.StatusMethodNotAllowed)
	}
}

func (a *App) ShowDialogue(w http.ResponseWriter, r *http.Request) {
	profile, err := a.getOwnerProfileFromContext(r)
	if err != nil {
		handleError("show dialogue", "get profile from context", err, w)
		return
	}
	ids := r.URL.Query()["with"]
	if len(ids) != 1 {
		http.Error(w, "with", http.StatusBadRequest)
		return
	}
	id, err := strconv.ParseUint(ids[0], 10, 64)
	if err != nil {
		handleError("show dialogue", "parse dialogue id", err, w)
		return
	}
	recipientID := models.ProfileID(id)
	if recipientID == profile.ID {
		http.Redirect(w, r, "/my/profile", http.StatusFound)
	}
	dialogueID := models.DialogueID{
		ProfileID1: profile.ID,
		ProfileID2: recipientID,
	}
	messages, err := a.dialogueRepo.GetMessages(
		dialogueID,
	)
	if err != nil {
		handleError("show dialogue", "get dialogue messages", err, w)
		return
	}
	recipientProfile, err := a.profileProvider.GetProfile(recipientID)
	if err != nil {
		handleError("show dialogue", "get recipient profile", err, w)
		return
	}
	dialogueData := models.DialogueData{
		Dialogue: models.Dialogue{
			ID:   dialogueID,
			Who:  profile.ID,
			With: recipientProfile.ProfileMain,
		},
		Messages: messages,
	}
	data := ViewData{
		Title: "Диалог с " + dialogueData.With.Name,
		Data:  dialogueData,
	}
	tmpl, err := loadTemplate("dialogue.html")
	if err != nil {
		log.Println("failed to load template: ", err)
		http.Error(w, errorMessage, http.StatusInternalServerError)
		return
	}
	err = tmpl.ExecuteTemplate(w, "base", data)
	if err != nil {
		log.Println("bad template ", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (a *App) sendMessage(w http.ResponseWriter, r *http.Request) {
	profile, err := a.getOwnerProfileFromContext(r)
	if err != nil {
		handleError("send message", "get profile from context", err, w)
		return
	}
	err = r.ParseForm()
	if err != nil {
		handleError("send message", "parse form", err, w)
		return
	}
	tos := r.PostForm["to"]
	if len(tos) != 1 {
		http.Error(w, "to", http.StatusBadRequest)
		return
	}
	to, err := strconv.ParseUint(tos[0], 10, 64)
	if err != nil {
		handleError("send message", "parse recipient id", err, w)
		return
	}
	recipientID := models.ProfileID(to)
	if profile.ID == recipientID {
		http.Redirect(w, r, "/my/profile", http.StatusFound)
		return
	}
	texts := r.PostForm["text"]
	if len(texts) > 0 {
		text := texts[0]
		if len(text) > 0 {
			err = a.dialogueRepo.SaveMessage(&models.MessageData{
				DialogueID: models.DialogueID{
					ProfileID1: profile.ID,
					ProfileID2: recipientID,
				},
				Message: &models.Message{
					Author: profile.ID,
					Text:   text,
				},
			})
			if err != nil {
				handleError("send message", "save message", err, w)
				return
			}
		}
	}
	http.Redirect(w, r, fmt.Sprintf("/dialogue?with=%d", recipientID), http.StatusFound)
}
