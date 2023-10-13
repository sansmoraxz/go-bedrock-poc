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



func EmbeddingsDemo(ctx context.Context, prompt string, bedrockSvc *bedrockruntime.Client) {
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

func ClaudeInvokeStreamingDemo(ctx context.Context, bedrockSvc *bedrockruntime.Client, prompt string, outChannel chan<- string) {
	var err error

	defer close(outChannel)
	jsonInput, err := json.Marshal(AnthropicInput{
		Prompt: fmt.Sprintf("\n\nHuman: %s\n\nAssistant: ```md", prompt),
		MaxTokensToSample: 2048,
		Temperature: 0.0,
		TopK: 250,
		TopP: 0.999,
		StopSequences: []string{"Human:"},
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
	var modelOutput AnthropicOutput
	for event := range eventsChannel {
		x := event.(*bedrockruntimetypes.ResponseStreamMemberChunk)
		json.Unmarshal(x.Value.Bytes, &modelOutput)
		outChannel <- modelOutput.Completion
	}

}

func main() {
	// 5 mins timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer func ()  {
		println("Canceling context")
		cancel()
	}()

	// start timer
	start := time.Now()

	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		panic(err)
	}
	bedrockSvc := bedrockruntime.NewFromConfig(cfg)

	
	prompts := []string{
		"Write an essay on the topic of 'The History of the United States'.",
		"Write an essay on the topic of 'The Battle of Gettysburg'.",
		"Why an essay on the topic of 'French Revolution'",
		"When was the Declaration of Independence signed?",
		"Who was the first president of the United States?",
		"What is the capital of the United States?",
		"Explain nuclear fusion.",
		"Tell me about Napoleon Bonaparte.",
	}

	k := len(prompts)


	// create output channels
	outchannels := make([]chan string, k)
	wg := sync.WaitGroup{}

	// create goroutines
	for i := 0; i < k; i++ {
		outchannels[i] = make(chan string)
		go ClaudeInvokeStreamingDemo(ctx, bedrockSvc, prompts[i], outchannels[i])
	}

	os.Mkdir("out", 0777)

	// read from the output channels
	for i := 0; i < k; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			outfile := fmt.Sprintf("out/output%d.md", i)
			f, err := os.Create(outfile)
			if err != nil {
				panic(err)
			}
			for out := range outchannels[i] {
				fmt.Println("Prompt ", i, ": ", out)
				f.WriteString(out)
			}
		}(i)
	}

	wg.Wait()

	// stop timer
	elapsed := time.Since(start)

	fmt.Printf("Time taken: %s\n", elapsed)

}
