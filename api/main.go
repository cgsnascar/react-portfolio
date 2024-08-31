package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	_ "github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"
	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
)

var (
	db             *sql.DB
	contactFormKey string
	reviewAPIKey   string
)

func loadEnv() error {
	if err := godotenv.Load(); err != nil {
		return err
	}

	contactFormKey = os.Getenv("SENDGRID_API_KEY")
	reviewAPIKey = os.Getenv("REVIEW_FORM_KEY")

	if contactFormKey == "" || reviewAPIKey == "" {
		return fmt.Errorf("environment variables not set")
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
	Key     string `json:"key"`
}

// Project represents a project in the database.
type Project struct {
	ID          int    `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Url         string `json:"url"`
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
		projects = append(projects, project)
	}

	jsonData, err := json.Marshal(projects)
	if err != nil {
		return nil, err
	}

	return jsonData, nil
}

func handleReviews(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	jsonData, err := fetchReviews(db)
	if err != nil {
		log.Println("Error fetching reviews from database:", err)
		http.Error(w, "Failed to fetch reviews", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonData)
}

func saveReviewHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

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
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var contactRequest ContactRequest
	if err := json.NewDecoder(r.Body).Decode(&contactRequest); err != nil {
		log.Println("Error decoding request body:", err)
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	if contactRequest.Key != contactFormKey {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	err := sendEmail(contactRequest.Email, "Contact Form Submission", fmt.Sprintf("Name: %s\nMessage: %s", contactRequest.Name, contactRequest.Message))
	if err != nil {
		log.Println("Error sending email:", err)
		http.Error(w, "Failed to send message", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Message sent successfully"))
}

// sendEmail sends an email using SendGrid
func sendEmail(submitterEmail, subject, body string) error {
	apiKey := os.Getenv("SENDGRID_API_KEY")
	sg := sendgrid.NewSendClient(apiKey)

	// Use a verified email address for the "From" field
	from := mail.NewEmail("Claudio Skala", "cskala@cgsnascar.dev")

	// Set the "To" field to your email address
	to := mail.NewEmail("Claudio Skala", "cskala@cgsnascar.dev")

	// Set the "Reply-To" field to the submitter's email address
	replyTo := mail.NewEmail("Submitter", submitterEmail)

	message := mail.NewSingleEmail(from, subject, to, body, body)
	message.ReplyTo = replyTo // Set the reply-to address to the submitter's email

	response, err := sg.Send(message)
	if err != nil {
		return err
	}

	log.Printf("SendGrid Response: %d", response.StatusCode)
	log.Printf("Headers: %v", response.Headers)
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

	// Define routes
	http.HandleFunc("/api/reviews", handleReviews)    // GET request to fetch reviews
	http.HandleFunc("/api/review", saveReviewHandler) // POST request to submit a review
	http.HandleFunc("/api/contact", contactHandler)
	http.HandleFunc("/api/projects", handleProjects)

	log.Println("Server started at :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
