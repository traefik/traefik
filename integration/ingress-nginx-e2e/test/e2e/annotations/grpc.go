/*
Copyright 2019 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package annotations

import (
	"context"
	"crypto/tls"
	"fmt"
	"strings"

	delaypb "github.com/Anddd7/pb/grpcbin"
	pb "github.com/moul/pb/grpcbin/go-grpc"
	"github.com/onsi/ginkgo/v2"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	"k8s.io/ingress-nginx/test/e2e/framework"
)

const (
	echoHost = "echo"
	host     = "grpc"
)

var _ = framework.DescribeAnnotation("backend-protocol - GRPC", func() {
	f := framework.NewDefaultFramework("grpc", framework.WithHTTPBunEnabled())

	ginkgo.It("should return OK for service with backend protocol GRPC", func() {
		f.NewGRPCBinDeployment()

		host := echoHost

		svc := &corev1.Service{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "grpcbin-test",
				Namespace: f.Namespace,
			},
			Spec: corev1.ServiceSpec{
				ExternalName: fmt.Sprintf("grpcbin.%v.svc.cluster.local", f.Namespace),
				Type:         corev1.ServiceTypeExternalName,
				Ports: []corev1.ServicePort{
					{
						Name:       host,
						Port:       9000,
						TargetPort: intstr.FromInt(9000),
						Protocol:   "TCP",
					},
				},
			},
		}
		f.EnsureService(svc)

		annotations := map[string]string{
			"nginx.ingress.kubernetes.io/backend-protocol": "GRPC",
		}

		ing := framework.NewSingleIngressWithTLS(host, "/", host, []string{host}, f.Namespace, "grpcbin-test", 9000, annotations)

		f.EnsureIngress(ing)

		f.WaitForNginxServer(host,
			func(server string) bool {
				return strings.Contains(server, "grpc_pass grpc://upstream_balancer;")
			})

		//nolint:goconst //string interpolation
		conn, err := grpc.NewClient(f.GetNginxIP()+":443",
			grpc.WithTransportCredentials(
				credentials.NewTLS(&tls.Config{
					ServerName:         echoHost,
					InsecureSkipVerify: true, //nolint:gosec // Ignore certificate validation in testing
				}),
			),
		)
		assert.Nil(ginkgo.GinkgoT(), err, "error creating a connection")
		defer conn.Close()

		client := pb.NewGRPCBinClient(conn)
		ctx := context.Background()

		res, err := client.HeadersUnary(ctx, &pb.EmptyMessage{})
		assert.Nil(ginkgo.GinkgoT(), err)

		metadata := res.GetMetadata()
		assert.Equal(ginkgo.GinkgoT(), metadata["content-type"].Values[0], "application/grpc")
		assert.Equal(ginkgo.GinkgoT(), metadata[":authority"].Values[0], host)
	})

	ginkgo.It("authorization metadata should be overwritten by external auth response headers", func() {
		f.NewGRPCBinDeployment()
		host := echoHost

		svc := &corev1.Service{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "grpcbin-test",
				Namespace: f.Namespace,
			},
			Spec: corev1.ServiceSpec{
				ExternalName: fmt.Sprintf("grpcbin.%v.svc.cluster.local", f.Namespace),
				Type:         corev1.ServiceTypeExternalName,
				Ports: []corev1.ServicePort{
					{
						Name:       host,
						Port:       9000,
						TargetPort: intstr.FromInt(9000),
						Protocol:   "TCP",
					},
				},
			},
		}
		f.EnsureService(svc)

		annotations := map[string]string{
			"nginx.ingress.kubernetes.io/auth-url":              fmt.Sprintf("http://%s/response-headers?authorization=foo", f.HTTPBunIP),
			"nginx.ingress.kubernetes.io/auth-response-headers": "Authorization",
			"nginx.ingress.kubernetes.io/backend-protocol":      "GRPC",
		}

		ing := framework.NewSingleIngressWithTLS(host, "/", host, []string{host}, f.Namespace, "grpcbin-test", 9000, annotations)

		f.EnsureIngress(ing)

		f.WaitForNginxServer(host,
			func(server string) bool {
				return strings.Contains(server, "grpc_pass grpc://upstream_balancer;")
			})

		conn, err := grpc.NewClient(f.GetNginxIP()+":443",
			grpc.WithTransportCredentials(
				credentials.NewTLS(&tls.Config{
					ServerName:         echoHost,
					InsecureSkipVerify: true, //nolint:gosec // Ignore certificate validation in testing
				}),
			),
		)
		assert.Nil(ginkgo.GinkgoT(), err)
		defer conn.Close()

		client := pb.NewGRPCBinClient(conn)
		ctx := metadata.AppendToOutgoingContext(context.Background(),
			"authorization", "bar")

		res, err := client.HeadersUnary(ctx, &pb.EmptyMessage{})
		assert.Nil(ginkgo.GinkgoT(), err)

		metadata := res.GetMetadata()
		assert.Equal(ginkgo.GinkgoT(), fooHost, metadata["authorization"].Values[0])
	})

	ginkgo.It("should return OK for service with backend protocol GRPCS", func() {
		f.NewGRPCBinDeployment()

		disableSnippet := f.AllowSnippetConfiguration()
		defer disableSnippet()

		host := echoHost

		svc := &corev1.Service{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "grpcbin-test",
				Namespace: f.Namespace,
			},
			Spec: corev1.ServiceSpec{
				ExternalName: fmt.Sprintf("grpcbin.%v.svc.cluster.local", f.Namespace),
				Type:         corev1.ServiceTypeExternalName,
				Ports: []corev1.ServicePort{
					{
						Name:       host,
						Port:       9001,
						TargetPort: intstr.FromInt(9001),
						Protocol:   "TCP",
					},
				},
			},
		}
		f.EnsureService(svc)

		annotations := map[string]string{
			"nginx.ingress.kubernetes.io/backend-protocol": "GRPCS",
			"nginx.ingress.kubernetes.io/configuration-snippet": fmt.Sprintf(`
			   # without this setting NGINX sends echo instead
			   grpc_ssl_name      		grpcbin.%v.svc.cluster.local;
			   grpc_ssl_server_name		on;
			   grpc_ssl_ciphers 		HIGH:!aNULL:!MD5;
			`, f.Namespace),
		}

		ing := framework.NewSingleIngressWithTLS(host, "/", host, []string{host}, f.Namespace, "grpcbin-test", 9001, annotations)
		f.EnsureIngress(ing)

		f.WaitForNginxServer(host,
			func(server string) bool {
				return strings.Contains(server, "grpc_pass grpcs://upstream_balancer;")
			})

		conn, err := grpc.NewClient(f.GetNginxIP()+":443",
			grpc.WithTransportCredentials(
				credentials.NewTLS(&tls.Config{
					ServerName:         echoHost,
					InsecureSkipVerify: true, //nolint:gosec // Ignore the gosec error in testing
				}),
			),
		)
		assert.Nil(ginkgo.GinkgoT(), err)
		defer conn.Close()

		client := pb.NewGRPCBinClient(conn)
		ctx := context.Background()

		res, err := client.HeadersUnary(ctx, &pb.EmptyMessage{})
		assert.Nil(ginkgo.GinkgoT(), err)

		metadata := res.GetMetadata()
		assert.Equal(ginkgo.GinkgoT(), metadata["content-type"].Values[0], "application/grpc")
	})

	ginkgo.It("should return OK when request not exceed timeout", func() {
		f.NewGRPCBinDelayDeployment()

		proxyTimeout := "10"
		ingressName := "grpcbin-delay"

		annotations := make(map[string]string)
		annotations["nginx.ingress.kubernetes.io/backend-protocol"] = "GRPC"
		annotations["nginx.ingress.kubernetes.io/proxy-connect-timeout"] = proxyTimeout
		annotations["nginx.ingress.kubernetes.io/proxy-send-timeout"] = proxyTimeout
		annotations["nginx.ingress.kubernetes.io/proxy-read-timeout"] = proxyTimeout

		ing := framework.NewSingleIngress(host, "/", host, f.Namespace, ingressName, 50051, annotations)

		f.EnsureIngress(ing)

		f.WaitForNginxServer(host,
			func(server string) bool {
				return strings.Contains(server, fmt.Sprintf("grpc_connect_timeout %ss;", proxyTimeout)) &&
					strings.Contains(server, fmt.Sprintf("grpc_send_timeout %ss;", proxyTimeout)) &&
					strings.Contains(server, fmt.Sprintf("grpc_read_timeout %ss;", proxyTimeout))
			})

		conn, err := grpc.NewClient(
			f.GetNginxIP()+":80",
			grpc.WithTransportCredentials(insecure.NewCredentials()),
			grpc.WithAuthority(host),
		)
		assert.Nil(ginkgo.GinkgoT(), err, "error creating a connection")
		defer conn.Close()

		client := delaypb.NewGrpcbinServiceClient(conn)

		res, err := client.Unary(context.Background(), &delaypb.UnaryRequest{
			Data: "hello",
		})
		assert.Nil(ginkgo.GinkgoT(), err)

		metadata := res.GetResponseAttributes().RequestHeaders
		assert.Equal(ginkgo.GinkgoT(), metadata["content-type"], "application/grpc")
		assert.Equal(ginkgo.GinkgoT(), metadata[":authority"], host)
	})

	ginkgo.It("should return Error when request exceed timeout", func() {
		f.NewGRPCBinDelayDeployment()

		proxyTimeout := "10"
		ingressName := "grpcbin-delay"

		annotations := make(map[string]string)
		annotations["nginx.ingress.kubernetes.io/backend-protocol"] = "GRPC"
		annotations["nginx.ingress.kubernetes.io/proxy-connect-timeout"] = proxyTimeout
		annotations["nginx.ingress.kubernetes.io/proxy-send-timeout"] = proxyTimeout
		annotations["nginx.ingress.kubernetes.io/proxy-read-timeout"] = proxyTimeout

		ing := framework.NewSingleIngress(host, "/", host, f.Namespace, ingressName, 50051, annotations)

		f.EnsureIngress(ing)

		f.WaitForNginxServer(host,
			func(server string) bool {
				return strings.Contains(server, fmt.Sprintf("grpc_connect_timeout %ss;", proxyTimeout)) &&
					strings.Contains(server, fmt.Sprintf("grpc_send_timeout %ss;", proxyTimeout)) &&
					strings.Contains(server, fmt.Sprintf("grpc_read_timeout %ss;", proxyTimeout))
			})

		conn, err := grpc.NewClient(
			f.GetNginxIP()+":80",
			grpc.WithTransportCredentials(insecure.NewCredentials()),
			grpc.WithAuthority(host),
		)
		assert.Nil(ginkgo.GinkgoT(), err, "error creating a connection")
		defer conn.Close()

		client := delaypb.NewGrpcbinServiceClient(conn)

		_, err = client.Unary(context.Background(), &delaypb.UnaryRequest{
			Data: "hello",
			RequestAttributes: &delaypb.RequestAttributes{
				Delay: 15,
			},
		})
		assert.Error(ginkgo.GinkgoT(), err)
	})
})
