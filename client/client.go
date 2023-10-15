package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

type Data struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

var mu sync.Mutex

func main() {
	endpoint := "http://localhost:8080/api/v1/get_something?id="
	totalRequests := 30
	r := gin.Default()
	r.GET("/", func(c *gin.Context) {
		s := c.DefaultQuery("limit", "30")
		if s == "" {
			s = fmt.Sprint(totalRequests)
		}
		limit, err := strconv.Atoi(s)
		if err != nil {
			c.JSON(400, gin.H{
				"message": err.Error(),
			})
			return
		}

		var response []Data
		responseCh := make(chan Data, limit)
		for i := 0; i < limit; i++ {
			go makeApiCall(endpoint+strconv.Itoa(i), responseCh)
			response = append(response, <-responseCh)
		}
		fmt.Println(len(response))
		c.JSON(200, response)
	})
	r.GET("/ping", func(c *gin.Context) {
		s := c.DefaultQuery("limit", "30")
		if s == "" {
			s = fmt.Sprint(totalRequests)
		}
		limit, err := strconv.Atoi(s)
		if err != nil {
			c.JSON(400, gin.H{
				"message": err.Error(),
			})
			return
		}

		var response []Data
		for i := 0; i < limit; i++ {
			body, err := httpGet(endpoint + strconv.Itoa(i))
			if err != nil {
				c.JSON(400, gin.H{
					"message": err.Error(),
				})
				return
			}
			var result Data
			json.Unmarshal(body, &result)
			response = append(response, result)
		}
		fmt.Println(len(response))
		c.JSON(200, response)
	})

	r.GET("/pong", func(c *gin.Context) {
		s := c.DefaultQuery("limit", "30")
		if s == "" {
			s = "30"
		}
		limit, err := strconv.Atoi(s)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"message": err.Error(),
			})
			return
		}

		var response []Data
		var wg sync.WaitGroup

		for i := 0; i < limit; i++ {
			wg.Add(1)

			go func(id int) {
				defer wg.Done()

				body, err := httpGetWithRetry(endpoint+strconv.Itoa(id), 3)
				if err != nil {
					return
				}
				var result Data
				err = json.Unmarshal(body, &result)
				if err != nil {
					return
				}

				mu.Lock()
				response = append(response, result)
				mu.Unlock()
			}(i)
		}

		wg.Wait()

		fmt.Println(len(response))

		c.JSON(http.StatusOK, response)
	})

	r.Run(":8081")
}

func makeApiCall(url string, responseCh chan Data) {
	body, err := httpGet(url)
	if err != nil {
		log.Println(err)
		return
	}
	var result Data
	err = json.Unmarshal(body, &result)
	if err != nil {
		return
	}
	responseCh <- result
}

func httpGet(url string) ([]byte, error) {
	httpReq, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpClient := &http.Client{
		Timeout: time.Second * 90,
	}
	resp, err := httpClient.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return body, nil
}

func httpGetWithRetry(url string, maxRetries int) ([]byte, error) {
	timeout := 30 * time.Second
	for attempt := 1; attempt <= maxRetries; attempt++ {
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()

		httpReq, err := http.NewRequest(http.MethodGet, url, nil)
		if err != nil {
			return nil, err
		}
		httpReq = httpReq.WithContext(ctx)
		httpReq.Header.Set("Content-Type", "application/json")
		httpClient := &http.Client{
			Timeout: timeout,
		}
		resp, err := httpClient.Do(httpReq)
		if err != nil {
			// Log the error
			fmt.Println("Attempt", attempt, "- Error:", err)
			continue
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			// Log the error
			fmt.Println("Attempt", attempt, "- Error:", err)
			continue
		}

		return body, nil
	}

	return nil, fmt.Errorf("max retries exceeded")
}
