package main

type AlfredFeedback struct {
	Items []AlfredItem `json:"items"`
	Rerun float64      `json:"rerun,omitempty"`
}

type AlfredItem struct {
	Title    string `json:"title"`
	Subtitle string `json:"subtitle,omitempty"`
}
