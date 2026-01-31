package controlplane

import "net/http"

type ControlPlaneServer struct {
	mux *http.ServeMux
	
}