package http

import (
	"fmt"
	"sync"
	"testing"
)

type testData struct {
	Errcode int `json:"errcode"`
	Data    []struct {
		Avatar      string `json:"avatar"`
		Comments    string `json:"comments"`
		CourseType  string `json:"courseType"`
		Description string `json:"description"`
		Grade       int    `json:"grade"`
		ID          int    `json:"id"`
		Level       int    `json:"level"`
		Location    bool   `json:"location"`
		Locked      bool   `json:"locked"`
		Name        string `json:"name"`
		Paid        bool   `json:"paid"`
	} `json:"data"`
}

func TestMiddlewares(t *testing.T) {
	UseMiddleware(NewTestMiddlewares())
	var data testData
	rep := NewFastRequest("https://api.testing.com/testing-logic/v1/demoapi?userId=123456").Get().ToJSON(&data)
	t.Log(rep)
	t.Log(data)
}

func NewTestMiddlewares() Handler {
	return func(middle Middleware) {
		fmt.Println("开始")
		middle.Next()
		fmt.Println("结束")
	}
}

func TestH1cPress(t *testing.T) {
	UseMiddleware(NewTestMiddlewares())
	var wait sync.WaitGroup
	for i := 0; i < 100; i++ {
		wait.Add(1)
		go func() {
			var data testData
			req := NewFastRequest("https://api.testing.com/testing-logic/v1/demoapi?userId=123456").Get()
			//req.Singleflight("hahahah", "uuuua", "fff")
			rep := req.ToJSON(&data)
			if rep.Error != nil {
				panic(rep.Error)
			}
			if data.Data[0].Grade != 6 {
				panic("?????????????")
			}
			fmt.Println(data)
			wait.Done()
		}()
	}
	wait.Wait()
}
