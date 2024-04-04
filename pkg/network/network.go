package network

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"net"
	"os"
	"time"

	grpc_opentracing "github.com/grpc-ecosystem/go-grpc-middleware/tracing/opentracing"
	"github.com/opentracing/opentracing-go"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"
)

type Node struct {
	Addr                  string `yaml:"addr"`
	SslTargetNameOverride string `yaml:"ssl_target_name_override"`
	TLSCACert             string `yaml:"tls_ca_cert"`
	Org                   string `yaml:"org"`
	TLSCAKey              string `yaml:"tls_ca_key"`
	TLSCARoot             string `yaml:"tls_ca_root"`
	TLSCACertByte         []byte
	TLSCAKeyByte          []byte
	TLSCARootByte         []byte
}

func (n *Node) LoadConfig() error {
	TLSCACert, err := GetTLSCACerts(n.TLSCACert)
	if err != nil {
		return fmt.Errorf("fail to load TLS CA Cert %s %w", n.TLSCACert, err)
	}
	certPEM, err := GetTLSCACerts(n.TLSCAKey)
	if err != nil {
		return fmt.Errorf("fail to load TLS CA Key %s %w", n.TLSCAKey, err)

	}
	TLSCARoot, err := GetTLSCACerts(n.TLSCARoot)
	if err != nil {
		return fmt.Errorf("fail to load TLS CA Root %s %w", n.TLSCARoot, err)
	}
	n.TLSCACertByte = TLSCACert
	n.TLSCAKeyByte = certPEM
	n.TLSCARootByte = TLSCARoot
	return nil
}

func GetTLSCACerts(file string) ([]byte, error) {
	if len(file) == 0 {
		return nil, nil
	}

	in, err := os.ReadFile(file)
	if err != nil {
		return nil, fmt.Errorf("error loading %s %w", file, err)
	}
	return in, nil
}

func CreateGRPCClient(node Node) (*GRPCClient, error) {
	var certs [][]byte
	if node.TLSCACertByte != nil {
		certs = append(certs, node.TLSCACertByte)
	}
	config := ClientConfig{}
	config.Timeout = 5 * time.Second
	config.SecOpts = SecureOptions{
		UseTLS:            false,
		RequireClientCert: false,
		ServerRootCAs:     certs,
	}

	if len(certs) > 0 {
		config.SecOpts.UseTLS = true
		if len(node.TLSCAKey) > 0 && len(node.TLSCARoot) > 0 {
			config.SecOpts.RequireClientCert = true
			config.SecOpts.Certificate = node.TLSCACertByte
			config.SecOpts.Key = node.TLSCAKeyByte
			if node.TLSCARootByte != nil {
				config.SecOpts.ClientRootCAs = append(config.SecOpts.ClientRootCAs, node.TLSCARootByte)
			}
		}
	}

	grpcClient, err := NewGRPCClient(config)
	//to do: unit test for this error, current fails to make case for this
	if err != nil {
		return nil, fmt.Errorf("error connecting to %s: %w", node.Addr, err)
	}

	return grpcClient, nil
}

type GRPCClient struct {
	// TLS configuration used by the grpc.ClientConn
	tlsConfig *tls.Config
	// Options for setting up new connections
	dialOpts []grpc.DialOption
	// Duration for which to block while established a new connection
	timeout time.Duration
	// Maximum message size the client can receive
	maxRecvMsgSize int
	// Maximum message size the client can send
	maxSendMsgSize int
}

var (
	MaxRecvMsgSize = 100 * 1024 * 1024
	MaxSendMsgSize = 100 * 1024 * 1024
)

// KeepaliveOptions is used to set the gRPC keepalive settings for both
// clients and servers
type KeepaliveOptions struct {
	// ClientInterval is the duration after which if the client does not see
	// any activity from the server it pings the server to see if it is alive
	ClientInterval time.Duration
	// ClientTimeout is the duration the client waits for a response
	// from the server after sending a ping before closing the connection
	ClientTimeout time.Duration
	// ServerInterval is the duration after which if the server does not see
	// any activity from the client it pings the client to see if it is alive
	ServerInterval time.Duration
	// ServerTimeout is the duration the server waits for a response
	// from the client after sending a ping before closing the connection
	ServerTimeout time.Duration
	// ServerMinInterval is the minimum permitted time between client pings.
	// If clients send pings more frequently, the server will disconnect them
	ServerMinInterval time.Duration
}

