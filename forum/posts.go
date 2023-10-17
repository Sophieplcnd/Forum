package forum

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"text/template"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// DEPRECATED
func redirectIfNotLoggedIn(w http.ResponseWriter, r *http.Request) {
	sessionCookie, err := r.Cookie("session")
	if err != nil {
		// If there is an error, it means the session cookie was not found
		// Redirect user to login page
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}

	if sessionCookie.Value == "" {
		// If the session cookie is empty, the user is not logged in
		// Redirect user to login page
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}
}

// CREATE POSTS FUNCTION
func CreatePostHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		categories, err := getCategories()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		var data struct {
			Categories []Category
		}
		data.Categories = categories
		// Serve create post page
		tmpl, err := template.ParseFiles("createPost.html")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		err = tmpl.Execute(w, data)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		return
	}

	// Check session cookie
	sessionID := GetSessionIDFromRequest(r)
	isLoggedIn := sessionID != ""
	if !isLoggedIn {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}

	err := r.ParseForm()
	if err != nil {
		http.Error(w, "Could not parse form", http.StatusBadRequest)
		return
	}

	userId, err := getUserIDFromSessionID(sessionID)
	if err != nil {
		fmt.Printf("failed to get user id from session id: %s\n", err.Error())
		http.Error(w, "cookie not found", http.StatusBadRequest)
		return
	}

	titleContent := r.Form.Get("postTitle")
	postContent := r.Form.Get("postContent")

	if titleContent == "" || postContent == "" {
		fmt.Fprintln(w, "Error - please ensure title and post content fields are not empty!")
		return
	}

	// Get selected categories
	categories := r.Form["postCategories"]

	//added
	dateCreated := time.Now()
	fmt.Println(userId)
	//userIdInt, err := strconv.Atoi(userId)
	if err != nil {
		http.Error(w, "Could not convert", http.StatusInternalServerError)
		return
	}
	// - added placeholders and userintid
	res, err := DB.Exec("INSERT INTO posts (title, user_id, content, created_at) VALUES (?, ?, ?, ?)", titleContent, userId, postContent, dateCreated)
	if err != nil {
		log.Println(err)
		http.Error(w, "Could not create post", http.StatusInternalServerError)
		return
	}

	id, err := res.LastInsertId()
	if err != nil {
		log.Println(err)
		http.Error(w, "Could not create post", http.StatusInternalServerError)
		return
	}

	for _, categoryID := range categories {
		catID, err := strconv.Atoi(categoryID)
		if err != nil {
			log.Println(err)
			http.Error(w, "Could not convert category ID", http.StatusInternalServerError)
			return
		}
		err = attachCategoryToPosts(catID, int(id))
		if err != nil {
			log.Println(err)
			http.Error(w, "Could not attach category to post", http.StatusInternalServerError)
			return
		}
	}

	fmt.Println("Post successfully created!")

	// Redirect the user to the homepage
	http.Redirect(w, r, "/", http.StatusFound)
}

// DEPRECATED use GetSessionIDFromRequest and then getUserIDFromSessionID instead
func GetCookieValue(r *http.Request) (string, string, error) {
	//- indices represent the split to cookie and value
	cookie, err := r.Cookie("session")
	if err != nil {
		return "", "", err
	}

	// TODO work out what is expected to be in the session
	return "", cookie.Value, nil
}

func GetSessionIDFromRequest(r *http.Request) string {
	cookie, _ := r.Cookie("session")
	if cookie == nil {
		return ""
	}
	return cookie.Value
}

