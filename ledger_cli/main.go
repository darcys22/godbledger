/*
 *
 * Copyright 2015 gRPC authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

// Package main implements a client for Greeter service.
package main

import (
	"context"
	"log"
	"time"

	pb "github.com/darcys22/godbledger/proto"
	"google.golang.org/grpc"
)

const (
	address = "localhost:50051"
)

func main() {
	// Set up a connection to the server.
	conn, err := grpc.Dial(address, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	c := pb.NewTransactorClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	date := "2011-03-15"
	desc := "Whole Food Market"

	transactionLines := make([]*pb.LineItem, 2)

	line1Account := "Expenses:Groceries"
	line1Desc := "Groceries"
	line1Amount := int64(7500)

	transactionLines[0] = &pb.LineItem{
		Accountname: line1Account,
		Description: line1Desc,
		Amount:      line1Amount,
	}

	line2Account := "Assets:Checking"
	line2Desc := "Groceries"
	line2Amount := int64(-7500)

	transactionLines[1] = &pb.LineItem{
		Accountname: line2Account,
		Description: line2Desc,
		Amount:      line2Amount,
	}

	req := &pb.TransactionRequest{
		Date:        date,
		Description: desc,
		Lines:       transactionLines,
	}
	r, err := c.AddTransaction(ctx, req)
	if err != nil {
		log.Fatalf("could not greet: %v", err)
	}
	log.Printf("Version: %s", r.GetMessage())
}
