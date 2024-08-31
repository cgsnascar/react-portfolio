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

// Function to fetch reviews from the database
func fetchReviews(db *sql.DB) ([]byte, error) {
	rows, err := db.Query("SELECT id, company, name, review FROM reviews")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	type Review struct {
		ID      int    `json:"id"`
		Company string `json:"company"`
		Name    string `json:"name"`
		Review  string `json:"review"`
	}

	var reviews []Review

	for rows.Next() {
		var review Review
		if err := rows.Scan(&review.ID, &review.Company, &review.Name, &review.Review); err != nil {
			return nil, err
		}
		reviews = append(reviews, review)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return json.Marshal(reviews)
}

func saveReview(db *sql.DB, name, company, review string) error {
	_, err := db.Exec("INSERT INTO reviews (name, company, review) VALUES (?, ?, ?)", name, company, review)
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
}

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

// sendEmail sends an email using the SMTP server

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
	// Load environment variables from .env file
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file:", err)
	}

	// Connect to the database
	db, dbErr := connectDB()
	if dbErr != nil {
		log.Fatal("Failed to connect to database:", dbErr)
	}
	defer db.Close()

	// Endpoint to get reviews
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

	// Endpoint to get projects
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

	// Endpoint to submit a new review
	http.HandleFunc("/api/submit-review", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		if r.Method == "POST" {
			// Extract the key and review from the request
			key := r.FormValue("key")
			name := r.FormValue("name")
			company := r.FormValue("company")
			review := r.FormValue("review")

			// Validate the key
			if key != "Testing123" { // Replace "your-secret-key" with your actual key
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			// Save the review to the database
			err := saveReview(db, name, company, review)
			if err != nil {
				log.Println("Error saving review to database:", err)
				http.Error(w, "Failed to save review", http.StatusInternalServerError)
				return
			}

			w.WriteHeader(http.StatusOK)
			return
		}

		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	})

	// Endpoint to handle contact requests
	http.HandleFunc("/api/contact", contactHandler)

	log.Println("Server started at :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
