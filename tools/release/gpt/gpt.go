package gpt

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
	"github.com/edgenesis/shifu/pkg/logger"
	"github.com/edgenesis/shifu/tools/release/prompts"
)

var (
	API_KEY = os.Getenv("OPENAI_API_KEY")
	MODEL   = getModel()
	VERSION = os.Getenv("VERSION")
)

func getModel() string {
	model := os.Getenv("OPENAI_MODEL")
	if model == "" {
		return "gpt-4" // default model
	}
	return model
}

type Helper struct {
	client   *openai.Client
	messages []openai.ChatCompletionMessageParamUnion
}

func Start(releaseNoteResp string) error {
	helper := &Helper{}
	client, err := newGPT()
	if err != nil {
		return err
	}

	helper.client = client

	helper.generateMessages(releaseNoteResp)

	err = helper.generateChangelog()
	if err != nil {
		return err
	}

	return nil
}

func newGPT() (*openai.Client, error) {
	if API_KEY == "" {
		return nil, fmt.Errorf("OPENAI_API_KEY environment variable is required")
	}

	client := openai.NewClient(
		option.WithAPIKey(API_KEY),
	)
	return &client, nil
}

func (h *Helper) generateMessages(releaseNoteResp string) {
	h.messages = []openai.ChatCompletionMessageParamUnion{
		openai.UserMessage(prompts.GreetingPrompts),
		openai.UserMessage(prompts.TemplateENPrompts),
		openai.UserMessage(prompts.TemplateZHPrompts),
		openai.UserMessage(prompts.GeneratePrompts),
		openai.UserMessage(releaseNoteResp),
	}
}

func (h *Helper) generateChangelog() error {
	resp, err := h.client.Chat.Completions.New(context.Background(), openai.ChatCompletionNewParams{
		Messages: h.messages,
		Model:    MODEL,
	})
	if err != nil {
		return fmt.Errorf("error get chat completions %s", err.Error())
	}

	if len(resp.Choices) == 0 {
		return fmt.Errorf("no response choices received from OpenAI")
	}

	content := resp.Choices[0].Message.Content
	parts := strings.Split(content, "--------")

	if len(parts) < 2 {
		return fmt.Errorf("error invalid content format")
	}
	enContent := removeChar(processContent(parts[0]))
	zhContent := removeChar(processContent(parts[1]))
	err = createMarkdownFile("CHANGELOG/CHANGELOG-"+VERSION+".md", enContent)
	if err != nil {
		return fmt.Errorf("error creating English changelog: %s", err.Error())
	}

	err = createMarkdownFile("CHANGELOG/CHANGELOG-"+VERSION+"-zh.md", zhContent)
	if err != nil {
		return fmt.Errorf("error creating Chinese changelog: %s", err.Error())
	}

	logger.Infof("successfully creating changelogs")
	return nil
}

func processContent(content string) string {
	lines := strings.Split(content, "\n")
	var processedLines []string

	for _, line := range lines {
		if trimmed := strings.TrimSpace(line); trimmed != "" {
			processedLines = append(processedLines, trimmed)
		}
	}

	return strings.Join(processedLines, "\n\n")
}

func createMarkdownFile(path, content string) error {
	return os.WriteFile(path, []byte(content), 0644)
}

func removeChar(s string) string {
	// remove all "```" and "```markdown"
	s = strings.ReplaceAll(s, "```markdown", "")
	s = strings.ReplaceAll(s, "```", "")
	s = strings.TrimSpace(s)
	
	// remove "markdown" at the beginning of the string if it exists
	if strings.HasPrefix(s, "markdown") {
		s = strings.TrimPrefix(s, "markdown")
		s = strings.TrimSpace(s)
	}
	
	return s
}