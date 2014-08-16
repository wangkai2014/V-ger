package download

import (
	"io/ioutil"
	"log"
	"net/http/httputil"
	"toutf8"
	// "bytes"
	"fmt"
	// "io"
	// "log"
	// "native"b
	"net/http"
	// "os"
	// "path/filepath"
	// "regexp"
	// "runtime"
	"strings"
	"time"
)

var NetworkTimeout time.Duration

func fetchN(req *http.Request, n int, quit chan struct{}) (resp *http.Response, err error) {
	finish := make(chan struct{})
	go func() {
		defer close(finish)

		for i := 0; i < n; i++ {
			resp, err = http.DefaultClient.Do(req)
			if err == nil {
				return
			}
			time.Sleep(100 * time.Millisecond)
		}
	}()

	select {
	case <-quit:
		cancelRequest(req)
		err = errStopFetch
		break
	case <-finish:
		break
	}

	return
}

func GetDownloadInfo(url string, readBody bool) (finalUrl string, name string, size int64, body []byte, err error) {
	return GetDownloadInfoN(url, nil, 3, readBody, nil)
}

func GetDownloadInfoN(url string, header http.Header, retryTimes int, readBody bool, quit chan struct{}) (finalUrl string, name string, size int64, body []byte, err error) {
	log.Printf("header: %v %s", header, url)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", "", 0, nil, err
	}
	if header != nil {
		req.Header = header
	}

	resp, err := fetchN(req, retryTimes, quit)
	if err != nil {
		return "", "", 0, nil, err
	}
	defer resp.Body.Close()

	if header != nil {
		data, _ := httputil.DumpRequest(req, false)
		log.Println(string(data))
		data, _ = httputil.DumpResponse(resp, false)
		log.Print(string(data))
	}

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		err = fmt.Errorf("response status code: %d", resp.StatusCode)
		return
	}

	if readBody {
		body, err = ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Print(err)
		}
	}

	finalUrl = resp.Request.URL.String()

	name, size = getFileInfo(resp.Header)
	if name == "" {
		name = getFileName(finalUrl)
	}

	if name == "" || size == 0 {
		err = fmt.Errorf("Broken resource")
	}

	name, err = toutf8.GB18030ToUTF8(name)
	if err != nil {
		log.Print(err)
	}

	name = strings.Replace(name, "/", "|", -1)
	name = strings.Replace(name, "\\", "|", -1)
	name = strings.TrimLeft(name, ".")

	return
}
