package main

import "fmt"

var failed = false

//go:noinline
func testMultiplication(r [100]int64, s int64) {
	if want, got := r[0], s*0; want != got {
		failed = true
		fmt.Printf("got %d * 0 == %d, wanted %d\n", s, got, want)
	}
	if want, got := r[1], s*1; want != got {
		failed = true
		fmt.Printf("got %d * 1 == %d, wanted %d\n", s, got, want)
	}
	if want, got := r[2], s*2; want != got {
		failed = true
		fmt.Printf("got %d * 2 == %d, wanted %d\n", s, got, want)
	}
	if want, got := r[3], s*3; want != got {
		failed = true
		fmt.Printf("got %d * 3 == %d, wanted %d\n", s, got, want)
	}
	if want, got := r[4], s*4; want != got {
		failed = true
		fmt.Printf("got %d * 4 == %d, wanted %d\n", s, got, want)
	}
	if want, got := r[5], s*5; want != got {
		failed = true
		fmt.Printf("got %d * 5 == %d, wanted %d\n", s, got, want)
	}
	if want, got := r[6], s*6; want != got {
		failed = true
		fmt.Printf("got %d * 6 == %d, wanted %d\n", s, got, want)
	}
	if want, got := r[7], s*7; want != got {
		failed = true
		fmt.Printf("got %d * 7 == %d, wanted %d\n", s, got, want)
	}
	if want, got := r[8], s*8; want != got {
		failed = true
		fmt.Printf("got %d * 8 == %d, wanted %d\n", s, got, want)
	}
	if want, got := r[9], s*9; want != got {
		failed = true
		fmt.Printf("got %d * 9 == %d, wanted %d\n", s, got, want)
	}
	if want, got := r[10], s*10; want != got {
		failed = true
		fmt.Printf("got %d * 10 == %d, wanted %d\n", s, got, want)
	}
	if want, got := r[11], s*11; want != got {
		failed = true
		fmt.Printf("got %d * 11 == %d, wanted %d\n", s, got, want)
	}
	if want, got := r[12], s*12; want != got {
		failed = true
		fmt.Printf("got %d * 12 == %d, wanted %d\n", s, got, want)
	}
	if want, got := r[13], s*13; want != got {
		failed = true
		fmt.Printf("got %d * 13 == %d, wanted %d\n", s, got, want)
	}
	if want, got := r[14], s*14; want != got {
		failed = true
		fmt.Printf("got %d * 14 == %d, wanted %d\n", s, got, want)
	}
	if want, got := r[15], s*15; want != got {
		failed = true
		fmt.Printf("got %d * 15 == %d, wanted %d\n", s, got, want)
	}
	if want, got := r[16], s*16; want != got {
		failed = true
		fmt.Printf("got %d * 16 == %d, wanted %d\n", s, got, want)
	}
	if want, got := r[17], s*17; want != got {
		failed = true
		fmt.Printf("got %d * 17 == %d, wanted %d\n", s, got, want)
	}
	if want, got := r[18], s*18; want != got {
		failed = true
		fmt.Printf("got %d * 18 == %d, wanted %d\n", s, got, want)
	}
	if want, got := r[19], s*19; want != got {
		failed = true
		fmt.Printf("got %d * 19 == %d, wanted %d\n", s, got, want)
	}
	if want, got := r[20], s*20; want != got {
		failed = true
		fmt.Printf("got %d * 20 == %d, wanted %d\n", s, got, want)
	}
	if want, got := r[21], s*21; want != got {
		failed = true
		fmt.Printf("got %d * 21 == %d, wanted %d\n", s, got, want)
	}
	if want, got := r[22], s*22; want != got {
		failed = true
		fmt.Printf("got %d * 22 == %d, wanted %d\n", s, got, want)
	}
	if want, got := r[23], s*23; want != got {
		failed = true
		fmt.Printf("got %d * 23 == %d, wanted %d\n", s, got, want)
	}
	if want, got := r[24], s*24; want != got {
		failed = true
		fmt.Printf("got %d * 24 == %d, wanted %d\n", s, got, want)
	}
	if want, got := r[25], s*25; want != got {
		failed = true
		fmt.Printf("got %d * 25 == %d, wanted %d\n", s, got, want)
	}
	if want, got := r[26], s*26; want != got {
		failed = true
		fmt.Printf("got %d * 26 == %d, wanted %d\n", s, got, want)
	}
	if want, got := r[27], s*27; want != got {
		failed = true
		fmt.Printf("got %d * 27 == %d, wanted %d\n", s, got, want)
	}
	if want, got := r[28], s*28; want != got {
		failed = true
		fmt.Printf("got %d * 28 == %d, wanted %d\n", s, got, want)
	}
	if want, got := r[29], s*29; want != got {
		failed = true
		fmt.Printf("got %d * 29 == %d, wanted %d\n", s, got, want)
	}
	if want, got := r[30], s*30; want != got {
		failed = true
		fmt.Printf("got %d * 30 == %d, wanted %d\n", s, got, want)
	}
	if want, got := r[31], s*31; want != got {
		failed = true
		fmt.Printf("got %d * 31 == %d, wanted %d\n", s, got, want)
	}
	if want, got := r[32], s*32; want != got {
		failed = true
		fmt.Printf("got %d * 32 == %d, wanted %d\n", s, got, want)
	}
	if want, got := r[33], s*33; want != got {
		failed = true
		fmt.Printf("got %d * 33 == %d, wanted %d\n", s, got, want)
	}
	if want, got := r[34], s*34; want != got {
		failed = true
		fmt.Printf("got %d * 34 == %d, wanted %d\n", s, got, want)
	}
	if want, got := r[35], s*35; want != got {
		failed = true
		fmt.Printf("got %d * 35 == %d, wanted %d\n", s, got, want)
	}
	if want, got := r[36], s*36; want != got {
		failed = true
		fmt.Printf("got %d * 36 == %d, wanted %d\n", s, got, want)
	}
	if want, got := r[37], s*37; want != got {
		failed = true
		fmt.Printf("got %d * 37 == %d, wanted %d\n", s, got, want)
	}
	if want, got := r[38], s*38; want != got {
		failed = true
		fmt.Printf("got %d * 38 == %d, wanted %d\n", s, got, want)
	}
	if want, got := r[39], s*39; want != got {
		failed = true
		fmt.Printf("got %d * 39 == %d, wanted %d\n", s, got, want)
	}
	if want, got := r[40], s*40; want != got {
		failed = true
		fmt.Printf("got %d * 40 == %d, wanted %d\n", s, got, want)
	}
	if want, got := r[41], s*41; want != got {
		failed = true
		fmt.Printf("got %d * 41 == %d, wanted %d\n", s, got, want)
	}
	if want, got := r[42], s*42; want != got {
		failed = true
		fmt.Printf("got %d * 42 == %d, wanted %d\n", s, got, want)
	}
	if want, got := r[43], s*43; want != got {
		failed = true
		fmt.Printf("got %d * 43 == %d, wanted %d\n", s, got, want)
	}
	if want, got := r[44], s*44; want != got {
		failed = true
		fmt.Printf("got %d * 44 == %d, wanted %d\n", s, got, want)
	}
	if want, got := r[45], s*45; want != got {
		failed = true
		fmt.Printf("got %d * 45 == %d, wanted %d\n", s, got, want)
	}
	if want, got := r[46], s*46; want != got {
		failed = true
		fmt.Printf("got %d * 46 == %d, wanted %d\n", s, got, want)
	}
	if want, got := r[47], s*47; want != got {
		failed = true
		fmt.Printf("got %d * 47 == %d, wanted %d\n", s, got, want)
	}
	if want, got := r[48], s*48; want != got {
		failed = true
		fmt.Printf("got %d * 48 == %d, wanted %d\n", s, got, want)
	}
	if want, got := r[49], s*49; want != got {
		failed = true
		fmt.Printf("got %d * 49 == %d, wanted %d\n", s, got, want)
	}
	if want, got := r[50], s*50; want != got {
		failed = true
		fmt.Printf("got %d * 50 == %d, wanted %d\n", s, got, want)
	}
	if want, got := r[51], s*51; want != got {
		failed = true
		fmt.Printf("got %d * 51 == %d, wanted %d\n", s, got, want)
	}
	if want, got := r[52], s*52; want != got {
		failed = true
		fmt.Printf("got %d * 52 == %d, wanted %d\n", s, got, want)
	}
	if want, got := r[53], s*53; want != got {
		failed = true
		fmt.Printf("got %d * 53 == %d, wanted %d\n", s, got, want)
	}
	if want, got := r[54], s*54; want != got {
		failed = true
		fmt.Printf("got %d * 54 == %d, wanted %d\n", s, got, want)
	}
	if want, got := r[55], s*55; want != got {
		failed = true
		fmt.Printf("got %d * 55 == %d, wanted %d\n", s, got, want)
	}
	if want, got := r[56], s*56; want != got {
		failed = true
		fmt.Printf("got %d * 56 == %d, wanted %d\n", s, got, want)
	}
	if want, got := r[57], s*57; want != got {
		failed = true
		fmt.Printf("got %d * 57 == %d, wanted %d\n", s, got, want)
	}
	if want, got := r[58], s*58; want != got {
		failed = true
		fmt.Printf("got %d * 58 == %d, wanted %d\n", s, got, want)
	}
	if want, got := r[59], s*59; want != got {
		failed = true
		fmt.Printf("got %d * 59 == %d, wanted %d\n", s, got, want)
	}
	if want, got := r[60], s*60; want != got {
		failed = true
		fmt.Printf("got %d * 60 == %d, wanted %d\n", s, got, want)
	}
	if want, got := r[61], s*61; want != got {
		failed = true
		fmt.Printf("got %d * 61 == %d, wanted %d\n", s, got, want)
	}
	if want, got := r[62], s*62; want != got {
		failed = true
		fmt.Printf("got %d * 62 == %d, wanted %d\n", s, got, want)
	}
	if want, got := r[63], s*63; want != got {
		failed = true
		fmt.Printf("got %d * 63 == %d, wanted %d\n", s, got, want)
	}
	if want, got := r[64], s*64; want != got {
		failed = true
		fmt.Printf("got %d * 64 == %d, wanted %d\n", s, got, want)
	}
	if want, got := r[65], s*65; want != got {
		failed = true
		fmt.Printf("got %d * 65 == %d, wanted %d\n", s, got, want)
	}
	if want, got := r[66], s*66; want != got {
		failed = true
		fmt.Printf("got %d * 66 == %d, wanted %d\n", s, got, want)
	}
	if want, got := r[67], s*67; want != got {
		failed = true
		fmt.Printf("got %d * 67 == %d, wanted %d\n", s, got, want)
	}
	if want, got := r[68], s*68; want != got {
		failed = true
		fmt.Printf("got %d * 68 == %d, wanted %d\n", s, got, want)
	}
	if want, got := r[69], s*69; want != got {
		failed = true
		fmt.Printf("got %d * 69 == %d, wanted %d\n", s, got, want)
	}
	if want, got := r[70], s*70; want != got {
		failed = true
		fmt.Printf("got %d * 70 == %d, wanted %d\n", s, got, want)
	}
	if want, got := r[71], s*71; want != got {
		failed = true
		fmt.Printf("got %d * 71 == %d, wanted %d\n", s, got, want)
	}
	if want, got := r[72], s*72; want != got {
		failed = true
		fmt.Printf("got %d * 72 == %d, wanted %d\n", s, got, want)
	}
	if want, got := r[73], s*73; want != got {
		failed = true
		fmt.Printf("got %d * 73 == %d, wanted %d\n", s, got, want)
	}
	if want, got := r[74], s*74; want != got {
		failed = true
		fmt.Printf("got %d * 74 == %d, wanted %d\n", s, got, want)
	}
	if want, got := r[75], s*75; want != got {
		failed = true
		fmt.Printf("got %d * 75 == %d, wanted %d\n", s, got, want)
	}
	if want, got := r[76], s*76; want != got {
		failed = true
		fmt.Printf("got %d * 76 == %d, wanted %d\n", s, got, want)
	}
	if want, got := r[77], s*77; want != got {
		failed = true
		fmt.Printf("got %d * 77 == %d, wanted %d\n", s, got, want)
	}
	if want, got := r[78], s*78; want != got {
		failed = true
		fmt.Printf("got %d * 78 == %d, wanted %d\n", s, got, want)
	}
	if want, got := r[79], s*79; want != got {
		failed = true
		fmt.Printf("got %d * 79 == %d, wanted %d\n", s, got, want)
	}
	if want, got := r[80], s*80; want != got {
		failed = true
		fmt.Printf("got %d * 80 == %d, wanted %d\n", s, got, want)
	}
	if want, got := r[81], s*81; want != got {
		failed = true
		fmt.Printf("got %d * 81 == %d, wanted %d\n", s, got, want)
	}
	if want, got := r[82], s*82; want != got {
		failed = true
		fmt.Printf("got %d * 82 == %d, wanted %d\n", s, got, want)
	}
	if want, got := r[83], s*83; want != got {
		failed = true
		fmt.Printf("got %d * 83 == %d, wanted %d\n", s, got, want)
	}
	if want, got := r[84], s*84; want != got {
		failed = true
		fmt.Printf("got %d * 84 == %d, wanted %d\n", s, got, want)
	}
	if want, got := r[85], s*85; want != got {
		failed = true
		fmt.Printf("got %d * 85 == %d, wanted %d\n", s, got, want)
	}
	if want, got := r[86], s*86; want != got {
		failed = true
		fmt.Printf("got %d * 86 == %d, wanted %d\n", s, got, want)
	}
	if want, got := r[87], s*87; want != got {
		failed = true
		fmt.Printf("got %d * 87 == %d, wanted %d\n", s, got, want)
	}
	if want, got := r[88], s*88; want != got {
		failed = true
		fmt.Printf("got %d * 88 == %d, wanted %d\n", s, got, want)
	}
	if want, got := r[89], s*89; want != got {
		failed = true
		fmt.Printf("got %d * 89 == %d, wanted %d\n", s, got, want)
	}
	if want, got := r[90], s*90; want != got {
		failed = true
		fmt.Printf("got %d * 90 == %d, wanted %d\n", s, got, want)
	}
	if want, got := r[91], s*91; want != got {
		failed = true
		fmt.Printf("got %d * 91 == %d, wanted %d\n", s, got, want)
	}
	if want, got := r[92], s*92; want != got {
		failed = true
		fmt.Printf("got %d * 92 == %d, wanted %d\n", s, got, want)
	}
	if want, got := r[93], s*93; want != got {
		failed = true
		fmt.Printf("got %d * 93 == %d, wanted %d\n", s, got, want)
	}
	if want, got := r[94], s*94; want != got {
		failed = true
		fmt.Printf("got %d * 94 == %d, wanted %d\n", s, got, want)
	}
	if want, got := r[95], s*95; want != got {
		failed = true
		fmt.Printf("got %d * 95 == %d, wanted %d\n", s, got, want)
	}
	if want, got := r[96], s*96; want != got {
		failed = true
		fmt.Printf("got %d * 96 == %d, wanted %d\n", s, got, want)
	}
	if want, got := r[97], s*97; want != got {
		failed = true
		fmt.Printf("got %d * 97 == %d, wanted %d\n", s, got, want)
	}
	if want, got := r[98], s*98; want != got {
		failed = true
		fmt.Printf("got %d * 98 == %d, wanted %d\n", s, got, want)
	}
	if want, got := r[99], s*99; want != got {
		failed = true
		fmt.Printf("got %d * 99 == %d, wanted %d\n", s, got, want)
	}
}

func main() {
	r := [100]int64{0, 15, 30, 45, 60, 75, 90, 105, 120, 135, 150, 165, 180, 195, 210, 225, 240, 255, 270, 285, 300, 315, 330, 345, 360, 375, 390, 405, 420, 435, 450, 465, 480, 495, 510, 525, 540, 555, 570, 585, 600, 615, 630, 645, 660, 675, 690, 705, 720, 735, 750, 765, 780, 795, 810, 825, 840, 855, 870, 885, 900, 915, 930, 945, 960, 975, 990, 1005, 1020, 1035, 1050, 1065, 1080, 1095, 1110, 1125, 1140, 1155, 1170, 1185, 1200, 1215, 1230, 1245, 1260, 1275, 1290, 1305, 1320, 1335, 1350, 1365, 1380, 1395, 1410, 1425, 1440, 1455, 1470, 1485}
	testMultiplication(r, 15)
	if failed {
		panic("failed multiplication")
	}
}
