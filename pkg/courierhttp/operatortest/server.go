package operatortest

import (
	"context"
	"github.com/octohelm/courier/pkg/courier"
	"github.com/octohelm/courier/pkg/courierhttp"
	"github.com/octohelm/courier/pkg/courierhttp/client"
	"github.com/octohelm/courier/pkg/courierhttp/handler"
	"github.com/octohelm/courier/pkg/courierhttp/handler/httprouter"
	"net/http"
	"net/http/httptest"
)

func Serve(ctx context.Context, o courier.Operator, middlewares ...handler.Middleware) *Server {
	r := courier.NewRouter(courierhttp.Group("/"))
	r.Register(courier.NewRouter(o))

	h, err := httprouter.New(r, "test")
	if err != nil {
		panic(err)
	}

	middlewares = append([]handler.Middleware{
		func(h http.Handler) http.Handler {
			return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
				h.ServeHTTP(rw, req.WithContext(ctx))
			})
		},
	}, middlewares...)

	s := httptest.NewServer(handler.ApplyMiddlewares(middlewares...)(h))

	return &Server{
		Server: s,
	}
}

type Server struct {
	*httptest.Server

	transports []client.HttpTransport
}

func (s *Server) ApplyHttpTransport(transports ...client.HttpTransport) {
	s.transports = transports
}

func (s *Server) Do(ctx context.Context, req any, meta ...courier.Metadata) courier.Result {
	c := &client.Client{
		Endpoint:       s.URL,
		HttpTransports: s.transports[:],
	}
	return c.Do(ctx, req, meta...)
}
