package main

import (
	"juhena-forum/forum"
	"log"
	"net/http"

	_ "github.com/mattn/go-sqlite3"
)

func main() {
	// var err error
	// forum.DB, err = sql.Open("sqlite3", "./database/database.db")
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// defer forum.DB.Close()

	// initialise database
	forum.Init()

	http.HandleFunc("/register", forum.RegisterHandler)
	http.HandleFunc("/login", forum.LoginHandler)
	http.HandleFunc("/", forum.HomeHandler)
	http.HandleFunc("/create-post", forum.CreatePostHandler)
	http.HandleFunc("/post/", forum.PostPageHandler)
	http.HandleFunc("/post-comment/", forum.PostCommentHandler)
	http.HandleFunc("/post-like/", forum.HandleLikesDislikes)
	http.HandleFunc("/comment-like/", forum.CommentLikesHandler)
	http.HandleFunc("/filtered-posts", forum.FilteredPostsHandler)
	http.HandleFunc("/logout", forum.LogoutHandler)
	// http.HandleFunc("/display-dislike-count", forum.DisplayDislikeCountHandler)

	forum.Logger.Println("Server running on port 8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
	forum.Shutdown()
}
