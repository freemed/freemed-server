package billing

import (
	"fmt"
	"log"

	"github.com/freemed/freemed-server/model"
	rc "github.com/freemed/remitt-server/client"
	rm "github.com/freemed/remitt-server/model"
)

// RemittClient is the local FreeMED wrapper around the REMITT base client code
type RemittClient struct {
	Client *rc.RemittClient
}

// init initializes the internal RemittClient instance with URLs, credentials, etc.
func (r *RemittClient) init() error {
	url, err := model.ConfigGetByKey("remitt_url")
	if err != nil {
		return err
	}
	user, err := model.ConfigGetByKey("remitt_user")
	if err != nil {
		return err
	}
	pass, err := model.ConfigGetByKey("remitt_pass")
	if err != nil {
		return err
	}
	r.Client, err = rc.NewClient(user.Value.String, pass.Value.String, url.Value.String)
	return err
}

// SubmitPayload submits a payload to a local REMITT instance
func (r *RemittClient) SubmitPayload(billkey int64, render, renderOption, transport, transportOption string) (int64, error) {
	payload, err := model.GetBillkeyPayload(billkey)
	if err != nil {
		log.Printf("SubmitPayload(%d): ERROR: %s", billkey, err.Error())
		return 0, err
	}
	return r.Client.PayloadInsert(rc.InputPayload{
		OriginalID:      rm.NewNullStringValue(fmt.Sprintf("%d", billkey)),
		RenderPlugin:    render,
		RenderOption:    renderOption,
		TransportPlugin: transport,
		TransportOption: transportOption,
		InputPayload:    payload,
	})
}
