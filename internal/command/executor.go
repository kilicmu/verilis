package command

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/user/go-scaffold/pkg/config"
)

const OPEN_ROUTER_REGISTRY = "https://openrouter.ai/api/v1/chat/completions"

// Executor 处理命令行命令执行
type Executor struct {
	config *config.Config
}

// NewExecutor 创建一个新的命令执行器
func NewExecutor(cfg *config.Config) *Executor {
	return &Executor{
		config: cfg,
	}
}

const DEFAULT_CONFIG_NAME = "verilis.config.json"

type VerilisConfig struct {
	AccessToken string `json:"access_token"`
	// Model        string            `json:"model"`
	Output       string            `json:"output"`
	SupportLangs []string          `json:"support_languages"`
	Resource     map[string]string `json:"resource"`
}

func (e *Executor) Init(args []string) {
	configFile := DEFAULT_CONFIG_NAME

	// Check if the file already exists
	if _, err := os.Stat(configFile); err == nil {
		fmt.Printf("\033[33mWarning: %s already exists.\033[0m\n", configFile)
		fmt.Print("Do you want to overwrite it? (y/N): ")

		// Read user input
		reader := bufio.NewReader(os.Stdin)
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(strings.ToLower(input))

		if input != "y" && input != "yes" {
			fmt.Println("Operation cancelled.")
			return
		}
	}

	// Create default config
	config := VerilisConfig{
		Output:       "./i18n/resources",
		SupportLangs: []string{"en", "zh-CN"},
		Resource:     map[string]string{"initial_example": "this is example"},
	}

	// Marshal to JSON with indentation
	configJSON, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		fmt.Printf("\033[31mError: Unable to create config: %v\033[0m\n", err)
		os.Exit(1)
	}

	// Create file if it doesn't exist
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		f, err := os.Create(configFile)
		if err != nil {
			fmt.Printf("\033[31mError: Unable to create config file: %v\033[0m\n", err)
			os.Exit(1)
		}
		f.Close()
	}

	// Write to file
	err = os.WriteFile(configFile, configJSON, 0644)
	if err != nil {
		fmt.Printf("\033[31mError: Unable to write config file: %v\033[0m\n", err)
		os.Exit(1)
	}

	fmt.Printf("\033[32mSuccess: %s has been created.\033[0m\n", configFile)
	fmt.Println("You can now edit this file to configure your i18n settings.")
}

// 20 most widely used languages globally with their ISO codes and English names
var SupportLangs = map[string]string{
	"en":    "English",
	"zh-CN": "Chinese (Simplified)",
	"zh-TW": "Chinese (Traditional)",
	"es":    "Spanish",
	"ar":    "Arabic",
	"hi":    "Hindi",
	"fr":    "French",
	"ru":    "Russian",
	"pt":    "Portuguese",
	"id":    "Indonesian",
	"de":    "German",
	"ja":    "Japanese",
	"bn":    "Bengali",
	"ur":    "Urdu",
	"tr":    "Turkish",
	"it":    "Italian",
	"ko":    "Korean",
	"vi":    "Vietnamese",
	"pl":    "Polish",
	"nl":    "Dutch",
	"th":    "Thai",
}

func (e *Executor) batchGenerateResource(c *VerilisConfig) {
	// Create output directory if it doesn't exist
	err := os.MkdirAll(c.Output, 0755)
	if err != nil {
		fmt.Printf("\033[31mError: Failed to create output directory %s: %v\033[0m\n", c.Output, err)
		os.Exit(1)
	}

	// Report total work to be done
	totalTranslations := len(c.SupportLangs) * len(c.Resource)
	fmt.Printf("Translating %d resources to %d languages (%d total translations)\n",
		len(c.Resource), len(c.SupportLangs), totalTranslations)

	// Initialize a map to store translations for each language
	translations := make(map[string]map[string]string)
	for _, lang := range c.SupportLangs {
		translations[lang] = make(map[string]string)
	}

	// Use a wait group to wait for all translations to complete
	var wg sync.WaitGroup
	// Use a mutex to protect access to the translations map
	var mu sync.Mutex
	// Channel for errors

	fmt.Println("\nStarting concurrent translation of resources...")

	// Process each language concurrently
	for _, lang := range c.SupportLangs {
		wg.Add(1)
		go func(language string) {
			defer wg.Done()

			fmt.Printf("\033[34mStarted processing language: %s (%s)\033[0m\n", SupportLangs[language], language)

			// Convert resource map to JSON
			resourceJSON, err := json.Marshal(c.Resource)
			if err != nil {
				errorMsg := fmt.Sprintf("\033[31mFailed to convert resources to JSON for %s: %v\033[0m\n", language, err)
				fmt.Print(errorMsg)
				return
			}

			// Make API request to OpenRouter with the entire resource JSON
			translatedJSON, err := e.translateText(c.AccessToken, string(resourceJSON), language)
			if err != nil {
				errorMsg := fmt.Sprintf("\033[31mFailed to translate %s: %v\033[0m\n", language, err)
				fmt.Print(errorMsg)
				return
			}

			// Parse the translated JSON with retry mechanism
			translatedResources := make(map[string]string)

			// Define max retries and current attempt
			maxRetries := 3
			for attempt := 1; attempt <= maxRetries; attempt++ {
				err = json.Unmarshal([]byte(translatedJSON), &translatedResources)
				if err == nil {
					// Successfully parsed JSON
					break
				}

				// Failed to parse JSON
				if attempt < maxRetries {
					// Not the last attempt, retry with modified prompt
					retryMsg := fmt.Sprintf("\033[33mAttempt %d/%d: Failed to parse JSON for %s: %v. Retrying...\033[0m\n",
						attempt, maxRetries, language, err)
					fmt.Print(retryMsg)

					// Modify the prompt to emphasize JSON format for retry
					modifiedPrompt := fmt.Sprintf("Fix this JSON to make it valid. Only return the fixed JSON with no explanation, no markdown formatting, no backticks: %s", translatedJSON)

					// Retry the translation with a focus on fixing JSON
					translatedJSON, err = e.translateText(c.AccessToken, modifiedPrompt, "en") // Use English for JSON fixing
					if err != nil {
						errorMsg := fmt.Sprintf("\033[31mFailed to fix JSON in retry attempt %d: %v\033[0m\n", attempt, err)
						fmt.Print(errorMsg)
						continue
					}
				} else {
					// Last attempt failed
					errorMsg := fmt.Sprintf("\033[31mAll %d attempts failed to parse JSON for %s: %v\033[0m\n",
						maxRetries, language, err)
					fmt.Print(errorMsg)
					return
				}
			}

			// Store translations in a thread-safe manner
			mu.Lock()
			translations[language] = translatedResources
			mu.Unlock()

			fmt.Println("\033[32mSuccess\033[0m")

			// Write translation file for this language
			outputFile := filepath.Join(c.Output, language+".json")
			jsonData, err := json.MarshalIndent(translatedResources, "", "  ")
			if err != nil {
				errorMsg := fmt.Sprintf("\033[31mError: Failed to create JSON for %s: %v\033[0m\n", language, err)
				fmt.Print(errorMsg)
				return
			}

			err = os.WriteFile(outputFile, jsonData, 0644)
			if err != nil {
				errorMsg := fmt.Sprintf("\033[31mError: Failed to write file %s: %v\033[0m\n", outputFile, err)
				fmt.Print(errorMsg)
				return
			}

			fmt.Printf("  \033[32mSaved translations to %s\033[0m\n", outputFile)
		}(lang)
	}

	// Wait for all goroutines to complete
	wg.Wait()

	fmt.Println("\n\033[32mTranslation process completed!\033[0m")
}

