package server

import (
	"context"
	"log"
	"net/http"
	"net/http/httputil"
	"time"
)

const (
	ProxyRetry        = 3
	proxyRetryTimeout = 10 * time.Millisecond
)

func NewProxy(server *Server, lbFn func(w http.ResponseWriter, r *http.Request)) *httputil.ReverseProxy {
	proxy := httputil.NewSingleHostReverseProxy(server.URL)
	proxy.ErrorHandler = func(writer http.ResponseWriter, request *http.Request, e error) {
		log.Printf("[%s] %s\n", server.URL.Host, e.Error())
		retries := GetRetryFromContext(request)
		if retries < ProxyRetry {
			select {
			case <-time.After(proxyRetryTimeout):
				ctx := context.WithValue(request.Context(), Retry, retries+1)
				proxy.ServeHTTP(writer, request.WithContext(ctx))
			}
		} else {
			// after N retries, mark this backend as down
			server.SetAlive(false)

			// if the same request routing for new attempts with different servers, increase the count
			attempts := GetAttemptsFromContext(request)
			log.Printf("%s(%s) Attempting retry %d\n", request.RemoteAddr, request.URL.Path, attempts)
			ctx := context.WithValue(request.Context(), Attempts, attempts+1)
			lbFn(writer, request.WithContext(ctx))
		}
	}
	return proxy
}
