package main

import (
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"

	config "github.com/Tangyd893/WorkPal/backend/configs"
	"github.com/Tangyd893/WorkPal/backend/internal/platform"
	"github.com/gin-gonic/gin"
)

type proxySet struct {
	user   *httputil.ReverseProxy
	im     *httputil.ReverseProxy
	file   *httputil.ReverseProxy
	search *httputil.ReverseProxy
}

func main() {
	cfg, err := platform.LoadConfig()
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	proxies, err := newProxySet(cfg)
	if err != nil {
		log.Fatalf("create gateway proxies: %v", err)
	}

	r := platform.NewRouter(cfg, "gateway")
	platform.RegisterHealth(r, nil, nil)
	r.NoRoute(func(c *gin.Context) {
		proxy := proxies.match(c.Request.URL.Path)
		if proxy == nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "route not found"})
			return
		}
		proxy.ServeHTTP(c.Writer, c.Request)
	})

	if err := platform.RunHTTP("gateway", cfg.Services.GatewayPort, r, nil); err != nil {
		log.Fatalf("gateway stopped: %v", err)
	}
}

func newProxySet(cfg *config.Config) (*proxySet, error) {
	userProxy, err := newReverseProxy(cfg.Services.UserURL)
	if err != nil {
		return nil, err
	}
	imProxy, err := newReverseProxy(cfg.Services.IMURL)
	if err != nil {
		return nil, err
	}
	fileProxy, err := newReverseProxy(cfg.Services.FileURL)
	if err != nil {
		return nil, err
	}
	searchProxy, err := newReverseProxy(cfg.Services.SearchURL)
	if err != nil {
		return nil, err
	}
	return &proxySet{
		user:   userProxy,
		im:     imProxy,
		file:   fileProxy,
		search: searchProxy,
	}, nil
}

func newReverseProxy(rawURL string) (*httputil.ReverseProxy, error) {
	target, err := url.Parse(rawURL)
	if err != nil {
		return nil, err
	}
	proxy := httputil.NewSingleHostReverseProxy(target)
	proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
		log.Printf("gateway proxy error for %s -> %s: %v", r.URL.Path, rawURL, err)
		http.Error(w, "upstream service unavailable", http.StatusBadGateway)
	}
	return proxy, nil
}

func (p *proxySet) match(path string) *httputil.ReverseProxy {
	switch {
	case path == "/ws":
		return p.im
	case strings.HasPrefix(path, "/api/v1/auth"):
		return p.user
	case strings.HasPrefix(path, "/api/v1/users"):
		return p.user
	case strings.HasPrefix(path, "/api/v1/departments"):
		return p.user
	case strings.HasPrefix(path, "/api/v1/files"):
		return p.file
	case strings.HasPrefix(path, "/api/v1/conversations/") && strings.HasSuffix(path, "/files"):
		return p.file
	case strings.HasPrefix(path, "/api/v1/search"):
		return p.search
	case strings.HasPrefix(path, "/api/v1/conversations"):
		return p.im
	case strings.HasPrefix(path, "/api/v1/messages"):
		return p.im
	default:
		return nil
	}
}
