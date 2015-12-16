// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package big

// TODO(gri) make correct precision decisions (rough sketch so far)

func (x *Float) mantExp10(mant *Float) int {
	a := makeAllocator(mant)

	// x = mant * 2**exp2
	exp2 := int64(x.MantExp(mant)) // 0.5 <= mant < 1.0

	// compute decimal exponent exp10 = exp2 * (ln2/ln10)
	exp10 := a.float().SetInt64(exp2)
	exp10.Mul(exp10, a.ln2())
	exp10.Quo(exp10, a.ln10())

	// We now have a floating-point base 10 exponent.
	// Break into the integer part and the fractional part.
	// The integer part is the exponent we use for the decimal
	// mantissa. The 10**(fractional part) will be multiplied
	// back into the mantissa, if one was provided.
	iexp10, _ := exp10.Int64()

	if mant != nil {
		frac := exp10.Sub(exp10, a.float().SetInt64(iexp10))
		// Now compute 10**(fractional part).
		// Fraction is in base 10. Move it to base e.
		frac.Mul(frac, a.ln10())
		scale := a.float().exponential(frac)
		mant.Mul(mant, scale)
	}

	return int(iexp10)
}

type allocator struct {
	prec uint
}

func makeAllocator(sample *Float) allocator {
	return allocator{uint(sample.prec)}
}

func (a allocator) float() *Float {
	return new(Float).SetPrec(a.prec)
}

func (a allocator) ln2() *Float {
	x, _ := a.float().SetPrec(a.prec).SetString(strLn2)
	return x
}

func (a allocator) ln10() *Float {
	x, _ := a.float().SetPrec(a.prec).SetString(strLn10)
	return x
}

const strLn2 = "0.6931471805599453094172321214581765680755001343602552541206800094933936219696947156058633269964186875" +
	"4200148102057068573368552023575813055703267075163507596193072757082837143519030703862389167347112335" +
	"0115364497955239120475172681574932065155524734139525882950453007095326366642654104239157814952043740" +
	"4303855008019441706416715186447128399681717845469570262716310645461502572074024816377733896385506952" +
	"6066834113727387372292895649354702576265209885969320196505855476470330679365443254763274495125040606" +
	"9438147104689946506220167720424524529612687946546193165174681392672504103802546259656869144192871608" +
	"2938031727143677826548775664850856740776484514644399404614226031930967354025744460703080960850474866" +
	"3852313818167675143866747664789088143714198549423151997354880375165861275352916610007105355824987941" +
	"4729509293113897155998205654392871700072180857610252368892132449713893203784393530887748259701715591" +
	"0708823683627589842589185353024363421436706118923678919237231467232172053401649256872747782344535347" +
	"6481149418642386776774406069562657379600867076257199184734022651462837904883062033061144630073719489" +
	"0027436439650025809365194430411911506080948793067865158870900605203468429736193841289652556539686022" +
	"1941229242075743217574890977067526871158170511370091589426654785959648906530584602586683829400228330" +
	"0538207400567705304678700184162404418833232798386349001563121889560650553151272199398332030751408426" +
	"0914790012651682434438935724727882054862715527418772430024897945401961872339808608316648114909306675" +
	"1933931289043164137068139777649817697486890388778999129650361927071088926410523092478391737350122984" +
	"2420499568935992206602204654941510613918788574424557751020683703086661948089641218680779020818158858" +
	"0001688115973056186676199187395200766719214592236720602539595436541655311295175989940056000366513567" +
	"5690512459268257439464831683326249018038242408242314523061409638057007025513877026817851630690255137" +
	"0323405380214501901537402950994226299577964742713815736380172987394070424217997226696297993931270693" +
	"5747240493386530879758721699645129446491883771156701678598804981838896784134938314014073166472765327" +
	"6359192335112333893387095132090592721854713289754707978913844454666761927028855334234298993218037691" +
	"5497334026754675887323677834291619181043011609169526554785973289176354555674286387746398710191243175" +
	"4255888301206779210280341206879759143081283307230300883494705792496591005860012341561757413272465943" +
	"0684354652111350215443415399553818565227502214245664400062761833032064727257219751529082785684213207" +
	"9598863896727711955221881904660395700977470651261950527893229608893140562543344255239206203034394177" +
	"7357945592125901992559114844024239012554259003129537051922061506434583787873002035414421785758013236" +
	"4516607099143831450049858966885772221486528821694181270488607589722032166631283783291567630749872985" +
	"7463892826937350984077804939500493399876264755070316221613903484529942491724837340613662263834936811" +
	"1684167056925214751383930638455371862687797328895558871634429756244755392366369488877823890174981027" +
	"3565524050"

