/*
Copyright IBM Corp. All Rights Reserved.
SPDX-License-Identifier: Apache-2.0
*/

package proposal

import (
	"fmt"

	"github.com/hyperledger/fabric-protos-go/common"
	"github.com/hyperledger/fabric-protos-go/peer"
)

func CheckSuccessfulResponse(proposalResponse *peer.ProposalResponse) error {
	response := proposalResponse.GetResponse()
	status := response.GetStatus()

	if status < int32(common.Status_SUCCESS) || status >= int32(common.Status_BAD_REQUEST) {
		return fmt.Errorf("unsuccessful response received with status %d (%s): %s", status, common.Status_name[status], response.GetMessage())
	}

	return nil
}
