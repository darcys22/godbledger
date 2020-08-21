// Package types includes important structs used by end to end tests, such
// as a configuration type, an evaluator type, and more.
package types

import (
	"google.golang.org/grpc"
)

// Evaluator defines the structure of the evaluators used to check the current server state during the E2E
type Evaluator struct {
	Name       string
	Evaluation func(conn ...*grpc.ClientConn) error // A variable amount of conns is allowed to be passed in for evaluations to check all nodes if needed.
}
