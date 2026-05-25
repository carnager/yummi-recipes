package main

import (
	"net/http"
	"strings"
)

type Lang string

const (
	LangDE Lang = "de"
	LangEN Lang = "en"
)

var translations = map[string]map[Lang]string{
	// Layout / Navigation
	"nav.recipes":    {LangDE: "Rezepte", LangEN: "Recipes"},
	"nav.categories": {LangDE: "Kategorien", LangEN: "Categories"},
	"nav.my_recipes": {LangDE: "Meine Rezepte", LangEN: "My Recipes"},
	"nav.import":     {LangDE: "Import", LangEN: "Import"},
	"nav.admin":      {LangDE: "Admin", LangEN: "Admin"},
	"nav.logout":     {LangDE: "Abmelden", LangEN: "Logout"},
	"nav.login":      {LangDE: "Anmelden", LangEN: "Login"},
	"nav.register":   {LangDE: "Registrieren", LangEN: "Register"},
	"footer.tagline": {LangDE: "Einfach lecker kochen", LangEN: "Simply delicious cooking"},

	// Auth
	"auth.login":             {LangDE: "Anmelden", LangEN: "Login"},
	"auth.register":          {LangDE: "Registrieren", LangEN: "Register"},
	"auth.username":          {LangDE: "Benutzername", LangEN: "Username"},
	"auth.password":          {LangDE: "Passwort", LangEN: "Password"},
	"auth.display_name":      {LangDE: "Anzeigename", LangEN: "Display Name"},
	"auth.login_title":       {LangDE: "Anmelden", LangEN: "Sign In"},
	"auth.register_title":    {LangDE: "Registrieren", LangEN: "Create Account"},
	"auth.no_account":        {LangDE: "Noch kein Konto?", LangEN: "No account yet?"},
	"auth.has_account":       {LangDE: "Bereits ein Konto?", LangEN: "Already have an account?"},
	"auth.err_invalid":       {LangDE: "Benutzername oder Passwort falsch.", LangEN: "Invalid username or password."},
	"auth.err_exists":        {LangDE: "Benutzername bereits vergeben.", LangEN: "Username already taken."},
	"auth.err_fields":        {LangDE: "Bitte alle Felder ausfuellen.", LangEN: "Please fill in all fields."},
	"auth.err_disabled":      {LangDE: "Registrierung ist deaktiviert.", LangEN: "Registration is disabled."},
	"auth.err_password_len":  {LangDE: "Passwort muss mindestens 6 Zeichen lang sein.", LangEN: "Password must be at least 6 characters."},
	"auth.err_password_match": {LangDE: "Passwoerter stimmen nicht ueberein.", LangEN: "Passwords do not match."},
	"auth.err_create":        {LangDE: "Fehler beim Erstellen des Kontos.", LangEN: "Error creating account."},

	// Recipes
	"recipe.new":             {LangDE: "Neues Rezept", LangEN: "New Recipe"},
	"recipe.edit":            {LangDE: "Rezept bearbeiten", LangEN: "Edit Recipe"},
	"recipe.delete":          {LangDE: "Loeschen", LangEN: "Delete"},
	"recipe.delete_confirm":  {LangDE: "Rezept wirklich loeschen?", LangEN: "Really delete this recipe?"},
	"recipe.save":            {LangDE: "Speichern", LangEN: "Save"},
	"recipe.cancel":          {LangDE: "Abbrechen", LangEN: "Cancel"},
	"recipe.title":           {LangDE: "Titel", LangEN: "Title"},
	"recipe.description":     {LangDE: "Beschreibung", LangEN: "Description"},
	"recipe.category":        {LangDE: "Kategorie", LangEN: "Category"},
	"recipe.no_category":     {LangDE: "Keine Kategorie", LangEN: "No Category"},
	"recipe.tags":            {LangDE: "Tags", LangEN: "Tags"},
	"recipe.tags_comma":      {LangDE: "Tags (kommagetrennt)", LangEN: "Tags (comma-separated)"},
	"recipe.prep_time":       {LangDE: "Vorbereitung", LangEN: "Prep Time"},
	"recipe.cook_time":       {LangDE: "Kochzeit", LangEN: "Cook Time"},
	"recipe.servings":        {LangDE: "Portionen", LangEN: "Servings"},
	"recipe.source_url":      {LangDE: "Quell-URL", LangEN: "Source URL"},
	"recipe.content":         {LangDE: "Inhalt (Markdown)", LangEN: "Content (Markdown)"},
	"recipe.tried":           {LangDE: "Probiert", LangEN: "Tried"},
	"recipe.not_tried":       {LangDE: "Noch nicht probiert", LangEN: "Not tried yet"},
	"recipe.share":           {LangDE: "Teilen", LangEN: "Share"},
	"recipe.edit_btn":        {LangDE: "Bearbeiten", LangEN: "Edit"},
	"recipe.source":          {LangDE: "Quelle", LangEN: "Source"},
	"recipe.none_found":      {LangDE: "Keine Rezepte gefunden", LangEN: "No recipes found"},
	"recipe.none_yet":        {LangDE: "Noch keine Rezepte", LangEN: "No recipes yet"},

	// Import
	"import.title":           {LangDE: "Rezept importieren", LangEN: "Import Recipe"},
	"import.url_label":       {LangDE: "URL eingeben", LangEN: "Enter URL"},
	"import.url_placeholder": {LangDE: "https://chefkoch.de/...", LangEN: "https://example.com/recipe/..."},
	"import.submit":          {LangDE: "Importieren", LangEN: "Import"},
	"import.manual":          {LangDE: "Manuell erstellen", LangEN: "Create Manually"},
	"import.manual_desc":     {LangDE: "Rezept selbst eingeben", LangEN: "Enter recipe yourself"},
	"import.url_desc":        {LangDE: "Rezept von einer Webseite importieren", LangEN: "Import recipe from a website"},
	"import.llm_title":       {LangDE: "KI-Import", LangEN: "AI Import"},
	"import.llm_desc":        {LangDE: "Rezept aus Text oder Foto extrahieren", LangEN: "Extract recipe from text or photo"},
	"import.llm_text":        {LangDE: "Rezepttext einfuegen", LangEN: "Paste recipe text"},
	"import.llm_photo":       {LangDE: "Foto hochladen", LangEN: "Upload photo"},

	// Categories
	"categories.title":       {LangDE: "Kategorien", LangEN: "Categories"},
	"categories.all":         {LangDE: "Alle Rezepte", LangEN: "All Recipes"},

	// My Recipes
	"my.title":               {LangDE: "Meine Rezepte", LangEN: "My Recipes"},
	"my.own":                 {LangDE: "Eigene Rezepte", LangEN: "Own Recipes"},
	"my.shared":              {LangDE: "Mit mir geteilt", LangEN: "Shared with me"},
	"my.tried":               {LangDE: "Probierte Rezepte", LangEN: "Tried Recipes"},

	// Share
	"share.title":            {LangDE: "Rezept teilen", LangEN: "Share Recipe"},
	"share.with_users":       {LangDE: "Mit Benutzern teilen", LangEN: "Share with users"},
	"share.no_users":         {LangDE: "Keine anderen Benutzer vorhanden", LangEN: "No other users available"},
	"share.link":             {LangDE: "Link teilen", LangEN: "Share link"},

	// Admin
	"admin.users":            {LangDE: "Benutzerverwaltung", LangEN: "User Management"},
	"admin.new_user":         {LangDE: "Benutzer anlegen", LangEN: "Create User"},
	"admin.edit_user":        {LangDE: "Benutzer bearbeiten", LangEN: "Edit User"},
	"admin.delete_user":      {LangDE: "Loeschen", LangEN: "Delete"},
	"admin.delete_confirm":   {LangDE: "Benutzer wirklich loeschen?", LangEN: "Really delete this user?"},
	"admin.is_admin":         {LangDE: "Admin", LangEN: "Admin"},
	"admin.registered":       {LangDE: "Registriert", LangEN: "Registered"},
	"admin.actions":          {LangDE: "Aktionen", LangEN: "Actions"},
	"admin.edit":             {LangDE: "Bearbeiten", LangEN: "Edit"},
	"admin.reg_enabled":      {LangDE: "Registrierung erlauben", LangEN: "Allow registration"},
	"admin.llm_classify":     {LangDE: "KI: Kategorie & Tags vorschlagen", LangEN: "AI: Suggest category & tags"},
	"admin.llm_clean":        {LangDE: "KI: Zubereitungsschritte bereinigen", LangEN: "AI: Clean up instructions"},
	"admin.llm_rephrase":     {LangDE: "KI: Schritte vereinheitlichen & umformulieren", LangEN: "AI: Standardize & rephrase steps"},
	"admin.export":           {LangDE: "Export (ZIP)", LangEN: "Export (ZIP)"},
	"admin.import":           {LangDE: "Import", LangEN: "Import"},
	"admin.password_new":     {LangDE: "Neues Passwort (leer lassen = nicht aendern)", LangEN: "New password (leave empty = no change)"},

	// Search
	"search.placeholder":     {LangDE: "Rezepte suchen...", LangEN: "Search recipes..."},
	"search.clear":           {LangDE: "Suche leeren", LangEN: "Clear search"},

	// General
	"general.yes":            {LangDE: "Ja", LangEN: "Yes"},
	"general.no":             {LangDE: "Nein", LangEN: "No"},
	"general.error":          {LangDE: "Fehler", LangEN: "Error"},
	"general.save":           {LangDE: "Speichern", LangEN: "Save"},
	"general.back":           {LangDE: "Zurueck", LangEN: "Back"},

	// Flash messages
	"flash.recipe_created":   {LangDE: "Rezept erstellt.", LangEN: "Recipe created."},
	"flash.recipe_updated":   {LangDE: "Rezept aktualisiert.", LangEN: "Recipe updated."},
	"flash.recipe_deleted":   {LangDE: "Rezept geloescht.", LangEN: "Recipe deleted."},
	"flash.user_created":     {LangDE: "Benutzer erstellt.", LangEN: "User created."},
	"flash.user_updated":     {LangDE: "Benutzer aktualisiert.", LangEN: "User updated."},
	"flash.user_deleted":     {LangDE: "Benutzer geloescht.", LangEN: "User deleted."},
	"flash.shared":           {LangDE: "Rezept geteilt.", LangEN: "Recipe shared."},
	"flash.unshared":         {LangDE: "Teilen aufgehoben.", LangEN: "Share removed."},
	"flash.import_error":     {LangDE: "Fehler beim Importieren", LangEN: "Import error"},
	"flash.connection_error": {LangDE: "Verbindungsfehler", LangEN: "Connection error"},
	"flash.cant_delete_self": {LangDE: "Du kannst dich nicht selbst loeschen.", LangEN: "You cannot delete yourself."},
	"flash.imported_count":   {LangDE: "%d Rezepte importiert.", LangEN: "%d recipes imported."},
}

