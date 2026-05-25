package main

import (
	"archive/zip"
	"bytes"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type PageData struct {
	Title         string
	User          *User
	Flash         string
	Error         string
	Recipe        Recipe
	Recipes       []Recipe
	OwnRecipes    []Recipe
	SharedRecipes []Recipe
	TriedRecipes  []Recipe
	Categories    []Category
	Tags          []Tag
	Users         []User
	SharedWith    []User
	PageTitle     string
	IsOwner       bool
	Tried         bool
	RecipeText    string
	EditUser      User
	Lang                    Lang
	RegEnabled              bool
	LLMClassify             bool
	LLMCleanInstructions    bool
	LLMRephraseInstructions bool
	HasOpenAIKey            bool
}

func (app *App) render(w http.ResponseWriter, r *http.Request, tmpl string, data *PageData) {
	if data == nil {
		data = &PageData{}
	}
	data.User = ctxUser(r)
	data.Lang = app.detectLang(r)
	data.RegEnabled = app.store.RegistrationEnabled()
	data.LLMClassify = app.store.GetSetting("llm_classify") != "0"
	data.LLMCleanInstructions = app.store.GetSetting("llm_clean_instructions") != "0"
	data.LLMRephraseInstructions = app.store.GetSetting("llm_rephrase_instructions") != "0"
	data.HasOpenAIKey = app.cfg.OpenAIKey != ""

	if cookie, err := r.Cookie("flash"); err == nil {
		data.Flash = cookie.Value
		http.SetCookie(w, &http.Cookie{Name: "flash", MaxAge: -1, Path: "/"})
	}

	t, ok := app.templates[tmpl]
	if !ok {
		log.Printf("Template nicht gefunden: %s", tmpl)
		http.Error(w, "Interner Fehler", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := t.Execute(w, data); err != nil {
		log.Printf("Template-Fehler %s: %v", tmpl, err)
	}
}

func (app *App) setFlash(w http.ResponseWriter, msg string) {
	http.SetCookie(w, &http.Cookie{Name: "flash", Value: msg, Path: "/", MaxAge: 5})
}

func (app *App) flash(w http.ResponseWriter, r *http.Request, key string) {
	app.setFlash(w, T(app.detectLang(r), key))
}

// --- Auth handlers ---

func (app *App) loginPage(w http.ResponseWriter, r *http.Request) {
	if ctxUser(r) != nil {
		http.Redirect(w, r, "/rezepte", http.StatusSeeOther)
		return
	}
	app.render(w, r, "login.html", &PageData{Title: "Anmelden"})
}

func (app *App) loginSubmit(w http.ResponseWriter, r *http.Request) {
	username := strings.TrimSpace(r.FormValue("username"))
	password := r.FormValue("password")

	user, err := app.store.GetUserByUsername(username)
	if err != nil || !checkPassword(user.PasswordHash, password) {
		lang := app.detectLang(r)
		app.render(w, r, "login.html", &PageData{Title: T(lang, "auth.login_title"), Error: T(lang, "auth.err_invalid")})
		return
	}

	token := generateToken()
	app.store.CreateSession(user.ID, token, sessionExpiry())
	http.SetCookie(w, &http.Cookie{
		Name:     "session",
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   30 * 24 * 3600,
	})
	http.Redirect(w, r, "/rezepte", http.StatusSeeOther)
}

func (app *App) registerPage(w http.ResponseWriter, r *http.Request) {
	if ctxUser(r) != nil {
		http.Redirect(w, r, "/rezepte", http.StatusSeeOther)
		return
	}
	if !app.store.RegistrationEnabled() {
		app.render(w, r, "login.html", &PageData{Title: "Anmelden", Error: "Registrierung ist deaktiviert."})
		return
	}
	app.render(w, r, "register.html", &PageData{Title: "Registrieren"})
}

func (app *App) registerSubmit(w http.ResponseWriter, r *http.Request) {
	if !app.store.RegistrationEnabled() {
		http.Error(w, "Registrierung ist deaktiviert", http.StatusForbidden)
		return
	}
	username := strings.TrimSpace(r.FormValue("username"))
	displayName := strings.TrimSpace(r.FormValue("display_name"))
	password := r.FormValue("password")
	password2 := r.FormValue("password2")

	lang := app.detectLang(r)
	if username == "" || password == "" {
		app.render(w, r, "register.html", &PageData{Title: T(lang, "auth.register_title"), Error: T(lang, "auth.err_fields")})
		return
	}
	if len(password) < 6 {
		app.render(w, r, "register.html", &PageData{Title: T(lang, "auth.register_title"), Error: T(lang, "auth.err_password_len")})
		return
	}
	if password != password2 {
		app.render(w, r, "register.html", &PageData{Title: T(lang, "auth.register_title"), Error: T(lang, "auth.err_password_match")})
		return
	}

	hash, err := hashPassword(password)
	if err != nil {
		app.render(w, r, "register.html", &PageData{Title: T(lang, "auth.register_title"), Error: T(lang, "auth.err_create")})
		return
	}

	userID, err := app.store.CreateUser(username, displayName, hash)
	if err != nil {
		app.render(w, r, "register.html", &PageData{Title: T(lang, "auth.register_title"), Error: T(lang, "auth.err_exists")})
		return
	}

	token := generateToken()
	app.store.CreateSession(userID, token, sessionExpiry())
	http.SetCookie(w, &http.Cookie{
		Name:     "session",
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   30 * 24 * 3600,
	})
	http.Redirect(w, r, "/rezepte", http.StatusSeeOther)
}

func (app *App) logoutSubmit(w http.ResponseWriter, r *http.Request) {
	if c, err := r.Cookie("session"); err == nil {
		app.store.DeleteSession(c.Value)
	}
	http.SetCookie(w, &http.Cookie{Name: "session", MaxAge: -1, Path: "/"})
	http.Redirect(w, r, "/rezepte", http.StatusSeeOther)
}

// --- Recipe handlers ---

func (app *App) recipeList(w http.ResponseWriter, r *http.Request) {
	recipes, _ := app.store.ListRecipes(nil, "", 100, 0)
	app.render(w, r, "recipe_list.html", &PageData{Title: "Rezepte", Recipes: recipes})
}

func (app *App) recipeDetail(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	recipe, err := app.store.GetRecipe(id)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	user := ctxUser(r)
	isOwner := user != nil && user.ID == recipe.CreatedBy
	tried := false
	if user != nil {
		tried = app.store.IsTried(user.ID, recipe.ID)
	}

	var users []User
	var sharedWith []User
	if isOwner {
		allUsers, _ := app.store.ListUsers()
		for _, u := range allUsers {
			if u.ID != user.ID {
				users = append(users, u)
			}
		}
		sharedWith, _ = app.store.ListSharesForRecipe(recipe.ID)
	}

	app.render(w, r, "recipe_detail.html", &PageData{
		Title:      recipe.Title,
		Recipe:     *recipe,
		IsOwner:    isOwner,
		Tried:      tried,
		Users:      users,
		SharedWith: sharedWith,
	})
}

func (app *App) recipeNew(w http.ResponseWriter, r *http.Request) {
	app.render(w, r, "recipe_new.html", &PageData{Title: "Neues Rezept"})
}

func (app *App) recipeNewManual(w http.ResponseWriter, r *http.Request) {
	cats, _ := app.store.ListCategories()
	app.render(w, r, "recipe_form.html", &PageData{Title: "Neues Rezept", Categories: cats})
}

func (app *App) recipeLLMPage(w http.ResponseWriter, r *http.Request) {
	app.render(w, r, "recipe_llm.html", &PageData{Title: "KI-Erkennung"})
}

func (app *App) recipeLLMSubmit(w http.ResponseWriter, r *http.Request) {
	if app.cfg.OpenAIKey == "" {
		app.render(w, r, "recipe_llm.html", &PageData{
			Title: "KI-Erkennung",
			Error: "OpenAI API-Key nicht konfiguriert (YUMMI_OPENAI_KEY).",
		})
		return
	}

	r.ParseMultipartForm(10 << 20) // 10 MB
	recipeText := strings.TrimSpace(r.FormValue("recipe_text"))

	// Check for uploaded source image (photo of recipe text)
	var imageBase64 string
	file, header, err := r.FormFile("source_image")
	if err == nil {
		defer file.Close()
		imgData, err := io.ReadAll(io.LimitReader(file, 5<<20))
		if err == nil && len(imgData) > 0 {
			mime := "image/jpeg"
			if strings.HasSuffix(strings.ToLower(header.Filename), ".png") {
				mime = "image/png"
			} else if strings.HasSuffix(strings.ToLower(header.Filename), ".webp") {
				mime = "image/webp"
			}
			imageBase64 = fmt.Sprintf("data:%s;base64,%s", mime, base64.StdEncoding.EncodeToString(imgData))
			log.Printf("LLM: Quellbild hochgeladen (%s, %d bytes)", header.Filename, len(imgData))
		}
	}

	if recipeText == "" && imageBase64 == "" {
		app.render(w, r, "recipe_llm.html", &PageData{
			Title: "KI-Erkennung",
			Error: "Bitte Text eingeben oder ein Foto hochladen.",
		})
		return
	}

	cats, _ := app.store.ListCategories()
	allTags, _ := app.store.ListTags()

	log.Printf("LLM: Erstelle Rezept aus %s...", func() string {
		if imageBase64 != "" {
			return "Bild"
		}
		return "Text"
	}())

	recipe, err := extractRecipeViaLLM(app.cfg.OpenAIKey, recipeText, imageBase64, cats, allTags)
	if err != nil {
		log.Printf("LLM: Fehler: %v", err)
		app.render(w, r, "recipe_llm.html", &PageData{
			Title:      "KI-Erkennung",
			Error:      fmt.Sprintf("KI-Erkennung fehlgeschlagen: %v", err),
			RecipeText: recipeText,
		})
		return
	}

	log.Printf("LLM: Rezept erkannt: '%s', Kategorie=%v, %d Tags", recipe.Title, recipe.CategoryID, len(recipe.Tags))

	app.render(w, r, "recipe_form.html", &PageData{
		Title:      "Erkanntes Rezept",
		Recipe:     *recipe,
		Categories: cats,
	})
}

func (app *App) recipeCreate(w http.ResponseWriter, r *http.Request) {
	user := ctxUser(r)
	recipe := Recipe{
		Title:       strings.TrimSpace(r.FormValue("title")),
		Description: strings.TrimSpace(r.FormValue("description")),
		SourceURL:   strings.TrimSpace(r.FormValue("source_url")),
		ImagePath:   strings.TrimSpace(r.FormValue("image_path")),
		PrepTime:    strings.TrimSpace(r.FormValue("prep_time")),
		CookTime:    strings.TrimSpace(r.FormValue("cook_time")),
		Servings:    strings.TrimSpace(r.FormValue("servings")),
		ContentMD:   r.FormValue("content_md"),
		CreatedBy:   user.ID,
	}

	if catID := r.FormValue("category_id"); catID != "" {
		id, _ := strconv.ParseInt(catID, 10, 64)
		if id > 0 {
			recipe.CategoryID = &id
		}
	}

	// Handle photo upload
	if photo, header, err := r.FormFile("recipe_photo"); err == nil {
		defer photo.Close()
		imgData, err := io.ReadAll(io.LimitReader(photo, 5<<20))
		if err == nil && len(imgData) > 0 {
			if path, err := saveImageBytes(app.cfg.DataDir, imgData, header.Filename); err == nil {
				recipe.ImagePath = path
			}
		}
	}

	if recipe.Title == "" || recipe.ContentMD == "" {
		cats, _ := app.store.ListCategories()
		app.render(w, r, "recipe_form.html", &PageData{
			Title:      "Neues Rezept",
			Recipe:     recipe,
			Categories: cats,
			Error:      "Titel und Rezeptinhalt sind erforderlich.",
		})
		return
	}

	recipeID, err := app.store.CreateRecipe(&recipe)
	if err != nil {
		log.Printf("Rezept erstellen: %v", err)
		cats, _ := app.store.ListCategories()
		app.render(w, r, "recipe_form.html", &PageData{
			Title:      "Neues Rezept",
			Recipe:     recipe,
			Categories: cats,
			Error:      "Fehler beim Erstellen des Rezepts.",
		})
		return
	}

	app.saveTags(recipeID, r.FormValue("tags"))
	app.flash(w, r, "flash.recipe_created")
	http.Redirect(w, r, fmt.Sprintf("/rezepte/%d", recipeID), http.StatusSeeOther)
}

func (app *App) recipeEdit(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	recipe, err := app.store.GetRecipe(id)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	user := ctxUser(r)
	if user.ID != recipe.CreatedBy {
		http.Error(w, "Nicht berechtigt", http.StatusForbidden)
		return
	}

	cats, _ := app.store.ListCategories()
	app.render(w, r, "recipe_form.html", &PageData{
		Title:      "Rezept bearbeiten",
		Recipe:     *recipe,
		Categories: cats,
	})
}

func (app *App) recipeUpdate(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	recipe, err := app.store.GetRecipe(id)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	user := ctxUser(r)
	if user.ID != recipe.CreatedBy {
		http.Error(w, "Nicht berechtigt", http.StatusForbidden)
		return
	}

	recipe.Title = strings.TrimSpace(r.FormValue("title"))
	recipe.Description = strings.TrimSpace(r.FormValue("description"))
	recipe.SourceURL = strings.TrimSpace(r.FormValue("source_url"))
	recipe.PrepTime = strings.TrimSpace(r.FormValue("prep_time"))
	recipe.CookTime = strings.TrimSpace(r.FormValue("cook_time"))
	recipe.Servings = strings.TrimSpace(r.FormValue("servings"))
	recipe.ContentMD = r.FormValue("content_md")

	// Handle photo upload
	if photo, header, err := r.FormFile("recipe_photo"); err == nil {
		defer photo.Close()
		imgData, err := io.ReadAll(io.LimitReader(photo, 5<<20))
		if err == nil && len(imgData) > 0 {
			if path, err := saveImageBytes(app.cfg.DataDir, imgData, header.Filename); err == nil {
				// Delete old image
				if recipe.ImagePath != "" {
					deleteImage(app.cfg.DataDir, recipe.ImagePath)
				}
				recipe.ImagePath = path
			}
		}
	}

	if catID := r.FormValue("category_id"); catID != "" {
		cid, _ := strconv.ParseInt(catID, 10, 64)
		if cid > 0 {
			recipe.CategoryID = &cid
		} else {
			recipe.CategoryID = nil
		}
	} else {
		recipe.CategoryID = nil
	}

	if err := app.store.UpdateRecipe(recipe); err != nil {
		cats, _ := app.store.ListCategories()
		app.render(w, r, "recipe_form.html", &PageData{
			Title:      "Rezept bearbeiten",
			Recipe:     *recipe,
			Categories: cats,
			Error:      "Fehler beim Speichern.",
		})
		return
	}

	app.saveTags(recipe.ID, r.FormValue("tags"))
	app.setFlash(w, "Rezept gespeichert!")
	http.Redirect(w, r, fmt.Sprintf("/rezepte/%d", recipe.ID), http.StatusSeeOther)
}

func (app *App) recipeDelete(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	recipe, err := app.store.GetRecipe(id)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	user := ctxUser(r)
	if user.ID != recipe.CreatedBy {
		http.Error(w, "Nicht berechtigt", http.StatusForbidden)
		return
	}

	if recipe.ImagePath != "" {
		deleteImage(app.cfg.DataDir, recipe.ImagePath)
	}
	app.store.DeleteRecipe(id)

	w.Header().Set("HX-Redirect", "/rezepte")
	w.WriteHeader(http.StatusOK)
}

// --- Category & Tag handlers ---

func (app *App) categoryList(w http.ResponseWriter, r *http.Request) {
	cats, _ := app.store.ListCategories()
	app.render(w, r, "category_list.html", &PageData{Title: "Kategorien", Categories: cats})
}

func (app *App) categoryRecipes(w http.ResponseWriter, r *http.Request) {
	slug := r.PathValue("slug")
	cat, err := app.store.GetCategoryBySlug(slug)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	recipes, _ := app.store.ListRecipes(&cat.ID, "", 100, 0)
	app.render(w, r, "recipe_list.html", &PageData{
		Title:     cat.Name,
		Recipes:   recipes,
		PageTitle: cat.Name,
	})
}

func (app *App) tagRecipes(w http.ResponseWriter, r *http.Request) {
	slug := r.PathValue("slug")
	tag, err := app.store.GetTagBySlug(slug)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	recipes, _ := app.store.ListRecipes(nil, slug, 100, 0)
	app.render(w, r, "recipe_list.html", &PageData{
		Title:     tag.Name,
		Recipes:   recipes,
		PageTitle: tag.Name,
	})
}

// --- Search ---

func (app *App) searchRecipes(w http.ResponseWriter, r *http.Request) {
	q := strings.TrimSpace(r.URL.Query().Get("q"))
	if q == "" {
		recipes, _ := app.store.ListRecipes(nil, "", 100, 0)
		app.renderPartial(w, "recipe_cards", recipes)
		return
	}

	recipes, _ := app.store.SearchRecipes(q, 50)
	app.renderPartial(w, "recipe_cards", recipes)
}

func (app *App) renderPartial(w http.ResponseWriter, name string, recipes []Recipe) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if len(recipes) == 0 {
		fmt.Fprint(w, `<p class="empty-state">Keine Rezepte gefunden.</p>`)
		return
	}
	for _, recipe := range recipes {
		t := app.templates["recipe_list.html"]
		t.ExecuteTemplate(w, "recipe_card", recipe)
	}
}

// --- Tried ---

func (app *App) toggleTried(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(r.PathValue("id"), 10, 64)
	user := ctxUser(r)
	tried := app.store.IsTried(user.ID, id)
	app.store.SetTried(user.ID, id, !tried)

	recipe, _ := app.store.GetRecipe(id)
	if recipe == nil {
		http.NotFound(w, r)
		return
	}

	data := &PageData{Recipe: *recipe, Tried: !tried}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	app.templates["recipe_detail.html"].ExecuteTemplate(w, "tried_button", data)
}

// --- Sharing ---

func (app *App) shareRecipe(w http.ResponseWriter, r *http.Request) {
	recipeID, _ := strconv.ParseInt(r.PathValue("id"), 10, 64)
	userID, _ := strconv.ParseInt(r.URL.Query().Get("user_id"), 10, 64)
	action := r.URL.Query().Get("action")
	owner := ctxUser(r)

	if action == "remove" {
		app.store.UnshareRecipe(recipeID, owner.ID, userID)
	} else {
		app.store.ShareRecipe(recipeID, owner.ID, userID)
	}

	recipe, _ := app.store.GetRecipe(recipeID)
	allUsers, _ := app.store.ListUsers()
	var users []User
	for _, u := range allUsers {
		if u.ID != owner.ID {
			users = append(users, u)
		}
	}
	sharedWith, _ := app.store.ListSharesForRecipe(recipeID)

	data := &PageData{
		Recipe:     *recipe,
		Users:      users,
		SharedWith: sharedWith,
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	app.templates["recipe_detail.html"].ExecuteTemplate(w, "share_list", data)
}

// --- My Recipes ---

func (app *App) myRecipes(w http.ResponseWriter, r *http.Request) {
	user := ctxUser(r)
	own, _ := app.store.ListUserRecipes(user.ID)
	shared, _ := app.store.ListSharedWithUser(user.ID)
	tried, _ := app.store.ListTriedRecipes(user.ID)
	app.render(w, r, "user_recipes.html", &PageData{
		Title:         "Meine Rezepte",
		OwnRecipes:    own,
		SharedRecipes: shared,
		TriedRecipes:  tried,
	})
}

// --- Tag suggestions ---

func (app *App) suggestTags(w http.ResponseWriter, r *http.Request) {
	q := strings.TrimSpace(r.URL.Query().Get("q"))
	if q == "" {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte("[]"))
		return
	}
	tags, _ := app.store.SearchTags(q, 10)
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte("["))
	for i, t := range tags {
		if i > 0 {
			w.Write([]byte(","))
		}
		fmt.Fprintf(w, `"%s"`, t.Name)
	}
	w.Write([]byte("]"))
}

// --- Remove from my recipes ---

func (app *App) removeFromMyRecipes(w http.ResponseWriter, r *http.Request) {
	user := ctxUser(r)
	recipeID, _ := strconv.ParseInt(r.URL.Query().Get("recipe_id"), 10, 64)
	removeType := r.URL.Query().Get("type")

	if removeType == "shared" {
		// Find who shared it and remove
		app.store.RemoveSharedWithUser(recipeID, user.ID)
	}

	w.WriteHeader(http.StatusOK)
}

// --- Helpers ---

func (app *App) saveTags(recipeID int64, tagsStr string) {
	parts := strings.Split(tagsStr, ",")
	var tagIDs []int64
	for _, p := range parts {
		name := strings.TrimSpace(p)
		if name == "" {
			continue
		}
		tagID, err := app.store.FindOrCreateTag(name)
		if err == nil {
			tagIDs = append(tagIDs, tagID)
		}
	}
	app.store.SetRecipeTags(recipeID, tagIDs)
}

// Template functions

func templateFuncs() template.FuncMap {
	return template.FuncMap{
		"t": func(lang Lang, key string) string {
			return T(lang, key)
		},
		"markdown": func(s string) template.HTML {
			return template.HTML(renderMarkdown(s))
		},
		"joinTags": func(tags []Tag) string {
			names := make([]string, len(tags))
			for i, t := range tags {
				names[i] = t.Name
			}
			return strings.Join(names, ", ")
		},
		"categorySelected": func(recipeCategory *int64, categoryID int64) bool {
			return recipeCategory != nil && *recipeCategory == categoryID
		},
		"isSharedWith": func(recipeID int64, userID int64, sharedWith []User) bool {
			for _, u := range sharedWith {
				if u.ID == userID {
					return true
				}
			}
			return false
		},
		"safeHTML": func(s string) template.HTML {
			return template.HTML(s)
		},
		"categoryIcon": func(name string) string {
			icons := map[string]string{
				"Frühstück":     "🍳",
				"Fruehstueck":   "🍳",
				"Vorspeisen":    "🥗",
				"Hauptgerichte": "🍽️",
				"Suppen":        "🍲",
				"Salate":        "🥬",
				"Beilagen":      "🥔",
				"Desserts":      "🍰",
				"Backen":        "🍞",
				"Snacks":        "🍿",
				"Getränke":      "🥤",
				"Getraenke":     "🥤",
				"Kochen":        "👨‍🍳",
			}
			if icon, ok := icons[name]; ok {
				return icon
			}
			return "🍴"
		},
	}
}

// newCategoryFromForm creates or finds a category from user input
func (app *App) findOrCreateCategoryFromForm(name string) *int64 {
	if name == "" {
		return nil
	}
	id, err := app.store.FindOrCreateCategory(name)
	if err != nil {
		return nil
	}
	return &id
}

// Ensure sql is imported for ErrNoRows usage
var _ = sql.ErrNoRows

// ==================== ADMIN ====================

func (app *App) adminUserList(w http.ResponseWriter, r *http.Request) {
	users, _ := app.store.ListUsers()
	app.render(w, r, "admin_users.html", &PageData{Title: "Benutzerverwaltung", Users: users})
}

func (app *App) adminUserNewPage(w http.ResponseWriter, r *http.Request) {
	app.render(w, r, "admin_user_form.html", &PageData{Title: "Neuer Benutzer"})
}

func (app *App) adminUserNewSubmit(w http.ResponseWriter, r *http.Request) {
	username := strings.TrimSpace(r.FormValue("username"))
	displayName := strings.TrimSpace(r.FormValue("display_name"))
	password := r.FormValue("password")
	isAdmin := r.FormValue("is_admin") == "1"

	if username == "" || password == "" {
		app.render(w, r, "admin_user_form.html", &PageData{
			Title:    "Neuer Benutzer",
			Error:    "Benutzername und Passwort sind erforderlich.",
			EditUser: User{Username: username, DisplayName: displayName, IsAdmin: isAdmin},
		})
		return
	}

	hash, err := hashPassword(password)
	if err != nil {
		app.render(w, r, "admin_user_form.html", &PageData{
			Title: "Neuer Benutzer",
			Error: "Fehler beim Erstellen.",
		})
		return
	}

	userID, err := app.store.CreateUser(username, displayName, hash)
	if err != nil {
		app.render(w, r, "admin_user_form.html", &PageData{
			Title:    "Neuer Benutzer",
			Error:    "Benutzername bereits vergeben.",
			EditUser: User{Username: username, DisplayName: displayName, IsAdmin: isAdmin},
		})
		return
	}

	if isAdmin {
		app.store.db.Exec("UPDATE users SET is_admin = 1 WHERE id = ?", userID)
	}

	app.setFlash(w, "Benutzer angelegt.")
	http.Redirect(w, r, "/admin/benutzer", http.StatusSeeOther)
}

func (app *App) adminUserEditPage(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	u, err := app.store.GetUserByID(id)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	app.render(w, r, "admin_user_form.html", &PageData{Title: "Benutzer bearbeiten", EditUser: *u})
}

func (app *App) adminUserEditSubmit(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	u, err := app.store.GetUserByID(id)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	u.Username = strings.TrimSpace(r.FormValue("username"))
	u.DisplayName = strings.TrimSpace(r.FormValue("display_name"))
	u.IsAdmin = r.FormValue("is_admin") == "1"

	if u.Username == "" {
		app.render(w, r, "admin_user_form.html", &PageData{
			Title:    "Benutzer bearbeiten",
			Error:    "Benutzername darf nicht leer sein.",
			EditUser: *u,
		})
		return
	}

	if err := app.store.UpdateUser(u); err != nil {
		app.render(w, r, "admin_user_form.html", &PageData{
			Title:    "Benutzer bearbeiten",
			Error:    "Fehler beim Speichern.",
			EditUser: *u,
		})
		return
	}

	// Update password if provided
	if pw := r.FormValue("password"); pw != "" {
		hash, err := hashPassword(pw)
		if err == nil {
			app.store.UpdateUserPassword(u.ID, hash)
		}
	}

	app.setFlash(w, "Benutzer aktualisiert.")
	http.Redirect(w, r, "/admin/benutzer", http.StatusSeeOther)
}

func (app *App) adminUserDelete(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	// Don't allow deleting yourself
	currentUser := ctxUser(r)
	if currentUser.ID == id {
		app.setFlash(w, "Du kannst dich nicht selbst loeschen.")
		http.Redirect(w, r, "/admin/benutzer", http.StatusSeeOther)
		return
	}

	app.store.DeleteUser(id)
	app.setFlash(w, "Benutzer geloescht.")
	http.Redirect(w, r, "/admin/benutzer", http.StatusSeeOther)
}

func (app *App) adminSettings(w http.ResponseWriter, r *http.Request) {
	regEnabled := "0"
	if r.FormValue("registration_enabled") == "1" {
		regEnabled = "1"
	}
	app.store.SetSetting("registration_enabled", regEnabled)

	llmClassify := "0"
	if r.FormValue("llm_classify") == "1" {
		llmClassify = "1"
	}
	app.store.SetSetting("llm_classify", llmClassify)

	llmCleanInstructions := "0"
	if r.FormValue("llm_clean_instructions") == "1" {
		llmCleanInstructions = "1"
	}
	app.store.SetSetting("llm_clean_instructions", llmCleanInstructions)

	llmRephraseInstructions := "0"
	if r.FormValue("llm_rephrase_instructions") == "1" {
		llmRephraseInstructions = "1"
	}
	app.store.SetSetting("llm_rephrase_instructions", llmRephraseInstructions)

	http.Redirect(w, r, "/admin/benutzer", http.StatusSeeOther)
}

// ==================== EXPORT / IMPORT ====================

type ExportRecipe struct {
	Title       string   `json:"title"`
	Description string   `json:"description,omitempty"`
	SourceURL   string   `json:"source_url,omitempty"`
	ImageFile   string   `json:"image_file,omitempty"`
	PrepTime    string   `json:"prep_time,omitempty"`
	CookTime    string   `json:"cook_time,omitempty"`
	Servings    string   `json:"servings,omitempty"`
	ContentMD   string   `json:"content_md"`
	Category    string   `json:"category,omitempty"`
	Tags        []string `json:"tags,omitempty"`
	Author      string   `json:"author,omitempty"`
	CreatedAt   string   `json:"created_at"`
}

func (app *App) adminExport(w http.ResponseWriter, r *http.Request) {
	recipes, err := app.store.ListAllRecipesFull()
	if err != nil {
		http.Error(w, "Export fehlgeschlagen", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/zip")
	w.Header().Set("Content-Disposition", `attachment; filename="yummi-export.zip"`)

	zw := zip.NewWriter(w)
	defer zw.Close()

	// Build export data
	var exported []ExportRecipe
	for _, rec := range recipes {
		er := ExportRecipe{
			Title:       rec.Title,
			Description: rec.Description,
			SourceURL:   rec.SourceURL,
			ImageFile:   rec.ImagePath,
			PrepTime:    rec.PrepTime,
			CookTime:    rec.CookTime,
			Servings:    rec.Servings,
			ContentMD:   rec.ContentMD,
			Author:      rec.AuthorName,
			CreatedAt:   rec.CreatedAt.Format("2006-01-02T15:04:05Z"),
		}
		if rec.Category != nil {
			er.Category = rec.Category.Name
		}
		for _, t := range rec.Tags {
			er.Tags = append(er.Tags, t.Name)
		}
		exported = append(exported, er)

		// Add image to ZIP
		if rec.ImagePath != "" {
			imgPath := filepath.Join(app.cfg.DataDir, "uploads", rec.ImagePath)
			if imgData, err := os.ReadFile(imgPath); err == nil {
				f, _ := zw.Create("images/" + rec.ImagePath)
				f.Write(imgData)
			}
		}
	}

	// Write JSON
	jsonFile, _ := zw.Create("recipes.json")
	enc := json.NewEncoder(jsonFile)
	enc.SetIndent("", "  ")
	enc.Encode(exported)
}

func (app *App) adminImportPage(w http.ResponseWriter, r *http.Request) {
	app.render(w, r, "admin_import.html", &PageData{Title: "Rezepte importieren"})
}

func (app *App) adminImportSubmit(w http.ResponseWriter, r *http.Request) {
	r.ParseMultipartForm(50 << 20) // 50 MB

	file, _, err := r.FormFile("import_file")
	if err != nil {
		app.render(w, r, "admin_import.html", &PageData{
			Title: "Rezepte importieren",
			Error: "Bitte eine ZIP-Datei auswaehlen.",
		})
		return
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		app.render(w, r, "admin_import.html", &PageData{
			Title: "Rezepte importieren",
			Error: "Fehler beim Lesen der Datei.",
		})
		return
	}

	zr, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		app.render(w, r, "admin_import.html", &PageData{
			Title: "Rezepte importieren",
			Error: "Ungueltige ZIP-Datei.",
		})
		return
	}

	// Read recipes.json
	var recipes []ExportRecipe
	for _, f := range zr.File {
		if f.Name == "recipes.json" {
			rc, err := f.Open()
			if err != nil {
				break
			}
			json.NewDecoder(rc).Decode(&recipes)
			rc.Close()
			break
		}
	}

	if len(recipes) == 0 {
		app.render(w, r, "admin_import.html", &PageData{
			Title: "Rezepte importieren",
			Error: "Keine Rezepte in der Datei gefunden.",
		})
		return
	}

	// Extract images
	uploadDir := filepath.Join(app.cfg.DataDir, "uploads")
	os.MkdirAll(uploadDir, 0755)
	for _, f := range zr.File {
		if strings.HasPrefix(f.Name, "images/") && f.Name != "images/" {
			rc, err := f.Open()
			if err != nil {
				continue
			}
			imgData, _ := io.ReadAll(rc)
			rc.Close()
			filename := strings.TrimPrefix(f.Name, "images/")
			os.WriteFile(filepath.Join(uploadDir, filename), imgData, 0644)
		}
	}

	// Import recipes
	user := ctxUser(r)
	imported := 0
	for _, er := range recipes {
		rec := &Recipe{
			Title:       er.Title,
			Description: er.Description,
			SourceURL:   er.SourceURL,
			ImagePath:   er.ImageFile,
			PrepTime:    er.PrepTime,
			CookTime:    er.CookTime,
			Servings:    er.Servings,
			ContentMD:   er.ContentMD,
			CreatedBy:   user.ID,
		}

		// Match category
		if er.Category != "" {
			catID, err := app.store.FindOrCreateCategory(er.Category)
			if err == nil {
				rec.CategoryID = &catID
			}
		}

		recipeID, err := app.store.CreateRecipe(rec)
		if err != nil {
			log.Printf("Import: Rezept '%s' fehlgeschlagen: %v", er.Title, err)
			continue
		}

		// Match/create tags
		if len(er.Tags) > 0 {
			app.saveTags(recipeID, strings.Join(er.Tags, ","))
		}
		imported++
	}

	log.Printf("Import: %d/%d Rezepte importiert", imported, len(recipes))
	app.setFlash(w, fmt.Sprintf(T(app.detectLang(r), "flash.imported_count"), imported))
	http.Redirect(w, r, "/admin/benutzer", http.StatusSeeOther)
}
