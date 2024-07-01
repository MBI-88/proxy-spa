package spa

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/spf13/viper"

)

type SPA struct {
	dir   string
	url   *url.URL
	port  int
	proxy *httputil.ReverseProxy
}

func (s SPA) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if strings.HasPrefix(r.URL.Path, "/api") {
		s.proxy.ServeHTTP(w, r)
	} else {
		path := filepath.Join(s.dir, r.URL.Path)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		http.FileServer(http.Dir(s.dir)).ServeHTTP(w, r)
	}
}

// Server runs the proxy and serves spa
func (s SPA) Server() {
	router := mux.NewRouter()
	router.PathPrefix("/").Handler(s)

	srv := &http.Server{
		Handler:      router,
		Addr:         fmt.Sprintf("0.0.0.0:%d", s.port),
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil {
			log.Println(err)
		}
	}()

	fmt.Printf("*************************************\n****Proxy running at 0.0.0.0:%d****\n*************************************\n", s.port)

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	<-c

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	srv.Shutdown(ctx)
	log.Println("Shutting down")
	os.Exit(0)
}

// Set environment variables and setting proxy
func (s *SPA) SetEnv(prod string) {
	if prod == "test" {
		viper.SetConfigFile("./.env")
		if err := viper.ReadInConfig(); err != nil {
			if _, ok := err.(viper.ConfigFileAlreadyExistsError); ok {
				panic(err)
			}
		}
	} else {
		viper.SetEnvPrefix("")
		viper.AutomaticEnv()
	}

	s.dir = viper.GetString("DIR")
	s.port = viper.GetInt("PORT")
	u := viper.GetString("URL")

	if s.dir == "" || s.port == 0 || u == "" {
		panic("Error in setting environment vars")
	}
	target, err := url.Parse(u)
	if err != nil {
		panic(err)
	}
	s.url = target
	s.proxy = &httputil.ReverseProxy{
		Director: func(r *http.Request) {
			r.Header.Set("Origin", u)
			r.URL.Host = s.url.Host
			r.URL.Scheme = s.url.Scheme
			r.URL.Path = strings.TrimPrefix(r.URL.Path, "/api")
		},
		ModifyResponse: func(r *http.Response) error {
			r.Header.Set("Control-Allow-Origin", "*")
			return nil
		},
	}
}
