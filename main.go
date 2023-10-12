package main

import (
	"context"
	"encoding/json"
	"fmt"

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



func embeddingsDemo(ctx context.Context, bedrockSvc *bedrockruntime.Client) {
	modelInput := &bedrockruntime.InvokeModelInput{
		ModelId: aws.String("amazon.titan-embed-text-v1"),
		Accept: aws.String("*/*"),
		ContentType: aws.String("application/json"),
		Body: []byte(`{"inputText": "This is a test"}`),
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

func claudeInvokeStreamingDemo(ctx context.Context, bedrockSvc *bedrockruntime.Client) {
	var err error

	jsonInput, err := json.Marshal(map[string]interface{}{
		"prompt": "\n\nHuman: What is 2+2?\n\nAssistant:",
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
	for {
		select {
		case event, ok := <-eventsChannel:
			if !ok {
				fmt.Println("Channel closed")
				return
			}
			x := event.(*bedrockruntimetypes.ResponseStreamMemberChunk)
			var response AnthropicOutput
			json.Unmarshal(x.Value.Bytes, &response)
			fmt.Println("..")
			fmt.Println("Stop reason: ", response.StopReason)
			fmt.Println("Completion: ", response.Completion)
		case <-ctx.Done():
			fmt.Println("Context done")
		}
	}

}

func main() {
	ctx := context.TODO()
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		panic(err)
	}

	bedrockSvc := bedrockruntime.NewFromConfig(cfg)

	embeddingsDemo(ctx, bedrockSvc)

	claudeInvokeStreamingDemo(ctx, bedrockSvc)
}
