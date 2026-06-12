package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/joho/godotenv"
)

func about(w http.ResponseWriter, r *http.Request) {
	payload := map[string]any{
		"operationName": "about",
		"query":         "query about {\n  about {\n    name\n    startTimeStamp\n    version\n    __typename\n  }\n}\n",
		"variables":     map[string]any{},
	}

	body, _ := json.Marshal(payload)

	req, err := http.NewRequestWithContext(r.Context(), http.MethodPost, cmtServiceURL+"?about", bytes.NewReader(body))
	if err != nil {
		http.Error(w, "failed to create request", http.StatusInternalServerError)
		return
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "1234")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		http.Error(w, "upstream request failed: "+err.Error(), http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	var upstreamBody any
	json.NewDecoder(resp.Body).Decode(&upstreamBody)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(resp.StatusCode)
	json.NewEncoder(w).Encode(map[string]any{
		"type":     "about",
		"response": upstreamBody,
	})
}

func chat(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Hello world"})
}

func getCustomer(w http.ResponseWriter, r *http.Request) {
	payload := map[string]any{
		"operationName": "getCustomer",
		"query":         "query getCustomer($input: GetCustomerInput!) {\n  getCustomer(input: $input) {\n    accountCreatedAt\n    accountUpdatedAt\n    alias\n    brCustomerId\n    unverifiedEmail\n    verifiedEmail\n    encryptedAddress\n    encryptedDateOfBirth\n    fullName\n    phone\n    profileStatus\n    lockingStatus\n    preferredLanguage\n    id\n    walletKycStatus\n    moneyAppActivatedDate\n    walletId\n    country\n    __typename\n  }\n}\n",
		"variables": map[string]any{
			"input": map[string]any{
				"id": "ffvtr91r2e2s",
			},
		},
	}

	body, _ := json.Marshal(payload)

	req, err := http.NewRequestWithContext(r.Context(), http.MethodPost, cmtServiceURL+"?customer", bytes.NewReader(body))
	if err != nil {
		http.Error(w, "failed to create request", http.StatusInternalServerError)
		return
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer eyJhbGciOiJIUzI1NiJ9.eyJncmFudHMiOltdLCJ1bmxvY2tIb3VycyI6MCwicm9sZXMiOlsiUk9MRV9BY2NvdW50IFJlY2VpdmFibGUiLCJST0xFX0N1c3RvbWVyU2VydmljZSBPcHMiLCJST0xFX0ZyYXVkIiwiUk9MRV9BZG1pbiIsIlJPTEVfRGV2ZWxvcGVyIl0sImNyZWRpdExpbWl0IjoiMCIsImV4cGlyYXRpb24iOiIyMDI2LTA2LTExVDE5OjM0OjQ1LjQ2NTA2MDA3OVoiLCJpZCI6IjMyMDUyNCIsImVtYWlsIjoiQWxpYWtzZWkuTmlha2xlc2FAaWR0Lm5ldCIsInVzZXJuYW1lIjoiYW5pYWtsZXNhIn0.5chT5fP9cm1VXeiKv-MuGBSYgI5Iw_sl4y1MfYNL8YQ")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		http.Error(w, "upstream request failed: "+err.Error(), http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body)
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}

var cmtServiceURL string

func main() {
	if err := godotenv.Load(); err != nil {
		fmt.Println("Error loading .env")
		return
	}
	cmtServiceURL = os.Getenv("CMT_SERVICE_URL")

	r := chi.NewRouter()
	r.Use(corsMiddleware)
	r.Post("/customer", getCustomer)
	r.Post("/test", about)
	r.Post("/chat", chat)
	http.ListenAndServe(":8080", r)
}

// import (
// 	"bufio"
// 	"context"
// 	"encoding/json"
// 	"fmt"
// 	"os"

// 	"github.com/joho/godotenv"
// 	"github.com/openai/openai-go"
// 	"github.com/openai/openai-go/option"
// )

// func getWeather(city string) string {
// 	return fmt.Sprintf("Current weather for %s is 20°C", city)
// }

// func main() {
// 	err := godotenv.Load()

// 	if err != nil {
// 		fmt.Println("Error env")
// 		return
// 	}

// 	apiKey := os.Getenv("OPENAI_API_KEY")
// 	client := openai.NewClient(
// 		option.WithAPIKey(apiKey),
// 	)
// 	tools := []openai.ChatCompletionToolParam{
// 		{
// 			Type: "function",
// 			Function: openai.FunctionDefinitionParam{
// 				Name:        "getWeather",
// 				Description: openai.String("Get the current weather for a location"),
// 				Parameters: openai.FunctionParameters{
// 					"type": "object",
// 					"properties": map[string]any{
// 						"location": map[string]any{
// 							"type":        "string",
// 							"description": "City name, e.g. Warsaw",
// 						},
// 					},
// 					"required": []string{"location"},
// 				},
// 			},
// 		},
// 	}

// 	reader := bufio.NewReader(os.Stdin)
// 	allMessages := []openai.ChatCompletionMessageParamUnion{}

// 	for {
// 		fmt.Print("Write: ")
// 		userInput, _ := reader.ReadString('\n')
// 		allMessages = append(allMessages, openai.UserMessage(userInput))

// 		// agentic loop: repeat until the model stops calling tools
// 		for {
// 			stream := client.Chat.Completions.NewStreaming(
// 				context.Background(),
// 				openai.ChatCompletionNewParams{
// 					Model:    openai.ChatModelGPT4oMini,
// 					Messages: allMessages,
// 					Tools:    tools,
// 				},
// 			)

// 			// accumulator merges all chunks into a complete message (incl. tool_calls)
// 			acc := openai.ChatCompletionAccumulator{}
// 			for stream.Next() {
// 				chunk := stream.Current()
// 				acc.AddChunk(chunk)
// 				// stream text to terminal as it arrives
// 				if len(chunk.Choices) > 0 {
// 					fmt.Print(chunk.Choices[0].Delta.Content)
// 				}
// 			}

// 			if err := stream.Err(); err != nil {
// 				fmt.Println("Stream error:", err)
// 				break
// 			}

// 			// ToParam() returns the full assistant message including tool_calls
// 			allMessages = append(allMessages, acc.Choices[0].Message.ToParam())

// 			if acc.Choices[0].FinishReason == "tool_calls" {
// 				for _, tc := range acc.Choices[0].Message.ToolCalls {
// 					var args map[string]string
// 					json.Unmarshal([]byte(tc.Function.Arguments), &args)

// 					var result string
// 					switch tc.Function.Name {
// 					case "getWeather":
// 						fmt.Printf("\n[calling getWeather(city=%s)]\n", args["location"])
// 						result = getWeather(args["location"])
// 					default:
// 						result = "unknown tool: " + tc.Function.Name
// 					}

// 					allMessages = append(allMessages, openai.ToolMessage(result, tc.ID))
// 				}
// 				continue
// 			}

// 			fmt.Println("\n--------------------------------------------------------")
// 			break
// 		}
// 	}
// }
