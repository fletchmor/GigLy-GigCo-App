package middleware

import (
	"encoding/json"
	"log"
	"net/http"
	"regexp"
	"strings"
	"text/template"
	"time"
)

var validEmails = make(map[string]time.Time)

var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)

// Logger Function
func Logger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		log.Printf("Started %s %s", r.Method, r.URL.Path)

		next.ServeHTTP(w, r)

		log.Printf("Completed %s in %v", r.URL.Path, time.Since(start))
	})
}

func IsValidEmailDomain(email string) bool {
	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		return false
	}

	domain := parts[1]

	// Check For Common Invalid Domains
	invalidDomains := []string{
		"test.com",
		"example.com",
		"fake.com",
		"invalid.com",
	}

	for _, invalid := range invalidDomains {
		if domain == invalid {
			return false
		}
	}

	return true
}

func ServeEmailForm(w http.ResponseWriter, r *http.Request) {
	// Set content type first
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	
	tmpl, err := template.ParseFiles("templates/email.html")
	if err != nil {
		// Fallback to simple HTML if template not found
		log.Printf("Template error, using fallback HTML: %v", err)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>GigCo - Email Form</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 40px; }
        form { max-width: 400px; }
        input, button { padding: 10px; margin: 5px 0; width: 100%; }
        button { background: #007bff; color: white; border: none; cursor: pointer; }
    </style>
</head>
<body>
    <h1>Email Subscription</h1>
    <p>Stay updated with our latest news and offers.</p>
    <form action="/submit-email" method="POST">
        <label for="email">Email Address:</label>
        <input type="email" id="email" name="email" required placeholder="your@email.com">
        <button type="submit">Subscribe</button>
    </form>
</body>
</html>`))
		return
	}

	// Template found, execute it
	err = tmpl.Execute(w, nil)
	if err != nil {
		log.Printf("Template execution error, using fallback: %v", err)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`<!DOCTYPE html><html><head><title>Email Form</title></head><body><h1>Email Form</h1><p>Template execution failed.</p></body></html>`))
	}
}

func HandleEmailSubmission(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var data struct {
		Email string `json:"email"`
	}

	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	email := strings.ToLower(strings.TrimSpace(data.Email))
	if !emailRegex.MatchString(email) {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "Please enter a valid email address",
		})
		return
	}

	if !IsValidEmailDomain(email) {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "Please enter a real email address",
		})
		return
	}

	validEmails[email] = time.Now()

	log.Printf("Valid email registered: %s", email)

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Email registered successfully",
	})
}
