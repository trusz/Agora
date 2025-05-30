package server

import (
	"context"
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"agora/src/db"
	"agora/src/log"
	"agora/src/post"
	"agora/src/post/comment"
	"agora/src/user"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"golang.org/x/oauth2"
)

// MSGraphUser represents a user object returned by Microsoft Graph API
type MSGraphUser struct {
	ODataContext      string   `json:"@odata.context"`
	BusinessPhones    []string `json:"businessPhones"`
	DisplayName       string   `json:"displayName"`
	GivenName         string   `json:"givenName"`
	JobTitle          *string  `json:"jobTitle"`
	Mail              string   `json:"mail"`
	MobilePhone       *string  `json:"mobilePhone"`
	OfficeLocation    *string  `json:"officeLocation"`
	PreferredLanguage string   `json:"preferredLanguage"`
	Surname           string   `json:"surname"`
	UserPrincipalName string   `json:"userPrincipalName"`
	ID                string   `json:"id"`
}

// Server provides an http server wrap around services
type Server struct {
	host    string
	port    string
	stop    chan os.Signal
	stopped chan struct{}
	server  *http.Server
}

// NewServer creates a new Server
// Basic Usage:
// srv := new Server("0.0.0.0","8080")
// srv.Start()
// srv.WaitTilRunning()
func NewServer(
	host string,
	port string,
) (s *Server) {

	return &Server{
		host:    host,
		port:    port,
		stop:    make(chan os.Signal, 1),
		stopped: make(chan struct{}, 1),
	}
}

//go:embed static/*
var staticFiles embed.FS

// Start starts the server in a new goroutine
// and returns a Stopper function
func (s *Server) Start() Stopper {

	db, _ := db.Open("tmp/agora_local.db")
	userHandler := user.NewUserHandler(db)
	userHandler.CreateDBTable()
	commentHandler := comment.NewCommentHandler(db)
	commentHandler.CreateDBTable()
	postHandler := post.NewPostHandler(db, commentHandler)
	postHandler.CreateDBTable()

	go func() {
		var router = mux.NewRouter()
		fs := http.FileServer(http.FS(staticFiles))

		//
		// ROUTES
		//
		router.StrictSlash(true)
		router.Use(loggingMiddleware)
		router.PathPrefix("/static/").Handler(fs)

		router.HandleFunc("/", makeHandleCallback(postHandler.PostListHandler, userHandler)).Methods("GET")
		router.HandleFunc("/", postHandler.PostListHandler).Methods("GET")

		router.HandleFunc("/login", startLogin).Methods("GET")

		router.HandleFunc("/posts/", postHandler.PostListHandler).Methods("GET")
		router.HandleFunc("/posts/submit", postHandler.PostSubmitGETHandler).Methods("GET")
		router.HandleFunc("/posts/submit", postHandler.PostSubmitPOSTHandler).Methods("POST")
		router.HandleFunc("/posts/{id}", postHandler.PostDetailGETHandler).Methods("GET")
		router.HandleFunc("/posts/{id}/comment", postHandler.PostCommentPOSTHandler).Methods("POST")

		var address = fmt.Sprintf("%s:%s", s.host, s.port)
		s.server = &http.Server{Addr: address, Handler: router}

		log.Info.Printf("state=http_listening address=%s", s.Address())
		go func() {
			if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				log.Error.Println(err)
			}
		}()
		signal.Notify(s.stop, os.Interrupt, syscall.SIGTERM)
		s.waitForStop(&s.stop, s.server)
	}()

	return s.Stop
}

func makeOuath2Config() *oauth2.Config {
	config, err := LoadAzureConfig()
	if err != nil {
		log.Error.Fatalf("msg='could not load azure config' err='%s'\n", err.Error())
	}

	return &oauth2.Config{
		ClientID:     config.ClientID,
		ClientSecret: config.ClientSecret,
		RedirectURL:  "http://localhost:54324/", //  in my case RedirectURL:  "http://localhost:8080/callback"
		Scopes:       []string{"User.Read"},
		Endpoint: oauth2.Endpoint{
			AuthURL:  "https://login.microsoftonline.com/" + config.TenantID + "/oauth2/v2.0/authorize",
			TokenURL: "https://login.microsoftonline.com/" + config.TenantID + "/oauth2/v2.0/token",
		},
	}
}