// T returns the translated string for the given key and language.
// Falls back to German if key not found in requested language.
func T(lang Lang, key string) string {
	if m, ok := translations[key]; ok {
		if s, ok := m[lang]; ok {
			return s
		}
		if s, ok := m[LangDE]; ok {
			return s
		}
	}
	return key
}

// detectLang determines the language from the request.
// Priority: 1) cookie "lang", 2) Accept-Language header, 3) global setting
func (app *App) detectLang(r *http.Request) Lang {
	// Check cookie
	if cookie, err := r.Cookie("lang"); err == nil {
		switch cookie.Value {
		case "en":
			return LangEN
		case "de":
			return LangDE
		}
	}

	// Check Accept-Language header
	accept := r.Header.Get("Accept-Language")
	if accept != "" {
		for _, part := range strings.Split(accept, ",") {
			lang := strings.TrimSpace(strings.SplitN(part, ";", 2)[0])
			if strings.HasPrefix(lang, "en") {
				return LangEN
			}
			if strings.HasPrefix(lang, "de") {
				return LangDE
			}
		}
	}

	// Default
	return LangDE
}

// setLangCookie sets the language preference cookie
func setLangCookie(w http.ResponseWriter, lang string) {
	http.SetCookie(w, &http.Cookie{
		Name:     "lang",
		Value:    lang,
		Path:     "/",
		MaxAge:   365 * 24 * 60 * 60,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
}
