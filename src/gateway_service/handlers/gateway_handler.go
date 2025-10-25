package handlers

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"time"

	"apps-hosting.com/messaging"

	"apps-hosting.com/logging"

	"gateway/utils"
)

type GatewayHandler struct {
	Logger logging.ServiceLogger
}

func NewGatewayHandler(logger logging.ServiceLogger) GatewayHandler {
	return GatewayHandler{
		Logger: logger,
	}
}

func (handler *GatewayHandler) HealthCheckHandler(w http.ResponseWriter, r *http.Request) {
	messaging.WriteSuccess(w, "OK", nil)
}

func (handler *GatewayHandler) ReverseProxyHandler(serviceURL string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !utils.HealthCheck(serviceURL) {
			http.Error(w, "Service is Down", http.StatusServiceUnavailable)
			return
		}

		targetURL, err := url.Parse(serviceURL)
		if err != nil {
			http.Error(w, "Bad gateway", http.StatusBadGateway)
			return
		}

		proxy := httputil.NewSingleHostReverseProxy(targetURL)
		r.Host = targetURL.Host
		proxy.ServeHTTP(w, r)
	}
}

func (handler *GatewayHandler) LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		handler.Logger.LogInfo(fmt.Sprintf("%s %s %d %s", r.Method, r.RequestURI, http.StatusOK, time.Since(start)))
	})
}
