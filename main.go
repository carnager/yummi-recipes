package main

import (
	"embed"
	"html/template"
	"io/fs"
	"log"
	"net/http"
	"strings"
	"time"
)

//go:embed templates/*.html templates/partials/*.html
var templatesFS embed.FS

//go:embed static/*
var staticFS embed.FS

type App struct {
	cfg       Config
	store     *Store
	templates map[string]*template.Template
}

func main() {
	cfg := loadConfig()
	db := openDB(cfg.DataDir)
	runMigrations(db)
	store := &Store{db: db}

	// Ensure at least one admin exists — promote first user if none
	store.EnsureAdmin()

	// CLI: create admin user
	if cfg.CreateAdmin != "" {
		parts := strings.SplitN(cfg.CreateAdmin, ":", 2)
		if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
			log.Fatal("Verwendung: -create-admin benutzername:passwort")
		}
		hash, err := hashPassword(parts[1])
		if err != nil {
			log.Fatalf("Passwort-Hash fehlgeschlagen: %v", err)
		}
		userID, err := store.CreateUser(parts[0], parts[0], hash)
		if err != nil {
			log.Fatalf("Benutzer erstellen fehlgeschlagen: %v", err)
		}
		db.Exec("UPDATE users SET is_admin = 1 WHERE id = ?", userID)
		log.Printf("Admin-Benutzer '%s' erstellt.", parts[0])
		return
	}

	app := &App{
		cfg:   cfg,
		store: store,
	}
	app.loadTemplates()

	mux := http.NewServeMux()

	// Static files
	staticSub, _ := fs.Sub(staticFS, "static")
	mux.Handle("GET /static/", http.StripPrefix("/static/", http.FileServer(http.FS(staticSub))))
	mux.Handle("GET /uploads/", http.StripPrefix("/uploads/", http.FileServer(http.Dir(cfg.DataDir+"/uploads"))))

	// Auth routes
	mux.HandleFunc("GET /anmelden", app.loginPage)
	mux.HandleFunc("POST /anmelden", app.loginSubmit)
	mux.HandleFunc("GET /registrieren", app.registerPage)
	mux.HandleFunc("POST /registrieren", app.registerSubmit)
	mux.HandleFunc("POST /abmelden", app.logoutSubmit)

	// Public recipe routes (read-only)
	mux.HandleFunc("GET /rezepte", app.recipeList)
	mux.HandleFunc("GET /rezepte/{id}", app.recipeDetail)
	mux.HandleFunc("GET /kategorien", app.categoryList)
	mux.HandleFunc("GET /kategorien/{slug}", app.categoryRecipes)
	mux.HandleFunc("GET /tags/{slug}", app.tagRecipes)
	mux.HandleFunc("GET /suche", app.searchRecipes)

	// Authenticated recipe routes (write)
	mux.HandleFunc("GET /rezepte/neu", requireAuth(app.recipeNew))
	mux.HandleFunc("GET /rezepte/neu/manuell", requireAuth(app.recipeNewManual))
	mux.HandleFunc("GET /rezepte/neu/import", requireAuth(app.importPage))
	mux.HandleFunc("POST /rezepte/neu/import", requireAuth(app.importSubmit))
	mux.HandleFunc("GET /rezepte/neu/llm", requireAuth(app.recipeLLMPage))
	mux.HandleFunc("POST /rezepte/neu/llm", requireAuth(app.recipeLLMSubmit))
	mux.HandleFunc("POST /rezepte", requireAuth(app.recipeCreate))
	mux.HandleFunc("GET /rezepte/{id}/bearbeiten", requireAuth(app.recipeEdit))
	mux.HandleFunc("POST /rezepte/{id}/bearbeiten", requireAuth(app.recipeUpdate))
	mux.HandleFunc("DELETE /rezepte/{id}", requireAuth(app.recipeDelete))

	// User features (authenticated)
	mux.HandleFunc("POST /rezepte/{id}/probiert", requireAuth(app.toggleTried))
	mux.HandleFunc("POST /rezepte/{id}/teilen", requireAuth(app.shareRecipe))
	mux.HandleFunc("GET /meine-rezepte", requireAuth(app.myRecipes))
	mux.HandleFunc("POST /meine-rezepte/entfernen", requireAuth(app.removeFromMyRecipes))
	mux.HandleFunc("GET /api/tags/suggest", requireAuth(app.suggestTags))

	// Admin
	mux.HandleFunc("GET /admin/benutzer", requireAdmin(app.adminUserList))
	mux.HandleFunc("GET /admin/benutzer/neu", requireAdmin(app.adminUserNewPage))
	mux.HandleFunc("POST /admin/benutzer/neu", requireAdmin(app.adminUserNewSubmit))
	mux.HandleFunc("GET /admin/benutzer/{id}", requireAdmin(app.adminUserEditPage))
	mux.HandleFunc("POST /admin/benutzer/{id}", requireAdmin(app.adminUserEditSubmit))
	mux.HandleFunc("POST /admin/benutzer/{id}/loeschen", requireAdmin(app.adminUserDelete))
	mux.HandleFunc("POST /admin/einstellungen", requireAdmin(app.adminSettings))
	mux.HandleFunc("GET /admin/export", requireAdmin(app.adminExport))
	mux.HandleFunc("GET /admin/import", requireAdmin(app.adminImportPage))
	mux.HandleFunc("POST /admin/import", requireAdmin(app.adminImportSubmit))

	// Language switch
	mux.HandleFunc("GET /sprache/{lang}", func(w http.ResponseWriter, r *http.Request) {
		lang := r.PathValue("lang")
		if lang == "en" || lang == "de" {
			setLangCookie(w, lang)
		}
		ref := r.Header.Get("Referer")
		if ref == "" {
			ref = "/rezepte"
		}
		http.Redirect(w, r, ref, http.StatusSeeOther)
	})

	// Home redirect
	mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		http.Redirect(w, r, "/rezepte", http.StatusSeeOther)
	})

	// --- API routes ---
	mux.HandleFunc("POST /api/v1/auth/login", app.apiLogin)
	mux.HandleFunc("POST /api/v1/auth/register", app.apiRegister)
	mux.HandleFunc("GET /api/v1/recipes", requireAuth(app.apiListRecipes))
	mux.HandleFunc("POST /api/v1/recipes", requireAuth(app.apiCreateRecipe))
	mux.HandleFunc("GET /api/v1/recipes/{id}", requireAuth(app.apiGetRecipe))
	mux.HandleFunc("PUT /api/v1/recipes/{id}", requireAuth(app.apiUpdateRecipe))
	mux.HandleFunc("DELETE /api/v1/recipes/{id}", requireAuth(app.apiDeleteRecipe))
	mux.HandleFunc("POST /api/v1/recipes/import", requireAuth(app.apiImportRecipe))
	mux.HandleFunc("POST /api/v1/recipes/{id}/tried", requireAuth(app.apiToggleTried))
	mux.HandleFunc("POST /api/v1/recipes/{id}/share", requireAuth(app.apiShareRecipe))
	mux.HandleFunc("GET /api/v1/my-recipes", requireAuth(app.apiMyRecipes))
	mux.HandleFunc("GET /api/v1/recipes/{id}/shares", requireAuth(app.apiListSharesForRecipe))
	mux.HandleFunc("GET /api/v1/categories", requireAuth(app.apiListCategories))
	mux.HandleFunc("GET /api/v1/tags", requireAuth(app.apiListTags))
	mux.HandleFunc("GET /api/v1/users", requireAuth(app.apiListUsers))

	// Session cleanup
	go func() {
		for {
			time.Sleep(1 * time.Hour)
			store.CleanExpiredSessions()
		}
	}()

	handler := app.authMiddleware(logMiddleware(mux))

	log.Printf("Yummi laeuft auf %s", cfg.Port)
	log.Fatal(http.ListenAndServe(cfg.Port, handler))
}

