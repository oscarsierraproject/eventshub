package v1rest

// Author: Sebastian Oleksiak (oscarsierraproject@protonmail.com)
// License: The Unlicense
// Created: August 18, 2024

import (
	logger "eventshub/logging"
	"context"
	"database/sql"
	"errors"
	"net/http"
	"os"
	"time"
)

const (
	VERSION string = "1.0.0"
)

type HTTPRestServer struct {
	db             DatabaseRepo
	log            *logger.ConsoleLogger
	server         *http.Server
	sigs           chan os.Signal
	deadly_package string
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
		err := errors.New("failed to obtain host")
		srv.log.Critical(err)
		panic(err)
	}

	port := os.Getenv("GOCALENDAR_PORT")
	if port == "" {
		err := errors.New("failed to obtain port")
		srv.log.Critical(err)
		panic(err)
	}

	if deadly_package := os.Getenv("GOCALENDAR_DEADLY_PACKAGE"); deadly_package == "" {
		err := errors.New("failed to obtain deadly package")
		srv.log.Critical(err)
		panic(err)
	} else {
		srv.deadly_package = deadly_package
	}

	srv.log.Info("Server will listen on ", host, ":", port)

	srv.server = &http.Server{
		Addr:    host + ":" + port,
		Handler: mux,
	}

	db, err = sql.Open("sqlite3", SQL_FILE)
	if err != nil {
		srv.log.Critical(err)
		panic(err)
	}

	srv.db = NewSQLiteRepository(db)
	srv.db.Migrate()

	/* Store hashed password for the user */
	admin_username := os.Getenv("GOCALENDAR_ADMIN_USERNAME")
	if admin_username == "" {
		err := errors.New("failed to obtain admin_username")
		srv.log.Critical(err)
		panic(err)
	}

	admin_hash := os.Getenv("GOCALENDAR_ADMIN_HASH")
	if admin_hash == "" {
		err := errors.New("failed to obtain admin_hash")
		srv.log.Critical(err)
		panic(err)
	}

	srv.db.AddUser(admin_username, admin_hash, true)
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
	shutdownCtx, shutdownRelease := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownRelease()

	if err := srv.server.Shutdown(shutdownCtx); err != nil {
		srv.log.Error("HTTP shutdown error: ", err)
	}
	srv.log.Info("Graceful shutdown complete.")
	return nil
}
