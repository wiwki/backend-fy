package main

import (
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	_ "github.com/go-sql-driver/mysql"
	"log"
	"net/http"
	"strings"
	"encoding/json"
	"time"
	"strconv"
)

var db, _ = gorm.Open(connstr)

//Все сущности
type Model struct {
	ID        uint       `gorm:"primary" json:"id"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at"`
}

type User struct {
	Model
	FirstName string    `gorm:"index" json:"first_name"`
	LastName  string    `gorm:"index" json:"last_name"`
	Username  string    `gorm:"index" json:"username"`
	FullName  string    `gorm:"index" json:"full_name"`
	PhotoUrl  string    `gorm:"size:3000" json:"photo_url"`
	Hash      string    `gorm:"index" json:"hash"`
	AuthDate  string    `json:"auth_date"`
	Posts     []Post    `json:"posts"`
	Role      string    `json:"role"`
	Sex       string    `json:"sex"`
	Info      string    `json:"info" gorm:"size:3000"`
	Birthday  time.Time `json:"birthday"`
	Age       uint      `json:"age"`
	CommentID uint      `json:"comment_id"`
}

type Post struct {
	Model
	Title         string    `gorm:"size:140" json:"title"`
	Caption       string    `gorm:"size:3000" json:"caption"`
	UserID        uint      `gorm:"index; column:user_id" json:"user_id"`
	Url           string    `gorm:"size:3000" json:"url"`
	ImgUrl        string    `gorm:"size:3000" json:"img_url"`
	LikesCount    int       `json:"likes_count"`
	ViewsCount    int       `json:"views_count"`
	CommentsCount int       `json:"comments_count"`
	Comments      []Comment `json:"comments"`
	Special       string    `json:"special"`
	Type          string    `json:"type"`
}

type Comment struct {
	Model
	Text   string `gorm:"size 500; index" json:"text"`
	User   User   `json:"user"`
	PostID uint   `gorm:"index" json:"post_id"`
	Date   string `json:"date"`
}

type Like struct {
	gorm.Model
	PostID uint `gorm:"index" json:"post_id"`
	UserID uint `gorm:"index" json:"user_id"`
}

type Chat struct {
	Model
	ImgUrl string `gorm:"size 3000" json:"img_url"`
	Title  string `gorm:"size:200" json:"title"`
	Url    string `gorm:"size:3000" json:"url"`
}

type FeedView struct {
	Posts []Post
}

type ChatsView struct {
	Chats []Chat
}

type UsersView struct {
	Users []User
}

//Инициализация БД
func InitDB() {
	var err error
	db, err = gorm.Open("mysql", connstr)
	if err != nil {
		log.Fatal(err)
	}
	db.AutoMigrate(&User{}, &Post{}, &Comment{}, &Chat{}, &Like{})
}

//Получение данных о юзере из бд
func GetUser(id uint) (User, error) {
	var user User
	if err := db.Where("id = ?", id).First(&user).Error; err != nil {
		return User{}, err
	}
	if err := db.Where("user_id = ?", id).First(&user.Posts).Error; err != nil {
		return User{}, err
	}
	return user, nil
}

//Создание или редактирование юзера в бд
func SaveUser(user User) error {
	var u User
	if db.Where("id = ?", user.ID).Find(&u).RecordNotFound() {
		return db.Create(&user).Error
	}
	u.LastName = user.LastName
	u.PhotoUrl = user.PhotoUrl
	u.Username = user.Username
	u.AuthDate = user.AuthDate
	u.FirstName = user.FirstName

	u.Hash = user.Hash
	return db.Save(&u).Error
}

