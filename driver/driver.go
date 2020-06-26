/*******************************************************************************
* Copyright 2020 VMware, Inc.
*
* Licensed under the Apache License, Version 2.0 (the "License");
* you may not use this file except in compliance with the License.
* You may obtain a copy of the License at
*
* http://www.apache.org/licenses/LICENSE-2.0
*
* Unless required by applicable law or agreed to in writing, software
* distributed under the License is distributed on an "AS IS" BASIS,
* WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
* See the License for the specific language governing permissions and
* limitations under the License.
*******************************************************************************/

package driver

import (
	"fmt"
	sdk "github.com/edgexfoundry/device-sdk-go/pkg/service"
	"io"
	"os"
	"sync"
	"time"

	evModels "github.com/CamusEnergy/kinney/controller/chargepoint/api/schema"

	dsModels "github.com/edgexfoundry/device-sdk-go/pkg/models"

	"github.com/edgexfoundry/go-mod-core-contracts/clients/logger"
	contract "github.com/edgexfoundry/go-mod-core-contracts/models"

	"github.com/pkg/errors"
)

var once sync.Once
var lock sync.Mutex

var evChargeClients map[string]*EVChargeClient
var driver *Driver

type Driver struct {
	lc              logger.LoggingClient
	asyncCh         chan<- *dsModels.AsyncValues
	deviceCh        chan<- []dsModels.DiscoveredDevice
	config          *configuration

	httpLogFileName string
	load            float64
}

// NewProtocolDriver initializes the singleton Driver and
// returns it to the caller
func NewProtocolDriver() *Driver {
	once.Do(func() {
		driver = new(Driver)
		evChargeClients = make(map[string]*EVChargeClient)
	})

	return driver
}

// Initialize performs protocol-specific initialization for the device
// service.
func (d *Driver) Initialize(lc logger.LoggingClient, asyncCh chan<- *dsModels.AsyncValues, deviceCh chan<- []dsModels.DiscoveredDevice) error {
	d.lc = lc
	d.asyncCh = asyncCh
	d.deviceCh = deviceCh

	config, err := loadEvChargerConfig()
	if err != nil {
		panic(fmt.Errorf("load evcharger configuration failed: %d", err))
	}
	d.config = config

	//TODO Handle Logging issues
	httpLogWriter, err := os.Create("/home/difince/git/ev-charger/httpLogger.json")
	if err != nil {
		return fmt.Errorf("error creating HTTP log file: %w", err)
	}
	defer httpLogWriter.Close()

	for _, dev := range sdk.RunningService().Devices() {
		initializeEVChargeClient(dev, config.evCharger.User, config.evCharger.Password, httpLogWriter)
	}

	return nil
}

func initializeEVChargeClient(device contract.Device, user string, password string, writer io.Writer) *EVChargeClient {
	addr := device.Protocols["HTTP"]["Address"]
	c := NewEVChargeClient(addr, user, password, writer)
	lock.Lock()
	evChargeClients[addr] = c
	lock.Unlock()
	return c
}
// HandleReadCommands triggers a protocol Read operation for the specified device.
func (d *Driver) HandleReadCommands(deviceName string, protocols map[string]contract.ProtocolProperties, reqs []dsModels.CommandRequest) (res []*dsModels.CommandValue, err error) {
	d.lc.Debug(fmt.Sprintf("Driver.HandleReadCommands: protocols: %v resource: %v attributes: %v", protocols, reqs[0].DeviceResourceName, reqs[0].Attributes))
	res = make([]*dsModels.CommandValue, len(reqs))
	addr, err := d.addrFromProtocols(protocols)
	if err != nil {
		return res, errors.Errorf("handleReadCommands: %v", err.Error())
	}

	// check for existence
	client, err := d.clientsFromAddr(addr, deviceName)
	if err != nil {
		return res, errors.Errorf("handleReadCommands: %v", err.Error())
	}
	//TODO SHOULD BE CHANGED
	var sgLoad *evModels.GetLoadResponse
	for i, r := range reqs {
		var cv *dsModels.CommandValue
		now := time.Now().UnixNano()
		switch r.DeviceResourceName {
		case "Load":
			sgLoad, err = client.getStationGroupLoad(deviceName)
			if err != nil {
				return nil, err
			}
			cv = dsModels.NewStringValue(r.DeviceResourceName, now, sgLoad.StationGroupLoadKW)
		case "GroupName":
			cv = dsModels.NewStringValue(r.DeviceResourceName, now, sgLoad.StationGroupName)
		case "NumStations":
			cv, err = dsModels.NewInt32Value(r.DeviceResourceName, now, sgLoad.StationGroupNumStations)
			if err != nil {
				return nil, err
			}
		}
		res[i] = cv
	}
	return
}

