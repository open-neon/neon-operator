package controlplane

import (
	"crypto/tls"
	"fmt"
	"log/slog"
	"net/http"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// ControlPlaneServer represents the control plane HTTP server
type ControlPlaneServer struct {
	mux     *http.ServeMux
	logger  *slog.Logger
	server  *http.Server
	nclient client.Client
	kclient kubernetes.Interface
	scheme  *runtime.Scheme
}

const (
	ServiceName         = "control-plane"
	controllerName      = "control-plane"
	httpPort            = ":9090"
	httpsPort           = ":9443"
	certPath            = "/etc/control-plane/certs/tls.crt"
	certKeyPath         = "/etc/control-plane/certs/tls.key"
	jwtPublicKeyPath    = "/etc/control-plane-jwt/jwt.pub"
	jwtPrivateKeyPath   = "/etc/control-plane-jwt/jwt.key"
)

var (
	enableTLS = false
	enableJWT = false
	protocol  = "http"
	port      = httpPort
)

// GetEnableTLS returns whether TLS is enabled for the control plane server
func GetEnableTLS() bool {
	return enableTLS
}

// GetEnableJWT returns whether JWT authentication is enabled for the control plane server
func GetEnableJWT() bool {
	return enableJWT
}

// GetProtocol returns the protocol (http or https) for the control plane server
func GetProtocol() string {
	return protocol
}

// GetPort returns the port for the control plane server
func GetPort() string {
	return port
}

// setTLSConfig is an internal function to set TLS configuration
func setTLSConfig(enabled bool, proto, p string) {
	enableTLS = enabled
	protocol = proto
	port = p
}

// NewControlPlaneServer creates a new control plane server instance
// The port is automatically selected based on enableTLS flag:
// - :9090 for HTTP (when TLS is disabled)
// - :9443 for HTTPS (when TLS is enabled)
func NewControlPlaneServer(enableTLSFlag bool, enableJWTFlag bool, logger *slog.Logger, nclient client.Client, config *rest.Config, scheme *runtime.Scheme) (*ControlPlaneServer, error) {
	// Select port based on TLS setting
	addr := httpPort
	if enableTLSFlag {
		setTLSConfig(true, "https", httpsPort)
		addr = httpsPort
	}
	enableJWT = enableJWTFlag
	logger = logger.With("component", controllerName)

	// Create kubernetes clientset for direct client-go operations
	kclient, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create kubernetes clientset: %w", err)
	}

	mux := http.NewServeMux()

	cps := &ControlPlaneServer{
		mux:     mux,
		logger:  logger,
		nclient: nclient,
		kclient: kclient,
		scheme:  scheme,
	}

	// Configure HTTP server
	cps.server = &http.Server{
		Addr:    addr,
		Handler: mux,
	}

	if enableTLS {
		// Load TLS certificate and configure in TLSConfig
		// When TLSConfig.Certificates is set, ListenAndServeTLS will use these instead of loading from files
		cert, err := tls.LoadX509KeyPair(certPath, certKeyPath)
		if err != nil {
			logger.Warn("failed to load TLS certificate, falling back to HTTP", "error", err)
			// Disable TLS and use HTTP instead
			setTLSConfig(false, "http", httpPort)
			cps.server.Addr = httpPort
		} else {
			cps.server.TLSConfig = &tls.Config{
				Certificates: []tls.Certificate{cert},
			}
			logger.Info("TLS enabled for control plane server", "certPath", certPath)
		}
	}

	return cps, nil
}

// Start starts the control plane server
func (cps *ControlPlaneServer) Start() error {

	cps.logger.Info("starting control plane server", "address", GetPort(), "protocol", GetProtocol())

	if GetEnableTLS() {
		return cps.server.ListenAndServeTLS("", "")
	}
	return cps.server.ListenAndServe()
}

// Stop gracefully shuts down the control plane server
func (cps *ControlPlaneServer) Stop() error {
	cps.logger.Info("shutting down control plane server")
	if cps.server != nil {
		return cps.server.Close()
	}
	return nil
}
