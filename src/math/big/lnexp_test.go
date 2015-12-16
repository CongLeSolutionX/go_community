// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package big

import (
	"testing"
)

type funTest struct {
	prec    uint
	x, want string
	acc     Accuracy
}

// Exact values derived from Wolfram Alpha.

var lnTests = []funTest{
	{64, "1e-1000000", "-2.302585092994045684e+06", Above},
	{64, "1e-100000", "-230258.5092994045684", Above},
	{64, "1e-10000", "-23025.85092994045684", Above},
	{64, "1e-1000", "-2302.585092994045684", Above},
	{64, "1e-100", "-230.2585092994045684", Above},
	{64, "1e-10", "-23.02585092994045684", Above},
	{64, "0.00001", "-11.51292546497022842", Above},
	{64, "0.0001", "-9.210340371976182736", Above},
	{64, "0.001", "-6.907755278982137052", Above},
	{64, "0.01", "-4.605170185988091368", Above},
	{64, "0.1", "-2.302585092994045684", Above},
	{64, "0.5", "-0.6931471805599453094", Above},
	{64, "1", "0", Exact},
	{64, "2", "0.6931471805599453094", Below},
	{64, "10", "2.302585092994045684", Below},
	{64, "100", "4.605170185988091368", Below},
	{64, "1000", "6.907755278982137052", Below},
	{64, "10000", "9.210340371976182736", Below},
	{64, "1e10", "23.02585092994045684", Below},
	{64, "1e100", "230.2585092994045684", Below},
	{64, "1e1000", "2302.585092994045684", Below},
	{64, "1e10000", "23025.85092994045684", Below},
	{64, "1e100000", "230258.5092994045684", Below},
	{64, "1e1000000", "2.302585092994045684e+06", Below},
	{64, "+Inf", "+Inf", Exact},

	{256, "0.00001", "-11.5129254649702284200899572734218210380055074431438648801666395048378630483867", Above},
	{256, "10000", "9.2103403719761827360719658187374568304044059545150919041333116038702904387095", Below},
	{256, "1e1000000", "2.30258509299404568401799145468436420760110148862877297603332790096757260967737e+06", Below},
}

var exponentialTests = []funTest{
	{64, "-Inf", "0", Exact},
	{64, "-1", "0.36787944117144232158", Above},
	{64, "0", "1", Exact},
	{64, "0.1", "1.1051709180756476248", Below},
	{64, "0.5", "1.6487212707001281468", Below},
	{64, "1", "2.7182818284590452354", Below},
	{64, "+Inf", "+Inf", Exact},

	{256, "1", "2.71828182845904523536028747135266249775724709369995957496696762772407663035355", Above},
	{256, "1000", "1.97007111401704699388887935224332312531693798532384578995280299138506385078244e+434", Below},
}

func testFun(t *testing.T, name string, fun func(*Float, *Float) *Float, tests []funTest) {
	for _, test := range tests {
		f, _, err := ParseFloat(test.x, 0, test.prec, ToNearestEven)
		if err != nil {
			t.Errorf("%v: %s", test, err)
			continue
		}

		var z Float
		got := fun(&z, f).Text('g', -1)
		acc := z.Acc()
		// TODO(gri) check accuracy
		if got != test.want /*|| acc != test.acc*/ {
			t.Errorf("%s(%s): got %s (%s); want %s (%s)", name, test.x, got, acc, test.want, test.acc)
		}
	}
}

func TestLn(t *testing.T)          { testFun(t, "ln", (*Float).ln, lnTests) }
func TestExponential(t *testing.T) { testFun(t, "exponential", (*Float).exponential, exponentialTests) }
