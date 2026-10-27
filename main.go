// Простой бэкенд: отдаёт готовые HTML-страницы (и их CSS/JS) из папки web/.
//
// Маршруты:
//
//	/              -> web/index.html
//	/models        -> web/models.html   ("красивые" URL без .html)
//	/css/style.css -> web/css/style.css (любые статические файлы)
//	всё остальное  -> web/404.html со статусом 404
package main

import (
	"flag"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"time"
)

func main() {
	addr := flag.String("addr", ":8080", "адрес сервера")
	dir := flag.String("dir", "web", "папка с HTML/CSS/JS")
	flag.Parse()

	mux := http.NewServeMux()
	// "GET /" ловит все GET- и HEAD-запросы; POST и прочие получат 405.
	mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		serveFile(w, r, *dir)
	})

	srv := &http.Server{
		Addr:              *addr,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second, // защита от "зависших" клиентов
	}

	log.Printf("сайт: http://localhost%s, файлы из папки %q", *addr, *dir)
	log.Fatal(srv.ListenAndServe())
}

// serveFile превращает URL в путь к файлу и отдаёт его.
func serveFile(w http.ResponseWriter, r *http.Request, dir string) {
	p := path.Clean(r.URL.Path)

	if p == "/" {
		p = "/index"
	}
	if path.Ext(p) == "" { // /models -> /models.html
		p += ".html"
	}

	file := filepath.Join(dir, filepath.FromSlash(p))
	if info, err := os.Stat(file); err != nil || info.IsDir() {
		notFound(w, dir)
		return
	}

	log.Printf("%s %s -> %s", r.Method, r.URL.Path, file)
	// ServeFile сам ставит Content-Type (text/html, text/css, ...),
	// поддерживает кэширование (Last-Modified) и HEAD-запросы.
	http.ServeFile(w, r, file)
}

func notFound(w http.ResponseWriter, dir string) {
	page, err := os.ReadFile(filepath.Join(dir, "404.html"))
	if err != nil {
		page = []byte("404 — страница не найдена")
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusNotFound)
	w.Write(page)
}
