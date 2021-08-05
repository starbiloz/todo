package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/gorilla/mux"
)

var index int
var tasks map[int]*Task = make(map[int]*Task)

type Task struct {
	Title string
	Done  bool
}

type NewTaskTodo struct {
	Task string `json:"task"`
}

func New(task string) {
	defer func() {
		index++
	}()

	tasks[index] = &Task{
		Title: task,
		Done:  false,
	}
}

func List() map[int]*Task {
	return tasks
}

func main() {
	r := mux.NewRouter()

	r.HandleFunc("/auth", func(rw http.ResponseWriter, r *http.Request) {
		mySigningKey := []byte("password")
		claims := &jwt.StandardClaims{
			ExpiresAt: time.Now().Add(2 * time.Minute).Unix(),
			Issuer:    "test",
		}

		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		ss, err := token.SignedString(mySigningKey)
		if err != nil {
			rw.WriteHeader(http.StatusInternalServerError)
			return
		}

		json.NewEncoder(rw).Encode(map[string]string{
			"token": ss,
		})
	})

	r.HandleFunc("/todos", func(rw http.ResponseWriter, r *http.Request) {
		tokenString := r.Header.Get("Authorization")

		tokenString = strings.ReplaceAll(tokenString, "Bearer ", "")

		fmt.Println(tokenString)

		mySigningKey := []byte("password")
		_, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
			}

			return mySigningKey, nil
		})

		if err != nil {
			rw.WriteHeader(http.StatusUnauthorized)
			return
		}

		defer r.Body.Close()
		var task NewTaskTodo
		if err := json.NewDecoder(r.Body).Decode(&task); err != nil {
			rw.WriteHeader(http.StatusInternalServerError)
			return
		}

		New(task.Task)
	}).Methods(http.MethodPut)

	r.HandleFunc("/todos/{index}", func(rw http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()

		vars := mux.Vars(r)
		index, err := strconv.Atoi(vars["index"])
		if err != nil {
			rw.WriteHeader(http.StatusBadRequest)
			return
		}

		tasks[index].Done = true
	}).Methods(http.MethodPut)

	r.HandleFunc("/todos", func(rw http.ResponseWriter, r *http.Request) {
		if err := json.NewEncoder(rw).Encode(tasks); err != nil {
			rw.WriteHeader(http.StatusInternalServerError)
			return
		}
	}).Methods(http.MethodGet)

	http.ListenAndServe(":9090", r)
}
