#ifndef MAIN_H
#define MAIN_H

#define SERVER_PORT 8080


void init_read(int socket_fd);

#include <sys/socket.h>
void init_accept(int epollfd ,int socket_fd , struct sockaddr* addr ,  socklen_t addr_len);


void log_conn(struct sockaddr* addr);

#endif 
