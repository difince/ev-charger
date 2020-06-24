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
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"time"

	evClient "github.com/CamusEnergy/kinney/controller/chargepoint/api"
	evModels "github.com/CamusEnergy/kinney/controller/chargepoint/api/schema"

	dsModels "github.com/edgexfoundry/device-sdk-go/pkg/models"

	"github.com/edgexfoundry/go-mod-core-contracts/clients/logger"
	contract "github.com/edgexfoundry/go-mod-core-contracts/models"
)

type EVCharger struct {
	lc              logger.LoggingClient
	asyncCh         chan<- *dsModels.AsyncValues
	deviceCh        chan<- []dsModels.DiscoveredDevice
	httpLogFileName string
	httpLogWriter   io.Writer
	apiKey          string
	apiPassword     string
	address         string
	load            float64
}

// Initialize performs protocol-specific initialization for the device
// service.
func (s *EVCharger) Initialize(lc logger.LoggingClient, asyncCh chan<- *dsModels.AsyncValues, deviceCh chan<- []dsModels.DiscoveredDevice) error {
	s.lc = lc
	s.asyncCh = asyncCh
	s.deviceCh = deviceCh
	//TODO Handle Security issues
	s.httpLogFileName = "/home/difince/git/ev-charger/httpLogger.json"
	s.apiKey = ""
	s.apiPassword = ""
	s.address = ""
	//TODO
	httpLogWriter, err := os.Create(s.httpLogFileName)
	if err != nil {
		return fmt.Errorf("error creating HTTP log file: %w", err)
	}
	defer httpLogWriter.Close()
	s.httpLogWriter = httpLogWriter
	return nil
}

