package main

import (
	"github.com/gorilla/mux"
	"net/http"
	"time"
	"log"
	"os"
	"github.com/creamdog/gonfig"
)

var (
	connstr  string
	port     string
	bottoken string
)

func main() {
	Init()
	Serve()
}
//Инициализация приложения
func Init() {
	f, err := os.Open("config.yml")
	if err != nil {
		log.Println(err)
	}
	defer f.Close();
	config, err := gonfig.FromYml(f)
	if err != nil {
		log.Println(err)
	}
	port, _ = config.GetString("port", "9000")
	user, _ := config.GetString("user", "user")
	pass, _ := config.GetString("pass", "pass")
	dbname, _ := config.GetString("dbname", "test")
	connstr = user + ":" + pass + "@/" + dbname + "?charset=utf8&parseTime=True&loc=Local"
	InitDB()
}
//Создание экземпляра сервера и определение endpoint-ов роутера.
func Serve() {
	r := mux.NewRouter()
	r.HandleFunc("/auth", TelegramAuthHandler).Methods("GET")
	r.HandleFunc("/signin", LoginViewHandler).Methods("GET")
	r.HandleFunc("/user/{id}", GetUserHandler).Methods("GET")
	r.HandleFunc("/post", AddPostHandler).Methods("PUT")
	r.HandleFunc("/post", DeletePostHandler).Methods("DELETE")
	r.HandleFunc("/post/{id}", GetPostHandler).Methods("GET")
	r.HandleFunc("/feed", GetFeedHandler).Methods("GET")
	r.HandleFunc("/comment", AddCommentHandler).Methods("PUT")
	r.HandleFunc("/comment", DeleteCommentHandler).Methods("DELETE")
	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	srv := &http.Server{
		Addr:         "0.0.0.0:" + port,
		WriteTimeout: time.Second * 15,
		ReadTimeout:  time.Second * 15,
		IdleTimeout:  time.Second * 15,
		Handler:      r,
	}
	log.Print("Listening on " + port + "...")
	log.Fatal(srv.ListenAndServe())
}