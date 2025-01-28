from mininet.log import setLogLevel
from mininet.net import Mininet
from mininet.link import TCLink
import mininet.node
import time

def create_topology():
    setLogLevel('info')
    net = Mininet(link=TCLink)

    h1 = net.addHost('h1') # Server
    h2 = net.addHost('h2') # Client

    net.addLink(h1, h2, cls=TCLink)
    
    net.build()
    
    h1.cmd("ifconfig h1-eth0 10.0.0.1 netmask 255.255.255.0")
    h2.cmd("ifconfig h2-eth0 10.0.0.2 netmask 255.255.255.0")

    return net

if __name__ == '__main__':

    net = create_topology()
    net.get('h1').cmd(f"sudo python3 -u ./server.py --host 10.0.0.1 --port 7272 > /tmp/server.log 2>&1 &")
    net.get('h2').cmd(f"sudo python3 -u ./client.py --host 10.0.0.1 --port 7272 > /tmp/client.log 2>&1 &") 

    time.sleep(10)
    net.stop()
