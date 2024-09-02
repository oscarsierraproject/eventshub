package v1rest

// Author: Sebastian Oleksiak (oscarsierraproject@protonmail.com)
// License: The Unlicense
// Created: August 18, 2024

import (
	"database/sql"
	logger "eventshub/logging"
	"time"

	// SQLite driver
	_ "github.com/mattn/go-sqlite3"
)

var (
	SQLFile = "file::memory:?cache=shared"
)

type DatabaseRepo interface {
	AddUser(user string, password string, hashed bool) error
	AuthenticateUser(user string, password string) (bool, error)
	Close()
	DeleteEvent(e *EventData) (bool, error)
	GetAllEvents() ([]EventData, error)
	GetEventsByTimeRange(start, end int64) ([]EventData, error)
	GetEventByUUID(uuid string) (EventData, error)
	GetStatus() (GetStatusResp, error)
	InsertEvent(e *EventData) (*EventData, error)
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

func (r *SQLiteRepository) insertEvent(e *EventData) (*EventData, error) {
	/* Insert event to database. */
	var (
		err            error
		result         sql.Result
		statement      *sql.Stmt
		insertEventSQL = `
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
		return nil, err
	}

	start, _ := dateTimeToUnix(&e.Start)
	end, _ := dateTimeToUnix(&e.End)
	done := Btoi(e.Done)
	important := Btoi(e.Important)
	urgent := Btoi(e.Urgent)

	result, err = statement.Exec(e.Version, e.UUID, e.Title, start, end, e.Address, e.Info, e.Reminder, done, important, urgent, e.Source)
	if err != nil {
		r.log.Error(err)
		return nil, err
	}

	id, err := result.LastInsertId()

	if err != nil {
		r.log.Error("Failed to get LastID.", err)

		return nil, err
	}

	e.ID = id

	err = r.updateStatus()
	if err != nil {
		r.log.Error(err)
		return nil, err
	}

	return e, nil
}

func (r *SQLiteRepository) updateEvent(e *EventData) (*EventData, error) {
	/* Update existing event with latest data */
	var (
		err            error
		statement      *sql.Stmt
		updateEventSQL = `
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
		return nil, err
	}

	start, _ := dateTimeToUnix(&e.Start)
	end, _ := dateTimeToUnix(&e.End)
	done := Btoi(e.Done)
	important := Btoi(e.Important)
	urgent := Btoi(e.Urgent)

	_, err = statement.Exec(e.Version, e.Title, start, end, e.Address, e.Info, e.Reminder, done, important, urgent, e.Source, e.UUID)
	if err != nil {
		r.log.Error(err)

		return nil, err
	}

	err = r.updateStatus()
	if err != nil {
		r.log.Error(err)

		return nil, err
	}

	return e, nil
}

func (r *SQLiteRepository) updateStatus() error {
	/* Update status table */
	var (
		err             error
		statement       *sql.Stmt
		updateStatusSQL = `INSERT INTO status (timestamp, version) VALUES (?, ?)`
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

func (r *SQLiteRepository) AddUser(user, password string, hashed bool) error {
	/* Add new user to database */
	var (
		err           error
		hash          string
		statement     *sql.Stmt
		insertUserSQL = "INSERT INTO users (username, password) VALUES (?, ?);"
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

func (r *SQLiteRepository) AuthenticateUser(username, password string) (bool, error) {
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

func (r *SQLiteRepository) DeleteEvent(e *EventData) (bool, error) {
	/* Delete event based on Event UUID */
	var (
		deleteEventSQL = "DELETE FROM events WHERE uuid = ?;"
		err            error
		statement      *sql.Stmt
	)

	statement, err = r.db.Prepare(deleteEventSQL)
	if err != nil {
		r.log.Error(err)
		return false, err
	}

	_, err = statement.Exec(e.UUID)
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

func (r *SQLiteRepository) GetEventByUUID(uuid string) (EventData, error) {
	/* Return events based on UUID. */
	rows, err := r.db.Query("SELECT * FROM events WHERE uuid = ?", uuid)

	if err != nil {
		return EventData{Common: Common{Type: EventDataStructName}}, err
	}

	defer rows.Close()

	if rows.Next() {
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

func (r *SQLiteRepository) InsertEvent(e *EventData) (*EventData, error) {
	/* Insert new event into database, or update existing one.
	 * Event will be updated if database contains different event with same UUID.
	 * Event will be inserted is event UUID is unique in database.
	 */
	var (
		err     error
		dbEvent EventData
	)

	rows, err := r.db.Query("SELECT * FROM events WHERE uuid = ?", e.UUID)
	if err != nil {
		r.log.Error(err)
		return e, err
	}

	if rows.Next() {
		/* Event exist in database. Check if update is needed */
		dbEvent, err = convertRawEventRecordToEventData(rows)
		if err != nil {
			r.log.Error(err)
			return e, err
		}

		rows.Close()

		e.ID = dbEvent.ID

		/* Check if passed event has some changes that requires update */
		if dbEvent.Sha256() == e.Sha256() {
			return e, nil
		}

		//nolint:govet //Event returned is same event that is passed with additional data like ID
		e, err := r.updateEvent(e)
		if err != nil {
			r.log.Error(err)
			return e, err
		}

		return e, nil
	}

	rows.Close()

	return r.insertEvent(e)
}

func (r *SQLiteRepository) Migrate() error {
	/* This database is in memory database. Create database structure from scratch. */
	var (
		err             error
		createEventsSQL = `
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
		createUsersSQL = `
		CREATE TABLE IF NOT EXISTS users (
			id INTEGER PRIMARY KEY,
			username VARCHAR(64),
			password VARCHAR(64));
		`
		createStatusSQL = `
		CREATE TABLE IF NOT EXISTS status (
			id INTEGER PRIMARY KEY,
			timestamp INTEGER,
			version VARCHAR(64));
		`
		statement *sql.Stmt
	)

	statement, err = r.db.Prepare(createEventsSQL)
	if err != nil {
		r.log.Critical("Failed to create table 'events'." + err.Error())
		return err
	}

	_, err = statement.Exec()
	if err != nil {
		r.log.Critical("Failed to create table 'events'." + err.Error())
		return err
	}

	r.log.Info("Successfully created table 'events'.")

	statement, err = r.db.Prepare(createUsersSQL)
	if err != nil {
		r.log.Critical("Failed to create table 'users'." + err.Error())
		return err
	}

	_, err = statement.Exec()
	if err != nil {
		r.log.Critical("Failed to create table 'users'." + err.Error())

		return err
	}

	r.log.Info("Successfully created table 'users'.")

	statement, err = r.db.Prepare(createStatusSQL)
	if err != nil {
		r.log.Critical("Failed to create table 'status'." + err.Error())
		return err
	}

	_, err = statement.Exec()
	if err != nil {
		r.log.Error(err)

		return err
	}

	r.log.Info("Successfully created table 'status'.")

	err = r.updateStatus()
	if err != nil {
		r.log.Error(err)

		return err
	}

	return nil
}
