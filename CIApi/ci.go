package CIApi

import (
	"context"

	"github.com/revan730/clipper-api/types"
	"github.com/revan730/clipper-api/log"
	commonTypes "github.com/revan730/clipper-common/types"
	"google.golang.org/grpc"
)

type CIClient struct {
	gClient commonTypes.CIAPIClient
	logger  log.Logger
}

func NewClient(address string, logger log.Logger) *CIClient {
	conn, err := grpc.Dial(address, grpc.WithInsecure())
	if err != nil {
		logger.Fatal("Couldn't connect to CI gRPC", err)
	}

	c := commonTypes.NewCIAPIClient(conn)
	client := &CIClient{
		gClient: c,
		logger:  logger,
	}
	return client
}

func (c *CIClient) GetBuild(buildID int64) (*commonTypes.Build, error) {
	return c.gClient.GetBuild(context.Background(),
		&commonTypes.Build{ID: buildID})
}

func (c *CIClient) GetBuildArtifact(buildID int64) (*commonTypes.BuildArtifact, error) {
	return c.gClient.GetBuildArtifact(context.Background(),
		&commonTypes.BuildArtifact{BuildID: buildID})
}

func (c *CIClient) GetAllBuilds(repoID int64, params types.BuildsQueryParams) (*commonTypes.BuildsArray, error) {
	return c.gClient.GetAllBuilds(context.Background(),
		&commonTypes.BuildsQuery{RepoID: repoID,
			Branch: params.Branch,
			Page:   int64(params.Page),
			Limit:  int64(params.Limit),
		})
}
