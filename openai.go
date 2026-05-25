package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type LLMSuggestion struct {
	Title        string   `json:"title"`
	CategoryName string   `json:"category"`
	TagNames     []string `json:"tags"`
	PrepTime     string   `json:"prep_time"`
	CookTime     string   `json:"cook_time"`
}

func suggestCategoryAndTags(apiKey string, recipe *Recipe, categories []Category, tags []Tag) (*LLMSuggestion, error) {
	if apiKey == "" {
		return nil, fmt.Errorf("kein OpenAI API-Key konfiguriert")
	}

	// Build category list
	var catNames []string
	for _, c := range categories {
		catNames = append(catNames, c.Name)
	}

	// Build tag list
	var tagNames []string
	for _, t := range tags {
		tagNames = append(tagNames, t.Name)
	}

	// Build recipe context (trim to avoid huge prompts)
	recipeText := fmt.Sprintf("Titel: %s\nBeschreibung: %s\n\n%s",
		recipe.Title, recipe.Description, recipe.ContentMD)
	if len(recipeText) > 4000 {
		recipeText = recipeText[:4000]
	}

	prompt := fmt.Sprintf(`Du bist ein Rezept-Klassifikator. Analysiere das folgende Rezept und schlage eine passende Kategorie und Tags vor.

Bereinige den Titel: entferne Benutzernamen (z.B. "von foobar", "by chef123"), Sonderzeichen-Spam, Seitennamen und andere Artefakte. Der Titel soll nur den Gerichtnamen enthalten.

Zeitangaben: Normalisiere prep_time und cook_time ins Format "X Min." oder "X Std. Y Min." (aber nie "0 Std." — lass die Stunden weg wenn 0). Falls eine Zeit im Rezept fehlt aber in den Zubereitungsschritten erwaehnt wird (z.B. "30 Minuten kochen"), trage sie ein.

Wichtig zur Kategorie-Unterscheidung:
- "Backen" = Kuchen, Kekse, Plätzchen, Brot, Gebäck, Torten, Muffins (alles was man im Ofen backt als Hauptzweck)
- "Desserts" = Nachspeisen die man NICHT backt: Mousse, Pudding, Eis, Panna Cotta, Tiramisu, Cremes

Verfügbare Kategorien (wähle genau EINE):
%s

Verfügbare Tags (wähle 3-8 passende):
%s

Rezept:
%s

Antworte NUR mit einem JSON-Objekt in diesem Format, ohne weitere Erklärung:
{"title": "Bereinigter Titel", "category": "Name der Kategorie", "tags": ["Tag1", "Tag2", "Tag3"], "prep_time": "X Min.", "cook_time": "X Min."}`,
		strings.Join(catNames, ", "),
		strings.Join(tagNames, ", "),
		recipeText,
	)

	reqBody, _ := json.Marshal(map[string]any{
		"model": "gpt-4.1-nano",
		"messages": []map[string]string{
			{"role": "user", "content": prompt},
		},
		"temperature": 0.3,
		"max_tokens":  200,
	})

	req, err := http.NewRequest("POST", "https://api.openai.com/v1/chat/completions", bytes.NewReader(reqBody))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("OpenAI-Anfrage fehlgeschlagen: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("OpenAI HTTP %d: %s", resp.StatusCode, string(body))
	}

	var result struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}
	if len(result.Choices) == 0 {
		return nil, fmt.Errorf("keine Antwort von OpenAI")
	}

	content := strings.TrimSpace(result.Choices[0].Message.Content)

	// Strip markdown code fences if present
	content = strings.TrimPrefix(content, "```json")
	content = strings.TrimPrefix(content, "```")
	content = strings.TrimSuffix(content, "```")
	content = strings.TrimSpace(content)

	var suggestion LLMSuggestion
	if err := json.Unmarshal([]byte(content), &suggestion); err != nil {
		return nil, fmt.Errorf("OpenAI-Antwort konnte nicht geparst werden: %s", content)
	}

	return &suggestion, nil
}

