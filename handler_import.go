package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"golang.org/x/net/html"
)

func (app *App) importPage(w http.ResponseWriter, r *http.Request) {
	app.render(w, r, "import.html", &PageData{Title: "Rezept importieren"})
}

func (app *App) importSubmit(w http.ResponseWriter, r *http.Request) {
	url := strings.TrimSpace(r.FormValue("url"))
	if url == "" {
		app.render(w, r, "import.html", &PageData{Title: "Rezept importieren", Error: "Bitte eine URL eingeben."})
		return
	}

	recipe, err := fetchRecipeFromURL(url)
	if err != nil {
		app.render(w, r, "import.html", &PageData{
			Title: "Rezept importieren",
			Error: fmt.Sprintf("Fehler beim Importieren: %v", err),
		})
		return
	}

	recipe.SourceURL = url

	// Download image if available
	if recipe.ImagePath != "" {
		imgURL := recipe.ImagePath
		localPath, err := downloadImage(app.cfg.DataDir, imgURL)
		if err != nil {
			recipe.ImagePath = ""
		} else {
			recipe.ImagePath = localPath
		}
	}

	// Clear import-only fields
	recipe.AuthorName = ""
	recipe.CategoryID = nil
	recipe.Tags = nil

	cats, _ := app.store.ListCategories()
	allTags, _ := app.store.ListTags()

	// LLM: clean instructions
	if app.cfg.OpenAIKey != "" && app.store.GetSetting("llm_clean_instructions") != "0" {
		log.Printf("LLM: Bereinige Zubereitungsschritte fuer '%s'...", recipe.Title)
		cleaned, err := cleanInstructions(app.cfg.OpenAIKey, recipe)
		if err != nil {
			log.Printf("LLM: Fehler bei Schrittbereinigung: %v", err)
		} else {
			recipe = cleaned
			log.Printf("LLM: Zubereitungsschritte bereinigt")
		}
	}

	// LLM: rephrase instructions
	if app.cfg.OpenAIKey != "" && app.store.GetSetting("llm_rephrase_instructions") != "0" {
		log.Printf("LLM: Formuliere Zubereitungsschritte um fuer '%s'...", recipe.Title)
		rephrased, err := rephraseInstructions(app.cfg.OpenAIKey, recipe)
		if err != nil {
			log.Printf("LLM: Fehler bei Umformulierung: %v", err)
		} else {
			recipe = rephrased
			log.Printf("LLM: Zubereitungsschritte umformuliert")
		}
	}

	// LLM suggestion for category and tags
	if app.cfg.OpenAIKey != "" && app.store.GetSetting("llm_classify") != "0" {
		log.Printf("LLM: Frage OpenAI nach Vorschlaegen fuer '%s'...", recipe.Title)
		suggestion, err := suggestCategoryAndTags(app.cfg.OpenAIKey, recipe, cats, allTags)
		if err != nil {
			log.Printf("LLM: Fehler: %v", err)
		} else if suggestion != nil {
			log.Printf("LLM: Vorschlag Titel=%q, Kategorie=%q, Tags=%v", suggestion.Title, suggestion.CategoryName, suggestion.TagNames)
			// Apply cleaned title
			if suggestion.Title != "" && suggestion.Title != recipe.Title {
				log.Printf("LLM: Titel bereinigt '%s' -> '%s'", recipe.Title, suggestion.Title)
				recipe.Title = suggestion.Title
			}
			// Match suggested category
			for _, c := range cats {
				if strings.EqualFold(c.Name, suggestion.CategoryName) {
					id := c.ID
					recipe.CategoryID = &id
					log.Printf("LLM: Kategorie '%s' zugeordnet", c.Name)
					break
				}
			}
			// Match suggested tags
			for _, sugName := range suggestion.TagNames {
				matched := false
				for _, t := range allTags {
					if strings.EqualFold(t.Name, sugName) {
						recipe.Tags = append(recipe.Tags, t)
						matched = true
						break
					}
				}
				if !matched {
					log.Printf("LLM: Tag '%s' nicht in DB gefunden, uebersprungen", sugName)
				}
			}
			// Apply normalized/inferred times
			if suggestion.PrepTime != "" {
				if recipe.PrepTime != suggestion.PrepTime {
					log.Printf("LLM: PrepTime '%s' -> '%s'", recipe.PrepTime, suggestion.PrepTime)
				}
				recipe.PrepTime = suggestion.PrepTime
			}
			if suggestion.CookTime != "" {
				if recipe.CookTime != suggestion.CookTime {
					log.Printf("LLM: CookTime '%s' -> '%s'", recipe.CookTime, suggestion.CookTime)
				}
				recipe.CookTime = suggestion.CookTime
			}
			log.Printf("LLM: %d Tags zugeordnet", len(recipe.Tags))
		}
	}

	app.render(w, r, "recipe_form.html", &PageData{
		Title:      "Importiertes Rezept",
		Recipe:     *recipe,
		Categories: cats,
	})
}

