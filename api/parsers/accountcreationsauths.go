package parsers

import (
	ethCommon "github.com/ethereum/go-ethereum/common"
	"github.com/gin-gonic/gin"
	"github.com/chainbing/node/common"
	"github.com/chainbing/tracerr"
)

// GetAccountCreationAuthFilter struct for parsing cbEthereumAddress from /account-creation-authorization/:cbEthereumAddress request
type GetAccountCreationAuthFilter struct {
	Addr string `uri:"cbEthereumAddress" binding:"required"`
}

// ParseGetAccountCreationAuthFilter parsing uri request to the eth address
func ParseGetAccountCreationAuthFilter(c *gin.Context) (*ethCommon.Address, error) {
	var getAccountCreationAuthFilter GetAccountCreationAuthFilter
	if err := c.ShouldBindUri(&getAccountCreationAuthFilter); err != nil {
		return nil, tracerr.Wrap(err)
	}
	return common.CbStringToEthAddr(getAccountCreationAuthFilter.Addr, "cbEthereumAddress")
}
