package v1rest

// Author: Sebastian Oleksiak (oscarsierraproject@protonmail.com)
// License: The Unlicense
// Created: August 18, 2024

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"syscall"
	"time"
)

// Send a JSON response to the client. It takes a response object and marshals it to JSON.
// If the marshaling fails, it logs the error and returns.
// If the write to the client fails, it logs the error.
func (srv *HTTPRestServer) send(resp any, w http.ResponseWriter, _ *http.Request) {
	var (
		byteResp []byte
		err      error
	)

	byteResp, err = json.Marshal(resp)
	if err != nil {
		srv.log.Error("Marshaling data failed:", err)
		return
	}

	_, err = w.Write(byteResp)
	if err != nil {
		srv.log.Error("Writing data failed:", err)
	}
}

// invalidTokenResponse sends a JSON response to the client with a 401 Unauthorized status code.
// The response body contains a JSON object with a single "status" field that describes the error.
func (srv *HTTPRestServer) invalidTokenResponse(w http.ResponseWriter, r *http.Request, reason error) {
	var (
		resp InvalidTokenResp
	)

	w.WriteHeader(http.StatusUnauthorized)
	w.Header().Set("Content-Type", "application/json")

	resp = InvalidTokenResp{
		Common: Common{
			Type: InvalidTokenRespName,
		},
		Status: ResponseStatus{
			Success: false,
			Message: fmt.Sprintf("%s", reason),
		},
	}

	srv.send(resp, w, r)
}

/*
loginHandler is an HTTP handler which handles login requests. It checks
if the provided user credentials are valid and returns a JWT token if
login is successful. Otherwise, it returns an error message.

Handler responds to POST requests only.

Example request body:

	{
		"username": "admin",
		"password": "admin"
	}

Example response:

	{
		"token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VybmFtZSI6ImFkbWluIiwiaWF0IjoxNjM4NjgzODUyfQ.0lq8Z2jwZp4J0hWZ9rj0X7g7nAa9JpFf4J6kT5mO3lA"
	}
*/
func (srv *HTTPRestServer) loginHandler(writer http.ResponseWriter, request *http.Request) {
	var (
		authenticated bool
		err           error
		user          User
	)

	switch request.Method {
	case "POST":
		err = json.NewDecoder(request.Body).Decode(&user)

		if err != nil {
			srv.log.Warning(err)
			fmt.Fprintf(writer, "Invalid or corrupted request!")

			return
		}

		authenticated, err = srv.db.AuthenticateUser(user.Username, user.Password)
		if !authenticated {
			srv.log.Info("Not enough mana!")
			fmt.Fprintf(writer, "Not enough mana!")

			return
		} else if err != nil {
			srv.log.Error(err)
			fmt.Fprintf(writer, "%s", err)

			return
		}

		writer.WriteHeader(http.StatusOK)

		token, err := createJWT(user.Username)
		if err != nil {
			srv.log.Error(err)
			fmt.Fprintf(writer, "%s", err)
		}

		data := TokenMsg{Token: token}

		jsonData, err := json.Marshal(data)
		if err != nil {
			srv.log.Error("Marshaling data failed:", err)
			return
		}

		_, err = writer.Write(jsonData)
		if err != nil {
			srv.log.Error("Writing data failed:", err)
			return
		}

		return

	default:
		srv.log.Error("Method not implemented!", request.Method)
		fmt.Fprintf(writer, "%s method not implemented!", request.Method)

		return
	}
}

/* Handle a request to the /api/v1/version endpoint. */
/* Returns server version in JSON format. */
/* If JWT token is invalid, returns 401 with error message. */
func (srv *HTTPRestServer) serverVersionHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	err := validateJWT(w, r)
	if err != nil {
		srv.invalidTokenResponse(w, r, err)
		return
	}

	w.WriteHeader(http.StatusOK)

	resp := VersionResp{
		Common: Common{
			Type: VersionRespName,
		},
		Status: ResponseStatus{
			Success: true,
			Message: "",
		},
		Version: Version,
	}

	srv.send(resp, w, r)
}

