
// This code is intended to trigger a DWARF "pointer to member" type DIE
struct C { int m; };
int main()
{
    int C::* p = &C::m;
    C c = {7};
    return c.*p;
}
