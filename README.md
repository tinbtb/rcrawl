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
### Description
This project presents a simple solution for an interview task and showcases common golang project structure 
and base crawling functionality. Additional features and functionality such as: 
Dockerfile, Makefile, Tests, support for replacing html links with their local alternatives 
can be added upon request.
