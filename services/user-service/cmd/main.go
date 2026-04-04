package main

import (
	"fmt"
	"net/http"
	"os"
)

func main() {
	port := os.Getenv("APP_PORT")
	if port == "" {
		port = "8060"
	}

	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "user-service ok")
	})

	fmt.Println("user-service started on port", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		panic(err)
	}
}
