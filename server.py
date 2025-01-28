#!/usr/bin/env python3
import socket
import argparse

def start_server(host, port):
    server_socket = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
    server_socket.bind((host, port))
    server_socket.setsockopt(socket.SOL_SOCKET, socket.SO_REUSEADDR, 1)
    server_socket.listen(1)  # Allow one connection
    print(f"Server listening on {host}:{port}")

    conn, addr = server_socket.accept()  # Accept a connection
    print(f"Connected by {addr}")
    
    # Receive message from client
    client_message = conn.recv(1024).decode('utf-8')
    print(f"Client says: {client_message}")

    # Send a response back to the client
    server_message = "Hello from server!"
    conn.sendall(server_message.encode('utf-8'))
    print("Server sent a hello message")
    
    # Close sockets
    conn.close()
    server_socket.close()
    print("Server socket closed")

if __name__ == "__main__":
    parser = argparse.ArgumentParser(description="Start a TCP server.")
    parser.add_argument('--host', type=str, default='127.0.0.1', help="Server IP address (default: 127.0.0.1)")
    parser.add_argument('--port', type=int, default=65432, help="Server port (default: 65432)")
    args = parser.parse_args()

    print("IN MAIN OF SERVER")
    print(f"args.host = {args.host}")
    print(f"args.port = {args.port}")
    start_server(args.host, args.port)
    print("AFTER SOCKET CLOSES")
