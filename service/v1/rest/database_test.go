package v1rest

// Author: Sebastian Oleksiak (oscarsierraproject@protonmail.com)
// License: The Unlicense
// Created: August 18, 2024

import (
	"database/sql"
	"log"
	"testing"

	"github.com/stretchr/testify/assert"
)

var (
	Test_event_1 EventData = EventData{
		Common{EventDataStructName},
		0, "1.1.1", "e0b2dd0f43614138995beafa87b6356b", "Ur. Mr X",
		DateTime{Common{DateTimeStructName}, 2021, 1, 12, 0, 0},
		DateTime{Common{DateTimeStructName}, 2021, 1, 12, 0, 0},
		"Warszawa, ul. Okrężna 26", "Likes beer", 7, false, true, false, "APP"}
	Test_event_2 EventData = EventData{
		Common{EventDataStructName},
		0, "1.1.1", "5bd8fa795fa04bf79c37dd1b9583709f", "Im. Miss Y",
		DateTime{Common{DateTimeStructName}, 2024, 2, 13, 12, 0},
		DateTime{Common{DateTimeStructName}, 2024, 2, 13, 12, 0},
		"Łódź, ul. Rzgowska 65", "Likes flowers", 7, false, true, false, "WEB"}
)

func Test_NewSqliteRepository(t *testing.T) {
	/* GIVEN a new SQLiteRepository structure
	 * WHEN NewSqliteRepository is called
	 * THEN a new SQLiteRepository structure should be returned
	 */
	db, err := sql.Open("sqlite3", SQL_FILE)
	if err != nil {
		log.Fatal(err)
	}

	sut := NewSQLiteRepository(db)

	assert.NotNil(t, sut.db)
}

func Test_Migrate(t *testing.T) {
	/* GIVEN fresh SQLiteRepository structure
	 * WHEN Migrate() is called
	 * THEN no errors should be returned
	 * AND all tables should be returned in database.
	 */
	var (
		err error
		db  *sql.DB
	)
	db, err = sql.Open("sqlite3", SQL_FILE)
	if err != nil {
		log.Fatal(err)
	}

	sut := NewSQLiteRepository(db)

	assert.NotNil(t, sut.db)
	err = sut.Migrate()
	assert.NoError(t, err)

	sut.Close()
}

func Test_GetAllEvents(t *testing.T) {
	/* GIVEN fresh SQLiteRepository structure
	 * WHEN Migrate() is called
	 * THEN no errors should be returned
	 * AND all tables should be returned in database.
	 */
	var (
		err error
		db  *sql.DB
	)
	db, err = sql.Open("sqlite3", SQL_FILE)
	if err != nil {
		log.Fatal(err)
	}

	sut := NewSQLiteRepository(db)
	assert.NotNil(t, sut.db)
	err = sut.Migrate()
	assert.NoError(t, err)

	_, err = sut.InsertEvent(Test_event_1)
	assert.NoError(t, err)

	_, err = sut.InsertEvent(Test_event_2)
	assert.NoError(t, err)

	result, err := sut.GetAllEvents()
	assert.Len(t, result, 2)

	assert.Equal(t, result[0].Urgent, Test_event_1.Urgent)
	assert.Equal(t, result[0].Important, Test_event_1.Important)
	assert.Equal(t, result[0].Done, Test_event_1.Done)

	assert.NotEqualf(t, result[0].Id, Test_event_1.Id, "Event ID should be populated with database value")
	assert.NotEqualf(t, result[1].Id, Test_event_2.Id, "Event ID should be populated with database value")
	assert.NotEqualf(t, result[0].Id, result[1].Id, "Retrieved events should have different ID's")

	sut.Close()
}
