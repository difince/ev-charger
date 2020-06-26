package driver

import (
	"fmt"
	sdk "github.com/edgexfoundry/device-sdk-go/pkg/service"
)

type configuration struct {
	evCharger evChargerInfo
}

type evChargerInfo struct {
	User       string
	Password   string
	//Address    string
}

const (
	APIUSER     = "ApiKey"
	APIPASSWORD = "ApiPassword"
	//ADDRESS     = "Address"
)

// loadEvChargerConfig loads the camera configuration
func loadEvChargerConfig() (*configuration, error) {
	config := new(configuration)
	if val, ok := sdk.DriverConfigs()[APIUSER]; ok {
		config.evCharger.User = val
	} else {
		return config, fmt.Errorf("driver config undefined: %s", APIUSER)
	}
	if val, ok := sdk.DriverConfigs()[APIPASSWORD]; ok {
		config.evCharger.Password = val
	} else {
		return config, fmt.Errorf("driver config undefined: %s", APIPASSWORD)
	}
	//if val, ok := sdk.DriverConfigs()[ADDRESS]; ok {
	//	config.evCharger.Address = val
	//} else {
	//	return config, fmt.Errorf("driver config undefined: %s", ADDRESS)
	//}

	return config, nil
}
