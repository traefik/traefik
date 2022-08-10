//go:build windows
// +build windows

package server

import "context"

func (s *Server) configureSignals() {}

func (s *Server) listenSignals(ctx context.Context) {}
