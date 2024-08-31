package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"net/smtp"
	"os"

	_ "github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"
)

func init() {
	// Load .env file
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file")
	}
}

// Function to connect to the database
func connectDB() (*sql.DB, error) {
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	dbName := os.Getenv("DB_NAME")
	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")

	dsn := user + ":" + password + "@tcp(" + host + ":" + port + ")/" + dbName
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	return db, nil
}

// Define the Review and ReviewRequest structs
type Review struct {
	ID      int    `json:"id"`
	Company string `json:"company"`
	Name    string `json:"name"`
	Review  string `json:"review"`
}

type ReviewRequest struct {
	Name    string `json:"name"`
	Company string `json:"company"`
	Review  string `json:"review"`
	Keyword string `json:"keyword"`
}

var db *sql.DB

// Function to fetch reviews from the database
func fetchReviews(db *sql.DB) ([]byte, error) {
	query := "SELECT id, company, name, review FROM reviews"
	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var reviews []Review
	for rows.Next() {
		var review Review
		if err := rows.Scan(&review.ID, &review.Company, &review.Name, &review.Review); err != nil {
			return nil, err
		}
		reviews = append(reviews, review)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return json.Marshal(reviews)
}

// Save Review Handler
func saveReviewHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	var reviewRequest ReviewRequest
	if err := json.NewDecoder(r.Body).Decode(&reviewRequest); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	expectedKeyword := os.Getenv("REVIEW_SUBMISSION_KEYWORD")
	if reviewRequest.Keyword != expectedKeyword {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	review := Review{
		Company: reviewRequest.Company,
		Name:    reviewRequest.Name,
		Review:  reviewRequest.Review,
	}

	if err := saveReview(review); err != nil {
		log.Println("Error saving review:", err)
		http.Error(w, "Failed to save review", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Review saved successfully"))
}

// Save Review Function
func saveReview(review Review) error {
	query := `INSERT INTO reviews (company, name, review) VALUES (?, ?, ?)`
	_, err := db.Exec(query, review.Company, review.Name, review.Review)
	return err
}

// Function to fetch projects from the database
func fetchProjects(db *sql.DB) ([]byte, error) {
	rows, err := db.Query("SELECT id, title, description, url FROM projects")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	type Project struct {
		ID          int    `json:"id"`
		Title       string `json:"title"`
		Description string `json:"description"`
		URL         string `json:"url"`
	}

	var projects []Project

	for rows.Next() {
		var project Project
		if err := rows.Scan(&project.ID, &project.Title, &project.Description, &project.URL); err != nil {
			return nil, err
		}
		projects = append(projects, project)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return json.Marshal(projects)
}

type ContactRequest struct {
	Name    string `json:"name"`
	Email   string `json:"email"`
	Message string `json:"message"`
	Key     string `json:"key"`
}

// contactHandler handles POST requests to /api/contact
// contactHandler handles POST requests to /api/contact
func contactHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != "POST" {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	var contactRequest struct {
		Name    string `json:"name"`
		Email   string `json:"email"`
		Message string `json:"message"`
	}

	err := json.NewDecoder(r.Body).Decode(&contactRequest)
	if err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	err = sendEmail(contactRequest.Email, "Contact Form Submission", contactRequest.Message)
	if err != nil {
		http.Error(w, "Failed to send email", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Message sent successfully"))
}

func sendEmail(senderEmail, subject, body string) error {
	from := os.Getenv("EMAIL")
	apiKey := os.Getenv("SENDGRID_API_KEY") // Use your SendGrid API Key here

	to := "cskala@cgsnascar.dev" // Replace with your email address

	msg := "From: " + from + "\n" +
		"To: " + to + "\n" +
		"Reply-To: " + senderEmail + "\n" +
		"Subject: " + subject + "\n\n" +
		body

	auth := smtp.PlainAuth("", "apikey", apiKey, "smtp.sendgrid.net")
	err := smtp.SendMail("smtp.sendgrid.net:587", auth, from, []string{to}, []byte(msg))

	if err != nil {
		log.Printf("Error sending email: %v", err)
		return err
	}

	return nil
}

func main() {
	// Connect to database
	db, dbErr := connectDB()
	if dbErr != nil {
		log.Fatal("Failed to connect to database:", dbErr)
	}
	defer db.Close()

	// Review Endpoint
	http.HandleFunc("/api/reviews", func(w http.ResponseWriter, r *http.Request) {
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
	})

	// Projects Endpoint
	http.HandleFunc("/api/projects", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
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
	})

	// Contact Endpoint
	http.HandleFunc("/api/contact", contactHandler)

	// Save Review Endpoint
	http.HandleFunc("/api/saveReview", saveReviewHandler)

	log.Println("Server started at :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
