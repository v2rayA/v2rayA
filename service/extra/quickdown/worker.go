package quickdown

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"sync"
)

type worker struct {
	id                int
	url               string
	filename          string
	min, max          int
	downloaded, total int
	done              bool
	wg                *sync.WaitGroup
}

func newWorker(id int, url string, min, max int, wg *sync.WaitGroup, filename string) *worker {
	return &worker{
		id:       id,
		url:      url,
		min:      min,
		max:      max,
		total:    max - min,
		wg:       wg,
		filename: filename,
	}
}

func (self *worker) run() {
	if self.done {
		panic(errors.New(fmt.Sprintf("worker(id=%d, url=%s) already done.", self.id, self.url)))
	}

	var tmpfile *os.File
	fInfo, err := os.Stat(self.filename)
	if err == nil {
		fSize := int(fInfo.Size())
		self.min += fSize
		self.downloaded = fSize

		if fSize >= self.total {
			self.setDone()
			return
		}

		tmpfile, err = os.OpenFile(self.filename, os.O_APPEND|os.O_WRONLY, 0600)
	} else {
		tmpfile, _ = os.Create(self.filename)
	}
	defer tmpfile.Close()

	request, _ := http.NewRequest("GET", self.url, nil)
	request.Header.Add("User-Agent", "github.com/xlzd/godown")
	if self.min >= 0 && self.max > 0 {
		request.Header.Add("Range", fmt.Sprintf("bytes=%d-%d", self.min, self.max))
	}

	response, err := client.Do(request)
	if err != nil {
		panic(err)
	}
	defer response.Body.Close()

	for i := 0; ; i++ {
		buf := make([]byte, 4096*16)
		n, err := response.Body.Read(buf)
		if n != 0 {
			self.downloaded += n
			tmpfile.Write(buf[:n])
		}

		if err != nil {
			if err == io.EOF {
				break
			}
			panic(err)
		}
	}

	tmpfile.Sync()
	self.setDone()
}

func (self *worker) setDone() {
	self.done = true
	self.wg.Done()
}
