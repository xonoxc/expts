#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <sys/socket.h>
#include <netinet/in.h>
#include <unistd.h>
#include <arpa/inet.h>
#include <asm-generic/errno-base.h>
#include<errno.h>



void init_read(int socket_fd);
void init_accept(int socket_fd , struct sockaddr* addr ,  socklen_t addr_len);


int main(int argc, char *argv[])
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
	address.sin_port = htons(8080);
	address.sin_addr.s_addr = INADDR_ANY;


	if (bind(socket_fd, (struct sockaddr *)&address, sizeof(address)) < 0){
		perror("socket bind failed");
		close(socket_fd);
		exit(EXIT_FAILURE);
	}


	printf("Socket successfully bound to port 8080\n");


	int listen_fd =  listen(socket_fd, 10);
	if (listen_fd < 0) {
		perror("listener init failed");
		close(socket_fd);
		exit(EXIT_FAILURE);
	}


	init_accept(
		socket_fd, 
		(struct sockaddr *)&address , sizeof(address)
	);

	close(socket_fd);

	return 0;
}


void init_accept(int socket_fd , struct sockaddr* addr ,  socklen_t addr_len) 
{
	for (;;) {
		int conn_sock_fd = accept(socket_fd , addr , &addr_len);
		if(conn_sock_fd < 0) {
			perror("error while acceptng connections");
			continue;
		}


		char * client_addr = inet_ntoa(((struct sockaddr_in*)addr)->sin_addr);
		int client_port = ntohs(((struct sockaddr_in*)addr)->sin_port);

		printf(
			"Connection accepted from %s:%d\n", 
			client_addr,
			client_port
		);

		init_read(conn_sock_fd);
	}
}


void init_read(int socket_fd) 
{
	char buffer[10324];

	int retry_count = 0;

	for(;;){
		int bytes_read = recv(socket_fd , buffer , sizeof(buffer) , 0);
		if (bytes_read < 0) {

			if (errno == EINTR) { 
				printf("connection interrrupted");
				continue; 
			}
			perror("failed to read bytes");
			exit(EXIT_FAILURE);
		}


		if  (bytes_read == 0){ return; }

		printf("bytes read: %d  with content: %s\n" ,(int)bytes_read , buffer);
	}
}




