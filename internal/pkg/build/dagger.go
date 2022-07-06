package build

import (
	"context"
	"dagger.io/dagger"
	"os"
)

// ConnectDaggerClient creates a Dagger client and connects to the Dagger engine
func ConnectDaggerClient() (context.Context, *dagger.Client) {
	ctx := context.Background()

	client, err := dagger.Connect(ctx, dagger.WithLogOutput(os.Stdout))
	if err != nil {
		panic(err)
	}

	return ctx, client
}
