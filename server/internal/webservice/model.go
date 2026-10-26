package webservice

type WebService struct {
	ID         int64  `json:"id"`
	SiteID     int64  `json:"site_id"`
	Name       string `json:"name"`
	Port       int    `json:"port"`
	WorkingDir string `json:"working_dir"`
	StartupCmd string `json:"startup_cmd"`
	EnvContent string `json:"env_content"`
	Comment    string `json:"comment"`
}
