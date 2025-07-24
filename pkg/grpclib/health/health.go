package health

import (
	"google.golang.org/grpc"

	healthgrpc "google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
)

// Server wraps grpc health server
type Server struct {
	server *healthgrpc.Server
}

// NewServer creates health server using default grpc health server.
func NewServer() *Server {
	return &Server{
		server: healthgrpc.NewServer(),
	}
}

// NewCustomServer creates health server using custom grpc health server.
func NewCustomServer(server *healthgrpc.Server) *Server {
	return &Server{
		server: server,
	}
}

// InitService initializes health service.
// Service name should be ServiceName defined in proto.
func (h *Server) InitService(serviceName string) {
	h.server.SetServingStatus(serviceName, healthpb.HealthCheckResponse_SERVING)
}

// Shutdown sets all serving status to NOT_SERVING.
func (h *Server) Shutdown() {
	h.server.Shutdown()
}

// Resume sets all serving status to SERVING.
func (h *Server) Resume() {
	h.server.Resume()
}

// Register registers health server.
func (h *Server) Register(grpc *grpc.Server) {
	healthpb.RegisterHealthServer(grpc, h.server)
}
