package xmlparser

// Author: Sebastian Oleksiak (oscarsierraproject@protonmail.com)
// License: The Unlicense
// Created: August 18, 2024

import (
	v1rest "eventshub/service/v1/rest"
	"strconv"
	"strings"
)

func yesNoToBool(s string) bool {
	return s == "Yes"
}

func stringToDateTimeConverter(s string) v1rest.DateTime {
	tmp := strings.Split(s, " ")
	date := strings.Split(tmp[0], "-")
	time := strings.Split(tmp[1], ":")
	year, _ := strconv.Atoi(date[0])
	month, _ := strconv.Atoi(date[1])
	day, _ := strconv.Atoi(date[2])
	hour, _ := strconv.Atoi(time[0])
	minute, _ := strconv.Atoi(time[1])

	dt := v1rest.DateTime{Year: int32(year), Month: int32(month), Day: int32(day), Hour: int32(hour), Minute: int32(minute)}
	dt.Type = "datetime"
	return dt
}

func xmlEventToEventDataConverter(xe Event) v1rest.EventData {
	var event v1rest.EventData
	event.Version = xe.Version
	event.Uuid = xe.Uuid
	event.Title = xe.Title
	event.Start = stringToDateTimeConverter(xe.Start)
	event.End = stringToDateTimeConverter(xe.End)
	event.Address = xe.Address
	event.Info = xe.Info
	i, err := strconv.Atoi(xe.Remind)
	if err != nil {
		panic(err)
	}
	event.Reminder = int32(i)
	event.Done = yesNoToBool(xe.Done)
	event.Important = yesNoToBool(xe.Important)
	event.Urgent = yesNoToBool(xe.Urgent)
	event.Source = "XML"
	return event
}
