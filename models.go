package main

type Collection struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Total    int    `json:"total"`
	Category string `json:"category"`
}

type IconResp struct {
	Icons  map[string]Icon `json:"icons"`
	Width  float64         `json:"width"`
	Height float64         `json:"height"`
}

type Icon struct {
	Body   string  `json:"body"`
	Width  float64 `json:"width"`
	Height float64 `json:"height"`
}
