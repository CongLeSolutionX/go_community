package big

import (
	"fmt"
	"github.com/jfcg/sponge"
	"testing"
)

func fillRnd(n nat) {
	for i := 0; i < len(n); i++ {
		n[i] = Word(mpr.I())
	}
	n[0] |= 1 // odd

	l := len(n) - 1
	for n[l] == 0 { // max Word != 0
		n[l] = Word(mpr.I())
	}
}

var mpr = sponge.NewPrng(3, 12, 3)

const nob = 1000

func timeMB(ben *testing.B, x *Int, nr int, res *[nob]bool) {
	var np int
	ben.ResetTimer()
	for k := 0; k < ben.N; k++ {
		mpr.Reset()
		np = 0
		for i := 0; i < nob; i++ {
			fillRnd(x.abs)
			res[i] = x.ProbablyPrime(nr)
			if res[i] {
				np++
			}
		}
	}
	ben.StopTimer()

	if ben.N > 1 {
		return
	}
	bn, bx := 9999, 0 // run these for ben.N=1
	mpr.Reset()
	for i := 0; i < nob; i++ {
		fillRnd(x.abs)
		bl := x.abs.bitLen()
		if bl > bx {
			bx = bl
		}
		if bl < bn {
			bn = bl
		}
	}
	fmt.Println("n=", nr, "minBits=", bn, "maxBits=", bx, "#ofPrimes=", np)
}

var mr1, mr2 [nob]bool

func BenchmarkManyBig1(ben *testing.B) {
	x := &Int{false, make(nat, 8)}
	timeMB(ben, x, 1, &mr1) // MR(n=1), 512 bits inputs
}

func BenchmarkManyBig2(ben *testing.B) {
	x := &Int{false, make(nat, 8)}
	timeMB(ben, x, 0, &mr2) // BPSW, 512 bits inputs

	for i := 0; i < nob; i++ {
		if mr1[i] != mr2[i] {
			ben.Fatal("512 bits results not identical", i)
		}
	}
}

func BenchmarkManyBig3(ben *testing.B) {
	x := &Int{false, make(nat, 16)}
	timeMB(ben, x, 1, &mr1) // MR(n=1), 1024 bits inputs
}

func BenchmarkManyBig4(ben *testing.B) {
	x := &Int{false, make(nat, 16)}
	timeMB(ben, x, 0, &mr2) // BPSW, 1024 bits inputs

	for i := 0; i < nob; i++ {
		if mr1[i] != mr2[i] {
			ben.Fatal("1024 bits results not identical", i)
		}
	}
}
