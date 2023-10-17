package forum

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

// var postID int
// var userID string

func getPostID(w http.ResponseWriter, r *http.Request) (int, error) {
	// Extract post ID from URL
	postIDStr := strings.TrimPrefix(r.URL.Path, "/post-like/")
	postID, err := strconv.Atoi(postIDStr)
	if err != nil { //added JB
		return 0, fmt.Errorf("invalid post ID: %w", err)
	}
	return postID, nil

	// if err != nil {
	// 	http.Error(w, "Invalid post ID", http.StatusBadRequest)
	// }
	// return postID
}

// DEPRECATE
func getUserID(w http.ResponseWriter, r *http.Request) string {
	// Get user ID from session cookie
	_, userID, err := GetCookieValue(r)
	if err != nil {
		http.Error(w, "Cookie not found", http.StatusBadRequest)
	}
	return userID
}

// Handler for handling like and dislike actions
func HandleLikesDislikes(w http.ResponseWriter, r *http.Request) {
	// Check session cookie

	// Retrieve the sessionID and userID from the cookie
	sessionID := GetSessionIDFromRequest(r)
	isLoggedIn := sessionID != ""
	if !isLoggedIn {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}
	userID, err := getUserIDFromSessionID(sessionID)
	if err != nil {
		fmt.Printf("ERR: %s\n", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Get postID
	postID, err := getPostID(w, r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Parse form data to retrieve the 'action' field
	r.ParseForm()
	action := r.FormValue("action")

	// Check if the user has already liked/disliked the post
	_, _, err = checkUserLikeDislike(userID, postID)
	if err != nil {
		http.Error(w, "Error checking user like/dislike", http.StatusInternalServerError)
		return
	}

	// Handle the like/dislike action based on user's previous interaction
	if action == "like" {
		err = addLike(userID, postID)
		if err != nil {
			http.Error(w, "Error adding like", http.StatusInternalServerError)
			return
		}
	} else if action == "dislike" {
		err = addDislike(userID, postID)
		if err != nil {
			http.Error(w, "Error adding dislike", http.StatusInternalServerError)
			return
		}
	}

	// Update the total likes and dislikes counts after the action is handled
	err = addTotalLikesDislikes(postID)
	if err != nil {
		http.Error(w, "Error updating total likes and dislikes", http.StatusInternalServerError)
		return
	}

	// Redirect back to the post page
	postIDStr := strings.TrimPrefix(r.URL.Path, "/post-like/")
	http.Redirect(w, r, "/post/"+postIDStr, http.StatusSeeOther)
}

func checkUserLikeDislike(userID int, postID int) (liked bool, disliked bool, err error) {
	// Execute a query to check if the user has liked or disliked the post
	row := DB.QueryRow("SELECT COUNT(*) FROM postlikes WHERE user_id = ? AND post_id = ? AND type = 1", userID, postID)
	var likeCount int
	if err := row.Scan(&likeCount); err != nil {
		return false, false, err
	}
	liked = likeCount > 0

	row = DB.QueryRow("SELECT COUNT(*) FROM postlikes WHERE user_id = ? AND post_id = ? AND type = -1", userID, postID)
	var dislikeCount int
	if err := row.Scan(&dislikeCount); err != nil {
		return false, false, err
	}
	disliked = dislikeCount > 0

	// check for data entries that have not been liked or disliked
	row = DB.QueryRow("SELECT COUNT(*) FROM postlikes WHERE user_id = ? AND post_id = ? AND type = 0", userID, postID)
	var neitherCount int
	if err := row.Scan(&neitherCount); err != nil {
		return false, false, err
	}

	// If there is a "neither" entry, treat it as a like and a dislike
	if neitherCount > 0 {
		liked = true
		disliked = true
	}

	return liked, disliked, nil
}

func addLike(userID int, postID int) error {
	// check if user has previously interacted with post
	existingLike, existingDislike, err := checkUserLikeDislike(userID, postID)
	if err != nil {
		return err
	}

	if existingDislike {
		// Replace value '-1' with '1'
		_, err = DB.Exec("UPDATE postlikes SET type = 1 WHERE user_id = ? AND post_id = ?", userID, postID)
		if err != nil {
			return err
		}
		fmt.Println("Replaced Dislike with Like")
	} else if existingLike {
		// Toggle value '1' to '0'
		_, err = DB.Exec("UPDATE postlikes SET type = 0 WHERE user_id = ? AND post_id = ?", userID, postID)
		if err != nil {
			return err
		}
		fmt.Println("Toggled Like to Neither")
	} else {
		// Insert new entry with value '1'
		_, err = DB.Exec("INSERT INTO postlikes (user_id, post_id, type) VALUES (?, ?, 1)", userID, postID)
		if err != nil {
			return err
		}
		fmt.Println("Added Like")
	}

	// Update the likes_count in the posts table
	_, err = DB.Exec("UPDATE posts SET likes_count = likes_count + 1 WHERE id = ?", postID)
	if err != nil {
		return err
	}
	fmt.Println("Updated likes_count for the post")

	return nil

}

func addDislike(userID int, postID int) error {

	existingLike, existingDislike, err := checkUserLikeDislike(userID, postID)
	if err != nil {
		return err
	}

	if existingLike {
		// Replace value '1' with '-1'
		_, err = DB.Exec("UPDATE postlikes SET type = -1 WHERE user_id = ? AND post_id = ?", userID, postID)
		if err != nil {
			return err
		}
		fmt.Println("Replaced Like with Dislike")
	} else if existingDislike {
		// Toggle value '-1' to '0'
		_, err = DB.Exec("UPDATE postlikes SET type = 0 WHERE user_id = ? AND post_id = ?", userID, postID)
		if err != nil {
			return err
		}
		fmt.Println("Toggled Dislike to Neither")
	} else {
		// Insert new entry with value '-1'
		_, err = DB.Exec("INSERT INTO postlikes (user_id, post_id, type) VALUES (?, ?, -1)", userID, postID)
		if err != nil {
			return err
		}
		fmt.Println("Added Dislike")
	}
	return nil
}

// create a function to add total likes from the database postlikes table
func getTotalLikes(postID int) (int, error) {
	//from the postlikes table, get the total number of likes for a post
	row := DB.QueryRow("SELECT COUNT(*) FROM postlikes WHERE post_id = ? AND type = 1", postID)
	var likeCount int
	if err := row.Scan(&likeCount); err != nil {
		return 0, err
	}
	return likeCount, nil
}

// create a function to add total dislikes from the database postlikes table
func getTotalDislikes(postID int) (int, error) {
	//from the postlikes table, get the total number of dislikes for a post
	row := DB.QueryRow("SELECT COUNT(*) FROM postlikes WHERE post_id = ? AND type = -1", postID)
	var dislikeCount int
	if err := row.Scan(&dislikeCount); err != nil {
		return 0, err
	}
	return dislikeCount, nil
}

// adds total likes to the likes_count column in the posts table and total dislikes to the dislikes_count column in the posts table
func addTotalLikesDislikes(postID int) error {
	//get the total number of likes for a post
	likeCount, err := getTotalLikes(postID)
	if err != nil {
		return err
	}
	// total number of dislikes for a post
	dislikeCount, err := getTotalDislikes(postID)
	if err != nil {
		return err
	}
	//total number of likes to the likes_count column in the posts table
	_, err = DB.Exec("UPDATE posts SET likes_count = ? WHERE id = ?", likeCount, postID)
	fmt.Println(likeCount, "likes count database")
	if err != nil {
		return err
	}
	//the total number of dislikes to the dislikes_count column and return the count
	_, err = DB.Exec("UPDATE posts SET dislikes_count = ? WHERE id = ?", dislikeCount, postID)
	fmt.Println(dislikeCount, "dislikes count database")
	if err != nil {
		return err
	}
	return nil

}
