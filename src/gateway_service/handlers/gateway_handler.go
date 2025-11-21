package handlers

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"time"

	"apps-hosting.com/messaging"
	"github.com/gorilla/mux"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"

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
	return otelhttp.NewHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		handler.Logger.LogInfo(fmt.Sprintf("%s %s %d %s", r.Method, r.RequestURI, http.StatusOK, time.Since(start)))
	}),
		"",
		otelhttp.WithSpanNameFormatter(func(operation string, r *http.Request) string {
			route := mux.CurrentRoute(r)
			path := r.URL.Path
			if route != nil {
				if tmpl, err := route.GetPathTemplate(); err == nil {
					path = tmpl
				}
			}

			return fmt.Sprintf("%s %s", strings.ToLower(r.Method), path)
		}))
}
