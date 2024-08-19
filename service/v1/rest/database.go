package v1rest

// Author: Sebastian Oleksiak (oscarsierraproject@protonmail.com)
// License: The Unlicense
// Created: August 18, 2024

import (
	logger "eventshub/logging"
	"database/sql"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

var (
	SQL_FILE string = "file::memory:?cache=shared"
	DB       *sql.DB
)

type DatabaseRepo interface {
	AddUser(user string, password string, hashed bool) error
	AuthenticateUser(user string, password string) (bool, error)
	Close()
	DeleteEvent(e EventData) (bool, error)
	GetAllEvents() ([]EventData, error)
	GetEventsByTimeRange(start, end int64) ([]EventData, error)
	GetEventByUuid(uuid string) (EventData, error)
	GetStatus() (GetStatusResp, error)
	InsertEvent(e EventData) (EventData, error)
	Migrate() error
}

type SQLiteRepository struct {
	db  *sql.DB
	log *logger.ConsoleLogger
}

func NewSQLiteRepository(db *sql.DB) *SQLiteRepository {
	return &SQLiteRepository{
		db:  db,
		log: logger.NewConsoleLogger("SQLite", logger.INFO),
	}
}

func (r *SQLiteRepository) insertEvent(e EventData) (EventData, error) {
	/* Insert event to database. */
	var (
		err            error
		result         sql.Result
		statement      *sql.Stmt
		insertEventSQL string = `
			INSERT INTO events (
				version, uuid, title, 
				start, end, address, 
				info, reminder, done, 
				important, urgent, source) 
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?);
		`
	)
	statement, err = r.db.Prepare(insertEventSQL)
	if err != nil {
		r.log.Error(err)
		return EventData{}, err
	}

	start, _ := dateTimeToUnix(&e.Start)
	end, _ := dateTimeToUnix(&e.End)
	done := Btoi(e.Done)
	important := Btoi(e.Important)
	urgent := Btoi(e.Urgent)

	result, err = statement.Exec(e.Version, e.Uuid, e.Title, start, end, e.Address, e.Info, e.Reminder, done, important, urgent, e.Source)
	if err != nil {
		r.log.Error(err)
		return EventData{}, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		r.log.Error("Failed to get LastID.", err)
		return EventData{}, err
	}

	e.Id = id
	r.updateStatus()
	return e, nil
}

func (r *SQLiteRepository) updateEvent(e EventData) (EventData, error) {
	/* Update existing event with latest data */
	var (
		err            error
		statement      *sql.Stmt
		updateEventSQL string = `
		UPDATE events
		SET
			version = ?, 
			title = ?,
			start = ?,
			end = ?,
			address = ?, 
			info = ?, 
			reminder = ?, 
			done = ?, 
			important = ?,
			urgent = ?,
			source = ? 
		WHERE
			uuid = ?;
		`
	)

	statement, err = r.db.Prepare(updateEventSQL)
	if err != nil {
		r.log.Error(err)
		return EventData{}, err
	}

	start, _ := dateTimeToUnix(&e.Start)
	end, _ := dateTimeToUnix(&e.End)
	done := Btoi(e.Done)
	important := Btoi(e.Important)
	urgent := Btoi(e.Urgent)

	_, err = statement.Exec(e.Version, e.Title, start, end, e.Address, e.Info, e.Reminder, done, important, urgent, e.Source, e.Uuid)
	if err != nil {
		r.log.Error(err)
		return EventData{}, err
	}
	r.updateStatus()
	return e, nil
}

func (r *SQLiteRepository) updateStatus() error {
	/* Update status table */
	var (
		err             error
		statement       *sql.Stmt
		updateStatusSQL string = `INSERT INTO status (timestamp, version) VALUES (?, ?)`
	)
	statement, err = r.db.Prepare(updateStatusSQL)
	if err != nil {
		r.log.Error(err)
		return err
	}

	t := time.Now().Unix()

	_, err = statement.Exec(t, VERSION)
	if err != nil {
		r.log.Error(err)
		return err
	}

	return nil
}

func (r *SQLiteRepository) AddUser(user string, password string, hashed bool) error {
	/* Add new user to database */
	var (
		err           error
		hash          string
		statement     *sql.Stmt
		insertUserSQL string = "INSERT INTO users (username, password) VALUES (?, ?);"
	)
	if !hashed {
		hash, err = hashPassword(password)
		if err != nil {
			r.log.Error(err)
			return err
		}
	} else {
		hash = password
	}

	statement, err = r.db.Prepare(insertUserSQL)
	if err != nil {
		r.log.Error(err)
		return err
	}

	_, err = statement.Exec(user, hash)
	if err != nil {
		r.log.Error(err)
		return err
	}
	return nil
}

func (r *SQLiteRepository) AuthenticateUser(username string, password string) (bool, error) {
	/* Authenticate user  */
	var (
		err  error
		rows *sql.Rows
		user User
	)

	rows, err = r.db.Query("SELECT username, password FROM users WHERE username = ?;", username)
	if err != nil {
		r.log.Error(err)
		return false, err
	}

	for rows.Next() {
		if err := rows.Scan(&user.Username, &user.Password); err != nil {
			r.log.Error(err)
			return false, err
		}
	}

	return checkPasswordHash(password, user.Password), nil
}

func (r *SQLiteRepository) Close() {
	/* Cleanup SQLiteRepository resources */
	r.log.Info("Closing database.")
	r.db.Close()
}

func (r *SQLiteRepository) DeleteEvent(e EventData) (bool, error) {
	/* Delete event based on Event UUID */
	var (
		deleteEventSQL string = "DELETE FROM events WHERE uuid = ?;"
		err            error
		statement      *sql.Stmt
	)
	statement, err = r.db.Prepare(deleteEventSQL)
	if err != nil {
		r.log.Error(err)
		return false, err
	}

	_, err = statement.Exec(e.Uuid)
	if err != nil {
		r.log.Error(err)
		return false, err
	}

	return true, err
}

func (r *SQLiteRepository) GetAllEvents() ([]EventData, error) {
	/* Return result events present in database. */
	var (
		result []EventData
	)

	rows, err := r.db.Query("SELECT * FROM events")
	if err != nil {
		r.log.Error(err)
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		e, err := convertRawEventRecordToEventData(rows)
		if err != nil {
			r.log.Error(err)
			continue
		}

		result = append(result, e)
	}
	return result, nil
}

func (r *SQLiteRepository) GetEventsByTimeRange(start, end int64) ([]EventData, error) {
	/* Return result events present in database listed by provided time range. */
	var (
		result []EventData
	)

	rows, err := r.db.Query("SELECT * FROM events WHERE end >= ? AND start <= ?", start, end)
	if err != nil {
		r.log.Error(err)
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		e, err := convertRawEventRecordToEventData(rows)
		if err != nil {
			r.log.Error(err)
			continue
		}

		result = append(result, e)
	}
	return result, nil
}

func (r *SQLiteRepository) GetEventByUuid(uuid string) (EventData, error) {
	/* Return events based on UUID. */
	rows, err := r.db.Query("SELECT * FROM events WHERE uuid = ?", uuid)
	if err != nil {
		return EventData{}, err
	}
	defer rows.Close()

	for rows.Next() {
		e, err := convertRawEventRecordToEventData(rows)
		if err != nil {
			r.log.Error(err)
			return EventData{Common: Common{Type: EventDataStructName}}, err
		}
		return e, nil
	}
	return EventData{Common: Common{Type: EventDataStructName}}, nil
}

func (r *SQLiteRepository) GetStatus() (GetStatusResp, error) {
	/* Return present server status */
	var (
		resp GetStatusResp
	)

	resp.Common = Common{Type: ResponseStatusName}
	rows, err := r.db.Query("SELECT timestamp, version FROM status WHERE ROWID IN ( SELECT max( ROWID ) FROM status);")
	if err != nil {
		r.log.Error(err)
		resp.Status = ResponseStatus{Common{ResponseStatusName}, false, err.Error()}
		return resp, err
	}
	defer rows.Close()

	for rows.Next() {
		if err := rows.Scan(&resp.Timestamp, &resp.Version); err != nil {
			r.log.Error(err)
			resp.Status = ResponseStatus{Common{ResponseStatusName}, false, err.Error()}
			return GetStatusResp{}, err
		}
	}

	resp.Status = ResponseStatus{Common{ResponseStatusName}, true, ""}
	return resp, nil
}

func (r *SQLiteRepository) InsertEvent(e EventData) (EventData, error) {
	/* Insert new event into database, or update existing one.
	 * Event will be updated if database contains different event with same UUID.
	 * Event will be inserted is event UUID is unique in database.
	 */
	var (
		err      error
		db_event EventData
	)

	rows, err := r.db.Query("SELECT * FROM events WHERE uuid = ?", e.Uuid)
	if err != nil {
		r.log.Error(err)
		return e, err
	}

	for rows.Next() {
		/* Event exist in database. Check if update is needed */
		db_event, err = convertRawEventRecordToEventData(rows)
		if err != nil {
			r.log.Error(err)
			return e, err
		}

		rows.Close()
		e.Id = db_event.Id

		/* Check if passed event has some changes that requires update */
		if db_event.Sha256() == e.Sha256() {
			return e, nil
		}

		return r.updateEvent(e)
	}

	rows.Close()
	return r.insertEvent(e)
}

func (r *SQLiteRepository) Migrate() error {
	/* This database is in memory database. Create database structure from scratch. */
	var (
		err             error
		createEventsSQL string = `
		CREATE TABLE IF NOT EXISTS events (
			id INTEGER PRIMARY KEY,
			version VARCHAR(16),
			uuid VARCHAR(32),
			title VARCHAR(255),
			start INTEGER,
			end	INTEGER,
			address VARCHAR(255),
			info VARCHAR(255),
			reminder INTEGER,
			done INTEGER,
			important INTEGER,
			urgent INTEGER,
			source VARCHAR(255))
		`
		createUsersSQL string = `
		CREATE TABLE IF NOT EXISTS users (
			id INTEGER PRIMARY KEY,
			username VARCHAR(64),
			password VARCHAR(64));
		`
		createStatusSQL string = `
		CREATE TABLE IF NOT EXISTS status (
			id INTEGER PRIMARY KEY,
			timestamp INTEGER,
			version VARCHAR(64));
		`
		statement *sql.Stmt
	)
	statement, err = r.db.Prepare(createEventsSQL)
	if err != nil {
		r.log.Critical("Failed to create table 'events'.")
	} else {
		r.log.Info("Successfully created table 'events'.")
	}
	statement.Exec()

	statement, err = r.db.Prepare(createUsersSQL)
	if err != nil {
		r.log.Critical("Failed to create table 'users'.")
	} else {
		r.log.Info("Successfully created table 'users'.")
	}
	statement.Exec()

	statement, err = r.db.Prepare(createStatusSQL)
	if err != nil {
		r.log.Critical("Failed to create table 'status'.")
	} else {
		r.log.Info("Successfully created table 'status'.")
	}
	statement.Exec()

	r.updateStatus()

	return nil
}
