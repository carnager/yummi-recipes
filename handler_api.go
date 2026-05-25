package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"
)

func (app *App) apiJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func (app *App) apiError(w http.ResponseWriter, status int, msg string) {
	app.apiJSON(w, status, map[string]string{"error": msg})
}

// --- Auth ---

func (app *App) apiLogin(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		app.apiError(w, http.StatusBadRequest, "Ungueltiger Request")
		return
	}

	user, err := app.store.GetUserByUsername(req.Username)
	if err != nil || !checkPassword(user.PasswordHash, req.Password) {
		app.apiError(w, http.StatusUnauthorized, "Benutzername oder Passwort falsch")
		return
	}

	token := generateToken()
	app.store.CreateSession(user.ID, token, sessionExpiry())

	app.apiJSON(w, http.StatusOK, map[string]any{
		"token": token,
		"user": map[string]any{
			"id":           user.ID,
			"username":     user.Username,
			"display_name": user.DisplayName,
		},
	})
}

func (app *App) apiRegister(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Username    string `json:"username"`
		DisplayName string `json:"display_name"`
		Password    string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		app.apiError(w, http.StatusBadRequest, "Ungueltiger Request")
		return
	}

	if req.Username == "" || len(req.Password) < 6 {
		app.apiError(w, http.StatusBadRequest, "Benutzername erforderlich, Passwort mind. 6 Zeichen")
		return
	}

	hash, err := hashPassword(req.Password)
	if err != nil {
		app.apiError(w, http.StatusInternalServerError, "Fehler")
		return
	}

	userID, err := app.store.CreateUser(req.Username, req.DisplayName, hash)
	if err != nil {
		app.apiError(w, http.StatusConflict, "Benutzername bereits vergeben")
		return
	}

	token := generateToken()
	app.store.CreateSession(userID, token, sessionExpiry())

	app.apiJSON(w, http.StatusCreated, map[string]any{
		"token": token,
		"user": map[string]any{
			"id":           userID,
			"username":     req.Username,
			"display_name": req.DisplayName,
		},
	})
}

// --- Recipes ---

func (app *App) apiListRecipes(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("q")
	if q != "" {
		recipes, _ := app.store.SearchRecipes(q, 50)
		app.apiJSON(w, http.StatusOK, recipes)
		return
	}

	var catID *int64
	if c := r.URL.Query().Get("category"); c != "" {
		cat, err := app.store.GetCategoryBySlug(c)
		if err == nil {
			catID = &cat.ID
		}
	}

	tagSlug := r.URL.Query().Get("tag")
	recipes, _ := app.store.ListRecipes(catID, tagSlug, 100, 0)
	app.apiJSON(w, http.StatusOK, recipes)
}

func (app *App) apiGetRecipe(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		app.apiError(w, http.StatusBadRequest, "Ungueltige ID")
		return
	}

	recipe, err := app.store.GetRecipe(id)
	if err != nil {
		app.apiError(w, http.StatusNotFound, "Rezept nicht gefunden")
		return
	}

	app.apiJSON(w, http.StatusOK, recipe)
}

func (app *App) apiCreateRecipe(w http.ResponseWriter, r *http.Request) {
	user := ctxUser(r)
	var req struct {
		Title       string `json:"title"`
		Description string `json:"description"`
		SourceURL   string `json:"source_url"`
		PrepTime    string `json:"prep_time"`
		CookTime    string `json:"cook_time"`
		Servings    string `json:"servings"`
		ContentMD   string `json:"content_md"`
		CategoryID  *int64 `json:"category_id"`
		Tags        string `json:"tags"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		app.apiError(w, http.StatusBadRequest, "Ungueltiger Request")
		return
	}

	recipe := &Recipe{
		Title:       req.Title,
		Description: req.Description,
		SourceURL:   req.SourceURL,
		PrepTime:    req.PrepTime,
		CookTime:    req.CookTime,
		Servings:    req.Servings,
		ContentMD:   req.ContentMD,
		CategoryID:  req.CategoryID,
		CreatedBy:   user.ID,
	}

	id, err := app.store.CreateRecipe(recipe)
	if err != nil {
		app.apiError(w, http.StatusInternalServerError, "Fehler beim Erstellen")
		return
	}

	if req.Tags != "" {
		app.saveTags(id, req.Tags)
	}

	recipe.ID = id
	app.apiJSON(w, http.StatusCreated, recipe)
}

func (app *App) apiUpdateRecipe(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		app.apiError(w, http.StatusBadRequest, "Ungueltige ID")
		return
	}

	recipe, err := app.store.GetRecipe(id)
	if err != nil {
		app.apiError(w, http.StatusNotFound, "Rezept nicht gefunden")
		return
	}

	user := ctxUser(r)
	if user.ID != recipe.CreatedBy {
		app.apiError(w, http.StatusForbidden, "Nicht berechtigt")
		return
	}

	var req struct {
		Title       string `json:"title"`
		Description string `json:"description"`
		SourceURL   string `json:"source_url"`
		PrepTime    string `json:"prep_time"`
		CookTime    string `json:"cook_time"`
		Servings    string `json:"servings"`
		ContentMD   string `json:"content_md"`
		CategoryID  *int64 `json:"category_id"`
		Tags        string `json:"tags"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		app.apiError(w, http.StatusBadRequest, "Ungueltiger Request")
		return
	}

	recipe.Title = req.Title
	recipe.Description = req.Description
	recipe.SourceURL = req.SourceURL
	recipe.PrepTime = req.PrepTime
	recipe.CookTime = req.CookTime
	recipe.Servings = req.Servings
	recipe.ContentMD = req.ContentMD
	recipe.CategoryID = req.CategoryID

	if err := app.store.UpdateRecipe(recipe); err != nil {
		app.apiError(w, http.StatusInternalServerError, "Fehler beim Speichern")
		return
	}

	if req.Tags != "" {
		app.saveTags(recipe.ID, req.Tags)
	}

	app.apiJSON(w, http.StatusOK, recipe)
}

