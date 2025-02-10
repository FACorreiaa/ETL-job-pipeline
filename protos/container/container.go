package container

import (
	score "esgbook-software-engineer-technical-test-2024/protos/modules/scoring/generated"
	"esgbook-software-engineer-technical-test-2024/protos/utils"
)

type Brokers struct {
	Score          score.ScoringServiceServer
	TransportUtils *utils.TransportUtils
}

func NewBrokers(transportUtils *utils.TransportUtils) *Brokers {
	if transportUtils == nil {
		return nil
	}

	brokers := &Brokers{}
	brokers.TransportUtils = transportUtils
	utils.Transport = transportUtils

	return brokers
}