func startLogin(w http.ResponseWriter, r *http.Request) {
	oauth2Config := makeOuath2Config()
	url := oauth2Config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	http.Redirect(w, r, url, http.StatusFound)
}

func makeHandleCallback(fallbackCallback http.HandlerFunc, userHandler *user.UserHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		oauth2Config := makeOuath2Config()
		code := r.URL.Query().Get("code")
		log.Debug.Printf("msg='received callback' code='%s'\n", code)
		if code == "" {
			fallbackCallback(w, r)
			return
		}

		token, err := oauth2Config.Exchange(context.Background(), code)
		if err != nil {
			http.Error(w, "Failed to exchange token: "+err.Error(), http.StatusInternalServerError)
			return
		}
		log.Debug.Printf("msg='token received' token='%s'\n", token.AccessToken)

		client := oauth2Config.Client(context.Background(), token)
		resp, err := client.Get("https://graph.microsoft.com/v1.0/me")
		if err != nil {
			http.Error(w, "Failed to get user info: "+err.Error(), http.StatusInternalServerError)
			return
		}
		defer resp.Body.Close()

		data, err := io.ReadAll(resp.Body)
		if err != nil {
			http.Error(w, "Failed to read user info: "+err.Error(), http.StatusInternalServerError)
			return
		}

		// Unmarshal the JSON data into the MSGraphUser struct
		var user MSGraphUser
		if err := json.Unmarshal(data, &user); err != nil {
			http.Error(w, "Failed to parse user info: "+err.Error(), http.StatusInternalServerError)
			return
		}

		log.Pretty(user)
		if !userHandler.UserExists(user.ID) {
			userHandler.AddUser(user.ID, user.DisplayName, user.Mail)
		}
		http.Redirect(w, r, "/", http.StatusPermanentRedirect)

	}
}

// Stops the server
// Users can wait until the server is stopped:
// ```go
// s := s.NewServer("localhost","8080")
// someOtherFunc()
// stop := s.Start()
// <-stop()
// ````
type Stopper = func() Stopped

// Address returns the full address of the server
func (s *Server) Address() string {
	return fmt.Sprintf("http://%s:%s", s.host, s.port)
}

// Stop stops the server and returns the Sopped chanel
func (s *Server) Stop() chan struct{} {
	s.stop <- os.Interrupt
	return s.stopped
}

// Stopped chan receives an empty struct when the server has stopped
type Stopped = chan struct{}

func (s *Server) waitForStop(stop *chan os.Signal, server *http.Server) {
	<-s.stop
	err := s.server.Close()
	if err != nil {
		log.Error.Println(err)
	}
	s.stopped <- struct{}{}
}

// WaitTilRunning waits until the server is stopped
func (s *Server) WaitTilRunning() {
	<-s.stopped
}

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Debug.Println(r.Method, r.URL.Path)
		next.ServeHTTP(w, r)
	})
}

type AzureConfig struct {
	TenantID     string
	ClientID     string
	ClientSecret string
}

func LoadAzureConfig() (AzureConfig, error) {

	err := godotenv.Load()
	if err != nil {
		log.Error.Fatalf("msg='could not load .env file' err='%s'\n", err.Error())
	}

	tenantID := os.Getenv("AZURE_TENANT_ID")
	if tenantID == "" {
		return AzureConfig{}, errors.New("missing AZURE_TENANT_ID environment variable")
	}

	clientID := os.Getenv("AZURE_CLIENT_ID")
	if clientID == "" {
		return AzureConfig{}, errors.New("missing AZURE_CLIENT_ID environment variable")
	}

	clientSecret := os.Getenv("AZURE_CLIENT_SECRET")
	if clientSecret == "" {
		return AzureConfig{}, errors.New("missing AZURE_CLIENT_SECRET environment variable")
	}

	return AzureConfig{
		TenantID:     tenantID,
		ClientID:     clientID,
		ClientSecret: clientSecret,
	}, nil
}
