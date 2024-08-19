package v1rest

// Author: Sebastian Oleksiak (oscarsierraproject@protonmail.com)
// License: The Unlicense
// Created: August 18, 2024

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"syscall"
	"time"
)

func (srv *HTTPRestServer) loginHandler(writer http.ResponseWriter, request *http.Request) {
	/* Check if sent user credentials are valid. Respond with JWT if
	 * login is successful, or with error message otherwise.
	 */
	var (
		authenticated bool
		err           error
		user          User
	)

	switch request.Method {
	case "POST":
		err = json.NewDecoder(request.Body).Decode(&user)
		if err != nil {
			fmt.Println(err)
			fmt.Fprintf(writer, "Invalid or corrupted request!")
			return
		}
		authenticated, _ = srv.db.AuthenticateUser(user.Username, user.Password)
		if !authenticated {
			fmt.Println("Not enough mana!")
			return
		}
		writer.WriteHeader(http.StatusOK)

		token, err := createJWT(user.Username)
		if err != nil {
			fmt.Fprintf(writer, "%s", err)
		}
		data := TokenMsg{Token: token}
		jsonData, err := json.Marshal(data)
		if err != nil {
			log.Fatalf("Marshaling data failed: %s", err.Error())
		}
		writer.Write(jsonData)
		return

	default:
		fmt.Fprintf(writer, "%s is not implemented.", request.Method)
		return
	}
}

func (srv *HTTPRestServer) serverVersionHandler(w http.ResponseWriter, r *http.Request) {
	/* Respond to server status request. */
	w.Header().Set("Content-Type", "application/json")

	err := validateJWT(w, r)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		resp := make(map[string]string)
		resp["message"] = fmt.Sprint(err)
		jsonResp, err := json.Marshal(resp)
		if err != nil {
			log.Fatalf("Error happened in JSON marshal. Err: %s", err)
		}
		w.Write(jsonResp)
		return
	}

	w.WriteHeader(http.StatusOK)
	data := VersionMsg{Version: Version}
	jsonData, err := json.Marshal(data)
	if err != nil {
		log.Fatalf("Marshaling data failed: %s", err.Error())
	}
	w.Write(jsonData)
}

func (srv *HTTPRestServer) getEventCheckSum(w http.ResponseWriter, r *http.Request) {
	/* Get event check sum */
	var (
		err      error
		event    EventData
		response GetEventCheckSumResp
	)
	err = validateJWT(w, r)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		resp := make(map[string]string)
		resp["message"] = fmt.Sprint(err)
		jsonResp, err := json.Marshal(resp)
		if err != nil {
			log.Fatalf("Error happened in JSON marshal. Err: %s", err)
		}
		w.Write(jsonResp)
		return
	}

	var msg_data GetEventCheckSumReq
	err = json.NewDecoder(r.Body).Decode(&msg_data)
	if err != nil {
		log.Fatal(err)
	}

	w.Header().Set("Content-Type", "application/json")

	event, err = srv.db.GetEventByUuid(msg_data.Uuid)
	if err != nil {
		log.Fatal(err)
	}

	response.Common = Common{Type: GetEventCheckSumRespName}
	response.Sum = fmt.Sprintf("%x", event.Sha256())
	response.Status = ResponseStatus{Common: Common{Type: ResponseStatusName}, Success: true, Message: ""}

	w.WriteHeader(http.StatusOK)
	jsonData, err := json.Marshal(response)
	if err != nil {
		log.Fatalf("Marshaling data failed: %s", err.Error())
	}
	w.Write(jsonData)
}

func (srv *HTTPRestServer) getStatus(w http.ResponseWriter, r *http.Request) {
	/* Get event check sum */
	var (
		err      error
		response GetStatusResp
	)

	w.Header().Set("Content-Type", "application/json")

	response, err = srv.db.GetStatus()
	if err != nil {
		log.Fatal(err)
		response.Status.Success = false
		response.Status.Message = fmt.Sprintf("%s", err)
	} else {
		response.Status.Success = true
		response.Status.Message = ""
	}
	response.Type = GetStatusRespName

	w.WriteHeader(http.StatusOK)
	jsonData, err := json.Marshal(response)
	if err != nil {
		log.Fatalf("Marshaling data failed: %s", err.Error())
	}
	w.Write(jsonData)
}

