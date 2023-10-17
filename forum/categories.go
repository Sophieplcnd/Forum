package forum

func getCategories() ([]Category, error) {
	rows, err := DB.Query("SELECT id, name FROM categories")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	categories := make([]Category, 0)
	for rows.Next() {
		var category Category
		if err := rows.Scan(&category.ID, &category.Name); err != nil {
			return nil, err
		}
		categories = append(categories, category)
	}
	return categories, nil
}

func attachCategoryToPosts(categoryID, postID int) error {
	_, err := DB.Exec("INSERT INTO categories_posts (post_id, category_id) VALUES (?, ?)", postID, categoryID)
	return err
}
