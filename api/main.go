package main

import (
	"crypto/tls"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/smtp"
	"os"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
)

func CORSMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Set CORS headers globally
		w.Header().Set("Access-Control-Allow-Origin", "*")                                              // Allow any origin, change to specific domain if needed
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")                            // Allow methods
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With") // Allow headers

		// Handle preflight OPTIONS request
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		// Pass the request to the next handler
		next.ServeHTTP(w, r)
	})
}

var (
	db             *sql.DB
	smtpHost       string
	smtpPort       string
	smtpUsername   string
	smtpPassword   string
	contactFormKey string
	reviewAPIKey   string
)

func loadEnv() error {
	// Attempt to load .env file
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: No .env file found, relying on system environment variables.")
	}

	// Load environment variables
	smtpHost = os.Getenv("SMTP_HOST")
	smtpPort = os.Getenv("SMTP_PORT")
	smtpUsername = os.Getenv("SMTP_USERNAME")
	smtpPassword = os.Getenv("SMTP_PASSWORD")

	reviewAPIKey = os.Getenv("REVIEW_FORM_KEY")

	// Check if critical variables are missing
	if smtpHost == "" || smtpPort == "" || smtpUsername == "" || smtpPassword == "" {
		return fmt.Errorf("SMTP environment variables not set")
	}

	if reviewAPIKey == "" {
		return fmt.Errorf("REVIEW_FORM_KEY environment variable not set")
	}

	return nil
}

func connectDB() (*sql.DB, error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s",
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_NAME"))

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Println("Error opening database connection:", err)
		return nil, err
	}

	if err := db.Ping(); err != nil {
		log.Println("Error pinging database:", err)
		return nil, err
	}

	return db, nil
}

// Review represents a review in the database.
type Review struct {
	ID      int    `json:"id"`
	Company string `json:"company"`
	Name    string `json:"name"`
	Review  string `json:"review"`
}

// ReviewRequest represents the request body for adding a review.
type ReviewRequest struct {
	Company string `json:"company"`
	Name    string `json:"name"`
	Review  string `json:"review"`
	Key     string `json:"key"`
}

// ContactRequest represents the request body for the contact form.
type ContactRequest struct {
	Name    string `json:"name"`
	Email   string `json:"email"`
	Message string `json:"message"`
}

