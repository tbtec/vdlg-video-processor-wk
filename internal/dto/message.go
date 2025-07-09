package dto

type Message struct {
	BucketName string
	Key        string
	Url        string
}

type OutputMessage struct {
	Filename string `json:"fileName"`
	Status   string `json:"status"` //COMPLETED, ERROR
	Reason   string `json:"reason"` //FILE_SIZE, PROCESSING_ERROR
}

type ProcessingResult struct {
	Success    bool
	Message    string
	ZipPath    string
	FrameCount int
	Images     []string
}
