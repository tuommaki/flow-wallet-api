package handlers

import (
	"log"
	"net/http"

	"github.com/flow-hydraulics/flow-wallet-api/templates"
)

// Templates is a HTTP server for template management.
type Templates struct {
	log     *log.Logger
	service *templates.Service
}

func NewTemplates(l *log.Logger, service *templates.Service) *Templates {
	return &Templates{l, service}
}

func (s *Templates) AddToken() http.Handler {
	h := http.HandlerFunc(s.AddTokenFunc)
	return UseJson(h)
}

func (s *Templates) ListTokens(tType templates.TokenType) http.Handler {
	return s.MakeListTokensFunc(tType)
}

func (s *Templates) GetToken() http.Handler {
	return http.HandlerFunc(s.GetTokenFunc)
}

func (s *Templates) RemoveToken() http.Handler {
	return http.HandlerFunc(s.RemoveTokenFunc)
}
