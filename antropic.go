package main

import "fmt"

type AnthropicOutput struct {
	Completion string `json:"completion"`
	StopReason string `json:"stop_reason"`
}


type AnthropicInput struct {
	Prompt string `json:"prompt"`
	MaxTokensToSample int `json:"max_tokens_to_sample"`
	Temperature float64 `json:"temperature"`
	TopK int `json:"top_k"`
	TopP float64 `json:"top_p"`
	StopSequences []string `json:"stop_sequences"`
}

type MessageModule struct {
	Speaker string
	Message string
}

type Conversation struct {
	Messages []MessageModule
}

func (c *Conversation) AddMessage(speaker string, message string) {
	c.Messages = append(c.Messages, MessageModule{
		Speaker: speaker,
		Message: message,
	})
}

func (c *Conversation) GetMessagesBySpeaker(speaker string) []MessageModule {
	var messages []MessageModule
	for _, message := range c.Messages {
		if message.Speaker == speaker {
			messages = append(messages, message)
		}
	}
	return messages
}

func (c *Conversation) ToString() string {
	var messages string
	for _, message := range c.Messages {
		messages += fmt.Sprintf("\n\n%s: %s\n", message.Speaker, message.Message)
	}
	return messages
}
