package server

import (
	"encoding/json"
	db "feed-chatte-backend/internal/database"
	"feed-chatte-backend/internal/database/models"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/google/uuid"
	"log"
	"math/rand"
	"net/http"
	"time"
)

func (s *Server) RegisterRoutes() http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.Logger)

	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"https://*", "http://*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// TODO: Add db scripts for creating tables
	// TODO(later): Limit the amount of users/events that can be created (since it's just for personal use)

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("Welcome to Feed Chatte Backend"))
	})

	r.Get("/users", s.GetUsers)

	r.Post("/users", s.PostUsers)

	r.Get("/events", s.GetEvents)

	r.Post("/events", s.PostEvents)

	r.Delete("/events/{id}", s.DeleteEvents)

	r.Get("/chatte-message", s.GetChatteMessage)

	r.Get("/health", s.GetHealth)

	return r
}

func (s *Server) GetUsers(w http.ResponseWriter, r *http.Request) {
	query := "SELECT * FROM users"
	if nameQueryParam := r.URL.Query().Get("name"); nameQueryParam != "" {
		query = fmt.Sprintf("SELECT * FROM users WHERE name LIKE '%%%s%%' LIMIT 1", nameQueryParam)
	}

	users, err := db.FindAll[models.User](s.db, query)
	if err != nil {
		log.Println("Error fetching users. Err: %v", err)
		http.Error(w, "Failed to fetch users", http.StatusInternalServerError)
		return
	}

	jsonResp, err := json.Marshal(users)
	if err != nil {
		log.Println("Error handling JSON marshal. Err: %v", err)
		http.Error(w, "Failed to marshal response", http.StatusInternalServerError)
		return
	}
	_, _ = w.Write(jsonResp)
}

type PostUsersRequest struct {
	Name string
}

func (s *Server) PostUsers(w http.ResponseWriter, r *http.Request) {
	var reqBody PostUsersRequest
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		log.Printf("Error decoding reqBody. Err: %v\n", err)
		http.Error(w, "Failed to decode reqBody", http.StatusBadRequest)
		return
	}

	users, err := db.FindAll[models.User](s.db, fmt.Sprintf("SELECT * FROM users WHERE name LIKE '%%%s%%' LIMIT 1", reqBody.Name))
	if err != nil {
		log.Printf("Error fetching user. Err: %v", err)
		http.Error(w, "Failed to decode reqBody", http.StatusBadRequest)
		return
	}
	if len(users) > 0 {
		http.Error(w, "User already exists", http.StatusBadRequest)
		return
	}

	id := uuid.New()
	name := reqBody.Name
	_, err = s.db.Exec(fmt.Sprintf("INSERT INTO users (id, name) VALUES ('%s', '%s')", id, name))
	if err != nil {
		log.Printf("Error inserting user. Err: %v", err)
		http.Error(w, "Failed to insert user", http.StatusInternalServerError)
		return
	}

	resp := map[string]string{"id": id.String(), "name": name}
	jsonResp, _ := json.Marshal(resp)
	w.WriteHeader(http.StatusCreated)
	_, _ = w.Write(jsonResp)
}

func (s *Server) GetEvents(w http.ResponseWriter, r *http.Request) {
	query := "SELECT * FROM events ORDER BY time ASC"
	if todayQueryParam := r.URL.Query().Get("today"); todayQueryParam == "true" {
		query = "SELECT * FROM events WHERE DATE(time) = CURDATE() ORDER BY time ASC"
	}

	events, err := db.FindAll[models.Event](s.db, query)
	if err != nil {
		log.Println("Error fetching events. Err: %v", err)
		http.Error(w, "Failed to fetch events", http.StatusInternalServerError)
		return
	}

	jsonResp, err := json.Marshal(events)
	if err != nil {
		log.Println("Error handling JSON marshal. Err: %v", err)
		http.Error(w, "Failed to marshal response", http.StatusInternalServerError)
		return
	}
	_, _ = w.Write(jsonResp)
}

var messages = []string{
	"%s fed that fat boy",
	"%s fed chungus",
	"%s has fed chonk",
}

var punctuations = []string{
	"!",
	".",
}

type PostEventsRequest struct {
	Name string
}

