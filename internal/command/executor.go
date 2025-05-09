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
)

const OPEN_ROUTER_REGISTRY = "https://openrouter.ai/api/v1/chat/completions"

// Executor 处理命令行命令执行
type Executor struct {
}

// NewExecutor 创建一个新的命令执行器
func NewExecutor() *Executor {
	return &Executor{}
}

const DEFAULT_CONFIG_NAME = "verilis.config.json"

type VerilisConfig struct {
	AccessToken string `json:"access_token"`
	// Model        string            `json:"model"`
	Output       string            `json:"output"`
	SupportLangs []string          `json:"support_languages"`
	Resource     map[string]string `json:"resource"`
}

func (e *Executor) Init() {
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
	err := os.MkdirAll(c.Output, 0755)
	if err != nil {
		fmt.Printf("\033[31mError: Failed to create output directory %s: %v\033[0m\n", c.Output, err)
		os.Exit(1)
	}

	totalTranslations := len(c.SupportLangs) * len(c.Resource)
	fmt.Printf("Translating %d resources to %d languages (%d total translations)\n",
		len(c.Resource), len(c.SupportLangs), totalTranslations)

	// Initialize a map to store translations for each language
	translations := make(map[string]map[string]string)
	for _, lang := range c.SupportLangs {
		// Check if language file exists in Resource directory
		langFile := filepath.Join(c.Output, lang+".json")
		if fileData, err := os.ReadFile(langFile); err == nil {
			// File exists, parse it
			var existingTranslations map[string]string

			if err := json.Unmarshal(fileData, &existingTranslations); err == nil {
				// remove unexisted resource fileds
				for k := range existingTranslations {
					if _, ok := c.Resource[k]; !ok {
						delete(existingTranslations, k)
					}
				}
				translations[lang] = existingTranslations
				fmt.Printf("Loaded existing translations for %s from %s\n", lang, langFile)
				continue
			} else {
				fmt.Printf("\033[33mWarning: Failed to parse existing translations for %s: %v\033[0m\n", lang, err)
				panic("unsupport translation resource:" + lang + ".json")
			}
		}
		// If file doesn't exist or couldn't be parsed, initialize as empty
		translations[lang] = make(map[string]string)
	}

	var wg sync.WaitGroup
	var mu sync.RWMutex

	fmt.Println("\nStarting concurrent translation of resources...")

	// Process each language concurrently
	for _, lang := range c.SupportLangs {
		wg.Add(1)
		go func(language string) {
			defer wg.Done()

			fmt.Printf("\033[34mStarted processing language: %s (%s)\033[0m\n", SupportLangs[language], language)

			untranslateFields := make(map[string]string)

			mu.RLock()
			translatedFields := translations[language]
			mu.RUnlock()

			for k, v := range c.Resource {
				if _, ok := translatedFields[k]; !ok {
					untranslateFields[k] = v
				}
			}

			if len(untranslateFields) != 0 {

				// Convert resource map to JSON
				untranslateFieldsJSON, err := json.Marshal(untranslateFields)
				if err != nil {
					errorMsg := fmt.Sprintf("\033[31mFailed to convert resources to JSON for %s: %v\033[0m\n", language, err)
					fmt.Println(errorMsg)
					return
				}

				// Make API request to OpenRouter with the entire resource JSON
				translatedJSON, err := e.translateText(c.AccessToken, string(untranslateFieldsJSON), language)
				if err != nil {
					errorMsg := fmt.Sprintf("\033[31mFailed to translate %s: %v\033[0m\n", language, err)
					fmt.Println(errorMsg)
					return
				}

				// Parse the translated JSON with retry mechanism
				modelTranslatedResources := make(map[string]string)

				// Define max retries and current attempt
				maxRetries := 3
				for attempt := 1; attempt <= maxRetries; attempt++ {
					err = json.Unmarshal([]byte(translatedJSON), &modelTranslatedResources)
					if err == nil {
						break
					}

					// Failed to parse JSON
					if attempt < maxRetries {
						// Not the last attempt, retry with modified prompt
						retryMsg := fmt.Sprintf("\033[33mAttempt %d/%d: Failed to parse JSON for %s: %v. Retrying...\033[0m\n",
							attempt, maxRetries, language, err)
						fmt.Println(retryMsg)

						// Modify the prompt to emphasize JSON format for retry
						modifiedPrompt := fmt.Sprintf("Fix this JSON to make it valid. Only return the fixed JSON with no explanation, no markdown formatting, no backticks: %s", translatedJSON)

						// Retry the translation with a focus on fixing JSON
						translatedJSON, err = e.translateText(c.AccessToken, modifiedPrompt, language)
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

				for k, v := range modelTranslatedResources {
					fmt.Printf("translate to「%s」, resource key「%s」: 【‘%s’ => ‘%s’】 \n", language, k, v, c.Resource[k])
					translatedFields[k] = v
				}

				mu.Lock()
				translations[language] = modelTranslatedResources
				mu.Unlock()

				fmt.Println("\033[32mSuccess\033[0m")

			} else {
				fmt.Printf("language %s has non resource to translate \n", language)
			}

			outputFile := filepath.Join(c.Output, language+".json")
			jsonData, err := json.MarshalIndent(translatedFields, "", "  ")
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

	wg.Wait()

	fmt.Println("\n\033[32mTranslation process completed!\033[0m")
}

func (e *Executor) translateText(apiKey, text, targetLang string) (string, error) {
	reqBody := map[string]interface{}{
		"model": "openai/chatgpt-4o-latest",
		"messages": []map[string]string{
			{
				"role": "user",
				"content": fmt.Sprintf(`
There is a json string, please translate the json-value following text to %s language, sprint. and And it is necessary to maintain the stability of the key. IMPORTANT: Only respond with the translation result json, nothing else (e.g.:markdown syntax, don't add unexist key value in result json).
When translating, if you encounter C-style formatting (such as %%s, %%d, %%.2f, etc.), do not alter the formatting tokens, and make sure to preserve any spaces or punctuation immediately before or after them. This ensures the placeholders work correctly at runtime.
Examples:
Original: Hello, %%s! → Translation: 你好，%%s！
Original: You have %%d new messages. → Translation: 你有 %%d 条新消息。
json: \n\n%s`, SupportLangs[targetLang], text),
			},
		},
	}

	data, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %v", err)
	}
	req, err := http.NewRequest("POST", OPEN_ROUTER_REGISTRY, bytes.NewBuffer(data))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("HTTP-Referer", "http://localhost")
	req.Header.Set("X-Title", "Verilis I18N")

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

	if len(response.Choices) == 0 || response.Choices[0].Message.Content == "" {
		return "", fmt.Errorf("no translation returned")
	}

	return strings.TrimSpace(response.Choices[0].Message.Content), nil
}

func (e *Executor) Generate() {

	configFile := DEFAULT_CONFIG_NAME

	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		fmt.Printf("\033[31mError: %s not found.\033[0m\n", configFile)
		fmt.Println("Run 'verilis init' to create a configuration file first.")
		os.Exit(1)
	}

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

	unsupportedLangs := []string{}

	for _, lang := range config.SupportLangs {
		if _, exists := SupportLangs[lang]; !exists {
			unsupportedLangs = append(unsupportedLangs, lang)
		}
	}

	if len(unsupportedLangs) > 0 {
		fmt.Printf("\033[33mError: The following languages are not officially supported: %v\033[0m\n", unsupportedLangs)
		fmt.Println("\nSupported languages:")

		for code, label := range SupportLangs {
			fmt.Printf("  %s - %s\n", code, label)
		}

		fmt.Println("\nYou can continue with unsupported languages, but we cannot guarantee full functionality.")
	}
	e.batchGenerateResource(config)
}
