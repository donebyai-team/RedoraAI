package portal

import (
	"context"
	pbportal "github.com/shank318/doota/pb/doota/portal/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"testing"
	"time"

	"google.golang.org/protobuf/types/known/emptypb"
)

func TestConnectRedditStream(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	conn, err := grpc.DialContext(ctx, "localhost:8787",
		grpc.WithTransportCredentials(insecure.NewCredentials()), // for plaintext h2c
		grpc.WithBlock(),
	)
	if err != nil {
		t.Fatalf("failed to connect: %v", err)
	}
	defer conn.Close()

	client := pbportal.NewPortalServiceClient(conn)

	stream, err := client.ConnectReddit(ctx, &emptypb.Empty{})
	if err != nil {
		t.Fatalf("failed to start stream: %v", err)
	}

	// Receive messages until error or context timeout
	for {
		resp, err := stream.Recv()
		if err != nil {
			// The stream ended or error occurred
			if err == context.DeadlineExceeded || err == context.Canceled {
				t.Logf("stream ended: %v", err)
				break
			}
			t.Fatalf("error receiving from stream: %v", err)
		}

		t.Logf("received response: %+v", resp)

		// Optional: you can add assertions here about the content of resp
		// For example:
		// if urlMsg := resp.GetUrl(); urlMsg != nil {
		//	   t.Logf("Live URL: %s", urlMsg.Url)
		// }
	}
}