// ClientConfig defines the parameters for configuring a GRPCClient instance
type ClientConfig struct {
	// SecOpts defines the security parameters
	SecOpts SecureOptions
	// KaOpts defines the keepalive parameters
	KaOpts KeepaliveOptions
	// Timeout specifies how long the client will block when attempting to
	// establish a connection
	Timeout time.Duration
	// AsyncConnect makes connection creation non blocking
	AsyncConnect bool
}

// NewGRPCClient creates a new implementation of GRPCClient given an address
// and client configuration
func NewGRPCClient(config ClientConfig) (*GRPCClient, error) {
	client := &GRPCClient{}

	// parse secure options
	err := client.parseSecureOptions(config.SecOpts)
	if err != nil {
		return client, err
	}

	// keepalive options

	kap := keepalive.ClientParameters{
		Time:                config.KaOpts.ClientInterval,
		Timeout:             config.KaOpts.ClientTimeout,
		PermitWithoutStream: true,
	}
	// set keepalive
	client.dialOpts = append(client.dialOpts, grpc.WithKeepaliveParams(kap))
	// Unless asynchronous connect is set, make connection establishment blocking.
	if !config.AsyncConnect {
		client.dialOpts = append(client.dialOpts, grpc.WithBlock())
		client.dialOpts = append(client.dialOpts, grpc.FailOnNonTempDialError(true))
	}
	client.timeout = config.Timeout
	// set send/recv message size to package defaults
	client.maxRecvMsgSize = MaxRecvMsgSize
	client.maxSendMsgSize = MaxSendMsgSize

	return client, nil
}

// SecureOptions defines the security parameters (e.g. TLS) for a
// GRPCServer or GRPCClient instance
type SecureOptions struct {
	// VerifyCertificate, if not nil, is called after normal
	// certificate verification by either a TLS client or server.
	// If it returns a non-nil error, the handshake is aborted and that error results.
	VerifyCertificate func(rawCerts [][]byte, verifiedChains [][]*x509.Certificate) error
	// PEM-encoded X509 public key to be used for TLS communication
	Certificate []byte
	// PEM-encoded private key to be used for TLS communication
	Key []byte
	// Set of PEM-encoded X509 certificate authorities used by clients to
	// verify server certificates
	ServerRootCAs [][]byte
	// Set of PEM-encoded X509 certificate authorities used by servers to
	// verify client certificates
	ClientRootCAs [][]byte
	// Whether or not to use TLS for communication
	UseTLS bool
	// Whether or not TLS client must present certificates for authentication
	RequireClientCert bool
	// CipherSuites is a list of supported cipher suites for TLS
	CipherSuites []uint16
	// TimeShift makes TLS handshakes time sampling shift to the past by a given duration
	TimeShift time.Duration
}

func (client *GRPCClient) parseSecureOptions(opts SecureOptions) error {
	// if TLS is not enabled, return
	if !opts.UseTLS {
		return nil
	}

	client.tlsConfig = &tls.Config{
		VerifyPeerCertificate: opts.VerifyCertificate,
		MinVersion:            tls.VersionTLS12} // TLS 1.2 only
	if len(opts.ServerRootCAs) > 0 {
		client.tlsConfig.RootCAs = x509.NewCertPool()
		for _, certBytes := range opts.ServerRootCAs {
			err := AddPemToCertPool(certBytes, client.tlsConfig.RootCAs)
			if err != nil {
				//commLogger.Debugf("error adding root certificate: %v", err)
				return fmt.Errorf("error adding root certificate: %w", err)
			}
		}
	}
	if opts.RequireClientCert {
		// make sure we have both Key and Certificate
		if opts.Key != nil &&
			opts.Certificate != nil {
			cert, err := tls.X509KeyPair(opts.Certificate,
				opts.Key)
			if err != nil {
				return fmt.Errorf("failed to load client certificate: %w", err)
			}
			client.tlsConfig.Certificates = append(
				client.tlsConfig.Certificates, cert)
		} else {
			return errors.New("both Key and Certificate are required when using mutual TLS")
		}
	}

	if opts.TimeShift > 0 {
		client.tlsConfig.Time = func() time.Time {
			return time.Now().Add((-1) * opts.TimeShift)
		}
	}

	return nil
}

// AddPemToCertPool adds PEM-encoded certs to a cert pool
func AddPemToCertPool(pemCerts []byte, pool *x509.CertPool) error {
	certs, _, err := pemToX509Certs(pemCerts)
	if err != nil {
		return err
	}
	for _, cert := range certs {
		pool.AddCert(cert)
	}
	return nil
}

