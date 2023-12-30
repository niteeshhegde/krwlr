package constant

const (
	// ParallelWorkerCount is the number of parallel workers to receieve links from already parsed web pages
	ParallelWorkerCount = 5

	// MaxLinksChannelBufferSize is the maximum number of links that can be in the links channel at any given time.
	// Set based on device bandwidth and io/network socket
	LinksChannelBufferSize = 5

	// SleepTimeOut in milliseconds when 429 (max requests) is returned from a site, we will sleep for this amount of time before trying again
	SleepTimeOut = 200

	// MaxRetries is the number of times we will retry a site if we get a 429 (max requests) error
	MaxRetries = 10

	// ReadOperationTimeoutError is the error message returned when a read operation times out
	ReadOperationTimeoutError = "read: operation timed out"

	// DefaultURL is the default url to crawl if none is provided
	DefaultURL = "https://www.example.com/"

	// DefaultMaxDepth is the default max depth to crawl if none is provided
	DefaultMaxDepth = 3

	// FailedToParseError is the error message returned when we fail to parse a web page
	FailedToParseError = "Failed to parse"

	// FailedHttpRequestError is the error message when status code is not 200
	FailedHttpRequestError = "Http request failed"
)
