/*
Package localip provides a function to retrieve the local IP address of the machine by establishing a TCP connection to an external address.

Key Features:
- LocalIP: Determines and returns the local IP address used by the machine for outgoing connections. It does so by dialing a remote address (Google DNS server at 8.8.8.8) and extracting the local IP from the connection's local address.

Functions:
- LocalIP: Establishes a TCP connection to a known remote IP address (8.8.8.8) on port 53 (DNS), then retrieves and returns the local IP address used for the connection. The result is cached after the first lookup to avoid redundant network requests.

Error Handling:
- The function returns an error if it fails to establish the TCP connection or if it encounters any issues while retrieving the local IP.

Usage:
- To obtain the local IP address, call the `LocalIP` function. It will return the local IP address as a string and any error encountered during the process.

Example:
	// Get local IP
	ip, err := localip.LocalIP()
	if err != nil {
		log.Fatal("Error retrieving local IP:", err)
	}
	fmt.Println("Local IP:", ip)

Note:
- This package uses a TCP connection to an external DNS server (8.8.8.8) to determine the local machine's IP, which works even when the machine is behind a router or firewall. It requires network access to the external server.

Credits: https://github.com/TTK4145/Network-go
*/
package localip


import (
	"net"
	"strings"
)

var localIP string

func LocalIP() (string, error) {
	if localIP == "" {
		conn, err := net.DialTCP("tcp4", nil, &net.TCPAddr{IP: []byte{8, 8, 8, 8}, Port: 53})
		if err != nil {
			return "", err
		}
		defer conn.Close()
		localIP = strings.Split(conn.LocalAddr().String(), ":")[0]
	}
	return localIP, nil
}
