package queue

import (
	commonTypes "github.com/revan730/clipper-common/types"
)

// Queue provides interface for message queue operations
type Queue interface {
	Close()
	PublishCIJob(jobMsg *commonTypes.CIJob) error
}
