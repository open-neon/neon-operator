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
	mux       *http.ServeMux
	logger    *slog.Logger
	addr      string
	enableTLS bool
	certPath  string
	certKey   string
	server    *http.Server
	nclient   client.Client
	kclient   kubernetes.Interface
	scheme    *runtime.Scheme
}

const (
	controllerName     = "control-plane"
	httpPort           = ":9090"
	httpsPort          = ":9443"
)

// NewControlPlaneServer creates a new control plane server instance
// The port is automatically selected based on enableTLS flag:
// - :9090 for HTTP (when TLS is disabled)
// - :9443 for HTTPS (when TLS is enabled)
func NewControlPlaneServer(enableTLS bool, certPath, certKey string, logger *slog.Logger, nclient client.Client, config *rest.Config, scheme *runtime.Scheme) (*ControlPlaneServer, error) {
	// Select port based on TLS setting
	addr := httpPort
	if enableTLS {
		addr = httpsPort
	}
	logger = logger.With("component", controllerName)

	// Create kubernetes clientset for direct client-go operations
	kclient, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create kubernetes clientset: %w", err)
	}

	mux := http.NewServeMux()

	cps := &ControlPlaneServer{
		mux:       mux,
		logger:    logger,
		addr:      addr,
		enableTLS: enableTLS,
		certPath:  certPath,
		certKey:   certKey,
		nclient:   nclient,
		kclient:   kclient,
		scheme:    scheme,
	}

	// Configure HTTP server
	cps.server = &http.Server{
		Addr:    addr,
		Handler: mux,
	}

	if enableTLS {
		// Load TLS certificate
		cert, err := tls.LoadX509KeyPair(certPath, certKey)
		if err != nil {
			return nil, fmt.Errorf("failed to load TLS certificate: %w", err)
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
	protocol := "HTTP"
	if cps.enableTLS {
		protocol = "HTTPS"
	}

	cps.logger.Info("starting control plane server", "address", cps.addr, "protocol", protocol)

	if cps.enableTLS {
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
