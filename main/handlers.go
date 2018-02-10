package main

import (
	"net/http"
	"html/template"
	"strconv"
	"log"
	"github.com/gin-gonic/gin/json"
	"github.com/gorilla/mux"
	"strings"
	"time"
)

//Обработка данных из авторизации telegram
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
	user.FullName = firstname + " " + lastname
	user.Username = username
	user.PhotoUrl = photourl
	user.Hash = hash
	user.AuthDate = authdate
	user.Role = "user"
	user.Birthday = time.Now().Add(-time.Hour * 876000)
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

//Страница авторизации telegram
func LoginViewHandler(w http.ResponseWriter, r *http.Request) {
	template.Must(template.ParseFiles("templates/index.html")).ExecuteTemplate(w, "index", nil)
}

//Контролер получения данных о юзере
func GetUserHandler(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(mux.Vars(r)["id"])
	user, err := GetUser(uint(id))
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		msg, _ := json.Marshal(map[string]string{"message": err.Error()})
		w.Write(msg)
		return
	}
	msg, err := json.Marshal(&user)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		msg, _ = json.Marshal(map[string]string{"message": err.Error()})
		w.Write(msg)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write(msg)
}

//Контроллер получения данных о юзере
func AddPostHandler(w http.ResponseWriter, r *http.Request) {
	url := r.URL.Query().Get("url")
	id, _ := strconv.Atoi(r.URL.Query().Get("user_id"))
	post, err := ParsePost(url)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		msg, _ := json.Marshal(map[string]string{"message": err.Error()})
		w.Write(msg)
		return
	}
	log.Println(post)
	post.UserID = uint(id)
	if err := AddPost(post); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		msg, _ := json.Marshal(map[string]string{"message": err.Error()})
		w.Write(msg)
		return
	}
	w.WriteHeader(http.StatusOK)
	msg, _ := json.Marshal(map[string]string{"message": "success"})
	w.Write(msg)
}

func EditUserHandler(w http.ResponseWriter, r *http.Request) {
	var user User
	id, _ := strconv.Atoi(r.URL.Query().Get("id"))
	user.ID = uint(id)
	user.Sex = r.URL.Query().Get("sex")
	user.Info = r.URL.Query().Get("info")
	unix, err := strconv.ParseInt(r.URL.Query().Get("birthday"), 10, 64)
	if err != nil {
		panic(err)
	}
	user.Birthday = time.Unix(unix, 0)
	user.Age = uint(time.Since(user.Birthday).Hours() / 8670)
	if err := EditUser(user); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		msg, _ := json.Marshal(map[string]string{"message":err.Error()})
		w.Write(msg)
		return
	}
	w.WriteHeader(http.StatusOK)
	msg, _ := json.Marshal(map[string]string{"message":"success"})
	w.Write(msg)
}

//Контроллер удаления поста из бд
func DeletePostHandler(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(r.URL.Query().Get("id"))
	err := DeletePost(uint(id))
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		msg, _ := json.Marshal(map[string]string{"message": err.Error()})
		w.Write(msg)
		return
	}
	w.WriteHeader(http.StatusOK)
	msg, _ := json.Marshal(map[string]string{"message": "success"})
	w.Write(msg)
}

func SearchUserHandler(w http.ResponseWriter, r *http.Request) {
	searchstring := r.URL.Query().Get("q")
	log.Println(searchstring)
	users, err := SearchUser(searchstring)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		msg, _ := json.Marshal(map[string]string{"message":err.Error()})
		w.Write(msg)
		return
	}
	msg, _ := json.Marshal(&users)
	w.WriteHeader(http.StatusOK)
	w.Write(msg)
}

//Контроллер получения данных поста
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

//Контроллер получения новостной ленты из бд постранично
func GetFeedHandler(w http.ResponseWriter, r *http.Request) {
	var feed []Post
	feed, err := GetFeed()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		msg, _ := json.Marshal(map[string]string{"message": err.Error()})
		w.Write(msg)

	}
	msg, _ := json.Marshal(&feed)
	w.WriteHeader(http.StatusOK)
	w.Write(msg)
}

//Контроллер добавления комментария в бд
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

//Контроллер удаления комментария из бд
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
