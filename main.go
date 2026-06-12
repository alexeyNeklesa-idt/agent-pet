package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/joho/godotenv"
	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
)

func fetchAbout(ctx context.Context) any {
	payload := map[string]any{
		"operationName": "about",
		"query":         "query about {\n  about {\n    name\n    startTimeStamp\n    version\n    __typename\n  }\n}\n",
		"variables":     map[string]any{},
	}
	body, _ := json.Marshal(payload)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, cmtServiceURL+"?about", bytes.NewReader(body))
	if err != nil {
		return map[string]string{"error": err.Error()}
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "1234")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return map[string]string{"error": err.Error()}
	}
	defer resp.Body.Close()

	var result any
	json.NewDecoder(resp.Body).Decode(&result)
	return result
}

func fetchCustomer(ctx context.Context, userID string) any {
	payload := map[string]any{
		"operationName": "getCustomer",
		"query":         "query getCustomer($input: GetCustomerInput!) {\n  getCustomer(input: $input) {\n    accountCreatedAt\n    accountUpdatedAt\n    alias\n    brCustomerId\n    unverifiedEmail\n    verifiedEmail\n    encryptedAddress\n    encryptedDateOfBirth\n    fullName\n    phone\n    profileStatus\n    lockingStatus\n    preferredLanguage\n    id\n    walletKycStatus\n    moneyAppActivatedDate\n    walletId\n    country\n    __typename\n  }\n}\n",
		"variables": map[string]any{
			"input": map[string]any{"id": userID},
		},
	}
	body, _ := json.Marshal(payload)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, cmtServiceURL+"?customer", bytes.NewReader(body))
	if err != nil {
		return map[string]string{"error": err.Error()}
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", cmtAuthToken)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return map[string]string{"error": err.Error()}
	}
	defer resp.Body.Close()

	var result any
	json.NewDecoder(resp.Body).Decode(&result)
	return result
}

func about(w http.ResponseWriter, r *http.Request) {
	result := fetchAbout(r.Context())
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{"type": "about", "response": result})
}

func getCustomer(w http.ResponseWriter, r *http.Request) {
	var body struct {
		UserID string `json:"userId"`
	}
	json.NewDecoder(r.Body).Decode(&body)
	result := fetchCustomer(r.Context(), body.UserID)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

var getDataTool = openai.ChatCompletionToolParam{
	Type: "function",
	Function: openai.FunctionDefinitionParam{
		Name:        "getData",
		Description: openai.String("Fetch data from the backend. Use type 'about' for service info, type 'customer' with a userId for customer profile."),
		Parameters: openai.FunctionParameters{
			"type": "object",
			"properties": map[string]any{
				"type": map[string]any{
					"type":        "string",
					"enum":        []string{"about", "customer"},
					"description": "Type of data to fetch: 'about' for system info, 'customer' for customer profile",
				},
				"userId": map[string]any{
					"type":        "string",
					"description": "Customer ID — required when type is 'customer'",
				},
			},
			"required": []string{"type"},
		},
	},
}

func chat(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Message string `json:"message"`
	}
	json.NewDecoder(r.Body).Decode(&body)

	client := openai.NewClient(option.WithAPIKey(os.Getenv("OPENAI_API_KEY")))
	messages := []openai.ChatCompletionMessageParamUnion{
		openai.UserMessage(body.Message),
	}

	for {
		resp, err := client.Chat.Completions.New(
			r.Context(),
			openai.ChatCompletionNewParams{
				Model:    openai.ChatModelGPT4oMini,
				Messages: messages,
				Tools:    []openai.ChatCompletionToolParam{getDataTool},
			},
		)
		if err != nil {
			http.Error(w, "AI error: "+err.Error(), http.StatusInternalServerError)
			return
		}

		msg := resp.Choices[0].Message
		messages = append(messages, msg.ToParam())

		if resp.Choices[0].FinishReason != "tool_calls" {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]any{"message": msg.Content})
			return
		}

		tc := msg.ToolCalls[0]
		var args struct {
			Type   string `json:"type"`
			UserID string `json:"userId"`
		}
		json.Unmarshal([]byte(tc.Function.Arguments), &args)

		var response any
		switch args.Type {
		case "about":
			response = fetchAbout(r.Context())
		case "customer":
			response = fetchCustomer(r.Context(), args.UserID)
		default:
			http.Error(w, "unknown type: "+args.Type, http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"type":     args.Type,
			"response": response,
		})
		return
	}
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
var cmtAuthToken string

func main() {
	if err := godotenv.Load(); err != nil {
		fmt.Println("Error loading .env")
		return
	}
	cmtServiceURL = os.Getenv("CMT_SERVICE_URL")
	cmtAuthToken = os.Getenv("CMT_AUTH_TOKEN")

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