func (app *App) loadTemplates() {
	funcMap := templateFuncs()
	app.templates = make(map[string]*template.Template)

	pages := []string{
		"login.html", "register.html",
		"recipe_list.html", "recipe_detail.html", "recipe_form.html",
		"recipe_new.html", "recipe_llm.html",
		"import.html", "category_list.html", "user_recipes.html",
		"admin_users.html", "admin_user_form.html", "admin_import.html",
	}

	layoutBytes, err := templatesFS.ReadFile("templates/layout.html")
	if err != nil {
		log.Fatalf("Layout-Template lesen: %v", err)
	}
	layoutStr := string(layoutBytes)

	// Load all partials
	partialEntries, _ := templatesFS.ReadDir("templates/partials")
	var partialSources []string
	for _, entry := range partialEntries {
		data, err := templatesFS.ReadFile("templates/partials/" + entry.Name())
		if err != nil {
			log.Fatalf("Partial %s lesen: %v", entry.Name(), err)
		}
		partialSources = append(partialSources, string(data))
	}

	for _, page := range pages {
		pageBytes, err := templatesFS.ReadFile("templates/" + page)
		if err != nil {
			log.Fatalf("Template %s lesen: %v", page, err)
		}

		t, err := template.New("layout").Funcs(funcMap).Parse(layoutStr)
		if err != nil {
			log.Fatalf("Layout parsen: %v", err)
		}

		// Parse all partials into each template
		for _, partial := range partialSources {
			t, err = t.Parse(partial)
			if err != nil {
				log.Fatalf("Partial parsen: %v", err)
			}
		}

		t, err = t.Parse(string(pageBytes))
		if err != nil {
			log.Fatalf("Template %s parsen: %v", page, err)
		}

		app.templates[page] = t
	}
}