// Project represents a project in the database.
type Project struct {
	ID          int    `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Url         string `json:"url"`
	ActionLabel string `json:"actionLabel"`
}

func fetchReviews(db *sql.DB) ([]byte, error) {
	rows, err := db.Query("SELECT company, name, review FROM reviews")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var reviews []Review
	for rows.Next() {
		var review Review
		if err := rows.Scan(&review.Company, &review.Name, &review.Review); err != nil {
			return nil, err
		}
		reviews = append(reviews, review)
	}

	jsonData, err := json.Marshal(reviews)
	if err != nil {
		return nil, err
	}

	return jsonData, nil
}

func saveReview(review ReviewRequest) error {
	_, err := db.Exec("INSERT INTO reviews (company, name, review) VALUES (?, ?, ?)", review.Company, review.Name, review.Review)
	return err
}

func fetchProjects(db *sql.DB) ([]byte, error) {
	rows, err := db.Query("SELECT id, title, description, url FROM projects")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var projects []Project
	for rows.Next() {
		var project Project
		if err := rows.Scan(&project.ID, &project.Title, &project.Description, &project.Url); err != nil {
			return nil, err
		}

		// Determine the action label based on the URL
		if containsGitHub(project.Url) {
			project.ActionLabel = "Show Code"
		} else {
			project.ActionLabel = "Show Website"
		}

		projects = append(projects, project)
	}

	jsonData, err := json.Marshal(projects)
	if err != nil {
		return nil, err
	}

	return jsonData, nil
}

// containsGitHub checks if the URL contains "github.com"
func containsGitHub(url string) bool {
	return strings.Contains(url, "github.com")
}

func handleReviews(w http.ResponseWriter, r *http.Request) {
	// Only handle the GET request (assuming you are fetching reviews here)
	if r.Method == "GET" {
		jsonData, err := fetchReviews(db)
		if err != nil {
			log.Println("Error fetching reviews from database:", err)
			http.Error(w, "Failed to fetch reviews", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write(jsonData)
		return
	}

	// Handle other methods like POST here if needed
	http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
}

func saveReviewHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	// Handle POST request for /api/review
	var reviewRequest ReviewRequest
	if err := json.NewDecoder(r.Body).Decode(&reviewRequest); err != nil {
		log.Println("Error decoding request body:", err)
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	// Verify the key
	if reviewRequest.Key != os.Getenv("REVIEW_FORM_KEY") {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	if err := saveReview(reviewRequest); err != nil {
		log.Println("Error saving review:", err)
		http.Error(w, "Failed to save review", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Review submitted successfully"))
}

func contactHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var contactRequest struct {
		Name    string `json:"name"`
		Email   string `json:"email"`
		Message string `json:"message"`
	}

	if err := json.NewDecoder(r.Body).Decode(&contactRequest); err != nil {
		log.Println("Error decoding request body:", err)
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	adminEmail := os.Getenv("ADMIN_EMAIL") // Your email where messages go
	if adminEmail == "" {
		log.Println("❌ ADMIN_EMAIL is not set in environment variables")
		http.Error(w, "Email server misconfiguration", http.StatusInternalServerError)
		return
	}

	subject := "New Contact Form Submission"
	plainTextContent := fmt.Sprintf("Name: %s\nMessage: %s", contactRequest.Name, contactRequest.Message)
	htmlContent := fmt.Sprintf("<p>Name: %s</p><p>Message: %s</p>", contactRequest.Name, contactRequest.Message)

	err := sendEmailSMTP(contactRequest.Email, adminEmail, subject, plainTextContent, htmlContent)
	if err != nil {
		log.Println("❌ Failed to send contact request email:", err)
		http.Error(w, "Failed to send message", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Message sent successfully"))
}

func sendEmailSMTP(fromEmail, toEmail, subject, plainTextContent, htmlContent string) error {
	smtpHost := os.Getenv("SMTP_HOST")
	smtpPort := os.Getenv("SMTP_PORT")
	smtpUsername := os.Getenv("SMTP_USERNAME")
	smtpPassword := os.Getenv("SMTP_PASSWORD")

	if smtpHost == "" || smtpPort == "" || smtpUsername == "" || smtpPassword == "" {
		log.Println("❌ SMTP configuration missing!")
		return fmt.Errorf("SMTP configuration is missing, check your environment variables")
	}

	addr := smtpHost + ":" + smtpPort
	log.Println("🔄 Connecting to SMTP server at:", addr)

	// Create a TCP connection to the SMTP server
	conn, err := net.DialTimeout("tcp", addr, 10*time.Second)
	if err != nil {
		log.Println("❌ SMTP Connection Error:", err)
		return err
	}
	defer conn.Close()

	// Create a new SMTP client
	client, err := smtp.NewClient(conn, smtpHost)
	if err != nil {
		log.Println("❌ SMTP Client Error:", err)
		return err
	}
	defer client.Close()

	// Upgrade to TLS using STARTTLS
	tlsConfig := &tls.Config{
		ServerName: smtpHost,
	}

	if err = client.StartTLS(tlsConfig); err != nil {
		log.Println("❌ SMTP STARTTLS Error:", err)
		return err
	}

	// Authenticate
	auth := smtp.PlainAuth("", smtpUsername, smtpPassword, smtpHost)
	if err = client.Auth(auth); err != nil {
		log.Println("❌ SMTP Authentication Error:", err)
		return err
	}

	// Set the sender and recipient
	if err = client.Mail(smtpUsername); err != nil {
		return err
	}
	if err = client.Rcpt(toEmail); err != nil {
		return err
	}

	// Write email data
	w, err := client.Data()
	if err != nil {
		return err
	}
	message := fmt.Sprintf(
		"From: %s\r\nReply-To: %s\r\nTo: %s\r\nSubject: %s\r\nMIME-Version: 1.0\r\nContent-Type: text/html; charset=\"UTF-8\"\r\n\r\n%s",
		smtpUsername, fromEmail, toEmail, subject, htmlContent,
	)
	_, err = w.Write([]byte(message))
	if err != nil {
		return err
	}
	err = w.Close()
	if err != nil {
		return err
	}

	// Quit SMTP client
	client.Quit()

	log.Println("✅ Email sent successfully to", toEmail, "with Reply-To:", fromEmail)
	return nil
}

func handleProjects(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	jsonData, err := fetchProjects(db)
	if err != nil {
		log.Println("Error fetching projects from database:", err)
		http.Error(w, "Failed to fetch projects", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonData)
}

func main() {
	r := mux.NewRouter().StrictSlash(true)

	// Load environment variables
	if err := loadEnv(); err != nil {
		log.Fatal("Error loading .env file:", err)
	}

	// Connect to the database
	var err error
	db, err = connectDB()
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	r.Use(CORSMiddleware)

	// Define routes
	r.HandleFunc("/api/review", saveReviewHandler).Methods("POST")
	r.HandleFunc("/api/reviews", handleReviews).Methods("GET", "POST")
	r.HandleFunc("/api/contact", contactHandler).Methods("POST", "OPTIONS")
	r.HandleFunc("/api/projects", handleProjects).Methods("GET")

	// Start the server
	log.Fatal(http.ListenAndServe(":8080", r))
}
