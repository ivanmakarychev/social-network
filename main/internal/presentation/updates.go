package presentation

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/websocket"

	"github.com/ivanmakarychev/social-network/internal/models"
)

type TapeData struct {
	Updates []*models.Update
}

func (a *App) Tape(w http.ResponseWriter, r *http.Request) {
	profile, err := a.getOwnerProfileFromContext(r)
	if err != nil {
		handleError("show tape", "get profile from context", err, w)
		return
	}
	tape, err := a.tapeProvider.GetTape(profile.ID)
	if err != nil {
		handleError("show tape", "get tape", err, w)
		return
	}
	data := ViewData{
		Title: "Твоя лента",
		Data: TapeData{
			Updates: tape,
		},
	}
	loadAndExecuteTemplate("tape.html", data, w)
}

func (a *App) PublishUpdate(w http.ResponseWriter, r *http.Request) {
	profile, err := a.getOwnerProfileFromContext(r)
	if err != nil {
		handleError("publish update", "get profile from context", err, w)
		return
	}
	err = r.ParseForm()
	if err != nil {
		handleError("publish update", "parse form", err, w)
		return
	}
	texts := r.PostForm["text"]
	if len(texts) != 1 {
		http.Error(w, "text", http.StatusBadRequest)
		return
	}
	err = a.publisher.Publish(&models.Update{
		Author: profile.ProfileMain,
		Text:   texts[0],
		TS:     time.Now(),
	})
	if err != nil {
		handleError("publish update", "publish", err, w)
		return
	}
	http.Redirect(w, r, "/success", http.StatusFound)
}

func (a *App) Subscribe(w http.ResponseWriter, r *http.Request) {
	profile, err := a.getOwnerProfileFromContext(r)
	if err != nil {
		handleError("subscribe", "get profile from context", err, w)
		return
	}
	err = r.ParseForm()
	if err != nil {
		handleError("subscribe", "parse form", err, w)
		return
	}
	ids := r.PostForm["profile_id"]
	if len(ids) != 1 {
		http.Error(w, "profile_id", http.StatusBadRequest)
		return
	}
	id, err := strconv.ParseUint(ids[0], 10, 64)
	if err != nil {
		handleError("subscribe", "parse profile id", err, w)
		return
	}
	otherID := models.ProfileID(id)
	if profile.ID == otherID {
		http.Redirect(w, r, "/my/profile", http.StatusFound)
		return
	}
	err = a.subscription.Subscribe(models.SubscriptionRq{
		SubscriberID: profile.ID,
		PublisherID:  otherID,
	})
	if err != nil {
		handleError("subscribe", "subscribe", err, w)
		return
	}
	http.Redirect(w, r, fmt.Sprintf("/profile?id=%d&subscribed=1", otherID), http.StatusFound)
}

var upgrader = websocket.Upgrader{} // use default options

func (a *App) UpdatesWS(w http.ResponseWriter, r *http.Request) {
	profile, err := a.getOwnerProfileFromContext(r)
	if err != nil {
		handleError("updates ws", "get profile from context", err, w)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		handleError("updates ws", "upgrade connection", err, w)
		return
	}
	defer conn.Close()

	updates, err := a.tapeProvider.SubscribeOnUpdates(profile.ID)
	if err != nil {
		handleError("updates ws", "subscribe on updates", err, w)
		return
	}

	conn.SetCloseHandler(func(code int, text string) error {
		err := a.tapeProvider.UnsubscribeFromUpdates(profile.ID)
		if err != nil {
			log.Println("updates ws failed to unsubscribe from updates:", err)
		}
		log.Println("updates ws conn closing")
		message := websocket.FormatCloseMessage(code, "")
		_ = conn.WriteControl(websocket.CloseMessage, message, time.Now().Add(time.Second))
		return nil
	})

	log.Println("updates ws: connection is ready")

	for upd := range updates {
		updForWS := UpdateForWS{
			AuthorName: upd.Author.Name,
			Text:       upd.Text,
			Datetime:   upd.TS.Format("15:04 01.02.2006"),
		}
		message, err := json.Marshal(updForWS)
		if err != nil {
			log.Println("updates ws failed to marshal message:", err)
			break
		}
		err = conn.WriteMessage(websocket.TextMessage, message)
		if err != nil {
			log.Println("updates ws failed to write message:", err)
			break
		}
		log.Println("updates ws sent message", string(message))
	}
}

type UpdateForWS struct {
	AuthorName string `json:"author_name"`
	Text       string `json:"text"`
	Datetime   string `json:"datetime"`
}
