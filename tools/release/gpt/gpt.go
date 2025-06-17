package gpt

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/edgenesis/shifu/pkg/logger"
	"github.com/edgenesis/shifu/tools/release/prompts"
	"github.com/openai/openai-go"
	"github.com/openai/openai-go/azure"
)

var (
	API_KEY         = os.Getenv("AZURE_OPENAI_APIKEY")
	HOST            = os.Getenv("AZURE_OPENAI_HOST")
	DEPLOYMENT_NAME = os.Getenv("DEPLOYMENT_NAME")
	VERSION         = os.Getenv("VERSION")
	API_VERSION     = "2024-06-01" // Latest Azure OpenAI API version
)

type Helper struct {
	client   *openai.Client
	messages []openai.ChatCompletionMessageParamUnion
}

func Start(releaseNoteResp string) error {
	helper := &Helper{}
	client := newGPT()

	helper.client = client

	helper.generateMessages(releaseNoteResp)

	err := helper.generateChangelog()
	if err != nil {
		return err
	}

	return nil
}

func newGPT() *openai.Client {
	client := openai.NewClient(
		azure.WithEndpoint(HOST, API_VERSION),
		azure.WithAPIKey(API_KEY),
	)
	return &client
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
		Messages:    h.messages,
		Model:       openai.ChatModel(DEPLOYMENT_NAME),
		Temperature: openai.Float(0.0),
	})
	if err != nil {
		return fmt.Errorf("error get chat completions %s", err.Error())
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
	// remove all "```"
	str := "```"
	return strings.ReplaceAll(s, str, "")
}
