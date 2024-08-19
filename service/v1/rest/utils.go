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
	if err := r.Scan(&e.Id, &e.Version, &e.Uuid, &e.Title,
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
	loc, _ := time.LoadLocation(timeZone)
	return time.Date(int(d.Year), time.Month(d.Month), int(d.Day), int(d.Hour), int(d.Minute), 0, 0, loc).Unix(), nil
}

func unixToDateTime(d *int64) (DateTime, error) {
	/* Convert Unix time to DateTime object*/
	timeZone := "Europe/Warsaw"
	loc, _ := time.LoadLocation(timeZone)
	t := time.Unix(*d, 0).In(loc)
	return DateTime{
		Common{Type: DateTimeStructName},
		int32(t.Year()), int32(t.Month()), int32(t.Day()), int32(t.Hour()), int32(t.Minute()),
	}, nil
}

func hashPassword(plain_password string) (string, error) {
	/* Generate a hash of a password */
	hash, err := bcrypt.GenerateFromPassword([]byte(plain_password), bcrypt.DefaultCost)
	return string(hash), err
}

func checkPasswordHash(plain_password, hash string) bool {
	/* Compare if provided plain text password matches with stored hash */
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(plain_password))
	return err == nil
}
