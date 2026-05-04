package main

import (
	"log"

	"github.com/Tangyd893/WorkPal/backend/internal/platform"
)

func main() {
	cfg, err := platform.LoadConfig()
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	app, err := newGatewayApp(cfg)
	if err != nil {
		log.Fatalf("create gateway app: %v", err)
	}

	r := platform.NewRouter(cfg, gatewayServiceName)
	app.Register(r)

	if err := platform.RunHTTP(gatewayServiceName, cfg.Services.GatewayPort, r, app.Shutdown); err != nil {
		log.Fatalf("gateway stopped: %v", err)
	}
}
