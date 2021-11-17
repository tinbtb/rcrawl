package urlcrawler

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/pkg/errors"
	"golang.org/x/net/html"
)

type uc struct {
	cl         *http.Client
	processing map[string]struct{}
	mu         *sync.Mutex
}

type URLCrawler interface {
	Crawl(parentCtx context.Context, urlToCrawl string, maxDepth uint) (err error)
}

func NewURLCrawler(reqTimeout time.Duration) URLCrawler {
	return &uc{
		cl: &http.Client{
			Transport: http.DefaultTransport,
			Timeout:   reqTimeout,
		},
		processing: make(map[string]struct{}),
		mu:         &sync.Mutex{},
	}
}

func checkURL(urlToParse string) (*url.URL, error) {
	if urlToParse == "" {
		return nil, errors.New("URL cannot be empty")
	}

	u, err := url.Parse(urlToParse)
	if err != nil {
		return nil, errors.Wrap(err, "incorrect URL: "+urlToParse)
	}

	return u, nil
}

func (u uc) Crawl(parentCtx context.Context, urlToCrawl string, maxDepth uint) (err error) {
	ur, err := checkURL(urlToCrawl)
	if err != nil {
		return errors.Wrap(err, "can't checkURL in Crawl")
	}

	if !ur.IsAbs() {
		return errors.New("URL should starts with http:// or https://")
	}

	baseSchema := ur.Scheme
	baseHost := ur.Host
	basePath := ur.Path

	ctx, cancel := context.WithCancel(parentCtx)
	errChan := make(chan error, 1)

	wg := &sync.WaitGroup{}
	wg.Add(1)
	go u.recursiveCrawl(ctx, cancel, wg, baseSchema, baseHost,
		basePath, urlToCrawl, errChan, maxDepth, 0)

	wg.Wait()
	close(errChan)
	err = <-errChan
	if err != nil {
		return errors.Wrap(err, "err from errChan in Crawl")
	}

	return nil
}

// nolint:gocritic // error should be passed as a pointer to catch later changes
func errHandler(ctx context.Context, cancel context.CancelFunc,
	errPtr *error, wg *sync.WaitGroup, errChan chan<- error) {
	if errPtr != nil {
		err := *errPtr
		if err != nil {
			select {
			case <-ctx.Done():
			case errChan <- err:
				cancel()
			}
		}
	}

	wg.Done()
}

func (u uc) recursiveCrawl(
	ctx context.Context, cancel context.CancelFunc, wg *sync.WaitGroup,
	baseSchema, baseHost, basePath, urlToCrawl string, errChan chan<- error,
	maxDepth, currentDepth uint,
) {
	var err error
	defer errHandler(ctx, cancel, &err, wg, errChan)

	if currentDepth > maxDepth {
		return
	}
	currentDepth++

	var ur *url.URL
	ur, err = checkURL(urlToCrawl)
	if err != nil {
		err = errors.Wrap(err, "can't checkURL in recursiveCrawl")

		return
	}

	if !ur.IsAbs() {
		if !strings.HasPrefix(ur.Path, "/") {
			// unsupported link type
			return
		}
		ur.Scheme = baseSchema
		ur.Host = baseHost
	}

	if (ur.Host == baseHost) && strings.HasPrefix(ur.Path, basePath) {
		subDirPath, fileName := path.Split(ur.Path)
		if fileName == "" {
			fileName = "index.html"
		}

		filePath := path.Join(ur.Host, subDirPath, fileName)
		u.mu.Lock()
		_, ok := u.processing[fileName]
		if ok {
			u.mu.Unlock()

			return
		}
		u.processing[fileName] = struct{}{}
		u.mu.Unlock()

		fmt.Println(ur.String(), filePath)

		var n *html.Node
		n, err = u.readHTMLNodeFromURLOrFile(ctx, filePath, ur.String())
		if err != nil {
			err = errors.Wrap(err, "can't readHTMLNodeFromURLOrFile in recursiveCrawl")

			return
		}

		if n != nil {
			// this is a html document
			links := pageLinks([]string{}, n)
			wg.Add(len(links))
			for _, link := range links {
				go u.recursiveCrawl(ctx, cancel, wg, baseSchema, baseHost, basePath,
					link, errChan, maxDepth, currentDepth)
			}
		}
	}
}

