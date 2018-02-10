package main

import (
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	_ "github.com/go-sql-driver/mysql"
	"log"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"strconv"
	"sort"
	"crypto/sha1"
	"net/http"
	"strings"
	"encoding/json"
)

var db, _ = gorm.Open(connstr)

type User struct {
	gorm.Model
	FirstName string `gorm:"index"`
	LastName  string `gorm:"index"`
	UserName  string `gorm:"index"`
	PhotoUrl  string `gorm:"size:3000"`
	Hash      string `gorm:"index"`
	AuthDate  string
	Posts     []Post
	Role      string
}

type Post struct {
	gorm.Model
	Title         string `gorm:"size:140"`
	UserID        uint   `gorm:"index"`
	Url           string `gorm:"size:3000"`
	ImgUrl        string `gorm:"size:3000"`
	LikesCount    int
	ViewsCount    int
	CommentsCount int
	Comments      []Comment
}

type Comment struct {
	gorm.Model
	Text   string `gorm:"size 300; index"`
	UserID uint   `gorm:"index"`
	PostID uint   `gorm:"index"`
}

type Like struct {
	gorm.Model
	PostID uint `gorm:"index"`
	UserID uint `gorm:"index"`
}

//Инициализация БД
func InitDB() {
	var err error
	db, err = gorm.Open("mysql", connstr)
	if err != nil {
		log.Fatal(err)
	}
	db.AutoMigrate(&User{}, &Post{}, &Comment{})
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
	u.UserName = user.UserName
	u.AuthDate = user.AuthDate
	u.FirstName = user.FirstName
	u.Hash = user.Hash
	return db.Save(&u).Error
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
			Title    string `json:"title"`
			ImageURL string `json:"image_url"`
			Views    int    `json:"views"`
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
	return post, nil
}

//Получение данных о посте из бд
func GetPost(id uint) (Post, error) {
	var post Post
	if err := db.Where("id = ?", id).First(&post).Error; err != nil {
		return Post{}, err
	}
	if err := db.Where("post_id = ?", id).Find(&post.Comments).Error; err != nil {
		return Post{}, err
	}
	return post, nil
}

//Удаление поста из бд
func DeletePost(id uint) error {
	return db.Where("id = ?", id).Delete(&Post{}).Error
}

//Получение новостной ленты из бд постранично
func GetFeed(page int) ([]Post, error) {
	var posts []Post
	if err := db.Order("created_at desc").Offset(page * 10).Limit(10).Find(&posts).Error; err != nil {
		return []Post{}, err
	}
	return posts, nil
}

//Добавление комментария в бд
func AddComment(comment Comment) error {
	return db.Create(&comment).Error
}

//Удаление комментария из бд
func DeleteComment(id uint) error {
	return db.Where("id = ?", id).Delete(&Comment{}).Error
}

func ValidateAuth(hash string) {
	var user User
	if err := db.Where("hash=?", hash).First(&user).Error; err != nil {
		log.Println(err)
	}
	log.Println(user)
	var data = make(map[string]string)
	var keys []string
	if user.AuthDate != "" {
		data["auth_date"] = user.AuthDate
	}
	if user.FirstName != "" {
		data["first_name"] = user.FirstName
	}
	if user.ID > 0 {
		data["id"] = strconv.Itoa(int(user.ID))
	}
	if user.LastName != "" {
		data["last_name"] = user.LastName
	}
	if user.PhotoUrl != "" {
		data["photo_url"] = user.PhotoUrl
	}
	if user.UserName != "" {
		data["username"] = user.UserName
	}
	var message string
	for key := range data {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		message += key + "=" + data[key] + "\n"
	}
	if len(message) > 2 {
		message = message[:len(message)-1]
	}
}

func ComputeHmac256(message string, secret string) string {
	key := []byte("cea09579535a26a3970e796eb48da1c6f17ce1a3a03f8f783abfe28869d8f065")
	h := hmac.New(sha256.New, key)
	h.Write([]byte(message))
	return hex.EncodeToString(h.Sum(nil))
}

func SHA256(secret string) []byte {
	key := []byte(secret)
	h := sha1.New()
	h.Write([]byte(key))
	return h.Sum(nil)
}
