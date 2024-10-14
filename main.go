package main

import (
	"crypto/tls"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"k8s.io/klog/v2"
)

func main() {
	var parameter CommandParameter

	// 获取命令行参数
	flag.IntVar(&parameter.port, "port", 443, "Webhook server port")
	flag.StringVar(&parameter.certFile, "certFile", "/etc/webhook/certs/cert.pem", "Path to the x509 certificate for https")
	flag.StringVar(&parameter.keyFile, "keyFile", "/etc/webhook/certs/key.pem", "Path to the x509 private key matching `CertFile`")
	pair, err := tls.LoadX509KeyPair(parameter.certFile, parameter.keyFile)
	if err != nil {
		klog.Errorf("Failed to load key pair: %v", err)
	}

	whsvr := &WebhookServer{
		server: &http.Server{
			Addr:      fmt.Sprintf(":%v", parameter.port),
			TLSConfig: &tls.Config{Certificates: []tls.Certificate{pair}},
		},
	}

	// define http server and server handler
	mux := http.NewServeMux()
	mux.HandleFunc("/mutate", whsvr.serveMutate)
	// mux.HandleFunc("/validate", whsvr.serve)
	whsvr.server.Handler = mux

	// start webhook server in new routine
	go func() {
		if err := whsvr.server.ListenAndServeTLS("", ""); err != nil {
			klog.Errorf("Failed to listen and serve webhook server: %v", err)
		}
	}()

	klog.Info("Server started")

	// listening OS shutdown singal
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)
	<-signalChan

	klog.Infof("Got OS shutdown signal, shutting down webhook server gracefully...")
}
