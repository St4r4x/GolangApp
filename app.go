package main

import (
	"embed"
	"encoding/json"
	"io/fs"
	"net/http"
	"os"
)

//go:embed swagger-ui
var content embed.FS

func logReq(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get server identification
		port := os.Getenv("PORT")
		if port == "" {
			port = "8080"
		}
		hostname, _ := os.Hostname()
		serverID := hostname + ":" + port

		// Add server identification to response headers
		w.Header().Set("X-Server-ID", serverID)
		w.Header().Set("X-Container-Name", hostname)
		w.Header().Set("X-Server-Port", port)

		Logger.Infof("🌐 [Server: %s] New request to: '%s %s' from %s",
			serverID, r.Method, r.RequestURI, r.RemoteAddr)
		next.ServeHTTP(w, r)
	})
}

func newApp() http.Handler {
	Logger.Info("Init the backend")

	router := http.NewServeMux()
	router.HandleFunc("GET /{$}", getHomeHandler)
	router.HandleFunc("POST /api/cats", makeHandlerFunc(createCat))
	router.HandleFunc("GET /api/cats", makeHandlerFunc(listCats))
	router.HandleFunc("GET /api/cats/{catId}", makeHandlerFunc(getCat))
	router.HandleFunc("DELETE /api/cats/{catId}", makeHandlerFunc(deleteCat))

	fsys, _ := fs.Sub(content, "swagger-ui")
	router.Handle("GET /swagger/", http.StripPrefix("/swagger", http.FileServer(http.FS(fsys))))

	return logReq(router)
}

// Simpler way to handle requests
type ServiceFunc func(*http.Request) (int, any)

// Wraps the ServiceFunc to make a http.HandlerFunc with panic handling and JSON response encoding
func makeHandlerFunc(svcFunc ServiceFunc) http.HandlerFunc {

	return func(res http.ResponseWriter, req *http.Request) {

		code, body := func(req *http.Request) (code int, body any) {
			// General panic/error handler to keep the server up
			defer func() {
				if recov := recover(); recov != nil {
					Logger.Error("Recovering from a panic: ", recov)
					// Using the named return values
					code = http.StatusInternalServerError
					body = http.StatusText(code)
				}
			}()
			return svcFunc(req)
		}(req)

		// Single response
		res.Header().Set("content-type", "application/json")
		res.WriteHeader(code)
		json.NewEncoder(res).Encode(body)
	}
}
