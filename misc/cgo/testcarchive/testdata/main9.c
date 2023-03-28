// Copyright 2023 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

#include "libgo9.h"

void callGoFWithDeepStack(int n) {
	if (n > 0) {
		callGoFWithDeepStack(n - 1);
	}
	GoF();
}

int main() {
	GoF();                        // call GoF without using much stack
	callGoFWithDeepStack(100000); // call GoF with a deep stack
}
