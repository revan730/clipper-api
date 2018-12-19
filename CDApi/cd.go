package CDApi

import (
	"context"
	"fmt"

	"github.com/revan730/clipper-api/types"
	commonTypes "github.com/revan730/clipper-common/types"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

type CDClient struct {
	gClient commonTypes.CDAPIClient
	logger  *zap.Logger
}

func NewClient(address string, logger *zap.Logger) *CDClient {
	conn, err := grpc.Dial(address, grpc.WithInsecure())
	if err != nil {
		panic(fmt.Sprintf("Couldn't connect to CD gRPC: %s", err))
	}

	c := commonTypes.NewCDAPIClient(conn)
	client := &CDClient{
		gClient: c,
		logger:  logger,
	}
	return client
}

func (c *CDClient) CreateDeployment(d *types.DeploymentMessage) error {
	protoMsg := types.ProtoFromDeploymentMsg(d)
	_, err := c.gClient.CreateDeployment(context.Background(), protoMsg)
	return err
}
