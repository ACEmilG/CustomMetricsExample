package main

import (
	"fmt"
	"math/rand"
	"net/http"
	"time"
)

func doWork() {
	sleepTime := rand.Uint32() % 10
	time.Sleep(time.Duration(sleepTime) * time.Second)
}

func handler(w http.ResponseWriter, r *http.Request) {
	doWork()
	fmt.Fprintf(w, "Hello, world...")
}

func main() {
	http.HandleFunc("/", handler)
	http.ListenAndServe(":8080", nil)
}
