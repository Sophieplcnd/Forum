package forum

import (
	"database/sql"
	"fmt"
	"net/http"
	"text/template"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

func executePosts() ([]Post, error) {
	var posts []Post //local struct - don't change as it duplicates the posts for some reason.
	rows, err := DB.Query("SELECT id, title, content, created_at FROM posts")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var post Post
		err := rows.Scan(&post.ID, &post.Title, &post.Content, &post.Time)
		if err != nil {
			return nil, err
		}
		// Format the datetime string
		t, err := time.Parse("2006-01-02T15:04:05.999999999-07:00", post.Time)
		if err != nil {
			return nil, err
		}
		post.Time = t.Format("January 2, 2006, 15:04:05")
		// make post URLs
		post.URL = "/post/" + post.ID
		posts = append(posts, post)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	// reverse posts
	posts = reverse(posts)
	return posts, nil
}

func getPostsLiked(userID int) ([]Post, error) {
	var posts []Post //local struct - don't change as it duplicates the posts for some reason.

	rows, err := DB.Query(`
SELECT id, title, content, created_at FROM posts WHERE id IN(
	SELECT post_id FROM postlikes WHERE user_id = 1 AND type = 1
);`, userID)
	if err != nil && err != sql.ErrNoRows {
		return nil, fmt.Errorf("failed to query liked posts: %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		var post Post
		err := rows.Scan(&post.ID, &post.Title, &post.Content, &post.Time)
		if err != nil {
			return nil, err
		}
		// Format the datetime string
		t, err := time.Parse("2006-01-02T15:04:05.999999999-07:00", post.Time)
		if err != nil {
			return nil, err
		}
		post.Time = t.Format("January 2, 2006, 15:04:05")
		// make post URLs
		post.URL = "/post/" + post.ID
		posts = append(posts, post)
	}
	// reverse posts
	posts = reverse(posts)
	return posts, nil
}

func getPostsUser(userID int) ([]Post, error) {
	var posts []Post //local struct - don't change as it duplicates the posts for some reason.

	rows, err := DB.Query(`SELECT id, title, content, created_at FROM posts WHERE user_id = $1`, userID)
	if err != nil && err != sql.ErrNoRows {
		return nil, fmt.Errorf("failed to query users posts: %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		var post Post
		err := rows.Scan(&post.ID, &post.Title, &post.Content, &post.Time)
		if err != nil {
			return nil, err
		}
		// Format the datetime string
		t, err := time.Parse("2006-01-02T15:04:05.999999999-07:00", post.Time)
		if err != nil {
			return nil, err
		}
		post.Time = t.Format("January 2, 2006, 15:04:05")
		// make post URLs
		post.URL = "/post/" + post.ID
		posts = append(posts, post)
	}
	// reverse posts
	posts = reverse(posts)
	return posts, nil
}

// reverse posts (latest first)
func reverse(s []Post) []Post {
	//runes := []rune(s)
	length := len(s)
	for i, j := 0, length-1; i < j; i, j = i+1, j-1 {
		s[i], s[j] = s[j], s[i]
	}
	return s
}

// serve homepage
func HomeHandler(w http.ResponseWriter, r *http.Request) {
	// Check if the user is already logged in
	sessionID := GetSessionIDFromRequest(r)
	isLoggedIn := sessionID != ""

	posts, err := executePosts()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	likedPosts := []Post{}
	yourPosts := []Post{}
	if isLoggedIn {
		userID, err := getUserIDFromSessionID(sessionID)
		if err != nil {
			fmt.Printf("ERR: %s\n", err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		likedPosts, err = getPostsLiked(userID)
		if err != nil {
			fmt.Printf("ERR: %s\n", err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		yourPosts, err = getPostsUser(userID)
		if err != nil {
			fmt.Printf("ERR: %s\n", err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	data := HomePageData{
		Posts:      posts,
		LikedPosts: likedPosts,
		YourPosts:  yourPosts,
		IsLoggedIn: isLoggedIn, // Pass the IsLoggedIn information to the template
	}

	tmpl, err := template.ParseFiles("home.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = tmpl.Execute(w, data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// handle filtered posts
// handle filtered posts
func FilteredPostsHandler(w http.ResponseWriter, r *http.Request) {
	category := r.URL.Query().Get("category")

	// Retrieve the posts based on the selected category
	filteredPosts, err := getPostsByCategory(category)
	if err != nil {
		http.Error(w, "Could not fetch posts", http.StatusInternalServerError)
		return
	}

	var data struct {
		Category      string
		FilteredPosts []Post // Use a slice of Post
	}

	data.Category = category
	data.FilteredPosts = filteredPosts

	tmpl, err := template.ParseFiles("filteredPosts.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Render the template with the filtered posts data
	err = tmpl.ExecuteTemplate(w, "filteredPosts.html", data)
	if err != nil {
		http.Error(w, "Internal Server Error - homepage", http.StatusInternalServerError)
		return
	}
}

// retrieve posts by their category
// func getPostsByCategory(category string) (*Post, error) {
// 	//added
// 	rows := DB.QueryRow("SELECT id, title, content, created_at FROM posts WHERE category_id = ?", category)
// 	var posts Post
// 	err := rows.Scan(&posts.id, &posts.Title, &posts.Content, &posts.Time)
// 	if err != nil {
// 		return nil, err
// 	}
// 	// Format the datetime string
// 	t, err := time.Parse("2006-01-02T15:04:05.999999999-07:00", posts.Time)
// 	if err != nil {
// 		return nil, err
// 	}
// 	posts.Time = t.Format("January 2, 2006, 15:04:05")
// 	// make post URLs
// 	posts.URL = "/post/" + posts.id
// 	return &posts, nil
// }

// retrieve posts by their category
func getPostsByCategory(category string) ([]Post, error) {
	rows, err := DB.Query(`
SELECT id, title, content, created_at FROM posts WHERE id IN(
	SELECT post_id FROM categories_posts WHERE category_id = (
		SELECT id FROM categories WHERE name = ?
	)
)
`, category)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var posts []Post

	for rows.Next() {
		var post Post
		err := rows.Scan(&post.ID, &post.Title, &post.Content, &post.Time)
		if err != nil {
			return nil, err
		}

		// Format the datetime string
		t, err := time.Parse("2006-01-02T15:04:05.999999999-07:00", post.Time)
		if err != nil {
			return nil, err
		}
		post.Time = t.Format("January 2, 2006, 15:04:05")

		// make post URLs
		post.URL = "/post/" + post.ID

		posts = append(posts, post)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return posts, nil
}
