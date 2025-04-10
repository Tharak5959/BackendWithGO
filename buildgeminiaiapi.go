package main

import (
        "context"
        "encoding/json"
        "fmt"
        "log"
        "net/http"
        "os"

        "github.com/google/generative-ai-go/genai"
        "google.golang.org/api/option"
)

type Request struct {
        Document string `json:"document"`
        Question string `json:"question"`
}

type Response struct {
        Answer string `json:"answer"`
}

var client *genai.Client
var model *genai.GenerativeModel

func init() {
        ctx := context.Background()
        apiKey := os.Getenv("GOOGLE_API_KEY")
        //  apiKey := os.Getenv("AIzaSyBbesOrOyfkV5qvS24Go_D26uwOgchbryU") // Use GOOGLE_API_KEY, not the literal API key
       
        if apiKey == "" {
                log.Fatal("GOOGLE_API_KEY environment variable not set")
        }

        var err error
        client, err = genai.NewClient(ctx, option.WithAPIKey(apiKey))
        if err != nil {
                log.Fatalf("Failed to create client: %v", err)
        }

        model = client.GenerativeModel("gemini-1.5-flash")
}

func main() {
        http.HandleFunc("/answer", answerHandler)
        port := os.Getenv("PORT")
        if port == "" {
                port = "8080"
                log.Printf("defaulting to port %s", port)
        }
        log.Printf("listening on port %s", port)
        log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", port), nil))
}

func answerHandler(w http.ResponseWriter, r *http.Request) {
        if r.Method != http.MethodPost {
                http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
                return
        }

        var req Request
        err := json.NewDecoder(r.Body).Decode(&req)
        if err != nil {
                http.Error(w, err.Error(), http.StatusBadRequest)
                return
        }
        

        answer, err := getAnswer(req.Document, req.Question)
        if err != nil {
                http.Error(w, err.Error(), http.StatusInternalServerError)
                return
        }

        resp := Response{Answer: answer}
        w.Header().Set("Content-Type", "application/json")
        err = json.NewEncoder(w).Encode(resp)
        if err != nil {
                http.Error(w, "Internal Server Error", http.StatusInternalServerError)
                return
        }
}

func getAnswer(document, question string) (string, error) {
        ctx := context.Background()
        prompt := fmt.Sprintf("Given the following document:\n\n%s\n\nAnswer the question: %s", document, question)

        resp, err := model.GenerateContent(ctx, genai.Text(prompt))
        if err != nil {
                return "", fmt.Errorf("failed to generate content: %w", err)
        }

        if len(resp.Candidates) == 0 {
                return "", fmt.Errorf("no candidates returned")
        }

        if len(resp.Candidates[0].Content.Parts) == 0 {
                return "", fmt.Errorf("no parts returned")
        }

        return fmt.Sprint(resp.Candidates[0].Content.Parts[0]), nil
}