func (srv *HTTPRestServer) insertEvent(w http.ResponseWriter, r *http.Request) {
	/* Create a random event */
	var (
		err      error
		response AddEventResp
	)

	err = validateJWT(w, r)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		resp := make(map[string]string)
		resp["message"] = fmt.Sprint(err)
		jsonResp, err := json.Marshal(resp)
		if err != nil {
			log.Fatalf("Error happened in JSON marshal. Err: %s", err)
		}
		w.Write(jsonResp)
		return
	}

	var msg_data AddEventReq
	err = json.NewDecoder(r.Body).Decode(&msg_data)
	if err != nil {
		log.Fatal(err)
	}

	w.Header().Set("Content-Type", "application/json")

	result, err := srv.db.InsertEvent(msg_data.Event)
	if err != nil {
		log.Panic(err)
	}

	response.Common = Common{Type: AddEventRespName}
	if result.Uuid == msg_data.Event.Uuid {
		response = AddEventResp{Success: true}
	} else {
		response = AddEventResp{Success: false}
	}

	w.WriteHeader(http.StatusOK)
	jsonData, err := json.Marshal(response)
	if err != nil {
		log.Fatalf("Marshaling data failed: %s", err.Error())
	}
	w.Write(jsonData)
}

func (srv *HTTPRestServer) getEventsWithinTimeRange(w http.ResponseWriter, r *http.Request) {
	/* Get all events within selected time range*/
	var (
		err      error
		response GetEventsResp
		status   ResponseStatus
	)

	responseWithError := func(w http.ResponseWriter, r *http.Request, msg string) {
		status = ResponseStatus{Common: Common{ResponseStatusName}, Success: false, Message: msg}
		response = GetEventsResp{Common: Common{Type: GetEventsRespName}, Status: status, Events: nil}
		jsonData, err := json.Marshal(response)
		if err != nil {
			log.Fatalf("Marshaling data failed: %s", err.Error())
		}
		w.WriteHeader(http.StatusOK)
		w.Write(jsonData)
	}

	err = validateJWT(w, r)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		resp := make(map[string]string)
		resp["message"] = fmt.Sprint(err)
		jsonResp, err := json.Marshal(resp)
		if err != nil {
			log.Fatalf("Error happened in JSON marshal. Err: %s", err)
		}
		w.Write(jsonResp)
		return
	}

	var msg_data GetEventsReq

	err = json.NewDecoder(r.Body).Decode(&msg_data)
	switch {
	case err == io.EOF || err != nil:
		responseWithError(w, r, "Missing body.")
		return
	}
	w.Header().Set("Content-Type", "application/json")

	start_unix, err := dateTimeToUnix(&msg_data.Start)
	if err != nil {
		responseWithError(w, r, "Start data error.")
		return
	}

	end_unix, err := dateTimeToUnix(&msg_data.End)
	if err != nil {
		responseWithError(w, r, "End data error.")
		return
	}

	result, err := srv.db.GetEventsByTimeRange(start_unix, end_unix)
	if err != nil {
		log.Printf("%v", err)
	}

	response = GetEventsResp{
		Common: Common{Type: GetEventsRespName},
		Status: ResponseStatus{
			Common:  Common{Type: ResponseStatusName},
			Success: true,
			Message: "",
		},
		Events: result,
	}

	w.WriteHeader(http.StatusOK)
	jsonData, err := json.Marshal(response)
	if err != nil {
		log.Fatalf("Marshaling data failed: %s", err.Error())
	}
	w.Write(jsonData)
}

func (srv *HTTPRestServer) killserver(w http.ResponseWriter, r *http.Request) {
	/* Kill running server from external source if correct deadly_package is provided. */

	var (
		request  KillReq
		response KillResp
	)

	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		log.Fatal(err)
	}

	if request.Payload == srv.deadly_package {
		srv.log.Critical("Received external kill signal.")

		w.WriteHeader(http.StatusOK)

		response.Status.setSuccess(true)
		response.Status.setMessage("Server will shutdown in 2 seconds!")
		jsonData, err := json.Marshal(response)
		if err != nil {
			srv.log.Error("Marshaling data failed: ", err.Error())
		}

		w.Write(jsonData)
		srv.log.Critical("Received external kill signal.")
		time.Sleep(2 * time.Second)
		srv.sigs <- syscall.SIGINT
	} else {
		srv.log.Error("Deadly package error.")

		w.WriteHeader(http.StatusOK)

		response.Status.setSuccess(false)
		response.Status.setMessage("Deadly package error.")
		jsonData, err := json.Marshal(response)
		if err != nil {
			srv.log.Error("Marshaling data failed: ", err.Error())
		}
		w.Write(jsonData)
	}
}
