package json

import (
	"testing"
)

func testEqual[T jsonType](t *testing.T, tt testType[T]) {
	eq, err := Equal(tt.s1, tt.s2, tt.sN...)

	if err != nil {
		t.Errorf("%T:%d; err:%s\n", tt, tt.id, err)
		return
	}

	if eq != tt.want {
		t.Errorf("%T:%d; want: %t got: %t\n", tt, tt.id, tt.want, eq)
	}
}

func TestEqual(t *testing.T) {

	byteCases := getTestcases[[]byte]()

	for i, tt := range byteCases {
		tt.id = i
		testEqual(t, tt)
	}

	strCases := getTestcases[string]()

	for i, tt := range strCases {
		tt.id = i
		testEqual(t, tt)
	}

}

type testType[T jsonType] struct {
	want bool
	s1   T
	s2   T
	sN   []T
	id   int
}

func getTestcases[T jsonType]() (testcases []testType[T]) {
	ipTestcases := []testType[string]{
		{
			want: false,
			s1:   `{"name":"Hiro","email":"laciferin@gmail.com","age":19}`,
			s2:   `{"name":"Hiro","age":19,"email":"laciferin@gmail.com"}`,
		},
	}

	for _, ip := range ipTestcases {

		testcase := testType[T]{
			want: ip.want,
			s1:   T(ip.s1),
			s2:   T(ip.s2),
		}

		for _, sI := range ip.sN {
			testcase.sN = append(testcase.sN, T(sI))
		}

		testcases = append(testcases, testcase)

	}

	return
}