/*
Get event check sum

Handler responds to GET requests only.

Example request:

	GET /api/v1/checksum?uuid=<uuid>

Example response:

	{
		"sum": "0b2dd0f43614138995beafa87b6356b",
		"status": {
			"type": "ResponseStatus",
			"success": true,
			"message": ""
		}
	}
*/
func (srv *HTTPRestServer) getEventCheckSum(w http.ResponseWriter, r *http.Request) {
	var (
		err      error
		event    EventData
		response GetEventCheckSumResp
	)

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")

	err = validateJWT(w, r)
	if err != nil {
		srv.invalidTokenResponse(w, r, err)

		return
	}

	var msgData GetEventCheckSumReq

	err = json.NewDecoder(r.Body).Decode(&msgData)
	if err != nil {
		srv.log.Error(err)
	}

	response.Common = Common{Type: GetEventCheckSumRespName}

	event, err = srv.db.GetEventByUUID(msgData.UUID)
	if err != nil {
		srv.log.Error(err)
		response.Status = ResponseStatus{Common: Common{Type: ResponseStatusName}, Success: false, Message: fmt.Sprintf("%s", err)}
		response.Sum = fmt.Sprintf("%x", 0)
	} else {
		response.Status = ResponseStatus{Common: Common{Type: ResponseStatusName}, Success: true, Message: ""}
		response.Sum = fmt.Sprintf("%x", event.Sha256())
	}

	srv.send(response, w, r)
}

// getStatus handles a request to the /api/v1/status endpoint.
// Returns current server status in JSON format.
// If any error occurs, returns 500 with error message
func (srv *HTTPRestServer) getStatus(w http.ResponseWriter, r *http.Request) {
	var (
		err  error
		resp GetStatusResp
	)

	responseWithError := func(w http.ResponseWriter, msg string) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Header().Set("Content-Type", "application/json")

		resp = GetStatusResp{
			Common:    Common{Type: GetStatusRespName},
			Timestamp: time.Now().Unix(),
			Status:    ResponseStatus{Common: Common{ResponseStatusName}, Success: false, Message: msg},
			Version:   Version,
		}

		srv.send(resp, w, r)
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")

	resp, err = srv.db.GetStatus()
	if err != nil {
		srv.log.Error(err)
		responseWithError(w, fmt.Sprintf("%s", err))
	}

	srv.send(resp, w, r)
}

/*
insertEvent handles a request to the /api/v1/insertEvent endpoint.
Takes EventData as JSON, inserts it into database and returns
response with inserted event UUID or error message.

Example request:

	POST /api/v1/insertEvent
	{
		"event": {
			"title": "New event",
			"start": "2024-02-13T12:00:00Z",
			"end": "2024-02-13T12:00:00Z",
			"address": "Warszawa, ul. Okrężna 26",
			"info": "Some event info",
			"reminder": 7,
			"done": false,
			"important": true,
			"urgent": false,
			"source": "APP"
		}
	}

Example response:

	{
		"common": {
			"type": "AddEventResp"
		},
		"status": {
			"type": "ResponseStatus",
			"success": true,
			"message": ""
		}
	}
*/
func (srv *HTTPRestServer) insertEvent(w http.ResponseWriter, r *http.Request) {
	var (
		err  error
		resp AddEventResp
	)

	responseWithError := func(w http.ResponseWriter, msg string) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Header().Set("Content-Type", "application/json")

		resp = AddEventResp{
			Common: Common{Type: AddEventRespName},
			Status: ResponseStatus{Common: Common{ResponseStatusName}, Success: false, Message: msg},
		}

		srv.send(resp, w, r)
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")

	err = validateJWT(w, r)
	if err != nil {
		srv.invalidTokenResponse(w, r, err)
		return
	}

	var msgData AddEventReq

	err = json.NewDecoder(r.Body).Decode(&msgData)
	if err != nil {
		responseWithError(w, fmt.Sprintf("%s", err))
		return
	}

	result, err := srv.db.InsertEvent(&msgData.Event)
	if err != nil {
		srv.log.Error(err)
		responseWithError(w, fmt.Sprintf("%s", err))

		return
	}

	resp.Common = Common{Type: AddEventRespName}
	if result.UUID == msgData.Event.UUID {
		resp.Status = ResponseStatus{Common: Common{Type: ResponseStatusName}, Success: true, Message: ""}
	} else {
		resp.Status = ResponseStatus{Common: Common{Type: ResponseStatusName}, Success: false, Message: fmt.Sprintf("%s", err)}
	}

	srv.send(resp, w, r)
}

