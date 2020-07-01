package driver

import (
	"context"
	"fmt"
	"github.com/pkg/errors"
	"io"
	"log"
	"strconv"

	evClient "github.com/CamusEnergy/kinney/controller/chargepoint/api"
	evModels "github.com/CamusEnergy/kinney/controller/chargepoint/api/schema"
)

type EVChargeClient struct {
	client EVChargePointClient
}

// NewEVChargeClient returns an EVChargeClient for a single evCharger
func NewEVChargeClient(addr string, key string, password string, writer io.Writer/*logger.LoggingClient*/) *EVChargeClient {
	c := EVChargeClient{client: evClient.NewClient(addr,key, password, writer)}
	return &c
}

// getStationGroupLoad makes an client GetLoadRequest request to the evCharger
func (c *EVChargeClient) getStationGroupLoad(device string) (*evModels.GetLoadResponse, error) {
	sgId, err := strconv.ParseInt(device, 10, 32)
	if err != nil {
		return nil, err
	}
	request := &evModels.GetLoadRequest{StationGroupID: int32(sgId)}

	resp, err := c.client.GetLoad(context.Background(), request)
	if err != nil {
		return nil, fmt.Errorf("error occured %s", err)
	}
	if resp.ResponseCode != "100" {
		log.Printf("Code: %s, Msg: %s ", resp.ResponseCode, resp.ResponseText)
		return nil, errors.New(resp.ResponseText)
	}

	return resp, nil
}

// ClearShed
func (c *EVChargeClient) ClearShed(sgId int32) error {
	request := &evModels.ClearShedStateRequest{StationGroupID: &sgId}
	resp, err := c.client.ClearShedState(context.Background(), request)
	if err != nil {
		return fmt.Errorf("error occured %s", err)
	}
	if !resp.Success {
		log.Printf("Code: %s, Msg: %s ", resp.ResponseCode, resp.ResponseText)
		return errors.New(resp.ResponseText)
	}
	return nil
}

// ShedLoadByAllowedLoad
func (c *EVChargeClient) ShedLoadByAllowedLoad(deviceName string, allowedLoad string) error {
	sgId, err := strconv.ParseInt(deviceName, 10, 32)
	if err != nil {
		return err
	}
	request := &evModels.ShedLoadRequest{StationGroupID: int32(sgId), StationGroupAllowedLoadKW: allowedLoad}
	resp, err := c.client.ShedLoad(context.Background(), request)
	if err != nil {
		return fmt.Errorf("error occured %s", err)
	}
	if resp.Success != 1 {
		log.Printf("Code: %s, Msg: %s ", resp.ResponseCode, resp.ResponseText)
		return errors.New(resp.ResponseText)
	}
	return nil
}

// ShedLoadByAllowedLoad
func (c *EVChargeClient) ShedLoadByPercentage(deviceName string, percentage int32) error {
	//sgId, err := strconv.ParseInt(deviceName, 10, 32)
	//if err != nil {
	//	return err
	//}
	//request := &evModels.ShedLoadRequest{StationGroupID: int32(sgId), StationGroupPercentShed: &percentage}
	//client := client.NewClient(c.address, c.key, c.password, c.writer)
	//resp, err := client.ShedLoad(context.Background(), request)
	//if err != nil {
	//	return fmt.Errorf("error occured %s", err)
	//}
	//if resp.Success != 1 {
	//	log.Printf("Code: %s, Msg: %s ", resp.ResponseCode, resp.ResponseText)
	//	return errors.New(resp.ResponseText)
	//}
	return nil
}