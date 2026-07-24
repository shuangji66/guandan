package main

import (
	"embed"
	"guandan/internal/hub"
	"io"
	"io/fs"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
)

//go:embed web
var staticFiles embed.FS

func main() {
	// 创建子文件系统，根目录为 web 目录的内容
	webFS, err := fs.Sub(staticFiles, "web")
	if err != nil {
		log.Fatalf("Failed to create sub filesystem: %v", err)
	}

	hubInstance := hub.NewHub()
	go hubInstance.Run()

	http.HandleFunc("/ws", hubInstance.HandleWebSocket)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimPrefix(r.URL.Path, "/")
		if path == "" {
			path = "index.html"
		}
		if strings.HasSuffix(path, "/") {
			http.NotFound(w, r)
			return
		}

		f, err := webFS.Open(path)
		if err != nil {
			http.NotFound(w, r)
			return
		}
		defer f.Close()

		stat, err := f.Stat()
		if err != nil || stat.IsDir() {
			http.NotFound(w, r)
			return
		}

		data, err := io.ReadAll(f)
		if err != nil {
			http.NotFound(w, r)
			return
		}

		// 根据扩展名设置 Content-Type
		ext := filepath.Ext(path)
		contentType := "text/plain"
		switch ext {
		case ".html":
			contentType = "text/html"
		case ".css":
			contentType = "text/css"
		case ".js":
			contentType = "application/javascript"
		case ".json":
			contentType = "application/json"
		case ".png":
			contentType = "image/png"
		case ".jpg", ".jpeg":
			contentType = "image/jpeg"
		case ".svg":
			contentType = "image/svg+xml"
		case ".ico":
			contentType = "image/x-icon"
		case ".woff":
			contentType = "font/woff"
		case ".woff2":
			contentType = "font/woff2"
		case ".ttf":
			contentType = "font/ttf"
		}
		w.Header().Set("Content-Type", contentType)
		w.Write(data)
	})

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	port := os.Getenv("PORT")
	if port == "" {
		port = "23000"
	}
	addr := ":" + port
	log.Printf("Server listening on %s", addr)
	server := &http.Server{Addr: addr}

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("ListenAndServe: %v", err)
		}
	}()

	<-stop
	log.Println("Shutting down server...")
}