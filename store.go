package main

import (
	"database/sql"
	"fmt"
	"strings"
	"time"
)

type Store struct {
	db *sql.DB
}

// --- Users ---

func (s *Store) CreateUser(username, displayName, passwordHash string) (int64, error) {
	if displayName == "" {
		displayName = username
	}
	res, err := s.db.Exec(
		"INSERT INTO users (username, display_name, password_hash) VALUES (?, ?, ?)",
		username, displayName, passwordHash,
	)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (s *Store) GetUserByUsername(username string) (*User, error) {
	u := &User{}
	err := s.db.QueryRow(
		"SELECT id, username, display_name, password_hash, is_admin, created_at FROM users WHERE username = ?",
		username,
	).Scan(&u.ID, &u.Username, &u.DisplayName, &u.PasswordHash, &u.IsAdmin, &u.CreatedAt)
	if err != nil {
		return nil, err
	}
	return u, nil
}

func (s *Store) GetUserByID(id int64) (*User, error) {
	u := &User{}
	err := s.db.QueryRow(
		"SELECT id, username, display_name, password_hash, is_admin, created_at FROM users WHERE id = ?",
		id,
	).Scan(&u.ID, &u.Username, &u.DisplayName, &u.PasswordHash, &u.IsAdmin, &u.CreatedAt)
	if err != nil {
		return nil, err
	}
	return u, nil
}

func (s *Store) ListUsers() ([]User, error) {
	rows, err := s.db.Query("SELECT id, username, display_name, is_admin, created_at FROM users ORDER BY username")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []User
	for rows.Next() {
		var u User
		if err := rows.Scan(&u.ID, &u.Username, &u.DisplayName, &u.IsAdmin, &u.CreatedAt); err != nil {
			return nil, err
		}
		users = append(users, u)
	}
	return users, nil
}

func (s *Store) EnsureAdmin() {
	var count int
	s.db.QueryRow("SELECT COUNT(*) FROM users WHERE is_admin = 1").Scan(&count)
	if count == 0 {
		// Promote first user if exists
		s.db.Exec("UPDATE users SET is_admin = 1 WHERE id = (SELECT MIN(id) FROM users)")
	}
}

func (s *Store) UpdateUser(u *User) error {
	_, err := s.db.Exec(
		"UPDATE users SET username = ?, display_name = ?, is_admin = ? WHERE id = ?",
		u.Username, u.DisplayName, u.IsAdmin, u.ID,
	)
	return err
}

func (s *Store) UpdateUserPassword(id int64, passwordHash string) error {
	_, err := s.db.Exec("UPDATE users SET password_hash = ? WHERE id = ?", passwordHash, id)
	return err
}

func (s *Store) DeleteUser(id int64) error {
	_, err := s.db.Exec("DELETE FROM users WHERE id = ?", id)
	return err
}

// --- Sessions ---

func (s *Store) CreateSession(userID int64, token string, expiresAt time.Time) error {
	_, err := s.db.Exec(
		"INSERT INTO sessions (user_id, token, expires_at) VALUES (?, ?, ?)",
		userID, token, expiresAt,
	)
	return err
}

func (s *Store) GetSessionByToken(token string) (*Session, error) {
	sess := &Session{}
	err := s.db.QueryRow(
		"SELECT id, user_id, token, expires_at FROM sessions WHERE token = ?",
		token,
	).Scan(&sess.ID, &sess.UserID, &sess.Token, &sess.ExpiresAt)
	if err != nil {
		return nil, err
	}
	return sess, nil
}

func (s *Store) DeleteSession(token string) error {
	_, err := s.db.Exec("DELETE FROM sessions WHERE token = ?", token)
	return err
}

func (s *Store) CleanExpiredSessions() error {
	_, err := s.db.Exec("DELETE FROM sessions WHERE expires_at < ?", time.Now())
	return err
}

// --- Categories ---

func (s *Store) CreateCategory(name, slug string) (int64, error) {
	res, err := s.db.Exec("INSERT INTO categories (name, slug) VALUES (?, ?)", name, slug)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (s *Store) FindOrCreateCategory(name string) (int64, error) {
	slug := slugify(name)
	var id int64
	err := s.db.QueryRow("SELECT id FROM categories WHERE slug = ?", slug).Scan(&id)
	if err == nil {
		return id, nil
	}
	return s.CreateCategory(name, slug)
}

func (s *Store) ListCategories() ([]Category, error) {
	rows, err := s.db.Query("SELECT id, name, slug FROM categories ORDER BY name")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var cats []Category
	for rows.Next() {
		var c Category
		if err := rows.Scan(&c.ID, &c.Name, &c.Slug); err != nil {
			return nil, err
		}
		cats = append(cats, c)
	}
	return cats, nil
}

func (s *Store) GetCategoryBySlug(slug string) (*Category, error) {
	c := &Category{}
	err := s.db.QueryRow("SELECT id, name, slug FROM categories WHERE slug = ?", slug).
		Scan(&c.ID, &c.Name, &c.Slug)
	if err != nil {
		return nil, err
	}
	return c, nil
}

func (s *Store) GetCategoryByID(id int64) (*Category, error) {
	c := &Category{}
	err := s.db.QueryRow("SELECT id, name, slug FROM categories WHERE id = ?", id).
		Scan(&c.ID, &c.Name, &c.Slug)
	if err != nil {
		return nil, err
	}
	return c, nil
}

// --- Tags ---

func (s *Store) CreateTag(name, slug string) (int64, error) {
	res, err := s.db.Exec("INSERT INTO tags (name, slug) VALUES (?, ?)", name, slug)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (s *Store) FindOrCreateTag(name string) (int64, error) {
	slug := slugify(name)
	var id int64
	err := s.db.QueryRow("SELECT id FROM tags WHERE slug = ?", slug).Scan(&id)
	if err == nil {
		return id, nil
	}
	return s.CreateTag(name, slug)
}

func (s *Store) SearchTags(query string, limit int) ([]Tag, error) {
	rows, err := s.db.Query(
		"SELECT id, name, slug FROM tags WHERE name LIKE ? ORDER BY name LIMIT ?",
		"%"+query+"%", limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tags []Tag
	for rows.Next() {
		var t Tag
		if err := rows.Scan(&t.ID, &t.Name, &t.Slug); err != nil {
			return nil, err
		}
		tags = append(tags, t)
	}
	return tags, nil
}

func (s *Store) ListTags() ([]Tag, error) {
	rows, err := s.db.Query("SELECT id, name, slug FROM tags ORDER BY name")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tags []Tag
	for rows.Next() {
		var t Tag
		if err := rows.Scan(&t.ID, &t.Name, &t.Slug); err != nil {
			return nil, err
		}
		tags = append(tags, t)
	}
	return tags, nil
}

func (s *Store) GetTagBySlug(slug string) (*Tag, error) {
	t := &Tag{}
	err := s.db.QueryRow("SELECT id, name, slug FROM tags WHERE slug = ?", slug).
		Scan(&t.ID, &t.Name, &t.Slug)
	if err != nil {
		return nil, err
	}
	return t, nil
}

// --- Recipes ---

func (s *Store) CreateRecipe(r *Recipe) (int64, error) {
	res, err := s.db.Exec(`
		INSERT INTO recipes (title, description, source_url, image_path, prep_time, cook_time, servings, content_md, category_id, created_by)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		r.Title, r.Description, r.SourceURL, r.ImagePath, r.PrepTime, r.CookTime, r.Servings, r.ContentMD, r.CategoryID, r.CreatedBy,
	)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (s *Store) UpdateRecipe(r *Recipe) error {
	_, err := s.db.Exec(`
		UPDATE recipes SET title=?, description=?, source_url=?, image_path=?, prep_time=?, cook_time=?, servings=?, content_md=?, category_id=?, updated_at=CURRENT_TIMESTAMP
		WHERE id=?`,
		r.Title, r.Description, r.SourceURL, r.ImagePath, r.PrepTime, r.CookTime, r.Servings, r.ContentMD, r.CategoryID, r.ID,
	)
	return err
}

func (s *Store) DeleteRecipe(id int64) error {
	_, err := s.db.Exec("DELETE FROM recipes WHERE id = ?", id)
	return err
}

func (s *Store) GetRecipe(id int64) (*Recipe, error) {
	r := &Recipe{}
	err := s.db.QueryRow(`
		SELECT r.id, r.title, r.description, r.source_url, r.image_path, r.prep_time, r.cook_time, r.servings, r.content_md, r.category_id, r.created_by, r.created_at, r.updated_at,
		       COALESCE(u.display_name, u.username, '') as author_name
		FROM recipes r
		LEFT JOIN users u ON u.id = r.created_by
		WHERE r.id = ?`, id,
	).Scan(&r.ID, &r.Title, &r.Description, &r.SourceURL, &r.ImagePath, &r.PrepTime, &r.CookTime, &r.Servings, &r.ContentMD, &r.CategoryID, &r.CreatedBy, &r.CreatedAt, &r.UpdatedAt, &r.AuthorName)
	if err != nil {
		return nil, err
	}

	if r.CategoryID != nil {
		r.Category, _ = s.GetCategoryByID(*r.CategoryID)
	}
	r.Tags, _ = s.GetRecipeTags(r.ID)

	return r, nil
}

func (s *Store) ListRecipes(categoryID *int64, tagSlug string, limit, offset int) ([]Recipe, error) {
	query := `
		SELECT DISTINCT r.id, r.title, r.description, r.image_path, r.prep_time, r.cook_time, r.category_id, r.created_by, r.created_at,
		       COALESCE(u.display_name, u.username, '') as author_name
		FROM recipes r
		LEFT JOIN users u ON u.id = r.created_by`
	var args []any

	if tagSlug != "" {
		query += ` JOIN recipe_tags rt ON rt.recipe_id = r.id JOIN tags t ON t.id = rt.tag_id AND t.slug = ?`
		args = append(args, tagSlug)
	}

	var where []string
	if categoryID != nil {
		where = append(where, "r.category_id = ?")
		args = append(args, *categoryID)
	}
	if len(where) > 0 {
		query += " WHERE " + strings.Join(where, " AND ")
	}

	query += " ORDER BY r.created_at DESC LIMIT ? OFFSET ?"
	args = append(args, limit, offset)

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var recipes []Recipe
	for rows.Next() {
		var r Recipe
		if err := rows.Scan(&r.ID, &r.Title, &r.Description, &r.ImagePath, &r.PrepTime, &r.CookTime, &r.CategoryID, &r.CreatedBy, &r.CreatedAt, &r.AuthorName); err != nil {
			return nil, err
		}
		recipes = append(recipes, r)
	}
	rows.Close()

	for i := range recipes {
		recipes[i].Tags, _ = s.GetRecipeTags(recipes[i].ID)
	}
	return recipes, nil
}

func (s *Store) SearchRecipes(query string, limit int) ([]Recipe, error) {
	// Append * for prefix matching so partial words work (e.g. "Kuch" matches "Kuchen")
	ftsQuery := strings.TrimSpace(query) + "*"
	rows, err := s.db.Query(`
		SELECT r.id, r.title, r.description, r.image_path, r.prep_time, r.cook_time, r.category_id, r.created_by, r.created_at,
		       COALESCE(u.display_name, u.username, '') as author_name
		FROM recipes_fts fts
		JOIN recipes r ON r.id = fts.rowid
		LEFT JOIN users u ON u.id = r.created_by
		WHERE recipes_fts MATCH ?
		ORDER BY rank
		LIMIT ?`, ftsQuery, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var recipes []Recipe
	for rows.Next() {
		var r Recipe
		if err := rows.Scan(&r.ID, &r.Title, &r.Description, &r.ImagePath, &r.PrepTime, &r.CookTime, &r.CategoryID, &r.CreatedBy, &r.CreatedAt, &r.AuthorName); err != nil {
			return nil, err
		}
		recipes = append(recipes, r)
	}
	rows.Close()

	for i := range recipes {
		recipes[i].Tags, _ = s.GetRecipeTags(recipes[i].ID)
	}
	return recipes, nil
}

// --- Recipe Tags ---

func (s *Store) SetRecipeTags(recipeID int64, tagIDs []int64) error {
	s.db.Exec("DELETE FROM recipe_tags WHERE recipe_id = ?", recipeID)
	for _, tagID := range tagIDs {
		s.db.Exec("INSERT OR IGNORE INTO recipe_tags (recipe_id, tag_id) VALUES (?, ?)", recipeID, tagID)
	}
	return nil
}

func (s *Store) GetRecipeTags(recipeID int64) ([]Tag, error) {
	rows, err := s.db.Query(`
		SELECT t.id, t.name, t.slug FROM tags t
		JOIN recipe_tags rt ON rt.tag_id = t.id
		WHERE rt.recipe_id = ?
		ORDER BY t.name`, recipeID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tags []Tag
	for rows.Next() {
		var t Tag
		if err := rows.Scan(&t.ID, &t.Name, &t.Slug); err != nil {
			return nil, err
		}
		tags = append(tags, t)
	}
	return tags, nil
}

// --- User Recipes ---

func (s *Store) SetTried(userID, recipeID int64, tried bool) error {
	_, err := s.db.Exec(`
		INSERT INTO user_recipes (user_id, recipe_id, tried) VALUES (?, ?, ?)
		ON CONFLICT(user_id, recipe_id) DO UPDATE SET tried = ?`,
		userID, recipeID, tried, tried,
	)
	return err
}

func (s *Store) GetUserRecipe(userID, recipeID int64) (*UserRecipe, error) {
	ur := &UserRecipe{}
	err := s.db.QueryRow(
		"SELECT user_id, recipe_id, tried, notes, created_at FROM user_recipes WHERE user_id = ? AND recipe_id = ?",
		userID, recipeID,
	).Scan(&ur.UserID, &ur.RecipeID, &ur.Tried, &ur.Notes, &ur.CreatedAt)
	if err != nil {
		return nil, err
	}
	return ur, nil
}

func (s *Store) ListTriedRecipes(userID int64) ([]Recipe, error) {
	rows, err := s.db.Query(`
		SELECT r.id, r.title, r.description, r.image_path, r.prep_time, r.cook_time, r.category_id, r.created_by, r.created_at,
		       COALESCE(u.display_name, u.username, '') as author_name
		FROM user_recipes ur
		JOIN recipes r ON r.id = ur.recipe_id
		LEFT JOIN users u ON u.id = r.created_by
		WHERE ur.user_id = ? AND ur.tried = 1 AND r.created_by != ?
		ORDER BY ur.created_at DESC`, userID, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var recipes []Recipe
	for rows.Next() {
		var r Recipe
		if err := rows.Scan(&r.ID, &r.Title, &r.Description, &r.ImagePath, &r.PrepTime, &r.CookTime, &r.CategoryID, &r.CreatedBy, &r.CreatedAt, &r.AuthorName); err != nil {
			return nil, err
		}
		r.Tried = true
		recipes = append(recipes, r)
	}
	rows.Close()

	for i := range recipes {
		recipes[i].Tags, _ = s.GetRecipeTags(recipes[i].ID)
	}
	return recipes, nil
}

func (s *Store) IsTried(userID, recipeID int64) bool {
	var tried bool
	s.db.QueryRow("SELECT tried FROM user_recipes WHERE user_id = ? AND recipe_id = ?", userID, recipeID).Scan(&tried)
	return tried
}

// --- Shares ---

func (s *Store) ShareRecipe(recipeID, ownerID, sharedWithID int64) error {
	_, err := s.db.Exec(
		"INSERT OR IGNORE INTO shares (recipe_id, owner_id, shared_with_id) VALUES (?, ?, ?)",
		recipeID, ownerID, sharedWithID,
	)
	return err
}

func (s *Store) UnshareRecipe(recipeID, ownerID, sharedWithID int64) error {
	_, err := s.db.Exec(
		"DELETE FROM shares WHERE recipe_id = ? AND owner_id = ? AND shared_with_id = ?",
		recipeID, ownerID, sharedWithID,
	)
	return err
}

func (s *Store) RemoveSharedWithUser(recipeID, userID int64) error {
	_, err := s.db.Exec(
		"DELETE FROM shares WHERE recipe_id = ? AND shared_with_id = ?",
		recipeID, userID,
	)
	return err
}

func (s *Store) ListSharedWithUser(userID int64) ([]Recipe, error) {
	rows, err := s.db.Query(`
		SELECT r.id, r.title, r.description, r.image_path, r.prep_time, r.cook_time, r.category_id, r.created_by, r.created_at,
		       COALESCE(u.display_name, u.username, '') as author_name
		FROM shares sh
		JOIN recipes r ON r.id = sh.recipe_id
		LEFT JOIN users u ON u.id = r.created_by
		WHERE sh.shared_with_id = ?
		ORDER BY sh.created_at DESC`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var recipes []Recipe
	for rows.Next() {
		var r Recipe
		if err := rows.Scan(&r.ID, &r.Title, &r.Description, &r.ImagePath, &r.PrepTime, &r.CookTime, &r.CategoryID, &r.CreatedBy, &r.CreatedAt, &r.AuthorName); err != nil {
			return nil, err
		}
		r.SharedByName = r.AuthorName
		recipes = append(recipes, r)
	}
	rows.Close()

	for i := range recipes {
		recipes[i].Tags, _ = s.GetRecipeTags(recipes[i].ID)
	}
	return recipes, nil
}

func (s *Store) ListSharesForRecipe(recipeID int64) ([]User, error) {
	rows, err := s.db.Query(`
		SELECT u.id, u.username, u.display_name, u.created_at
		FROM shares sh
		JOIN users u ON u.id = sh.shared_with_id
		WHERE sh.recipe_id = ?
		ORDER BY u.display_name`, recipeID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []User
	for rows.Next() {
		var u User
		if err := rows.Scan(&u.ID, &u.Username, &u.DisplayName, &u.CreatedAt); err != nil {
			return nil, err
		}
		users = append(users, u)
	}
	return users, nil
}

// --- User's own recipes ---

func (s *Store) ListUserRecipes(userID int64) ([]Recipe, error) {
	rows, err := s.db.Query(`
		SELECT r.id, r.title, r.description, r.image_path, r.prep_time, r.cook_time, r.category_id, r.created_by, r.created_at,
		       COALESCE(u.display_name, u.username, '') as author_name
		FROM recipes r
		LEFT JOIN users u ON u.id = r.created_by
		WHERE r.created_by = ?
		ORDER BY r.created_at DESC`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var recipes []Recipe
	for rows.Next() {
		var r Recipe
		if err := rows.Scan(&r.ID, &r.Title, &r.Description, &r.ImagePath, &r.PrepTime, &r.CookTime, &r.CategoryID, &r.CreatedBy, &r.CreatedAt, &r.AuthorName); err != nil {
			return nil, err
		}
		recipes = append(recipes, r)
	}
	rows.Close()

	for i := range recipes {
		recipes[i].Tags, _ = s.GetRecipeTags(recipes[i].ID)
	}
	return recipes, nil
}

// --- Helpers ---

func slugify(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	replacer := strings.NewReplacer(
		"ä", "ae", "ö", "oe", "ü", "ue", "ß", "ss",
		" ", "-", "_", "-",
	)
	s = replacer.Replace(s)

	var result strings.Builder
	for _, r := range s {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' {
			result.WriteRune(r)
		}
	}

	// Remove double dashes
	final := result.String()
	for strings.Contains(final, "--") {
		final = strings.ReplaceAll(final, "--", "-")
	}
	return strings.Trim(final, "-")
}

func (s *Store) RecipeCount() int {
	var count int
	s.db.QueryRow("SELECT COUNT(*) FROM recipes").Scan(&count)
	return count
}

func (s *Store) CategoryRecipeCount(categoryID int64) int {
	var count int
	s.db.QueryRow("SELECT COUNT(*) FROM recipes WHERE category_id = ?", categoryID).Scan(&count)
	return count
}

func (s *Store) TagRecipeCount(tagID int64) int {
	var count int
	s.db.QueryRow("SELECT COUNT(*) FROM recipes r JOIN recipe_tags rt ON rt.recipe_id = r.id WHERE rt.tag_id = ?", tagID).Scan(&count)
	return count
}

func (s *Store) ListAllRecipesFull() ([]Recipe, error) {
	rows, err := s.db.Query(`
		SELECT r.id, r.title, r.description, r.source_url, r.image_path, r.prep_time, r.cook_time, r.servings, r.content_md, r.category_id, r.created_by, r.created_at, r.updated_at,
		       COALESCE(u.display_name, u.username, '') as author_name
		FROM recipes r
		LEFT JOIN users u ON u.id = r.created_by
		ORDER BY r.id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var recipes []Recipe
	for rows.Next() {
		var r Recipe
		if err := rows.Scan(&r.ID, &r.Title, &r.Description, &r.SourceURL, &r.ImagePath, &r.PrepTime, &r.CookTime, &r.Servings, &r.ContentMD, &r.CategoryID, &r.CreatedBy, &r.CreatedAt, &r.UpdatedAt, &r.AuthorName); err != nil {
			return nil, err
		}
		recipes = append(recipes, r)
	}
	rows.Close()

	for i := range recipes {
		if recipes[i].CategoryID != nil {
			recipes[i].Category, _ = s.GetCategoryByID(*recipes[i].CategoryID)
		}
		recipes[i].Tags, _ = s.GetRecipeTags(recipes[i].ID)
	}
	return recipes, nil
}

// --- Settings ---

func (s *Store) GetSetting(key string) string {
	var val string
	s.db.QueryRow("SELECT value FROM settings WHERE key = ?", key).Scan(&val)
	return val
}

func (s *Store) SetSetting(key, value string) {
	s.db.Exec("INSERT INTO settings (key, value) VALUES (?, ?) ON CONFLICT(key) DO UPDATE SET value = ?", key, value, value)
}

func (s *Store) RegistrationEnabled() bool {
	return s.GetSetting("registration_enabled") != "0"
}

// formatDuration converts a count to a display string for recipe counts
func formatCount(n int) string {
	if n == 1 {
		return "1 Rezept"
	}
	return fmt.Sprintf("%d Rezepte", n)
}
