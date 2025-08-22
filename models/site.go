package models

import (
	// "time"
	"sync"
)
type Site struct {
	Url 	string
	Name    string
	Status  string
}

type SitesList struct {
	Websites []Site `json:"websites"`
}

type Runner struct {
	Mu sync.Mutex
	Running map[string] bool
}

