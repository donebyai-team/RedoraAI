package server

import (
	"net/http"
	"time"

	"github.com/rs/cors"
	"go.uber.org/zap"
)

func (s *Server) corsOption() *cors.Cors {
	return cors.New(cors.Options{
		AllowedMethods: []string{
			http.MethodHead,
			http.MethodGet,
			http.MethodPost,
			http.MethodPut,
			http.MethodPatch,
			http.MethodDelete,
		},
		AllowOriginFunc: s.allowedOrigin,
		AllowedHeaders:  []string{"*"},
		ExposedHeaders: []string{
			// Content-MessageType is in the default safelist.
			"Accept",
			"Accept-Encoding",
			"Accept-Post",
			"Connect-Accept-Encoding",
			"Connect-Content-Encoding",
			"Content-Encoding",
			"Grpc-Accept-Encoding",
			"Grpc-Encoding",
			"Grpc-Message",
			"Grpc-State",
			"Grpc-State-Details-Bin",
		},
		MaxAge:           int(2 * time.Hour / time.Second),
		AllowCredentials: true,
	})
}

func (s *Server) allowedOrigin(origin string) bool {
	if s.corsURLRegexAllow == nil {
		s.logger.Warn("allowed origin, no URL regex allowed filter specify denying origin", zap.String("origin", origin))
		return false
	}

	return s.corsURLRegexAllow.MatchString(origin)
}
