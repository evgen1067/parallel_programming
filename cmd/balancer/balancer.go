package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"sync"
	"sync/atomic"
	"time"
)

const Retry = "retry-count"

type backend struct {
	url   *url.URL
	mu    *sync.RWMutex
	alive bool
	proxy *httputil.ReverseProxy
}

func (b *backend) SetAlive(alive bool) {
	b.mu.Lock()
	b.alive = alive
	b.mu.Unlock()
}

func (b *backend) IsAlive() (alive bool) {
	b.mu.RLock()
	alive = b.alive
	b.mu.RUnlock()

	return alive
}

type pool struct {
	backends []*backend
	current  uint64
}

type BackendPool interface {
	AddBackend(url *url.URL, proxy *httputil.ReverseProxy)
	SetBackendStatus(url *url.URL, alive bool)
	Next() *backend
	HealthCheck()
	Start(port int) error
	NextBackend() int
}

func (p *pool) AddBackend(url *url.URL, proxy *httputil.ReverseProxy) {
	b := &backend{
		url:   url,
		mu:    &sync.RWMutex{},
		alive: true,
		proxy: proxy,
	}
	proxy.ErrorHandler = b.errorHandler
	p.backends = append(p.backends, b)
}

func (p *pool) NextBackend() int {
	return int(atomic.AddUint64(&p.current, uint64(1)) % uint64(len(p.backends)))
}

func (p *pool) SetBackendStatus(url *url.URL, alive bool) {
	for _, b := range p.backends {
		if b.url.String() == url.String() {
			b.SetAlive(alive)
			return
		}
	}
}

func (p *pool) Next() *backend {
	next := p.NextBackend()
	length := len(p.backends) + next

	for i := next; i < length; i++ {
		idx := i % len(p.backends)
		if p.backends[idx].IsAlive() {
			if i != next {
				atomic.StoreUint64(&p.current, uint64(idx))
			}
			return p.backends[idx]
		}
	}

	return nil
}

func (p *pool) HealthCheck() {
	t := time.NewTicker(time.Minute)
	for {
		select {
		case <-t.C:
			log.Println("Проверка работоспособности [START]")
			for _, b := range p.backends {
				alive := b.IsAlive()
				timeout := 2 * time.Second
				conn, err := net.DialTimeout("tcp", b.url.Host, timeout)
				if err != nil {
					alive = false
				}
				conn.Close()
				b.SetAlive(alive)

				var status string
				if b.IsAlive() {
					status = "сервер не отвечает"
				} else {
					status = "сервер работает"
				}
				log.Print(fmt.Sprintf("%s [%s]\n", b.url, status))
			}
			log.Println("Проверка работоспособности [END]")
		}
	}
}

func (p *pool) Start(port int) error {
	server := http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: http.HandlerFunc(p.lb),
	}

	// start health checking
	go p.HealthCheck()

	log.Printf("Балансировщик нагрузки запущен: %s:%d\n", "http://localhost", port)
	if err := server.ListenAndServe(); err != nil {
		return err
	}

	return nil
}

func (p *pool) lb(w http.ResponseWriter, r *http.Request) {
	peer := p.Next()
	if peer != nil {
		peer.proxy.ServeHTTP(w, r)
		return
	}
	http.Error(w, "Сервис не доступен", http.StatusServiceUnavailable)
}

func (b *backend) errorHandler(writer http.ResponseWriter, request *http.Request, e error) {
	log.Printf("[%s] %s\n", b.url.Host, e.Error())
	retries := 0
	if retry, ok := request.Context().Value(Retry).(int); ok {
		retries = retry
	}
	if retries < 5 {
		select {
		case <-time.After(10 * time.Millisecond):
			ctx := context.WithValue(request.Context(), Retry, retries+1)
			b.proxy.ServeHTTP(writer, request.WithContext(ctx))
		}
		return
	}

	b.SetAlive(false)
}

func NewPool(urls []string) (BackendPool, error) {
	p := &pool{}
	for _, u := range urls {
		parsedUrl, err := url.Parse(u)
		if err != nil {
			return nil, err
		}

		// обработчик HTTP, который берёт входящие запросы и отправляет на другой сервер,
		// проксируя ответы обратно клиенту.
		proxy := httputil.NewSingleHostReverseProxy(parsedUrl)
		p.AddBackend(parsedUrl, proxy)

		log.Printf("Настроенный сервер: %s\n", parsedUrl)
	}

	return p, nil
}
