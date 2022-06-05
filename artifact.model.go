package main

import "gorm.io/gorm"

type Artifact struct {
	gorm.Model
	Sha256    string `json:"sha256"`
	Name      string `json:"name"`
	SourceURL string `json:"sourceURL"`
	Size      int64  `json:"size"`
	Data      []byte `json:"data"`
}