func cleanInstructions(apiKey string, recipe *Recipe) (*Recipe, error) {
	if apiKey == "" {
		return recipe, nil
	}

	prompt := fmt.Sprintf(`Du bist ein Rezept-Editor. Bereinige die folgenden Zubereitungsschritte eines Rezepts.

Regeln:
- Entferne Schritte die nur Ueberschriften/Abschnittsnamen sind (z.B. "Zubereitung", "Fuellung", "Teig") und keine echte Anweisung enthalten
- Wenn ein Abschnittsname relevant ist, integriere ihn in den naechsten Schritt (z.B. "Fuer den Teig: Mehl und Zucker mischen")
- Korrigiere offensichtliche Formatierungsfehler (abgeschnittene Saetze, HTML-Artefakte)
- Behalte die Reihenfolge bei
- Aendere den Inhalt der Schritte NICHT inhaltlich, nur strukturell bereinigen
- Gib die bereinigten Schritte als JSON-Array von Strings zurueck

Rezepttitel: %s

Aktuelle Schritte:
%s

Antworte NUR mit einem JSON-Array, ohne weitere Erklaerung:
["Schritt 1", "Schritt 2", ...]`,
		recipe.Title,
		recipe.ContentMD,
	)

	reqBody, _ := json.Marshal(map[string]any{
		"model": "gpt-4.1-nano",
		"messages": []map[string]string{
			{"role": "user", "content": prompt},
		},
		"temperature": 0.2,
		"max_tokens":  2000,
	})

	req, err := http.NewRequest("POST", "https://api.openai.com/v1/chat/completions", bytes.NewReader(reqBody))
	if err != nil {
		return recipe, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return recipe, fmt.Errorf("OpenAI-Anfrage fehlgeschlagen: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return recipe, err
	}

	if resp.StatusCode != 200 {
		return recipe, fmt.Errorf("OpenAI HTTP %d: %s", resp.StatusCode, string(body))
	}

	var result struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return recipe, err
	}
	if len(result.Choices) == 0 {
		return recipe, fmt.Errorf("keine Antwort von OpenAI")
	}

	content := strings.TrimSpace(result.Choices[0].Message.Content)
	content = strings.TrimPrefix(content, "```json")
	content = strings.TrimPrefix(content, "```")
	content = strings.TrimSuffix(content, "```")
	content = strings.TrimSpace(content)

	var steps []string
	if err := json.Unmarshal([]byte(content), &steps); err != nil {
		return recipe, fmt.Errorf("KI-Antwort konnte nicht geparst werden: %s", content)
	}

	// Rebuild markdown with cleaned steps
	var md strings.Builder

	// Preserve ingredients section
	if idx := strings.Index(recipe.ContentMD, "## Zubereitung"); idx > 0 {
		md.WriteString(recipe.ContentMD[:idx])
	} else if idx := strings.Index(recipe.ContentMD, "## Zutaten"); idx >= 0 {
		// Find end of ingredients section
		afterIngredients := recipe.ContentMD[idx:]
		if nextSection := strings.Index(afterIngredients[1:], "## "); nextSection > 0 {
			md.WriteString(recipe.ContentMD[:idx+1+nextSection])
		} else {
			md.WriteString(recipe.ContentMD)
			md.WriteString("\n")
		}
	}

	md.WriteString("## Zubereitung\n\n")
	for i, step := range steps {
		md.WriteString(fmt.Sprintf("%d. %s\n", i+1, strings.TrimSpace(step)))
	}
	md.WriteString("\n")

	// Preserve notes section if present
	if idx := strings.Index(recipe.ContentMD, "## Notizen"); idx >= 0 {
		md.WriteString(recipe.ContentMD[idx:])
	}

	recipe.ContentMD = md.String()
	return recipe, nil
}

