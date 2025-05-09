# Verilis

<p align="center">
  <img src="logo/logo.png" alt="Verilis Logo" width="200" height="200">
</p>

<p align="center">
  <strong>Verilis</strong> - An AI-driven internationalization (i18n) solution
</p>

<p align="center">
  <a href="#features">Features</a> •
  <a href="#installation">Installation</a> •
  <a href="#getting-started">Getting Started</a> •
  <a href="#commands">Commands</a> •
  <a href="#configuration">Configuration</a> •
  <a href="#license">License</a>
</p>

---

## Features

- 🌍 **Multi-language Support**: Translate your content into multiple languages
- 🤖 **AI-powered**: Leverage OpenRouter API for high-quality translations
- ⚡ **Concurrent Processing**: Process multiple language translations in parallel using goroutines
- 🔄 **Format Preservation**: Maintain C-style formatting tokens (%s, %d, etc.) during translation
- 📦 **Cross-platform**: Support for Linux, macOS, and Windows
- 🛠️ **Simple Configuration**: Easy JSON-based configuration file

## Installation

### Using the Automatic Installation Script

The easiest way to install Verilis is using our installation script, which automatically detects your platform and installs the appropriate binary:

```bash
# Using curl
curl -sSL https://raw.githubusercontent.com/kilicmu/verilis/main/verilis-install.sh | bash

# Using wget
wget -qO- https://raw.githubusercontent.com/kilicmu/verilis/main/verilis-install.sh | bash
```

### Manual Installation

1. Download the appropriate binary for your platform from [GitHub Releases](https://github.com/user/verilis/releases)
2. Extract the downloaded archive
3. Move the binary to a directory in your PATH

### Building from Source

```bash
# Clone the repository
git clone https://github.com/user/verilis.git
cd verilis

# Build the application
make build

# Install it
make install
```

## Getting Started

### Initialize Your Project

To start using Verilis, first initialize your project:

```bash
verilis init
```

This command creates a `verilis.config.json` file in your current directory with default settings.

### Edit the Configuration File

Edit the generated configuration file to add your OpenRouter API key and translation resources:

```json
{
  "access_token": "your-openrouter-api-key",
  "output": "./i18n/resources",
  "support_languages": ["en", "zh-CN", "fr", "es"],
  "resource": {
    "hello": "Hello",
    "welcome": "Welcome to Verilis",
    "greeting": "Hello, %s!",
    "items_count": "You have %d items"
  }
}
```

### Generate Translations

Once your configuration is set up, generate translations with:

```bash
verilis generate
```

This command will use the OpenRouter API to translate your resources into all supported languages and save them in the specified output directory.

## Commands

### init

```bash
verilis init
```

Initializes a new Verilis project by creating a configuration file.

### generate

```bash
verilis generate
```

Generates translations based on the configuration file.

## Configuration

The `verilis.config.json` file contains the following fields:

| Field | Description |
|-------|-------------|
| `access_token` | Your OpenRouter API access token, required for translation services |
| `output` | Directory where translation files will be saved |
| `support_languages` | Array of language codes to translate into (e.g., "en", "zh-CN", "fr") |
| `resource` | Key-value pairs of strings to translate |

### Access Token

To obtain an OpenRouter API access token, visit [OpenRouter](https://openrouter.ai/settings/keys) and create an API key.

### Supported Languages

Verilis supports a wide range of language codes, including but not limited to:

- `en` - English
- `zh-CN` - Chinese (Simplified)
- `zh-TW` - Chinese (Traditional)
- `fr` - French
- `es` - Spanish
- `de` - German
- `ja` - Japanese
- `ko` - Korean
- `ru` - Russian
- `pt` - Portuguese
- `it` - Italian

### Resource Format

The `resource` field contains key-value pairs where:
- Keys are unique identifiers for your strings
- Values are the source text to be translated

C-style formatting tokens (%s, %d, etc.) are preserved during translation.

## License

Distributed under the MIT License. See `LICENSE` for more information.

---