const strLn10 = "2.30258509299404568401799145468436420760110148862877297603332790096757260967735248023599720508959829" +
	"83419677840422862486334095254650828067566662873690987816894829072083255546808437998948262331985283" +
	"93505308965377732628846163366222287698219886746543667474404243274365155048934314939391479619404400" +
	"22210510171417480036880840126470806855677432162283552201148046637156591213734507478569476834636167" +
	"92101806445070648000277502684916746550586856935673420670581136429224554405758925724208241314695689" +
	"01675894025677631135691929203337658714166023010570308963457207544037084746994016826928280848118428" +
	"93148485249486448719278096762712757753970276686059524967166741834857044225071979650047149510504922" +
	"14776567636938662976979522110718264549734772662425709429322582798502585509785265383207606726317164" +
	"30950599508780752371033310119785754733154142180842754386359177811705430982748238504564801909561029" +
	"92918243182375253577097505395651876975103749708886921802051893395072385392051446341972652872869651" +
	"10862571492198849978748873771345686209167058498078280597511938544450099781311469159346662410718466" +
	"92310107598438319191292230792503747298650929009880391941702654416816335727555703151596113564846546" +
	"19089704281976336583698371632898217440736600916217785054177927636773114504178213766011101073104239" +
	"78325218948988175979217986663943195239368559164471182467532456309125287783309636042629821530408745" +
	"60927760726641354787576616262926568298704957954913954918049209069438580790032763017941503117866862" +
	"09240853794986126493347935487173745167580953708828106745244010589244497647968607512027572418187498" +
	"93959716431055188481952883307466993178146349300003212003277656541304726218839705967944579434683432" +
	"18395304414844803701305753674262153675579814770458031413637793236291560128185336498466942261465206" +
	"45994207291711937060244492935803700771898109736253322454836698850552828596619280509844717519850366" +
	"66808749704969822732202448233430971691111368135884186965493237149969419796878030088504089796185987" +
	"56579894836445212043698216415292987811742973332588607915912510967187510929248475023930572665446276" +
	"20092306879151813580347770129559364629841236649702335517458619556477246185771736936840467657704787" +
	"43197805738532718109338834963388130699455693993461010907456160333122479493604553618491233330637047" +
	"51724871276379140924398331810164737823379692265637682071706935846394531616949411701841938119405416" +
	"44946611127471281970581778329384174223140993002291150236219218672333726838568827353337192510341293" +
	"07056325444266114297653883018223840910261985828884335874559604530045483707890525784731662837019533" +
	"92231047527564998119228742789713715713228319641003422124210082180679525276689858180956119208391760" +
	"72108091992346151695259909947378278064812805879273199389345341532018596971102140754228279629823706" +
	"89417647406422257572124553925261793736524344405605953365915391603125244801493132345724538795243890" +
	"36839236450507881731359711238145323701508413491122324390927681724749607955799151363982881058285740" +
	"5380006533716555530141963322419180876210182049194926514838926922937079"

// ----------------------------------------------------------------------------
// The code below is from Rob Pike's "ivy" APL-like calculator, copied from
// https://github.com/robpike/ivy, adjusted for the conventions and environ-
// ment in this package.
//
// Notable changes from ivy implementation:
//
// - signatures now use incoming z for result as is customary for Float methods
// - result precision is determined from incoming argument rather than config
// - new Floats are allocated via the allocator rather than the config object
// - operations deal with special operands (±Inf, 0, etc.) as necessary
// - config argument was removed
// - log renamed to ln