func rephraseInstructions(apiKey string, recipe *Recipe) (*Recipe, error) {
	if apiKey == "" {
		return recipe, nil
	}

	prompt := fmt.Sprintf(`Du bist ein strenger Rezept-Redakteur fuer ein Kochbuch. Deine Aufgabe: Die Schritte STARK kuerzen und vereinheitlichen.

WICHTIGSTE REGEL: Wenn mehrere Schritte das gleiche Muster wiederholen (z.B. Schichten bei Lasagne, wiederholtes Ruehren, mehrfaches Wenden), fasse sie zu EINEM einzigen Schritt zusammen. Beschreibe das Muster, nicht jeden Durchgang.

Beispiel VORHER (schlecht):
1. Tomatensauce auf den Boden streichen
2. Drei Lasagneplatten auflegen
3. Gemuese darauf verteilen
4. Sauce darueber geben
5. Mozzarella streuen
6. Drei Lasagneplatten auflegen
7. Restliches Gemuese verteilen
8. Sauce darueber
9. Mozzarella streuen
10. Letzte Lasagneplatten auflegen

Beispiel NACHHER (gut):
1. In einer Auflaufform abwechselnd schichten: Tomatensauce, Lasagneplatten, Gemuesemischung und Mozzarella (3 Lagen). Mit Lasagneplatten und restlicher Sauce abschliessen.

Weitere Regeln:
- Durchgehend Imperativ ("Die Butter schmelzen", nicht "Man schmilzt die Butter")
- Maximal 8-12 Schritte fuer ein normales Rezept
- Mengen, Temperaturen und Zeiten MUESSEN erhalten bleiben
- Keine Floskeln, kein "nun", "dann", "anschliessend" am Satzanfang
- Gib die Schritte als JSON-Array von Strings zurueck

Rezepttitel: %s

Aktuelle Schritte:
%s

Antworte NUR mit einem JSON-Array, ohne weitere Erklaerung:
["Schritt 1", "Schritt 2", ...]`,
		recipe.Title,
		recipe.ContentMD,
	)

	reqBody, _ := json.Marshal(map[string]any{
		"model": "gpt-4.1-mini",
		"messages": []map[string]string{
			{"role": "user", "content": prompt},
		},
		"temperature": 0.4,
		"max_tokens":  2000,
	})

	req, err := http.NewRequest("POST", "https://api.openai.com/v1/chat/completions", bytes.NewReader(reqBody))
	if err != nil {
		return recipe, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return recipe, fmt.Errorf("OpenAI-Anfrage fehlgeschlagen: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return recipe, err
	}

	if resp.StatusCode != 200 {
		return recipe, fmt.Errorf("OpenAI HTTP %d: %s", resp.StatusCode, string(body))
	}

	var result struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return recipe, err
	}
	if len(result.Choices) == 0 {
		return recipe, fmt.Errorf("keine Antwort von OpenAI")
	}

	content := strings.TrimSpace(result.Choices[0].Message.Content)
	content = strings.TrimPrefix(content, "```json")
	content = strings.TrimPrefix(content, "```")
	content = strings.TrimSuffix(content, "```")
	content = strings.TrimSpace(content)

	var steps []string
	if err := json.Unmarshal([]byte(content), &steps); err != nil {
		return recipe, fmt.Errorf("KI-Antwort konnte nicht geparst werden: %s", content)
	}

	// Rebuild markdown with rephrased steps
	var md strings.Builder

	// Preserve ingredients section
	if idx := strings.Index(recipe.ContentMD, "## Zubereitung"); idx > 0 {
		md.WriteString(recipe.ContentMD[:idx])
	} else if idx := strings.Index(recipe.ContentMD, "## Zutaten"); idx >= 0 {
		afterIngredients := recipe.ContentMD[idx:]
		if nextSection := strings.Index(afterIngredients[1:], "## "); nextSection > 0 {
			md.WriteString(recipe.ContentMD[:idx+1+nextSection])
		} else {
			md.WriteString(recipe.ContentMD)
			md.WriteString("\n")
		}
	}

	md.WriteString("## Zubereitung\n\n")
	for i, step := range steps {
		md.WriteString(fmt.Sprintf("%d. %s\n", i+1, strings.TrimSpace(step)))
	}
	md.WriteString("\n")

	// Preserve notes section if present
	if idx := strings.Index(recipe.ContentMD, "## Notizen"); idx >= 0 {
		md.WriteString(recipe.ContentMD[idx:])
	}

	recipe.ContentMD = md.String()
	return recipe, nil
}

type LLMRecipe struct {
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Servings    string   `json:"servings"`
	PrepTime    string   `json:"prep_time"`
	CookTime    string   `json:"cook_time"`
	Ingredients []string `json:"ingredients"`
	Steps       []string `json:"steps"`
	Category    string   `json:"category"`
	Tags        []string `json:"tags"`
}

