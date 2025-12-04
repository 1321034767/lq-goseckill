package main

import (
	"fmt"
	"net/http"
	"sync"
)

var sum int64 = 0

var productNum int64 = 10000

var mutex sync.Mutex

var count int64 = 0

func GetOneProduct() bool {
	mutex.Lock()
	defer mutex.Unlock()
	count += 1
	if count%100 == 0 {
		if sum < productNum {
			sum += 1
			fmt.Println(sum)
			return true
		}
	}
	return false

}

func GetProduct(w http.ResponseWriter, req *http.Request) {
	if GetOneProduct() {
		_, _ = w.Write([]byte("true"))
		return
	}
	_, _ = w.Write([]byte("false"))
	return
}

func main() {
	http.HandleFunc("/getOne", GetProduct)
	_ = http.ListenAndServe(":8084", nil)
}
