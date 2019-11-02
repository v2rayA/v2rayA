package quickdown

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path"
	"strconv"
	"sync"
	"time"

	"gopkg.in/cheggaaa/pb.v1"
)

type DownloadTask struct {
	url      string
	filename string

	workerCount int
	workers     []*worker

	headOnce      sync.Once
	contentLength int
	wg            sync.WaitGroup
	done          chan bool
}

func NewDefaultDownloadTask(url string) *DownloadTask {
	return NewDownloadTask(url, "", 5)
}

func NewDownlosadTaskWithWorkers(url string, workerCount int) *DownloadTask {
	return NewDownloadTask(url, "", workerCount)
}

func NewDownloadTask(url, filename string, workerCount int) *DownloadTask {
	if workerCount < 1 {
		panic(errors.New("[DownloadTask] worker count must greater than 0."))
	}

	if filename == "" {
		filename = path.Base(url)
	}

	return &DownloadTask{
		url:         url,
		workerCount: workerCount,
		filename:    filename,
		done:        make(chan bool),
	}
}

func (self *DownloadTask) getContentLength() int {
	self.headOnce.Do(func() {
		request, _ := http.NewRequest("HEAD", self.url, nil)
		request.Header.Add("User-Agent", "github.com/xlzd/godown")
		response, err := client.Do(request)
		if err != nil {
			msg := fmt.Sprintf("[DownloadTask] Check %s head fail.", self.url)
			panic(errors.New(msg))
		}

		if contentLengthStr := response.Header.Get("Content-Length"); contentLengthStr != "" {
			self.contentLength, _ = strconv.Atoi(contentLengthStr)
		} else {
			log.Print(fmt.Sprintf("[DownloadTask] Get %s content length fail, download via one gorouting.", self.url))
		}
	})
	return self.contentLength
}

func (self *DownloadTask) initWorkers() {
	cl := self.getContentLength()
	if cl <= 4096 {
		self.workerCount = 1
		log.Println("[DownloadTask] content length too small, use one goroutine.")
	} else if cl < self.workerCount*4096 {
		self.workerCount = (cl + 4095) / 4096
		log.Printf("[DownloadTask] content length %d, use %d goroutine.\n", cl, self.workerCount)
	} else {
		log.Printf("[DownloadTask] content length %d, use %d goroutine.\n", cl, self.workerCount)
	}

	persize := (cl + self.workerCount - 1) / self.workerCount
	self.wg.Add(self.workerCount)
	for i := 0; i < self.workerCount; i++ {
		min := i * persize
		max := (i+1)*persize - 1
		if max > cl {
			max = cl
		}
		filename := fmt.Sprintf("%s.part.%d.%d", self.filename, self.workerCount, i)
		self.workers = append(self.workers, newWorker(i, self.url, min, max, &self.wg, filename))
	}
}

func (self *DownloadTask) merge() {
	err := os.Rename(self.workers[0].filename, self.filename)
	if err != nil {
		panic(err)
	}

	if self.workerCount == 1 {
		return
	}

	out, err := os.OpenFile(self.filename, os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		panic(err)
	}
	defer out.Close()

	for _, w := range self.workers[1:] {
		in, err := os.Open(w.filename)
		if err != nil {
			panic(err)
		}
		if _, err = io.Copy(out, in); err != nil {
			panic(err)
		}
		in.Close()
		out.Sync()
		os.Remove(w.filename)
	}
}

func (self *DownloadTask) wait() {
	self.wg.Wait()
	close(self.done)
}

func (self *DownloadTask) showProgress() {
	pbs := make([]*pb.ProgressBar, self.workerCount)
	full := 10000

	for i := range self.workers {
		bar := pb.New(full).Prefix(fmt.Sprintf("\033[0;32m[Worker %d]\033[0m", i+1))
		//bar.Format("[=>~]")
		bar.ShowCounters = false
		bar.ShowFinalTime = true
		bar.ShowElapsedTime = true
		//bar.AlwaysUpdate = true
		//bar.SetRefreshRate(time.Millisecond * 100)
		pbs[i] = bar
	}

	pool, err := pb.StartPool(pbs...)
	if err != nil {
		panic(err)
	}

	for {
		select {
		case <-self.done:
			for _, p := range pbs {
				if !p.IsFinished() {
					p.Set(full)
					p.Finish()
				}
			}
			pool.Stop()
			return
		case <-time.After(time.Millisecond * 200):
			for i, w := range self.workers {
				if w.done {
					if !pbs[i].IsFinished() {
						pbs[i].Set(full)
						pbs[i].Finish()
					}
				} else {
					current := int(float64(full) * float64(w.downloaded) / float64(w.total))
					pbs[i].Set(current)
				}
			}
		}
	}
}

func (self *DownloadTask) Download() {
	select {
	case <-self.done:
		panic(errors.New("[DownloadTask] already downloaded."))
	default:
	}

	self.initWorkers()
	for _, w := range self.workers {
		go w.run()
	}

	go self.showProgress()
	self.wait()
	self.merge()
	log.Printf("[DownloadTask] download %s done.", self.url)
}
