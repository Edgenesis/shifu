package gpt

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/edgenesis/shifu/pkg/logger"
	"github.com/edgenesis/shifu/tools/release/prompts"
	"github.com/openai/openai-go"
	"github.com/openai/openai-go/azure"
)

const (
	// Azure OpenAI API configuration
	DefaultAPIVersion = "2024-06-01"
	DefaultTimeout    = 60 * time.Second
	
	// File configuration
	DefaultChangelogDir = "CHANGELOG"
	FilePermissions     = 0644
	
	// Content processing
	LanguageSeparator = "--------"
	CodeBlockMarker   = "```"
)

// Config holds all configuration for the changelog generator
type Config struct {
	APIKey         string
	Host           string
	DeploymentName string
	Version        string
	APIVersion     string
	Timeout        time.Duration
	ChangelogDir   string
}

// NewConfigFromEnv creates a new config from environment variables
func NewConfigFromEnv() (*Config, error) {
	config := &Config{
		APIKey:         os.Getenv("AZURE_OPENAI_APIKEY"),
		Host:           os.Getenv("AZURE_OPENAI_HOST"),
		DeploymentName: os.Getenv("DEPLOYMENT_NAME"),
		Version:        os.Getenv("VERSION"),
		APIVersion:     DefaultAPIVersion,
		Timeout:        DefaultTimeout,
		ChangelogDir:   DefaultChangelogDir,
	}

	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("configuration validation failed: %w", err)
	}

	return config, nil
}

// Validate checks if all required configuration is present
func (c *Config) Validate() error {
	if c.APIKey == "" {
		return fmt.Errorf("AZURE_OPENAI_APIKEY environment variable is required")
	}
	if c.Host == "" {
		return fmt.Errorf("AZURE_OPENAI_HOST environment variable is required")
	}
	if c.DeploymentName == "" {
		return fmt.Errorf("DEPLOYMENT_NAME environment variable is required")
	}
	if c.Version == "" {
		return fmt.Errorf("VERSION environment variable is required")
	}
	
	// Validate host URL format
	if !strings.HasPrefix(c.Host, "https://") {
		return fmt.Errorf("host must be a valid HTTPS URL, got: %s", c.Host)
	}
	
	return nil
}

// ChangelogGenerator handles the generation of changelogs using OpenAI
type ChangelogGenerator struct {
	config *Config
	client *openai.Client
}

// NewChangelogGenerator creates a new changelog generator
func NewChangelogGenerator(config *Config) (*ChangelogGenerator, error) {
	client := openai.NewClient(
		azure.WithEndpoint(config.Host, config.APIVersion),
		azure.WithAPIKey(config.APIKey),
	)

	return &ChangelogGenerator{
		config: config,
		client: &client,
	}, nil
}

// GenerateChangelog generates changelog files from release notes
func (cg *ChangelogGenerator) GenerateChangelog(releaseNotes string) error {
	if strings.TrimSpace(releaseNotes) == "" {
		return fmt.Errorf("release notes cannot be empty")
	}

	logger.Infof("Starting changelog generation for version %s", cg.config.Version)

	// Generate messages for the AI
	messages := cg.buildMessages(releaseNotes)

	// Call OpenAI API
	content, err := cg.callOpenAI(messages)
	if err != nil {
		return fmt.Errorf("failed to call OpenAI API: %w", err)
	}

	// Process and save the generated content
	if err := cg.processAndSaveContent(content); err != nil {
		return fmt.Errorf("failed to process and save content: %w", err)
	}

	logger.Infof("Successfully generated changelogs for version %s", cg.config.Version)
	return nil
}

// buildMessages constructs the message chain for OpenAI
func (cg *ChangelogGenerator) buildMessages(releaseNotes string) []openai.ChatCompletionMessageParamUnion {
	return []openai.ChatCompletionMessageParamUnion{
		openai.UserMessage(prompts.GreetingPrompts),
		openai.UserMessage(prompts.TemplateENPrompts),
		openai.UserMessage(prompts.TemplateZHPrompts),
		openai.UserMessage(prompts.GeneratePrompts),
		openai.UserMessage(releaseNotes),
	}
}

