package gpt

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/ai/azopenai"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/edgenesis/shifu/pkg/logger"
	"github.com/edgenesis/shifu/tools/release/prompts"
)

var (
	API_KEY = os.Getenv("AZURE_OPENAI_APIKEY")
	HOST    = os.Getenv("AZURE_OPENAI_HOST")
	VERSION = os.Getenv("VERSION")
	GPT4    = "gpt-4"
)

type Helper struct {
	client   *azopenai.Client
	messages []azopenai.ChatMessage
}

func Start(releaseNoteResp string) error {
	var helper *Helper

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

func newGPT() (*azopenai.Client, error) {
	ky, err := azopenai.NewKeyCredential(API_KEY)
	if err != nil {
		return nil, fmt.Errorf("error new key credential %s", err.Error())
	}

	client, err := azopenai.NewClientWithKeyCredential("https://"+HOST+".openai.azure.com", ky, nil)
	if err != nil {
		return nil, fmt.Errorf("error new azure client %s", err.Error())
	}
	return client, nil
}

func (h *Helper) generateMessages(releaseNoteResp string) {
	h.messages = []azopenai.ChatMessage{
		{
			Role:    to.Ptr(azopenai.ChatRoleUser),
			Content: to.Ptr(prompts.GreetingPrompts),
		},
		{
			Role:    to.Ptr(azopenai.ChatRoleUser),
			Content: to.Ptr(prompts.TemplateENPrompts),
		},
		{
			Role:    to.Ptr(azopenai.ChatRoleUser),
			Content: to.Ptr(prompts.TemplateZHPrompts),
		},
		{
			Role:    to.Ptr(azopenai.ChatRoleUser),
			Content: to.Ptr(prompts.GeneratePrompts),
		},
		{
			Role:    to.Ptr(azopenai.ChatRoleUser),
			Content: to.Ptr(releaseNoteResp),
		},
	}
}

func (h *Helper) generateChangelog() error {
	resp, err := h.client.GetChatCompletions(context.Background(), azopenai.ChatCompletionsOptions{
		Messages:    h.messages,
		Deployment:  GPT4,
		Temperature: toPointer(float32(0)),
	}, nil)
	if err != nil {
		return fmt.Errorf("error get chat completions %s", err.Error())
	}

	content := *resp.Choices[0].Message.Content
	parts := strings.Split(content, "--------")

	if len(parts) < 2 {
		return fmt.Errorf("error invalid content format")
	}
	enContent := processContent(parts[0])
	zhContent := processContent(parts[1])
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

func toPointer[T any](v T) *T {
	return &v
}
