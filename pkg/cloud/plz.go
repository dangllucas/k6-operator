package cloud

import (
	"fmt"

	"go.k6.io/k6/cloudapi"
)

func RegisterPLZ(client *cloudapi.Client, data PLZRegistrationData) error {
	// url := fmt.Sprintf("https://%s/v1/load-zones", client.GetURL())
	url := fmt.Sprintf("http://%s/v1/load-zones", "mock-cloud.k6-operator-system.svc.cluster.local:8080")

	req, err := client.NewRequest("POST", url, data)
	if err != nil {
		return err
	}

	var resp struct {
		Error struct {
			Message string `json:"message"`
		} `json:"error"`
	}
	if err = client.Do(req, &resp); err != nil {
		return fmt.Errorf("Received error `%s`. Message from server `%s`", err.Error(), resp.Error.Message)
	}

	return nil
}

func DeRegisterPLZ(client *cloudapi.Client, name string) error {
	// url := fmt.Sprintf("https://%s/v1/load-zones/%s", client.GetURL(), name)
	url := fmt.Sprintf("http://%s/v1/load-zones/%s", "mock-cloud.k6-operator-system.svc.cluster.local:8080", name)

	req, err := client.NewRequest("DELETE", url, nil)
	if err != nil {
		return err
	}

	return client.Do(req, nil)
}
