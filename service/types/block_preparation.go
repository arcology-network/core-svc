package types

import (
	"github.com/HPISTechnologies/common-lib/types"
)

type BlockPreparation struct {
	NodeRole      *types.NodeRole
	SelectedTx    *[][]byte
	VertifyHeader *types.PartialHeader
}
