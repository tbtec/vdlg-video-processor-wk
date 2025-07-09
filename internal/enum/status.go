package enum

type OutputStatus string

const (
	StatusCompleted        OutputStatus = "COMPLETED"
	StatusError            OutputStatus = "ERROR"
	StatusErrorContentType OutputStatus = "CONTENT_TYPE"
	StatusErrorFileSize    OutputStatus = "FILE_SIZE"
	StatusErrorProcessing  OutputStatus = "PROCESSING_ERROR"
	StatusErrorFileCheck   OutputStatus = "FILE_CHECK_ERROR"
)

func (s OutputStatus) String() string {
	return string(s)
}
