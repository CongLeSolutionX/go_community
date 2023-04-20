#include <windows.h>

int main(int argc, char **argv) {
    if (argc < 3) {
        return 1;
    }
    // Allocate more than 64 TLS indices
    // so the Go runtime doesn't find
    // enough space in the TEB TLS slots.
    for (int i = 0; i < 65; i++) {
        TlsAlloc();
    }
    HMODULE hlib = LoadLibrary(argv[1]);
    if (hlib == NULL) {
        return 2;
    }
    FARPROC proc = GetProcAddress(hlib, argv[2]);
    if (proc == NULL) {
        return 3;
    }
    proc();
    return 0;
}