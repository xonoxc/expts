#include <fcntl.h>
#include "epoll.h"

int set_non_blocking(int sockfd) {
	int flag = fcntl(sockfd, F_GETFL , 0);
	if (flag == -1)
		return -1;

	return fcntl(sockfd, F_SETFL, flag | O_NONBLOCK);
}



