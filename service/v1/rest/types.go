package v1rest

// Author: Sebastian Oleksiak (oscarsierraproject@protonmail.com)
// License: MIT
// Created: August 18, 2024

import (
	"crypto/sha256"
	"fmt"
	"time"
)

const (
	DateTimeStructName       string        = "DateTime"
	EventDataStructName      string        = "EventData"
	ResponseStatusName       string        = "ResponseStatus"
	AddEventRespName         string        = "AddEventResp"
	GetEventCheckSumRespName string        = "GetEventCheckSumResp"
	GetEventsRespName        string        = "GetEventsResp"
	GetStatusRespName        string        = "GetStatusResp"
	InvalidTokenRespName     string        = "InvalidTokenResp"
	KillRespName             string        = "KillResp"
	Version                  string        = "v1.1.0"
	VersionRespName          string        = "VersionResp"
	GracefulShutdownTimeout  time.Duration = 2 * time.Second
)

type Common struct {
	Type string `json:"__type__,omitempty"`
}

type User struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type DateTime struct {
	Common
	Year   int32 `json:"year"`
	Month  int32 `json:"month"`
	Day    int32 `json:"day"`
	Hour   int32 `json:"hour"`
	Minute int32 `json:"minute"`
}

//nolint:govet //All structs should have similar attributes order
type EventData struct {
	Common
	ID        int64    `json:"id"`
	Version   string   `json:"version"`
	UUID      string   `json:"uuid"`
	Title     string   `json:"title"`
	Start     DateTime `json:"start"`
	End       DateTime `json:"end"`
	Address   string   `json:"address"`
	Info      string   `json:"info"`
	Reminder  int32    `json:"reminder"`
	Done      bool     `json:"done"`
	Important bool     `json:"important"`
	Urgent    bool     `json:"urgent"`
	Source    string   `json:"source"`
}

func (e *EventData) Sha256() [32]byte {
	// Sha256 returns the SHA256 hash of the EventData.
	//
	// Parameter: EventData object.
	// Return type: [32]byte.
	hash := sha256.Sum256([]byte(e.ToString()))
	return hash
}

func (e *EventData) ToString() string {
	// ToString converts EventData object to a string representation.
	//
	// Parameter: EventData object (self).
	// Return type: string.
	result := fmt.Sprintf(
		"Version: %s, UUID: %s, Title: %s, Start: %v, End: %v, Address: %s, Info: %s, Reminder: %d, Done: %t, Important: %t, Urgent: %t",
		e.Version, e.UUID, e.Title, e.Start, e.End, e.Address, e.Info, e.Reminder, e.Done, e.Important, e.Urgent)

	return result
}

//nolint:govet //All structs should have similar attributes order
type ResponseStatus struct {
	Common
	Success bool   `json:"success"`
	Message string `json:"message"`
}

type AddEventReq struct {
	Event EventData `json:"event"`
}

type AddEventResp struct {
	Common
	Status ResponseStatus `json:"status"`
}

type GetEventCheckSumReq struct {
	UUID string `json:"uuid"`
}

type GetEventCheckSumResp struct {
	Common
	Sum    string         `json:"sum"`
	Status ResponseStatus `json:"status"`
}

type GetEventsReq struct {
	Start DateTime `json:"start"`
	End   DateTime `json:"end"`
}

//nolint:govet //All structs should have similar attributes order
type GetEventsResp struct {
	Common
	Events []EventData    `json:"events"`
	Status ResponseStatus `json:"status"`
}

type GetStatusReq struct {
}

//nolint:govet //All structs should have similar attributes order
type GetStatusResp struct {
	Common
	Timestamp int64          `json:"timestamp"`
	Status    ResponseStatus `json:"status"`
	Version   string         `json:"version"`
}

type InvalidTokenResp struct {
	Common
	Status ResponseStatus `json:"status"`
}

type KillReq struct {
	Payload string `json:"payload"`
}

type KillResp struct {
	Common
	Status ResponseStatus `json:"status"`
}

type TokenMsg struct {
	Token string `json:"token"`
}

type VersionResp struct {
	Common
	Status  ResponseStatus `json:"status"`
	Version string         `json:"version"`
}