// parse PEM-encoded certs
func pemToX509Certs(pemCerts []byte) ([]*x509.Certificate, []string, error) {
	var certs []*x509.Certificate
	var subjects []string

	// it's possible that multiple certs are encoded
	for len(pemCerts) > 0 {
		var block *pem.Block
		block, pemCerts = pem.Decode(pemCerts)
		if block == nil {
			break
		}

		cert, err := x509.ParseCertificate(block.Bytes)
		if err != nil {
			return nil, []string{}, err
		}

		certs = append(certs, cert)
		subjects = append(subjects, string(cert.RawSubject))
	}

	return certs, subjects, nil
}

func DialConnection(node Node) (*grpc.ClientConn, error) {
	gRPCClient, err := CreateGRPCClient(node)
	if err != nil {
		return nil, err
	}
	var connError error
	var conn *grpc.ClientConn
	for i := 1; i <= 3; i++ {
		conn, connError = gRPCClient.NewConnection(node.Addr, func(tlsConfig *tls.Config) {
			tlsConfig.InsecureSkipVerify = true
			tlsConfig.ServerName = node.SslTargetNameOverride
		})
		if connError == nil {
			return conn, nil
		} else {
			fmt.Printf("%d of 3 attempts to make connection to %s, details: %s", i, node.Addr, connError)
		}
	}
	return nil, fmt.Errorf("error connecting to %s: %w", node.Addr, connError)
}

type DynamicClientCredentials struct {
	TLSConfig  *tls.Config
	TLSOptions []TLSOption
}

func (dtc *DynamicClientCredentials) latestConfig() *tls.Config {
	tlsConfigCopy := dtc.TLSConfig.Clone()
	for _, tlsOption := range dtc.TLSOptions {
		tlsOption(tlsConfigCopy)
	}
	return tlsConfigCopy
}

func (dtc *DynamicClientCredentials) ClientHandshake(ctx context.Context, authority string, rawConn net.Conn) (net.Conn, credentials.AuthInfo, error) {
	return credentials.NewTLS(dtc.latestConfig()).ClientHandshake(ctx, authority, rawConn)
}

var ErrServerHandshakeNotImplemented = errors.New("core/comm: server handshakes are not implemented with clientCreds")

func (dtc *DynamicClientCredentials) ServerHandshake(rawConn net.Conn) (net.Conn, credentials.AuthInfo, error) {
	return nil, nil, ErrServerHandshakeNotImplemented
}

func (dtc *DynamicClientCredentials) Info() credentials.ProtocolInfo {
	return credentials.NewTLS(dtc.latestConfig()).Info()
}

func (dtc *DynamicClientCredentials) Clone() credentials.TransportCredentials {
	return credentials.NewTLS(dtc.latestConfig())
}

func (dtc *DynamicClientCredentials) OverrideServerName(name string) error {
	dtc.TLSConfig.ServerName = name
	return nil
}

type TLSOption func(tlsConfig *tls.Config)

// NewConnection returns a grpc.ClientConn for the target address and
// overrides the server name used to verify the hostname on the
// certificate returned by a server when using TLS
func (client *GRPCClient) NewConnection(address string, tlsOptions ...TLSOption) (*grpc.ClientConn, error) {

	dialOpts := client.dialOpts

	// set transport credentials and max send/recv message sizes
	// immediately before creating a connection in order to allow
	// SetServerRootCAs / SetMaxRecvMsgSize / SetMaxSendMsgSize
	//  to take effect on a per connection basis
	if client.tlsConfig != nil {
		dialOpts = append(dialOpts, grpc.WithTransportCredentials(
			&DynamicClientCredentials{
				TLSConfig:  client.tlsConfig,
				TLSOptions: tlsOptions,
			},
		))
	} else {
		dialOpts = append(dialOpts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	}

	dialOpts = append(dialOpts, grpc.WithDefaultCallOptions(
		grpc.MaxCallRecvMsgSize(client.maxRecvMsgSize),
		grpc.MaxCallSendMsgSize(client.maxSendMsgSize),
	))

	tracer := opentracing.GlobalTracer()
	opts := []grpc_opentracing.Option{
		grpc_opentracing.WithTracer(tracer),
	}

	dialOpts = append(dialOpts,
		grpc.WithUnaryInterceptor(grpc_opentracing.UnaryClientInterceptor(opts...)),
		grpc.WithStreamInterceptor(grpc_opentracing.StreamClientInterceptor(opts...)),
	)

	conn, err := grpc.NewClient(address, dialOpts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create new connection: %w", err)
	}
	return conn, nil
}