/* getEventsWithinTimeRange handles a request to the /api/v1/getEventsWithinTimeRange endpoint.
 * Takes GetEventsReq as JSON, retrieves events within the specified time range and returns
 * response with events or error message.
 *
 * Example request:
 *
 *	POST /api/v1/getEventsWithinTimeRange
 *	{
 *		"start": "2024-02-13T12:00:00Z",
 *		"end": "2024-02-13T12:00:00Z"
 *	}
 *
 * Example response:
 *
 *	{
 *		"common": {
 *			"type": "GetEventsResp"
 *		},
 *		"status": {
 *			"type": "ResponseStatus",
 *			"success": true,
 *			"message": ""
 *		},
 *		"events": []
 *	}
 */
func (srv *HTTPRestServer) getEventsWithinTimeRange(w http.ResponseWriter, r *http.Request) {
	/* Get all events within selected time range*/
	var (
		err  error
		resp GetEventsResp
	)

	responseWithError := func(w http.ResponseWriter, msg string) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Header().Set("Content-Type", "application/json")

		resp = GetEventsResp{Common: Common{Type: GetEventsRespName},
			Status: ResponseStatus{Common: Common{ResponseStatusName}, Success: false, Message: msg},
			Events: nil,
		}

		srv.send(resp, w, r)
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")

	err = validateJWT(w, r)
	if err != nil {
		srv.invalidTokenResponse(w, r, err)
		return
	}

	var msgData GetEventsReq

	err = json.NewDecoder(r.Body).Decode(&msgData)
	if err == io.EOF || err != nil {
		responseWithError(w, "Missing body.")

		return
	}

	startUnix, err := dateTimeToUnix(&msgData.Start)
	if err != nil {
		responseWithError(w, "Start data error.")

		return
	}

	endUnix, err := dateTimeToUnix(&msgData.End)
	if err != nil {
		responseWithError(w, "End data error.")

		return
	}

	result, err := srv.db.GetEventsByTimeRange(startUnix, endUnix)
	if err != nil {
		srv.log.Warning(err)
	}

	resp = GetEventsResp{
		Common: Common{Type: GetEventsRespName},
		Status: ResponseStatus{
			Common:  Common{ResponseStatusName},
			Success: false, Message: "",
		},
		Events: result,
	}

	srv.send(resp, w, r)
}

func (srv *HTTPRestServer) killserver(w http.ResponseWriter, r *http.Request) {
	/* Kill running server from external source if correct deadlyPackage is provided. */
	var (
		request  KillReq
		response KillResp
	)

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")

	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		srv.log.Error(err)
	}

	if request.Payload == srv.deadlyPackage {
		srv.log.Critical("Received external kill signal.")

		response = KillResp{
			Common: Common{Type: KillRespName},
			Status: ResponseStatus{
				Common:  Common{ResponseStatusName},
				Success: true,
				Message: "Server will shutdown in 2 seconds!",
			},
		}

		srv.send(response, w, r)

		srv.log.Critical("Received external kill signal.")
		time.Sleep(GracefulShutdownTimeout)
		srv.sigs <- syscall.SIGINT
	} else {
		srv.log.Error("Deadly package error.")

		response = KillResp{
			Common: Common{Type: KillRespName},
			Status: ResponseStatus{
				Common:  Common{ResponseStatusName},
				Success: false,
				Message: "Package error!",
			},
		}

		srv.send(response, w, r)
	}
}