func extractRecipeViaLLM(apiKey, text, imageBase64 string, categories []Category, tags []Tag) (*Recipe, error) {
	var catNames []string
	for _, c := range categories {
		catNames = append(catNames, c.Name)
	}
	var tagNames []string
	for _, t := range tags {
		tagNames = append(tagNames, t.Name)
	}

	prompt := fmt.Sprintf(`Extrahiere ein Rezept aus dem folgenden Input. Der Input kann ein Foto, kopierter Text, oder beides sein. Erstelle daraus ein strukturiertes Rezept auf Deutsch.

Zeitangaben: Normalisiere prep_time und cook_time ins Format "X Min." oder "X Std. Y Min." (aber nie "0 Std." — lass die Stunden weg wenn 0). Falls eine Zeit nicht direkt angegeben ist aber aus den Zubereitungsschritten ableitbar (z.B. "30 Minuten kochen"), trage sie ein.

Wichtig zur Kategorie-Unterscheidung:
- "Backen" = Kuchen, Kekse, Plätzchen, Brot, Gebäck, Torten, Muffins (alles was man im Ofen backt als Hauptzweck)
- "Desserts" = Nachspeisen die man NICHT backt: Mousse, Pudding, Eis, Panna Cotta, Tiramisu, Cremes

Verfügbare Kategorien (wähle genau EINE):
%s

Verfügbare Tags (wähle 3-8 passende):
%s

Antworte NUR mit einem JSON-Objekt:
{"title": "Titel", "description": "Kurzbeschreibung", "servings": "4 Portionen", "prep_time": "15 Min.", "cook_time": "30 Min.", "ingredients": ["200g Mehl", "..."], "steps": ["Schritt 1", "Schritt 2"], "category": "Kategoriename", "tags": ["Tag1", "Tag2"]}`,
		strings.Join(catNames, ", "),
		strings.Join(tagNames, ", "),
	)

	if text != "" {
		prompt += "\n\nRezepttext:\n" + text
	}

	// Build messages with optional image
	var messages []any
	if imageBase64 != "" {
		content := []any{
			map[string]string{"type": "text", "text": prompt},
			map[string]any{
				"type": "image_url",
				"image_url": map[string]string{
					"url": imageBase64,
				},
			},
		}
		messages = []any{
			map[string]any{"role": "user", "content": content},
		}
	} else {
		messages = []any{
			map[string]string{"role": "user", "content": prompt},
		}
	}

	model := "gpt-4.1-nano"
	if imageBase64 != "" {
		model = "gpt-4.1-mini"
	}

	reqBody, _ := json.Marshal(map[string]any{
		"model":       model,
		"messages":    messages,
		"temperature": 0.3,
		"max_tokens":  1500,
	})

	req, err := http.NewRequest("POST", "https://api.openai.com/v1/chat/completions", bytes.NewReader(reqBody))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("OpenAI-Anfrage fehlgeschlagen: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("OpenAI HTTP %d: %s", resp.StatusCode, string(body))
	}

	var result struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}
	if len(result.Choices) == 0 {
		return nil, fmt.Errorf("keine Antwort von OpenAI")
	}

	content := strings.TrimSpace(result.Choices[0].Message.Content)
	content = strings.TrimPrefix(content, "```json")
	content = strings.TrimPrefix(content, "```")
	content = strings.TrimSuffix(content, "```")
	content = strings.TrimSpace(content)

	var llmRecipe LLMRecipe
	if err := json.Unmarshal([]byte(content), &llmRecipe); err != nil {
		return nil, fmt.Errorf("KI-Antwort konnte nicht geparst werden: %s", content)
	}

	// Build markdown
	var md strings.Builder
	if len(llmRecipe.Ingredients) > 0 {
		md.WriteString("## Zutaten\n\n")
		for _, ing := range llmRecipe.Ingredients {
			md.WriteString("- " + strings.TrimSpace(ing) + "\n")
		}
		md.WriteString("\n")
	}
	if len(llmRecipe.Steps) > 0 {
		md.WriteString("## Zubereitung\n\n")
		for i, step := range llmRecipe.Steps {
			md.WriteString(fmt.Sprintf("%d. %s\n", i+1, strings.TrimSpace(step)))
		}
		md.WriteString("\n")
	}

	recipe := &Recipe{
		Title:       llmRecipe.Title,
		Description: llmRecipe.Description,
		Servings:    llmRecipe.Servings,
		PrepTime:    llmRecipe.PrepTime,
		CookTime:    llmRecipe.CookTime,
		ContentMD:   md.String(),
	}

	// Match category
	for _, c := range categories {
		if strings.EqualFold(c.Name, llmRecipe.Category) {
			id := c.ID
			recipe.CategoryID = &id
			break
		}
	}

	// Match tags
	for _, sugName := range llmRecipe.Tags {
		for _, t := range tags {
			if strings.EqualFold(t.Name, sugName) {
				recipe.Tags = append(recipe.Tags, t)
				break
			}
		}
	}

	return recipe, nil
}
