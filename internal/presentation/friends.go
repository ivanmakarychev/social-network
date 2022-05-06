package presentation

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/ivanmakarychev/social-network/internal/models"
)

func (a *App) MakeFriend(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		profile, err := a.getOwnerProfileFromContext(r)
		if err != nil {
			handleError("make friend", "get profile from context", err, w)
			return
		}
		err = r.ParseForm()
		if err != nil {
			handleError("make friend", "parse form", err, w)
			return
		}
		ids := r.PostForm["other_profile_id"]
		if len(ids) != 1 {
			http.Error(w, "other_profile_id", http.StatusBadRequest)
			return
		}
		id, err := strconv.ParseUint(ids[0], 10, 64)
		if err != nil {
			handleError("make friend", "parse profile id", err, w)
			return
		}
		otherID := models.ProfileID(id)
		if profile.ID == otherID {
			http.Redirect(w, r, "/my/profile", http.StatusFound)
			return
		}
		for _, proposal := range profile.FriendshipApplications {
			if proposal.ID == models.ProfileID(id) {
				err = a.friendsRepo.ConfirmFriendship(profile.ID, otherID)
				if err != nil {
					handleError("make friend", "confirm friendship", err, w)
					return
				}
				http.Redirect(w, r, fmt.Sprintf("/profile?friendship_proposed=1&id=%d", id), http.StatusFound)
				return
			}
		}
		for _, friend := range profile.Friends {
			if friend.ID == otherID {
				http.Redirect(w, r, fmt.Sprintf("/profile?friendship_proposed=1&id=%d", id), http.StatusFound)
				return
			}
		}
		err = a.friendsRepo.MakeFriend(profile.ID, otherID)
		if err != nil {
			handleError("make friend", "make friend", err, w)
			return
		}
		http.Redirect(w, r, fmt.Sprintf("/profile?friendship_proposed=1&id=%d", id), http.StatusFound)
	default:
		http.Error(w, "", http.StatusMethodNotAllowed)
	}
}

func (a *App) ConfirmFriendship(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		profile, err := a.getOwnerProfileFromContext(r)
		if err != nil {
			handleError("confirm friend", "get profile from context", err, w)
			return
		}
		err = r.ParseForm()
		if err != nil {
			handleError("confirm friend", "parse form", err, w)
			return
		}
		ids := r.PostForm["other_profile_id"]
		if len(ids) != 1 {
			http.Error(w, "other_profile_id", http.StatusBadRequest)
			return
		}
		id, err := strconv.ParseUint(ids[0], 10, 64)
		if err != nil {
			handleError("confirm friend", "parse profile id", err, w)
			return
		}
		otherID := models.ProfileID(id)
		if profile.ID == otherID {
			http.Redirect(w, r, "/my/profile", http.StatusFound)
			return
		}
		proposalTakesPlace := false
		for _, proposal := range profile.FriendshipApplications {
			if proposal.ID == otherID {
				proposalTakesPlace = true
				break
			}
		}
		if !proposalTakesPlace {
			http.Error(w, "", http.StatusBadRequest)
			return
		}
		err = a.friendsRepo.ConfirmFriendship(profile.ID, otherID)
		if err != nil {
			handleError("confirm friend", "confirm friend", err, w)
			return
		}
		http.Redirect(w, r, "/my/profile", http.StatusFound)
	default:
		http.Error(w, "", http.StatusMethodNotAllowed)
	}
}
