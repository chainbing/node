package coordinator

import (
	"context"

	"github.com/chainbing/node/common"
	"github.com/chainbing/node/log"
	"github.com/chainbing/node/prover"
	"github.com/chainbing/tracerr"
)

// ProversPool contains the multiple prover clients
type ProversPool struct {
	pool chan prover.Client
}

// NewProversPool creates a new pool of provers.
func NewProversPool(maxServerProofs int) *ProversPool {
	return &ProversPool{
		pool: make(chan prover.Client, maxServerProofs),
	}
}

// Add a prover to the pool
func (p *ProversPool) Add(ctx context.Context, serverProof prover.Client) {
	select {
	case p.pool <- serverProof:
	case <-ctx.Done():
	}
}

// Get returns the next available prover
func (p *ProversPool) Get(ctx context.Context) (prover.Client, error) {
	select {
	case <-ctx.Done():
		log.Info("ServerProofPool.Get done")
		return nil, tracerr.Wrap(common.ErrDone)
	case serverProof := <-p.pool:
		return serverProof, nil
	}
}
