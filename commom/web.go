package common

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

var client *http.Client

//InitWebClient 创建与后台连接web client 实例
func InitWebClient() {
	client = &http.Client{}
}

//HTTPGetData send http request to get data
func HTTPGetData(url string) ([]byte, error) {
	fmt.Println("url is :", url)
	result, err := Get(url)
	if err != nil {
		return result, err
	}

	return result, nil
}

//HTTPPostData send http request
func HTTPPostData(body []byte, url string) ([]byte, error) {
	fmt.Println("url is :", url)
	result, err := Post(body, url)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func retry(attempts int, sleep time.Duration, f func() ([]byte, error)) ([]byte, error) {
	ret, err := f()
	if err != nil {
		if attempts--; attempts > 0 {
			time.Sleep(sleep)
			return retry(attempts, sleep, f)
		}
	}
	return ret, err
}

func Post(body []byte, url string) ([]byte, error) {

	req, err := http.NewRequest("POST", url, bytes.NewBuffer([]byte(body)))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded; charset=utf-8")

	return retry(1, time.Second, func() ([]byte, error) { //尝试发送三次请求
		resp, err := client.Do(req)

		if err != nil {
			return nil, err
		}

		if resp.StatusCode != 200 {
			log.Println("status not 200, is", resp.StatusCode)
			result, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				log.Println("read body error :", err)
			} else {
				log.Println("result is", string(result))
			}

			return result, errors.New("resp status code error")
		}

		defer func() {
			err = resp.Body.Close()
			if err != nil {
				log.Println("close resp error :", err)
			}
		}()
		return ioutil.ReadAll(resp.Body)
	})

}

func Get(url string) ([]byte, error) {

	req, err := http.NewRequest("GET", url, nil)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded; charset=utf-8")

	if err != nil {
		return nil, err
	}

	return retry(3, time.Second, func() ([]byte, error) { //尝试发送三次请求
		resp, err := client.Do(req)

		if err != nil {
			return nil, err
		}

		if resp.StatusCode != 200 {
			log.Println("status not 200, is", resp.StatusCode)
			result, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				log.Println("read body error :", err)
			} else {
				log.Println("result is", string(result))
			}

			return result, errors.New("resp status code error")
		}

		fmt.Println("body is :", resp.Body)
		defer func() {
			err = resp.Body.Close()
			if err != nil {
				log.Println("close resp error :", err)
			}
		}()
		return ioutil.ReadAll(resp.Body)
	})

}