// callOpenAI makes the API call to OpenAI and returns the response content
func (cg *ChangelogGenerator) callOpenAI(messages []openai.ChatCompletionMessageParamUnion) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), cg.config.Timeout)
	defer cancel()

	resp, err := cg.client.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
		Messages:    messages,
		Model:       openai.ChatModel(cg.config.DeploymentName),
		Temperature: openai.Float(0.0),
	})
	if err != nil {
		return "", fmt.Errorf("API call failed: %w", err)
	}

	if len(resp.Choices) == 0 {
		return "", fmt.Errorf("no response choices returned from API")
	}

	content := resp.Choices[0].Message.Content
	if content == "" {
		return "", fmt.Errorf("empty content returned from API")
	}

	return content, nil
}

// processAndSaveContent processes the AI response and saves changelog files
func (cg *ChangelogGenerator) processAndSaveContent(content string) error {
	// Split content into English and Chinese versions
	parts := strings.Split(content, LanguageSeparator)
	if len(parts) < 2 {
		return fmt.Errorf("invalid content format: expected separator '%s' to split English and Chinese versions", LanguageSeparator)
	}

	// Process content for both languages
	enContent := cg.processContent(parts[0])
	zhContent := cg.processContent(parts[1])

	// Ensure changelog directory exists
	if err := cg.ensureChangelogDir(); err != nil {
		return fmt.Errorf("failed to create changelog directory: %w", err)
	}

	// Save English changelog
	enPath := filepath.Join(cg.config.ChangelogDir, fmt.Sprintf("CHANGELOG-%s.md", cg.config.Version))
	if err := cg.saveFile(enPath, enContent); err != nil {
		return fmt.Errorf("failed to save English changelog: %w", err)
	}

	// Save Chinese changelog
	zhPath := filepath.Join(cg.config.ChangelogDir, fmt.Sprintf("CHANGELOG-%s-zh.md", cg.config.Version))
	if err := cg.saveFile(zhPath, zhContent); err != nil {
		return fmt.Errorf("failed to save Chinese changelog: %w", err)
	}

	logger.Infof("Saved changelogs: %s, %s", enPath, zhPath)
	return nil
}

// processContent cleans and formats the content
func (cg *ChangelogGenerator) processContent(content string) string {
	// Remove code block markers
	content = strings.ReplaceAll(content, CodeBlockMarker, "")
	
	// Split into lines and remove empty lines
	lines := strings.Split(content, "\n")
	var processedLines []string

	for _, line := range lines {
		if trimmed := strings.TrimSpace(line); trimmed != "" {
			processedLines = append(processedLines, trimmed)
		}
	}

	// Join with double newlines for proper markdown spacing
	return strings.Join(processedLines, "\n\n")
}

// ensureChangelogDir creates the changelog directory if it doesn't exist
func (cg *ChangelogGenerator) ensureChangelogDir() error {
	return os.MkdirAll(cg.config.ChangelogDir, 0755)
}

// saveFile saves content to a file with proper error handling
func (cg *ChangelogGenerator) saveFile(path, content string) error {
	if err := os.WriteFile(path, []byte(content), FilePermissions); err != nil {
		return fmt.Errorf("failed to write file %s: %w", path, err)
	}
	return nil
}

// Start is the main entry point that maintains backward compatibility
func Start(releaseNoteResp string) error {
	config, err := NewConfigFromEnv()
	if err != nil {
		return fmt.Errorf("failed to initialize configuration: %w", err)
	}

	generator, err := NewChangelogGenerator(config)
	if err != nil {
		return fmt.Errorf("failed to create changelog generator: %w", err)
	}

	return generator.GenerateChangelog(releaseNoteResp)
}