package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"google.golang.org/grpc/credentials/insecure"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	sparkv1 "github.com/zcubbs/spark/gen/proto/go/spark/v1"

	"google.golang.org/grpc"
)

func main() {
	var serverGrpcAddr, serverRestAddr, image, jobId, mode, cmd string
	var timeout int

	flag.StringVar(&serverGrpcAddr, "grpcAddr", "localhost:9000", "gRPC server address")
	flag.StringVar(&serverRestAddr, "restAddr", "localhost:8000", "REST server address")
	flag.StringVar(&jobId, "id", "", "Job ID (optional)")
	flag.StringVar(&image, "image", "", "Image of the job")
	flag.StringVar(&cmd, "cmd", "", "Command to run, wrapped in quotes, space-separated")
	flag.IntVar(&timeout, "timeout", 30, "Timeout in seconds")
	flag.StringVar(&mode, "mode", "grpc", "Mode of operation: 'grpc' or 'rest'")
	flag.Parse()

	if image == "" || cmd == "" {
		fmt.Println("Image and command must be specified.")
		flag.Usage()
		os.Exit(1)
	}

	commandParts := strings.Split(cmd, " ")

	if mode == "grpc" {
		doGrpcCall(serverGrpcAddr, jobId, image, commandParts, timeout)
	} else if mode == "rest" {
		doRestCall(serverRestAddr, jobId, image, commandParts, timeout)
	} else {
		log.Fatalf("Unsupported mode: %s", mode)
	}
}

func doGrpcCall(serverAddr, jobId, image string, commandParts []string, timeout int) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout+5)*time.Second)
	defer cancel()

	conn, err := grpc.DialContext(
		ctx,
		serverAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
	)
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	client := sparkv1.NewSparkServiceClient(conn)

	response, err := client.QueueJob(ctx, &sparkv1.QueueJobRequest{
		JobId:   jobId,
		Image:   image,
		Command: commandParts,
		Timeout: int32(timeout),
	})
	if err != nil {
		log.Fatalf("could not queue job: %v", err)
	}

	fmt.Printf("Job queued successfully via gRPC: %v\n", response)
}

func doRestCall(serverAddr, jobId, image string, commandParts []string, timeout int) {
	url := fmt.Sprintf("http://%s/v1/queue_job", serverAddr)
	requestBody, err := json.Marshal(map[string]interface{}{
		"job_id":  jobId,
		"image":   image,
		"command": commandParts,
		"timeout": timeout,
	})
	if err != nil {
		log.Fatalf("could not create request body: %v", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(requestBody))
	if err != nil {
		log.Fatalf("could not create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{
		Timeout: time.Duration(timeout+5) * time.Second,
	}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatalf("could not send HTTP request: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("could not read response body: %v", err)
	}

	fmt.Printf("Job queued successfully via REST: %s\n", body)
}
