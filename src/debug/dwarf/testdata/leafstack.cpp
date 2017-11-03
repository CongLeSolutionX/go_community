
inline int leaf(int lx, int ly)
{
  int lv[64];
  for (unsigned i = 0; i < 64; ++i)
    lv[i] = i;
  lv[lx&3] += 3;
  return lv[ly&3];
}

inline int mid(int mx, int my)
{
  int mv[64];
  for (unsigned i = 0; i < 64; ++i)
    mv[i] = i;
  mv[mx&3] += leaf(mx, my);
  return mv[my&3];
}

int main(int argc, char **argv) {
  return mid(1, argc);
}

typedef int (*mfn)(int, int);

mfn G = mid;