// HandleWriteCommands passes a slice of CommandRequest struct each representing
// a ResourceOperation for a specific device resource.
// Since the commands are actuation commands, params provide parameters for the individual
// command.
func (d *Driver) HandleWriteCommands(deviceName string, protocols map[string]contract.ProtocolProperties, reqs []dsModels.CommandRequest,
	params []*dsModels.CommandValue) error {
	var err error
	addr, err := d.addrFromProtocols(protocols)
	if err != nil {
		return errors.Errorf("handleReadCommands: %v", err.Error())
	}

	// check for existence
	client, err := d.clientsFromAddr(addr, deviceName)
	if err != nil {
		return errors.Errorf("handleReadCommands: %v", err.Error())
	}

	for i, r := range reqs {
		d.lc.Info(fmt.Sprintf("Driver.HandleWriteCommands: protocols: %v, resource: %v, parameters: %v", protocols, reqs[i].DeviceResourceName, params[i]))
		switch r.DeviceResourceName {
		case "StationGroupId": //ClearShed
			if sgId, err := params[i].Int32Value(); err == nil {
				return client.ClearShed(sgId)
			}
		case "AllowedLoad": //ShedByAllowedLoad
			var allowedLoad string
			if allowedLoad, err = params[i].StringValue(); err == nil {
				return client.ShedLoadByAllowedLoad(deviceName, allowedLoad)
			}
		case "PercentShed": //ShedByPercentage
			if percentShed, err := params[i].Int32Value(); err == nil {
				return client.ShedLoadByPercentage(deviceName, percentShed)
			}
			//case "Load":
			//	if sgId, err := params[i].Int32Value(); err == nil {
			//		request := &evModels.GetLoadRequest{
			//			StationGroupID: sgId,
			//		}
			//		c := evClient.NewClient(d.address, d.apiKey, d.apiPassword, d.httpLogWriter)
			//		resp, err := c.GetLoad(context.Background(), request)
			//		if err != nil {
			//			return fmt.Errorf("error occured %s", err)
			//		}
			//		if resp.ResponseCode != "100" {
			//			log.Printf("Code: %s, Msg: %s ", resp.ResponseCode, resp.ResponseText)
			//			return errors.New(resp.ResponseText)
			//		}
			//		return nil
			//	}
		}
	}

	return nil
}

// Stop the protocol-specific DS code to shutdown gracefully, or
// if the force parameter is 'true', immediately. The driver is responsible
// for closing any in-use channels, including the channel used to send async
// readings (if supported).
func (d *Driver) Stop(force bool) error {
	// Then Logging Client might not be initialized
	if d.lc != nil {
		d.lc.Debug(fmt.Sprintf("Driver.Stop called: force=%v", force))
	}
	return nil
}

// AddDevice is a callback function that is invoked
// when a new Device associated with this Device Service is added
func (d *Driver) AddDevice(deviceName string, protocols map[string]contract.ProtocolProperties, adminState contract.AdminState) error {
	d.lc.Debug(fmt.Sprintf("a new Device is added: %s", deviceName))
	return nil
}

// UpdateDevice is a callback function that is invoked
// when a Device associated with this Device Service is updated
func (d *Driver) UpdateDevice(deviceName string, protocols map[string]contract.ProtocolProperties, adminState contract.AdminState) error {
	d.lc.Debug(fmt.Sprintf("Device %s is updated", deviceName))
	return nil
}

// RemoveDevice is a callback function that is invoked
// when a Device associated with this Device Service is removed
func (d *Driver) RemoveDevice(deviceName string, protocols map[string]contract.ProtocolProperties) error {
	d.lc.Debug(fmt.Sprintf("Device %s is removed", deviceName))
	return nil
}

// Discover triggers protocol specific device discovery, which is an asynchronous operation.
// Devices found as part of this discovery operation are written to the channel devices.
func (d *Driver) Discover() {
	proto := make(map[string]contract.ProtocolProperties)
	proto["other"] = map[string]string{"Address": "simple02", "Port": "301"}

	device2 := dsModels.DiscoveredDevice{
		Name:        "Simple-Device02",
		Protocols:   proto,
		Description: "found by discovery",
		Labels:      []string{"auto-discovery"},
	}

	proto = make(map[string]contract.ProtocolProperties)
	proto["other"] = map[string]string{"Address": "simple03", "Port": "399"}

	device3 := dsModels.DiscoveredDevice{
		Name:        "Simple-Device03",
		Protocols:   proto,
		Description: "found by discovery",
		Labels:      []string{"auto-discovery"},
	}

	res := []dsModels.DiscoveredDevice{device2, device3}

	time.Sleep(10 * time.Second)
	d.deviceCh <- res
}

func (d *Driver) clientsFromAddr(addr string, deviceName string) (*EVChargeClient, error) {
	echChargeClient, ok := getEVChargeClient(addr)
	if !ok {
		//TODO if does not exists try to initialize it, and if it again fails then return an error
		err := fmt.Errorf("device not found: %s", deviceName)
		d.lc.Error(err.Error())
		return nil,  err
	}
	return echChargeClient, nil
}

func getEVChargeClient(addr string) (*EVChargeClient, bool) {
	lock.Lock()
	c, ok := evChargeClients[addr]
	lock.Unlock()
	return c, ok
}

func (d *Driver) addrFromProtocols(protocols map[string]contract.ProtocolProperties) (string, error) {
	if _, ok := protocols["HTTP"]; !ok {
		d.lc.Error("No HTTP address found for device. Check configuration file.")
		return "", fmt.Errorf("no HTTP address in protocols map")
	}

	var addr string
	addr, ok := protocols["HTTP"]["Address"]
	if !ok {
		d.lc.Error("No HTTP address found for device. Check configuration file.")
		return "", fmt.Errorf("no HTTP address in protocols map")
	}
	return addr, nil

}