func fetchRecipeFromURL(url string) (*Recipe, error) {
	client := &http.Client{Timeout: 15 * time.Second}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("ungueltige URL")
	}
	req.Header.Set("User-Agent", "Yummi/1.0 Recipe Importer")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("Seite nicht erreichbar")
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("HTTP-Status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, 5*1024*1024))
	if err != nil {
		return nil, fmt.Errorf("Fehler beim Lesen der Seite")
	}

	return parseRecipeFromHTML(string(body))
}

func parseRecipeFromHTML(htmlContent string) (*Recipe, error) {
	jsonLDs := extractJSONLD(htmlContent)

	for _, jsonStr := range jsonLDs {
		recipe, err := parseRecipeJSON(jsonStr)
		if err == nil && recipe != nil {
			return recipe, nil
		}
	}

	return nil, fmt.Errorf("Kein Rezept auf dieser Seite gefunden. Die Seite muss Schema.org Rezept-Markup enthalten.")
}

func extractJSONLD(htmlContent string) []string {
	tokenizer := html.NewTokenizer(strings.NewReader(htmlContent))
	var results []string
	inScript := false

	for {
		tt := tokenizer.Next()
		switch tt {
		case html.ErrorToken:
			return results
		case html.StartTagToken:
			t := tokenizer.Token()
			if t.Data == "script" {
				for _, attr := range t.Attr {
					if attr.Key == "type" && attr.Val == "application/ld+json" {
						inScript = true
						break
					}
				}
			}
		case html.TextToken:
			if inScript {
				text := strings.TrimSpace(tokenizer.Token().Data)
				if text != "" {
					results = append(results, text)
				}
			}
		case html.EndTagToken:
			if inScript {
				inScript = false
			}
		}
	}
}

func parseRecipeJSON(jsonStr string) (*Recipe, error) {
	// Try as single object first
	var obj map[string]any
	if err := json.Unmarshal([]byte(jsonStr), &obj); err == nil {
		if recipe := extractRecipeFromObject(obj); recipe != nil {
			return recipe, nil
		}
	}

	// Try as array
	var arr []map[string]any
	if err := json.Unmarshal([]byte(jsonStr), &arr); err == nil {
		for _, item := range arr {
			if recipe := extractRecipeFromObject(item); recipe != nil {
				return recipe, nil
			}
		}
	}

	return nil, fmt.Errorf("kein Rezept gefunden")
}

func extractRecipeFromObject(obj map[string]any) *Recipe {
	// Check @graph
	if graph, ok := obj["@graph"]; ok {
		if items, ok := graph.([]any); ok {
			for _, item := range items {
				if m, ok := item.(map[string]any); ok {
					if r := extractRecipeFromObject(m); r != nil {
						return r
					}
				}
			}
		}
	}

	// Check @type
	typeStr := getStringField(obj, "@type")
	if typeStr != "Recipe" {
		// Could be an array of types
		if types, ok := obj["@type"].([]any); ok {
			found := false
			for _, t := range types {
				if s, ok := t.(string); ok && s == "Recipe" {
					found = true
					break
				}
			}
			if !found {
				return nil
			}
		} else {
			return nil
		}
	}

	recipe := &Recipe{}
	recipe.Title = getStringField(obj, "name")
	recipe.Description = cleanDescription(getStringField(obj, "description"))
	recipe.Servings = getStringField(obj, "recipeYield")
	recipe.PrepTime = formatISODuration(getStringField(obj, "prepTime"))
	recipe.CookTime = formatISODuration(getStringField(obj, "cookTime"))

	// Image
	recipe.ImagePath = extractImageURL(obj)

	// Category (stored temporarily in AuthorName for import flow)
	recipe.AuthorName = getStringField(obj, "recipeCategory")

	// Build markdown content
	var md strings.Builder

	// Ingredients
	if ingredients := getStringArray(obj, "recipeIngredient"); len(ingredients) > 0 {
		md.WriteString("## Zutaten\n\n")
		for _, ing := range ingredients {
			md.WriteString("- " + strings.TrimSpace(ing) + "\n")
		}
		md.WriteString("\n")
	}

	// Instructions
	instructions := extractInstructions(obj)
	if len(instructions) > 0 {
		md.WriteString("## Zubereitung\n\n")
		for i, step := range instructions {
			md.WriteString(fmt.Sprintf("%d. %s\n", i+1, strings.TrimSpace(step)))
		}
		md.WriteString("\n")
	}

	// Notes
	if notes := getStringField(obj, "recipeNotes"); notes != "" {
		md.WriteString("## Notizen\n\n" + notes + "\n")
	}

	recipe.ContentMD = md.String()

	// Tags from keywords
	if keywords := getStringField(obj, "keywords"); keywords != "" {
		parts := strings.Split(keywords, ",")
		for _, p := range parts {
			name := strings.TrimSpace(p)
			if name != "" {
				recipe.Tags = append(recipe.Tags, Tag{Name: name})
			}
		}
	}

	return recipe
}

