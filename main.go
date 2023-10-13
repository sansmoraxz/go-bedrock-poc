package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"

	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime"
	bedrockruntimetypes "github.com/aws/aws-sdk-go-v2/service/bedrockruntime/types"
)

type Embeddings struct {
	Embeddings []float64 `json:"embedding"`
	InputTextTokenCount int `json:"inputTextTokenCount"`
}

type AnthropicOutput struct {
	Completion string `json:"completion"`
	StopReason string `json:"stop_reason"`
}



func embeddingsDemo(ctx context.Context, prompt string, bedrockSvc *bedrockruntime.Client) {
	// {"inputText": "This is a test"} as bytes
	body, err := json.Marshal(map[string]interface{}{
		"inputText": prompt,
	})
	if err != nil {
		panic(err)
	}
	modelInput := &bedrockruntime.InvokeModelInput{
		ModelId: aws.String("amazon.titan-embed-text-v1"),
		Accept: aws.String("*/*"),
		ContentType: aws.String("application/json"),
		Body: body,
	}
	result, err := bedrockSvc.InvokeModel(ctx, modelInput)
	if err != nil {
		panic(err)
	}
	var response Embeddings
	json.Unmarshal(result.Body, &response)

	// Print the embeddings
	fmt.Println("Embeddings: ", response.Embeddings)

}

func claudeInvokeStreamingDemo(ctx context.Context, bedrockSvc *bedrockruntime.Client, prompt string, outChannel chan<- string) {
	var err error

	defer close(outChannel)

	jsonInput, err := json.Marshal(map[string]interface{}{
		"prompt": fmt.Sprintf("\n\nHuman: %s\n\nAssistant: ```md", prompt),
		"max_tokens_to_sample": 2048,
		"temperature": 0.0,
		"top_k": 250,
		"top_p": 0.999,
		"stop_sequences": []string{"Human:"},
	})
	if err != nil {
		panic(err)
	}
	modelInput := &bedrockruntime.InvokeModelWithResponseStreamInput{
		ModelId: aws.String("anthropic.claude-v2"),
		Accept: aws.String("*/*"),
		ContentType: aws.String("application/json"),
		Body: jsonInput,
	}
	result, err := bedrockSvc.InvokeModelWithResponseStream(ctx, modelInput)
	if err != nil {
		panic(err)
	}
	fmt.Println("Response: ", result)
	reader := result.GetStream().Reader
	eventsChannel := reader.Events()
	for event := range eventsChannel {
		x := event.(*bedrockruntimetypes.ResponseStreamMemberChunk)
		var response AnthropicOutput
		json.Unmarshal(x.Value.Bytes, &response)
		outChannel <- response.Completion
	}

}

func main() {
	// 5 mins timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer func ()  {
		println("Canceling context")
		cancel()
	}()

	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		panic(err)
	}
	bedrockSvc := bedrockruntime.NewFromConfig(cfg)

	prompt1 := "Write an essay on the topic of 'The History of the United States'."
	prompt2 := "Write an essay on the topic of 'The Battle of Gettysburg'."

	// embeddingsDemo(ctx, prompt1, bedrockSvc)

	// output channel for streaming demo
	outChannel1 := make(chan string)
	outChannel2 := make(chan string)
	go claudeInvokeStreamingDemo(ctx, bedrockSvc, prompt1, outChannel1)
	go claudeInvokeStreamingDemo(ctx, bedrockSvc, prompt2, outChannel2)

	// read from the output channels
	wg := sync.WaitGroup{}
	wg.Add(2)
	go func() {
		defer wg.Done()
		f, err := os.Create("output1.txt")
		if err != nil {
			panic(err)
		}
		defer f.Close()
		for out := range outChannel1 {
			fmt.Println("Prompt 1: ", out)
			f.WriteString(out)
		}
	}()

	go func() {
		defer wg.Done()
		f, err := os.Create("output2.txt")
		if err != nil {
			panic(err)
		}
		defer f.Close()
		for out := range outChannel2 {
			fmt.Println("Prompt 2: ", out)
			f.WriteString(out)
		}
	}()

	wg.Wait()

}
