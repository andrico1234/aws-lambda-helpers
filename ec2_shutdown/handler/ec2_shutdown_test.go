package main

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-lambda-go/lambdacontext"
)

func TestHandleRequest(t *testing.T) {
	ctx := context.Background()
	lc := new(lambdacontext.LambdaContext)
	ctx = lambdacontext.NewContext(ctx, lc)

	t.Run("Unable to describe instances", func(t *testing.T) {
		_, err := handleLambdaEvent(ctx)

		fmt.Println("Error", err)

		if err != nil {
			t.Fatal("")
		}
	})
}