var floatOne = new(Float).SetUint64(1)

// exponential sets z to e**x, computed using Taylor series, and returns z.
// It converges quickly as long as it is called only for small values of x.
// If z's precision is 0, it is changed to the precision of x before setting z.
func (z *Float) exponential(x *Float) *Float {
	if z.prec == 0 {
		z.prec = x.prec
	}

	// handle special cases
	if x.form == inf {
		if x.neg {
			z.form = zero
		} else {
			z.form = inf
			z.neg = false
		}
		return z
	}

	if x == z {
		t := new(Float).Set(x)
		x = t
	}

	// perform computation with extra bits
	const extra = 64 // TODO(gri) what is the right number?
	z.prec += extra
	z.mode = ToZero
	a := makeAllocator(z)

	// The Taylor series for e**x, exp(x), is 1 + x + x²/2! + x³/3! ...

	z.SetUint64(1)
	term := a.float()
	xN := a.float().Set(x)
	nFactorial := a.float().SetUint64(1)

	for loop := newLoop("exponential", x, 1, 10); ; {
		term.Set(xN)
		term.neg = false // ignore sign
		term.Quo(term, nFactorial)
		z.Add(z, term)

		if loop.done(z) {
			break
		}

		// Advance x**index (multiply by x).
		xN.Mul(xN, x)
		// Advance n!.
		nFactorial.Mul(nFactorial, loop.index())
	}

	// e**(-x) == 1/(e**x)
	if x.neg {
		// Cannot use floatOne below due to initialization cycle
		// TODO(gri) fix this
		one := a.float().SetUint64(1)
		z.Quo(one, z)
	}

	return z.SetMode(ToNearestEven).SetPrec(z.Prec() - extra)

}

// ln sets z to the the natural logarithm of x, computed using the
// Maclaurin series for ln(1-x), and returns z. x must be > 0.
// If z's precision is 0, it is changed to the precision of x before
// setting z.
func (z *Float) ln(x *Float) *Float {
	if z.prec == 0 {
		z.prec = x.prec
	}

	if x.Sign() <= 0 {
		panic(ErrNaN{"ln(x) undefined for x <= 0"})
	}

	if x.IsInf() {
		return z.SetInf(false)
	}

	if x == z {
		t := new(Float).Set(x)
		x = t
	}

	// perform computation with extra bits
	const extra = 64 // TODO(gri) what is the right number?
	z.prec += extra
	a := makeAllocator(z)

	// The series wants x < 1, and ln(1/x) == -ln(x), so exploit that.
	xx := a.float().Set(x)
	invert := false
	switch d := x.Cmp(floatOne); {
	case d == 0:
		// x == 1
		return z.SetUint64(0)
	case d > 0:
		// x > 1
		invert = true
		xx.Quo(floatOne, xx)
	}
	// xx < 1

	// x = mant * 2**exp, and 0.5 <= mant < 1.
	// So ln(x) is ln(mant)+exp*ln(2), and 1-x will be
	// between 0 and 0.5, so the series for 1-x will converge well.
	// (The series converges slowly in general.)
	var mant Float
	exp2 := int64(xx.MantExp(&mant))
	exp := a.float().SetInt64(exp2)
	exp.Mul(exp, a.ln2())
	if invert {
		exp.Neg(exp)
	}

	// y = 1-x (whereupon x = 1-y and we use that in the series).
	y := a.float().SetUint64(1)
	y.Sub(y, &mant)

	// The Maclaurin series for ln(1-y) == ln(x) is: -y - y²/2 - y³/3 ...

	yN := a.float().Set(y)
	term := a.float()

	// This is the slowest-converging series, so we add a factor of ten to the cutoff.
	// Only necessary when FloatPrec is at or beyond constPrecisionInBits.

	for loop := newLoop("ln", x, 1, 10); ; {
		term.Quo(yN, loop.index())
		z.Sub(z, term)
		if loop.done(z) {
			break
		}
		// Advance y**index (multiply by y).
		yN.Mul(yN, y)
	}

	if invert {
		z.Neg(z)
	}
	z.Add(z, exp)

	return z.SetPrec(z.Prec() - extra)
}

