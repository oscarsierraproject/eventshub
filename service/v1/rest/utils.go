package v1rest

// Author: Sebastian Oleksiak (oscarsierraproject@protonmail.com)
// License: The Unlicense
// Created: August 18, 2024

import (
	"database/sql"
	"time"

	"golang.org/x/crypto/bcrypt"
)

func Btoi(b bool) int {
	if b {
		return 1
	}

	return 0
}

func convertRawEventRecordToEventData(r *sql.Rows) (EventData, error) {
	/* Convert SQL row data into EventData structure */
	var (
		e  EventData
		t1 int64
		t2 int64
	)

	if err := r.Scan(&e.ID, &e.Version, &e.UUID, &e.Title,
		&t1, &t2, &e.Address, &e.Info, &e.Reminder,
		&e.Done, &e.Important, &e.Urgent, &e.Source); err != nil {
		return e, err
	}

	e.Type = EventDataStructName
	e.Start, _ = unixToDateTime(&t1)
	e.End, _ = unixToDateTime(&t2)

	return e, nil
}

func dateTimeToUnix(d *DateTime) (int64, error) {
	/* Convert DateTime object value to Unix time */
	timeZone := "Europe/Warsaw"

	loc, err := time.LoadLocation(timeZone)
	if err != nil {
		return 0, err
	}

	return time.Date(int(d.Year), time.Month(d.Month), int(d.Day), int(d.Hour), int(d.Minute), 0, 0, loc).Unix(), nil
}

//nolint:gosec // Only integers used for date are for conversion so no integer overflow possible
func unixToDateTime(d *int64) (DateTime, error) {
	/* Convert Unix time to DateTime object*/
	timeZone := "Europe/Warsaw"

	loc, err := time.LoadLocation(timeZone)
	if err != nil {
		return DateTime{
			Common: Common{
				Type: DateTimeStructName},
			Year:   0,
			Month:  0,
			Day:    0,
			Hour:   0,
			Minute: 0,
		}, err
	}

	t := time.Unix(*d, 0).In(loc)

	return DateTime{
		Common{Type: DateTimeStructName},
		int32(t.Year()), int32(t.Month()), int32(t.Day()), int32(t.Hour()), int32(t.Minute()),
	}, nil
}

func hashPassword(plainPassword string) (string, error) {
	/* Generate a hash of a password */
	hash, err := bcrypt.GenerateFromPassword([]byte(plainPassword), bcrypt.DefaultCost)

	return string(hash), err
}

func checkPasswordHash(plainPassword, hash string) bool {
	/* Compare if provided plain text password matches with stored hash */
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(plainPassword))

	return err == nil
}
