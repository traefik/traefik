package middlewares

import (
	"github.com/containous/traefik/types"
	"github.com/rs/cors"
)

//NewCORS returns new CORS middleware
func NewCORS(options *types.CORS) *cors.Cors {
	return cors.New(cors.Options{
		AllowedOrigins:     options.AllowedOrigins,
		AllowedMethods:     options.AllowedMethods,
		AllowedHeaders:     options.AllowedHeaders,
		ExposedHeaders:     options.ExposedHeaders,
		AllowCredentials:   options.AllowCredentials,
		MaxAge:             options.MaxAge,
		OptionsPassthrough: options.OptionsPassthrough,
	})
}