func getStringField(obj map[string]any, key string) string {
	v, ok := obj[key]
	if !ok {
		return ""
	}
	switch val := v.(type) {
	case string:
		return val
	case []any:
		if len(val) > 0 {
			if s, ok := val[0].(string); ok {
				return s
			}
		}
	}
	return fmt.Sprintf("%v", v)
}

func getStringArray(obj map[string]any, key string) []string {
	v, ok := obj[key]
	if !ok {
		return nil
	}
	arr, ok := v.([]any)
	if !ok {
		return nil
	}
	var result []string
	for _, item := range arr {
		if s, ok := item.(string); ok {
			result = append(result, s)
		}
	}
	return result
}

func extractInstructions(obj map[string]any) []string {
	v, ok := obj["recipeInstructions"]
	if !ok {
		return nil
	}

	switch val := v.(type) {
	case string:
		return splitInstructionText(val)
	case []any:
		var steps []string
		for _, item := range val {
			switch step := item.(type) {
			case string:
				steps = append(steps, step)
			case map[string]any:
				stepType := getStringField(step, "@type")
				// HowToSection: only extract sub-steps from itemListElement
				if stepType == "HowToSection" {
					if items, ok := step["itemListElement"].([]any); ok {
						for _, subItem := range items {
							if m, ok := subItem.(map[string]any); ok {
								if text := getStringField(m, "text"); text != "" {
									steps = append(steps, text)
								}
							}
						}
					}
				} else {
					// HowToStep or untyped
					if text := getStringField(step, "text"); text != "" {
						steps = append(steps, text)
					} else if name := getStringField(step, "name"); name != "" {
						steps = append(steps, name)
					}
				}
			}
		}
		return steps
	}
	return nil
}

func splitInstructionText(text string) []string {
	lines := strings.Split(text, "\n")
	var steps []string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" {
			steps = append(steps, line)
		}
	}
	return steps
}

func extractImageURL(obj map[string]any) string {
	v, ok := obj["image"]
	if !ok {
		return ""
	}
	switch val := v.(type) {
	case string:
		return val
	case map[string]any:
		return getStringField(val, "url")
	case []any:
		if len(val) > 0 {
			switch first := val[0].(type) {
			case string:
				return first
			case map[string]any:
				return getStringField(first, "url")
			}
		}
	}
	return ""
}

func cleanDescription(desc string) string {
	// Remove common SEO spam from recipe sites
	junk := []string{
		"Portionsrechner",
		"Kochbuch",
		"Video-Tipps",
		"Jetzt entdecken und ausprobieren",
		"schmackhaft befunden",
		"Bewertungen und für",
	}
	for _, j := range junk {
		if strings.Contains(desc, j) {
			// Cut at the first junk match
			idx := strings.Index(desc, "Über")
			if idx > 0 {
				desc = strings.TrimSpace(desc[:idx])
			} else {
				// Try cutting at ►
				idx = strings.Index(desc, "►")
				if idx > 0 {
					desc = strings.TrimSpace(desc[:idx])
				}
			}
			break
		}
	}
	// Strip trailing punctuation artifacts
	desc = strings.TrimRight(desc, " .!,;:-–")
	return strings.TrimSpace(desc)
}

func formatISODuration(iso string) string {
	if iso == "" || !strings.HasPrefix(iso, "PT") {
		if iso != "" && !strings.HasPrefix(iso, "P") {
			return iso
		}
		if iso == "" {
			return ""
		}
	}

	iso = strings.TrimPrefix(iso, "P")
	iso = strings.TrimPrefix(iso, "T")

	var parts []string
	current := ""
	for _, c := range iso {
		if c >= '0' && c <= '9' {
			current += string(c)
		} else {
			switch c {
			case 'H':
				if current != "" {
					parts = append(parts, current+" Std.")
				}
			case 'M':
				if current != "" {
					parts = append(parts, current+" Min.")
				}
			case 'S':
				if current != "" {
					parts = append(parts, current+" Sek.")
				}
			}
			current = ""
		}
	}

	if len(parts) == 0 {
		return iso
	}
	return strings.Join(parts, " ")
}
