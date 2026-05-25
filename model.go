package main

import "time"

type User struct {
	ID           int64
	Username     string
	DisplayName  string
	PasswordHash string
	IsAdmin      bool
	CreatedAt    time.Time
}

type Session struct {
	ID        int64
	UserID    int64
	Token     string
	ExpiresAt time.Time
}

type Category struct {
	ID   int64
	Name string
	Slug string
}

type Tag struct {
	ID   int64
	Name string
	Slug string
}

type Recipe struct {
	ID          int64
	Title       string
	Description string
	SourceURL   string
	ImagePath   string
	PrepTime    string
	CookTime    string
	Servings    string
	ContentMD   string
	CategoryID  *int64
	CreatedBy   int64
	CreatedAt   time.Time
	UpdatedAt   time.Time

	// Joined fields (not always populated)
	Category     *Category
	Tags         []Tag
	AuthorName   string
	Tried        bool
	SharedByName string
}

type UserRecipe struct {
	UserID    int64
	RecipeID  int64
	Tried     bool
	Notes     string
	CreatedAt time.Time
}

type Share struct {
	ID           int64
	RecipeID     int64
	OwnerID      int64
	SharedWithID int64
	CreatedAt    time.Time
}
