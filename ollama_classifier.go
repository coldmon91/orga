package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

const (
	OllamaURL          = "http://localhost:11434/api/generate"
	ollamaDefaultModel = "gemma3n:e4b"
)

type OllamaRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
	Stream bool   `json:"stream"`
}

type OllamaResponse struct {
	Response string `json:"response"`
}

// ollamaClassifier asks a local Ollama model; a reply outside the categories counts as inconclusive.
type ollamaClassifier struct {
	model string
}

func (c ollamaClassifier) classify(item classifyItem) (string, bool, error) {
	subject, unsureRule := ollamaPromptParts(item)
	category, err := c.requestCategory(item.Kind, subject, unsureRule)
	if err != nil {
		return "", false, err
	}
	return category, isValidCategory(category), nil
}

func ollamaPromptParts(item classifyItem) (subject, unsureRule string) {
	switch {
	case item.Kind == "file":
		return fmt.Sprintf("Filename: %q", item.Name), `If unsure, use "others".`
	case len(item.SampleFiles) == 0:
		return fmt.Sprintf("Directory name: %q", item.Name),
			`If the name alone is not enough to decide, reply "unknown".`
	default:
		return fmt.Sprintf("Directory name: %q\nSample of files inside:\n- %s", item.Name, strings.Join(item.SampleFiles, "\n- ")),
			`Judge by the dominant type of the files inside. If unsure, use "others".`
	}
}

// requestCategory asks Ollama to classify one item described by subject.
func (c ollamaClassifier) requestCategory(kind, subject, unsureRule string) (string, error) {
	prompt := fmt.Sprintf(`You are a file organization assistant.
Classify the following %s into exactly one of these categories:
- documents
- images
- videos
- audio
- archives
- executables
- installers
- others

%s

Rules:
1. Reply ONLY with the category name in lowercase.
2. Do not add any punctuation or explanation.
3. %s

	Category:`, kind, subject, unsureRule)

	reqBody := OllamaRequest{
		Model:  c.model,
		Prompt: prompt,
		Stream: false,
	}
	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", err
	}

	if verbose {
		fmt.Printf("[DEBUG] Requesting classification for %s from Ollama (%s)...\n", subject, OllamaURL)
	}

	resp, err := http.Post(OllamaURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("ollama API returned status: %s", resp.Status)
	}

	var ollamaResp OllamaResponse
	if err := json.NewDecoder(resp.Body).Decode(&ollamaResp); err != nil {
		return "", err
	}

	rawResponse := strings.TrimSpace(ollamaResp.Response)
	if verbose {
		fmt.Printf("[DEBUG] Ollama raw response: %q\n", rawResponse)
	}

	return strings.ToLower(rawResponse), nil
}
