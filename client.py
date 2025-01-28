#!/usr/bin/env python3
import socket
import argparse

def start_client(host, port):
    client_socket = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
    client_socket.connect((host, port))  # Connect to the server

    # Send a hello message to the server
    client_message = "Hello from client!"
    client_socket.sendall(client_message.encode('utf-8'))
    print("Client sent a hello message")

    # Receive a message from the server
    server_message = client_socket.recv(1024).decode('utf-8')
    print(f"Server says: {server_message}")

    # Close socket
    client_socket.close()
    print("Client socket closed")

if __name__ == "__main__":
    print("IN MAIN OF CLIENT")
    parser = argparse.ArgumentParser(description="Start a TCP client.")
    parser.add_argument('--host', type=str, default='127.0.0.1', help="Server IP address (default: 127.0.0.1)")
    parser.add_argument('--port', type=int, default=65432, help="Server port (default: 65432)")
    args = parser.parse_args()

    start_client(args.host, args.port)
    print("END OF CLIENT MAIN")
