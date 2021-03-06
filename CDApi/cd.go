package CDApi

import (
	"context"

	"github.com/revan730/clipper-api/log"
	"github.com/revan730/clipper-api/types"
	commonTypes "github.com/revan730/clipper-common/types"
	"google.golang.org/grpc"
)

type CDClient struct {
	gClient commonTypes.CDAPIClient
	log     log.Logger
}

func NewClient(address string, logger log.Logger) *CDClient {
	conn, err := grpc.Dial(address, grpc.WithInsecure())
	if err != nil {
		logger.Fatal("Couldn't connect to CD gRPC", err)
	}

	c := commonTypes.NewCDAPIClient(conn)
	client := &CDClient{
		gClient: c,
		log:     logger,
	}
	return client
}

func (c *CDClient) CreateDeployment(d *types.DeploymentMessage) error {
	protoMsg := types.ProtoFromDeploymentMsg(d)
	_, err := c.gClient.CreateDeployment(context.Background(), protoMsg)
	return err
}

func (c *CDClient) GetDeployment(deploymentID int64) (*commonTypes.Deployment, error) {
	protoMsg := &commonTypes.Deployment{
		ID: deploymentID,
	}
	return c.gClient.GetDeployment(context.Background(), protoMsg)
}

func (c *CDClient) GetAllDeployments(params types.PaginationQueryParams) (*commonTypes.DeploymentsArray, error) {
	return c.gClient.GetAllDeployments(context.Background(),
	&commonTypes.DeploymentsQuery{
		Page:   int64(params.Page),
		Limit:  int64(params.Limit),
	})
}

func (c *CDClient) DeleteDeployment(deploymentID int64) error {
	protoMsg := &commonTypes.Deployment{
		ID: deploymentID,
	}
	_, err := c.gClient.DeleteDeployment(context.Background(), protoMsg)
	return err
}

func (c *CDClient) UpdateImage(d *types.DeploymentMessage) error {
	protoMsg := types.ProtoFromDeploymentMsg(d)
	_, err := c.gClient.ChangeImage(context.Background(), protoMsg)
	return err
}

func (c *CDClient) ScaleDeployment(d *types.DeploymentMessage) error {
	protoMsg := types.ProtoFromDeploymentMsg(d)
	_, err := c.gClient.ScaleDeployment(context.Background(), protoMsg)
	return err
}

func (c *CDClient) UpdateManifest(d *types.DeploymentMessage) error {
	protoMsg := types.ProtoFromDeploymentMsg(d)
	_, err := c.gClient.UpdateManifest(context.Background(), protoMsg)
	return err
}

func (c *CDClient) GetRevision(revisionID int64) (*commonTypes.Revision, error) {
	protoMsg := &commonTypes.Revision{
		ID: revisionID,
	}
	return c.gClient.GetRevision(context.Background(), protoMsg)
}

func (c *CDClient) GetRevisions(deploymentID int64, params types.PaginationQueryParams) (*commonTypes.RevisionsArray, error) {
	return c.gClient.GetRevisions(context.Background(),
	&commonTypes.RevisionsQuery{
		DeploymentID: deploymentID,
		Page:   int64(params.Page),
		Limit:  int64(params.Limit),
	})
}
