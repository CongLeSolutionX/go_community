package main

func main() {
	var v int = 5
	var sum int64
	var i int32

	for i = 0; i < int32(v); i++ {
		sum += int64(i)
	}
	i = 0
	if sum != 10 {
		sum = sum / int64(i)
	}
}
