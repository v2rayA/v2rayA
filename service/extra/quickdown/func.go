package quickdown

import (
	"errors"
	"log"
	"net/http"
)

var client *http.Client

func Download(url string) error {
	return download(url, "", 5)
}

func DownloadTo(url, filename string) error {
	return download(url, filename, 5)
}

func DownloadWithWorkers(url string, workerCount int) error {
	return download(url, "", workerCount)
}

func DownloadWithWorkersTo(url string, workerCount int, filename string) error {
	return download(url, filename, workerCount)
}

func download(url, filename string, workerCount int) (err error) {
	defer func() {
		if r := recover(); r != nil {
			switch x := r.(type) {
			case string:
				err = errors.New(x)
			case error:
				err = x
			default:
				err = errors.New("Unknown panic")
			}
			log.Println(err)
		}
	}()
	NewDownloadTask(url, filename, workerCount).Download()
	return
}

func SetHttpClient(c *http.Client) {
	client = c
}

func GetHttpClient() (c *http.Client) {
	return client
}
