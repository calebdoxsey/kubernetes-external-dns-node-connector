package main

import (
	"encoding/gob"
	"flag"
	"log"
	"net"

	"github.com/kubernetes-incubator/external-dns/endpoint"
	core_v1 "k8s.io/api/core/v1"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

var options = struct {
	Address string
	DNSName string
}{
	Address: ":8080",
}

func main() {
	log.SetFlags(0)

	flag.StringVar(&options.Address, "address", options.Address, "the address to bind")
	flag.StringVar(&options.DNSName, "dns-name", options.DNSName, "the dns name to use for the nodes")
	flag.Parse()

	cfg, err := rest.InClusterConfig()
	if err != nil {
		log.Fatalln(err)
	}

	client, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		log.Fatalln(err)
	}

	log.Println("starting server on", options.Address)
	li, err := net.Listen("tcp", options.Address)
	if err != nil {
		log.Fatalln(err)
	}
	defer li.Close()

	for {
		conn, err := li.Accept()
		if err != nil {
			log.Fatalln(err)
		}
		go handle(conn, client)
	}
}

func handle(conn net.Conn, client kubernetes.Interface) {
	defer conn.Close()

	nodes, err := client.CoreV1().Nodes().List(meta_v1.ListOptions{})
	if err != nil {
		log.Fatalln(err)
	}

	var ips []string
	for _, node := range nodes.Items {
		ip := ""
		for _, addr := range node.Status.Addresses {
			switch addr.Type {
			case core_v1.NodeExternalIP:
				ip = addr.Address
			}
		}
		if ip != "" {
			ips = append(ips, ip)
		}
	}

	endpoints := []*endpoint.Endpoint{
		endpoint.NewEndpoint(options.DNSName, "A", ips...),
	}

	log.Println("sending endpoints", endpoints, "to", conn.RemoteAddr())

	err = gob.NewEncoder(conn).Encode(endpoints)
	if err != nil {
		log.Fatalln(err)
	}
}
