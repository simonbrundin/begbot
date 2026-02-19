package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"regexp"
	"strings"
	"time"
)

type OpenRouterClient struct {
	apiKey   string
	siteURL  string
	siteName string
	baseURL  string
}

type openRouterRequest struct {
	Model       string              `json:"model"`
	Messages    []openRouterMessage `json:"messages"`
	MaxTokens   int                 `json:"max_tokens,omitempty"`
	Temperature float64             `json:"temperature,omitempty"`
}

type openRouterMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type openRouterResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error"`
}

func NewOpenRouterClient(apiKey, siteURL, siteName string) *OpenRouterClient {
	return &OpenRouterClient{
		apiKey:   apiKey,
		siteURL:  siteURL,
		siteName: siteName,
		baseURL:  "https://openrouter.ai/api/v1/chat/completions",
	}
}

func (c *OpenRouterClient) Chat(ctx context.Context, model, prompt string) (string, error) {
	reqBody := openRouterRequest{
		Model: model,
		Messages: []openRouterMessage{
			{Role: "user", Content: prompt},
		},
		MaxTokens:   200,
		Temperature: 0.1,
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL, bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	log.Printf("DEBUG OpenRouter: Using API key: %s...", c.apiKey[:20])
	log.Printf("DEBUG OpenRouter: SiteURL: %s, SiteName: %s, Model: %s", c.siteURL, c.siteName, model)
	req.Header.Set("HTTP-Referer", c.siteURL)
	req.Header.Set("X-Title", c.siteName)
	req.Header.Set("User-Agent", "Begbot/1.0")

	client := &http.Client{
		Timeout: 60 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		if len(respBody) > 500 {
			respBody = respBody[:500]
		}
		log.Printf("OpenRouter error: %s", string(respBody))
		return "", fmt.Errorf("API request failed with status %d", resp.StatusCode)
	}

	respBody, _ := io.ReadAll(resp.Body)

	var result openRouterResponse
	if err := json.NewDecoder(bytes.NewReader(respBody)).Decode(&result); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	if result.Error != nil {
		return "", fmt.Errorf("OpenRouter error: %s", result.Error.Message)
	}

	if len(result.Choices) == 0 {
		return "", fmt.Errorf("no response from OpenRouter")
	}

	return result.Choices[0].Message.Content, nil
}

func (c *OpenRouterClient) GetModel(functionName string, defaultModel string, models map[string]string) string {
	if model, ok := models[functionName]; ok && model != "" {
		return model
	}
	return defaultModel
}

func cleanupMarkdownJSON(content string) string {
	content = strings.TrimSpace(content)
	re := regexp.MustCompile("^```json\\s*")
	content = re.ReplaceAllString(content, "")
	re = regexp.MustCompile("\\s*```$")
	content = re.ReplaceAllString(content, "")
	return strings.TrimSpace(content)
}
