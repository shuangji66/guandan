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
	// 创建子文件系统，使得根目录直接为 web 目录的内容
	webFS, err := fs.Sub(staticFiles, "web")
	if err != nil {
		log.Fatalf("Failed to create sub filesystem: %v", err)
	}

	// 调试：递归列出嵌入的文件（现在根目录就是 web 下的内容）
	log.Println("Embedded files recursively:")
	fs.WalkDir(webFS, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() {
			log.Printf("  %s", path)
		}
		return nil
	})

	// 列出根目录
	entries, err := fs.ReadDir(webFS, ".")
	if err != nil {
		log.Printf("ReadDir error: %v", err)
	} else {
		log.Printf("Embedded files in root: %v", entries)
		for _, e := range entries {
			log.Printf("  %s", e.Name())
		}
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

		log.Printf("Serving: %s", path)

		// 使用 webFS 打开文件
		f, err := webFS.Open(path)
		if err == nil {
			defer f.Close()
			stat, err := f.Stat()
			if err == nil && stat.IsDir() {
				http.NotFound(w, r)
				return
			}
			data, err := io.ReadAll(f)
			if err != nil {
				log.Printf("Error reading %s: %v", path, err)
				http.NotFound(w, r)
				return
			}
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
			return
		}

		log.Printf("File not found: %s, falling back to index.html", path)
		// 使用 fs.ReadFile 从 webFS 读取 index.html
		indexData, err := fs.ReadFile(webFS, "index.html")
		if err != nil {
			log.Printf("Error reading index.html: %v", err)
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "text/html")
		w.Write(indexData)
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