// get post ID
func getPostByID(postID string) (*Post, error) {
	//added
	// Adjusted the SELECT query to also get the `dislike_count`
	row := DB.QueryRow("SELECT id, title, content, created_at, likes_count, dislikes_count FROM posts WHERE id = ?", postID)
	if err := row.Err(); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	var post Post
	// Added &post.DislikeCount at the end
	err := row.Scan(&post.ID, &post.Title, &post.Content, &post.Time, &post.LikesCount, &post.DislikeCount)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("post not found")
		}
		return nil, err
	}
	// row := DB.QueryRow("SELECT id, title, content, created_at, likes_count FROM posts WHERE id = ?", postID)
	// var post Post
	// err := row.Scan(&post.ID, &post.Title, &post.Content, &post.Time, &post.LikesCount)
	// if err != nil {
	// 	if err == sql.ErrNoRows {
	// 		return nil, errors.New("post not found")
	// 	}
	// 	return nil, err
	// }
	// Format the datetime string
	t, err := time.Parse("2006-01-02T15:04:05.999999999-07:00", post.Time)
	if err != nil {
		return nil, err
	}
	post.Time = t.Format("January 2, 2006, 15:04:05")
	// make post URLs
	post.URL = "/post/" + post.ID
	return &post, nil
}

func getCategoriesByPostID(postID int) ([]Category, error) {
	categories := []Category{}
	rows, err := DB.Query("SELECT categories.name FROM categories INNER JOIN categories_posts ON categories.id = categories_posts.category_id WHERE categories_posts.post_id = ?", postID)
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var category Category
		err := rows.Scan(&category.Name)
		if err != nil {
			return nil, err
		}
		categories = append(categories, category)
	}

	return categories, nil
}

func getUserIDFromSessionID(sessionID string) (int, error) {
	var userID int
	err := DB.QueryRow("SELECT id FROM users WHERE sessionID = ?", sessionID).Scan(&userID)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, nil
		}
		return 0, fmt.Errorf("failed to get user id from session id: %v", err)
	}

	return userID, nil
}

func getCommentsByPostID(postID string) ([]Comment, error) {
	comments := []Comment{} // creating an empty slice to store comments from the database //i've also added postID and userID to the comment struct
	rows, err := DB.Query("SELECT id, user_id, post_id, content, created_at, commentlikes_count, commentdislikes_count FROM comments WHERE post_id = ?", postID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var comment Comment
		err := rows.Scan(&comment.CommentID, &comment.UserID, &comment.PostID, &comment.Content, &comment.Time, &comment.LikesCount, &comment.DislikeCount)
		if err != nil {
			return nil, err
		}
		t, err := time.Parse("2006-01-02T15:04:05.999999999-07:00", comment.Time)
		if err != nil {
			return nil, err
		}
		comment.Time = t.Format("January 2, 2006, 15:04:05")
		comments = append(comments, comment)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return comments, nil
}

func PostPageHandler(w http.ResponseWriter, r *http.Request) {
	// Get post id from the URL path
	postIDStr := strings.TrimPrefix(r.URL.Path, "/post/")

	postID, err := strconv.Atoi(postIDStr)
	if err != nil {
		http.Error(w, "Invalid post ID", http.StatusBadRequest)
		return
	}
	var comments []Comment

	// Get the post data by calling the getPostByID function or fetching it from the database
	post, err := getPostByID(postIDStr)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	categories, err := getCategoriesByPostID(postID)
	if err != nil {
		fmt.Printf("ERR: %s\n", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	post.Categories = categories

	// Get the likes count from the post variable
	likesCount := post.LikesCount
	dislikeCount := post.DislikeCount

	//get comments by postID -
	comments, err = getCommentsByPostID(postIDStr)
	if err != nil {
		log.Println(err)
		http.Error(w, "Could not fetch comments", http.StatusInternalServerError)
		return
	}

	// Assuming your Post struct has a field named PostID
	var data struct {
		PostID   int
		Post     Post
		Comments []Comment
		Likes    int
		Dislikes int
		Success  bool // Add the Success field to indicate if the comment was successfully posted
	}

	data.PostID = postID
	data.Post = *post // Use the dereferenced post pointer
	data.Comments = comments
	data.Success = r.URL.Query().Get("success") == "1"
	data.Likes = likesCount
	data.Dislikes = dislikeCount

	fmt.Println(likesCount, "likes count")
	fmt.Println(dislikeCount, "dislike count")

	tmpl, err := template.ParseFiles("postPage.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Render the template with the data
	// ! error flag
	err = tmpl.ExecuteTemplate(w, "postPage.html", data)
	if err != nil {
		http.Error(w, "Internal Server Error - posts", http.StatusInternalServerError)
		return
	}
}
