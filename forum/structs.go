package forum

import (
	"database/sql"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

var DB *sql.DB

// struct for individual posts
type Post struct {
	ID           string
	Title        string
	Content      string
	Time         string
	LikesCount   int // added JB
	DislikeCount int // added JB
	URL          string
	Categories   []Category
	// Author  string
}

type Category struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

// struct for comments
type Comment struct {
	UserID       string //
	PostID       string //
	CommentID    string
	Content      string
	Time         string
	LikesCount   int
	DislikeCount int
	// Author  string
}

// struct for posts
type HomePageData struct {
	Posts      []Post // Replace with your actual Post type
	LikedPosts []Post
	YourPosts  []Post
	IsLoggedIn bool // Add this field to indicate whether the user is logged in
}

type PostPageData struct {
	Post     *Post
	Comments []Comment // added
	Success  bool      // For displaying the success message
}

// struct to contain comments
type CommentsData struct {
	Comment []Comment
}

// var comments []Comment

//var posts []Post

// initialise DB
func Init() {
	var err error
	DB, err = sql.Open("sqlite3", "./database/database.db")
	if err != nil {
		log.Fatal(err)
	}
	// defer DB.Close()
}

func Shutdown() {
	if DB != nil {
		if err := DB.Close(); err != nil {
			log.Fatal(err)
		}
	}
}
