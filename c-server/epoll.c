#include "epoll.h" 
#include <stdio.h>
#include <sys/epoll.h>
#include <stdlib.h>

#include "funcs.h"


struct epoll_event ev, events[MAX_EVENTS];

int must_setup_epoll(int listening_soc_fd)
{
	int epollfd = epoll_create(MAX_CANDIDATE_SOCKETS);
	if(epollfd == -1) {
		perror("cannot initialize epoll");
		exit(EXIT_FAILURE);
	}

	ev.events = EPOLLIN;
	ev.data.fd = listening_soc_fd;


	if(epoll_ctl(epollfd, EPOLL_CTL_ADD, listening_soc_fd, &ev)==-1) {
		perror("epoll sock registeration failed");
		exit(EXIT_FAILURE);
	}

	return epollfd;
}




void must_register_conn_sock(int epollfd , int conn_sockfd)
{
	set_non_blocking(conn_sockfd);

	ev.events = EPOLLIN | EPOLLET;
	ev.data.fd = conn_sockfd;


	if (epoll_ctl(epollfd, EPOLL_CTL_ADD, conn_sockfd, &ev) == -1) {
		perror("epoll_ctl: conn_sock");
		exit(EXIT_FAILURE);
	}
}



int must_watch_epoll(int epollfd) {
	int nfds = epoll_wait(epollfd, events, MAX_EVENTS, -1);
	if (nfds == -1) {
		perror("epoll_wait");
		exit(EXIT_FAILURE);
	}
	return nfds;
}
