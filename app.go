package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/dgrijalva/jwt-go"
	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
)

type App struct {
	Router *mux.Router
	DB     *sql.DB
}

// Initialize method
func (app *App) Initialize(DBUser, DBPassword, DBName string) error {

	// Initialize database connection
	var err error
	connectionString := fmt.Sprintf("%v:%v@tcp(192.168.1.13:3306)/%v", DBUser, DBPassword, DBName)
	app.DB, err = sql.Open("mysql", connectionString)
	if err != nil {
		return err
	}

	// Initialize routes
	app.Router = mux.NewRouter().StrictSlash(true)
	app.handleRoutes()
	return nil
}

// LoginHandler handles the login requests
func (app *App) LoginHandler(w http.ResponseWriter, r *http.Request) {
	var user User
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		errorResponse(w, http.StatusBadRequest, err.Error())
		return
	}
	defer r.Body.Close()

	if Authenticate(user.Username, user.Password) {
		token, err := GenerateToken(user.Username)
		if err != nil {
			errorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}
		response := json.NewEncoder(w).Encode(map[string]string{"token": token})

		sendResponse(w, http.StatusOK, response)
	} else {
		sendResponse(w, http.StatusUnauthorized, map[string]string{"result": "access denied"})
	}
}

// Run method
func (app *App) Run(address string) {
	fmt.Println("Listening on port 8080")
	fmt.Println("Press Ctrl+C to quit")
	log.Fatal(http.ListenAndServe(address, app.Router))
}

// handleRoutes method
func sendResponse(w http.ResponseWriter, statusCode int, payload interface{}) {
	response, _ := json.Marshal(payload)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	w.Write(response)
}

// errorResponse method
func errorResponse(w http.ResponseWriter, statusCode int, err string) {
	error_message := map[string]string{"error": err}
	sendResponse(w, statusCode, error_message)
}

// handleRoutes method
func (app *App) getProducts(w http.ResponseWriter, r *http.Request) {
	products, err := getProducts(app.DB)
	if err != nil {
		errorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}
	sendResponse(w, http.StatusOK, products)
}

// getProductByID method
func (app *App) getProductByID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi((vars["id"]))
	if err != nil {
		errorResponse(w, http.StatusBadRequest, "Invalid product ID")
		return
	}
	p := product{ID: id}
	err = p.getProductByID(app.DB)
	if err != nil {
		switch err {
		case sql.ErrNoRows:
			errorResponse(w, http.StatusNotFound, "Product not found")
		default:
			errorResponse(w, http.StatusInternalServerError, err.Error())
		}
		return
	}
	sendResponse(w, http.StatusOK, p)
}

// addProduct method
func (app *App) addProduct(w http.ResponseWriter, r *http.Request) {
	var p product
	error := json.NewDecoder(r.Body).Decode(&p)
	if error != nil {
		errorResponse(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	defer r.Body.Close()
	err := p.addProduct(app.DB)
	if err != nil {
		errorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}
	sendResponse(w, http.StatusCreated, p)
}

// updateProduct method
func (app *App) updateProduct(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	key, err := strconv.Atoi((vars["id"]))
	if err != nil {
		errorResponse(w, http.StatusBadRequest, "Invalid product ID")
		return
	}
	var p product
	err = json.NewDecoder(r.Body).Decode(&p)
	if err != nil {
		errorResponse(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	defer r.Body.Close()
	p.ID = key
	err = p.updateProduct(app.DB)
	if err != nil {
		errorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}
	sendResponse(w, http.StatusOK, p)
}

// deleteProductByID method
func (app *App) deleteProductByID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	key, err := strconv.Atoi((vars["id"]))
	if err != nil {
		errorResponse(w, http.StatusBadRequest, "Invalid product ID")
		return
	}
	p := product{ID: key}
	err = p.deleteProductByID(app.DB)
	if err != nil {
		errorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}
	sendResponse(w, http.StatusOK, map[string]string{"result": "success"})
}

// handleRoutes method
func (app *App) handleRoutes() {
	app.Router.HandleFunc("/login", app.LoginHandler).Methods("POST")
	app.Router.HandleFunc("/products", jwtMiddleware(app.getProducts)).Methods("GET")
	app.Router.HandleFunc("/product", jwtMiddleware(app.addProduct)).Methods("POST")
	app.Router.HandleFunc("/product/{id}", jwtMiddleware(app.updateProduct)).Methods("PUT")
	app.Router.HandleFunc("/product/{id}", jwtMiddleware(app.getProductByID)).Methods("GET")
	app.Router.HandleFunc("/product/{id}", jwtMiddleware(app.deleteProductByID)).Methods("DELETE")
}

// jwtMiddleware is a middleware function to validate the JWT token
func jwtMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tokenString := r.Header.Get("Authorization")
		if tokenString == "" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
			return secretKey, nil
		})

		if err != nil || !token.Valid {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		next(w, r)
	}
}
