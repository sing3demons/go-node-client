package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type Data struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

var mu sync.Mutex

func main() {
	url := os.Getenv("SERVER_URL")
	if url == "" {
		url = "http://localhost:8080/api/v1/get_something"
	}
	endpoint := url + "?id="
	totalRequests := 30
	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "ok",
		})
	})
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
		var wg sync.WaitGroup
		responseCh := make(chan Data, limit)
		poolSize := 10
		semaphore := make(chan struct{}, poolSize)
		for i := 0; i < limit; i++ {
			wg.Add(1)
			go func(i int) {
				defer wg.Done()
				semaphore <- struct{}{}
				defer func() { <-semaphore }()
				makeApiCall(endpoint+strconv.Itoa(i), responseCh)
			}(i)
		}

		go func() {
			wg.Wait()
			close(responseCh)
		}()

		for data := range responseCh {
			response = append(response, data)
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

	ServeHttp(":8081", "client", r)
}

func ServeHttp(addr, serviceName string, router http.Handler) {
	srv := &http.Server{
		Addr:    addr,
		Handler: router,
	}

	go func() {
		logrus.Infof("[%s] http listen: %v", serviceName, srv.Addr)

		if err := srv.ListenAndServe(); err != nil && errors.Is(err, http.ErrServerClosed) {
			logrus.Error("server listen err: ", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logrus.Warn("shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logrus.Fatal("server forced to shutdown: ", err)
	}

	logrus.Warn("server exited")
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
