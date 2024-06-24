package handlers;

import (


	"html/template"
	"log"
	"net/http"

	_ "github.com/lib/pq"
)
type PageData struct {
	UserID   string
	UserEmail string
}


func PharmacyHandler(w http.ResponseWriter, r *http.Request) {
	userIDCookie, err := r.Cookie("userID")
	if err != nil {
		http.Error(w, "Missing user ID", http.StatusBadRequest)
		return
	}
	userEmailCookie, err := r.Cookie("userEmail")
	if err != nil {
		http.Error(w, "Missing user email", http.StatusBadRequest)
		return
	}

	data := PageData{
		UserID:   userIDCookie.Value,
		UserEmail: userEmailCookie.Value,
	}

	tmpl, err := template.ParseFiles("Static/appointment.html")
if err != nil {
    log.Printf("Error parsing pharmacy template: %v\n", err)
    http.Error(w, "Error loading pharmacy page", http.StatusInternalServerError)
    return
}
log.Println("Pharmacy called")

	tmpl.Execute(w, data)
}