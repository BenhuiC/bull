package worker

// todo api
// todo queue attributes


import (
	"{{ .ProjectName }}/pkg/log"
)

type Worker interface {
	WorkerName() string
	Perform(ctx Context, logger *log.Logger) error
}