package main

import (
	"html/template"
	"net/http"
	"strconv"
)

type ComplaintView struct {
	ID          int
	StudentName string
	RollNo      string
	RoomNo      string
	Title       string
	Description string
	Status      string
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	username := r.FormValue("username")
	password := r.FormValue("password")
	role := r.FormValue("role")

	if role == "admin" {
		var id int
		err := db.QueryRow(
			"SELECT id FROM admins WHERE username=? AND password=?",
			username, password,
		).Scan(&id)

		if err == nil {
			http.SetCookie(w, &http.Cookie{
				Name:  "admin",
				Value: "true",
				Path:  "/",
			})

			http.Redirect(w, r, "/admin", http.StatusSeeOther)
			return
		}

	}

	if role == "student" {
		var id int
		err := db.QueryRow(
			"SELECT id FROM students WHERE username=? AND password=?",
			username, password,
		).Scan(&id)

		if err == nil {
			http.SetCookie(w, &http.Cookie{
				Name:  "student_id",
				Value: strconv.Itoa(id),
				Path:  "/",
			})

			http.Redirect(w, r, "/student", http.StatusSeeOther)
			return
		}
	}

	w.Write([]byte("Invalid login credentials"))
}

func adminDashboard(w http.ResponseWriter, r *http.Request) {
	// STEP D2: Admin access check
	_, err := r.Cookie("admin")
	if err != nil {
		// No admin cookie → not logged in as admin
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	// ---- EXISTING CODE CONTINUES BELOW ----

	rows, err := db.Query(`
		SELECT c.id, s.name, s.roll_no, s.room_no,
		       c.title, c.description, c.status
		FROM complaints c
		JOIN students s ON c.student_id = s.id
		ORDER BY c.created_at DESC
	`)
	if err != nil {
		w.Write([]byte("Error loading complaints"))
		return
	}
	defer rows.Close()

	var complaints []ComplaintView

	for rows.Next() {
		var c ComplaintView
		rows.Scan(
			&c.ID,
			&c.StudentName,
			&c.RollNo,
			&c.RoomNo,
			&c.Title,
			&c.Description,
			&c.Status,
		)
		complaints = append(complaints, c)
	}

	tmpl := template.Must(template.ParseFiles("templates/admin.html"))
	tmpl.Execute(w, complaints)
}

func studentDashboard(w http.ResponseWriter, r *http.Request) {
	_, err := r.Cookie("student_id")
	cookie, err := r.Cookie("student_id")
	if _, err := r.Cookie("admin"); err == nil {
		http.Redirect(w, r, "/admin", http.StatusSeeOther)
		return
	}

	if err != nil {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	rows, err := db.Query(`
		SELECT id, title, description, status
		FROM complaints
		WHERE student_id = ?
		ORDER BY created_at DESC
	`, cookie.Value)

	if err != nil {
		w.Write([]byte("Error loading complaints"))
		return
	}
	defer rows.Close()

	type StudentComplaint struct {
		ID          int
		Title       string
		Description string
		Status      string
	}

	var complaints []StudentComplaint

	for rows.Next() {
		var c StudentComplaint
		rows.Scan(&c.ID, &c.Title, &c.Description, &c.Status)
		complaints = append(complaints, c)
	}

	tmpl := template.Must(template.ParseFiles("templates/student.html"))
	tmpl.Execute(w, complaints)
}

func addStudent(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/admin", http.StatusSeeOther)
		return
	}

	name := r.FormValue("name")
	roll := r.FormValue("roll")
	room := r.FormValue("room")
	username := r.FormValue("username")
	password := r.FormValue("password")

	_, err := db.Exec(
		`INSERT INTO students (name, roll_no, room_no, username, password)
		 VALUES (?, ?, ?, ?, ?)`,
		name, roll, room, username, password,
	)

	if err != nil {
		w.Write([]byte("Error adding student"))
		return
	}

	http.Redirect(w, r, "/admin", http.StatusSeeOther)
}
func addComplaint(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/student", http.StatusSeeOther)
		return
	}

	cookie, err := r.Cookie("student_id")
	if err != nil {
		w.Write([]byte("Student not logged in"))
		return
	}

	studentID := cookie.Value
	title := r.FormValue("title")
	description := r.FormValue("description")

	_, err = db.Exec(
		`INSERT INTO complaints (student_id, title, description)
		 VALUES (?, ?, ?)`,
		studentID, title, description,
	)

	if err != nil {
		w.Write([]byte("Error submitting complaint"))
		return
	}

	http.Redirect(w, r, "/student", http.StatusSeeOther)
}
func logout(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:   "student_id",
		Value:  "",
		Path:   "/",
		MaxAge: -1,
	})

	http.SetCookie(w, &http.Cookie{
		Name:   "admin",
		Value:  "",
		Path:   "/",
		MaxAge: -1,
	})

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func updateComplaintStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/admin", http.StatusSeeOther)
		return
	}

	id := r.FormValue("id")
	status := r.FormValue("status")

	db.Exec(
		"UPDATE complaints SET status=? WHERE id=?",
		status, id,
	)

	http.Redirect(w, r, "/admin", http.StatusSeeOther)
}
