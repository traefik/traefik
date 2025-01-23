package conformance

import (
	"knative.dev/networking/test/conformance/ingress"
	"testing"
)

func TestYourIngressConformance(t *testing.T) {
	ingress.RunConformance(t)
}

//func TestT2(t *testing.T) {
//	const suffix = "- pong"
//
//	// Establish a TCP connection
//	tcpConn, err := net.Dial("tcp", "localhost:30874")
//	if err != nil {
//		fmt.Printf("did not connect: %v", err)
//	}
//	defer tcpConn.Close()
//	conn, err := grpc.Dial("your-ingress-conformance-grpc-ifdqevhr.example.com:80",
//		grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithContextDialer(
//			func(ctx context.Context, addr string) (net.Conn, error) {
//				return tcpConn, nil
//			}))
//	if err != nil {
//		fmt.Printf("did not connect: %v", err)
//	}
//	defer conn.Close()
//	c := pb.NewPingServiceClient(conn)
//
//	// Increase the timeout duration
//	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
//	defer cancel()
//
//	r, err := c.PingStream(ctx)
//	if err != nil {
//		fmt.Printf("could not ping: %v", err)
//	}
//	fmt.Printf("Ping response: %v", r)
//	for i := 0; i < 100; i++ {
//		fmt.Printf("Iteration: %v\n", i)
//		checkGRPCRoundTrip(t, r, suffix)
//	}
//	//fmt.Printf(t.Name())
//}
//
//func checkGRPCRoundTrip(t *testing.T, stream pb.PingService_PingStreamClient, suffix string) {
//	message := fmt.Sprint("ping -", rand.Intn(1000))
//	if err := stream.Send(&pb.Request{Msg: message}); err != nil {
//		t.Error("Error sending request:", err)
//		return
//	}
//
//	// Read back the echoed message and compared with sent.
//	if resp, err := stream.Recv(); err != nil {
//		t.Error("Error receiving response:", err)
//	} else if got, want := resp.Msg, message; got != want {
//		t.Errorf("Recv() = %s, wanted %s", got, want)
//	}
//}