func (s *Server) PostEvents(w http.ResponseWriter, r *http.Request) {
	var reqBody PostEventsRequest
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		log.Printf("Error decoding reqBody reqBody. Err: %v\n", err)
		http.Error(w, "Failed to decode reqBody reqBody", http.StatusBadRequest)
		return
	}

	users, err := db.FindAll[models.User](s.db, fmt.Sprintf("SELECT * FROM users WHERE name LIKE '%%%s%%' LIMIT 1", reqBody.Name))
	if err != nil {
		log.Printf("Error fetching user. Err: %v", err)
		http.Error(w, "Failed to decode reqBody reqBody", http.StatusBadRequest)
		return
	}
	userId := users[0].Id

	messageTemplate := fmt.Sprintf(messages[rand.Intn(len(messages))], reqBody.Name)
	punctuation := punctuations[rand.Intn(len(punctuations))]
	message := messageTemplate + punctuation

	id := uuid.New()
	_, err = s.db.Exec(fmt.Sprintf("INSERT INTO events (id, user_id, message) VALUES ('%s', '%s', '%s')", id, userId, message))
	if err != nil {
		log.Printf("Error inserting event. Err: %v", err)
		http.Error(w, "Failed to insert event", http.StatusInternalServerError)
		return
	}

	resp := map[string]string{"id": id.String()}
	jsonResp, _ := json.Marshal(resp)
	w.WriteHeader(http.StatusCreated)
	_, _ = w.Write(jsonResp)

}

func (s *Server) DeleteEvents(w http.ResponseWriter, r *http.Request) {
	eventId := chi.URLParam(r, "id")
	events, err := db.FindAll[models.Event](s.db, fmt.Sprintf("SELECT * FROM events WHERE id = '%s'", eventId))
	if err != nil {
		log.Printf("Error fetching event. Err: %v", err)
		http.Error(w, "Failed to fetch event", http.StatusInternalServerError)
		return
	}
	if len(events) == 0 {
		http.Error(w, "Event does not exist", http.StatusNotFound)
		return
	}

	if _, err = s.db.Exec(fmt.Sprintf("DELETE FROM events WHERE id = '%s'", eventId)); err != nil {
		log.Printf("Error deleting event. Err: %v", err)
		http.Error(w, "Failed to delete event", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

var chatteMessages = map[string]string{
	"starving":          "Feed me, I'm starving!!! 😱🍔",
	"hungry":            "Hey, feed me! >:|",
	"slightlySatisfied": "My belly is satisfied, for now...",
	"satisfied":         "My belly is satisfied",
	"full":              "Ah! That was a good meal! 😋",
}

func (s *Server) GetChatteMessage(w http.ResponseWriter, r *http.Request) {
	events, err := db.FindAll[models.Event](s.db, fmt.Sprintf("SELECT * FROM events ORDER BY time DESC LIMIT 1"))
	if err != nil {
		log.Printf("Error fetching events. Err: %v", err)
		http.Error(w, "Failed to fetch events", http.StatusInternalServerError)
		return
	}

	var response struct {
		Message string `json:"message"`
		Status  string `json:"status"`
	}

	switch timeLastFed := events[0].Time; {
	case timeLastFed.Before(time.Now().Add(2 * time.Hour)):
		response.Message = chatteMessages["full"]
		response.Status = "full"
	case timeLastFed.Before(time.Now().Add(3 * time.Hour)):
		response.Message = chatteMessages["satisfied"]
		response.Status = "satisfied"
	case timeLastFed.Before(time.Now().Add(4 * time.Hour)):
		response.Message = chatteMessages["slightlySatisfied"]
		response.Status = "slightlySatisfied"
	case timeLastFed.Before(time.Now().Add(6 * time.Hour)):
		response.Message = chatteMessages["hungry"]
		response.Status = "hungry"
	default:
		response.Message = chatteMessages["starving"]
		response.Status = "starving"
	}

	jsonResp, err := json.Marshal(response)
	if err != nil {
		log.Printf("Error handling JSON marshal. Err: %v", err)
		http.Error(w, "Failed to marshal response", http.StatusInternalServerError)
		return
	}
	_, _ = w.Write(jsonResp)
}

func (s *Server) GetHealth(w http.ResponseWriter, r *http.Request) {
	jsonResp, _ := json.Marshal(s.DBHealth())
	_, _ = w.Write(jsonResp)
}
