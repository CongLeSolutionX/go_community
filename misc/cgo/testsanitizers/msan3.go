package main

/*
extern int *GoFn(void);

// Yes, you can have definitions if you use //export, as long as they are weak.
int f(void) __attribute__ ((weak));
int g(int *) __attribute__ ((weak));

int f() {
  int *p = GoFn();
  return g(p);
}

int g(int *p) {
  if (*p != 12345)
    return 0;
  return 1;
}
*/
import "C"

//export GoFn
func GoFn() *C.int {
	i := C.int(12345)
	return &i
}

func main() {
	if r := C.f(); r != 1 {
		panic(r)
	}
}
