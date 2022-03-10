package model

type AckJob struct {
	JobId    int    `json:"job_id"`
	Username string `json:"username"`
}