// translateText calls the OpenRouter API to translate text to the target language
func (e *Executor) translateText(apiKey, text, targetLang string) (string, error) {
	// Prepare the request body
	reqBody := map[string]interface{}{
		"model": "google/gemini-2.5-flash-preview",
		"messages": []map[string]string{
			{
				"role": "user",
				"content": fmt.Sprintf(`
There is a json string, please translate the json-value following text to %s language, sprint. and And it is necessary to maintain the stability of the key. IMPORTANT: Only respond with the translation result json, nothing else (e.g.:markdown syntax).
When translating, if you encounter C-style formatting (such as %%s, %%d, %%.2f, etc.), do not alter the formatting tokens, and make sure to preserve any spaces or punctuation immediately before or after them. This ensures the placeholders work correctly at runtime.
Examples:
Original: Hello, %%s! → Translation: 你好，%%s！
Original: You have %%d new messages. → Translation: 你有 %%d 条新消息。
json: %s`, SupportLangs[targetLang], text),
			},
		},
	}

	// Marshal request body to JSON
	data, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %v", err)
	}

	// Create HTTP request
	req, err := http.NewRequest("POST", OPEN_ROUTER_REGISTRY, bytes.NewBuffer(data))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %v", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("HTTP-Referer", "http://localhost")
	req.Header.Set("X-Title", "Verilis I18N")

	// Make request
	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("request failed: %v", err)
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(bodyBytes))
	}

	// Parse response
	var response struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return "", fmt.Errorf("failed to parse response: %v", err)
	}

	// Get the translated text from the response
	if len(response.Choices) == 0 || response.Choices[0].Message.Content == "" {
		return "", fmt.Errorf("no translation returned")
	}

	return strings.TrimSpace(response.Choices[0].Message.Content), nil
}

func (e *Executor) Generate(args []string) {

	// Config file name
	configFile := DEFAULT_CONFIG_NAME

	// Check if the config file exists
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		fmt.Printf("\033[31mError: %s not found.\033[0m\n", configFile)
		fmt.Println("Run 'verilis init' to create a configuration file first.")
		os.Exit(1)
	}

	// Read the config file
	configData, err := os.ReadFile(configFile)
	if err != nil {
		fmt.Printf("\033[31mError: Failed to read %s: %v\033[0m\n", configFile, err)
		os.Exit(1)
	}

	// Parse the JSON config
	config := &VerilisConfig{}
	err = json.Unmarshal(configData, config)
	if err != nil {
		fmt.Printf("\033[31mError: Failed to parse %s: %v\033[0m\n", configFile, err)
		os.Exit(1)
	}

	fmt.Printf("\033[32mSuccessfully loaded configuration from %s\033[0m\n", configFile)

	// Check that the access token is provided
	if config.AccessToken == "" {
		fmt.Printf("\033[31mError: Access token not provided in %s\033[0m\n", configFile)
		fmt.Println("Please provide an access token from https://openrouter.ai/settings/keys to continue.")
		os.Exit(1)
	}

	// Validate that all configured languages are supported
	unsupportedLangs := []string{}

	// Check each language in config against supported languages
	for _, lang := range config.SupportLangs {
		if _, exists := SupportLangs[lang]; !exists {
			unsupportedLangs = append(unsupportedLangs, lang)
		}
	}

	// If unsupported languages were found, print a warning
	if len(unsupportedLangs) > 0 {
		fmt.Printf("\033[33mError: The following languages are not officially supported: %v\033[0m\n", unsupportedLangs)
		fmt.Println("\nSupported languages:")

		// Print all supported languages with their labels
		for code, label := range SupportLangs {
			fmt.Printf("  %s - %s\n", code, label)
		}

		fmt.Println("\nYou can continue with unsupported languages, but we cannot guarantee full functionality.")
	}
	e.batchGenerateResource(config)
}