// HandleReadCommands triggers a protocol Read operation for the specified device.
func (s *EVCharger) HandleReadCommands(deviceName string, protocols map[string]contract.ProtocolProperties, reqs []dsModels.CommandRequest) (res []*dsModels.CommandValue, err error) {
	s.lc.Debug(fmt.Sprintf("EVCharger.HandleReadCommands: protocols: %v resource: %v attributes: %v", protocols, reqs[0].DeviceResourceName, reqs[0].Attributes))
	res = make([]*dsModels.CommandValue, 3)
	var sgLoad *evModels.GetLoadResponse
	for i, r := range reqs {
		var cv *dsModels.CommandValue
		now := time.Now().UnixNano()
		switch r.DeviceResourceName {
		case "Load":
			sgLoad, err = s.getStationGroupLoad(deviceName)
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

func (s *EVCharger) getStationGroupLoad(deviceName string) (*evModels.GetLoadResponse, error) {
	sgId, err := strconv.ParseInt(deviceName, 10, 32)
	if err != nil {
		return nil, err
	}
	request := &evModels.GetLoadRequest{StationGroupID: int32(sgId)}
	c := evClient.NewClient(s.address, s.apiKey, s.apiPassword, s.httpLogWriter)
	resp, err := c.GetLoad(context.Background(), request)
	if err != nil {
		return nil, fmt.Errorf("error occured %s", err)
	}
	if resp.ResponseCode != "100" {
		log.Printf("Code: %s, Msg: %s ", resp.ResponseCode, resp.ResponseText)
		return nil, errors.New(resp.ResponseText)
	}
	return resp, nil
}

// HandleWriteCommands passes a slice of CommandRequest struct each representing
// a ResourceOperation for a specific device resource.
// Since the commands are actuation commands, params provide parameters for the individual
// command.
func (s *EVCharger) HandleWriteCommands(deviceName string, protocols map[string]contract.ProtocolProperties, reqs []dsModels.CommandRequest,
	params []*dsModels.CommandValue) error {
	var err error

	for i, r := range reqs {
		s.lc.Info(fmt.Sprintf("EVCharger.HandleWriteCommands: protocols: %v, resource: %v, parameters: %v", protocols, reqs[i].DeviceResourceName, params[i]))
		switch r.DeviceResourceName {
		case "StationGroupId": //ClearShed
			if sgId, err := params[i].Int32Value(); err == nil {
				request := &evModels.ClearShedStateRequest{StationGroupID: &sgId}
				c := evClient.NewClient(s.address, s.apiKey, s.apiPassword, s.httpLogWriter)
				resp, err := c.ClearShedState(context.Background(), request)
				if err != nil {
					return fmt.Errorf("error occured %s", err)
				}
				if !resp.Success {
					log.Printf("Code: %s, Msg: %s ", resp.ResponseCode, resp.ResponseText)
					return errors.New(resp.ResponseText)
				}
				return nil
			}
		case "Load":
			if sgId, err := params[i].Int32Value(); err == nil {
				request := &evModels.GetLoadRequest{
					StationGroupID: sgId,
				}
				c := evClient.NewClient(s.address, s.apiKey, s.apiPassword, s.httpLogWriter)
				resp, err := c.GetLoad(context.Background(), request)
				if err != nil {
					return fmt.Errorf("error occured %s", err)
				}
				if resp.ResponseCode != "100" {
					log.Printf("Code: %s, Msg: %s ", resp.ResponseCode, resp.ResponseText)
					return errors.New(resp.ResponseText)
				}
				return nil
			}
		case "AllowedLoad": //ShedByAllowedLoad
			var allowedload string
			if allowedload, err = params[i].StringValue(); err == nil {
				sgId, err := strconv.ParseInt(deviceName, 10, 32)
				if err != nil {
					return err
				}
				request := &evModels.ShedLoadRequest{StationGroupID: int32(sgId), StationGroupAllowedLoadKW: allowedload}
				c := evClient.NewClient(s.address, s.apiKey, s.apiPassword, s.httpLogWriter)
				resp, err := c.ShedLoad(context.Background(), request)
				if err != nil {
					return fmt.Errorf("error occured %s", err)
				}
				if resp.Success != 1 {
					log.Printf("Code: %s, Msg: %s ", resp.ResponseCode, resp.ResponseText)
					return errors.New(resp.ResponseText)
				}
				return nil
			}
		case "PercentShed": //ShedByPercentage
			if percentShed, err := params[i].Int32Value(); err == nil {
				sgId, err := strconv.ParseInt(deviceName, 10, 32)
				if err != nil {
					return err
				}
				request := &evModels.ShedLoadRequest{StationGroupID: int32(sgId), StationGroupPercentShed: &percentShed}
				c := evClient.NewClient(s.address, s.apiKey, s.apiPassword, s.httpLogWriter)
				resp, err := c.ShedLoad(context.Background(), request)
				if err != nil {
					return fmt.Errorf("error occured %s", err)
				}
				if resp.Success != 1 {
					log.Printf("Code: %s, Msg: %s ", resp.ResponseCode, resp.ResponseText)
					return errors.New(resp.ResponseText)
				}
				return nil
			}
		}
	}

	return nil
}

// Stop the protocol-specific DS code to shutdown gracefully, or
// if the force parameter is 'true', immediately. The driver is responsible
// for closing any in-use channels, including the channel used to send async
// readings (if supported).
func (s *EVCharger) Stop(force bool) error {
	// Then Logging Client might not be initialized
	if s.lc != nil {
		s.lc.Debug(fmt.Sprintf("EVCharger.Stop called: force=%v", force))
	}
	return nil
}

// AddDevice is a callback function that is invoked
// when a new Device associated with this Device Service is added
func (s *EVCharger) AddDevice(deviceName string, protocols map[string]contract.ProtocolProperties, adminState contract.AdminState) error {
	s.lc.Debug(fmt.Sprintf("a new Device is added: %s", deviceName))
	return nil
}

// UpdateDevice is a callback function that is invoked
// when a Device associated with this Device Service is updated
func (s *EVCharger) UpdateDevice(deviceName string, protocols map[string]contract.ProtocolProperties, adminState contract.AdminState) error {
	s.lc.Debug(fmt.Sprintf("Device %s is updated", deviceName))
	return nil
}

// RemoveDevice is a callback function that is invoked
// when a Device associated with this Device Service is removed
func (s *EVCharger) RemoveDevice(deviceName string, protocols map[string]contract.ProtocolProperties) error {
	s.lc.Debug(fmt.Sprintf("Device %s is removed", deviceName))
	return nil
}

// Discover triggers protocol specific device discovery, which is an asynchronous operation.
// Devices found as part of this discovery operation are written to the channel devices.
func (s *EVCharger) Discover() {
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
	s.deviceCh <- res
}
