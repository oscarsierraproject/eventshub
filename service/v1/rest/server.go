package v1rest

// Author: Sebastian Oleksiak (oscarsierraproject@protonmail.com)
// License: The Unlicense
// Created: August 18, 2024

import (
	"context"
	"database/sql"
	"errors"
	logger "eventshub/logging"
	"net/http"
	"os"
	"time"
)

const (
	IdleTimeout       time.Duration = 60 * time.Second
	ReadHeaderTimeout time.Duration = 2 * time.Second
	ReadTimeout       time.Duration = 1 * time.Second
	ShutdownTimeout   time.Duration = 10 * time.Second
	WriteTimeout      time.Duration = 5 * time.Second
	VERSION           string        = "1.1.0"
)

type HTTPRestServer struct {
	db            DatabaseRepo
	log           *logger.ConsoleLogger
	server        *http.Server
	sigs          chan os.Signal
	deadlyPackage string
}

func (srv *HTTPRestServer) Configure(sigs chan os.Signal) {
	var (
		err error
		db  *sql.DB
	)

	srv.sigs = sigs

	srv.log = logger.NewConsoleLogger("SERVER", logger.DEBUG)
	srv.log.Info("Configuring server.")

	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/version", srv.serverVersionHandler)
	mux.HandleFunc("/api/v1/login", srv.loginHandler)
	mux.HandleFunc("/api/v1/insertEvent", srv.insertEvent)
	mux.HandleFunc("/api/v1/getEventCheckSum", srv.getEventCheckSum)
	mux.HandleFunc("/api/v1/getEventsWithinTimeRange", srv.getEventsWithinTimeRange)
	mux.HandleFunc("/api/v1/status", srv.getStatus)
	mux.HandleFunc("/api/v1/ki11s3rv3rn0w", srv.killserver)

	host := os.Getenv("GOCALENDAR_HOST")

	if host == "" {
		err = errors.New("failed to obtain host")
		srv.log.Critical(err)
		panic(err)
	}

	port := os.Getenv("GOCALENDAR_PORT")

	if port == "" {
		err = errors.New("failed to obtain port")
		srv.log.Critical(err)
		panic(err)
	}

	if deadlyPackage := os.Getenv("GOCALENDAR_DEADLY_PACKAGE"); deadlyPackage == "" {
		err = errors.New("failed to obtain deadly package")
		srv.log.Critical(err)
	} else {
		srv.deadlyPackage = deadlyPackage
	}

	srv.log.Info("Server will listen on ", host, ":", port)

	srv.server = &http.Server{
		ReadTimeout:       ReadTimeout,
		WriteTimeout:      WriteTimeout,
		IdleTimeout:       IdleTimeout,
		ReadHeaderTimeout: ReadHeaderTimeout,
		Addr:              host + ":" + port,
		Handler:           mux,
	}

	db, err = sql.Open("sqlite3", SQLFile)
	if err != nil {
		srv.log.Critical(err)
		panic(err)
	}

	srv.db = NewSQLiteRepository(db)

	err = srv.db.Migrate()
	if err != nil {
		srv.log.Critical(err)
		panic(err)
	}

	/* Store hashed password for the user */
	adminUsername := os.Getenv("GOCALENDAR_ADMIN_USERNAME")
	if adminUsername == "" {
		err = errors.New("failed to obtain adminUsername")
		srv.log.Critical(err)
		panic(err)
	}

	adminHash := os.Getenv("GOCALENDAR_ADMIN_HASH")
	if adminHash == "" {
		err = errors.New("failed to obtain adminHash")
		srv.log.Critical(err)
		panic(err)
	}

	err = srv.db.AddUser(adminUsername, adminHash, true)
	if err != nil {
		srv.log.Critical(err)
		panic(err)
	}
}

func (srv *HTTPRestServer) Start() {
	/* Starts HTTPRestServer as a goroutine. */
	srv.log.Warning("USING NOT SECURE PROTOCOL.")

	go func() {
		err := srv.server.ListenAndServe()
		if errors.Is(err, http.ErrServerClosed) {
			srv.log.Error("HTTP REST Server is closed. ", err)
		} else if err != nil {
			srv.log.Error("HTTP REST Server error while listening. ", err)
		}
	}()
}

func (srv *HTTPRestServer) StartTLS() {
	/* Starts HTTPRestServer as a goroutine. */
	srv.log.Info("Starting TLS server.")

	go func() {
		certificatePath := os.Getenv("GOCALENDAR_OPENSSL_CALENDAR_CERTIFICATE")
		privatekeyPath := os.Getenv("GOCALENDAR_OPENSSL_CALENDAR_SIGNING_KEY")

		err := srv.server.ListenAndServeTLS(certificatePath, privatekeyPath)
		if errors.Is(err, http.ErrServerClosed) {
			srv.log.Error("HTTP REST Server is closed. ", err)
		} else if err != nil {
			srv.log.Error("HTTP REST Server error while listening. ", err)
		}

		srv.log.Warning("Stopped serving new connections")
	}()
}

func (srv *HTTPRestServer) Stop() error {
	srv.log.Warning("Shutting down server.")

	shutdownCtx, shutdownRelease := context.WithTimeout(context.Background(), ShutdownTimeout)

	defer shutdownRelease()

	if err := srv.server.Shutdown(shutdownCtx); err != nil {
		srv.log.Error("HTTP shutdown error: ", err)
	}

	srv.log.Info("Graceful shutdown complete.")

	return nil
}
