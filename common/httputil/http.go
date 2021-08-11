package httputil

import (
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	fp "path/filepath"
	"strings"
)

var (
	client = &http.Client{}
)

// HTTPDoPost http post
func HTTPDoPost(body interface{}, url string) ([]byte, error) {
	jsonBody, err := json.Marshal(body)
	if err != nil {
		return nil, errors.Wrap(err, "marshal body")
	}
	request, err := http.NewRequest("POST", url, strings.NewReader(string(jsonBody)))
	if err != nil {
		return nil, errors.Wrapf(err, "new http request, url [%s]", url)
	}
	res, err := client.Do(request)
	if err != nil {
		return nil, errors.Wrap(err, "do http post")
	}
	if res == nil {
		return nil, errors.New("http response is nil")
	}
	defer res.Body.Close()

	data, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, errors.Wrap(err, "read response body")
	}
	if res.StatusCode != http.StatusOK {
		return nil, errors.Errorf("http response code : [%d], data : [%s]", res.StatusCode, string(data))
	}
	return data, nil
}

// HTTPDoGet http get
func HTTPDoGet(url string) ([]byte, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, errors.Wrapf(err, "new get request, url [%s]", url)
	}
	response, err := client.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "do http get")
	} else if response == nil {
		return nil, errors.New("http response is nil")
	}
	defer response.Body.Close()
	data, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, errors.Wrap(err, "read response body")
	}
	if response.StatusCode != http.StatusOK {
		return nil, errors.Errorf("http response code : [%d], data : [%s]", response.StatusCode, string(data))
	}
	return data, nil
}

// HTTPDoUploadFile http upload file
func HTTPDoUploadFile(filepath string, url string) error {
	file, err := os.Open(filepath)
	if err != nil {
		return err
	}
	defer file.Close()

	res, err := http.Post(url, "binary/octet-stream", file) // 第二个参数用来指定 "Content-Type"
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("http response code : %d", res.StatusCode)
	}
	return nil
}

// HTTPDoDownloadFile download file through http
// saveFilepath : path + target file name, if target file not exist, create it
// url : remote download file url
func HTTPDoDownloadFile(saveFilepath, url string) error {
	filepath, err := fp.Abs(saveFilepath)
	if err != nil {
		return err
	}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}
	response, err := client.Do(req)
	if err != nil {
		return err
	} else if response == nil {
		return errors.New("http response is nil")
	} else if response.StatusCode != http.StatusOK {
		return fmt.Errorf("http response code : %d", response.StatusCode)
	}
	defer response.Body.Close()
	f, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = io.Copy(f, response.Body)
	if err != nil {
		return err
	}
	_ = f.Sync()
	return nil
}