type loop struct {
	name        string // name of function we are evaluating, for diagnostic only
	x           *Float // function argument, for precision and diagnostic
	prev, last  *Float // result from previous and last iteration, rounded to desired result precision
	index_      *Float
	i           uint64 // loop count
	itersPerBit uint64
	fixedCount  uint64
}

func newLoop(name string, x *Float, i, itersPerBit uint64) *loop {
	prec := x.Prec()
	return &loop{
		name:        name,
		x:           x,
		prev:        new(Float).SetPrec(prec),
		last:        new(Float).SetPrec(prec),
		index_:      new(Float),
		i:           i,
		itersPerBit: itersPerBit,
	}
}

func (l *loop) index() *Float {
	return l.index_.SetUint64(l.i)
}

func (l *loop) done(z *Float) bool {
	l.last.Set(z) // round z
	if l.prev.Cmp(l.last) == 0 {
		l.fixedCount++
		if l.fixedCount >= l.itersPerBit {
			//println(l.name, l.x.String(), l.i, "iterations")
			return true
		}
	} else {
		l.fixedCount = 0
	}
	l.i++
	if l.i >= 10000 {
		println(l.name, l.x.String(), l.last.String(), l.i)
		panic(0)
	}
	l.prev, l.last = l.last, l.prev
	return false
}

type loop_ struct {
	name          string // The name of the function we are evaluating.
	i             uint64 // Loop count.
	maxIterations uint64 // When to give up.
	stallCount    int    // Iterations since |delta| changed.
	start         *Float // starting value.
	prevZ         *Float // Result from the previous iteration.
	delta         *Float // |Change| from previous iteration.
	prevDelta     *Float // Delta from the previous iteration.
	index_        *Float
}

// newLoop returns a new loop checker. The arguments are the name
// of the function being evaluated, the argument to the function, and
// the maximum number of iterations to perform before giving up.
// The last number in terms of iterations per bit, so the caller can
// ignore the precision setting.
func newLoop_(name string, x *Float, i, itersPerBit uint64) *loop_ {
	a := makeAllocator(x)
	return &loop_{
		name:          name,
		i:             i,
		start:         a.float().Set(x),
		maxIterations: 10 + itersPerBit*uint64(x.prec),
		prevZ:         a.float(),
		delta:         a.float(),
		prevDelta:     a.float(),
		index_:        a.float(),
	}
}

// done reports whether the loop is done. If it does not converge
// after the maximum number of iterations, it errors out.
func (l *loop_) done(z *Float) bool {
	l.delta.Sub(l.prevZ, z)
	if l.delta.Sign() == 0 {
		return true
	}
	if l.delta.Sign() < 0 {
		// Convergence can oscillate when the calculation is nearly
		// done and we're running out of bits. This stops that.
		// See next comment.
		l.delta.Neg(l.delta)
	}
	if l.delta.Cmp(l.prevDelta) == 0 {
		// In freaky cases (like e**3) we can hit the same large positive
		// and then  large negative value (4.5, -4.5) so we count a few times
		// to see that it really has stalled. Avoids having to do hard math,
		// but it means we may iterate a few extra times. Usually, though,
		// iteration is stopped by the zero check above, so this is fine.
		l.stallCount++
		if l.stallCount > 3 {
			// Convergence has stopped.
			return true
		}
	} else {
		l.stallCount = 0
	}
	l.i++
	if l.i == l.maxIterations {
		// Users should never see this.
		panic("unimplemented")
		//Errorf("%s %s: did not converge after %d iterations; prev,last result %s,%s delta %s", l.name, l.start, l.maxIterations, BigFloat{z}, BigFloat{l.prevZ}, BigFloat{l.delta})
	}
	l.prevDelta.Set(l.delta)
	l.prevZ.Set(z)
	return false

}

func (l *loop_) index() *Float {
	return l.index_.SetUint64(l.i)
}
