package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"maps"
	"net/http"
	"os"
	"strings"
	"time"
)

const (
	jevURL          = "https://api.typesafe.ai/v1/systemone"
	jevDefaultModel = "jev-latest"
	jevAPIKeyEnv    = "TYPESAFE_API_KEY"
	jevQuestionID   = "category"

	// Below this confidence a directory name alone is treated as inconclusive
	jevMinConfidence = 0.7

	jevRequestTimeout   = 30 * time.Second
	jevMaxAttempts      = 3
	jevMaxResponseBytes = 1 << 20
	jevMaxErrorBodyLen  = 300
)

const jevInstructions = "Which category should this item be filed under? " +
	"`kind` says whether it is a file or a directory. " +
	"Judge a file by its name and extension. " +
	"Judge a directory by its name, and when `sample_files` is present, by the dominant type of those files."

var jevCategoryCriteria = map[string]string{
	"documents":   "Readable documents such as text, PDF, office files, spreadsheets, presentations, e-books",
	"images":      "Photos, pictures, screenshots, and other graphics",
	"videos":      "Video and movie files",
	"audio":       "Music, voice recordings, and other sound files",
	"archives":    "Compressed or bundled archives such as zip, tar, gz, 7z, rar",
	"executables": "Programs or scripts that run directly, such as binaries, .exe, .app, shell scripts",
	"installers":  "Packages that install software, such as .dmg, .pkg, .msi, setup programs",
	"others":      "Anything that fits none of the categories above",
}

// Offered only for a directory name alone; without it Jev confidently picks "others"
// for uninformative names and the contents are never inspected.
const (
	jevUnknownOption      = "unknown"
	jevUnknownDescription = "The directory name alone does not reveal what kind of files it contains"
)

type jevQuestion struct {
	Type         string            `json:"type"`
	Instructions string            `json:"instructions"`
	Criteria     map[string]string `json:"criteria"`
}

type jevRequest struct {
	Model     string                 `json:"model"`
	State     classifyItem           `json:"state"`
	Questions map[string]jevQuestion `json:"questions"`
}

type jevChoiceAnswer struct {
	Choice     string  `json:"choice"`
	Confidence float64 `json:"confidence"`
}

type jevResponse struct {
	Answers map[string]jevChoiceAnswer `json:"answers"`
}

// jevClassifier asks TypeSafe Jev with a Choice question; low confidence counts as inconclusive.
type jevClassifier struct {
	model  string
	apiKey string
	client *http.Client
}

func newJevClassifier(model string) (*jevClassifier, error) {
	apiKey := os.Getenv(jevAPIKeyEnv)
	if apiKey == "" {
		return nil, fmt.Errorf("%s is not set (environment or .env)", jevAPIKeyEnv)
	}
	return &jevClassifier{
		model:  model,
		apiKey: apiKey,
		client: &http.Client{Timeout: jevRequestTimeout},
	}, nil
}

func (c *jevClassifier) classify(item classifyItem) (string, bool, error) {
	criteria := jevCategoryCriteria
	if item.Kind == "directory" && len(item.SampleFiles) == 0 {
		criteria = maps.Clone(jevCategoryCriteria)
		criteria[jevUnknownOption] = jevUnknownDescription
	}

	body, err := json.Marshal(jevRequest{
		Model: c.model,
		State: item,
		Questions: map[string]jevQuestion{
			jevQuestionID: {Type: "choice", Instructions: jevInstructions, Criteria: criteria},
		},
	})
	if err != nil {
		return "", false, fmt.Errorf("encode jev request: %w", err)
	}

	if verbose {
		fmt.Printf("[DEBUG] Requesting classification for %s %q (samples: %d) from Jev (%s)...\n",
			item.Kind, item.Name, len(item.SampleFiles), c.model)
	}

	respBody, err := c.post(body)
	if err != nil {
		return "", false, err
	}

	var parsed jevResponse
	if err := json.Unmarshal(respBody, &parsed); err != nil {
		return "", false, fmt.Errorf("decode jev response: %w", err)
	}
	answer, ok := parsed.Answers[jevQuestionID]
	if !ok {
		return "", false, fmt.Errorf("jev response has no %q answer", jevQuestionID)
	}

	if verbose {
		fmt.Printf("[DEBUG] Jev answer: %q (confidence %.2f)\n", answer.Choice, answer.Confidence)
	}
	conclusive := answer.Choice != jevUnknownOption && answer.Confidence >= jevMinConfidence
	return answer.Choice, conclusive, nil
}

// post sends body to Jev, retrying with exponential backoff on rate limit (429) and overload (529).
func (c *jevClassifier) post(body []byte) ([]byte, error) {
	for attempt := 1; ; attempt++ {
		req, err := http.NewRequest(http.MethodPost, jevURL, bytes.NewReader(body))
		if err != nil {
			return nil, fmt.Errorf("build jev request: %w", err)
		}
		req.Header.Set("Authorization", "Bearer "+c.apiKey)
		req.Header.Set("Content-Type", "application/json")

		resp, err := c.client.Do(req)
		if err != nil {
			return nil, fmt.Errorf("request jev: %w", err)
		}
		respBody, err := io.ReadAll(io.LimitReader(resp.Body, jevMaxResponseBytes))
		resp.Body.Close()
		if err != nil {
			return nil, fmt.Errorf("read jev response: %w", err)
		}

		if resp.StatusCode == http.StatusOK {
			return respBody, nil
		}
		retryable := resp.StatusCode == http.StatusTooManyRequests || resp.StatusCode == 529
		if !retryable || attempt == jevMaxAttempts {
			return nil, fmt.Errorf("jev API returned status %s: %s", resp.Status, truncate(strings.TrimSpace(string(respBody)), jevMaxErrorBodyLen))
		}

		backoff := time.Duration(1<<(attempt-1)) * time.Second
		if verbose {
			fmt.Printf("[DEBUG] Jev returned %s; retrying in %s\n", resp.Status, backoff)
		}
		time.Sleep(backoff)
	}
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
