package v1rest

// Author: Sebastian Oleksiak (oscarsierraproject@protonmail.com)
// License: MIT
// Created: August 18, 2024

import (
	"crypto/sha256"
	"fmt"
)

const (
	DateTimeStructName       string = "DateTime"
	EventDataStructName      string = "EventData"
	ResponseStatusName       string = "ResponseStatus"
	AddEventRespName         string = "AddEventResp"
	GetEventCheckSumRespName string = "GetEventCheckSumResp"
	GetEventsRespName        string = "GetEventsResp"
	GetStatusRespName        string = "GetStatusResp"
	Version                  string = "v1.1.0"
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

type EventData struct {
	Common
	Id        int64    `json:"id"`
	Version   string   `json:"version"`
	Uuid      string   `json:"uuid"`
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
		"Version: %s, Uuid: %s, Title: %s, Start: %v, End: %v, Address: %s, Info: %s, Reminder: %d, Done: %t, Important: %t, Urgent: %t",
		e.Version, e.Uuid, e.Title, e.Start, e.End, e.Address, e.Info, e.Reminder, e.Done, e.Important, e.Urgent)
	return result
}

type ResponseStatus struct {
	Common
	Success bool   `json:"success"`
	Message string `json:"message"`
}

func (rs *ResponseStatus) setSuccess(value bool) {
	// setSuccess sets the success status of the ResponseStatus object.
	//
	// Parameter: value (bool) - the success status to be set.
	// Return type: none.
	rs.Success = value
}

func (rs *ResponseStatus) setMessage(value string) {
	// setMessage sets the message field of the ResponseStatus struct.
	//
	// value: the new value to set.
	rs.Message = value
}

type AddEventReq struct {
	Event EventData `json:"event"`
}

type AddEventResp struct {
	Common
	Success bool `json:"success"`
}

type GetEventCheckSumReq struct {
	Uuid string `json:"uuid"`
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

type GetEventsResp struct {
	Common
	Events []EventData    `json:"events"`
	Status ResponseStatus `json:"status"`
}

type GetStatusReq struct {
}

type GetStatusResp struct {
	Common
	Timestamp int64          `json:"timestamp"`
	Status    ResponseStatus `json:"status"`
	Version   string         `json:"version"`
}

type KillReq struct {
	Payload string `json:"payload"`
}

type KillResp struct {
	Status ResponseStatus `json:"status"`
}

type TokenMsg struct {
	Token string `json:"token"`
}

type VersionMsg struct {
	Version string `json:"version"`
}
