package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	jwt "github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/mux"
	"github.com/rs/cors"
)

type Server struct {
	listenAddr string
	store      Storage
}

type apiFunc func(w http.ResponseWriter, r *http.Request) error

func NewApiServer(addr string, store Storage) *Server {
	return &Server{
		listenAddr: addr,
		store:      store,
	}
}

func (s *Server) Run() {
	router := mux.NewRouter()
	router.HandleFunc("/account", makeApiHandleFunc(s.handleGetAccount)).Methods("GET")
	router.HandleFunc("/account", makeApiHandleFunc(s.handleCreateAccount)).Methods("POST")
	router.HandleFunc("/account/{id}", withJWTAuth(makeApiHandleFunc(s.handleGetAccountByID), s.store)).Methods("GET")
	router.HandleFunc("/account/{id}", withJWTAuth(makeApiHandleFunc(s.handleDeleteAccount), s.store)).Methods("DELETE")
	router.HandleFunc("/login", makeApiHandleFunc(s.handleLogin)).Methods("POST")
	router.HandleFunc("/update/{id}", withJWTAuth(makeApiHandleFunc(s.handleUpdateAccount), s.store)).Methods("POST")
	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"http://localhost:3000"},
		AllowCredentials: true,
	})
	handler := c.Handler(router)
	fmt.Println("listening on port", s.listenAddr)
	http.ListenAndServe(s.listenAddr, handler)

}

func (s *Server) handleGetAccount(w http.ResponseWriter, r *http.Request) error {
	accounts, err := s.store.GetAccounts()
	if err != nil {
		return err
	}

	WriteJSON(w, http.StatusOK, accounts)
	return nil
}

func getID(r *http.Request) (int, error) {
	idStr := mux.Vars(r)["id"]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		return 0, fmt.Errorf("invalid id given")
	}
	return id, nil
}

func (s *Server) handleLogin(w http.ResponseWriter, r *http.Request) error {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return err
	}
	acc, err := s.store.GetAccountByNumber(int(req.Number))
	if err != nil {
		return err
	}

	if !acc.ValidPassword(req.Password) {
		return fmt.Errorf("authentication failed")
	}

	token, err := createJWT(acc)
	if err != nil {
		return err
	}

	resp := LoginResponse{
		Token:  token,
		Number: acc.Number,
	}

	return WriteJSON(w, http.StatusOK, resp)
}

func (s *Server) handleGetAccountByID(w http.ResponseWriter, r *http.Request) error {
	id, _ := getID(r)
	account, err := s.store.GetAccountByID(id)
	if err != nil {
		return err
	}

	token, err := createJWT(account)
	if err != nil {
		return err
	}
	fmt.Println("jwt token: ", token)

	return WriteJSON(w, http.StatusOK, account)
}

func (s *Server) handleCreateAccount(w http.ResponseWriter, r *http.Request) error {
	CreateAccountReq := new(CreateAccountRequest)
	//decode request body into above^
	if err := json.NewDecoder(r.Body).Decode(CreateAccountReq); err != nil {
		return err
	}
	//create account object (returns obj)
	account, err := NewAccount(CreateAccountReq.FirstName, CreateAccountReq.LastName, CreateAccountReq.Password)
	if err != nil {
		return err
	}

	if err := s.store.CreateAccount(account); err != nil {
		return err
	}

	return WriteJSON(w, http.StatusOK, account)
}

func (s *Server) handleDeleteAccount(w http.ResponseWriter, r *http.Request) error {
	vars := mux.Vars(r)
	id, err := strconv.ParseInt(vars["id"], 10, 0)
	if err != nil {
		http.Error(w, "unable to convert string to int", http.StatusInternalServerError)
	}

	err = s.store.DeleteAccount(int(id))
	if err != nil {
		return err
	}
	return nil
}

func (s *Server) handleTransfer(w http.ResponseWriter, r *http.Request) error {
	transferReq := new(TransferRequest)
	if err := json.NewDecoder(r.Body).Decode(transferReq); err != nil {
		return err
	}
	defer r.Body.Close()
	return WriteJSON(w, http.StatusOK, transferReq)
}

func (s *Server) handleUpdateAccount(w http.ResponseWriter, r *http.Request) error {
	updateReq := new(UpdatePassword)
	err := json.NewDecoder(r.Body).Decode(updateReq)
	if err != nil {
		log.Fatal(err)
	}
	id, err := getID(r)
	if err != nil {
		return err
	}

	err = s.store.UpdatePassword(id, updateReq.NewPass)
	if err != nil {
		return err
	}
	return nil
}

func withJWTAuth(handlerFunc http.HandlerFunc, s Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("calling jwt auth middleware")
		tokenString := r.Header.Get("jwt-token")
		token, err := validateJWT(tokenString)
		if err != nil {
			WriteJSON(w, http.StatusForbidden, "invalid token")
			return
		}

		if !token.Valid {
			WriteJSON(w, http.StatusForbidden, "invalid token")
			return
		}

		userID, err := getID(r)
		if err != nil {
			return
		}
		account, err := s.GetAccountByID(userID)
		if err != nil {
			return
		}

		claims := token.Claims.(jwt.MapClaims)
		if account.Number != int64((claims["accountNumber"]).(float64)) {
			WriteJSON(w, http.StatusForbidden, "permission denied")
			return
		}

		handlerFunc(w, r)
	}

}

func createJWT(account *Account) (string, error) {
	claims := &jwt.MapClaims{
		"expiresAt":     15000,
		"accountNumber": account.Number,
	}
	jwtToken := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	secret := os.Getenv("JWT_SECRET")
	tokenString, err := jwtToken.SignedString([]byte(secret))
	fmt.Println("token: ", tokenString)
	if err != nil {
		return "", err
	}

	return tokenString, nil

}

func validateJWT(tokenString string) (*jwt.Token, error) {
	return jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {

		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		// hmacSampleSecret is a []byte containing your secret, e.g. []byte("my_secret_key")
		return []byte(os.Getenv("JWT_SECRET")), nil
	})
}

func WriteJSON(w http.ResponseWriter, status int, v any) error {
	w.WriteHeader(status)
	w.Header().Set("Content-Type", "application/json")
	return json.NewEncoder(w).Encode(v)
}

func makeApiHandleFunc(f apiFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := f(w, r); err != nil {
			WriteJSON(w, http.StatusBadRequest, err.Error())
		}
	}
}
