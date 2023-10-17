package forum

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// CREATE COMMENTS FUNCTION
func PostCommentHandler(w http.ResponseWriter, r *http.Request) {
	// Check session cookie

	// Get postID from URL path
	postIDStr := strings.TrimPrefix(r.URL.Path, "/post-comment/")
	postID, err := strconv.Atoi(postIDStr)
	if err != nil {
		http.Error(w, "Invalid post ID", http.StatusBadRequest)
		return
	}

	// Retrieve the sessionID and userID from the cookie
	sessionID := GetSessionIDFromRequest(r)
	if sessionID == "" {
		http.Error(w, "session not found", http.StatusBadRequest)
		return
	}
	userId, err := getUserIDFromSessionID(sessionID)
	if err != nil {
		http.Error(w, "cookie not found", http.StatusBadRequest)
		return
	}

	err = r.ParseForm()
	if err != nil {
		http.Error(w, "Could not parse form", http.StatusBadRequest)
		return
	}

	postComment := r.Form.Get("commentContent")
	fmt.Println(postComment)

	if postComment == "" {
		fmt.Fprintln(w, "Error - please ensure comment box is not empty!")
		return
	}

	dateCreated := time.Now()

	// Use userID and postID to create a new comment
	//user_ID gets excecuted to the database
	_, err = DB.Exec("INSERT INTO comments (post_id, user_id, content, created_at, updated_at) VALUES (?, ?, ?, ?, ?)",
		postID, userId, postComment, dateCreated, dateCreated)
	if err != nil {
		log.Println(err)
		http.Error(w, "Could not post comment", http.StatusInternalServerError)
		return
	}

	fmt.Println("Comment successfully posted!")

	http.Redirect(w, r, "/post/"+postIDStr, http.StatusFound)
}
