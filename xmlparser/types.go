package xmlparser

// Author: Sebastian Oleksiak (oscarsierraproject@protonmail.com)
// License: The Unlicense
// Created: August 18, 2024

import "encoding/xml"

type Config struct {
	Host               string   `json:"host"`
	Port               int      `json:"port"`
	Source_files_paths []string `json:"source_files_paths"`
}

type Root struct {
	XMLName xml.Name `xml:"root"`
	Events  []Event  `xml:"event"`
}

type Event struct {
	XMLName   xml.Name `xml:"event"`
	Version   string   `xml:"ver,attr"`
	Uuid      string   `xml:"uuid,attr"`
	Start     string   `xml:"start,attr"`
	End       string   `xml:"end,attr"`
	Remind    string   `xml:"remind,attr"`
	Done      string   `xml:"done,attr"`
	Urgent    string   `xml:"urgent,attr"`
	Important string   `xml:"important,attr"`
	Title     string   `xml:"title,attr"`
	Address   string   `xml:"address,attr"`
	Info      string   `xml:"info,attr"`
}
