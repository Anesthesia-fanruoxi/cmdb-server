package kubePod

type PodScaleRequest struct {
	Namespace string `json:"namespace"`
	Action    bool   `json:"action"`
}

type PodBatchScaleRequest struct {
	Namespace []string `json:"namespace"`
	Action    bool     `json:"action"`
}
