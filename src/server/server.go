package server

import (
	"embed"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"agora/src/db"
	"agora/src/log"
	"agora/src/post"
	"agora/src/post/comment"
	"agora/src/server/auth"
	"agora/src/user"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
)

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

	env := LoadEnv()
	address := fmt.Sprintf("%s:%s", s.host, s.port)

	db, _ := db.Open("tmp/agora_local.db")
	userHandler := user.NewUserHandler(db)
	userHandler.CreateDBTable()

	// TODO: too many arguments, refactor
	authHandler := auth.NewAuthHandler(
		env.JWTSecret,
		address,
		env.AzureTenantID,
		env.AzureClientID,
		env.AzureClientSecret,
		userHandler,
	)

	commentHandler := comment.NewCommentHandler(db)
	commentHandler.CreateDBTable()

	postHandler := post.NewPostHandler(db, commentHandler)
	postHandler.CreateDBTable()

	go func() {
		var router = mux.NewRouter()
		fs := http.FileServer(http.FS(staticFiles))

		s.server = &http.Server{Addr: address, Handler: router}

		//
		// ROUTES
		//
		router.StrictSlash(true)
		// router.Use(loggingMiddleware)
		router.Use(authHandler.Middleware)
		router.PathPrefix("/static/").Handler(fs)

		router.HandleFunc("/", authHandler.MakeHandleCallback(postHandler.PostListHandler)).Methods("GET")
		router.HandleFunc("/", postHandler.PostListHandler).Methods("GET")

		router.HandleFunc("/login", authHandler.HandleLogin).Methods("GET")

		router.HandleFunc("/posts/", postHandler.PostListHandler).Methods("GET")
		router.HandleFunc("/posts/submit", postHandler.PostSubmitGETHandler).Methods("GET")
		router.HandleFunc("/posts/submit", postHandler.PostSubmitPOSTHandler).Methods("POST")
		router.HandleFunc("/posts/{id}", postHandler.PostDetailGETHandler).Methods("GET")
		router.HandleFunc("/posts/{id}/comment", postHandler.PostCommentPOSTHandler).Methods("POST")

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

type Env struct {
	AzureTenantID     string
	AzureClientID     string
	AzureClientSecret string
	JWTSecret         string
}

func LoadEnv() Env {
	err := godotenv.Load()
	if err != nil {
		log.Error.Fatalf("msg='could not load .env file' err='%s'\n", err.Error())
	}

	env := Env{
		AzureTenantID:     os.Getenv("AZURE_TENANT_ID"),
		AzureClientID:     os.Getenv("AZURE_CLIENT_ID"),
		AzureClientSecret: os.Getenv("AZURE_CLIENT_SECRET"),
		JWTSecret:         os.Getenv("JWT_SECRET"),
	}

	return env
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
