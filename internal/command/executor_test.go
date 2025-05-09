package command

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
	"testing"

	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"
)

const configfilename = "verilis.config.json"

// TestInit tests the Init function
func TestInit(t *testing.T) {
	exec := NewExecutor()
	assert.NotNil(t, exec)

	exec.Init()

	assert.FileExistsf(t, configfilename, "config file must exist after init")
	os.Remove(configfilename)

	assert.NoFileExists(t, configfilename)
}

func TestGenerate(t *testing.T) {
	godotenv.Load("../../.env")
	exec := NewExecutor()
	assert.NotNil(t, exec)

	exec.Init()
	assert.FileExistsf(t, configfilename, "config file must exist after init")

	defer func() {
		os.Remove(configfilename)
	}()

	configJSON, err := os.ReadFile(configfilename)
	assert.NoError(t, err)
	conf := &VerilisConfig{}
	json.Unmarshal(configJSON, conf)
	conf.AccessToken = os.Getenv("TEST_TOKEN")
	conf.Resource["hello"] = "hello"
	conf.Resource["whats_up"] = "whats up"

	conf.Output = "tmp_resources"

	defer os.RemoveAll(conf.Output)

	configJSON, err = json.Marshal(conf)
	assert.NoError(t, err)
	err = os.WriteFile(configfilename, configJSON, 0x777)
	assert.NoError(t, err)

	// test translate
	exec.Generate()

	translatedresourceMap := make(map[string]map[string]string)

	for _, lang := range conf.SupportLangs {
		fullpath := path.Join(conf.Output, lang+".json")
		assert.FileExistsf(t, fullpath, "translate result file is must exist after translate")

		translatedresourceJson, err := os.ReadFile(fullpath)
		assert.NoError(t, err)

		_translateresource := make(map[string]string)
		assert.NoError(t, json.Unmarshal(translatedresourceJson, &_translateresource))

		for k := range conf.Resource {
			_, ok := _translateresource[k]
			assert.Truef(t, ok, "translated field must exist after translate")
		}

		translatedresourceMap[lang] = _translateresource
	}

	// Test incremental translation
	conf.Resource["incremental_new_test"] = "new test"
	conf.Resource["c_sprint_test"] = "this number is %d"
	configJSON, err = json.Marshal(conf)
	assert.NoError(t, err)
	err = os.WriteFile(configfilename, configJSON, 0x777)
	assert.NoError(t, err)

	exec.Generate()
	for _, lang := range conf.SupportLangs {
		fullpath := path.Join(conf.Output, lang+".json")
		assert.FileExistsf(t, fullpath, "translate result file is must exist after translate")

		aftertranslatedresourceJson, err := os.ReadFile(fullpath)
		assert.NoError(t, err)
		aftertranslatedresource := make(map[string]string)

		assert.NoError(t, json.Unmarshal(aftertranslatedresourceJson, &aftertranslatedresource))

		for k, v := range translatedresourceMap[lang] {
			// history keep no change
			assert.Equal(t, aftertranslatedresource[k], v)
		}
		fmt.Printf("%+v\n", aftertranslatedresource)
		for k := range conf.Resource {
			_, ok := aftertranslatedresource[k]
			assert.Truef(t, ok, "translated field must exist after translate")
		}
	}

}

// TestLoadConfig tests loading a configuration file
// func TestLoadConfig(t *testing.T) {
// 	// Setup test directory
// 	testDir := t.TempDir()
// 	originalWd, err := os.Getwd()
// 	if err != nil {
// 		t.Fatalf("Failed to get current directory: %v", err)
// 	}

// 	// Change to test directory
// 	err = os.Chdir(testDir)
// 	if err != nil {
// 		t.Fatalf("Failed to change to test directory: %v", err)
// 	}
// 	defer func() {
// 		// Cleanup: change back to original directory
// 		os.Chdir(originalWd)
// 	}()

// 	// Create a test config file
// 	config := VerilisConfig{
// 		AccessToken:  os.Getenv("TEST_TOKEN"),
// 		Output:       "./output",
// 		SupportLangs: []string{"en", "zh-CN"},
// 		Resource:     map[string]string{"hello": "Hello", "world": "World"},
// 	}

// 	configData, err := json.MarshalIndent(config, "", "  ")
// 	if err != nil {
// 		t.Fatalf("Failed to marshal config: %v", err)
// 	}

// 	err = os.WriteFile(DEFAULT_CONFIG_NAME, configData, 0644)
// 	if err != nil {
// 		t.Fatalf("Failed to write config file: %v", err)
// 	}

// 	// Test loading the config
// 	// Note: We can't directly test this since the Generate function calls os.Exit
// 	// In a real test, you would refactor the code to return errors instead
// 	// of calling os.Exit
// }

// // TestReadExistingTranslations tests reading existing translation files
// func TestReadExistingTranslations(t *testing.T) {
// 	// Setup test directory
// 	testDir := t.TempDir()

// 	// Create output directory
// 	outputDir := filepath.Join(testDir, "output")
// 	err := os.MkdirAll(outputDir, 0755)
// 	if err != nil {
// 		t.Fatalf("Failed to create output directory: %v", err)
// 	}

// 	// Create a test translation file
// 	translations := map[string]string{
// 		"hello": "Hello",
// 		"world": "World",
// 	}

// 	translationData, err := json.MarshalIndent(translations, "", "  ")
// 	if err != nil {
// 		t.Fatalf("Failed to marshal translations: %v", err)
// 	}

// 	translationFile := filepath.Join(outputDir, "en.json")
// 	err = os.WriteFile(translationFile, translationData, 0644)
// 	if err != nil {
// 		t.Fatalf("Failed to write translation file: %v", err)
// 	}

// 	// In a real test, we would create a config like this and pass it to batchGenerateResource
// 	// But since we can't call that function directly in tests (due to API calls and panic),
// 	// we'll just document what the config would look like
// 	_ = &VerilisConfig{
// 		AccessToken:  "test-token",
// 		Output:       outputDir,
// 		SupportLangs: []string{"en", "fr"},
// 		Resource:     map[string]string{"hello": "Hello", "world": "World", "new": "New"},
// 	}

// 	// We can't directly test batchGenerateResource as it makes API calls
// 	// and has a panic statement for invalid translations
// 	// In a real test, you would mock the API calls and handle the panic

// 	// Instead, we'll verify the translation file exists
// 	if _, err := os.Stat(translationFile); os.IsNotExist(err) {
// 		t.Fatalf("Translation file should exist: %v", err)
// 	}

// 	// Read the translation file
// 	fileData, err := os.ReadFile(translationFile)
// 	if err != nil {
// 		t.Fatalf("Failed to read translation file: %v", err)
// 	}

// 	// Parse the translation file
// 	var existingTranslations map[string]string
// 	err = json.Unmarshal(fileData, &existingTranslations)
// 	if err != nil {
// 		t.Fatalf("Failed to parse translation file: %v", err)
// 	}

// 	// Verify the translations
// 	if existingTranslations["hello"] != "Hello" {
// 		t.Errorf("Expected 'hello' to be 'Hello', got '%s'", existingTranslations["hello"])
// 	}

// 	if existingTranslations["world"] != "World" {
// 		t.Errorf("Expected 'world' to be 'World', got '%s'", existingTranslations["world"])
// 	}
// }
