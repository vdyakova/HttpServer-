package models

import (
	"github.com/lib/pq"
	"time"
)

type Document struct {
	ID      int            `json:"id"`
	Name    string         `json:"name"`
	Mime    string         `json:"mime"`
	File    bool           `json:"file"`
	Public  bool           `json:"public"`
	Created time.Time      `json:"created"`
	Grant   pq.StringArray `json:"grant"`
}
