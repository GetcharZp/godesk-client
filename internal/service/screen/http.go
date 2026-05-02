package screen

import (
	"bytes"
	"context"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"sync"
	"time"

	"godesk-client/internal/logger"
	"godesk-client/internal/service/session"

	"go.uber.org/zap"
)

const preferredHTTPPort = 55080

type HTTPService struct{}

var (
	httpServerMux sync.RWMutex
	httpServer    *http.Server
	httpBaseURL   string
)

func (s *HTTPService) EnsureRunningForControlSessions() {
	if session.CountSessionsByType("control") == 0 {
		s.Stop(context.Background())
		return
	}

	if err := s.start(); err != nil {
		logger.Error("[screen] failed to start local image server.", zap.Error(err))
	}
}

func (s *HTTPService) Stop(ctx context.Context) {
	httpServerMux.Lock()
	currentServer := httpServer
	httpServer = nil
	httpBaseURL = ""
	httpServerMux.Unlock()

	if currentServer != nil {
		_ = currentServer.Shutdown(ctx)
	}
}

func (s *HTTPService) GetSessionImageURL(sessionID string, sequence uint64) string {
	httpServerMux.RLock()
	currentBaseURL := httpBaseURL
	httpServerMux.RUnlock()

	if currentBaseURL == "" || sessionID == "" {
		return ""
	}

	params := url.Values{}
	params.Set("sessionId", sessionID)
	params.Set("t", fmt.Sprintf("%d", sequence))

	return fmt.Sprintf("%s/api/session-image?%s", currentBaseURL, params.Encode())
}

func (s *HTTPService) start() error {
	httpServerMux.Lock()
	defer httpServerMux.Unlock()

	if httpServer != nil && httpBaseURL != "" {
		return nil
	}

	listener, err := s.listen()
	if err != nil {
		return err
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/api/session-image", serveSessionImage)

	httpBaseURL = fmt.Sprintf("http://%s", listener.Addr().String())
	httpServer = &http.Server{
		Handler: withCORS(mux),
	}

	go func(localServer *http.Server, localListener net.Listener) {
		if err := localServer.Serve(localListener); err != nil && err != http.ErrServerClosed {
			logger.Error("[screen] local image server stopped unexpectedly.", zap.Error(err))
		}
	}(httpServer, listener)

	logger.Info("[screen] local image server started.", zap.String("baseURL", httpBaseURL))
	return nil
}

func (s *HTTPService) listen() (net.Listener, error) {
	preferredAddr := fmt.Sprintf("127.0.0.1:%d", preferredHTTPPort)
	listener, err := net.Listen("tcp", preferredAddr)
	if err == nil {
		return listener, nil
	}

	logger.Warn("[screen] preferred port unavailable, falling back to system-assigned port.",
		zap.Int("preferredPort", preferredHTTPPort),
		zap.Error(err))

	return net.Listen("tcp", "127.0.0.1:0")
}

func withCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func serveSessionImage(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}

	sessionID := r.URL.Query().Get("sessionId")
	if sessionID == "" {
		http.Error(w, "missing sessionId", http.StatusBadRequest)
		return
	}

	sess := session.GetSession(sessionID)
	if sess == nil {
		http.NotFound(w, r)
		return
	}

	if frame := sess.GetLastFrameData(); frame != nil {
		if frame.Codec != "jpeg" || len(frame.FrameData) == 0 {
			http.Error(w, "frame not available", http.StatusNotFound)
			return
		}

		writeImageResponse(w, r, frame.FrameData, frame.Timestamp)
		return
	}

	imageData := sess.GetLastImageData()
	if len(imageData) == 0 {
		http.NotFound(w, r)
		return
	}

	writeImageResponse(w, r, imageData, sess.UpdatedAt*1000)
}

func writeImageResponse(w http.ResponseWriter, r *http.Request, data []byte, timestampMillis int64) {
	w.Header().Set("Content-Type", "image/jpeg")
	w.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "0")

	http.ServeContent(w, r, "", time.UnixMilli(timestampMillis), bytes.NewReader(data))
}
