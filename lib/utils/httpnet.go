package utils

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"sync"
)

func RequestHandler(bodyRequest map[string]interface{}, url, method string) (*http.Request, error) {
	dataValues, err := json.Marshal(bodyRequest)
	if err != nil {
		return nil, err
	}

	reqBody := []byte(string(dataValues))
	request, err := http.NewRequest(method, url, bytes.NewBuffer(reqBody))
	if err != nil {
		return request, err
	}

	return request, nil
}

func ResponseHandler(request *http.Request) (map[string]interface{}, error) {
	var result map[string]interface{}

	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		return result, err
	}
	defer response.Body.Close()

	body, _ := ioutil.ReadAll(response.Body)
	log.Println("Response Status:", response.Status)
	log.Println("Response Body:", string(body))

	json.Unmarshal([]byte(string(body)), &result)
	if err != nil {
		return result, err
	}

	return result, nil
}

func DoAsyncRequest(request *http.Request, ch chan<- string, wg *sync.WaitGroup) {
	defer wg.Done()

	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		log.Println("err handle it")
	}

	defer response.Body.Close()

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		log.Println("err handle it")
	}
	log.Println("Response Status:", response.Status)
	log.Println("Response Body:", string(body))

	ch <- string(body)
}

func ResponseAsyncHandler(request *http.Request) ([]string, error) {
	// make a channel
	ch := make(chan string)
	var wg sync.WaitGroup

	// do async request with channel
	wg.Add(1)
	go DoAsyncRequest(request, ch, &wg)

	// close the channel in the background
	go func() {
		wg.Wait()
		close(ch)
	}()

	// read from channel as they come in until its closed
	var responses []string
	for res := range ch {
		responses = append(responses, res)
	}

	return responses, nil
}

func RequestHandlerEntity(entity interface{}, url, method string) (*http.Request, error) {
	dataValues, err := json.Marshal(entity)
	if err != nil {
		return nil, err
	}

	reqBody := []byte(string(dataValues))
	request, err := http.NewRequest(method, url, bytes.NewBuffer(reqBody))
	if err != nil {
		return request, err
	}

	return request, nil
}
