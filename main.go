package main

import (
	"encoding/json"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/bedrockruntime"
)

type Embeddings struct {
	Embeddings []float64 `json:"embedding"`
	InputTextTokenCount int `json:"inputTextTokenCount"`
}


func embeddingsDemo(bedrockSvc *bedrockruntime.BedrockRuntime) {
	modelInput := &bedrockruntime.InvokeModelInput{
		ModelId: aws.String("amazon.titan-embed-text-v1"),
		Accept: aws.String("*/*"),
		ContentType: aws.String("application/json"),
		Body: []byte(`{"inputText": "This is a test"}`),
	}
	result, err := bedrockSvc.InvokeModel(modelInput)
	if err != nil {
		panic(err)
	}
	var response Embeddings
	json.Unmarshal(result.Body, &response)

	// Print the embeddings
	fmt.Println("Embeddings: ", response.Embeddings)

}

func claudeInvokeStreamingDemo(bedrockSvc *bedrockruntime.BedrockRuntime) {
	var err error

	jsonInput, err := json.Marshal(map[string]interface{}{
		"prompt": "\n\nHuman: What is 2+2?\n\nAssistant",
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
	result, err := bedrockSvc.InvokeModelWithResponseStream(modelInput)
	if err != nil {
		panic(err)
	}
	fmt.Println("Response: ", result)
	reader := result.GetStream().Reader
	reader.Err()
	eventsChannel := reader.Events()
	for {
		select {
		case event, ok := <-eventsChannel:
			if !ok {
				fmt.Println("Channel closed")
				return
			}
			fmt.Println("Event: ", event)
		}
	}

}

func main() {
	mySession := session.Must(session.NewSession())
	bedrockSvc := bedrockruntime.New(mySession, aws.NewConfig().WithRegion("us-east-1"))

	// embeddingsDemo(bedrockSvc)

	claudeInvokeStreamingDemo(bedrockSvc)
}
