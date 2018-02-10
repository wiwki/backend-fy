package main

import (
	"net/http"
	"html/template"
	"strconv"
	"log"
	"github.com/gin-gonic/gin/json"
	"github.com/gorilla/mux"
	"strings"
)

//
func TelegramAuthHandler(w http.ResponseWriter, r *http.Request) {
	var user User
	id, _ := strconv.Atoi(r.URL.Query().Get("id"))
	firstname := r.URL.Query().Get("first_name")
	lastname := r.URL.Query().Get("last_name")
	username := r.URL.Query().Get("username")
	photourl := r.URL.Query().Get("photo_url")
	authdate := r.URL.Query().Get("auth_date")
	hash := r.URL.Query().Get("hash")
	user.ID = uint(id)
	user.FirstName = firstname
	user.LastName = lastname
	user.UserName = username
	user.PhotoUrl = photourl
	user.Hash = hash
	user.AuthDate = authdate
	user.Role = "user"
	if err := SaveUser(user); err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		msg, _ := json.Marshal(map[string]string{"message": err.Error()})
		w.Write(msg)
		return
	}
	parameters := strings.Split(r.URL.String(), "/")[len(strings.Split(r.URL.String(), "/"))-1]
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("<script>window.location.replace('lolkek://" + parameters + "');</script>"))
}

func LoginViewHandler(w http.ResponseWriter, r *http.Request) {
	template.Must(template.ParseFiles("templates/index.html")).ExecuteTemplate(w, "index", nil)
}

func GetUserHandler(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(mux.Vars(r)["id"])
	user, err := GetUser(uint(id))
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		msg, _ := json.Marshal(map[string]string{"message": err.Error()})
		w.Write(msg)
	}
	msg, err := json.Marshal(&user)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		msg, _ = json.Marshal(map[string]string{"message": err.Error()})
		w.Write(msg)
	}
	w.WriteHeader(http.StatusOK)
	w.Write(msg)
}

func AddPostHandler(w http.ResponseWriter, r *http.Request) {
	url := r.URL.Query().Get("url")
	id, _ := strconv.Atoi(r.URL.Query().Get("user_id"))
	post, err := ParsePost(url)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		msg, _ := json.Marshal(map[string]string{"message": err.Error()})
		w.Write(msg)
	}
	post.UserID = uint(id)
	if err := AddPost(post); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		msg, _ := json.Marshal(map[string]string{"message": err.Error()})
		w.Write(msg)
	}
	w.WriteHeader(http.StatusOK)
	msg, _ := json.Marshal(map[string]string{"message": "success"})
	w.Write(msg)
}

func DeletePostHandler(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(r.URL.Query().Get("id"))
	err := DeletePost(uint(id))
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		msg, _ := json.Marshal(map[string]string{"message": err.Error()})
		w.Write(msg)
	}
	w.WriteHeader(http.StatusOK)
	msg, _ := json.Marshal(map[string]string{"message": "success"})
	w.Write(msg)
}

func GetPostHandler(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(mux.Vars(r)["id"])
	post, err := GetPost(uint(id))
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		msg, _ := json.Marshal(map[string]string{"message": err.Error()})
		w.Write(msg)
	}
	msg, _ := json.Marshal(&post)
	w.WriteHeader(http.StatusOK)
	w.Write(msg)
}

func GetFeedHandler(w http.ResponseWriter, r *http.Request) {
	var page int
	if mux.Vars(r)["page"] != "" {
		page, _ = strconv.Atoi(mux.Vars(r)["page"])
	} else {
		page = 1
	}
	var feed []Post
	feed, err := GetFeed(page - 1)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		msg, _ := json.Marshal(map[string]string{"message": err.Error()})
		w.Write(msg)

	}
	msg, _ := json.Marshal(&feed)
	w.WriteHeader(http.StatusOK)
	w.Write(msg)
}

func AddCommentHandler(w http.ResponseWriter, r *http.Request) {
	var comment Comment
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&comment)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		msg, _ := json.Marshal(map[string]string{"message": err.Error()})
		w.Write(msg)
	}
	if err := AddComment(comment); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		msg, _ := json.Marshal(map[string]string{"message": err.Error()})
		w.Write(msg)
	}
	w.WriteHeader(http.StatusOK)
	msg, _ := json.Marshal(map[string]string{"message": "success"})
	w.Write(msg)
}

func DeleteCommentHandler(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(r.URL.Query().Get("id"))
	err := DeleteComment(uint(id))
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		msg, _ := json.Marshal(map[string]string{"message": err.Error()})
		w.Write(msg)
	}
	w.WriteHeader(http.StatusOK)
	msg, _ := json.Marshal(map[string]string{"message": "success"})
	w.Write(msg)
}
