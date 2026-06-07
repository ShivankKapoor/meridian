package models

type Location struct {
	Country     string `json:"country"`
	CountryCode string `json:"countryCode"`
	City        string `json:"city"`
	Region      string `json:"regionName"`
}
