package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"github.com/rs/cors"
)

var projectDescription string
var applicationName string = "default-app-name" // Default app name
var step int = 1                                // Variável para controlar o step

type ChatRequest struct {
	Message string `json:"message"`
	Step    int    `json:"step"`
}

type ChatGPTRequest struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ChatGPTResponse struct {
	Choices []struct {
		Message Message `json:"message"`
	} `json:"choices"`
}

type ChatResponse struct {
	Response string `json:"response"`
	Step     int    `json:"step"`
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	r := mux.NewRouter()
	r.HandleFunc("/api/chat", chatHandler).Methods("POST")

	corsMiddleware := cors.New(cors.Options{
		AllowedOrigins:   []string{"http://localhost:3001"}, // Porta correta do frontend
		AllowedMethods:   []string{"GET", "POST", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Content-Type", "X-Requested-With"},
		AllowCredentials: true, // Permitir credenciais, se necessário
		Debug:            true,
	})

	log.Println("Server started on port 8000...")
	log.Fatal(http.ListenAndServe(":8000", corsMiddleware.Handler(r)))
}

func chatHandler(w http.ResponseWriter, r *http.Request) {
	var req ChatRequest
	_ = json.NewDecoder(r.Body).Decode(&req)

	log.Println("Received message from user:", req.Message)
	log.Println("Step:", step)

	switch step {
	case 1:
		initialMessage := createInitialMessage(req.Message)
		projectDescription = getChatGPTResponse(initialMessage)
		sendResponseToUser(w, "This is the project description:\n\n"+projectDescription+"\n\nIs this what you were thinking? Can we move forward?")
		step++

	case 2:
		if req.Message == "YES" {
			log.Println("Starting backend setup for application:", applicationName)

			// Commands for yaml
			commandoyaml := `I want you to provide the simplest structure of all the files in the project in YAML format. The structure should follow this exact model:

            - Each folder is a key.
            - Each file is listed under its corresponding folder key.
            - Keep the structure as minimal as possible, with everything inside a single file called main.go.

            The YAML model should look like this:

            simple-go-api:
              - "main.go"

            Provide the response exactly in this format.`

			newPrompt := projectDescription + getBackendSetupInstructions() + commandoyaml + " with ```yaml in the beginning ``` in the end."
			structureAndFiles := getChatGPTResponse(newPrompt)

			text := `The system should be implemented in the main.go file and contain the following parts:

            1. Product Structure (Struct):
                - Create a Go struct called Product with the following fields:
                - ID (int)
                - Name (string)
                - Price (float64)
                - Description (string)

            2. Add Product Function (POST):
                - Create a function called addProduct that:
                - Handles HTTP POST requests.
                - Decodes the product data received via JSON.
                - Stores the product in a list of products.
                - Responds with the created product in JSON format.

            3. Get Products Function (GET):
                - Create a function called getProducts that:
                - Handles HTTP GET requests.
                - Returns the list of stored products in JSON format.

            4. Web Interface:
                - The interface should be built using HTML and JavaScript (with fetch).
                - It should allow users to add products through a form.
                - The list of added products should be displayed in real-time on the interface.
                - The interface should interact directly with the APIs (POST and GET) using JavaScript.

            5. Execution:
                - The APIs should run on port 4000 and the web interface on port 4001.
                - Use the net/http package to serve both the APIs and the web interface.`

			log.Println("YAML content extracted:", structureAndFiles)

			// Extrair o conteúdo YAML
			yamlContent := extractYAMLContent(structureAndFiles)
			log.Println("YAML content extracted:", yamlContent)

			// Criar o arquivo project.yaml e escrever o conteúdo YAML nele usando a função de codificação
			err := saveYAMLFile("project.yaml", yamlContent)
			if err != nil {
				log.Fatalf("Error writing YAML to file: %v", err)
			}
			log.Println("YAML content written to project.yaml")

			// Leia o conteúdo do arquivo Code_template.txt
			codeTemplateFileName := "Code_template.txt"
			codeTemplateContent, err := ioutil.ReadFile(codeTemplateFileName)
			if err != nil {
				log.Fatalf("Error reading code template file: %v", err)
			}

			// Concatene o conteúdo do arquivo com o prompt existente
			promptToPython := projectDescription + text + " I want the code like this: " + string(codeTemplateContent)

			// Write prompt to a file in the same directory as main.go
			promptFileName := "promptFile.txt"
			err = ioutil.WriteFile(promptFileName, []byte(promptToPython), 0644)
			if err != nil {
				log.Fatalf("Error writing prompt to file: %v", err)
			}

			// Enviar o conteúdo do YAML de volta para o front-end para confirmar
			sendResponseToUser(w, "```yaml\n"+yamlContent+"\n```")
			step++

		} else {
			sendResponseToUser(w, "Please provide the necessary details to start.")
		}

	case 3:
		if req.Message == "YES" {

			log.Println("User confirmed YAML structure. Executing Python script to create directory and files.")

			// Executa o script Python para criar pastas e arquivos
			executePythonScript("project.yaml", "promptFile.txt")

			// Executa o projeto Go
			runGoProject()

			// Gera a URL após execução do projeto Go
			url := "http://localhost:4001" // Substitua pela URL gerada dinamicamente, se necessário

			// Cria o link HTML clicável
			link := "<a href='" + url + "' target='_blank'>Access the system here</a>"

			// Enviar a resposta final com o link HTML ao frontend
			finalMessage := "The directory structure has been created successfully. " + link
			sendResponseToUser(w, finalMessage)

			step++

		} else {
			sendResponseToUser(w, "Please provide the necessary details to start.")
		}

	default:
		sendResponseToUser(w, "Please provide the necessary details to start.")
	}
}

// Função para salvar o arquivo YAML com a codificação adequada
func saveYAMLFile(filename, content string) error {
	// Cria ou abre o arquivo
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	// Codifica o conteúdo em UTF-8, se necessário
	writer := bufio.NewWriter(f)
	_, err = writer.WriteString(content)
	if err != nil {
		return err
	}
	return writer.Flush()
}

func executePythonScript(yamlFile, promptFile string) {
	log.Println("Executing Python script to create project structure...")

	pythonPath := "/Users/jeovahsimoes/Documents/vtkl/ai-sdlc/backend/venv/bin/python3"
	cmd := exec.Command(pythonPath, "create_project_files.py", yamlFile, promptFile)

	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		log.Fatalf("Python script error: %v: %v", err, stderr.String())
	}

	log.Printf("Python script output: %s", out.String())
}

