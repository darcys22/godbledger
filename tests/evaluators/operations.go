package evaluators

import (
	"context"

	pb "github.com/darcys22/godbledger/proto"

	"github.com/darcys22/godbledger/tests/types"

	"google.golang.org/grpc"
)

// SingleTransaction submits a single transaction to the server and expects no errors as a response
var SingleTransaction = types.Evaluator{
	Name:       "single_transaction",
	Evaluation: singleTransaction,
}

func singleTransaction(conns ...*grpc.ClientConn) error {
	client := pb.NewTransactorClient(conns[0])
	req := &pb.VersionRequest{
		Message: "Test",
	}
	_, err := client.NodeVersion(context.Background(), req)
	if err != nil {
		return err
	}

	return nil
}
