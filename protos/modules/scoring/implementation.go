package scoring

import (
	"context"
	"errors"

	"google.golang.org/grpc"

	"esgbook-software-engineer-technical-test-2024/protos/modules/scoring/generated"

	"esgbook-software-engineer-technical-test-2024/protos/core"
	"esgbook-software-engineer-technical-test-2024/protos/utils"
)

type Broker struct {
	serverAddr string
	conn       *grpc.ClientConn
	client     generated.ScoringServiceClient
}

var (
	_ generated.ScoringServiceClient = (*Broker)(nil)
	_ core.Broker                    = (*Broker)(nil)
)

func NewBroker(serverAddr string) (*Broker, error) {
	b := new(Broker)
	b.serverAddr = serverAddr

	if b.serverAddr == "" {
		return nil, errors.New("null routed upstream host")
	}

	return b, nil
}

func (b *Broker) NewConnection() (*grpc.ClientConn, error) {
	conn, err := utils.NewConnection(b.serverAddr)
	if err != nil {
		return nil, errors.New("could not open connection")
	}

	b.conn = conn
	b.client = generated.NewScoringServiceClient(b.conn)

	return b.conn, nil
}

func (b *Broker) GetAddress() string {
	return b.serverAddr
}

func (b *Broker) CalculateScores(ctx context.Context, in *generated.CalculateRequest, opts ...grpc.CallOption) (*generated.CalculateResponse, error) {
	return b.client.CalculateScores(ctx, in, opts...)
}
