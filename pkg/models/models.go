package models

type Params struct {
	Source        string
	Destination   string
	Compression   int
	SkipUserInput bool // Flag to bypass user input
	DeleteSource  bool // Flag to delete source files after processing
	EnableLog     bool // Flag to enable logging
}
