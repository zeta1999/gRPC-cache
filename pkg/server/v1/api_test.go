package v1

import (
	"context"
	"log"
	"net"
	"strconv"
	"testing"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/test/bufconn"

	"github.com/golang/protobuf/ptypes/empty"
	apis "github.com/knrt10/percona-cache/pkg/api/v1"
)

const (
	bufSize = 1024 * 1024
	expire  = 10
	cleanup = 5
)

var lis *bufconn.Listener

func init() {
	lis = bufconn.Listen(bufSize)
	s := grpc.NewServer()
	apis.RegisterCacheServiceServer(s, NewCacheService(time.Duration(expire)*time.Minute, time.Duration(cleanup)*time.Minute))
	go func() {
		if err := s.Serve(lis); err != nil {
			log.Fatalf("Server exited with error: %v", err)
		}
	}()
}

func bufDialer(context.Context, string) (net.Conn, error) {
	return lis.Dial()
}

func connectClient() (apis.CacheServiceClient, error) {
	ctx := context.Background()
	conn, err := grpc.DialContext(ctx, "bufnet", grpc.WithContextDialer(bufDialer), grpc.WithInsecure())
	if err != nil {
		return nil, err
	}
	defer conn.Close()
	c := apis.NewCacheServiceClient(conn)
	return c, nil
}

func TestAdd(t *testing.T) {
	ctx := context.Background()
	conn, err := grpc.DialContext(ctx, "bufnet", grpc.WithContextDialer(bufDialer), grpc.WithInsecure())
	if err != nil {
		t.Fatalf("Failed to dial bufnet: %v", err)
	}
	defer conn.Close()
	c := apis.NewCacheServiceClient(conn)
	keyVal1 := &apis.Item{
		Key:        "kautilya",
		Value:      "knrt10",
		Expiration: "1m",
	}

	keyVal2 := &apis.Item{
		Key:        "2006",
		Value:      "percona",
		Expiration: "1m",
	}

	keyVal3 := &apis.Item{
		Key:        "foo",
		Value:      "bar",
		Expiration: "1m",
	}

	keyVal4 := &apis.Item{
		Key:        "temp",
		Value:      "bar",
		Expiration: "1µs",
	}

	c.Add(context.Background(), keyVal2)
	c.Add(context.Background(), keyVal3)
	c.Add(context.Background(), keyVal4)

	resp, err := c.Add(context.Background(), keyVal1)
	if err != nil {
		t.Fatalf("Adding key Failed: %v", err)
	}
	if resp.Key != "kautilya" {
		t.Errorf("handler returned unexpected body: got %v want %v",
			resp.Key, "kautilya")
	}
	if resp.Value != "knrt10" {
		t.Errorf("handler returned unexpected body: got %v want %v",
			resp.Key, "knrt10")
	}

	// Save 900000 keys
	for i := 0; i < 40000; i++ {
		c.Add(context.Background(), &apis.Item{
			Key:        strconv.Itoa(i),
			Value:      "Value of i is ",
			Expiration: strconv.Itoa(i),
		})
	}

}

func TestGet(t *testing.T) {
	ctx := context.Background()
	conn, err := grpc.DialContext(ctx, "bufnet", grpc.WithContextDialer(bufDialer), grpc.WithInsecure())
	if err != nil {
		t.Fatalf("Failed to dial bufnet: %v", err)
	}
	defer conn.Close()
	c := apis.NewCacheServiceClient(conn)

	keyGet := &apis.GetKey{
		Key: "kautilya",
	}
	resp, err := c.Get(context.Background(), keyGet)
	if err != nil {
		t.Fatalf("Adding key Failed: %v", err)
	}
	if resp.Key != "kautilya" {
		t.Errorf("handler returned unexpected body: got %v want %v",
			resp.Key, "kautilya")
	}
	if resp.Value != "knrt10" {
		t.Errorf("handler returned unexpected body: got %v want %v",
			resp.Key, "knrt10")
	}
}

func TestGetAllItems(t *testing.T) {
	ctx := context.Background()
	conn, err := grpc.DialContext(ctx, "bufnet", grpc.WithContextDialer(bufDialer), grpc.WithInsecure())
	if err != nil {
		t.Fatalf("Failed to dial bufnet: %v", err)
	}
	defer conn.Close()
	c := apis.NewCacheServiceClient(conn)

	resp, err := c.GetAllItems(context.Background(), &empty.Empty{})
	if err != nil {
		t.Fatalf("Adding key Failed: %v", err)
	}

	if len(resp.Items) != 40002 {
		t.Errorf("handler returned unexpected body: got %v want %v",
			len(resp.Items), 40002)
	}
}

func TestDeleteKey(t *testing.T) {
	ctx := context.Background()
	conn, err := grpc.DialContext(ctx, "bufnet", grpc.WithContextDialer(bufDialer), grpc.WithInsecure())
	if err != nil {
		t.Fatalf("Failed to dial bufnet: %v", err)
	}
	defer conn.Close()
	c := apis.NewCacheServiceClient(conn)

	keyGet := &apis.GetKey{
		Key: "22",
	}
	resp, err := c.DeleteKey(context.Background(), keyGet)
	if err != nil {
		t.Fatalf("Adding key Failed: %v", err)
	}
	if resp.Success != true {
		t.Errorf("handler returned unexpected body: got %v want %v",
			resp.Success, true)
	}
}

func TestDeleteAll(t *testing.T) {
	ctx := context.Background()
	conn, err := grpc.DialContext(ctx, "bufnet", grpc.WithContextDialer(bufDialer), grpc.WithInsecure())
	if err != nil {
		t.Fatalf("Failed to dial bufnet: %v", err)
	}
	defer conn.Close()
	c := apis.NewCacheServiceClient(conn)

	resp, err := c.DeleteAll(context.Background(), &empty.Empty{})
	if err != nil {
		t.Fatalf("Adding key Failed: %v", err)
	}
	if resp.Success != true {
		t.Errorf("handler returned unexpected body: got %v want %v",
			resp.Success, true)
	}
}

// Testing deleted Key
func TestGetDeletedKey(t *testing.T) {
	ctx := context.Background()
	conn, err := grpc.DialContext(ctx, "bufnet", grpc.WithContextDialer(bufDialer), grpc.WithInsecure())
	if err != nil {
		t.Fatalf("Failed to dial bufnet: %v", err)
	}
	defer conn.Close()
	c := apis.NewCacheServiceClient(conn)

	// Geting expired key
	keyGet := &apis.GetKey{
		Key: "temp",
	}
	_, err = c.Get(context.Background(), keyGet)
	if err.Error() != "rpc error: code = Unknown desc = No key found" {
		t.Errorf("Key not deleted")
	}
}
