package main

import (
	"fmt"
	"log"
	"net/http"
)

func main() {

	fileServer := http.FileServer(http.Dir("./Register_User"))
	http.Handle("/", fileServer)

	http.HandleFunc("/register", loginHandler)
	http.HandleFunc("/submit", submitHandler)
	fmt.Println("Server started at http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	err := r.ParseForm()
	if err != nil {
		fmt.Fprintf(w, "Parse form error : %v", err)
		return
	}
}

func submitHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	email := r.FormValue("email")
	password := r.FormValue("password")

	// For simplicity, we are just printing the received data.
	// In a real application, you would validate and store this data securely.
	fmt.Fprintf(w, "Email: %s\n", email)
	fmt.Fprintf(w, "Password: %s\n", password)
}
