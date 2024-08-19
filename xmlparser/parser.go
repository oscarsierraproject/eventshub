package xmlparser

// Author: Sebastian Oleksiak (oscarsierraproject@protonmail.com)
// License: The Unlicense
// Created: August 18, 2024

import (
	"bytes"
	logger "eventshub/logging"
	v1rest "eventshub/service/v1/rest"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

type XMLEventsParser struct {
	config Config
	log    *logger.ConsoleLogger
	token  string
}

func NewXMLEventsParser(config_path string, logging_lvl int) XMLEventsParser {
	var (
		config Config
		log    *logger.ConsoleLogger
	)
	log = logger.NewConsoleLogger("XMLParser", logging_lvl)
	log.Info("Crating and configuring XMLEventsParser.")

	// Let's first read the `config.json` file
	content, err := os.ReadFile(config_path)
	if err != nil {
		log.Critical("Error when opening configuration file: ", err)
		panic(err)
	}

	// Now let's unmarshall the data into `payload`
	err = json.Unmarshal(content, &config)
	if err != nil {
		log.Critical("Error during Unmarshal(): ", err)
		panic(err)
	}

	return XMLEventsParser{
		config: config,
		log:    log,
		token:  "",
	}
}

func (parser *XMLEventsParser) getTransportConfiguration() (*http.Transport, error) {
	/* Prepare request transport configuration */

	caCert, err := os.ReadFile(os.Getenv("GOCALENDAR_OPENSSL_CA_CERTIFICATE"))
	if err != nil {
		return nil, err
	}

	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)

	tlsConfig := &tls.Config{
		RootCAs: caCertPool,
	}

	transport := &http.Transport{
		TLSClientConfig: tlsConfig,
		MaxIdleConns:    10,
		IdleConnTimeout: 30 * time.Second,
	}
	return transport, nil
}

func (parser *XMLEventsParser) getToken() {
	/* Login and get JWT */
	parser.log.Info("Begin requesting the token.")
	url := fmt.Sprintf("https://%s:%d/api/v1/login", parser.config.Host, parser.config.Port)

	var (
		err       error
		token_msg v1rest.TokenMsg
		user      v1rest.User = v1rest.User{
			Username: os.Getenv("GOCALENDAR_ADMIN_USERNAME"),
			Password: os.Getenv("GOCALENDAR_ADMIN_PASSWORD"),
		}
	)

	if user.Username == "" || user.Password == "" {
		parser.log.Critical("Missing user data.")
	}

	userData, err := json.Marshal(&user)
	if err != nil {
		parser.log.Error(err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(userData))
	if err != nil {
		parser.log.Error(err)
	}

	transport, err := parser.getTransportConfiguration()
	if err != nil {
		parser.log.Error(err)
	}

	client := &http.Client{Transport: transport}
	resp, err := client.Do(req)
	if err != nil {
		parser.log.Error(err)
	}
	defer resp.Body.Close()

	responseData, err := io.ReadAll(resp.Body)
	if err != nil {
		parser.log.Error(err)
	}

	err = json.Unmarshal(responseData, &token_msg)
	if err != nil {
		parser.log.Error(err)
	}

	parser.log.Info("Successfully obtained the token.")
	parser.token = token_msg.Token
}

func (parser *XMLEventsParser) postEvent(e v1rest.EventData) {
	url := fmt.Sprintf("https://%s:%d/api/v1/insertEvent", parser.config.Host, parser.config.Port)

	addEventReq := v1rest.AddEventReq{Event: e}
	data, err := json.Marshal(addEventReq)
	if err != nil {
		parser.log.Error(err)
		return
	}

	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(data))
	req.Header.Set("Token", parser.token)
	req.Header.Set("Content-Type", "application/json")

	transport, err := parser.getTransportConfiguration()
	if err != nil {
		parser.log.Error(err)
		panic(err)
	}

	client := &http.Client{Transport: transport}
	resp, err := client.Do(req)
	if err != nil {
		parser.log.Error(err)
		panic(err)
	}
	defer resp.Body.Close()

	for retry := 3; retry <= 3; retry++ {
		switch resp.StatusCode {
		case http.StatusOK:
			parser.log.Debug("Successfully added event with UUID ", e.Uuid)
		case http.StatusUnauthorized:
			parser.getToken()
			parser.log.Info("Unauthorized. Refreshing token.")
		default:
			parser.log.Info("Failed to add event with UUID ", e.Uuid)
		}
	}
}

func (parser *XMLEventsParser) UploadStoredEvents() {
	for _, path := range parser.config.Source_files_paths {
		parser.log.Info("Reading data from ", path)
		xmlFile, err := os.Open(path)
		if err != nil {
			log.Fatalf("%v", err)
		}
		defer xmlFile.Close()

		byteValue, _ := io.ReadAll(xmlFile)

		var root Root
		xml.Unmarshal(byteValue, &root)

		parser.log.Debug("Uploading data from ", path)
		for i := 0; i < len(root.Events); i++ {
			e := xmlEventToEventDataConverter(root.Events[i])
			parser.postEvent(e)
		}
	}

}
