package main

import (
	"bufio"
	"context"
	"fmt"
	"os"

	"github.com/joho/godotenv"
	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
)

func main() {
	err := godotenv.Load()

	if err != nil {
		fmt.Println("Error env")
		return
	}

	apiKey := os.Getenv("OPENAI_API_KEY")

	client := openai.NewClient(
		option.WithAPIKey(apiKey),
	)

	reader := bufio.NewReader(os.Stdin)

	allMessages := []openai.ChatCompletionMessageParamUnion{
		openai.SystemMessage("Add at start of every message 'HELLO THERE' text"),
	}

	for {
		fmt.Printf("Write: ")

		userInput, _ := reader.ReadString('\n')

		allMessages = append(allMessages, openai.UserMessage(userInput))

		stream := client.Chat.Completions.NewStreaming(
			context.Background(),
			openai.ChatCompletionNewParams{
				Model:       openai.ChatModelGPT4oMini,
				Messages:    allMessages,
				Temperature: openai.Float(2.0),
			},
		)

		var assistantResponse string
		for stream.Next() {
			chunk := stream.Current()
			if len(chunk.Choices) > 0 {
				delta := chunk.Choices[0].Delta.Content
				fmt.Print(delta)
				assistantResponse += delta
			}
		}
		fmt.Println("\n--------------------------------------------------------")

		if err := stream.Err(); err != nil {
			fmt.Println("Error: ", err)
			continue
		}

		allMessages = append(allMessages, openai.AssistantMessage(assistantResponse))
	}
}
