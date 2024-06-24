package handlers

import (
    "context"
    "html/template"
    "log"
    "net/http"

    pb "shubam/proto"
)

var appointmentClient pb.HospitalServiceClient

func AppointmentHandler(w http.ResponseWriter, r *http.Request) {
    log.Println("AppointmentHandler called")  // Debug log

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

    log.Println("Cookies:", userIDCookie, userEmailCookie)  // Debug log

    if r.Method == http.MethodGet {
        tmpl, err := template.ParseFiles("Static/appointment.html")
        if err != nil {
            log.Printf("Error parsing appointment template: %v\n", err)
            http.Error(w, "Error loading appointment page", http.StatusInternalServerError)
            return
        }

        data := struct {
            UserID    string
            UserEmail string
        }{
            UserID:    userIDCookie.Value,
            UserEmail: userEmailCookie.Value,
        }

        tmpl.Execute(w, data)
        return
    }

    if r.Method == http.MethodPost {
        err := r.ParseForm()
        if err != nil {
            http.Error(w, "Parse form error", http.StatusInternalServerError)
            return
        }

        doctorName := r.FormValue("doctor")
        date := r.FormValue("date")
        time := r.FormValue("time")

        if date == "" {
            http.Error(w, "Date is required", http.StatusBadRequest)
            return
        }

        if time == "" {
            http.Error(w, "Time is required", http.StatusBadRequest)
            return
        }

        // Add debug log for form values
        log.Println("Form values:", doctorName, date, time)

        req := &pb.AppointmentRequest{
            DoctorName: doctorName,
            UserId:     userIDCookie.Value,
            Email:      userEmailCookie.Value,
            Date:       date,
            Time:       time,
        }

        log.Println("AppointmentClient:", appointmentClient)  // Debug log

        resp, err := appointmentClient.Appointment(context.Background(), req)
        if err != nil {
            log.Printf("Failed to create appointment: %v", err)
            http.Error(w, "Failed to create appointment", http.StatusInternalServerError)
            return
        }

        log.Println(resp.Message)
        http.Redirect(w, r, "/service", http.StatusSeeOther)
    }
}
