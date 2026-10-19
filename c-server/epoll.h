#include <sys/epoll.h>

#ifndef EPOLL_H
#define EPOLL_H

#define MAX_EVENTS 10
#define MAX_CANDIDATE_SOCKETS 100

extern struct epoll_event ev, events[MAX_EVENTS];

int must_setup_epoll(int listening_soc_fd);
int must_watch_epoll(int epollfd);


void must_register_conn_sock(int epollfd , int conn_sockfd);

#endif