func (u uc) readHTMLNodeFromURLOrFile(
	ctx context.Context, filePath, urlToDownload string) (
	n *html.Node, err error,
) {
	var b []byte
	b, err = os.ReadFile(filePath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			// need to download the file first
			n, err = u.downloadFile(ctx, filePath, urlToDownload)
			if err != nil {
				return nil, errors.Wrap(err, "can't downloadFile in recursiveCrawl")
			}
		} else if !errors.Is(err, syscall.EISDIR) {
			// syscall.EISDIR = not supported url structure,
			// both file and dir have same names,
			// same behavior as wget
			return nil, errors.Wrap(err, "can't os.OpenFile in recursiveCrawl")
		}
	} else {
		// nolint:errcheck // if not html n is nil, error can be omitted
		n, _ = html.Parse(bytes.NewReader(b))
	}

	return n, nil
}

func (u uc) downloadFile(ctx context.Context, filePath, urlToDownload string) (
	n *html.Node, err error) {
	var req *http.Request
	req, err = http.NewRequestWithContext(ctx, http.MethodGet, urlToDownload, nil)
	if err != nil {
		return nil, errors.Wrap(err, "can't NewRequestWithContext in downloadFile")
	}

	var resp *http.Response
	resp, err = u.cl.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "can't cl.Do in downloadFile")
	}
	defer func() {
		err = resp.Body.Close()
		if err != nil {
			err = errors.Wrap(err, "can't resp.Body.Close for downloadFile")
		}
	}()

	// Could read to a buffer here to save some RAM
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.Wrap(err, "can't io.ReadAll in downloadFile")
	}

	n, err = html.Parse(bytes.NewReader(b))
	if err == nil {
		// this is a html file
		if !strings.HasSuffix(filePath, "html") &&
			!strings.HasSuffix(filePath, "htm") {
			filePath += ".html"
		}
	}

	var tempFile *os.File
	tempFile, err = os.CreateTemp("", "rcrawl_*")
	if err != nil {
		return nil, errors.Wrap(err, "can't CreateTemp in downloadFile")
	}
	defer func() {
		err = tempFile.Close()
		if err != nil {
			err = errors.Wrap(err, "can't tempFile.Close in downloadFile")
		}
	}()

	_, err = tempFile.Write(b)
	if err != nil {
		return nil, errors.Wrap(err, "can't file.Write in downloadFile")
	}

	// returns nil if already existed
	err = os.MkdirAll(path.Dir(filePath), os.ModePerm)
	if err != nil {
		if errors.Is(err, syscall.ENOTDIR) {
			// syscall.ENOTDIR = not supported url structure,
			// both file and dir have same names,
			// same behavior as wget
			return nil, nil
		}

		return nil, errors.Wrap(err, "can't MkdirAll in recursiveCrawl")
	}

	err = os.Rename(tempFile.Name(), filePath)
	if err != nil {
		return nil, errors.Wrap(err, "can't os.Rename in downloadFile")
	}

	return n, nil
}

// pageLinks will recursively scan a `html.Node` and will return
// a list of links found, with no duplicates
func pageLinks(links []string, n *html.Node) []string {
	if n.Type == html.ElementNode && n.Data == "a" {
		for _, a := range n.Attr {
			if a.Key == "href" {
				if !sliceContains(links, a.Val) {
					links = append(links, a.Val)
				}
			}
		}
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		links = pageLinks(links, c)
	}

	return links
}

// sliceContains returns true if `slice` contains `value`
func sliceContains(slice []string, value string) bool {
	for _, v := range slice {
		if v == value {
			return true
		}
	}

	return false
}
