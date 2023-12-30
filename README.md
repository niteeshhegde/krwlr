# Krwlr

Krwlr is a web page parsing tool that retrieves URLs from web pages based on a given URL and depth parameter.

## Features

- Lists the nested urls belonging to same domain.
- Retries if timeout

## TODO
- Dynamically retry with exponental time-out
- configure parallel parsers and timeout via paramers

## Usage

To use Krwlr, you must have Go installed on your system. Follow the steps below:

1. Update the PATH environment variable to include the Go bin directory.
2. Clone the Krwlr repository to your local machine.
3. Open a terminal and navigate to the project directory.
4. Run go build 
5. Update path to include the binary
6. use ```krwlr --url "https://example.com" --depth 2```


## Architecture
Implemented via go routines and channels to fetch urls in parallel. 
Configured number of workers spin up and parse the urls in links channel.

`net/http` package does the web page fetch and sends the parsed HTML dom to `parseLinks`.
`parseLinks` filters for `a` `href` attributes and sends the valid lins to links channel
`parseLinks` recursively calls itself for all child nodes.

`CrawlWebpage` has the workers listening to links channel, validates new links and updates a map. Links are sorted and returned to the user.
`CrawlWebpage` also listens to erros on error channel, but doesnt return immediately. All the parsed urls and depth is printed first. If there are more than 1 errors, last error is returned along with no links. (This feature of not retuning of errors while continuing parsing links is still under consideration) 