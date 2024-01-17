package endpoints

type UpdateRequest struct {
}

type UpdateResponse struct {
	//	Documents []internal.Document `json:"documents"`
	Err string `json:"err,omitempty"`
}

type ServiceStatusRequest struct{}

type ServiceStatusResponse struct {
	Code int    `json:"status"`
	Err  string `json:"err,omitempty"`
}
