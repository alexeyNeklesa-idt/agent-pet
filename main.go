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

	for {
		fmt.Printf("Write: ")

		userInput, _ := reader.ReadString('\n')

		response, err := client.Chat.Completions.New(
			context.Background(),
			openai.ChatCompletionNewParams{
				Model: openai.ChatModelGPT4oMini,
				Messages: []openai.ChatCompletionMessageParamUnion{
					openai.UserMessage(userInput),
				},
			},
		)

		if err != nil {
			fmt.Println("Error: ", err)
		}

		fmt.Println(response)
	}
}
