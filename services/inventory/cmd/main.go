package main

import (
    "fmt"
    "net/http"
    "os"
)

func healthHandler(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodGet {
        http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
        return
    }
    w.Header().Set("Content-Type", "application/json")
    fmt.Fprint(w, `{"status":"ok"}`)
}

func main() {
    port := os.Getenv("PORT")
    if port == "" {
        port = "8080"
    }

    mux := http.NewServeMux()
    mux.HandleFunc("/health", healthHandler)

    addr := ":" + port
    fmt.Printf("service started on %s\n", addr)
    if err := http.ListenAndServe(addr, mux); err != nil {
        panic(err)
    }
}
