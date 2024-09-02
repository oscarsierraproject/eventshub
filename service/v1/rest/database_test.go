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
	TestEvent1 = EventData{
		Common{EventDataStructName},
		0, "1.1.1", "e0b2dd0f43614138995beafa87b6356b", "Ur. Mr X",
		DateTime{Common{DateTimeStructName}, 2021, 1, 12, 0, 0},
		DateTime{Common{DateTimeStructName}, 2021, 1, 12, 0, 0},
		"Warszawa, ul. Okrężna 26", "Likes beer", 7, false, true, false, "APP"}
	TestEvent2 = EventData{
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
	db, err := sql.Open("sqlite3", SQLFile)
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

	db, err = sql.Open("sqlite3", SQLFile)
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

	db, err = sql.Open("sqlite3", SQLFile)
	if err != nil {
		log.Fatal(err)
	}

	sut := NewSQLiteRepository(db)
	assert.NotNil(t, sut.db)
	err = sut.Migrate()
	assert.NoError(t, err)

	assert.Equal(t, int64(0), TestEvent1.ID)
	_, err = sut.InsertEvent(&TestEvent1)
	assert.NoError(t, err)

	assert.Equal(t, int64(0), TestEvent2.ID)
	_, err = sut.InsertEvent(&TestEvent2)
	assert.NoError(t, err)

	result, err := sut.GetAllEvents()
	assert.NoError(t, err)
	assert.Len(t, result, 2)

	assert.Equal(t, result[0].Urgent, TestEvent1.Urgent)
	assert.Equal(t, result[0].Important, TestEvent1.Important)
	assert.Equal(t, result[0].Done, TestEvent1.Done)

	assert.Equalf(t, result[0].ID, TestEvent1.ID, "Event ID should be populated with database value, %d != %d", result[0].ID, TestEvent1.ID)
	assert.Equalf(t, result[1].ID, TestEvent2.ID, "Event ID should be populated with database value, %d != %d", result[1].ID, TestEvent2.ID)
	assert.NotEqualf(t, result[0].ID, result[1].ID, "Retrieved events should have different ID's")

	sut.Close()
}
