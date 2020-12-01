package a
func f() bool { return true }
func G() func() func() bool { return func() func() bool { return f } }
