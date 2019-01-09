
int i = 3;
double d = 3;

// anonymous reference type
int &bad = i;

// named reference type
typedef double &dref;
dref dr = d;

int main() {
        return 0;
}
