package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"net/http"
	"slices"
	"sync"
	"time"

	"crawler/constant"
	"crawler/logger"

	"golang.org/x/net/html"
)

// CrawlWebpage craws the given rootURL looking for <a href=""> tags
// that are targeting the current web page, either via an absolute url like http://mysite.com/mypath or by a relative url like /mypath
// and returns a sorted list of absolute urls  (eg: []string{"http://mysite.com/1","http://mysite.com/2"})

// Link is a struct that represents a link in a webpage
type Link struct {
	URL   string
	Depth int
}

func CrawlWebpage(rootURL string, maxDepth int) ([]string, error) {
	var (
		wg        sync.WaitGroup
		linksChan = make(chan Link, 5)
		errorChan = make(chan error, 0)
		links     = make([]string, 0)
		err       error
		visited   = make(map[string]bool)
		mapMutex  = sync.RWMutex{}
	)

	// validating the input parameters
	if len(rootURL) < 10 || maxDepth == 0 {
		logger.LogWarn("invalid input parameters")
		return nil, nil
	}

	if rootURL[:7] != "http://" && rootURL[:8] != "https://" {
		logger.LogWarn("invalid root url. please provide a valid url starting with http:// or https://")
		return nil, nil
	}

	wg.Add(1)
	go func() {
		defer wg.Done()
		getWebPageAndParse(rootURL, 1, linksChan, &wg)
	}()

	go func() {
		for errI := range errorChan {
			// TODO: clarity needed if we need to send no list if any occurance of error on any page or send the list of links till the error occured
			// For now, code returns the last error message and prints all the urls parsed
			err = errI
			logger.LogError("error ", err.Error())
		}
	}()

	// running parallel workers to receive links from already parsed web pages. This may not be needed if the load is not high as we fetch the web page in a separate go routine.
	for i := 0; i < constant.ParallelWorkerCount; i++ {
		go func() {
			for link := range linksChan {
				if link.URL[0] == '/' {
					link.URL = rootURL + link.URL // need root urls to be concatenated with relative urls as expected
				} else if len(link.URL) <= len(rootURL) || link.URL[:len(rootURL)] != rootURL { // ignoring urls that are not part of the root url
					wg.Done()
					continue
				}

				mapMutex.RLock() // locking the map to avoid race condition from multiple workers trying to access the same key
				_, ok := visited[link.URL]
				mapMutex.RUnlock()

				if ok {
					wg.Done()
					continue
				}

				mapMutex.Lock()
				visited[link.URL] = true
				mapMutex.Unlock()

				logger.LogInfo("Crawled depth ", link.Depth, " url - ", link.URL)

				if link.Depth < maxDepth {
					go fethChildLinks(link, linksChan, errorChan, &wg)
				} else {
					wg.Done()
				}
			}
		}()
	}

	wg.Wait()
	close(linksChan)
	close(errorChan)

	if err != nil {
		// TODO: Ideally, we should return the list of links till the error occured, but as we cannot ignore the links parsed so far incase of any error more clarity is needed on this
		// sending empty list with error as of now to comply with the test cases
		return links, err
	}

	links = append(links, rootURL)
	for key, _ := range visited {
		links = append(links, key)
	}
	slices.Sort(links)
	return links, err
}

func fethChildLinks(link Link, linkChan chan Link, errorChan chan (error), wg *sync.WaitGroup) {
	defer wg.Done()
	err := getWebPageAndParse(link.URL, link.Depth+1, linkChan, wg)
	if err != nil {
		errorChan <- err
		return
	}
}

func getWebPageAndParse(url string, depth int, linkChan chan Link, wg *sync.WaitGroup) error {
	retryCount := 0
	resp, err := http.Get(url)

	if err != nil {
		if err.Error() != constant.ReadOperationTimeoutError {
			logger.LogError("failed to crawl %s: %v", url, err)
			return err
		}
		// dns ratelimitting error - unable to connect to the server.
		// These errors are due to temporary ratelimiting of client ip by the cloudflare or any dns proxy. ignoring timeout errors as we retry for timeout errors.
		for err.Error() == constant.ReadOperationTimeoutError && retryCount < constant.MaxRetries {
			logger.LogWarn("retrying ", url)
			retryCount += 1
			time.Sleep(constant.SleepTimeOut * time.Millisecond)

			resp, err = http.Get(url)
			if err != nil && err.Error() != constant.ReadOperationTimeoutError {
				logger.LogError("failed to crawl %s: %v", url, err)
				return err
			}
		}

	}
	defer resp.Body.Close()

	// rate-limitted by server. These errors are due to temporary ratelimiting of client ip by the server after successful connection (could be based on client ip).
	// retrying for a maximum of constant.MaxRetries times
	for resp.StatusCode == http.StatusTooManyRequests && retryCount < constant.MaxRetries {
		retryCount += 1
		time.Sleep(constant.SleepTimeOut * time.Millisecond)
		logger.LogWarn("retrying ", url)

		resp, err = http.Get(url)
		if err != nil && err.Error() != constant.ReadOperationTimeoutError {
			logger.LogError("failed to crawl %s: %v", url, err)
			return err
		}
	}

	if resp.StatusCode != http.StatusOK {
		logger.LogError("failed to crawl %s: %s", url, resp.Status)
		err = errors.New(constant.FailedHttpRequestError)
		return err
	}

	// getting the html document from the response body
	doc, err := html.Parse(resp.Body)
	if err != nil {
		logger.LogError("failed to parse %s: %v", url, err)
		err = errors.New(constant.FailedToParseError)
		return err
	}

	parseLinks(doc, depth, linkChan, wg)
	return nil
}

func parseLinks(n *html.Node, depth int, linkChan chan Link, wg *sync.WaitGroup) {

	// parsing the html document for <a href=""> tags
	if n.Type == html.ElementNode && n.Data == "a" {
		for _, attr := range n.Attr {
			if attr.Key == "href" && attr.Val != "" && attr.Val[0] == '/' {
				wg.Add(1)
				linkChan <- Link{URL: attr.Val, Depth: depth}
			}
		}
	}

	// recursively parsing the child nodes
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		parseLinks(c, depth, linkChan, wg)
	}
}

// --- DO NOT MODIFY BELOW ---

func main() {
	const (
		defaultURL      = constant.DefaultURL
		defaultMaxDepth = constant.DefaultMaxDepth
	)
	urlFlag := flag.String("url", defaultURL, "the url that you want to crawl")
	maxDepth := flag.Int("depth", defaultMaxDepth, "the maximum number of links deep to traverse")
	flag.Parse()
	fmt.Println(`Crawling URL: "` + *urlFlag + `" to a depth of ` + fmt.Sprint(*maxDepth) + ` links`)

	links, err := CrawlWebpage(*urlFlag, *maxDepth)
	if err != nil {
		log.Fatalln("ERROR:", err)
	}
	fmt.Println("Links")
	fmt.Println("-----")
	for i, l := range links {
		fmt.Printf("%03d. %s\n", i+1, l)
	}
	fmt.Println()
}