//Поиск юзеров
func SearchUser(searchstring string) ([]User, error) {
	var users, u []User
	templ := "%" + searchstring + "%"
	if err := db.Where("full_name LIKE ? or username LIKE ? or info LIKE ?", templ, templ, templ).Find(&users).Error; err != nil {
		return []User{}, err
	}
	for i := range users {
		if strings.HasPrefix(strings.ToLower(users[i].Username), strings.ToLower(searchstring)) ||
			strings.HasPrefix(strings.ToLower(users[i].FirstName), strings.ToLower(searchstring)) ||
			strings.HasPrefix(strings.ToLower(users[i].LastName), strings.ToLower(searchstring)) ||
			strings.HasPrefix(strings.ToLower(users[i].FullName), strings.ToLower(searchstring)) ||
			strings.Contains(strings.ToLower(users[i].Info), strings.ToLower(searchstring)) {
			u = append(u, users[i])
		}
	}
	return u, nil
}

//Редатирование данных юзера
func EditUser(u User) error {
	var user User
	if err := db.Where("id = ?", u.ID).Find(&user).Error; err != nil {
		return err
	}
	user.Sex = u.Sex
	user.Info = u.Info
	user.Birthday = u.Birthday
	user.Age = u.Age
	if err := db.Save(&user).Error; err != nil {
		return err
	}
	return nil
}

//Создание поста в бд
func AddPost(post Post) error {
	return db.Create(&post).Error
}

//Парсинг статьи(поста) в Telegra.ph
func ParsePost(url string) (Post, error) {
	var post Post
	url = "https://api.telegra.ph/getPage/" +
		strings.Split(url, "/")[len(strings.Split(url, "/"))-1] +
		"?return_content=true"
	var TelegraphPage struct {
		Result struct {
			Title       string `json:"title"`
			ImageURL    string `json:"image_url"`
			URL         string `json:"url"`
			Views       int    `json:"views"`
			Description string `json:"description"`
		}
	}
	resp, err := http.Get(url)
	if err != nil {
		return Post{}, err
	}
	decoder := json.NewDecoder(resp.Body)
	err = decoder.Decode(&TelegraphPage)
	if err != nil {
		return Post{}, err
	}
	post.Title = TelegraphPage.Result.Title
	post.ImgUrl = TelegraphPage.Result.ImageURL
	post.ViewsCount = TelegraphPage.Result.Views
	post.Url = TelegraphPage.Result.URL
	post.Caption = TelegraphPage.Result.Description
	return post, nil
}

//Получение данных о посте из бд
func GetPost(id uint) (Post, error) {
	var post Post
	if err := db.Where("id = ?", id).First(&post).Error; err != nil {
		return Post{}, err
	}
	db.Where("post_id = ?", id).Order("created_at desc").Find(&post.Comments)
	for i := range post.Comments {
		db.Where("comment_id = ?", post.Comments[i].ID).First(&post.Comments[i].User)
	}
	post.CommentsCount = len(post.Comments)
	db.Where("post_id = ?", id).Find(&Like{}).Count(&post.LikesCount)
	return post, nil
}

//Удаление поста из бд
func DeletePost(id uint) error {
	return db.Where("id = ?", id).Delete(&Post{}).Error
}

//Получение новостной ленты из бд постранично
func GetFeed() ([]Post, error) {
	var posts []Post
	if err := db.Order("created_at desc").Find(&posts).Error; err != nil {
		return []Post{}, err
	}
	for i := range posts {
		db.Where("post_id = ?", posts[i].ID).Find(&[]Comment{}).Count(&posts[i].CommentsCount)
	}
	return posts, nil
}

//Добавление комментария в бд
func AddComment(comment Comment) error {
	comment.Date = strconv.Itoa(time.Now().Day()) + " " + time.Now().Month().String() + " " + strconv.Itoa(time.Now().Year())
	db.Where("id = ?", comment.User.ID).First(&comment.User)
	return db.Create(&comment).Error
}

//Удаление комментария из бд
func DeleteComment(id uint) error {
	return db.Where("id = ?", id).Delete(&Comment{}).Error
}

//Получение списка сайтов из бд
func GetAllChats() ([]Chat, error) {
	var chats []Chat
	err := db.Order("title asc").Find(&chats).Error
	return chats, err
}

//Добавление чата
func AddChat(chat Chat) error {
	return db.Create(&chat).Error
}
