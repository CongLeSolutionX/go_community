package reflect_test

import (
	. "reflect"
	"testing"
)

type structField struct {
	name  string
	index []int
}

var fieldsTests = []struct {
	testName string
	val      interface{}
	expect   []structField
}{{
	testName: "simple-struct",
	val: struct {
		A int
		B string
		C bool
	}{},
	expect: []structField{{
		name:  "A",
		index: []int{0},
	}, {
		name:  "B",
		index: []int{1},
	}, {
		name:  "C",
		index: []int{2},
	}},
}, {
	testName: "non-embedded-struct-member",
	val: struct {
		A struct {
			X int
		}
	}{},
	expect: []structField{{
		name:  "A",
		index: []int{0},
	}},
}, {
	testName: "embedded-exported-struct",
	val: struct {
		SFG
	}{},
	expect: []structField{{
		name:  "SFG",
		index: []int{0},
	}, {
		name:  "F",
		index: []int{0, 0},
	}, {
		name:  "G",
		index: []int{0, 1},
	}},
}, {
	testName: "embedded-unexported-struct",
	val: struct {
		sFG
	}{},
	expect: []structField{{
		name:  "sFG",
		index: []int{0},
	}, {
		name:  "F",
		index: []int{0, 0},
	}, {
		name:  "G",
		index: []int{0, 1},
	}},
}, {
	testName: "two-embedded-structs-with-cancelling-members",
	val: struct {
		SFG
		SF
	}{},
	expect: []structField{{
		name:  "SFG",
		index: []int{0},
	}, {
		name:  "G",
		index: []int{0, 1},
	}, {
		name:  "SF",
		index: []int{1},
	}},
}, {
	testName: "embedded-structs-with-same-fields-at-different-depths",
	val: struct {
		SFGH3
		SG1
		SFG2
		SF2
		L int
	}{},
	expect: []structField{{
		name:  "SFGH3",
		index: []int{0},
	}, {
		name:  "SFGH2",
		index: []int{0, 0},
	}, {
		name:  "SFGH1",
		index: []int{0, 0, 0},
	}, {
		name:  "SFGH",
		index: []int{0, 0, 0, 0},
	}, {
		name:  "H",
		index: []int{0, 0, 0, 0, 2},
	}, {
		name:  "SG1",
		index: []int{1},
	}, {
		name:  "SG",
		index: []int{1, 0},
	}, {
		name:  "G",
		index: []int{1, 0, 0},
	}, {
		name:  "SFG2",
		index: []int{2},
	}, {
		name:  "SFG1",
		index: []int{2, 0},
	}, {
		name:  "SFG",
		index: []int{2, 0, 0},
	}, {
		name:  "SF2",
		index: []int{3},
	}, {
		name:  "SF1",
		index: []int{3, 0},
	}, {
		name:  "SF",
		index: []int{3, 0, 0},
	}, {
		name:  "L",
		index: []int{4},
	}},
}, {
	testName: "embedded-pointer-struct",
	val: struct {
		*SF
	}{},
	expect: []structField{{
		name:  "SF",
		index: []int{0},
	}, {
		name:  "F",
		index: []int{0, 0},
	}},
}, {
	testName: "embedded-not-a-pointer",
	val: struct {
		M
	}{},
	expect: []structField{{
		name:  "M",
		index: []int{0},
	}},
}, {
	testName: "recursive-embedding",
	val:      Rec1{},
	expect: []structField{{
		name:  "Rec2",
		index: []int{0},
	}, {
		name:  "F",
		index: []int{0, 0},
	}, {
		name:  "Rec1",
		index: []int{0, 1},
	}},
}, {
	testName: "recursive-embedding-2",
	val:      Rec2{},
	expect: []structField{{
		name:  "F",
		index: []int{0},
	}, {
		name:  "Rec1",
		index: []int{1},
	}, {
		name:  "Rec2",
		index: []int{1, 0},
	}},
}}

type SFG struct {
	F int `httprequest:",form"`
	G int `httprequest:",form"`
}

type SFG1 struct {
	SFG
}

type SFG2 struct {
	SFG1
}

type SFGH struct {
	F int `httprequest:",form"`
	G int `httprequest:",form"`
	H int `httprequest:",form"`
}

type SFGH1 struct {
	SFGH
}

type SFGH2 struct {
	SFGH1
}

type SFGH3 struct {
	SFGH2
}

type SF struct {
	F int `httprequest:",form"`
}

type SF1 struct {
	SF
}

type SF2 struct {
	SF1
}

type SG struct {
	G int `httprequest:",form"`
}

type SG1 struct {
	SG
}

type sFG struct {
	F int `httprequest:",form"`
	G int `httprequest:",form"`
}

type M map[string]interface{}

type Rec1 struct {
	*Rec2
}

type Rec2 struct {
	F string
	*Rec1
}

func TestFields(t *testing.T) {
	for _, test := range fieldsTests {
		test := test
		t.Run(test.testName, func(t *testing.T) {
			typ := TypeOf(test.val)
			fields := VisibleFields(typ)
			if got, want := len(fields), len(test.expect); got != want {
				t.Fatalf("unexpected field count; got %d want %d", got, want)
			}

			for j, field := range fields {
				expect := test.expect[j]
				t.Logf("field %d: %s", j, expect.name)
				gotField := typ.FieldByIndex(field.Index)
				// Unfortunately, FieldByIndex does not return
				// a field with the same index that we passed in,
				// so we set it to the expected value so that
				// it can be compared later with the result of FieldByName.
				gotField.Index = field.Index
				expectField := typ.FieldByIndex(expect.index)
				// ditto.
				expectField.Index = expect.index
				if !DeepEqual(gotField, expectField) {
					t.Fatalf("unexpected field result; got %#v want %#v", gotField, expectField)
				}

				// Sanity check that we can actually access the field by the
				// expected name.
				gotField1, ok := typ.FieldByName(expect.name)
				if !ok {
					t.Fatalf("field %q not accessible by name", expect.name)
				}
				if !DeepEqual(gotField1, expectField) {
					t.Fatalf("unexpected FieldByName result; got %#v want %#v", gotField1, expectField)
				}
			}
		})
	}
}
