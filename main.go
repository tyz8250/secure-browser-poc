package main

import (
	"context"
	"fmt"
	"log"

	"github.com/moby/moby/client"
)

func main() {
	// Dockerへ処理を依頼するためのcontextを作る
	ctx := context.Background()

	// Docker Engineへ接続するclientを作る
	apiClient, err := client.New(client.FromEnv)
	if err != nil {
		log.Fatal(err)
	}

	// Docker clientを最後に閉じる

	// Docker EngineへPingする
	pingResult, err := apiClient.Ping(ctx, client.PingOptions{})

	if err != nil {
		log.Fatal(err)
	}

	// Pingに成功したことを表示する
	fmt.Println("Docker API Version:", pingResult.APIVersion)
	fmt.Println("OS Type:", pingResult.OSType)

	// Docker create
	resp, err := apiClient.ContainerCreate(
		ctx, client.ContainerCreateOptions{
			Image: "nginx:alpine",
		},
	)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Container created:", resp.ID)

	// Docker start
	_, err = apiClient.ContainerStart(
		ctx,
		resp.ID,
		client.ContainerStartOptions{},
	)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Container started")
}
