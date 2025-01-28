
#!/usr/bin/python3

from mininet.topo import Topo
from mininet.net import Mininet
from mininet.util import dumpNodeConnections
from mininet.node import OVSController
from mininet.log import setLogLevel
import time

class ExpTopo( Topo ):
    def __init__( self ):
            # Initialize topology
            Topo.__init__( self )
            # Add hosts and switches
            h1 = self.addHost( 'h1' )
            h2 = self.addHost( 'h2' )
            h3 = self.addHost( 'h3')
            switch = self.addSwitch( 's1' )
            # Add links
            self.addLink( h1, switch )
            self.addLink( switch, h2 )
            self.addLink( switch, h3)
topos = {'topoExp': (lambda: ExpTopo())}

def runExperiment():
    topo = ExpTopo()
    #net = Mininet(topo)
    net = Mininet(topo= topo, controller=OVSController)
    net.start()
    print("Host Information")
    for host in net.hosts:
        print(f"Host: {host.name}, IP: {host.IP()}, MAC: {host.MAC()}")   
    
    print("Test running executables")
    # Start server
    net.get('h1').cmd(f"python3 -u ./server.py --host 10.0.0.1 --port 7272 > /tmp/server.log 2>&1 &") 
    # Start client
    net.get('h2').cmd(f"python3 -u ./client.py --host 10.0.0.1 --port 7272 > /tmp/client.log 2>&1 &")

    time.sleep(10)

    net.stop()

if __name__ == '__main__':
    setLogLevel( 'info' )
    runExperiment()
