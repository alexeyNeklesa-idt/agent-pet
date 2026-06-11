package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/joho/godotenv"
	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
)

func getWeather(city string) string {
	return fmt.Sprintf("Current weather for %s is 20°C", city)
}

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
	tools := []openai.ChatCompletionToolParam{
		{
			Type: "function",
			Function: openai.FunctionDefinitionParam{
				Name:        "getWeather",
				Description: openai.String("Get the current weather for a location"),
				Parameters: openai.FunctionParameters{
					"type": "object",
					"properties": map[string]any{
						"location": map[string]any{
							"type":        "string",
							"description": "City name, e.g. Warsaw",
						},
					},
					"required": []string{"location"},
				},
			},
		},
	}

	reader := bufio.NewReader(os.Stdin)
	allMessages := []openai.ChatCompletionMessageParamUnion{}

	for {
		fmt.Print("Write: ")
		userInput, _ := reader.ReadString('\n')
		allMessages = append(allMessages, openai.UserMessage(userInput))

		// agentic loop: repeat until the model stops calling tools
		for {
			stream := client.Chat.Completions.NewStreaming(
				context.Background(),
				openai.ChatCompletionNewParams{
					Model:    openai.ChatModelGPT4oMini,
					Messages: allMessages,
					Tools:    tools,
				},
			)

			// accumulator merges all chunks into a complete message (incl. tool_calls)
			acc := openai.ChatCompletionAccumulator{}
			for stream.Next() {
				chunk := stream.Current()
				acc.AddChunk(chunk)
				// stream text to terminal as it arrives
				if len(chunk.Choices) > 0 {
					fmt.Print(chunk.Choices[0].Delta.Content)
				}
			}

			if err := stream.Err(); err != nil {
				fmt.Println("Stream error:", err)
				break
			}

			// ToParam() returns the full assistant message including tool_calls
			allMessages = append(allMessages, acc.Choices[0].Message.ToParam())

			if acc.Choices[0].FinishReason == "tool_calls" {
				for _, tc := range acc.Choices[0].Message.ToolCalls {
					var args map[string]string
					json.Unmarshal([]byte(tc.Function.Arguments), &args)

					var result string
					switch tc.Function.Name {
					case "getWeather":
						fmt.Printf("\n[calling getWeather(city=%s)]\n", args["location"])
						result = getWeather(args["location"])
					default:
						result = "unknown tool: " + tc.Function.Name
					}

					allMessages = append(allMessages, openai.ToolMessage(result, tc.ID))
				}
				continue
			}

			fmt.Println("\n--------------------------------------------------------")
			break
		}
	}
}
