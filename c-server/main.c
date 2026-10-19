#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <sys/socket.h>
#include <netinet/in.h>
#include <unistd.h>
#include <arpa/inet.h>
#include <asm-generic/errno-base.h>
#include<errno.h>
#include <sys/epoll.h>

#include "main.h"
#include "epoll.h"


int main(void)
{
	int socket_fd; 
	struct sockaddr_in address;

	socket_fd=socket(AF_INET , SOCK_STREAM, 0);
	if (socket_fd < 0) {
		perror("socket creation failed");
		exit(EXIT_FAILURE);
	}


	memset(&address, 0, sizeof(address));
	address.sin_family = AF_INET;
	address.sin_port = htons(SERVER_PORT);
	address.sin_addr.s_addr = INADDR_ANY;


	if (bind(socket_fd, (struct sockaddr *)&address, sizeof(address)) < 0){
		perror("socket bind failed");
		close(socket_fd);
		exit(EXIT_FAILURE);
	}


	printf("Socket successfully bound to port %d\n" , SERVER_PORT);


	int listen_fd =  listen(socket_fd, 10);
	if (listen_fd < 0) {
		perror("listener init failed");
		close(socket_fd);
		exit(EXIT_FAILURE);
	}

	int epollfd = must_setup_epoll(socket_fd);

	init_accept(
		epollfd,
		socket_fd, 
		(struct sockaddr *)&address , sizeof(address)
	);

	close(socket_fd);

	return 0;
}


void init_accept(int epollfd ,int socket_fd , struct sockaddr* addr ,  socklen_t addr_len) 
{
	for (;;) {
		int ndfs = must_watch_epoll(epollfd);

		for (int i = 0 ; i < ndfs ; i++) {
			int curr_fd = events[i].data.fd;

			if(curr_fd== socket_fd) {

				int conn_sockfd = accept(socket_fd , addr , &addr_len);
				if(conn_sockfd < 0) {
					perror("error while acceptng connections");
					continue;
				}

				log_conn(addr);
				must_register_conn_sock(epollfd, conn_sockfd);
			}else {
				char buffer[10324];

				int bytes_read = recv(curr_fd , buffer , sizeof(buffer) , 0);
				if (bytes_read < 0) {

					if (errno == EINTR) { 
						printf("connection interrrupted");
						continue; 
					}
					perror("failed to read bytes");
					exit(EXIT_FAILURE);
				}

				if (bytes_read == 0) return; 

				printf("bytes read: %d  with content: %s\n" ,(int)bytes_read , buffer);

			}
		}


	}
}

void log_conn(struct sockaddr* addr) 
{
	char * client_addr = inet_ntoa(((struct sockaddr_in*)addr)->sin_addr);
	int client_port = ntohs(((struct sockaddr_in*)addr)->sin_port);

	printf(
		"Connection accepted from %s:%d\n", 
		client_addr,
		client_port
	);
}






