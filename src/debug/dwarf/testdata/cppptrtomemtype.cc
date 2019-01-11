// Copyright 2018 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// This code is intended to trigger a DWARF "pointer to member" type DIE
struct CS { int dm; };
int main()
{
  int CS::* pdm = &CS::dm;
  CS cs = {42};
  return cs.*pdm;
}
