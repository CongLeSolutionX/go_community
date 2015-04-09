// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

// #include <pthread.h>
// #include <signal.h>
// #include <stdlib.h>
// #include <stdio.h>
//
// static void sigsegv() {
// 	pthread_kill(pthread_self(), SIGSEGV);
// 	fprintf(stderr, "ERROR: SIGSEGV not thrown on caught?.\n");
// 	exit(2);
// }
//
// static void sighandler(int signum) {
// 	if (signum == SIGSEGV) {
// 		exit(0);  // success
// 	}
// }
//
// static void __attribute__ ((constructor)) sigsetup(void) {
// 	struct sigaction act;
// 	act.sa_handler = &sighandler;
// 	sigaction(SIGSEGV, &act, 0);
// }
import "C"

func main() {
	C.sigsegv()
}
