package driver

import (
	"context"

	"github.com/CamusEnergy/kinney/controller/chargepoint/api/schema"
)

type EVChargePointClient interface {
	GetCPNInstances(ctx context.Context, req *schema.GetCPNInstancesRequest) (*schema.GetCPNInstancesResponse, error)
	GetStations(ctx context.Context, req *schema.GetStationsRequest) (*schema.GetStationsResponse, error)
	GetStationGroups(ctx context.Context, req *schema.GetStationGroupsRequest) (*schema.GetStationGroupsResponse, error)
	ShedLoad(ctx context.Context, req *schema.ShedLoadRequest) (*schema.ShedLoadResponse, error)
	ClearShedState(ctx context.Context, req *schema.ClearShedStateRequest) (*schema.ClearShedStateResponse, error)
	GetLoad(ctx context.Context, req *schema.GetLoadRequest) (*schema.GetLoadResponse, error)
}