func runGoProject() {
	log.Println("Running 'go mod init', 'go mod tidy', and 'go run main.go'...")

	// Step 1: Run 'go mod init'
	cmd := exec.Command("go", "mod", "init", "api-default-app-name")
	cmd.Dir = "../simple-go-api"
	runCommand(cmd, "go mod init")

	// Step 2: Run 'go mod tidy'
	cmd = exec.Command("go", "mod", "tidy")
	cmd.Dir = "../simple-go-api"
	runCommand(cmd, "go mod tidy")

	// Step 3: Run 'go run main.go'
	cmd = exec.Command("go", "run", "main.go")
	cmd.Dir = "../simple-go-api"
	startGoRunCommand(cmd, "go run main.go")
}

func startGoRunCommand(cmd *exec.Cmd, description string) {
	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr

	// Run the command in the background
	err := cmd.Start()
	if err != nil {
		log.Fatalf("Failed to start '%s': %v: %v", description, err, stderr.String())
	}

	log.Printf("%s started: %s", description, out.String())

	// Optional: Wait for the process to finish if you want to block until it's done
	go func() {
		err = cmd.Wait()
		if err != nil {
			log.Fatalf("Process '%s' finished with error: %v", description, err)
		} else {
			log.Printf("Process '%s' finished successfully", description)
		}
	}()
}

func runCommand(cmd *exec.Cmd, description string) {
	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		log.Fatalf("Failed to run '%s': %v: %v", description, err, stderr.String())
	}

	log.Printf("%s output: %s", description, out.String())
}

func createInitialMessage(userMessage string) string {

	return "The user wants to build: " + userMessage + ". At this point, I want you to create a user story with acceptance criteria, then build a web application with both frontend and backend, define the necessary APIs and their URLs and show its structure, and finally, display a list with the following: User Story: As a [user type], I want [desired action or feature] so that [benefit or expected outcome]. Acceptance Criteria: Given [initial situation], when [user action], then [expected outcome]. Effort Estimate: This feature is classified as [low, medium, high] effort for development. Time Estimate: We estimate it will take approximately [number of hours/days] for a developer to complete this user story. API Functionality: Each API will have the following responsibility."

}

func getBackendSetupInstructions() string {
	request := `
        I want to build a simple backend system in Go based on the user story provided. Here are the final requirements:

        1. API Endpoints:
        - Create two API endpoints:
            - Insert Data (POST): This will accept JSON data and store it in memory.
            - Read Data (GET): This will return all stored data in JSON format.
        - Use Gorilla Mux for routing.

        2. Data Handling:
        - Store the data in memory (no need for a database). The data will be lost if the application is restarted.
        - Implement the logic directly in the handler functions for storing and retrieving data.

        3. Project Structure:
        - Everything should be in one single main.go file.
        - The code should handle both inserting and reading data within this file.

        4. Interface and Usage:
        - Include a simple web interface (HTML form) within the Go code to allow testing of the APIs via a URL.
        - The system should be accessible via a browser for inserting data and viewing the stored data.
        `

	return request
}

func getChatGPTResponse(message string) string {
	chatGPTReq := ChatGPTRequest{
		Model: "gpt-4", // Using GPT-4
		Messages: []Message{
			{
				Role:    "user",
				Content: message,
			},
		},
	}

	body, err := json.Marshal(chatGPTReq)
	if err != nil {
		log.Fatal("Error marshalling request: ", err)
	}

	apiKey := os.Getenv("CHATGPT_API_KEY")
	req, err := http.NewRequest("POST", "https://api.openai.com/v1/chat/completions", bytes.NewBuffer(body))
	if err != nil {
		log.Fatal("Error creating request: ", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal("Error making API request: ", err)
	}
	defer resp.Body.Close()

	responseBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal("Error reading response: ", err)
	}

	var chatGPTResp ChatGPTResponse
	err = json.Unmarshal(responseBody, &chatGPTResp)
	if err != nil {
		log.Fatal("Error unmarshalling response: ", err)
	}

	return chatGPTResp.Choices[0].Message.Content
}

func extractYAMLContent(response string) string {
	start := strings.Index(response, "```yaml")
	if start == -1 {
		return ""
	}
	start += len("```yaml")
	end := strings.Index(response[start:], "```")
	if end == -1 {
		return ""
	}
	return strings.TrimSpace(response[start : start+end])
}

func sendResponseToUser(w http.ResponseWriter, response string) {
	chatResponse := ChatResponse{Response: response, Step: step}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(chatResponse)
}
