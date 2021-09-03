package parsers

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/chainbing/node/common"
	"github.com/chainbing/tracerr"
)

// AtomicGroupFilter struct for filtering atomic group request
type AtomicGroupFilter struct {
	ID string `uri:"id" binding:"required"`
}

// ParseParamAtomicGroupID func for parsing AtomicGroupID
func ParseParamAtomicGroupID(c *gin.Context) (common.AtomicGroupID, error) {
	var atomicGroupFilter AtomicGroupFilter
	if err := c.ShouldBindUri(&atomicGroupFilter); err != nil {
		return common.AtomicGroupID{}, tracerr.Wrap(err)
	}

	atomicGroupID, err := common.NewAtomicGroupIDFromString(atomicGroupFilter.ID)
	if err != nil {
		return common.AtomicGroupID{}, tracerr.Wrap(fmt.Errorf("invalid id"))
	}

	return atomicGroupID, nil
}
