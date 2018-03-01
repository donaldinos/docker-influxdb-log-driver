package commons

import "time"

// JSONLogLine -
type JSONLogLine struct {
	Message          string            `json:"message"`
	ContainerID      string            `json:"container_id"`
	ContainerName    string            `json:"container_name"`
	ContainerCreated time.Time         `json:"container_created"`
	ImageID          string            `json:"image_id"`
	ImageName        string            `json:"image_name"`
	Command          string            `json:"command"`
	Tag              string            `json:"tag"`
	Extra            map[string]string `json:"extra"`
	Host             string            `json:"host"`
}
