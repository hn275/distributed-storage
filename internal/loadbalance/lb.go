package loadbalance

import "net"

type LBAlgo interface {
	// adding a new node to the cluster
	NodeJoin(net.Conn)

	// feel free to define more functions here
}

type LoadBalancer struct {
	net.Listener
	Engine LBAlgo
}

func NewBalancer(protocol, addr string, algo LBAlgo) (*LoadBalancer, error) {
	soc, err := net.Listen(protocol, addr)
	if err != nil {
		return nil, err
	}

	loadBalancer := &LoadBalancer{soc, algo}
	return loadBalancer, nil
}
