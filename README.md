# Go Calendar API
======================

A RESTful API for managing calendar events built using Go.

## Table of Contents
-----------------

* [Overview](#overview)
* [Features](#features)
* [Requirements](#requirements)
* [Installation](#installation)
* [Usage](#usage)
* [Authentication](#authentication)
* [API Endpoints](#api-endpoints)
* [Security](#security)
* [Contributing](#contributing)
* [License](#license)
* [Contact](#contact)


## Overview
------------

This project provides a simple calendar API that allows users to create, read, update, and delete (CRUD) events. The API is built using Go and uses a JSON Web Token (JWT) for authentication.

## Features
------------

* User authentication using JWT
* CRUD operations for calendar events
* Support for multiple users

## Requirements
------------

* Go 1.14 or higher
* A database (e.g., MySQL, PostgreSQL)
* A secret key for JWT signing (set as an environment variable `GOCALENDAR_TOKEN_SECRET`)

## Installation
------------

1. Clone the repository: `git clone https://git@github.com:oscarsierraproject/eventshub.git`
2. Install dependencies: `go get -u ./...`
3. Set the environment variables from [Authentication](#authentication) section
4. Run the API: `go run main.go`

### Configuring Your Environment: Essential Variables

Before diving into the implementation details, it's essential to ensure your environment is properly configured to support the application. This chapter outlines the critical environment variables that must be set to guarantee seamless execution. These variables play a vital role in defining the application's behavior, security, and connectivity. 

- GOCALENDAR_HOST
Description: The host IP address or hostname where the server will listen.
- GOCALENDAR_PORT
Description: The port number where the server will listen.
- GOCALENDAR_ADMIN_USERNAME
Description: The username of the administrator account.
- GOCALENDAR_ADMIN_PASSWORD
Description: The password of the administrator account. Not used by the server itself, but for auxiliary tools.
- GOCALENDAR_ADMIN_HASH
Description: The hashed password of the administrator account.
- GOCALENDAR_TOKEN_SECRET
Description: Secret used for JWT.
- GOCALENDAR_OPENSSL_CALENDAR_CERTIFICATE
Description: The path to the SSL/TLS certificate file used for secure connections.
- GOCALENDAR_OPENSSL_CALENDAR_SIGNING_KEY
Description: The path to the SSL/TLS private key file used for secure connections.
- GOCALENDAR_DEADLY_PACKAGE
Description: A package content which allow remote server kill.


## Usage
------------

### Authentication

To use the API, you need to obtain a JWT token by sending a POST request to the `/login` endpoint with your username and password.

### API Endpoints

Endpoints are defined in the Configure method of the HTTPRestServer struct, located in service/v1/rest/server.go.

* `GET /api/v1/version`: Retrieve the version of the server.
* `GET /api/v1/getEventCheckSum`: Retrieve the checksum of an event.
* `GET /api/v1/getEventsWithinTimeRange`: Retrieve a list of events within a specified time range.
* `GET api/v1/status`: Get the status of the server.
* `POST /api/v1/login`: Authenticate a user and obtain a session token.
* `POST /api/v1/insertEvent`: Insert a new event into the system.
* `POST /api/v1/ki11s3rv3rn0w`: Gracefully shut down the server.

## Security
------------

### User management
This application does not currently support user management. It is designed to support only one, configured, user. 

Future versions of the application may introduce user management, but at the moment, the application is designed to support only one user.

Configuration of user is done with GOCALENDAR_ADMIN_USERNAME and GOCALENDAR_ADMIN_HASH variables. No need to store plaintext password do user.

### API

* The API uses JWT for authentication and authorization.
* All API endpoints require a valid JWT token to be passed in the `Authorization` header.

## Contributing
------------

Contributions are welcome! Please submit a pull request with your changes.

Note: This is a basic README file, and you may want to add more details specific to your project.

## License
------------
This project is licensed under the `The Unlicense` License. See the LICENSE file for more details.

## Contact
------------
For any questions or inquiries, please contact oscarsierraproject@protonmail.com
