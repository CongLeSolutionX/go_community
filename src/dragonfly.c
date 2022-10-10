#define _XOPEN_SOURCE 600
#include <fcntl.h>
#include <stdlib.h>
#include <stdio.h>
#include <sys/types.h>
#include <sys/wait.h>
#include <unistd.h>

// Very similar to os/signal.TestDragonfly, but does not reproduce hang.

int main(int argc, char **argv) {
  if (getenv("GO_TEST_DRAGONFLY")) {
    printf("child exiting\n");
    exit(0);
  }

  int fd = posix_openpt(O_RDWR);
  if (fd < 0) {
    perror("posix_openpt");
    exit(1);
  }

  int ret = grantpt(fd);
  if (ret < 0) {
    perror("grantpt");
    exit(1);
  }

  ret = unlockpt(fd);
  if (ret < 0) {
    perror("unlockpt");
    exit(1);
  }

  char *name = ptsname(fd);
  int pts_fd = open(name, O_RDWR, 0);
  if (pts_fd < 0) {
    perror("open");
    exit(1);
  }

  pid_t pid = fork();
  if (pid == 0) {
    // child
    char *child_argv[] = {argv[0], NULL};
    char *child_envv[] = {"GO_TEST_DRAGONFLY=1", NULL};
    execve(argv[0], child_argv, child_envv);
    exit(42); // unreachable
  }

  char b = '\n';
  ret = write(fd, &b, 1);
  if (ret < 0) {
    perror("write");
    exit(1);
  }

  // Matches src/os/wait_wait6.go.
  pid_t child = wait6(P_PID, pid, NULL, WEXITED|WNOWAIT, NULL, NULL);
  if (child < 0) {
    perror("wait6");
    exit(1);
  }

  return 0;
}
