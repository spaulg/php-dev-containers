package main

import (
	"context"
	"dagger.io/dagger"
	"os"
)

// ConnectDaggerClient creates a Dagger client and connects to the Dagger engine
func ConnectDaggerClient(ctx context.Context) *dagger.Client {
	client, err := dagger.Connect(ctx, dagger.WithLogOutput(os.Stdout))
	if err != nil {
		panic(err)
	}

	return client
}
