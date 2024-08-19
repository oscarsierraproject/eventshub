package main

import (
	logger "eventshub/logging"
	xmlparser "eventshub/xmlparser"
)

func main() {
	parser := xmlparser.NewXMLEventsParser("./xmlparser/config.json", logger.INFO)
	parser.UploadStoredEvents()
}
