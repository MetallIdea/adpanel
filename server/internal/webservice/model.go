package webservice

type WebService struct {
	ID     int64  `json:"id"`
	SiteID int64  `json:"site_id"`
	Name   string `json:"name"`
	Port   int    `json:"port"`
}
