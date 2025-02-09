package core

import "google.golang.org/grpc"

// Broker is a common that we mainly use for testing that should be implemented by any concrete broker.
type Broker interface {
	NewConnection() (*grpc.ClientConn, error)
	GetAddress() string
}