func (app *App) apiDeleteRecipe(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		app.apiError(w, http.StatusBadRequest, "Ungueltige ID")
		return
	}

	recipe, err := app.store.GetRecipe(id)
	if err != nil {
		app.apiError(w, http.StatusNotFound, "Rezept nicht gefunden")
		return
	}

	user := ctxUser(r)
	if user.ID != recipe.CreatedBy {
		app.apiError(w, http.StatusForbidden, "Nicht berechtigt")
		return
	}

	if recipe.ImagePath != "" {
		deleteImage(app.cfg.DataDir, recipe.ImagePath)
	}
	app.store.DeleteRecipe(id)
	w.WriteHeader(http.StatusNoContent)
}

func (app *App) apiImportRecipe(w http.ResponseWriter, r *http.Request) {
	user := ctxUser(r)
	var req struct {
		URL string `json:"url"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		app.apiError(w, http.StatusBadRequest, "Ungueltiger Request")
		return
	}

	recipe, err := fetchRecipeFromURL(req.URL)
	if err != nil {
		app.apiError(w, http.StatusBadRequest, err.Error())
		return
	}

	recipe.SourceURL = req.URL
	recipe.CreatedBy = user.ID

	// Handle image
	if recipe.ImagePath != "" {
		imgURL := recipe.ImagePath
		localPath, err := downloadImage(app.cfg.DataDir, imgURL)
		if err == nil {
			recipe.ImagePath = localPath
		} else {
			recipe.ImagePath = ""
		}
	}

	// Clear import-only fields
	recipe.AuthorName = ""
	recipe.CategoryID = nil
	recipe.Tags = nil

	// LLM: clean instructions
	if app.cfg.OpenAIKey != "" && app.store.GetSetting("llm_clean_instructions") != "0" {
		log.Printf("LLM API: Bereinige Zubereitungsschritte fuer '%s'...", recipe.Title)
		cleaned, err := cleanInstructions(app.cfg.OpenAIKey, recipe)
		if err != nil {
			log.Printf("LLM API: Fehler bei Schrittbereinigung: %v", err)
		} else {
			recipe = cleaned
			log.Printf("LLM API: Zubereitungsschritte bereinigt")
		}
	}

	// LLM: rephrase instructions
	if app.cfg.OpenAIKey != "" && app.store.GetSetting("llm_rephrase_instructions") != "0" {
		log.Printf("LLM API: Formuliere Zubereitungsschritte um fuer '%s'...", recipe.Title)
		rephrased, err := rephraseInstructions(app.cfg.OpenAIKey, recipe)
		if err != nil {
			log.Printf("LLM API: Fehler bei Umformulierung: %v", err)
		} else {
			recipe = rephrased
			log.Printf("LLM API: Zubereitungsschritte umformuliert")
		}
	}

	// LLM suggestion for category and tags
	if app.cfg.OpenAIKey != "" && app.store.GetSetting("llm_classify") != "0" {
		log.Printf("LLM API: Frage OpenAI nach Vorschlaegen fuer '%s'...", recipe.Title)
		cats, _ := app.store.ListCategories()
		allTags, _ := app.store.ListTags()
		suggestion, err := suggestCategoryAndTags(app.cfg.OpenAIKey, recipe, cats, allTags)
		if err != nil {
			log.Printf("LLM API: Fehler: %v", err)
		} else if suggestion != nil {
			log.Printf("LLM API: Vorschlag Titel=%q, Kategorie=%q, Tags=%v", suggestion.Title, suggestion.CategoryName, suggestion.TagNames)
			if suggestion.Title != "" && suggestion.Title != recipe.Title {
				log.Printf("LLM API: Titel bereinigt '%s' -> '%s'", recipe.Title, suggestion.Title)
				recipe.Title = suggestion.Title
			}
			for _, c := range cats {
				if strings.EqualFold(c.Name, suggestion.CategoryName) {
					catID := c.ID
					recipe.CategoryID = &catID
					break
				}
			}
			for _, sugName := range suggestion.TagNames {
				for _, t := range allTags {
					if strings.EqualFold(t.Name, sugName) {
						recipe.Tags = append(recipe.Tags, t)
						break
					}
				}
			}
			if suggestion.PrepTime != "" {
				recipe.PrepTime = suggestion.PrepTime
			}
			if suggestion.CookTime != "" {
				recipe.CookTime = suggestion.CookTime
			}
			log.Printf("LLM API: %d Tags zugeordnet", len(recipe.Tags))
		}
	}

	// Save tags
	importTags := recipe.Tags
	recipe.Tags = nil

	id, err := app.store.CreateRecipe(recipe)
	if err != nil {
		app.apiError(w, http.StatusInternalServerError, "Fehler beim Erstellen")
		return
	}

	if len(importTags) > 0 {
		names := make([]string, len(importTags))
		for i, t := range importTags {
			names[i] = t.Name
		}
		app.saveTags(id, strings.Join(names, ","))
	}

	recipe.ID = id
	app.apiJSON(w, http.StatusCreated, recipe)
}

func (app *App) apiToggleTried(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(r.PathValue("id"), 10, 64)
	user := ctxUser(r)
	tried := app.store.IsTried(user.ID, id)
	app.store.SetTried(user.ID, id, !tried)
	app.apiJSON(w, http.StatusOK, map[string]bool{"tried": !tried})
}

func (app *App) apiShareRecipe(w http.ResponseWriter, r *http.Request) {
	recipeID, _ := strconv.ParseInt(r.PathValue("id"), 10, 64)
	user := ctxUser(r)

	var req struct {
		UserID int64  `json:"user_id"`
		Action string `json:"action"`
	}
	json.NewDecoder(r.Body).Decode(&req)

	if req.Action == "remove" {
		app.store.UnshareRecipe(recipeID, user.ID, req.UserID)
	} else {
		app.store.ShareRecipe(recipeID, user.ID, req.UserID)
	}

	app.apiJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (app *App) apiListCategories(w http.ResponseWriter, r *http.Request) {
	cats, _ := app.store.ListCategories()
	app.apiJSON(w, http.StatusOK, cats)
}

func (app *App) apiListTags(w http.ResponseWriter, r *http.Request) {
	tags, _ := app.store.ListTags()
	app.apiJSON(w, http.StatusOK, tags)
}

func (app *App) apiMyRecipes(w http.ResponseWriter, r *http.Request) {
	user := ctxUser(r)
	own, _ := app.store.ListUserRecipes(user.ID)
	shared, _ := app.store.ListSharedWithUser(user.ID)
	app.apiJSON(w, http.StatusOK, map[string]any{
		"own":    own,
		"shared": shared,
	})
}

func (app *App) apiListSharesForRecipe(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		app.apiError(w, http.StatusBadRequest, "Ungueltige ID")
		return
	}
	users, _ := app.store.ListSharesForRecipe(id)
	type apiUser struct {
		ID          int64  `json:"id"`
		Username    string `json:"username"`
		DisplayName string `json:"display_name"`
	}
	var result []apiUser
	for _, u := range users {
		result = append(result, apiUser{ID: u.ID, Username: u.Username, DisplayName: u.DisplayName})
	}
	app.apiJSON(w, http.StatusOK, result)
}

func (app *App) apiListUsers(w http.ResponseWriter, r *http.Request) {
	users, _ := app.store.ListUsers()
	type apiUser struct {
		ID          int64  `json:"id"`
		Username    string `json:"username"`
		DisplayName string `json:"display_name"`
	}
	var result []apiUser
	for _, u := range users {
		result = append(result, apiUser{ID: u.ID, Username: u.Username, DisplayName: u.DisplayName})
	}
	app.apiJSON(w, http.StatusOK, result)
}
