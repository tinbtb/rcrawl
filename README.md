# rcrawl

---
## Recursive web-crawler.

- #### To build:
```go build -o bin/rcrawl cmd/rcrawl/main.go```
- #### To run
```./bin/rcrawl --url=https://webscraper.io/test-sites/e-commerce/allinone --max_depth=3 --req_timeout_sec=5```
- #### Help
```./bin/rcrawl --help```
- #### Most closely matches:
```wget -r -np -l 3 -E -e robots=off https://webscraper.io/test-sites/e-commerce/allinone```

---
### Service description
This project presents a simple solution for an interview task and showcases common golang project structure 
and base crawling functionality. Additional features and functionality such as: 
Dockerfile, Makefile, Tests, support for replacing html links with their local alternatives 
can be added upon request.

---
### Task description
- implement recursive web-crawler of the site.
- crawler is a command-line tool that accept starting URL and destination directory
- crawler download the initial URL and look to links inside the original document (recursively)
- crawler does not walk to link outside initial url (if starting link is https://start.url/abc, then it goes to https://start.url/abc/123 and https://start.url/abc/456, but skip https://another.domain/ and https://start.url/def)
- crawler should correctly process Ctrl+C hotkey
- crawler should be parallel
- crawler should support continue to load if the destination directory already has loaded data (if we cancel the download and then continue).
