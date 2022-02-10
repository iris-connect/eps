package channels

import (
	"github.com/iris-connect/eps"
	"github.com/iris-connect/eps/forms"
)

func parseConnectionRequest(request *eps.Request) (*eps.ConnectionRequest, error) {
	var connectionRequest eps.ConnectionRequest
	if params, err := forms.ConnectionRequestForm.Validate(request.Params); err != nil {
		return nil, err
	} else if err := forms.ConnectionRequestForm.Coerce(&connectionRequest, params); err != nil {
		return nil, err
	} else {
		return &connectionRequest, nil
	}
}

func parseRequestConnectionResponse(response map[string]interface{}) (*RequestConnectionResponse, error) {
	var requestConnectionResponse RequestConnectionResponse
	if params, err := RequestConnectionResponseForm.Validate(response); err != nil {
		return nil, err
	} else if err := RequestConnectionResponseForm.Coerce(&requestConnectionResponse, params); err != nil {
		return nil, err
	} else {
		return &requestConnectionResponse, nil
	}
}
