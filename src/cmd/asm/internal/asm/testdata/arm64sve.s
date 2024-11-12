// Copyright 2024 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

TEXT svetest(SB),$0

// ABS     <Zd>.<T>, <Pg>/M, <Zn>.<T>
    ZABS P0.M, Z0.B, Z0.B      // 00a01604
    ZABS P3.M, Z12.B, Z10.B    // 8aad1604
    ZABS P7.M, Z31.B, Z31.B    // ffbf1604
    ZABS P0.M, Z0.H, Z0.H      // 00a05604
    ZABS P3.M, Z12.H, Z10.H    // 8aad5604
    ZABS P7.M, Z31.H, Z31.H    // ffbf5604
    ZABS P0.M, Z0.S, Z0.S      // 00a09604
    ZABS P3.M, Z12.S, Z10.S    // 8aad9604
    ZABS P7.M, Z31.S, Z31.S    // ffbf9604
    ZABS P0.M, Z0.D, Z0.D      // 00a0d604
    ZABS P3.M, Z12.D, Z10.D    // 8aadd604
    ZABS P7.M, Z31.D, Z31.D    // ffbfd604

// ADD     <Zdn>.<T>, <Pg>/M, <Zdn>.<T>, <Zm>.<T>
    ZADD P0.M, Z0.B, Z0.B, Z0.B       // 00000004
    ZADD P3.M, Z10.B, Z12.B, Z10.B    // 8a0d0004
    ZADD P7.M, Z31.B, Z31.B, Z31.B    // ff1f0004
    ZADD P0.M, Z0.H, Z0.H, Z0.H       // 00004004
    ZADD P3.M, Z10.H, Z12.H, Z10.H    // 8a0d4004
    ZADD P7.M, Z31.H, Z31.H, Z31.H    // ff1f4004
    ZADD P0.M, Z0.S, Z0.S, Z0.S       // 00008004
    ZADD P3.M, Z10.S, Z12.S, Z10.S    // 8a0d8004
    ZADD P7.M, Z31.S, Z31.S, Z31.S    // ff1f8004
    ZADD P0.M, Z0.D, Z0.D, Z0.D       // 0000c004
    ZADD P3.M, Z10.D, Z12.D, Z10.D    // 8a0dc004
    ZADD P7.M, Z31.D, Z31.D, Z31.D    // ff1fc004

// ADD     <Zdn>.<T>, <Zdn>.<T>, #<imm>, <shift>
    ZADD Z0.B, $0, $0, Z0.B        // 00c02025
    ZADD Z10.B, $85, $0, Z10.B     // aaca2025
    ZADD Z31.B, $255, $0, Z31.B    // ffdf2025
    ZADD Z0.H, $0, $8, Z0.H        // 00e06025
    ZADD Z10.H, $85, $8, Z10.H     // aaea6025
    ZADD Z31.H, $255, $0, Z31.H    // ffdf6025
    ZADD Z0.S, $0, $8, Z0.S        // 00e0a025
    ZADD Z10.S, $85, $8, Z10.S     // aaeaa025
    ZADD Z31.S, $255, $0, Z31.S    // ffdfa025
    ZADD Z0.D, $0, $8, Z0.D        // 00e0e025
    ZADD Z10.D, $85, $8, Z10.D     // aaeae025
    ZADD Z31.D, $255, $0, Z31.D    // ffdfe025

// ADD     <Zd>.<T>, <Zn>.<T>, <Zm>.<T>
    ZADD Z0.B, Z0.B, Z0.B       // 00002004
    ZADD Z11.B, Z12.B, Z10.B    // 6a012c04
    ZADD Z31.B, Z31.B, Z31.B    // ff033f04
    ZADD Z0.H, Z0.H, Z0.H       // 00006004
    ZADD Z11.H, Z12.H, Z10.H    // 6a016c04
    ZADD Z31.H, Z31.H, Z31.H    // ff037f04
    ZADD Z0.S, Z0.S, Z0.S       // 0000a004
    ZADD Z11.S, Z12.S, Z10.S    // 6a01ac04
    ZADD Z31.S, Z31.S, Z31.S    // ff03bf04
    ZADD Z0.D, Z0.D, Z0.D       // 0000e004
    ZADD Z11.D, Z12.D, Z10.D    // 6a01ec04
    ZADD Z31.D, Z31.D, Z31.D    // ff03ff04

// ADDPL   <Xd|SP>, <Xn|SP>, #<imm>
    ZADDPL R0, $-32, R0      // 00546004
    ZADDPL R11, $-11, R10    // aa566b04
    ZADDPL R30, $31, R30     // fe537e04

// ADDVL   <Xd|SP>, <Xn|SP>, #<imm>
    ZADDVL R0, $-32, R0      // 00542004
    ZADDVL R11, $-11, R10    // aa562b04
    ZADDVL R30, $31, R30     // fe533e04

// ADR     <Zd>.D, [<Zn>.D, <Zm>.D, SXTW <amount>]
    ZADR (Z0.D)(Z0.D.SXTW), Z0.D          // 00a02004
    ZADR (Z11.D)(Z12.D.SXTW), Z10.D       // 6aa12c04
    ZADR (Z31.D)(Z31.D.SXTW), Z31.D       // ffa33f04
    ZADR (Z0.D)(Z0.D.SXTW<<1), Z0.D       // 00a42004
    ZADR (Z11.D)(Z12.D.SXTW<<1), Z10.D    // 6aa52c04
    ZADR (Z31.D)(Z31.D.SXTW<<1), Z31.D    // ffa73f04
    ZADR (Z0.D)(Z0.D.SXTW<<2), Z0.D       // 00a82004
    ZADR (Z11.D)(Z12.D.SXTW<<2), Z10.D    // 6aa92c04
    ZADR (Z31.D)(Z31.D.SXTW<<2), Z31.D    // ffab3f04
    ZADR (Z0.D)(Z0.D.SXTW<<3), Z0.D       // 00ac2004
    ZADR (Z11.D)(Z12.D.SXTW<<3), Z10.D    // 6aad2c04
    ZADR (Z31.D)(Z31.D.SXTW<<3), Z31.D    // ffaf3f04

// ADR     <Zd>.D, [<Zn>.D, <Zm>.D, UXTW <amount>]
    ZADR (Z0.D)(Z0.D.UXTW), Z0.D          // 00a06004
    ZADR (Z11.D)(Z12.D.UXTW), Z10.D       // 6aa16c04
    ZADR (Z31.D)(Z31.D.UXTW), Z31.D       // ffa37f04
    ZADR (Z0.D)(Z0.D.UXTW<<1), Z0.D       // 00a46004
    ZADR (Z11.D)(Z12.D.UXTW<<1), Z10.D    // 6aa56c04
    ZADR (Z31.D)(Z31.D.UXTW<<1), Z31.D    // ffa77f04
    ZADR (Z0.D)(Z0.D.UXTW<<2), Z0.D       // 00a86004
    ZADR (Z11.D)(Z12.D.UXTW<<2), Z10.D    // 6aa96c04
    ZADR (Z31.D)(Z31.D.UXTW<<2), Z31.D    // ffab7f04
    ZADR (Z0.D)(Z0.D.UXTW<<3), Z0.D       // 00ac6004
    ZADR (Z11.D)(Z12.D.UXTW<<3), Z10.D    // 6aad6c04
    ZADR (Z31.D)(Z31.D.UXTW<<3), Z31.D    // ffaf7f04

// ADR     <Zd>.<T>, [<Zn>.<T>, <Zm>.<T>, <extend> <amount>]
    ZADR (Z0.S)(Z0.S), Z0.S              // 00a0a004
    ZADR (Z11.S)(Z12.S), Z10.S           // 6aa1ac04
    ZADR (Z31.S)(Z31.S), Z31.S           // ffa3bf04
    ZADR (Z0.S)(Z0.S.LSL<<1), Z0.S       // 00a4a004
    ZADR (Z11.S)(Z12.S.LSL<<1), Z10.S    // 6aa5ac04
    ZADR (Z31.S)(Z31.S.LSL<<1), Z31.S    // ffa7bf04
    ZADR (Z0.S)(Z0.S.LSL<<2), Z0.S       // 00a8a004
    ZADR (Z11.S)(Z12.S.LSL<<2), Z10.S    // 6aa9ac04
    ZADR (Z31.S)(Z31.S.LSL<<2), Z31.S    // ffabbf04
    ZADR (Z0.S)(Z0.S.LSL<<3), Z0.S       // 00aca004
    ZADR (Z11.S)(Z12.S.LSL<<3), Z10.S    // 6aadac04
    ZADR (Z31.S)(Z31.S.LSL<<3), Z31.S    // ffafbf04
    ZADR (Z0.D)(Z0.D), Z0.D              // 00a0e004
    ZADR (Z11.D)(Z12.D), Z10.D           // 6aa1ec04
    ZADR (Z31.D)(Z31.D), Z31.D           // ffa3ff04
    ZADR (Z0.D)(Z0.D.LSL<<1), Z0.D       // 00a4e004
    ZADR (Z11.D)(Z12.D.LSL<<1), Z10.D    // 6aa5ec04
    ZADR (Z31.D)(Z31.D.LSL<<1), Z31.D    // ffa7ff04
    ZADR (Z0.D)(Z0.D.LSL<<2), Z0.D       // 00a8e004
    ZADR (Z11.D)(Z12.D.LSL<<2), Z10.D    // 6aa9ec04
    ZADR (Z31.D)(Z31.D.LSL<<2), Z31.D    // ffabff04
    ZADR (Z0.D)(Z0.D.LSL<<3), Z0.D       // 00ace004
    ZADR (Z11.D)(Z12.D.LSL<<3), Z10.D    // 6aadec04
    ZADR (Z31.D)(Z31.D.LSL<<3), Z31.D    // ffafff04

// AND     <Pd>.B, <Pg>/Z, <Pn>.B, <Pm>.B
    PAND P0.Z, P0.B, P0.B, P0.B        // 00400025
    PAND P6.Z, P7.B, P8.B, P5.B        // e5580825
    PAND P15.Z, P15.B, P15.B, P15.B    // ef7d0f25

// AND     <Zdn>.<T>, <Pg>/M, <Zdn>.<T>, <Zm>.<T>
    ZAND P0.M, Z0.B, Z0.B, Z0.B       // 00001a04
    ZAND P3.M, Z10.B, Z12.B, Z10.B    // 8a0d1a04
    ZAND P7.M, Z31.B, Z31.B, Z31.B    // ff1f1a04
    ZAND P0.M, Z0.H, Z0.H, Z0.H       // 00005a04
    ZAND P3.M, Z10.H, Z12.H, Z10.H    // 8a0d5a04
    ZAND P7.M, Z31.H, Z31.H, Z31.H    // ff1f5a04
    ZAND P0.M, Z0.S, Z0.S, Z0.S       // 00009a04
    ZAND P3.M, Z10.S, Z12.S, Z10.S    // 8a0d9a04
    ZAND P7.M, Z31.S, Z31.S, Z31.S    // ff1f9a04
    ZAND P0.M, Z0.D, Z0.D, Z0.D       // 0000da04
    ZAND P3.M, Z10.D, Z12.D, Z10.D    // 8a0dda04
    ZAND P7.M, Z31.D, Z31.D, Z31.D    // ff1fda04

// AND     <Zdn>.D, <Zdn>.D, #<const>
    ZAND Z0.S, $1, Z0.S                   // 00008005
    ZAND Z10.S, $4192256, Z10.S           // 4aa98005
    ZAND Z0.H, $1, Z0.H                   // 00048005
    ZAND Z10.H, $63489, Z10.H             // aa2c8005
    ZAND Z0.B, $1, Z0.B                   // 00068005
    ZAND Z10.B, $56, Z10.B                // 4a2e8005
    ZAND Z0.B, $17, Z0.B                  // 00078005
    ZAND Z10.B, $153, Z10.B               // 2a0f8005
    ZAND Z0.D, $4294967297, Z0.D          // 00008005
    ZAND Z10.D, $-8787503089663, Z10.D    // aaaa8005

// AND     <Zd>.D, <Zn>.D, <Zm>.D
    ZAND Z0.D, Z0.D, Z0.D       // 00302004
    ZAND Z11.D, Z12.D, Z10.D    // 6a312c04
    ZAND Z31.D, Z31.D, Z31.D    // ff333f04

// ANDS    <Pd>.B, <Pg>/Z, <Pn>.B, <Pm>.B
    PANDS P0.Z, P0.B, P0.B, P0.B        // 00404025
    PANDS P6.Z, P7.B, P8.B, P5.B        // e5584825
    PANDS P15.Z, P15.B, P15.B, P15.B    // ef7d4f25

// ANDV    <V><d>, <Pg>, <Zn>.<T>
    ZANDV P0, Z0.B, V0      // 00201a04
    ZANDV P3, Z12.B, V10    // 8a2d1a04
    ZANDV P7, Z31.B, V31    // ff3f1a04
    ZANDV P0, Z0.H, V0      // 00205a04
    ZANDV P3, Z12.H, V10    // 8a2d5a04
    ZANDV P7, Z31.H, V31    // ff3f5a04
    ZANDV P0, Z0.S, V0      // 00209a04
    ZANDV P3, Z12.S, V10    // 8a2d9a04
    ZANDV P7, Z31.S, V31    // ff3f9a04
    ZANDV P0, Z0.D, V0      // 0020da04
    ZANDV P3, Z12.D, V10    // 8a2dda04
    ZANDV P7, Z31.D, V31    // ff3fda04

// ASR     <Zdn>.<T>, <Pg>/M, <Zdn>.<T>, #<const>
    ZASR P0.M, Z0.B, $1, Z0.B       // e0810004
    ZASR P3.M, Z10.B, $3, Z10.B     // aa8d0004
    ZASR P7.M, Z31.B, $8, Z31.B     // 1f9d0004
    ZASR P0.M, Z0.H, $1, Z0.H       // e0830004
    ZASR P3.M, Z10.H, $6, Z10.H     // 4a8f0004
    ZASR P7.M, Z31.H, $16, Z31.H    // 1f9e0004
    ZASR P0.M, Z0.S, $1, Z0.S       // e0834004
    ZASR P3.M, Z10.S, $11, Z10.S    // aa8e4004
    ZASR P7.M, Z31.S, $32, Z31.S    // 1f9c4004
    ZASR P0.M, Z0.D, $1, Z0.D       // e083c004
    ZASR P3.M, Z10.D, $22, Z10.D    // 4a8dc004
    ZASR P7.M, Z31.D, $64, Z31.D    // 1f9c8004

// ASR     <Zdn>.<T>, <Pg>/M, <Zdn>.<T>, <Zm>.D
    ZASR P0.M, Z0.B, Z0.D, Z0.B       // 00801804
    ZASR P3.M, Z10.B, Z12.D, Z10.B    // 8a8d1804
    ZASR P7.M, Z31.B, Z31.D, Z31.B    // ff9f1804
    ZASR P0.M, Z0.H, Z0.D, Z0.H       // 00805804
    ZASR P3.M, Z10.H, Z12.D, Z10.H    // 8a8d5804
    ZASR P7.M, Z31.H, Z31.D, Z31.H    // ff9f5804
    ZASR P0.M, Z0.S, Z0.D, Z0.S       // 00809804
    ZASR P3.M, Z10.S, Z12.D, Z10.S    // 8a8d9804
    ZASR P7.M, Z31.S, Z31.D, Z31.S    // ff9f9804

// ASR     <Zdn>.<T>, <Pg>/M, <Zdn>.<T>, <Zm>.<T>
    ZASR P0.M, Z0.B, Z0.B, Z0.B       // 00801004
    ZASR P3.M, Z10.B, Z12.B, Z10.B    // 8a8d1004
    ZASR P7.M, Z31.B, Z31.B, Z31.B    // ff9f1004
    ZASR P0.M, Z0.H, Z0.H, Z0.H       // 00805004
    ZASR P3.M, Z10.H, Z12.H, Z10.H    // 8a8d5004
    ZASR P7.M, Z31.H, Z31.H, Z31.H    // ff9f5004
    ZASR P0.M, Z0.S, Z0.S, Z0.S       // 00809004
    ZASR P3.M, Z10.S, Z12.S, Z10.S    // 8a8d9004
    ZASR P7.M, Z31.S, Z31.S, Z31.S    // ff9f9004
    ZASR P0.M, Z0.D, Z0.D, Z0.D       // 0080d004
    ZASR P3.M, Z10.D, Z12.D, Z10.D    // 8a8dd004
    ZASR P7.M, Z31.D, Z31.D, Z31.D    // ff9fd004

// ASR     <Zd>.<T>, <Zn>.<T>, #<const>
    ZASR Z0.B, $1, Z0.B       // 00902f04
    ZASR Z11.B, $3, Z10.B     // 6a912d04
    ZASR Z31.B, $8, Z31.B     // ff932804
    ZASR Z0.H, $1, Z0.H       // 00903f04
    ZASR Z11.H, $6, Z10.H     // 6a913a04
    ZASR Z31.H, $16, Z31.H    // ff933004
    ZASR Z0.S, $1, Z0.S       // 00907f04
    ZASR Z11.S, $11, Z10.S    // 6a917504
    ZASR Z31.S, $32, Z31.S    // ff936004
    ZASR Z0.D, $1, Z0.D       // 0090ff04
    ZASR Z11.D, $22, Z10.D    // 6a91ea04
    ZASR Z31.D, $64, Z31.D    // ff93a004

// ASR     <Zd>.<T>, <Zn>.<T>, <Zm>.D
    ZASR Z0.B, Z0.D, Z0.B       // 00802004
    ZASR Z11.B, Z12.D, Z10.B    // 6a812c04
    ZASR Z31.B, Z31.D, Z31.B    // ff833f04
    ZASR Z0.H, Z0.D, Z0.H       // 00806004
    ZASR Z11.H, Z12.D, Z10.H    // 6a816c04
    ZASR Z31.H, Z31.D, Z31.H    // ff837f04
    ZASR Z0.S, Z0.D, Z0.S       // 0080a004
    ZASR Z11.S, Z12.D, Z10.S    // 6a81ac04
    ZASR Z31.S, Z31.D, Z31.S    // ff83bf04

// ASRD    <Zdn>.<T>, <Pg>/M, <Zdn>.<T>, #<const>
    ZASRD P0.M, Z0.B, $1, Z0.B       // e0810404
    ZASRD P3.M, Z10.B, $3, Z10.B     // aa8d0404
    ZASRD P7.M, Z31.B, $8, Z31.B     // 1f9d0404
    ZASRD P0.M, Z0.H, $1, Z0.H       // e0830404
    ZASRD P3.M, Z10.H, $6, Z10.H     // 4a8f0404
    ZASRD P7.M, Z31.H, $16, Z31.H    // 1f9e0404
    ZASRD P0.M, Z0.S, $1, Z0.S       // e0834404
    ZASRD P3.M, Z10.S, $11, Z10.S    // aa8e4404
    ZASRD P7.M, Z31.S, $32, Z31.S    // 1f9c4404
    ZASRD P0.M, Z0.D, $1, Z0.D       // e083c404
    ZASRD P3.M, Z10.D, $22, Z10.D    // 4a8dc404
    ZASRD P7.M, Z31.D, $64, Z31.D    // 1f9c8404

// ASRR    <Zdn>.<T>, <Pg>/M, <Zdn>.<T>, <Zm>.<T>
    ZASRR P0.M, Z0.B, Z0.B, Z0.B       // 00801404
    ZASRR P3.M, Z10.B, Z12.B, Z10.B    // 8a8d1404
    ZASRR P7.M, Z31.B, Z31.B, Z31.B    // ff9f1404
    ZASRR P0.M, Z0.H, Z0.H, Z0.H       // 00805404
    ZASRR P3.M, Z10.H, Z12.H, Z10.H    // 8a8d5404
    ZASRR P7.M, Z31.H, Z31.H, Z31.H    // ff9f5404
    ZASRR P0.M, Z0.S, Z0.S, Z0.S       // 00809404
    ZASRR P3.M, Z10.S, Z12.S, Z10.S    // 8a8d9404
    ZASRR P7.M, Z31.S, Z31.S, Z31.S    // ff9f9404
    ZASRR P0.M, Z0.D, Z0.D, Z0.D       // 0080d404
    ZASRR P3.M, Z10.D, Z12.D, Z10.D    // 8a8dd404
    ZASRR P7.M, Z31.D, Z31.D, Z31.D    // ff9fd404

// BFCVT   <Zd>.H, <Pg>/M, <Zn>.S
    ZBFCVT P0.M, Z0.S, Z0.H      // 00a08a65
    ZBFCVT P3.M, Z12.S, Z10.H    // 8aad8a65
    ZBFCVT P7.M, Z31.S, Z31.H    // ffbf8a65

// BFCVTNT <Zd>.H, <Pg>/M, <Zn>.S
    ZBFCVTNT P0.M, Z0.S, Z0.H      // 00a08a64
    ZBFCVTNT P3.M, Z12.S, Z10.H    // 8aad8a64
    ZBFCVTNT P7.M, Z31.S, Z31.H    // ffbf8a64

// BFDOT   <Zda>.S, <Zn>.H, <Zm>.H
    ZBFDOT Z0.H, Z0.H, Z0.S       // 00806064
    ZBFDOT Z11.H, Z12.H, Z10.S    // 6a816c64
    ZBFDOT Z31.H, Z31.H, Z31.S    // ff837f64

// BFDOT   <Zda>.S, <Zn>.H, <Zm>.H[<imm>]
    ZBFDOT Z0.H, Z0.H[0], Z0.S      // 00406064
    ZBFDOT Z11.H, Z4.H[1], Z10.S    // 6a416c64
    ZBFDOT Z31.H, Z7.H[3], Z31.S    // ff437f64

// BFMLALB <Zda>.S, <Zn>.H, <Zm>.H
    ZBFMLALB Z0.H, Z0.H, Z0.S       // 0080e064
    ZBFMLALB Z11.H, Z12.H, Z10.S    // 6a81ec64
    ZBFMLALB Z31.H, Z31.H, Z31.S    // ff83ff64

// BFMLALB <Zda>.S, <Zn>.H, <Zm>.H[<imm>]
    ZBFMLALB Z0.H, Z0.H[0], Z0.S      // 0040e064
    ZBFMLALB Z11.H, Z4.H[2], Z10.S    // 6a41ec64
    ZBFMLALB Z31.H, Z7.H[7], Z31.S    // ff4bff64

// BFMLALT <Zda>.S, <Zn>.H, <Zm>.H
    ZBFMLALT Z0.H, Z0.H, Z0.S       // 0084e064
    ZBFMLALT Z11.H, Z12.H, Z10.S    // 6a85ec64
    ZBFMLALT Z31.H, Z31.H, Z31.S    // ff87ff64

// BFMLALT <Zda>.S, <Zn>.H, <Zm>.H[<imm>]
    ZBFMLALT Z0.H, Z0.H[0], Z0.S      // 0044e064
    ZBFMLALT Z11.H, Z4.H[2], Z10.S    // 6a45ec64
    ZBFMLALT Z31.H, Z7.H[7], Z31.S    // ff4fff64

// BFMMLA  <Zda>.S, <Zn>.H, <Zm>.H
    ZBFMMLA Z0.H, Z0.H, Z0.S       // 00e46064
    ZBFMMLA Z11.H, Z12.H, Z10.S    // 6ae56c64
    ZBFMMLA Z31.H, Z31.H, Z31.S    // ffe77f64

// BIC     <Pd>.B, <Pg>/Z, <Pn>.B, <Pm>.B
    PBIC P0.Z, P0.B, P0.B, P0.B        // 10400025
    PBIC P6.Z, P7.B, P8.B, P5.B        // f5580825
    PBIC P15.Z, P15.B, P15.B, P15.B    // ff7d0f25

// BIC     <Zdn>.<T>, <Pg>/M, <Zdn>.<T>, <Zm>.<T>
    ZBIC P0.M, Z0.B, Z0.B, Z0.B       // 00001b04
    ZBIC P3.M, Z10.B, Z12.B, Z10.B    // 8a0d1b04
    ZBIC P7.M, Z31.B, Z31.B, Z31.B    // ff1f1b04
    ZBIC P0.M, Z0.H, Z0.H, Z0.H       // 00005b04
    ZBIC P3.M, Z10.H, Z12.H, Z10.H    // 8a0d5b04
    ZBIC P7.M, Z31.H, Z31.H, Z31.H    // ff1f5b04
    ZBIC P0.M, Z0.S, Z0.S, Z0.S       // 00009b04
    ZBIC P3.M, Z10.S, Z12.S, Z10.S    // 8a0d9b04
    ZBIC P7.M, Z31.S, Z31.S, Z31.S    // ff1f9b04
    ZBIC P0.M, Z0.D, Z0.D, Z0.D       // 0000db04
    ZBIC P3.M, Z10.D, Z12.D, Z10.D    // 8a0ddb04
    ZBIC P7.M, Z31.D, Z31.D, Z31.D    // ff1fdb04

// BIC     <Zd>.D, <Zn>.D, <Zm>.D
    ZBIC Z0.D, Z0.D, Z0.D       // 0030e004
    ZBIC Z11.D, Z12.D, Z10.D    // 6a31ec04
    ZBIC Z31.D, Z31.D, Z31.D    // ff33ff04

// BICS    <Pd>.B, <Pg>/Z, <Pn>.B, <Pm>.B
    PBICS P0.Z, P0.B, P0.B, P0.B        // 10404025
    PBICS P6.Z, P7.B, P8.B, P5.B        // f5584825
    PBICS P15.Z, P15.B, P15.B, P15.B    // ff7d4f25

// BRKA    <Pd>.B, <Pg>/<ZM>, <Pn>.B
    PBRKA P0.Z, P0.B, P0.B       // 00401025
    PBRKA P6.Z, P7.B, P5.B       // e5581025
    PBRKA P15.Z, P15.B, P15.B    // ef7d1025
    PBRKA P0.M, P0.B, P0.B       // 10401025
    PBRKA P6.M, P7.B, P5.B       // f5581025
    PBRKA P15.M, P15.B, P15.B    // ff7d1025

// BRKAS   <Pd>.B, <Pg>/Z, <Pn>.B
    PBRKAS P0.Z, P0.B, P0.B       // 00405025
    PBRKAS P6.Z, P7.B, P5.B       // e5585025
    PBRKAS P15.Z, P15.B, P15.B    // ef7d5025

// BRKB    <Pd>.B, <Pg>/<ZM>, <Pn>.B
    PBRKB P0.Z, P0.B, P0.B       // 00409025
    PBRKB P6.Z, P7.B, P5.B       // e5589025
    PBRKB P15.Z, P15.B, P15.B    // ef7d9025
    PBRKB P0.M, P0.B, P0.B       // 10409025
    PBRKB P6.M, P7.B, P5.B       // f5589025
    PBRKB P15.M, P15.B, P15.B    // ff7d9025

// BRKBS   <Pd>.B, <Pg>/Z, <Pn>.B
    PBRKBS P0.Z, P0.B, P0.B       // 0040d025
    PBRKBS P6.Z, P7.B, P5.B       // e558d025
    PBRKBS P15.Z, P15.B, P15.B    // ef7dd025

// BRKN    <Pdm>.B, <Pg>/Z, <Pn>.B, <Pdm>.B
    PBRKN P0.Z, P0.B, P0.B, P0.B        // 00401825
    PBRKN P6.Z, P7.B, P5.B, P5.B        // e5581825
    PBRKN P15.Z, P15.B, P15.B, P15.B    // ef7d1825

// BRKNS   <Pdm>.B, <Pg>/Z, <Pn>.B, <Pdm>.B
    PBRKNS P0.Z, P0.B, P0.B, P0.B        // 00405825
    PBRKNS P6.Z, P7.B, P5.B, P5.B        // e5585825
    PBRKNS P15.Z, P15.B, P15.B, P15.B    // ef7d5825

// BRKPA   <Pd>.B, <Pg>/Z, <Pn>.B, <Pm>.B
    PBRKPA P0.Z, P0.B, P0.B, P0.B        // 00c00025
    PBRKPA P6.Z, P7.B, P8.B, P5.B        // e5d80825
    PBRKPA P15.Z, P15.B, P15.B, P15.B    // effd0f25

// BRKPAS  <Pd>.B, <Pg>/Z, <Pn>.B, <Pm>.B
    PBRKPAS P0.Z, P0.B, P0.B, P0.B        // 00c04025
    PBRKPAS P6.Z, P7.B, P8.B, P5.B        // e5d84825
    PBRKPAS P15.Z, P15.B, P15.B, P15.B    // effd4f25

// BRKPB   <Pd>.B, <Pg>/Z, <Pn>.B, <Pm>.B
    PBRKPB P0.Z, P0.B, P0.B, P0.B        // 10c00025
    PBRKPB P6.Z, P7.B, P8.B, P5.B        // f5d80825
    PBRKPB P15.Z, P15.B, P15.B, P15.B    // fffd0f25

// BRKPBS  <Pd>.B, <Pg>/Z, <Pn>.B, <Pm>.B
    PBRKPBS P0.Z, P0.B, P0.B, P0.B        // 10c04025
    PBRKPBS P6.Z, P7.B, P8.B, P5.B        // f5d84825
    PBRKPBS P15.Z, P15.B, P15.B, P15.B    // fffd4f25

// CLASTA  <R><dn>, <Pg>, <R><dn>, <Zm>.<T>
    ZCLASTA P0, R0, Z0.B, R0       // 00a03005
    ZCLASTA P3, R10, Z12.B, R10    // 8aad3005
    ZCLASTA P7, R30, Z31.B, R30    // febf3005
    ZCLASTA P0, R0, Z0.H, R0       // 00a07005
    ZCLASTA P3, R10, Z12.H, R10    // 8aad7005
    ZCLASTA P7, R30, Z31.H, R30    // febf7005
    ZCLASTA P0, R0, Z0.S, R0       // 00a0b005
    ZCLASTA P3, R10, Z12.S, R10    // 8aadb005
    ZCLASTA P7, R30, Z31.S, R30    // febfb005
    ZCLASTA P0, R0, Z0.D, R0       // 00a0f005
    ZCLASTA P3, R10, Z12.D, R10    // 8aadf005
    ZCLASTA P7, R30, Z31.D, R30    // febff005

// CLASTA  <V><dn>, <Pg>, <V><dn>, <Zm>.<T>
    ZCLASTA P0, V0, Z0.B, V0       // 00802a05
    ZCLASTA P3, V10, Z12.B, V10    // 8a8d2a05
    ZCLASTA P7, V31, Z31.B, V31    // ff9f2a05
    ZCLASTA P0, V0, Z0.H, V0       // 00806a05
    ZCLASTA P3, V10, Z12.H, V10    // 8a8d6a05
    ZCLASTA P7, V31, Z31.H, V31    // ff9f6a05
    ZCLASTA P0, V0, Z0.S, V0       // 0080aa05
    ZCLASTA P3, V10, Z12.S, V10    // 8a8daa05
    ZCLASTA P7, V31, Z31.S, V31    // ff9faa05
    ZCLASTA P0, V0, Z0.D, V0       // 0080ea05
    ZCLASTA P3, V10, Z12.D, V10    // 8a8dea05
    ZCLASTA P7, V31, Z31.D, V31    // ff9fea05

// CLASTA  <Zdn>.<T>, <Pg>, <Zdn>.<T>, <Zm>.<T>
    ZCLASTA P0, Z0.B, Z0.B, Z0.B       // 00802805
    ZCLASTA P3, Z10.B, Z12.B, Z10.B    // 8a8d2805
    ZCLASTA P7, Z31.B, Z31.B, Z31.B    // ff9f2805
    ZCLASTA P0, Z0.H, Z0.H, Z0.H       // 00806805
    ZCLASTA P3, Z10.H, Z12.H, Z10.H    // 8a8d6805
    ZCLASTA P7, Z31.H, Z31.H, Z31.H    // ff9f6805
    ZCLASTA P0, Z0.S, Z0.S, Z0.S       // 0080a805
    ZCLASTA P3, Z10.S, Z12.S, Z10.S    // 8a8da805
    ZCLASTA P7, Z31.S, Z31.S, Z31.S    // ff9fa805
    ZCLASTA P0, Z0.D, Z0.D, Z0.D       // 0080e805
    ZCLASTA P3, Z10.D, Z12.D, Z10.D    // 8a8de805
    ZCLASTA P7, Z31.D, Z31.D, Z31.D    // ff9fe805

// CLASTB  <R><dn>, <Pg>, <R><dn>, <Zm>.<T>
    ZCLASTB P0, R0, Z0.B, R0       // 00a03105
    ZCLASTB P3, R10, Z12.B, R10    // 8aad3105
    ZCLASTB P7, R30, Z31.B, R30    // febf3105
    ZCLASTB P0, R0, Z0.H, R0       // 00a07105
    ZCLASTB P3, R10, Z12.H, R10    // 8aad7105
    ZCLASTB P7, R30, Z31.H, R30    // febf7105
    ZCLASTB P0, R0, Z0.S, R0       // 00a0b105
    ZCLASTB P3, R10, Z12.S, R10    // 8aadb105
    ZCLASTB P7, R30, Z31.S, R30    // febfb105
    ZCLASTB P0, R0, Z0.D, R0       // 00a0f105
    ZCLASTB P3, R10, Z12.D, R10    // 8aadf105
    ZCLASTB P7, R30, Z31.D, R30    // febff105

// CLASTB  <V><dn>, <Pg>, <V><dn>, <Zm>.<T>
    ZCLASTB P0, V0, Z0.B, V0       // 00802b05
    ZCLASTB P3, V10, Z12.B, V10    // 8a8d2b05
    ZCLASTB P7, V31, Z31.B, V31    // ff9f2b05
    ZCLASTB P0, V0, Z0.H, V0       // 00806b05
    ZCLASTB P3, V10, Z12.H, V10    // 8a8d6b05
    ZCLASTB P7, V31, Z31.H, V31    // ff9f6b05
    ZCLASTB P0, V0, Z0.S, V0       // 0080ab05
    ZCLASTB P3, V10, Z12.S, V10    // 8a8dab05
    ZCLASTB P7, V31, Z31.S, V31    // ff9fab05
    ZCLASTB P0, V0, Z0.D, V0       // 0080eb05
    ZCLASTB P3, V10, Z12.D, V10    // 8a8deb05
    ZCLASTB P7, V31, Z31.D, V31    // ff9feb05

// CLASTB  <Zdn>.<T>, <Pg>, <Zdn>.<T>, <Zm>.<T>
    ZCLASTB P0, Z0.B, Z0.B, Z0.B       // 00802905
    ZCLASTB P3, Z10.B, Z12.B, Z10.B    // 8a8d2905
    ZCLASTB P7, Z31.B, Z31.B, Z31.B    // ff9f2905
    ZCLASTB P0, Z0.H, Z0.H, Z0.H       // 00806905
    ZCLASTB P3, Z10.H, Z12.H, Z10.H    // 8a8d6905
    ZCLASTB P7, Z31.H, Z31.H, Z31.H    // ff9f6905
    ZCLASTB P0, Z0.S, Z0.S, Z0.S       // 0080a905
    ZCLASTB P3, Z10.S, Z12.S, Z10.S    // 8a8da905
    ZCLASTB P7, Z31.S, Z31.S, Z31.S    // ff9fa905
    ZCLASTB P0, Z0.D, Z0.D, Z0.D       // 0080e905
    ZCLASTB P3, Z10.D, Z12.D, Z10.D    // 8a8de905
    ZCLASTB P7, Z31.D, Z31.D, Z31.D    // ff9fe905

// CLS     <Zd>.<T>, <Pg>/M, <Zn>.<T>
    ZCLS P0.M, Z0.B, Z0.B      // 00a01804
    ZCLS P3.M, Z12.B, Z10.B    // 8aad1804
    ZCLS P7.M, Z31.B, Z31.B    // ffbf1804
    ZCLS P0.M, Z0.H, Z0.H      // 00a05804
    ZCLS P3.M, Z12.H, Z10.H    // 8aad5804
    ZCLS P7.M, Z31.H, Z31.H    // ffbf5804
    ZCLS P0.M, Z0.S, Z0.S      // 00a09804
    ZCLS P3.M, Z12.S, Z10.S    // 8aad9804
    ZCLS P7.M, Z31.S, Z31.S    // ffbf9804
    ZCLS P0.M, Z0.D, Z0.D      // 00a0d804
    ZCLS P3.M, Z12.D, Z10.D    // 8aadd804
    ZCLS P7.M, Z31.D, Z31.D    // ffbfd804

// CLZ     <Zd>.<T>, <Pg>/M, <Zn>.<T>
    ZCLZ P0.M, Z0.B, Z0.B      // 00a01904
    ZCLZ P3.M, Z12.B, Z10.B    // 8aad1904
    ZCLZ P7.M, Z31.B, Z31.B    // ffbf1904
    ZCLZ P0.M, Z0.H, Z0.H      // 00a05904
    ZCLZ P3.M, Z12.H, Z10.H    // 8aad5904
    ZCLZ P7.M, Z31.H, Z31.H    // ffbf5904
    ZCLZ P0.M, Z0.S, Z0.S      // 00a09904
    ZCLZ P3.M, Z12.S, Z10.S    // 8aad9904
    ZCLZ P7.M, Z31.S, Z31.S    // ffbf9904
    ZCLZ P0.M, Z0.D, Z0.D      // 00a0d904
    ZCLZ P3.M, Z12.D, Z10.D    // 8aadd904
    ZCLZ P7.M, Z31.D, Z31.D    // ffbfd904

// CMPEQ   <Pd>.<T>, <Pg>/Z, <Zn>.<T>, #<imm>
    ZCMPEQ P0.Z, Z0.B, $-16, P0.B     // 00801025
    ZCMPEQ P3.Z, Z12.B, $-6, P5.B     // 858d1a25
    ZCMPEQ P7.Z, Z31.B, $15, P15.B    // ef9f0f25
    ZCMPEQ P0.Z, Z0.H, $-16, P0.H     // 00805025
    ZCMPEQ P3.Z, Z12.H, $-6, P5.H     // 858d5a25
    ZCMPEQ P7.Z, Z31.H, $15, P15.H    // ef9f4f25
    ZCMPEQ P0.Z, Z0.S, $-16, P0.S     // 00809025
    ZCMPEQ P3.Z, Z12.S, $-6, P5.S     // 858d9a25
    ZCMPEQ P7.Z, Z31.S, $15, P15.S    // ef9f8f25
    ZCMPEQ P0.Z, Z0.D, $-16, P0.D     // 0080d025
    ZCMPEQ P3.Z, Z12.D, $-6, P5.D     // 858dda25
    ZCMPEQ P7.Z, Z31.D, $15, P15.D    // ef9fcf25

// CMPEQ   <Pd>.<T>, <Pg>/Z, <Zn>.<T>, <Zm>.D
    ZCMPEQ P0.Z, Z0.B, Z0.D, P0.B       // 00200024
    ZCMPEQ P3.Z, Z12.B, Z13.D, P5.B     // 852d0d24
    ZCMPEQ P7.Z, Z31.B, Z31.D, P15.B    // ef3f1f24
    ZCMPEQ P0.Z, Z0.H, Z0.D, P0.H       // 00204024
    ZCMPEQ P3.Z, Z12.H, Z13.D, P5.H     // 852d4d24
    ZCMPEQ P7.Z, Z31.H, Z31.D, P15.H    // ef3f5f24
    ZCMPEQ P0.Z, Z0.S, Z0.D, P0.S       // 00208024
    ZCMPEQ P3.Z, Z12.S, Z13.D, P5.S     // 852d8d24
    ZCMPEQ P7.Z, Z31.S, Z31.D, P15.S    // ef3f9f24

// CMPEQ   <Pd>.<T>, <Pg>/Z, <Zn>.<T>, <Zm>.<T>
    ZCMPEQ P0.Z, Z0.B, Z0.B, P0.B       // 00a00024
    ZCMPEQ P3.Z, Z12.B, Z13.B, P5.B     // 85ad0d24
    ZCMPEQ P7.Z, Z31.B, Z31.B, P15.B    // efbf1f24
    ZCMPEQ P0.Z, Z0.H, Z0.H, P0.H       // 00a04024
    ZCMPEQ P3.Z, Z12.H, Z13.H, P5.H     // 85ad4d24
    ZCMPEQ P7.Z, Z31.H, Z31.H, P15.H    // efbf5f24
    ZCMPEQ P0.Z, Z0.S, Z0.S, P0.S       // 00a08024
    ZCMPEQ P3.Z, Z12.S, Z13.S, P5.S     // 85ad8d24
    ZCMPEQ P7.Z, Z31.S, Z31.S, P15.S    // efbf9f24
    ZCMPEQ P0.Z, Z0.D, Z0.D, P0.D       // 00a0c024
    ZCMPEQ P3.Z, Z12.D, Z13.D, P5.D     // 85adcd24
    ZCMPEQ P7.Z, Z31.D, Z31.D, P15.D    // efbfdf24

// CMPGE   <Pd>.<T>, <Pg>/Z, <Zn>.<T>, #<imm>
    ZCMPGE P0.Z, Z0.B, $-16, P0.B     // 00001025
    ZCMPGE P3.Z, Z12.B, $-6, P5.B     // 850d1a25
    ZCMPGE P7.Z, Z31.B, $15, P15.B    // ef1f0f25
    ZCMPGE P0.Z, Z0.H, $-16, P0.H     // 00005025
    ZCMPGE P3.Z, Z12.H, $-6, P5.H     // 850d5a25
    ZCMPGE P7.Z, Z31.H, $15, P15.H    // ef1f4f25
    ZCMPGE P0.Z, Z0.S, $-16, P0.S     // 00009025
    ZCMPGE P3.Z, Z12.S, $-6, P5.S     // 850d9a25
    ZCMPGE P7.Z, Z31.S, $15, P15.S    // ef1f8f25
    ZCMPGE P0.Z, Z0.D, $-16, P0.D     // 0000d025
    ZCMPGE P3.Z, Z12.D, $-6, P5.D     // 850dda25
    ZCMPGE P7.Z, Z31.D, $15, P15.D    // ef1fcf25

// CMPGE   <Pd>.<T>, <Pg>/Z, <Zn>.<T>, <Zm>.D
    ZCMPGE P0.Z, Z0.B, Z0.D, P0.B       // 00400024
    ZCMPGE P3.Z, Z12.B, Z13.D, P5.B     // 854d0d24
    ZCMPGE P7.Z, Z31.B, Z31.D, P15.B    // ef5f1f24
    ZCMPGE P0.Z, Z0.H, Z0.D, P0.H       // 00404024
    ZCMPGE P3.Z, Z12.H, Z13.D, P5.H     // 854d4d24
    ZCMPGE P7.Z, Z31.H, Z31.D, P15.H    // ef5f5f24
    ZCMPGE P0.Z, Z0.S, Z0.D, P0.S       // 00408024
    ZCMPGE P3.Z, Z12.S, Z13.D, P5.S     // 854d8d24
    ZCMPGE P7.Z, Z31.S, Z31.D, P15.S    // ef5f9f24

// CMPGE   <Pd>.<T>, <Pg>/Z, <Zn>.<T>, <Zm>.<T>
    ZCMPGE P0.Z, Z0.B, Z0.B, P0.B       // 00800024
    ZCMPGE P3.Z, Z12.B, Z13.B, P5.B     // 858d0d24
    ZCMPGE P7.Z, Z31.B, Z31.B, P15.B    // ef9f1f24
    ZCMPGE P0.Z, Z0.H, Z0.H, P0.H       // 00804024
    ZCMPGE P3.Z, Z12.H, Z13.H, P5.H     // 858d4d24
    ZCMPGE P7.Z, Z31.H, Z31.H, P15.H    // ef9f5f24
    ZCMPGE P0.Z, Z0.S, Z0.S, P0.S       // 00808024
    ZCMPGE P3.Z, Z12.S, Z13.S, P5.S     // 858d8d24
    ZCMPGE P7.Z, Z31.S, Z31.S, P15.S    // ef9f9f24
    ZCMPGE P0.Z, Z0.D, Z0.D, P0.D       // 0080c024
    ZCMPGE P3.Z, Z12.D, Z13.D, P5.D     // 858dcd24
    ZCMPGE P7.Z, Z31.D, Z31.D, P15.D    // ef9fdf24

// CMPGT   <Pd>.<T>, <Pg>/Z, <Zn>.<T>, #<imm>
    ZCMPGT P0.Z, Z0.B, $-16, P0.B     // 10001025
    ZCMPGT P3.Z, Z12.B, $-6, P5.B     // 950d1a25
    ZCMPGT P7.Z, Z31.B, $15, P15.B    // ff1f0f25
    ZCMPGT P0.Z, Z0.H, $-16, P0.H     // 10005025
    ZCMPGT P3.Z, Z12.H, $-6, P5.H     // 950d5a25
    ZCMPGT P7.Z, Z31.H, $15, P15.H    // ff1f4f25
    ZCMPGT P0.Z, Z0.S, $-16, P0.S     // 10009025
    ZCMPGT P3.Z, Z12.S, $-6, P5.S     // 950d9a25
    ZCMPGT P7.Z, Z31.S, $15, P15.S    // ff1f8f25
    ZCMPGT P0.Z, Z0.D, $-16, P0.D     // 1000d025
    ZCMPGT P3.Z, Z12.D, $-6, P5.D     // 950dda25
    ZCMPGT P7.Z, Z31.D, $15, P15.D    // ff1fcf25

// CMPGT   <Pd>.<T>, <Pg>/Z, <Zn>.<T>, <Zm>.D
    ZCMPGT P0.Z, Z0.B, Z0.D, P0.B       // 10400024
    ZCMPGT P3.Z, Z12.B, Z13.D, P5.B     // 954d0d24
    ZCMPGT P7.Z, Z31.B, Z31.D, P15.B    // ff5f1f24
    ZCMPGT P0.Z, Z0.H, Z0.D, P0.H       // 10404024
    ZCMPGT P3.Z, Z12.H, Z13.D, P5.H     // 954d4d24
    ZCMPGT P7.Z, Z31.H, Z31.D, P15.H    // ff5f5f24
    ZCMPGT P0.Z, Z0.S, Z0.D, P0.S       // 10408024
    ZCMPGT P3.Z, Z12.S, Z13.D, P5.S     // 954d8d24
    ZCMPGT P7.Z, Z31.S, Z31.D, P15.S    // ff5f9f24

// CMPGT   <Pd>.<T>, <Pg>/Z, <Zn>.<T>, <Zm>.<T>
    ZCMPGT P0.Z, Z0.B, Z0.B, P0.B       // 10800024
    ZCMPGT P3.Z, Z12.B, Z13.B, P5.B     // 958d0d24
    ZCMPGT P7.Z, Z31.B, Z31.B, P15.B    // ff9f1f24
    ZCMPGT P0.Z, Z0.H, Z0.H, P0.H       // 10804024
    ZCMPGT P3.Z, Z12.H, Z13.H, P5.H     // 958d4d24
    ZCMPGT P7.Z, Z31.H, Z31.H, P15.H    // ff9f5f24
    ZCMPGT P0.Z, Z0.S, Z0.S, P0.S       // 10808024
    ZCMPGT P3.Z, Z12.S, Z13.S, P5.S     // 958d8d24
    ZCMPGT P7.Z, Z31.S, Z31.S, P15.S    // ff9f9f24
    ZCMPGT P0.Z, Z0.D, Z0.D, P0.D       // 1080c024
    ZCMPGT P3.Z, Z12.D, Z13.D, P5.D     // 958dcd24
    ZCMPGT P7.Z, Z31.D, Z31.D, P15.D    // ff9fdf24

// CMPHI   <Pd>.<T>, <Pg>/Z, <Zn>.<T>, #<imm>
    ZCMPHI P0.Z, Z0.B, $0, P0.B        // 10002024
    ZCMPHI P3.Z, Z12.B, $42, P5.B      // 958d2a24
    ZCMPHI P7.Z, Z31.B, $127, P15.B    // ffdf3f24
    ZCMPHI P0.Z, Z0.H, $0, P0.H        // 10006024
    ZCMPHI P3.Z, Z12.H, $42, P5.H      // 958d6a24
    ZCMPHI P7.Z, Z31.H, $127, P15.H    // ffdf7f24
    ZCMPHI P0.Z, Z0.S, $0, P0.S        // 1000a024
    ZCMPHI P3.Z, Z12.S, $42, P5.S      // 958daa24
    ZCMPHI P7.Z, Z31.S, $127, P15.S    // ffdfbf24
    ZCMPHI P0.Z, Z0.D, $0, P0.D        // 1000e024
    ZCMPHI P3.Z, Z12.D, $42, P5.D      // 958dea24
    ZCMPHI P7.Z, Z31.D, $127, P15.D    // ffdfff24

// CMPHI   <Pd>.<T>, <Pg>/Z, <Zn>.<T>, <Zm>.D
    ZCMPHI P0.Z, Z0.B, Z0.D, P0.B       // 10c00024
    ZCMPHI P3.Z, Z12.B, Z13.D, P5.B     // 95cd0d24
    ZCMPHI P7.Z, Z31.B, Z31.D, P15.B    // ffdf1f24
    ZCMPHI P0.Z, Z0.H, Z0.D, P0.H       // 10c04024
    ZCMPHI P3.Z, Z12.H, Z13.D, P5.H     // 95cd4d24
    ZCMPHI P7.Z, Z31.H, Z31.D, P15.H    // ffdf5f24
    ZCMPHI P0.Z, Z0.S, Z0.D, P0.S       // 10c08024
    ZCMPHI P3.Z, Z12.S, Z13.D, P5.S     // 95cd8d24
    ZCMPHI P7.Z, Z31.S, Z31.D, P15.S    // ffdf9f24

// CMPHI   <Pd>.<T>, <Pg>/Z, <Zn>.<T>, <Zm>.<T>
    ZCMPHI P0.Z, Z0.B, Z0.B, P0.B       // 10000024
    ZCMPHI P3.Z, Z12.B, Z13.B, P5.B     // 950d0d24
    ZCMPHI P7.Z, Z31.B, Z31.B, P15.B    // ff1f1f24
    ZCMPHI P0.Z, Z0.H, Z0.H, P0.H       // 10004024
    ZCMPHI P3.Z, Z12.H, Z13.H, P5.H     // 950d4d24
    ZCMPHI P7.Z, Z31.H, Z31.H, P15.H    // ff1f5f24
    ZCMPHI P0.Z, Z0.S, Z0.S, P0.S       // 10008024
    ZCMPHI P3.Z, Z12.S, Z13.S, P5.S     // 950d8d24
    ZCMPHI P7.Z, Z31.S, Z31.S, P15.S    // ff1f9f24
    ZCMPHI P0.Z, Z0.D, Z0.D, P0.D       // 1000c024
    ZCMPHI P3.Z, Z12.D, Z13.D, P5.D     // 950dcd24
    ZCMPHI P7.Z, Z31.D, Z31.D, P15.D    // ff1fdf24

// CMPHS   <Pd>.<T>, <Pg>/Z, <Zn>.<T>, #<imm>
    ZCMPHS P0.Z, Z0.B, $0, P0.B        // 00002024
    ZCMPHS P3.Z, Z12.B, $42, P5.B      // 858d2a24
    ZCMPHS P7.Z, Z31.B, $127, P15.B    // efdf3f24
    ZCMPHS P0.Z, Z0.H, $0, P0.H        // 00006024
    ZCMPHS P3.Z, Z12.H, $42, P5.H      // 858d6a24
    ZCMPHS P7.Z, Z31.H, $127, P15.H    // efdf7f24
    ZCMPHS P0.Z, Z0.S, $0, P0.S        // 0000a024
    ZCMPHS P3.Z, Z12.S, $42, P5.S      // 858daa24
    ZCMPHS P7.Z, Z31.S, $127, P15.S    // efdfbf24
    ZCMPHS P0.Z, Z0.D, $0, P0.D        // 0000e024
    ZCMPHS P3.Z, Z12.D, $42, P5.D      // 858dea24
    ZCMPHS P7.Z, Z31.D, $127, P15.D    // efdfff24

// CMPHS   <Pd>.<T>, <Pg>/Z, <Zn>.<T>, <Zm>.D
    ZCMPHS P0.Z, Z0.B, Z0.D, P0.B       // 00c00024
    ZCMPHS P3.Z, Z12.B, Z13.D, P5.B     // 85cd0d24
    ZCMPHS P7.Z, Z31.B, Z31.D, P15.B    // efdf1f24
    ZCMPHS P0.Z, Z0.H, Z0.D, P0.H       // 00c04024
    ZCMPHS P3.Z, Z12.H, Z13.D, P5.H     // 85cd4d24
    ZCMPHS P7.Z, Z31.H, Z31.D, P15.H    // efdf5f24
    ZCMPHS P0.Z, Z0.S, Z0.D, P0.S       // 00c08024
    ZCMPHS P3.Z, Z12.S, Z13.D, P5.S     // 85cd8d24
    ZCMPHS P7.Z, Z31.S, Z31.D, P15.S    // efdf9f24

// CMPHS   <Pd>.<T>, <Pg>/Z, <Zn>.<T>, <Zm>.<T>
    ZCMPHS P0.Z, Z0.B, Z0.B, P0.B       // 00000024
    ZCMPHS P3.Z, Z12.B, Z13.B, P5.B     // 850d0d24
    ZCMPHS P7.Z, Z31.B, Z31.B, P15.B    // ef1f1f24
    ZCMPHS P0.Z, Z0.H, Z0.H, P0.H       // 00004024
    ZCMPHS P3.Z, Z12.H, Z13.H, P5.H     // 850d4d24
    ZCMPHS P7.Z, Z31.H, Z31.H, P15.H    // ef1f5f24
    ZCMPHS P0.Z, Z0.S, Z0.S, P0.S       // 00008024
    ZCMPHS P3.Z, Z12.S, Z13.S, P5.S     // 850d8d24
    ZCMPHS P7.Z, Z31.S, Z31.S, P15.S    // ef1f9f24
    ZCMPHS P0.Z, Z0.D, Z0.D, P0.D       // 0000c024
    ZCMPHS P3.Z, Z12.D, Z13.D, P5.D     // 850dcd24
    ZCMPHS P7.Z, Z31.D, Z31.D, P15.D    // ef1fdf24

// CMPLE   <Pd>.<T>, <Pg>/Z, <Zn>.<T>, #<imm>
    ZCMPLE P0.Z, Z0.B, $-16, P0.B     // 10201025
    ZCMPLE P3.Z, Z12.B, $-6, P5.B     // 952d1a25
    ZCMPLE P7.Z, Z31.B, $15, P15.B    // ff3f0f25
    ZCMPLE P0.Z, Z0.H, $-16, P0.H     // 10205025
    ZCMPLE P3.Z, Z12.H, $-6, P5.H     // 952d5a25
    ZCMPLE P7.Z, Z31.H, $15, P15.H    // ff3f4f25
    ZCMPLE P0.Z, Z0.S, $-16, P0.S     // 10209025
    ZCMPLE P3.Z, Z12.S, $-6, P5.S     // 952d9a25
    ZCMPLE P7.Z, Z31.S, $15, P15.S    // ff3f8f25
    ZCMPLE P0.Z, Z0.D, $-16, P0.D     // 1020d025
    ZCMPLE P3.Z, Z12.D, $-6, P5.D     // 952dda25
    ZCMPLE P7.Z, Z31.D, $15, P15.D    // ff3fcf25

// CMPLE   <Pd>.<T>, <Pg>/Z, <Zn>.<T>, <Zm>.D
    ZCMPLE P0.Z, Z0.B, Z0.D, P0.B       // 10600024
    ZCMPLE P3.Z, Z12.B, Z13.D, P5.B     // 956d0d24
    ZCMPLE P7.Z, Z31.B, Z31.D, P15.B    // ff7f1f24
    ZCMPLE P0.Z, Z0.H, Z0.D, P0.H       // 10604024
    ZCMPLE P3.Z, Z12.H, Z13.D, P5.H     // 956d4d24
    ZCMPLE P7.Z, Z31.H, Z31.D, P15.H    // ff7f5f24
    ZCMPLE P0.Z, Z0.S, Z0.D, P0.S       // 10608024
    ZCMPLE P3.Z, Z12.S, Z13.D, P5.S     // 956d8d24
    ZCMPLE P7.Z, Z31.S, Z31.D, P15.S    // ff7f9f24

// CMPLO   <Pd>.<T>, <Pg>/Z, <Zn>.<T>, #<imm>
    ZCMPLO P0.Z, Z0.B, $0, P0.B        // 00202024
    ZCMPLO P3.Z, Z12.B, $42, P5.B      // 85ad2a24
    ZCMPLO P7.Z, Z31.B, $127, P15.B    // efff3f24
    ZCMPLO P0.Z, Z0.H, $0, P0.H        // 00206024
    ZCMPLO P3.Z, Z12.H, $42, P5.H      // 85ad6a24
    ZCMPLO P7.Z, Z31.H, $127, P15.H    // efff7f24
    ZCMPLO P0.Z, Z0.S, $0, P0.S        // 0020a024
    ZCMPLO P3.Z, Z12.S, $42, P5.S      // 85adaa24
    ZCMPLO P7.Z, Z31.S, $127, P15.S    // efffbf24
    ZCMPLO P0.Z, Z0.D, $0, P0.D        // 0020e024
    ZCMPLO P3.Z, Z12.D, $42, P5.D      // 85adea24
    ZCMPLO P7.Z, Z31.D, $127, P15.D    // efffff24

// CMPLO   <Pd>.<T>, <Pg>/Z, <Zn>.<T>, <Zm>.D
    ZCMPLO P0.Z, Z0.B, Z0.D, P0.B       // 00e00024
    ZCMPLO P3.Z, Z12.B, Z13.D, P5.B     // 85ed0d24
    ZCMPLO P7.Z, Z31.B, Z31.D, P15.B    // efff1f24
    ZCMPLO P0.Z, Z0.H, Z0.D, P0.H       // 00e04024
    ZCMPLO P3.Z, Z12.H, Z13.D, P5.H     // 85ed4d24
    ZCMPLO P7.Z, Z31.H, Z31.D, P15.H    // efff5f24
    ZCMPLO P0.Z, Z0.S, Z0.D, P0.S       // 00e08024
    ZCMPLO P3.Z, Z12.S, Z13.D, P5.S     // 85ed8d24
    ZCMPLO P7.Z, Z31.S, Z31.D, P15.S    // efff9f24

// CMPLS   <Pd>.<T>, <Pg>/Z, <Zn>.<T>, #<imm>
    ZCMPLS P0.Z, Z0.B, $0, P0.B        // 10202024
    ZCMPLS P3.Z, Z12.B, $42, P5.B      // 95ad2a24
    ZCMPLS P7.Z, Z31.B, $127, P15.B    // ffff3f24
    ZCMPLS P0.Z, Z0.H, $0, P0.H        // 10206024
    ZCMPLS P3.Z, Z12.H, $42, P5.H      // 95ad6a24
    ZCMPLS P7.Z, Z31.H, $127, P15.H    // ffff7f24
    ZCMPLS P0.Z, Z0.S, $0, P0.S        // 1020a024
    ZCMPLS P3.Z, Z12.S, $42, P5.S      // 95adaa24
    ZCMPLS P7.Z, Z31.S, $127, P15.S    // ffffbf24
    ZCMPLS P0.Z, Z0.D, $0, P0.D        // 1020e024
    ZCMPLS P3.Z, Z12.D, $42, P5.D      // 95adea24
    ZCMPLS P7.Z, Z31.D, $127, P15.D    // ffffff24

// CMPLS   <Pd>.<T>, <Pg>/Z, <Zn>.<T>, <Zm>.D
    ZCMPLS P0.Z, Z0.B, Z0.D, P0.B       // 10e00024
    ZCMPLS P3.Z, Z12.B, Z13.D, P5.B     // 95ed0d24
    ZCMPLS P7.Z, Z31.B, Z31.D, P15.B    // ffff1f24
    ZCMPLS P0.Z, Z0.H, Z0.D, P0.H       // 10e04024
    ZCMPLS P3.Z, Z12.H, Z13.D, P5.H     // 95ed4d24
    ZCMPLS P7.Z, Z31.H, Z31.D, P15.H    // ffff5f24
    ZCMPLS P0.Z, Z0.S, Z0.D, P0.S       // 10e08024
    ZCMPLS P3.Z, Z12.S, Z13.D, P5.S     // 95ed8d24
    ZCMPLS P7.Z, Z31.S, Z31.D, P15.S    // ffff9f24

// CMPLT   <Pd>.<T>, <Pg>/Z, <Zn>.<T>, #<imm>
    ZCMPLT P0.Z, Z0.B, $-16, P0.B     // 00201025
    ZCMPLT P3.Z, Z12.B, $-6, P5.B     // 852d1a25
    ZCMPLT P7.Z, Z31.B, $15, P15.B    // ef3f0f25
    ZCMPLT P0.Z, Z0.H, $-16, P0.H     // 00205025
    ZCMPLT P3.Z, Z12.H, $-6, P5.H     // 852d5a25
    ZCMPLT P7.Z, Z31.H, $15, P15.H    // ef3f4f25
    ZCMPLT P0.Z, Z0.S, $-16, P0.S     // 00209025
    ZCMPLT P3.Z, Z12.S, $-6, P5.S     // 852d9a25
    ZCMPLT P7.Z, Z31.S, $15, P15.S    // ef3f8f25
    ZCMPLT P0.Z, Z0.D, $-16, P0.D     // 0020d025
    ZCMPLT P3.Z, Z12.D, $-6, P5.D     // 852dda25
    ZCMPLT P7.Z, Z31.D, $15, P15.D    // ef3fcf25

// CMPLT   <Pd>.<T>, <Pg>/Z, <Zn>.<T>, <Zm>.D
    ZCMPLT P0.Z, Z0.B, Z0.D, P0.B       // 00600024
    ZCMPLT P3.Z, Z12.B, Z13.D, P5.B     // 856d0d24
    ZCMPLT P7.Z, Z31.B, Z31.D, P15.B    // ef7f1f24
    ZCMPLT P0.Z, Z0.H, Z0.D, P0.H       // 00604024
    ZCMPLT P3.Z, Z12.H, Z13.D, P5.H     // 856d4d24
    ZCMPLT P7.Z, Z31.H, Z31.D, P15.H    // ef7f5f24
    ZCMPLT P0.Z, Z0.S, Z0.D, P0.S       // 00608024
    ZCMPLT P3.Z, Z12.S, Z13.D, P5.S     // 856d8d24
    ZCMPLT P7.Z, Z31.S, Z31.D, P15.S    // ef7f9f24

// CMPNE   <Pd>.<T>, <Pg>/Z, <Zn>.<T>, #<imm>
    ZCMPNE P0.Z, Z0.B, $-16, P0.B     // 10801025
    ZCMPNE P3.Z, Z12.B, $-6, P5.B     // 958d1a25
    ZCMPNE P7.Z, Z31.B, $15, P15.B    // ff9f0f25
    ZCMPNE P0.Z, Z0.H, $-16, P0.H     // 10805025
    ZCMPNE P3.Z, Z12.H, $-6, P5.H     // 958d5a25
    ZCMPNE P7.Z, Z31.H, $15, P15.H    // ff9f4f25
    ZCMPNE P0.Z, Z0.S, $-16, P0.S     // 10809025
    ZCMPNE P3.Z, Z12.S, $-6, P5.S     // 958d9a25
    ZCMPNE P7.Z, Z31.S, $15, P15.S    // ff9f8f25
    ZCMPNE P0.Z, Z0.D, $-16, P0.D     // 1080d025
    ZCMPNE P3.Z, Z12.D, $-6, P5.D     // 958dda25
    ZCMPNE P7.Z, Z31.D, $15, P15.D    // ff9fcf25

// CMPNE   <Pd>.<T>, <Pg>/Z, <Zn>.<T>, <Zm>.D
    ZCMPNE P0.Z, Z0.B, Z0.D, P0.B       // 10200024
    ZCMPNE P3.Z, Z12.B, Z13.D, P5.B     // 952d0d24
    ZCMPNE P7.Z, Z31.B, Z31.D, P15.B    // ff3f1f24
    ZCMPNE P0.Z, Z0.H, Z0.D, P0.H       // 10204024
    ZCMPNE P3.Z, Z12.H, Z13.D, P5.H     // 952d4d24
    ZCMPNE P7.Z, Z31.H, Z31.D, P15.H    // ff3f5f24
    ZCMPNE P0.Z, Z0.S, Z0.D, P0.S       // 10208024
    ZCMPNE P3.Z, Z12.S, Z13.D, P5.S     // 952d8d24
    ZCMPNE P7.Z, Z31.S, Z31.D, P15.S    // ff3f9f24

// CMPNE   <Pd>.<T>, <Pg>/Z, <Zn>.<T>, <Zm>.<T>
    ZCMPNE P0.Z, Z0.B, Z0.B, P0.B       // 10a00024
    ZCMPNE P3.Z, Z12.B, Z13.B, P5.B     // 95ad0d24
    ZCMPNE P7.Z, Z31.B, Z31.B, P15.B    // ffbf1f24
    ZCMPNE P0.Z, Z0.H, Z0.H, P0.H       // 10a04024
    ZCMPNE P3.Z, Z12.H, Z13.H, P5.H     // 95ad4d24
    ZCMPNE P7.Z, Z31.H, Z31.H, P15.H    // ffbf5f24
    ZCMPNE P0.Z, Z0.S, Z0.S, P0.S       // 10a08024
    ZCMPNE P3.Z, Z12.S, Z13.S, P5.S     // 95ad8d24
    ZCMPNE P7.Z, Z31.S, Z31.S, P15.S    // ffbf9f24
    ZCMPNE P0.Z, Z0.D, Z0.D, P0.D       // 10a0c024
    ZCMPNE P3.Z, Z12.D, Z13.D, P5.D     // 95adcd24
    ZCMPNE P7.Z, Z31.D, Z31.D, P15.D    // ffbfdf24

// CNOT    <Zd>.<T>, <Pg>/M, <Zn>.<T>
    ZCNOT P0.M, Z0.B, Z0.B      // 00a01b04
    ZCNOT P3.M, Z12.B, Z10.B    // 8aad1b04
    ZCNOT P7.M, Z31.B, Z31.B    // ffbf1b04
    ZCNOT P0.M, Z0.H, Z0.H      // 00a05b04
    ZCNOT P3.M, Z12.H, Z10.H    // 8aad5b04
    ZCNOT P7.M, Z31.H, Z31.H    // ffbf5b04
    ZCNOT P0.M, Z0.S, Z0.S      // 00a09b04
    ZCNOT P3.M, Z12.S, Z10.S    // 8aad9b04
    ZCNOT P7.M, Z31.S, Z31.S    // ffbf9b04
    ZCNOT P0.M, Z0.D, Z0.D      // 00a0db04
    ZCNOT P3.M, Z12.D, Z10.D    // 8aaddb04
    ZCNOT P7.M, Z31.D, Z31.D    // ffbfdb04

// CNT     <Zd>.<T>, <Pg>/M, <Zn>.<T>
    ZCNT P0.M, Z0.B, Z0.B      // 00a01a04
    ZCNT P3.M, Z12.B, Z10.B    // 8aad1a04
    ZCNT P7.M, Z31.B, Z31.B    // ffbf1a04
    ZCNT P0.M, Z0.H, Z0.H      // 00a05a04
    ZCNT P3.M, Z12.H, Z10.H    // 8aad5a04
    ZCNT P7.M, Z31.H, Z31.H    // ffbf5a04
    ZCNT P0.M, Z0.S, Z0.S      // 00a09a04
    ZCNT P3.M, Z12.S, Z10.S    // 8aad9a04
    ZCNT P7.M, Z31.S, Z31.S    // ffbf9a04
    ZCNT P0.M, Z0.D, Z0.D      // 00a0da04
    ZCNT P3.M, Z12.D, Z10.D    // 8aadda04
    ZCNT P7.M, Z31.D, Z31.D    // ffbfda04

// CNTB    <Xd>{, <pattern>{, MUL #<imm>}}
    ZCNTB POW2, $1, R0      // 00e02004
    ZCNTB VL1, $1, R1       // 21e02004
    ZCNTB VL2, $2, R2       // 42e02104
    ZCNTB VL3, $2, R3       // 63e02104
    ZCNTB VL4, $3, R4       // 84e02204
    ZCNTB VL5, $3, R5       // a5e02204
    ZCNTB VL6, $4, R6       // c6e02304
    ZCNTB VL7, $4, R7       // e7e02304
    ZCNTB VL8, $5, R8       // 08e12404
    ZCNTB VL16, $5, R8      // 28e12404
    ZCNTB VL32, $6, R9      // 49e12504
    ZCNTB VL64, $6, R10     // 6ae12504
    ZCNTB VL128, $7, R11    // 8be12604
    ZCNTB VL256, $7, R12    // ace12604
    ZCNTB $14, $8, R13      // cde12704
    ZCNTB $15, $8, R14      // eee12704
    ZCNTB $16, $9, R15      // 0fe22804
    ZCNTB $17, $9, R16      // 30e22804
    ZCNTB $18, $9, R17      // 51e22804
    ZCNTB $19, $10, R17     // 71e22904
    ZCNTB $20, $10, R19     // 93e22904
    ZCNTB $21, $11, R20     // b4e22a04
    ZCNTB $22, $11, R21     // d5e22a04
    ZCNTB $23, $12, R22     // f6e22b04
    ZCNTB $24, $12, R22     // 16e32b04
    ZCNTB $25, $13, R23     // 37e32c04
    ZCNTB $26, $13, R24     // 58e32c04
    ZCNTB $27, $14, R25     // 79e32d04
    ZCNTB $28, $14, R26     // 9ae32d04
    ZCNTB MUL4, $15, R27    // bbe32e04
    ZCNTB MUL3, $15, R27    // dbe32e04
    ZCNTB ALL, $16, R30     // fee32f04

// CNTD    <Xd>{, <pattern>{, MUL #<imm>}}
    ZCNTD POW2, $1, R0      // 00e0e004
    ZCNTD VL1, $1, R1       // 21e0e004
    ZCNTD VL2, $2, R2       // 42e0e104
    ZCNTD VL3, $2, R3       // 63e0e104
    ZCNTD VL4, $3, R4       // 84e0e204
    ZCNTD VL5, $3, R5       // a5e0e204
    ZCNTD VL6, $4, R6       // c6e0e304
    ZCNTD VL7, $4, R7       // e7e0e304
    ZCNTD VL8, $5, R8       // 08e1e404
    ZCNTD VL16, $5, R8      // 28e1e404
    ZCNTD VL32, $6, R9      // 49e1e504
    ZCNTD VL64, $6, R10     // 6ae1e504
    ZCNTD VL128, $7, R11    // 8be1e604
    ZCNTD VL256, $7, R12    // ace1e604
    ZCNTD $14, $8, R13      // cde1e704
    ZCNTD $15, $8, R14      // eee1e704
    ZCNTD $16, $9, R15      // 0fe2e804
    ZCNTD $17, $9, R16      // 30e2e804
    ZCNTD $18, $9, R17      // 51e2e804
    ZCNTD $19, $10, R17     // 71e2e904
    ZCNTD $20, $10, R19     // 93e2e904
    ZCNTD $21, $11, R20     // b4e2ea04
    ZCNTD $22, $11, R21     // d5e2ea04
    ZCNTD $23, $12, R22     // f6e2eb04
    ZCNTD $24, $12, R22     // 16e3eb04
    ZCNTD $25, $13, R23     // 37e3ec04
    ZCNTD $26, $13, R24     // 58e3ec04
    ZCNTD $27, $14, R25     // 79e3ed04
    ZCNTD $28, $14, R26     // 9ae3ed04
    ZCNTD MUL4, $15, R27    // bbe3ee04
    ZCNTD MUL3, $15, R27    // dbe3ee04
    ZCNTD ALL, $16, R30     // fee3ef04

// CNTH    <Xd>{, <pattern>{, MUL #<imm>}}
    ZCNTH POW2, $1, R0      // 00e06004
    ZCNTH VL1, $1, R1       // 21e06004
    ZCNTH VL2, $2, R2       // 42e06104
    ZCNTH VL3, $2, R3       // 63e06104
    ZCNTH VL4, $3, R4       // 84e06204
    ZCNTH VL5, $3, R5       // a5e06204
    ZCNTH VL6, $4, R6       // c6e06304
    ZCNTH VL7, $4, R7       // e7e06304
    ZCNTH VL8, $5, R8       // 08e16404
    ZCNTH VL16, $5, R8      // 28e16404
    ZCNTH VL32, $6, R9      // 49e16504
    ZCNTH VL64, $6, R10     // 6ae16504
    ZCNTH VL128, $7, R11    // 8be16604
    ZCNTH VL256, $7, R12    // ace16604
    ZCNTH $14, $8, R13      // cde16704
    ZCNTH $15, $8, R14      // eee16704
    ZCNTH $16, $9, R15      // 0fe26804
    ZCNTH $17, $9, R16      // 30e26804
    ZCNTH $18, $9, R17      // 51e26804
    ZCNTH $19, $10, R17     // 71e26904
    ZCNTH $20, $10, R19     // 93e26904
    ZCNTH $21, $11, R20     // b4e26a04
    ZCNTH $22, $11, R21     // d5e26a04
    ZCNTH $23, $12, R22     // f6e26b04
    ZCNTH $24, $12, R22     // 16e36b04
    ZCNTH $25, $13, R23     // 37e36c04
    ZCNTH $26, $13, R24     // 58e36c04
    ZCNTH $27, $14, R25     // 79e36d04
    ZCNTH $28, $14, R26     // 9ae36d04
    ZCNTH MUL4, $15, R27    // bbe36e04
    ZCNTH MUL3, $15, R27    // dbe36e04
    ZCNTH ALL, $16, R30     // fee36f04

// CNTP    <Xd>, <Pg>, <Pn>.<T>
    PCNTP P0, P0.B, R0       // 00802025
    PCNTP P6, P7.B, R10      // ea982025
    PCNTP P15, P15.B, R30    // febd2025
    PCNTP P0, P0.H, R0       // 00806025
    PCNTP P6, P7.H, R10      // ea986025
    PCNTP P15, P15.H, R30    // febd6025
    PCNTP P0, P0.S, R0       // 0080a025
    PCNTP P6, P7.S, R10      // ea98a025
    PCNTP P15, P15.S, R30    // febda025
    PCNTP P0, P0.D, R0       // 0080e025
    PCNTP P6, P7.D, R10      // ea98e025
    PCNTP P15, P15.D, R30    // febde025

// CNTW    <Xd>{, <pattern>{, MUL #<imm>}}
    ZCNTW POW2, $1, R0      // 00e0a004
    ZCNTW VL1, $1, R1       // 21e0a004
    ZCNTW VL2, $2, R2       // 42e0a104
    ZCNTW VL3, $2, R3       // 63e0a104
    ZCNTW VL4, $3, R4       // 84e0a204
    ZCNTW VL5, $3, R5       // a5e0a204
    ZCNTW VL6, $4, R6       // c6e0a304
    ZCNTW VL7, $4, R7       // e7e0a304
    ZCNTW VL8, $5, R8       // 08e1a404
    ZCNTW VL16, $5, R8      // 28e1a404
    ZCNTW VL32, $6, R9      // 49e1a504
    ZCNTW VL64, $6, R10     // 6ae1a504
    ZCNTW VL128, $7, R11    // 8be1a604
    ZCNTW VL256, $7, R12    // ace1a604
    ZCNTW $14, $8, R13      // cde1a704
    ZCNTW $15, $8, R14      // eee1a704
    ZCNTW $16, $9, R15      // 0fe2a804
    ZCNTW $17, $9, R16      // 30e2a804
    ZCNTW $18, $9, R17      // 51e2a804
    ZCNTW $19, $10, R17     // 71e2a904
    ZCNTW $20, $10, R19     // 93e2a904
    ZCNTW $21, $11, R20     // b4e2aa04
    ZCNTW $22, $11, R21     // d5e2aa04
    ZCNTW $23, $12, R22     // f6e2ab04
    ZCNTW $24, $12, R22     // 16e3ab04
    ZCNTW $25, $13, R23     // 37e3ac04
    ZCNTW $26, $13, R24     // 58e3ac04
    ZCNTW $27, $14, R25     // 79e3ad04
    ZCNTW $28, $14, R26     // 9ae3ad04
    ZCNTW MUL4, $15, R27    // bbe3ae04
    ZCNTW MUL3, $15, R27    // dbe3ae04
    ZCNTW ALL, $16, R30     // fee3af04

// COMPACT <Zd>.<T>, <Pg>, <Zn>.<T>
    ZCOMPACT P0, Z0.S, Z0.S      // 0080a105
    ZCOMPACT P3, Z12.S, Z10.S    // 8a8da105
    ZCOMPACT P7, Z31.S, Z31.S    // ff9fa105
    ZCOMPACT P0, Z0.D, Z0.D      // 0080e105
    ZCOMPACT P3, Z12.D, Z10.D    // 8a8de105
    ZCOMPACT P7, Z31.D, Z31.D    // ff9fe105

// CPY     <Zd>.<T>, <Pg>/Z, #<imm>, <shift>
    ZCPY P0.Z, $-128, $0, Z0.B     // 00101005
    ZCPY P6.Z, $-43, $0, Z10.B     // aa1a1605
    ZCPY P15.Z, $127, $0, Z31.B    // ff0f1f05
    ZCPY P0.Z, $-128, $8, Z0.H     // 00305005
    ZCPY P6.Z, $-43, $8, Z10.H     // aa3a5605
    ZCPY P15.Z, $127, $0, Z31.H    // ff0f5f05
    ZCPY P0.Z, $-128, $8, Z0.S     // 00309005
    ZCPY P6.Z, $-43, $8, Z10.S     // aa3a9605
    ZCPY P15.Z, $127, $0, Z31.S    // ff0f9f05
    ZCPY P0.Z, $-128, $8, Z0.D     // 0030d005
    ZCPY P6.Z, $-43, $8, Z10.D     // aa3ad605
    ZCPY P15.Z, $127, $0, Z31.D    // ff0fdf05

// CPY     <Zd>.<T>, <Pg>/M, #<imm>, <shift>
    ZCPY P0.M, $-128, $0, Z0.B     // 00501005
    ZCPY P6.M, $-43, $0, Z10.B     // aa5a1605
    ZCPY P15.M, $127, $0, Z31.B    // ff4f1f05
    ZCPY P0.M, $-128, $8, Z0.H     // 00705005
    ZCPY P6.M, $-43, $8, Z10.H     // aa7a5605
    ZCPY P15.M, $127, $0, Z31.H    // ff4f5f05
    ZCPY P0.M, $-128, $8, Z0.S     // 00709005
    ZCPY P6.M, $-43, $8, Z10.S     // aa7a9605
    ZCPY P15.M, $127, $0, Z31.S    // ff4f9f05
    ZCPY P0.M, $-128, $8, Z0.D     // 0070d005
    ZCPY P6.M, $-43, $8, Z10.D     // aa7ad605
    ZCPY P15.M, $127, $0, Z31.D    // ff4fdf05

// CPY     <Zd>.<T>, <Pg>/M, <R><n|SP>
    ZCPY P0.M, R0, Z0.B      // 00a02805
    ZCPY P3.M, R12, Z10.B    // 8aad2805
    ZCPY P7.M, R30, Z31.B    // dfbf2805
    ZCPY P0.M, R0, Z0.H      // 00a06805
    ZCPY P3.M, R12, Z10.H    // 8aad6805
    ZCPY P7.M, R30, Z31.H    // dfbf6805
    ZCPY P0.M, R0, Z0.S      // 00a0a805
    ZCPY P3.M, R12, Z10.S    // 8aada805
    ZCPY P7.M, R30, Z31.S    // dfbfa805
    ZCPY P0.M, R0, Z0.D      // 00a0e805
    ZCPY P3.M, R12, Z10.D    // 8aade805
    ZCPY P7.M, R30, Z31.D    // dfbfe805

// CPY     <Zd>.<T>, <Pg>/M, <V><n>
    ZCPY P0.M, V0, Z0.B      // 00802005
    ZCPY P3.M, V12, Z10.B    // 8a8d2005
    ZCPY P7.M, V31, Z31.B    // ff9f2005
    ZCPY P0.M, V0, Z0.H      // 00806005
    ZCPY P3.M, V12, Z10.H    // 8a8d6005
    ZCPY P7.M, V31, Z31.H    // ff9f6005
    ZCPY P0.M, V0, Z0.S      // 0080a005
    ZCPY P3.M, V12, Z10.S    // 8a8da005
    ZCPY P7.M, V31, Z31.S    // ff9fa005
    ZCPY P0.M, V0, Z0.D      // 0080e005
    ZCPY P3.M, V12, Z10.D    // 8a8de005
    ZCPY P7.M, V31, Z31.D    // ff9fe005

// CTERMEQ <R><n>, <R><m>
    ZCTERMEQW R0, R0      // 0020a025
    ZCTERMEQW R10, R11    // 4021ab25
    ZCTERMEQW R30, R30    // c023be25
    ZCTERMEQ R0, R0       // 0020e025
    ZCTERMEQ R10, R11     // 4021eb25
    ZCTERMEQ R30, R30     // c023fe25

// CTERMNE <R><n>, <R><m>
    ZCTERMNEW R0, R0      // 1020a025
    ZCTERMNEW R10, R11    // 5021ab25
    ZCTERMNEW R30, R30    // d023be25
    ZCTERMNE R0, R0       // 1020e025
    ZCTERMNE R10, R11     // 5021eb25
    ZCTERMNE R30, R30     // d023fe25

// DECB    <Xdn>{, <pattern>{, MUL #<imm>}}
    ZDECB POW2, $1, R0      // 00e43004
    ZDECB VL1, $1, R1       // 21e43004
    ZDECB VL2, $2, R2       // 42e43104
    ZDECB VL3, $2, R3       // 63e43104
    ZDECB VL4, $3, R4       // 84e43204
    ZDECB VL5, $3, R5       // a5e43204
    ZDECB VL6, $4, R6       // c6e43304
    ZDECB VL7, $4, R7       // e7e43304
    ZDECB VL8, $5, R8       // 08e53404
    ZDECB VL16, $5, R8      // 28e53404
    ZDECB VL32, $6, R9      // 49e53504
    ZDECB VL64, $6, R10     // 6ae53504
    ZDECB VL128, $7, R11    // 8be53604
    ZDECB VL256, $7, R12    // ace53604
    ZDECB $14, $8, R13      // cde53704
    ZDECB $15, $8, R14      // eee53704
    ZDECB $16, $9, R15      // 0fe63804
    ZDECB $17, $9, R16      // 30e63804
    ZDECB $18, $9, R17      // 51e63804
    ZDECB $19, $10, R17     // 71e63904
    ZDECB $20, $10, R19     // 93e63904
    ZDECB $21, $11, R20     // b4e63a04
    ZDECB $22, $11, R21     // d5e63a04
    ZDECB $23, $12, R22     // f6e63b04
    ZDECB $24, $12, R22     // 16e73b04
    ZDECB $25, $13, R23     // 37e73c04
    ZDECB $26, $13, R24     // 58e73c04
    ZDECB $27, $14, R25     // 79e73d04
    ZDECB $28, $14, R26     // 9ae73d04
    ZDECB MUL4, $15, R27    // bbe73e04
    ZDECB MUL3, $15, R27    // dbe73e04
    ZDECB ALL, $16, R30     // fee73f04

// DECD    <Xdn>{, <pattern>{, MUL #<imm>}}
    ZDECD POW2, $1, R0      // 00e4f004
    ZDECD VL1, $1, R1       // 21e4f004
    ZDECD VL2, $2, R2       // 42e4f104
    ZDECD VL3, $2, R3       // 63e4f104
    ZDECD VL4, $3, R4       // 84e4f204
    ZDECD VL5, $3, R5       // a5e4f204
    ZDECD VL6, $4, R6       // c6e4f304
    ZDECD VL7, $4, R7       // e7e4f304
    ZDECD VL8, $5, R8       // 08e5f404
    ZDECD VL16, $5, R8      // 28e5f404
    ZDECD VL32, $6, R9      // 49e5f504
    ZDECD VL64, $6, R10     // 6ae5f504
    ZDECD VL128, $7, R11    // 8be5f604
    ZDECD VL256, $7, R12    // ace5f604
    ZDECD $14, $8, R13      // cde5f704
    ZDECD $15, $8, R14      // eee5f704
    ZDECD $16, $9, R15      // 0fe6f804
    ZDECD $17, $9, R16      // 30e6f804
    ZDECD $18, $9, R17      // 51e6f804
    ZDECD $19, $10, R17     // 71e6f904
    ZDECD $20, $10, R19     // 93e6f904
    ZDECD $21, $11, R20     // b4e6fa04
    ZDECD $22, $11, R21     // d5e6fa04
    ZDECD $23, $12, R22     // f6e6fb04
    ZDECD $24, $12, R22     // 16e7fb04
    ZDECD $25, $13, R23     // 37e7fc04
    ZDECD $26, $13, R24     // 58e7fc04
    ZDECD $27, $14, R25     // 79e7fd04
    ZDECD $28, $14, R26     // 9ae7fd04
    ZDECD MUL4, $15, R27    // bbe7fe04
    ZDECD MUL3, $15, R27    // dbe7fe04
    ZDECD ALL, $16, R30     // fee7ff04

// DECD    <Zdn>.D{, <pattern>{, MUL #<imm>}}
    ZDECD POW2, $1, Z0.D      // 00c4f004
    ZDECD VL1, $1, Z1.D       // 21c4f004
    ZDECD VL2, $2, Z2.D       // 42c4f104
    ZDECD VL3, $2, Z3.D       // 63c4f104
    ZDECD VL4, $3, Z4.D       // 84c4f204
    ZDECD VL5, $3, Z5.D       // a5c4f204
    ZDECD VL6, $4, Z6.D       // c6c4f304
    ZDECD VL7, $4, Z7.D       // e7c4f304
    ZDECD VL8, $5, Z8.D       // 08c5f404
    ZDECD VL16, $5, Z9.D      // 29c5f404
    ZDECD VL32, $6, Z10.D     // 4ac5f504
    ZDECD VL64, $6, Z11.D     // 6bc5f504
    ZDECD VL128, $7, Z12.D    // 8cc5f604
    ZDECD VL256, $7, Z13.D    // adc5f604
    ZDECD $14, $8, Z14.D      // cec5f704
    ZDECD $15, $8, Z15.D      // efc5f704
    ZDECD $16, $9, Z16.D      // 10c6f804
    ZDECD $17, $9, Z16.D      // 30c6f804
    ZDECD $18, $9, Z17.D      // 51c6f804
    ZDECD $19, $10, Z18.D     // 72c6f904
    ZDECD $20, $10, Z19.D     // 93c6f904
    ZDECD $21, $11, Z20.D     // b4c6fa04
    ZDECD $22, $11, Z21.D     // d5c6fa04
    ZDECD $23, $12, Z22.D     // f6c6fb04
    ZDECD $24, $12, Z23.D     // 17c7fb04
    ZDECD $25, $13, Z24.D     // 38c7fc04
    ZDECD $26, $13, Z25.D     // 59c7fc04
    ZDECD $27, $14, Z26.D     // 7ac7fd04
    ZDECD $28, $14, Z27.D     // 9bc7fd04
    ZDECD MUL4, $15, Z28.D    // bcc7fe04
    ZDECD MUL3, $15, Z29.D    // ddc7fe04
    ZDECD ALL, $16, Z31.D     // ffc7ff04

// DECH    <Xdn>{, <pattern>{, MUL #<imm>}}
    ZDECH POW2, $1, R0      // 00e47004
    ZDECH VL1, $1, R1       // 21e47004
    ZDECH VL2, $2, R2       // 42e47104
    ZDECH VL3, $2, R3       // 63e47104
    ZDECH VL4, $3, R4       // 84e47204
    ZDECH VL5, $3, R5       // a5e47204
    ZDECH VL6, $4, R6       // c6e47304
    ZDECH VL7, $4, R7       // e7e47304
    ZDECH VL8, $5, R8       // 08e57404
    ZDECH VL16, $5, R8      // 28e57404
    ZDECH VL32, $6, R9      // 49e57504
    ZDECH VL64, $6, R10     // 6ae57504
    ZDECH VL128, $7, R11    // 8be57604
    ZDECH VL256, $7, R12    // ace57604
    ZDECH $14, $8, R13      // cde57704
    ZDECH $15, $8, R14      // eee57704
    ZDECH $16, $9, R15      // 0fe67804
    ZDECH $17, $9, R16      // 30e67804
    ZDECH $18, $9, R17      // 51e67804
    ZDECH $19, $10, R17     // 71e67904
    ZDECH $20, $10, R19     // 93e67904
    ZDECH $21, $11, R20     // b4e67a04
    ZDECH $22, $11, R21     // d5e67a04
    ZDECH $23, $12, R22     // f6e67b04
    ZDECH $24, $12, R22     // 16e77b04
    ZDECH $25, $13, R23     // 37e77c04
    ZDECH $26, $13, R24     // 58e77c04
    ZDECH $27, $14, R25     // 79e77d04
    ZDECH $28, $14, R26     // 9ae77d04
    ZDECH MUL4, $15, R27    // bbe77e04
    ZDECH MUL3, $15, R27    // dbe77e04
    ZDECH ALL, $16, R30     // fee77f04

// DECH    <Zdn>.H{, <pattern>{, MUL #<imm>}}
    ZDECH POW2, $1, Z0.H      // 00c47004
    ZDECH VL1, $1, Z1.H       // 21c47004
    ZDECH VL2, $2, Z2.H       // 42c47104
    ZDECH VL3, $2, Z3.H       // 63c47104
    ZDECH VL4, $3, Z4.H       // 84c47204
    ZDECH VL5, $3, Z5.H       // a5c47204
    ZDECH VL6, $4, Z6.H       // c6c47304
    ZDECH VL7, $4, Z7.H       // e7c47304
    ZDECH VL8, $5, Z8.H       // 08c57404
    ZDECH VL16, $5, Z9.H      // 29c57404
    ZDECH VL32, $6, Z10.H     // 4ac57504
    ZDECH VL64, $6, Z11.H     // 6bc57504
    ZDECH VL128, $7, Z12.H    // 8cc57604
    ZDECH VL256, $7, Z13.H    // adc57604
    ZDECH $14, $8, Z14.H      // cec57704
    ZDECH $15, $8, Z15.H      // efc57704
    ZDECH $16, $9, Z16.H      // 10c67804
    ZDECH $17, $9, Z16.H      // 30c67804
    ZDECH $18, $9, Z17.H      // 51c67804
    ZDECH $19, $10, Z18.H     // 72c67904
    ZDECH $20, $10, Z19.H     // 93c67904
    ZDECH $21, $11, Z20.H     // b4c67a04
    ZDECH $22, $11, Z21.H     // d5c67a04
    ZDECH $23, $12, Z22.H     // f6c67b04
    ZDECH $24, $12, Z23.H     // 17c77b04
    ZDECH $25, $13, Z24.H     // 38c77c04
    ZDECH $26, $13, Z25.H     // 59c77c04
    ZDECH $27, $14, Z26.H     // 7ac77d04
    ZDECH $28, $14, Z27.H     // 9bc77d04
    ZDECH MUL4, $15, Z28.H    // bcc77e04
    ZDECH MUL3, $15, Z29.H    // ddc77e04
    ZDECH ALL, $16, Z31.H     // ffc77f04

// DECP    <Xdn>, <Pm>.<T>
    PDECP P0.B, R0      // 00882d25
    PDECP P6.B, R10     // ca882d25
    PDECP P15.B, R30    // fe892d25
    PDECP P0.H, R0      // 00886d25
    PDECP P6.H, R10     // ca886d25
    PDECP P15.H, R30    // fe896d25
    PDECP P0.S, R0      // 0088ad25
    PDECP P6.S, R10     // ca88ad25
    PDECP P15.S, R30    // fe89ad25
    PDECP P0.D, R0      // 0088ed25
    PDECP P6.D, R10     // ca88ed25
    PDECP P15.D, R30    // fe89ed25

// DECP    <Zdn>.<T>, <Pm>.<T>
    ZDECP P0.H, Z0.H      // 00806d25
    ZDECP P6.H, Z10.H     // ca806d25
    ZDECP P15.H, Z31.H    // ff816d25
    ZDECP P0.S, Z0.S      // 0080ad25
    ZDECP P6.S, Z10.S     // ca80ad25
    ZDECP P15.S, Z31.S    // ff81ad25
    ZDECP P0.D, Z0.D      // 0080ed25
    ZDECP P6.D, Z10.D     // ca80ed25
    ZDECP P15.D, Z31.D    // ff81ed25

// DECW    <Xdn>{, <pattern>{, MUL #<imm>}}
    ZDECW POW2, $1, R0      // 00e4b004
    ZDECW VL1, $1, R1       // 21e4b004
    ZDECW VL2, $2, R2       // 42e4b104
    ZDECW VL3, $2, R3       // 63e4b104
    ZDECW VL4, $3, R4       // 84e4b204
    ZDECW VL5, $3, R5       // a5e4b204
    ZDECW VL6, $4, R6       // c6e4b304
    ZDECW VL7, $4, R7       // e7e4b304
    ZDECW VL8, $5, R8       // 08e5b404
    ZDECW VL16, $5, R8      // 28e5b404
    ZDECW VL32, $6, R9      // 49e5b504
    ZDECW VL64, $6, R10     // 6ae5b504
    ZDECW VL128, $7, R11    // 8be5b604
    ZDECW VL256, $7, R12    // ace5b604
    ZDECW $14, $8, R13      // cde5b704
    ZDECW $15, $8, R14      // eee5b704
    ZDECW $16, $9, R15      // 0fe6b804
    ZDECW $17, $9, R16      // 30e6b804
    ZDECW $18, $9, R17      // 51e6b804
    ZDECW $19, $10, R17     // 71e6b904
    ZDECW $20, $10, R19     // 93e6b904
    ZDECW $21, $11, R20     // b4e6ba04
    ZDECW $22, $11, R21     // d5e6ba04
    ZDECW $23, $12, R22     // f6e6bb04
    ZDECW $24, $12, R22     // 16e7bb04
    ZDECW $25, $13, R23     // 37e7bc04
    ZDECW $26, $13, R24     // 58e7bc04
    ZDECW $27, $14, R25     // 79e7bd04
    ZDECW $28, $14, R26     // 9ae7bd04
    ZDECW MUL4, $15, R27    // bbe7be04
    ZDECW MUL3, $15, R27    // dbe7be04
    ZDECW ALL, $16, R30     // fee7bf04

// DECW    <Zdn>.S{, <pattern>{, MUL #<imm>}}
    ZDECW POW2, $1, Z0.S      // 00c4b004
    ZDECW VL1, $1, Z1.S       // 21c4b004
    ZDECW VL2, $2, Z2.S       // 42c4b104
    ZDECW VL3, $2, Z3.S       // 63c4b104
    ZDECW VL4, $3, Z4.S       // 84c4b204
    ZDECW VL5, $3, Z5.S       // a5c4b204
    ZDECW VL6, $4, Z6.S       // c6c4b304
    ZDECW VL7, $4, Z7.S       // e7c4b304
    ZDECW VL8, $5, Z8.S       // 08c5b404
    ZDECW VL16, $5, Z9.S      // 29c5b404
    ZDECW VL32, $6, Z10.S     // 4ac5b504
    ZDECW VL64, $6, Z11.S     // 6bc5b504
    ZDECW VL128, $7, Z12.S    // 8cc5b604
    ZDECW VL256, $7, Z13.S    // adc5b604
    ZDECW $14, $8, Z14.S      // cec5b704
    ZDECW $15, $8, Z15.S      // efc5b704
    ZDECW $16, $9, Z16.S      // 10c6b804
    ZDECW $17, $9, Z16.S      // 30c6b804
    ZDECW $18, $9, Z17.S      // 51c6b804
    ZDECW $19, $10, Z18.S     // 72c6b904
    ZDECW $20, $10, Z19.S     // 93c6b904
    ZDECW $21, $11, Z20.S     // b4c6ba04
    ZDECW $22, $11, Z21.S     // d5c6ba04
    ZDECW $23, $12, Z22.S     // f6c6bb04
    ZDECW $24, $12, Z23.S     // 17c7bb04
    ZDECW $25, $13, Z24.S     // 38c7bc04
    ZDECW $26, $13, Z25.S     // 59c7bc04
    ZDECW $27, $14, Z26.S     // 7ac7bd04
    ZDECW $28, $14, Z27.S     // 9bc7bd04
    ZDECW MUL4, $15, Z28.S    // bcc7be04
    ZDECW MUL3, $15, Z29.S    // ddc7be04
    ZDECW ALL, $16, Z31.S     // ffc7bf04

// DUP     <Zd>.<T>, #<imm>, <shift>
    ZDUP $-128, $0, Z0.B    // 00d03825
    ZDUP $-43, $0, Z10.B    // aada3825
    ZDUP $127, $0, Z31.B    // ffcf3825
    ZDUP $-128, $8, Z0.H    // 00f07825
    ZDUP $-43, $8, Z10.H    // aafa7825
    ZDUP $127, $0, Z31.H    // ffcf7825
    ZDUP $-128, $8, Z0.S    // 00f0b825
    ZDUP $-43, $8, Z10.S    // aafab825
    ZDUP $127, $0, Z31.S    // ffcfb825
    ZDUP $-128, $8, Z0.D    // 00f0f825
    ZDUP $-43, $8, Z10.D    // aafaf825
    ZDUP $127, $0, Z31.D    // ffcff825

// DUP     <Zd>.<T>, <R><n|SP>
    ZDUP R0, Z0.B      // 00382005
    ZDUP R11, Z10.B    // 6a392005
    ZDUP R30, Z31.B    // df3b2005
    ZDUP R0, Z0.H      // 00386005
    ZDUP R11, Z10.H    // 6a396005
    ZDUP R30, Z31.H    // df3b6005
    ZDUP R0, Z0.S      // 0038a005
    ZDUP R11, Z10.S    // 6a39a005
    ZDUP R30, Z31.S    // df3ba005
    ZDUP R0, Z0.D      // 0038e005
    ZDUP R11, Z10.D    // 6a39e005
    ZDUP R30, Z31.D    // df3be005

// DUP     <Zd>.<T>, <Zn>.<T>[<imm>]
    ZDUP Z0.Q[0], Z0.Q       // 00203005
    ZDUP Z11.Q[1], Z10.Q     // 6a217005
    ZDUP Z31.Q[3], Z31.Q     // ff23f005
    ZDUP Z0.D[0], Z0.D       // 00202805
    ZDUP Z11.D[2], Z10.D     // 6a216805
    ZDUP Z31.D[7], Z31.D     // ff23f805
    ZDUP Z0.S[0], Z0.S       // 00202405
    ZDUP Z11.S[5], Z10.S     // 6a216c05
    ZDUP Z31.S[15], Z31.S    // ff23fc05
    ZDUP Z0.H[0], Z0.H       // 00202205
    ZDUP Z11.H[10], Z10.H    // 6a216a05
    ZDUP Z31.H[31], Z31.H    // ff23fe05
    ZDUP Z0.B[0], Z0.B       // 00202105
    ZDUP Z11.B[21], Z10.B    // 6a216b05
    ZDUP Z31.B[63], Z31.B    // ff23ff05

// DUPM    <Zd>.D, #<const>
    ZDUPM $1, Z0.S                  // 0000c005
    ZDUPM $4192256, Z10.S           // 4aa9c005
    ZDUPM $1, Z0.H                  // 0004c005
    ZDUPM $63489, Z10.H             // aa2cc005
    ZDUPM $1, Z0.B                  // 0006c005
    ZDUPM $56, Z10.B                // 4a2ec005
    ZDUPM $17, Z0.B                 // 0007c005
    ZDUPM $153, Z10.B               // 2a0fc005
    ZDUPM $4294967297, Z0.D         // 0000c005
    ZDUPM $-8787503089663, Z10.D    // aaaac005

// EOR     <Pd>.B, <Pg>/Z, <Pn>.B, <Pm>.B
    PEOR P0.Z, P0.B, P0.B, P0.B        // 00420025
    PEOR P6.Z, P7.B, P8.B, P5.B        // e55a0825
    PEOR P15.Z, P15.B, P15.B, P15.B    // ef7f0f25

// EOR     <Zdn>.<T>, <Pg>/M, <Zdn>.<T>, <Zm>.<T>
    ZEOR P0.M, Z0.B, Z0.B, Z0.B       // 00001904
    ZEOR P3.M, Z10.B, Z12.B, Z10.B    // 8a0d1904
    ZEOR P7.M, Z31.B, Z31.B, Z31.B    // ff1f1904
    ZEOR P0.M, Z0.H, Z0.H, Z0.H       // 00005904
    ZEOR P3.M, Z10.H, Z12.H, Z10.H    // 8a0d5904
    ZEOR P7.M, Z31.H, Z31.H, Z31.H    // ff1f5904
    ZEOR P0.M, Z0.S, Z0.S, Z0.S       // 00009904
    ZEOR P3.M, Z10.S, Z12.S, Z10.S    // 8a0d9904
    ZEOR P7.M, Z31.S, Z31.S, Z31.S    // ff1f9904
    ZEOR P0.M, Z0.D, Z0.D, Z0.D       // 0000d904
    ZEOR P3.M, Z10.D, Z12.D, Z10.D    // 8a0dd904
    ZEOR P7.M, Z31.D, Z31.D, Z31.D    // ff1fd904

// EOR     <Zdn>.D, <Zdn>.D, #<const>
    ZEOR Z0.S, $1, Z0.S                   // 00004005
    ZEOR Z10.S, $4192256, Z10.S           // 4aa94005
    ZEOR Z0.H, $1, Z0.H                   // 00044005
    ZEOR Z10.H, $63489, Z10.H             // aa2c4005
    ZEOR Z0.B, $1, Z0.B                   // 00064005
    ZEOR Z10.B, $56, Z10.B                // 4a2e4005
    ZEOR Z0.B, $17, Z0.B                  // 00074005
    ZEOR Z10.B, $153, Z10.B               // 2a0f4005
    ZEOR Z0.D, $4294967297, Z0.D          // 00004005
    ZEOR Z10.D, $-8787503089663, Z10.D    // aaaa4005

// EOR     <Zd>.D, <Zn>.D, <Zm>.D
    ZEOR Z0.D, Z0.D, Z0.D       // 0030a004
    ZEOR Z11.D, Z12.D, Z10.D    // 6a31ac04
    ZEOR Z31.D, Z31.D, Z31.D    // ff33bf04

// EORS    <Pd>.B, <Pg>/Z, <Pn>.B, <Pm>.B
    PEORS P0.Z, P0.B, P0.B, P0.B        // 00424025
    PEORS P6.Z, P7.B, P8.B, P5.B        // e55a4825
    PEORS P15.Z, P15.B, P15.B, P15.B    // ef7f4f25

// EORV    <V><d>, <Pg>, <Zn>.<T>
    ZEORV P0, Z0.B, V0      // 00201904
    ZEORV P3, Z12.B, V10    // 8a2d1904
    ZEORV P7, Z31.B, V31    // ff3f1904
    ZEORV P0, Z0.H, V0      // 00205904
    ZEORV P3, Z12.H, V10    // 8a2d5904
    ZEORV P7, Z31.H, V31    // ff3f5904
    ZEORV P0, Z0.S, V0      // 00209904
    ZEORV P3, Z12.S, V10    // 8a2d9904
    ZEORV P7, Z31.S, V31    // ff3f9904
    ZEORV P0, Z0.D, V0      // 0020d904
    ZEORV P3, Z12.D, V10    // 8a2dd904
    ZEORV P7, Z31.D, V31    // ff3fd904

// EXT     <Zdn>.B, <Zdn>.B, <Zm>.B, #<imm>
    ZEXT Z0.B, Z0.B, $0, Z0.B         // 00002005
    ZEXT Z10.B, Z11.B, $85, Z10.B     // 6a152a05
    ZEXT Z31.B, Z31.B, $255, Z31.B    // ff1f3f05

// FABD    <Zdn>.<T>, <Pg>/M, <Zdn>.<T>, <Zm>.<T>
    ZFABD P0.M, Z0.H, Z0.H, Z0.H       // 00804865
    ZFABD P3.M, Z10.H, Z12.H, Z10.H    // 8a8d4865
    ZFABD P7.M, Z31.H, Z31.H, Z31.H    // ff9f4865
    ZFABD P0.M, Z0.S, Z0.S, Z0.S       // 00808865
    ZFABD P3.M, Z10.S, Z12.S, Z10.S    // 8a8d8865
    ZFABD P7.M, Z31.S, Z31.S, Z31.S    // ff9f8865
    ZFABD P0.M, Z0.D, Z0.D, Z0.D       // 0080c865
    ZFABD P3.M, Z10.D, Z12.D, Z10.D    // 8a8dc865
    ZFABD P7.M, Z31.D, Z31.D, Z31.D    // ff9fc865

// FABS    <Zd>.<T>, <Pg>/M, <Zn>.<T>
    ZFABS P0.M, Z0.H, Z0.H      // 00a05c04
    ZFABS P3.M, Z12.H, Z10.H    // 8aad5c04
    ZFABS P7.M, Z31.H, Z31.H    // ffbf5c04
    ZFABS P0.M, Z0.S, Z0.S      // 00a09c04
    ZFABS P3.M, Z12.S, Z10.S    // 8aad9c04
    ZFABS P7.M, Z31.S, Z31.S    // ffbf9c04
    ZFABS P0.M, Z0.D, Z0.D      // 00a0dc04
    ZFABS P3.M, Z12.D, Z10.D    // 8aaddc04
    ZFABS P7.M, Z31.D, Z31.D    // ffbfdc04

// FACGE   <Pd>.<T>, <Pg>/Z, <Zn>.<T>, <Zm>.<T>
    ZFACGE P0.Z, Z0.H, Z0.H, P0.H       // 10c04065
    ZFACGE P3.Z, Z12.H, Z13.H, P5.H     // 95cd4d65
    ZFACGE P7.Z, Z31.H, Z31.H, P15.H    // ffdf5f65
    ZFACGE P0.Z, Z0.S, Z0.S, P0.S       // 10c08065
    ZFACGE P3.Z, Z12.S, Z13.S, P5.S     // 95cd8d65
    ZFACGE P7.Z, Z31.S, Z31.S, P15.S    // ffdf9f65
    ZFACGE P0.Z, Z0.D, Z0.D, P0.D       // 10c0c065
    ZFACGE P3.Z, Z12.D, Z13.D, P5.D     // 95cdcd65
    ZFACGE P7.Z, Z31.D, Z31.D, P15.D    // ffdfdf65

// FACGT   <Pd>.<T>, <Pg>/Z, <Zn>.<T>, <Zm>.<T>
    ZFACGT P0.Z, Z0.H, Z0.H, P0.H       // 10e04065
    ZFACGT P3.Z, Z12.H, Z13.H, P5.H     // 95ed4d65
    ZFACGT P7.Z, Z31.H, Z31.H, P15.H    // ffff5f65
    ZFACGT P0.Z, Z0.S, Z0.S, P0.S       // 10e08065
    ZFACGT P3.Z, Z12.S, Z13.S, P5.S     // 95ed8d65
    ZFACGT P7.Z, Z31.S, Z31.S, P15.S    // ffff9f65
    ZFACGT P0.Z, Z0.D, Z0.D, P0.D       // 10e0c065
    ZFACGT P3.Z, Z12.D, Z13.D, P5.D     // 95edcd65
    ZFACGT P7.Z, Z31.D, Z31.D, P15.D    // ffffdf65

// FADD    <Zdn>.<T>, <Pg>/M, <Zdn>.<T>, <const>
    ZFADD P0.M, Z0.H, $(0.5), Z0.H      // 00805865
    ZFADD P3.M, Z10.H, $(0.5), Z10.H    // 0a8c5865
    ZFADD P7.M, Z31.H, $(1.0), Z31.H    // 3f9c5865
    ZFADD P0.M, Z0.S, $(0.5), Z0.S      // 00809865
    ZFADD P3.M, Z10.S, $(0.5), Z10.S    // 0a8c9865
    ZFADD P7.M, Z31.S, $(1.0), Z31.S    // 3f9c9865
    ZFADD P0.M, Z0.D, $(0.5), Z0.D      // 0080d865
    ZFADD P3.M, Z10.D, $(0.5), Z10.D    // 0a8cd865
    ZFADD P7.M, Z31.D, $(1.0), Z31.D    // 3f9cd865

// FADD    <Zdn>.<T>, <Pg>/M, <Zdn>.<T>, <Zm>.<T>
    ZFADD P0.M, Z0.H, Z0.H, Z0.H       // 00804065
    ZFADD P3.M, Z10.H, Z12.H, Z10.H    // 8a8d4065
    ZFADD P7.M, Z31.H, Z31.H, Z31.H    // ff9f4065
    ZFADD P0.M, Z0.S, Z0.S, Z0.S       // 00808065
    ZFADD P3.M, Z10.S, Z12.S, Z10.S    // 8a8d8065
    ZFADD P7.M, Z31.S, Z31.S, Z31.S    // ff9f8065
    ZFADD P0.M, Z0.D, Z0.D, Z0.D       // 0080c065
    ZFADD P3.M, Z10.D, Z12.D, Z10.D    // 8a8dc065
    ZFADD P7.M, Z31.D, Z31.D, Z31.D    // ff9fc065

// FADD    <Zd>.<T>, <Zn>.<T>, <Zm>.<T>
    ZFADD Z0.H, Z0.H, Z0.H       // 00004065
    ZFADD Z11.H, Z12.H, Z10.H    // 6a014c65
    ZFADD Z31.H, Z31.H, Z31.H    // ff035f65
    ZFADD Z0.S, Z0.S, Z0.S       // 00008065
    ZFADD Z11.S, Z12.S, Z10.S    // 6a018c65
    ZFADD Z31.S, Z31.S, Z31.S    // ff039f65
    ZFADD Z0.D, Z0.D, Z0.D       // 0000c065
    ZFADD Z11.D, Z12.D, Z10.D    // 6a01cc65
    ZFADD Z31.D, Z31.D, Z31.D    // ff03df65

// FADDA   <V><dn>, <Pg>, <V><dn>, <Zm>.<T>
    ZFADDA P0, F0, Z0.H, F0       // 00205865
    ZFADDA P3, F10, Z12.H, F10    // 8a2d5865
    ZFADDA P7, F31, Z31.H, F31    // ff3f5865
    ZFADDA P0, F0, Z0.S, F0       // 00209865
    ZFADDA P3, F10, Z12.S, F10    // 8a2d9865
    ZFADDA P7, F31, Z31.S, F31    // ff3f9865
    ZFADDA P0, F0, Z0.D, F0       // 0020d865
    ZFADDA P3, F10, Z12.D, F10    // 8a2dd865
    ZFADDA P7, F31, Z31.D, F31    // ff3fd865

// FADDV   <V><d>, <Pg>, <Zn>.<T>
    ZFADDV P0, Z0.H, F0      // 00204065
    ZFADDV P3, Z12.H, F10    // 8a2d4065
    ZFADDV P7, Z31.H, F31    // ff3f4065
    ZFADDV P0, Z0.S, F0      // 00208065
    ZFADDV P3, Z12.S, F10    // 8a2d8065
    ZFADDV P7, Z31.S, F31    // ff3f8065
    ZFADDV P0, Z0.D, F0      // 0020c065
    ZFADDV P3, Z12.D, F10    // 8a2dc065
    ZFADDV P7, Z31.D, F31    // ff3fc065

// FCADD   <Zdn>.<T>, <Pg>/M, <Zdn>.<T>, <Zm>.<T>, <const>
    ZFCADD P0.M, Z0.H, Z0.H, $90, Z0.H        // 00804064
    ZFCADD P3.M, Z10.H, Z12.H, $90, Z10.H     // 8a8d4064
    ZFCADD P7.M, Z31.H, Z31.H, $270, Z31.H    // ff9f4164
    ZFCADD P0.M, Z0.S, Z0.S, $90, Z0.S        // 00808064
    ZFCADD P3.M, Z10.S, Z12.S, $90, Z10.S     // 8a8d8064
    ZFCADD P7.M, Z31.S, Z31.S, $270, Z31.S    // ff9f8164
    ZFCADD P0.M, Z0.D, Z0.D, $90, Z0.D        // 0080c064
    ZFCADD P3.M, Z10.D, Z12.D, $90, Z10.D     // 8a8dc064
    ZFCADD P7.M, Z31.D, Z31.D, $270, Z31.D    // ff9fc164

// FCMEQ   <Pd>.<T>, <Pg>/Z, <Zn>.<T>, #0.0
    ZFCMEQ P0.Z, Z0.H, $(0.0), P0.H      // 00205265
    ZFCMEQ P3.Z, Z12.H, $(0.0), P5.H     // 852d5265
    ZFCMEQ P7.Z, Z31.H, $(0.0), P15.H    // ef3f5265
    ZFCMEQ P0.Z, Z0.S, $(0.0), P0.S      // 00209265
    ZFCMEQ P3.Z, Z12.S, $(0.0), P5.S     // 852d9265
    ZFCMEQ P7.Z, Z31.S, $(0.0), P15.S    // ef3f9265
    ZFCMEQ P0.Z, Z0.D, $(0.0), P0.D      // 0020d265
    ZFCMEQ P3.Z, Z12.D, $(0.0), P5.D     // 852dd265
    ZFCMEQ P7.Z, Z31.D, $(0.0), P15.D    // ef3fd265

// FCMEQ   <Pd>.<T>, <Pg>/Z, <Zn>.<T>, <Zm>.<T>
    ZFCMEQ P0.Z, Z0.H, Z0.H, P0.H       // 00604065
    ZFCMEQ P3.Z, Z12.H, Z13.H, P5.H     // 856d4d65
    ZFCMEQ P7.Z, Z31.H, Z31.H, P15.H    // ef7f5f65
    ZFCMEQ P0.Z, Z0.S, Z0.S, P0.S       // 00608065
    ZFCMEQ P3.Z, Z12.S, Z13.S, P5.S     // 856d8d65
    ZFCMEQ P7.Z, Z31.S, Z31.S, P15.S    // ef7f9f65
    ZFCMEQ P0.Z, Z0.D, Z0.D, P0.D       // 0060c065
    ZFCMEQ P3.Z, Z12.D, Z13.D, P5.D     // 856dcd65
    ZFCMEQ P7.Z, Z31.D, Z31.D, P15.D    // ef7fdf65

// FCMGE   <Pd>.<T>, <Pg>/Z, <Zn>.<T>, #0.0
    ZFCMGE P0.Z, Z0.H, $(0.0), P0.H      // 00205065
    ZFCMGE P3.Z, Z12.H, $(0.0), P5.H     // 852d5065
    ZFCMGE P7.Z, Z31.H, $(0.0), P15.H    // ef3f5065
    ZFCMGE P0.Z, Z0.S, $(0.0), P0.S      // 00209065
    ZFCMGE P3.Z, Z12.S, $(0.0), P5.S     // 852d9065
    ZFCMGE P7.Z, Z31.S, $(0.0), P15.S    // ef3f9065
    ZFCMGE P0.Z, Z0.D, $(0.0), P0.D      // 0020d065
    ZFCMGE P3.Z, Z12.D, $(0.0), P5.D     // 852dd065
    ZFCMGE P7.Z, Z31.D, $(0.0), P15.D    // ef3fd065

// FCMGE   <Pd>.<T>, <Pg>/Z, <Zn>.<T>, <Zm>.<T>
    ZFCMGE P0.Z, Z0.H, Z0.H, P0.H       // 00404065
    ZFCMGE P3.Z, Z12.H, Z13.H, P5.H     // 854d4d65
    ZFCMGE P7.Z, Z31.H, Z31.H, P15.H    // ef5f5f65
    ZFCMGE P0.Z, Z0.S, Z0.S, P0.S       // 00408065
    ZFCMGE P3.Z, Z12.S, Z13.S, P5.S     // 854d8d65
    ZFCMGE P7.Z, Z31.S, Z31.S, P15.S    // ef5f9f65
    ZFCMGE P0.Z, Z0.D, Z0.D, P0.D       // 0040c065
    ZFCMGE P3.Z, Z12.D, Z13.D, P5.D     // 854dcd65
    ZFCMGE P7.Z, Z31.D, Z31.D, P15.D    // ef5fdf65

// FCMGT   <Pd>.<T>, <Pg>/Z, <Zn>.<T>, #0.0
    ZFCMGT P0.Z, Z0.H, $(0.0), P0.H      // 10205065
    ZFCMGT P3.Z, Z12.H, $(0.0), P5.H     // 952d5065
    ZFCMGT P7.Z, Z31.H, $(0.0), P15.H    // ff3f5065
    ZFCMGT P0.Z, Z0.S, $(0.0), P0.S      // 10209065
    ZFCMGT P3.Z, Z12.S, $(0.0), P5.S     // 952d9065
    ZFCMGT P7.Z, Z31.S, $(0.0), P15.S    // ff3f9065
    ZFCMGT P0.Z, Z0.D, $(0.0), P0.D      // 1020d065
    ZFCMGT P3.Z, Z12.D, $(0.0), P5.D     // 952dd065
    ZFCMGT P7.Z, Z31.D, $(0.0), P15.D    // ff3fd065

// FCMGT   <Pd>.<T>, <Pg>/Z, <Zn>.<T>, <Zm>.<T>
    ZFCMGT P0.Z, Z0.H, Z0.H, P0.H       // 10404065
    ZFCMGT P3.Z, Z12.H, Z13.H, P5.H     // 954d4d65
    ZFCMGT P7.Z, Z31.H, Z31.H, P15.H    // ff5f5f65
    ZFCMGT P0.Z, Z0.S, Z0.S, P0.S       // 10408065
    ZFCMGT P3.Z, Z12.S, Z13.S, P5.S     // 954d8d65
    ZFCMGT P7.Z, Z31.S, Z31.S, P15.S    // ff5f9f65
    ZFCMGT P0.Z, Z0.D, Z0.D, P0.D       // 1040c065
    ZFCMGT P3.Z, Z12.D, Z13.D, P5.D     // 954dcd65
    ZFCMGT P7.Z, Z31.D, Z31.D, P15.D    // ff5fdf65

// FCMLA   <Zda>.<T>, <Pg>/M, <Zn>.<T>, <Zm>.<T>, <const>
    ZFCMLA P0.M, Z0.H, Z0.H, $0, Z0.H         // 00004064
    ZFCMLA P3.M, Z10.H, Z11.H, $90, Z8.H      // 482d4b64
    ZFCMLA P5.M, Z18.H, Z19.H, $180, Z16.H    // 50565364
    ZFCMLA P7.M, Z31.H, Z31.H, $270, Z31.H    // ff7f5f64
    ZFCMLA P0.M, Z0.S, Z0.S, $0, Z0.S         // 00008064
    ZFCMLA P3.M, Z10.S, Z11.S, $90, Z8.S      // 482d8b64
    ZFCMLA P5.M, Z18.S, Z19.S, $180, Z16.S    // 50569364
    ZFCMLA P7.M, Z31.S, Z31.S, $270, Z31.S    // ff7f9f64
    ZFCMLA P0.M, Z0.D, Z0.D, $0, Z0.D         // 0000c064
    ZFCMLA P3.M, Z10.D, Z11.D, $90, Z8.D      // 482dcb64
    ZFCMLA P5.M, Z18.D, Z19.D, $180, Z16.D    // 5056d364
    ZFCMLA P7.M, Z31.D, Z31.D, $270, Z31.D    // ff7fdf64

// FCMLA   <Zda>.H, <Zn>.H, <Zm>.H[<imm>], <const>
    ZFCMLA Z0.H, Z0.H[0], $0, Z0.H        // 0010a064
    ZFCMLA Z9.H, Z4.H[1], $90, Z8.H       // 2815ac64
    ZFCMLA Z17.H, Z6.H[2], $180, Z16.H    // 301ab664
    ZFCMLA Z31.H, Z7.H[3], $270, Z31.H    // ff1fbf64

// FCMLA   <Zda>.S, <Zn>.S, <Zm>.S[<imm>], <const>
    ZFCMLA Z0.S, Z0.S[0], $0, Z0.S         // 0010e064
    ZFCMLA Z9.S, Z6.S[0], $90, Z8.S        // 2815e664
    ZFCMLA Z17.S, Z10.S[0], $180, Z16.S    // 301aea64
    ZFCMLA Z31.S, Z15.S[1], $270, Z31.S    // ff1fff64

// FCMLE   <Pd>.<T>, <Pg>/Z, <Zn>.<T>, #0.0
    ZFCMLE P0.Z, Z0.H, $(0.0), P0.H      // 10205165
    ZFCMLE P3.Z, Z12.H, $(0.0), P5.H     // 952d5165
    ZFCMLE P7.Z, Z31.H, $(0.0), P15.H    // ff3f5165
    ZFCMLE P0.Z, Z0.S, $(0.0), P0.S      // 10209165
    ZFCMLE P3.Z, Z12.S, $(0.0), P5.S     // 952d9165
    ZFCMLE P7.Z, Z31.S, $(0.0), P15.S    // ff3f9165
    ZFCMLE P0.Z, Z0.D, $(0.0), P0.D      // 1020d165
    ZFCMLE P3.Z, Z12.D, $(0.0), P5.D     // 952dd165
    ZFCMLE P7.Z, Z31.D, $(0.0), P15.D    // ff3fd165

// FCMLT   <Pd>.<T>, <Pg>/Z, <Zn>.<T>, #0.0
    ZFCMLT P0.Z, Z0.H, $(0.0), P0.H      // 00205165
    ZFCMLT P3.Z, Z12.H, $(0.0), P5.H     // 852d5165
    ZFCMLT P7.Z, Z31.H, $(0.0), P15.H    // ef3f5165
    ZFCMLT P0.Z, Z0.S, $(0.0), P0.S      // 00209165
    ZFCMLT P3.Z, Z12.S, $(0.0), P5.S     // 852d9165
    ZFCMLT P7.Z, Z31.S, $(0.0), P15.S    // ef3f9165
    ZFCMLT P0.Z, Z0.D, $(0.0), P0.D      // 0020d165
    ZFCMLT P3.Z, Z12.D, $(0.0), P5.D     // 852dd165
    ZFCMLT P7.Z, Z31.D, $(0.0), P15.D    // ef3fd165

// FCMNE   <Pd>.<T>, <Pg>/Z, <Zn>.<T>, #0.0
    ZFCMNE P0.Z, Z0.H, $(0.0), P0.H      // 00205365
    ZFCMNE P3.Z, Z12.H, $(0.0), P5.H     // 852d5365
    ZFCMNE P7.Z, Z31.H, $(0.0), P15.H    // ef3f5365
    ZFCMNE P0.Z, Z0.S, $(0.0), P0.S      // 00209365
    ZFCMNE P3.Z, Z12.S, $(0.0), P5.S     // 852d9365
    ZFCMNE P7.Z, Z31.S, $(0.0), P15.S    // ef3f9365
    ZFCMNE P0.Z, Z0.D, $(0.0), P0.D      // 0020d365
    ZFCMNE P3.Z, Z12.D, $(0.0), P5.D     // 852dd365
    ZFCMNE P7.Z, Z31.D, $(0.0), P15.D    // ef3fd365

// FCMNE   <Pd>.<T>, <Pg>/Z, <Zn>.<T>, <Zm>.<T>
    ZFCMNE P0.Z, Z0.H, Z0.H, P0.H       // 10604065
    ZFCMNE P3.Z, Z12.H, Z13.H, P5.H     // 956d4d65
    ZFCMNE P7.Z, Z31.H, Z31.H, P15.H    // ff7f5f65
    ZFCMNE P0.Z, Z0.S, Z0.S, P0.S       // 10608065
    ZFCMNE P3.Z, Z12.S, Z13.S, P5.S     // 956d8d65
    ZFCMNE P7.Z, Z31.S, Z31.S, P15.S    // ff7f9f65
    ZFCMNE P0.Z, Z0.D, Z0.D, P0.D       // 1060c065
    ZFCMNE P3.Z, Z12.D, Z13.D, P5.D     // 956dcd65
    ZFCMNE P7.Z, Z31.D, Z31.D, P15.D    // ff7fdf65

// FCMUO   <Pd>.<T>, <Pg>/Z, <Zn>.<T>, <Zm>.<T>
    ZFCMUO P0.Z, Z0.H, Z0.H, P0.H       // 00c04065
    ZFCMUO P3.Z, Z12.H, Z13.H, P5.H     // 85cd4d65
    ZFCMUO P7.Z, Z31.H, Z31.H, P15.H    // efdf5f65
    ZFCMUO P0.Z, Z0.S, Z0.S, P0.S       // 00c08065
    ZFCMUO P3.Z, Z12.S, Z13.S, P5.S     // 85cd8d65
    ZFCMUO P7.Z, Z31.S, Z31.S, P15.S    // efdf9f65
    ZFCMUO P0.Z, Z0.D, Z0.D, P0.D       // 00c0c065
    ZFCMUO P3.Z, Z12.D, Z13.D, P5.D     // 85cdcd65
    ZFCMUO P7.Z, Z31.D, Z31.D, P15.D    // efdfdf65

// FCPY    <Zd>.<T>, <Pg>/M, #<const>
    ZFCPY P0.M, $(-2.0), Z0.H         // 00d05005
    ZFCPY P6.M, $(-0.40625), Z10.H    // 4adb5605
    ZFCPY P15.M, $(1.9375), Z31.H     // ffcf5f05
    ZFCPY P0.M, $(-2.0), Z0.S         // 00d09005
    ZFCPY P6.M, $(-0.40625), Z10.S    // 4adb9605
    ZFCPY P15.M, $(1.9375), Z31.S     // ffcf9f05
    ZFCPY P0.M, $(-2.0), Z0.D         // 00d0d005
    ZFCPY P6.M, $(-0.40625), Z10.D    // 4adbd605
    ZFCPY P15.M, $(1.9375), Z31.D     // ffcfdf05

// FCVT    <Zd>.H, <Pg>/M, <Zn>.D
    ZFCVT P0.M, Z0.D, Z0.H      // 00a0c865
    ZFCVT P3.M, Z12.D, Z10.H    // 8aadc865
    ZFCVT P7.M, Z31.D, Z31.H    // ffbfc865

// FCVT    <Zd>.S, <Pg>/M, <Zn>.D
    ZFCVT P0.M, Z0.D, Z0.S      // 00a0ca65
    ZFCVT P3.M, Z12.D, Z10.S    // 8aadca65
    ZFCVT P7.M, Z31.D, Z31.S    // ffbfca65

// FCVT    <Zd>.D, <Pg>/M, <Zn>.H
    ZFCVT P0.M, Z0.H, Z0.D      // 00a0c965
    ZFCVT P3.M, Z12.H, Z10.D    // 8aadc965
    ZFCVT P7.M, Z31.H, Z31.D    // ffbfc965

// FCVT    <Zd>.S, <Pg>/M, <Zn>.H
    ZFCVT P0.M, Z0.H, Z0.S      // 00a08965
    ZFCVT P3.M, Z12.H, Z10.S    // 8aad8965
    ZFCVT P7.M, Z31.H, Z31.S    // ffbf8965

// FCVT    <Zd>.D, <Pg>/M, <Zn>.S
    ZFCVT P0.M, Z0.S, Z0.D      // 00a0cb65
    ZFCVT P3.M, Z12.S, Z10.D    // 8aadcb65
    ZFCVT P7.M, Z31.S, Z31.D    // ffbfcb65

// FCVT    <Zd>.H, <Pg>/M, <Zn>.S
    ZFCVT P0.M, Z0.S, Z0.H      // 00a08865
    ZFCVT P3.M, Z12.S, Z10.H    // 8aad8865
    ZFCVT P7.M, Z31.S, Z31.H    // ffbf8865

// FCVTZS  <Zd>.S, <Pg>/M, <Zn>.D
    ZFCVTZS P0.M, Z0.D, Z0.S      // 00a0d865
    ZFCVTZS P3.M, Z12.D, Z10.S    // 8aadd865
    ZFCVTZS P7.M, Z31.D, Z31.S    // ffbfd865

// FCVTZS  <Zd>.D, <Pg>/M, <Zn>.D
    ZFCVTZS P0.M, Z0.D, Z0.D      // 00a0de65
    ZFCVTZS P3.M, Z12.D, Z10.D    // 8aadde65
    ZFCVTZS P7.M, Z31.D, Z31.D    // ffbfde65

// FCVTZS  <Zd>.H, <Pg>/M, <Zn>.H
    ZFCVTZS P0.M, Z0.H, Z0.H      // 00a05a65
    ZFCVTZS P3.M, Z12.H, Z10.H    // 8aad5a65
    ZFCVTZS P7.M, Z31.H, Z31.H    // ffbf5a65

// FCVTZS  <Zd>.S, <Pg>/M, <Zn>.H
    ZFCVTZS P0.M, Z0.H, Z0.S      // 00a05c65
    ZFCVTZS P3.M, Z12.H, Z10.S    // 8aad5c65
    ZFCVTZS P7.M, Z31.H, Z31.S    // ffbf5c65

// FCVTZS  <Zd>.D, <Pg>/M, <Zn>.H
    ZFCVTZS P0.M, Z0.H, Z0.D      // 00a05e65
    ZFCVTZS P3.M, Z12.H, Z10.D    // 8aad5e65
    ZFCVTZS P7.M, Z31.H, Z31.D    // ffbf5e65

// FCVTZS  <Zd>.S, <Pg>/M, <Zn>.S
    ZFCVTZS P0.M, Z0.S, Z0.S      // 00a09c65
    ZFCVTZS P3.M, Z12.S, Z10.S    // 8aad9c65
    ZFCVTZS P7.M, Z31.S, Z31.S    // ffbf9c65

// FCVTZS  <Zd>.D, <Pg>/M, <Zn>.S
    ZFCVTZS P0.M, Z0.S, Z0.D      // 00a0dc65
    ZFCVTZS P3.M, Z12.S, Z10.D    // 8aaddc65
    ZFCVTZS P7.M, Z31.S, Z31.D    // ffbfdc65

// FCVTZU  <Zd>.S, <Pg>/M, <Zn>.D
    ZFCVTZU P0.M, Z0.D, Z0.S      // 00a0d965
    ZFCVTZU P3.M, Z12.D, Z10.S    // 8aadd965
    ZFCVTZU P7.M, Z31.D, Z31.S    // ffbfd965

// FCVTZU  <Zd>.D, <Pg>/M, <Zn>.D
    ZFCVTZU P0.M, Z0.D, Z0.D      // 00a0df65
    ZFCVTZU P3.M, Z12.D, Z10.D    // 8aaddf65
    ZFCVTZU P7.M, Z31.D, Z31.D    // ffbfdf65

// FCVTZU  <Zd>.H, <Pg>/M, <Zn>.H
    ZFCVTZU P0.M, Z0.H, Z0.H      // 00a05b65
    ZFCVTZU P3.M, Z12.H, Z10.H    // 8aad5b65
    ZFCVTZU P7.M, Z31.H, Z31.H    // ffbf5b65

// FCVTZU  <Zd>.S, <Pg>/M, <Zn>.H
    ZFCVTZU P0.M, Z0.H, Z0.S      // 00a05d65
    ZFCVTZU P3.M, Z12.H, Z10.S    // 8aad5d65
    ZFCVTZU P7.M, Z31.H, Z31.S    // ffbf5d65

// FCVTZU  <Zd>.D, <Pg>/M, <Zn>.H
    ZFCVTZU P0.M, Z0.H, Z0.D      // 00a05f65
    ZFCVTZU P3.M, Z12.H, Z10.D    // 8aad5f65
    ZFCVTZU P7.M, Z31.H, Z31.D    // ffbf5f65

// FCVTZU  <Zd>.S, <Pg>/M, <Zn>.S
    ZFCVTZU P0.M, Z0.S, Z0.S      // 00a09d65
    ZFCVTZU P3.M, Z12.S, Z10.S    // 8aad9d65
    ZFCVTZU P7.M, Z31.S, Z31.S    // ffbf9d65

// FCVTZU  <Zd>.D, <Pg>/M, <Zn>.S
    ZFCVTZU P0.M, Z0.S, Z0.D      // 00a0dd65
    ZFCVTZU P3.M, Z12.S, Z10.D    // 8aaddd65
    ZFCVTZU P7.M, Z31.S, Z31.D    // ffbfdd65

// FDIV    <Zdn>.<T>, <Pg>/M, <Zdn>.<T>, <Zm>.<T>
    ZFDIV P0.M, Z0.H, Z0.H, Z0.H       // 00804d65
    ZFDIV P3.M, Z10.H, Z12.H, Z10.H    // 8a8d4d65
    ZFDIV P7.M, Z31.H, Z31.H, Z31.H    // ff9f4d65
    ZFDIV P0.M, Z0.S, Z0.S, Z0.S       // 00808d65
    ZFDIV P3.M, Z10.S, Z12.S, Z10.S    // 8a8d8d65
    ZFDIV P7.M, Z31.S, Z31.S, Z31.S    // ff9f8d65
    ZFDIV P0.M, Z0.D, Z0.D, Z0.D       // 0080cd65
    ZFDIV P3.M, Z10.D, Z12.D, Z10.D    // 8a8dcd65
    ZFDIV P7.M, Z31.D, Z31.D, Z31.D    // ff9fcd65

// FDIVR   <Zdn>.<T>, <Pg>/M, <Zdn>.<T>, <Zm>.<T>
    ZFDIVR P0.M, Z0.H, Z0.H, Z0.H       // 00804c65
    ZFDIVR P3.M, Z10.H, Z12.H, Z10.H    // 8a8d4c65
    ZFDIVR P7.M, Z31.H, Z31.H, Z31.H    // ff9f4c65
    ZFDIVR P0.M, Z0.S, Z0.S, Z0.S       // 00808c65
    ZFDIVR P3.M, Z10.S, Z12.S, Z10.S    // 8a8d8c65
    ZFDIVR P7.M, Z31.S, Z31.S, Z31.S    // ff9f8c65
    ZFDIVR P0.M, Z0.D, Z0.D, Z0.D       // 0080cc65
    ZFDIVR P3.M, Z10.D, Z12.D, Z10.D    // 8a8dcc65
    ZFDIVR P7.M, Z31.D, Z31.D, Z31.D    // ff9fcc65

// FDUP    <Zd>.<T>, #<const>
    ZFDUP $(-2.0), Z0.H         // 00d07925
    ZFDUP $(-0.40625), Z10.H    // 4adb7925
    ZFDUP $(1.9375), Z31.H      // ffcf7925
    ZFDUP $(-2.0), Z0.S         // 00d0b925
    ZFDUP $(-0.40625), Z10.S    // 4adbb925
    ZFDUP $(1.9375), Z31.S      // ffcfb925
    ZFDUP $(-2.0), Z0.D         // 00d0f925
    ZFDUP $(-0.40625), Z10.D    // 4adbf925
    ZFDUP $(1.9375), Z31.D      // ffcff925

// FEXPA   <Zd>.<T>, <Zn>.<T>
    ZFEXPA Z0.H, Z0.H      // 00b86004
    ZFEXPA Z11.H, Z10.H    // 6ab96004
    ZFEXPA Z31.H, Z31.H    // ffbb6004
    ZFEXPA Z0.S, Z0.S      // 00b8a004
    ZFEXPA Z11.S, Z10.S    // 6ab9a004
    ZFEXPA Z31.S, Z31.S    // ffbba004
    ZFEXPA Z0.D, Z0.D      // 00b8e004
    ZFEXPA Z11.D, Z10.D    // 6ab9e004
    ZFEXPA Z31.D, Z31.D    // ffbbe004

// FMAD    <Zdn>.<T>, <Pg>/M, <Zm>.<T>, <Za>.<T>
    ZFMAD P0.M, Z0.H, Z0.H, Z0.H       // 00806065
    ZFMAD P3.M, Z12.H, Z13.H, Z10.H    // 8a8d6d65
    ZFMAD P7.M, Z31.H, Z31.H, Z31.H    // ff9f7f65
    ZFMAD P0.M, Z0.S, Z0.S, Z0.S       // 0080a065
    ZFMAD P3.M, Z12.S, Z13.S, Z10.S    // 8a8dad65
    ZFMAD P7.M, Z31.S, Z31.S, Z31.S    // ff9fbf65
    ZFMAD P0.M, Z0.D, Z0.D, Z0.D       // 0080e065
    ZFMAD P3.M, Z12.D, Z13.D, Z10.D    // 8a8ded65
    ZFMAD P7.M, Z31.D, Z31.D, Z31.D    // ff9fff65

// FMAX    <Zdn>.<T>, <Pg>/M, <Zdn>.<T>, <const>
    ZFMAX P0.M, Z0.H, $(0.0), Z0.H      // 00805e65
    ZFMAX P3.M, Z10.H, $(0.0), Z10.H    // 0a8c5e65
    ZFMAX P7.M, Z31.H, $(1.0), Z31.H    // 3f9c5e65
    ZFMAX P0.M, Z0.S, $(0.0), Z0.S      // 00809e65
    ZFMAX P3.M, Z10.S, $(0.0), Z10.S    // 0a8c9e65
    ZFMAX P7.M, Z31.S, $(1.0), Z31.S    // 3f9c9e65
    ZFMAX P0.M, Z0.D, $(0.0), Z0.D      // 0080de65
    ZFMAX P3.M, Z10.D, $(0.0), Z10.D    // 0a8cde65
    ZFMAX P7.M, Z31.D, $(1.0), Z31.D    // 3f9cde65

// FMAX    <Zdn>.<T>, <Pg>/M, <Zdn>.<T>, <Zm>.<T>
    ZFMAX P0.M, Z0.H, Z0.H, Z0.H       // 00804665
    ZFMAX P3.M, Z10.H, Z12.H, Z10.H    // 8a8d4665
    ZFMAX P7.M, Z31.H, Z31.H, Z31.H    // ff9f4665
    ZFMAX P0.M, Z0.S, Z0.S, Z0.S       // 00808665
    ZFMAX P3.M, Z10.S, Z12.S, Z10.S    // 8a8d8665
    ZFMAX P7.M, Z31.S, Z31.S, Z31.S    // ff9f8665
    ZFMAX P0.M, Z0.D, Z0.D, Z0.D       // 0080c665
    ZFMAX P3.M, Z10.D, Z12.D, Z10.D    // 8a8dc665
    ZFMAX P7.M, Z31.D, Z31.D, Z31.D    // ff9fc665

// FMAXNM  <Zdn>.<T>, <Pg>/M, <Zdn>.<T>, <const>
    ZFMAXNM P0.M, Z0.H, $(0.0), Z0.H      // 00805c65
    ZFMAXNM P3.M, Z10.H, $(0.0), Z10.H    // 0a8c5c65
    ZFMAXNM P7.M, Z31.H, $(1.0), Z31.H    // 3f9c5c65
    ZFMAXNM P0.M, Z0.S, $(0.0), Z0.S      // 00809c65
    ZFMAXNM P3.M, Z10.S, $(0.0), Z10.S    // 0a8c9c65
    ZFMAXNM P7.M, Z31.S, $(1.0), Z31.S    // 3f9c9c65
    ZFMAXNM P0.M, Z0.D, $(0.0), Z0.D      // 0080dc65
    ZFMAXNM P3.M, Z10.D, $(0.0), Z10.D    // 0a8cdc65
    ZFMAXNM P7.M, Z31.D, $(1.0), Z31.D    // 3f9cdc65

// FMAXNM  <Zdn>.<T>, <Pg>/M, <Zdn>.<T>, <Zm>.<T>
    ZFMAXNM P0.M, Z0.H, Z0.H, Z0.H       // 00804465
    ZFMAXNM P3.M, Z10.H, Z12.H, Z10.H    // 8a8d4465
    ZFMAXNM P7.M, Z31.H, Z31.H, Z31.H    // ff9f4465
    ZFMAXNM P0.M, Z0.S, Z0.S, Z0.S       // 00808465
    ZFMAXNM P3.M, Z10.S, Z12.S, Z10.S    // 8a8d8465
    ZFMAXNM P7.M, Z31.S, Z31.S, Z31.S    // ff9f8465
    ZFMAXNM P0.M, Z0.D, Z0.D, Z0.D       // 0080c465
    ZFMAXNM P3.M, Z10.D, Z12.D, Z10.D    // 8a8dc465
    ZFMAXNM P7.M, Z31.D, Z31.D, Z31.D    // ff9fc465

// FMAXNMV <V><d>, <Pg>, <Zn>.<T>
    ZFMAXNMV P0, Z0.H, F0      // 00204465
    ZFMAXNMV P3, Z12.H, F10    // 8a2d4465
    ZFMAXNMV P7, Z31.H, F31    // ff3f4465
    ZFMAXNMV P0, Z0.S, F0      // 00208465
    ZFMAXNMV P3, Z12.S, F10    // 8a2d8465
    ZFMAXNMV P7, Z31.S, F31    // ff3f8465
    ZFMAXNMV P0, Z0.D, F0      // 0020c465
    ZFMAXNMV P3, Z12.D, F10    // 8a2dc465
    ZFMAXNMV P7, Z31.D, F31    // ff3fc465

// FMAXV   <V><d>, <Pg>, <Zn>.<T>
    ZFMAXV P0, Z0.H, F0      // 00204665
    ZFMAXV P3, Z12.H, F10    // 8a2d4665
    ZFMAXV P7, Z31.H, F31    // ff3f4665
    ZFMAXV P0, Z0.S, F0      // 00208665
    ZFMAXV P3, Z12.S, F10    // 8a2d8665
    ZFMAXV P7, Z31.S, F31    // ff3f8665
    ZFMAXV P0, Z0.D, F0      // 0020c665
    ZFMAXV P3, Z12.D, F10    // 8a2dc665
    ZFMAXV P7, Z31.D, F31    // ff3fc665

// FMIN    <Zdn>.<T>, <Pg>/M, <Zdn>.<T>, <const>
    ZFMIN P0.M, Z0.H, $(0.0), Z0.H      // 00805f65
    ZFMIN P3.M, Z10.H, $(0.0), Z10.H    // 0a8c5f65
    ZFMIN P7.M, Z31.H, $(1.0), Z31.H    // 3f9c5f65
    ZFMIN P0.M, Z0.S, $(0.0), Z0.S      // 00809f65
    ZFMIN P3.M, Z10.S, $(0.0), Z10.S    // 0a8c9f65
    ZFMIN P7.M, Z31.S, $(1.0), Z31.S    // 3f9c9f65
    ZFMIN P0.M, Z0.D, $(0.0), Z0.D      // 0080df65
    ZFMIN P3.M, Z10.D, $(0.0), Z10.D    // 0a8cdf65
    ZFMIN P7.M, Z31.D, $(1.0), Z31.D    // 3f9cdf65

// FMIN    <Zdn>.<T>, <Pg>/M, <Zdn>.<T>, <Zm>.<T>
    ZFMIN P0.M, Z0.H, Z0.H, Z0.H       // 00804765
    ZFMIN P3.M, Z10.H, Z12.H, Z10.H    // 8a8d4765
    ZFMIN P7.M, Z31.H, Z31.H, Z31.H    // ff9f4765
    ZFMIN P0.M, Z0.S, Z0.S, Z0.S       // 00808765
    ZFMIN P3.M, Z10.S, Z12.S, Z10.S    // 8a8d8765
    ZFMIN P7.M, Z31.S, Z31.S, Z31.S    // ff9f8765
    ZFMIN P0.M, Z0.D, Z0.D, Z0.D       // 0080c765
    ZFMIN P3.M, Z10.D, Z12.D, Z10.D    // 8a8dc765
    ZFMIN P7.M, Z31.D, Z31.D, Z31.D    // ff9fc765

// FMINNM  <Zdn>.<T>, <Pg>/M, <Zdn>.<T>, <const>
    ZFMINNM P0.M, Z0.H, $(0.0), Z0.H      // 00805d65
    ZFMINNM P3.M, Z10.H, $(0.0), Z10.H    // 0a8c5d65
    ZFMINNM P7.M, Z31.H, $(1.0), Z31.H    // 3f9c5d65
    ZFMINNM P0.M, Z0.S, $(0.0), Z0.S      // 00809d65
    ZFMINNM P3.M, Z10.S, $(0.0), Z10.S    // 0a8c9d65
    ZFMINNM P7.M, Z31.S, $(1.0), Z31.S    // 3f9c9d65
    ZFMINNM P0.M, Z0.D, $(0.0), Z0.D      // 0080dd65
    ZFMINNM P3.M, Z10.D, $(0.0), Z10.D    // 0a8cdd65
    ZFMINNM P7.M, Z31.D, $(1.0), Z31.D    // 3f9cdd65

// FMINNM  <Zdn>.<T>, <Pg>/M, <Zdn>.<T>, <Zm>.<T>
    ZFMINNM P0.M, Z0.H, Z0.H, Z0.H       // 00804565
    ZFMINNM P3.M, Z10.H, Z12.H, Z10.H    // 8a8d4565
    ZFMINNM P7.M, Z31.H, Z31.H, Z31.H    // ff9f4565
    ZFMINNM P0.M, Z0.S, Z0.S, Z0.S       // 00808565
    ZFMINNM P3.M, Z10.S, Z12.S, Z10.S    // 8a8d8565
    ZFMINNM P7.M, Z31.S, Z31.S, Z31.S    // ff9f8565
    ZFMINNM P0.M, Z0.D, Z0.D, Z0.D       // 0080c565
    ZFMINNM P3.M, Z10.D, Z12.D, Z10.D    // 8a8dc565
    ZFMINNM P7.M, Z31.D, Z31.D, Z31.D    // ff9fc565

// FMINNMV <V><d>, <Pg>, <Zn>.<T>
    ZFMINNMV P0, Z0.H, F0      // 00204565
    ZFMINNMV P3, Z12.H, F10    // 8a2d4565
    ZFMINNMV P7, Z31.H, F31    // ff3f4565
    ZFMINNMV P0, Z0.S, F0      // 00208565
    ZFMINNMV P3, Z12.S, F10    // 8a2d8565
    ZFMINNMV P7, Z31.S, F31    // ff3f8565
    ZFMINNMV P0, Z0.D, F0      // 0020c565
    ZFMINNMV P3, Z12.D, F10    // 8a2dc565
    ZFMINNMV P7, Z31.D, F31    // ff3fc565

// FMINV   <V><d>, <Pg>, <Zn>.<T>
    ZFMINV P0, Z0.H, F0      // 00204765
    ZFMINV P3, Z12.H, F10    // 8a2d4765
    ZFMINV P7, Z31.H, F31    // ff3f4765
    ZFMINV P0, Z0.S, F0      // 00208765
    ZFMINV P3, Z12.S, F10    // 8a2d8765
    ZFMINV P7, Z31.S, F31    // ff3f8765
    ZFMINV P0, Z0.D, F0      // 0020c765
    ZFMINV P3, Z12.D, F10    // 8a2dc765
    ZFMINV P7, Z31.D, F31    // ff3fc765

// FMLA    <Zda>.<T>, <Pg>/M, <Zn>.<T>, <Zm>.<T>
    ZFMLA P0.M, Z0.H, Z0.H, Z0.H       // 00006065
    ZFMLA P3.M, Z12.H, Z13.H, Z10.H    // 8a0d6d65
    ZFMLA P7.M, Z31.H, Z31.H, Z31.H    // ff1f7f65
    ZFMLA P0.M, Z0.S, Z0.S, Z0.S       // 0000a065
    ZFMLA P3.M, Z12.S, Z13.S, Z10.S    // 8a0dad65
    ZFMLA P7.M, Z31.S, Z31.S, Z31.S    // ff1fbf65
    ZFMLA P0.M, Z0.D, Z0.D, Z0.D       // 0000e065
    ZFMLA P3.M, Z12.D, Z13.D, Z10.D    // 8a0ded65
    ZFMLA P7.M, Z31.D, Z31.D, Z31.D    // ff1fff65

// FMLA    <Zda>.D, <Zn>.D, <Zm>.D[<imm>]
    ZFMLA Z0.D, Z0.D[0], Z0.D       // 0000e064
    ZFMLA Z11.D, Z7.D[0], Z10.D     // 6a01e764
    ZFMLA Z31.D, Z15.D[1], Z31.D    // ff03ff64

// FMLA    <Zda>.H, <Zn>.H, <Zm>.H[<imm>]
    ZFMLA Z0.H, Z0.H[0], Z0.H      // 00002064
    ZFMLA Z11.H, Z4.H[2], Z10.H    // 6a013464
    ZFMLA Z31.H, Z7.H[7], Z31.H    // ff037f64

// FMLA    <Zda>.S, <Zn>.S, <Zm>.S[<imm>]
    ZFMLA Z0.S, Z0.S[0], Z0.S      // 0000a064
    ZFMLA Z11.S, Z4.S[1], Z10.S    // 6a01ac64
    ZFMLA Z31.S, Z7.S[3], Z31.S    // ff03bf64

// FMLS    <Zda>.<T>, <Pg>/M, <Zn>.<T>, <Zm>.<T>
    ZFMLS P0.M, Z0.H, Z0.H, Z0.H       // 00206065
    ZFMLS P3.M, Z12.H, Z13.H, Z10.H    // 8a2d6d65
    ZFMLS P7.M, Z31.H, Z31.H, Z31.H    // ff3f7f65
    ZFMLS P0.M, Z0.S, Z0.S, Z0.S       // 0020a065
    ZFMLS P3.M, Z12.S, Z13.S, Z10.S    // 8a2dad65
    ZFMLS P7.M, Z31.S, Z31.S, Z31.S    // ff3fbf65
    ZFMLS P0.M, Z0.D, Z0.D, Z0.D       // 0020e065
    ZFMLS P3.M, Z12.D, Z13.D, Z10.D    // 8a2ded65
    ZFMLS P7.M, Z31.D, Z31.D, Z31.D    // ff3fff65

// FMLS    <Zda>.D, <Zn>.D, <Zm>.D[<imm>]
    ZFMLS Z0.D, Z0.D[0], Z0.D       // 0004e064
    ZFMLS Z11.D, Z7.D[0], Z10.D     // 6a05e764
    ZFMLS Z31.D, Z15.D[1], Z31.D    // ff07ff64

// FMLS    <Zda>.H, <Zn>.H, <Zm>.H[<imm>]
    ZFMLS Z0.H, Z0.H[0], Z0.H      // 00042064
    ZFMLS Z11.H, Z4.H[2], Z10.H    // 6a053464
    ZFMLS Z31.H, Z7.H[7], Z31.H    // ff077f64

// FMLS    <Zda>.S, <Zn>.S, <Zm>.S[<imm>]
    ZFMLS Z0.S, Z0.S[0], Z0.S      // 0004a064
    ZFMLS Z11.S, Z4.S[1], Z10.S    // 6a05ac64
    ZFMLS Z31.S, Z7.S[3], Z31.S    // ff07bf64

// FMMLA   <Zda>.D, <Zn>.D, <Zm>.D
    ZFMMLA Z0.D, Z0.D, Z0.D       // 00e4e064
    ZFMMLA Z11.D, Z12.D, Z10.D    // 6ae5ec64
    ZFMMLA Z31.D, Z31.D, Z31.D    // ffe7ff64

// FMMLA   <Zda>.S, <Zn>.S, <Zm>.S
    ZFMMLA Z0.S, Z0.S, Z0.S       // 00e4a064
    ZFMMLA Z11.S, Z12.S, Z10.S    // 6ae5ac64
    ZFMMLA Z31.S, Z31.S, Z31.S    // ffe7bf64

// FMSB    <Zdn>.<T>, <Pg>/M, <Zm>.<T>, <Za>.<T>
    ZFMSB P0.M, Z0.H, Z0.H, Z0.H       // 00a06065
    ZFMSB P3.M, Z12.H, Z13.H, Z10.H    // 8aad6d65
    ZFMSB P7.M, Z31.H, Z31.H, Z31.H    // ffbf7f65
    ZFMSB P0.M, Z0.S, Z0.S, Z0.S       // 00a0a065
    ZFMSB P3.M, Z12.S, Z13.S, Z10.S    // 8aadad65
    ZFMSB P7.M, Z31.S, Z31.S, Z31.S    // ffbfbf65
    ZFMSB P0.M, Z0.D, Z0.D, Z0.D       // 00a0e065
    ZFMSB P3.M, Z12.D, Z13.D, Z10.D    // 8aaded65
    ZFMSB P7.M, Z31.D, Z31.D, Z31.D    // ffbfff65

// FMUL    <Zdn>.<T>, <Pg>/M, <Zdn>.<T>, <const>
    ZFMUL P0.M, Z0.H, $(0.5), Z0.H      // 00805a65
    ZFMUL P3.M, Z10.H, $(0.5), Z10.H    // 0a8c5a65
    ZFMUL P7.M, Z31.H, $(2.0), Z31.H    // 3f9c5a65
    ZFMUL P0.M, Z0.S, $(0.5), Z0.S      // 00809a65
    ZFMUL P3.M, Z10.S, $(0.5), Z10.S    // 0a8c9a65
    ZFMUL P7.M, Z31.S, $(2.0), Z31.S    // 3f9c9a65
    ZFMUL P0.M, Z0.D, $(0.5), Z0.D      // 0080da65
    ZFMUL P3.M, Z10.D, $(0.5), Z10.D    // 0a8cda65
    ZFMUL P7.M, Z31.D, $(2.0), Z31.D    // 3f9cda65

// FMUL    <Zdn>.<T>, <Pg>/M, <Zdn>.<T>, <Zm>.<T>
    ZFMUL P0.M, Z0.H, Z0.H, Z0.H       // 00804265
    ZFMUL P3.M, Z10.H, Z12.H, Z10.H    // 8a8d4265
    ZFMUL P7.M, Z31.H, Z31.H, Z31.H    // ff9f4265
    ZFMUL P0.M, Z0.S, Z0.S, Z0.S       // 00808265
    ZFMUL P3.M, Z10.S, Z12.S, Z10.S    // 8a8d8265
    ZFMUL P7.M, Z31.S, Z31.S, Z31.S    // ff9f8265
    ZFMUL P0.M, Z0.D, Z0.D, Z0.D       // 0080c265
    ZFMUL P3.M, Z10.D, Z12.D, Z10.D    // 8a8dc265
    ZFMUL P7.M, Z31.D, Z31.D, Z31.D    // ff9fc265

// FMUL    <Zd>.<T>, <Zn>.<T>, <Zm>.<T>
    ZFMUL Z0.H, Z0.H, Z0.H       // 00084065
    ZFMUL Z11.H, Z12.H, Z10.H    // 6a094c65
    ZFMUL Z31.H, Z31.H, Z31.H    // ff0b5f65
    ZFMUL Z0.S, Z0.S, Z0.S       // 00088065
    ZFMUL Z11.S, Z12.S, Z10.S    // 6a098c65
    ZFMUL Z31.S, Z31.S, Z31.S    // ff0b9f65
    ZFMUL Z0.D, Z0.D, Z0.D       // 0008c065
    ZFMUL Z11.D, Z12.D, Z10.D    // 6a09cc65
    ZFMUL Z31.D, Z31.D, Z31.D    // ff0bdf65

// FMUL    <Zd>.D, <Zn>.D, <Zm>.D[<imm>]
    ZFMUL Z0.D, Z0.D[0], Z0.D       // 0020e064
    ZFMUL Z11.D, Z7.D[0], Z10.D     // 6a21e764
    ZFMUL Z31.D, Z15.D[1], Z31.D    // ff23ff64

// FMUL    <Zd>.H, <Zn>.H, <Zm>.H[<imm>]
    ZFMUL Z0.H, Z0.H[0], Z0.H      // 00202064
    ZFMUL Z11.H, Z4.H[2], Z10.H    // 6a213464
    ZFMUL Z31.H, Z7.H[7], Z31.H    // ff237f64

// FMUL    <Zd>.S, <Zn>.S, <Zm>.S[<imm>]
    ZFMUL Z0.S, Z0.S[0], Z0.S      // 0020a064
    ZFMUL Z11.S, Z4.S[1], Z10.S    // 6a21ac64
    ZFMUL Z31.S, Z7.S[3], Z31.S    // ff23bf64

// FMULX   <Zdn>.<T>, <Pg>/M, <Zdn>.<T>, <Zm>.<T>
    ZFMULX P0.M, Z0.H, Z0.H, Z0.H       // 00804a65
    ZFMULX P3.M, Z10.H, Z12.H, Z10.H    // 8a8d4a65
    ZFMULX P7.M, Z31.H, Z31.H, Z31.H    // ff9f4a65
    ZFMULX P0.M, Z0.S, Z0.S, Z0.S       // 00808a65
    ZFMULX P3.M, Z10.S, Z12.S, Z10.S    // 8a8d8a65
    ZFMULX P7.M, Z31.S, Z31.S, Z31.S    // ff9f8a65
    ZFMULX P0.M, Z0.D, Z0.D, Z0.D       // 0080ca65
    ZFMULX P3.M, Z10.D, Z12.D, Z10.D    // 8a8dca65
    ZFMULX P7.M, Z31.D, Z31.D, Z31.D    // ff9fca65

// FNEG    <Zd>.<T>, <Pg>/M, <Zn>.<T>
    ZFNEG P0.M, Z0.H, Z0.H      // 00a05d04
    ZFNEG P3.M, Z12.H, Z10.H    // 8aad5d04
    ZFNEG P7.M, Z31.H, Z31.H    // ffbf5d04
    ZFNEG P0.M, Z0.S, Z0.S      // 00a09d04
    ZFNEG P3.M, Z12.S, Z10.S    // 8aad9d04
    ZFNEG P7.M, Z31.S, Z31.S    // ffbf9d04
    ZFNEG P0.M, Z0.D, Z0.D      // 00a0dd04
    ZFNEG P3.M, Z12.D, Z10.D    // 8aaddd04
    ZFNEG P7.M, Z31.D, Z31.D    // ffbfdd04

// FNMAD   <Zdn>.<T>, <Pg>/M, <Zm>.<T>, <Za>.<T>
    ZFNMAD P0.M, Z0.H, Z0.H, Z0.H       // 00c06065
    ZFNMAD P3.M, Z12.H, Z13.H, Z10.H    // 8acd6d65
    ZFNMAD P7.M, Z31.H, Z31.H, Z31.H    // ffdf7f65
    ZFNMAD P0.M, Z0.S, Z0.S, Z0.S       // 00c0a065
    ZFNMAD P3.M, Z12.S, Z13.S, Z10.S    // 8acdad65
    ZFNMAD P7.M, Z31.S, Z31.S, Z31.S    // ffdfbf65
    ZFNMAD P0.M, Z0.D, Z0.D, Z0.D       // 00c0e065
    ZFNMAD P3.M, Z12.D, Z13.D, Z10.D    // 8acded65
    ZFNMAD P7.M, Z31.D, Z31.D, Z31.D    // ffdfff65

// FNMLA   <Zda>.<T>, <Pg>/M, <Zn>.<T>, <Zm>.<T>
    ZFNMLA P0.M, Z0.H, Z0.H, Z0.H       // 00406065
    ZFNMLA P3.M, Z12.H, Z13.H, Z10.H    // 8a4d6d65
    ZFNMLA P7.M, Z31.H, Z31.H, Z31.H    // ff5f7f65
    ZFNMLA P0.M, Z0.S, Z0.S, Z0.S       // 0040a065
    ZFNMLA P3.M, Z12.S, Z13.S, Z10.S    // 8a4dad65
    ZFNMLA P7.M, Z31.S, Z31.S, Z31.S    // ff5fbf65
    ZFNMLA P0.M, Z0.D, Z0.D, Z0.D       // 0040e065
    ZFNMLA P3.M, Z12.D, Z13.D, Z10.D    // 8a4ded65
    ZFNMLA P7.M, Z31.D, Z31.D, Z31.D    // ff5fff65

// FNMLS   <Zda>.<T>, <Pg>/M, <Zn>.<T>, <Zm>.<T>
    ZFNMLS P0.M, Z0.H, Z0.H, Z0.H       // 00606065
    ZFNMLS P3.M, Z12.H, Z13.H, Z10.H    // 8a6d6d65
    ZFNMLS P7.M, Z31.H, Z31.H, Z31.H    // ff7f7f65
    ZFNMLS P0.M, Z0.S, Z0.S, Z0.S       // 0060a065
    ZFNMLS P3.M, Z12.S, Z13.S, Z10.S    // 8a6dad65
    ZFNMLS P7.M, Z31.S, Z31.S, Z31.S    // ff7fbf65
    ZFNMLS P0.M, Z0.D, Z0.D, Z0.D       // 0060e065
    ZFNMLS P3.M, Z12.D, Z13.D, Z10.D    // 8a6ded65
    ZFNMLS P7.M, Z31.D, Z31.D, Z31.D    // ff7fff65

// FNMSB   <Zdn>.<T>, <Pg>/M, <Zm>.<T>, <Za>.<T>
    ZFNMSB P0.M, Z0.H, Z0.H, Z0.H       // 00e06065
    ZFNMSB P3.M, Z12.H, Z13.H, Z10.H    // 8aed6d65
    ZFNMSB P7.M, Z31.H, Z31.H, Z31.H    // ffff7f65
    ZFNMSB P0.M, Z0.S, Z0.S, Z0.S       // 00e0a065
    ZFNMSB P3.M, Z12.S, Z13.S, Z10.S    // 8aedad65
    ZFNMSB P7.M, Z31.S, Z31.S, Z31.S    // ffffbf65
    ZFNMSB P0.M, Z0.D, Z0.D, Z0.D       // 00e0e065
    ZFNMSB P3.M, Z12.D, Z13.D, Z10.D    // 8aeded65
    ZFNMSB P7.M, Z31.D, Z31.D, Z31.D    // ffffff65

// FRECPE  <Zd>.<T>, <Zn>.<T>
    ZFRECPE Z0.H, Z0.H      // 00304e65
    ZFRECPE Z11.H, Z10.H    // 6a314e65
    ZFRECPE Z31.H, Z31.H    // ff334e65
    ZFRECPE Z0.S, Z0.S      // 00308e65
    ZFRECPE Z11.S, Z10.S    // 6a318e65
    ZFRECPE Z31.S, Z31.S    // ff338e65
    ZFRECPE Z0.D, Z0.D      // 0030ce65
    ZFRECPE Z11.D, Z10.D    // 6a31ce65
    ZFRECPE Z31.D, Z31.D    // ff33ce65

// FRECPS  <Zd>.<T>, <Zn>.<T>, <Zm>.<T>
    ZFRECPS Z0.H, Z0.H, Z0.H       // 00184065
    ZFRECPS Z11.H, Z12.H, Z10.H    // 6a194c65
    ZFRECPS Z31.H, Z31.H, Z31.H    // ff1b5f65
    ZFRECPS Z0.S, Z0.S, Z0.S       // 00188065
    ZFRECPS Z11.S, Z12.S, Z10.S    // 6a198c65
    ZFRECPS Z31.S, Z31.S, Z31.S    // ff1b9f65
    ZFRECPS Z0.D, Z0.D, Z0.D       // 0018c065
    ZFRECPS Z11.D, Z12.D, Z10.D    // 6a19cc65
    ZFRECPS Z31.D, Z31.D, Z31.D    // ff1bdf65

// FRECPX  <Zd>.<T>, <Pg>/M, <Zn>.<T>
    ZFRECPX P0.M, Z0.H, Z0.H      // 00a04c65
    ZFRECPX P3.M, Z12.H, Z10.H    // 8aad4c65
    ZFRECPX P7.M, Z31.H, Z31.H    // ffbf4c65
    ZFRECPX P0.M, Z0.S, Z0.S      // 00a08c65
    ZFRECPX P3.M, Z12.S, Z10.S    // 8aad8c65
    ZFRECPX P7.M, Z31.S, Z31.S    // ffbf8c65
    ZFRECPX P0.M, Z0.D, Z0.D      // 00a0cc65
    ZFRECPX P3.M, Z12.D, Z10.D    // 8aadcc65
    ZFRECPX P7.M, Z31.D, Z31.D    // ffbfcc65

// FRINTA  <Zd>.<T>, <Pg>/M, <Zn>.<T>
    ZFRINTA P0.M, Z0.H, Z0.H      // 00a04465
    ZFRINTA P3.M, Z12.H, Z10.H    // 8aad4465
    ZFRINTA P7.M, Z31.H, Z31.H    // ffbf4465
    ZFRINTA P0.M, Z0.S, Z0.S      // 00a08465
    ZFRINTA P3.M, Z12.S, Z10.S    // 8aad8465
    ZFRINTA P7.M, Z31.S, Z31.S    // ffbf8465
    ZFRINTA P0.M, Z0.D, Z0.D      // 00a0c465
    ZFRINTA P3.M, Z12.D, Z10.D    // 8aadc465
    ZFRINTA P7.M, Z31.D, Z31.D    // ffbfc465

// FRINTI  <Zd>.<T>, <Pg>/M, <Zn>.<T>
    ZFRINTI P0.M, Z0.H, Z0.H      // 00a04765
    ZFRINTI P3.M, Z12.H, Z10.H    // 8aad4765
    ZFRINTI P7.M, Z31.H, Z31.H    // ffbf4765
    ZFRINTI P0.M, Z0.S, Z0.S      // 00a08765
    ZFRINTI P3.M, Z12.S, Z10.S    // 8aad8765
    ZFRINTI P7.M, Z31.S, Z31.S    // ffbf8765
    ZFRINTI P0.M, Z0.D, Z0.D      // 00a0c765
    ZFRINTI P3.M, Z12.D, Z10.D    // 8aadc765
    ZFRINTI P7.M, Z31.D, Z31.D    // ffbfc765

// FRINTM  <Zd>.<T>, <Pg>/M, <Zn>.<T>
    ZFRINTM P0.M, Z0.H, Z0.H      // 00a04265
    ZFRINTM P3.M, Z12.H, Z10.H    // 8aad4265
    ZFRINTM P7.M, Z31.H, Z31.H    // ffbf4265
    ZFRINTM P0.M, Z0.S, Z0.S      // 00a08265
    ZFRINTM P3.M, Z12.S, Z10.S    // 8aad8265
    ZFRINTM P7.M, Z31.S, Z31.S    // ffbf8265
    ZFRINTM P0.M, Z0.D, Z0.D      // 00a0c265
    ZFRINTM P3.M, Z12.D, Z10.D    // 8aadc265
    ZFRINTM P7.M, Z31.D, Z31.D    // ffbfc265

// FRINTN  <Zd>.<T>, <Pg>/M, <Zn>.<T>
    ZFRINTN P0.M, Z0.H, Z0.H      // 00a04065
    ZFRINTN P3.M, Z12.H, Z10.H    // 8aad4065
    ZFRINTN P7.M, Z31.H, Z31.H    // ffbf4065
    ZFRINTN P0.M, Z0.S, Z0.S      // 00a08065
    ZFRINTN P3.M, Z12.S, Z10.S    // 8aad8065
    ZFRINTN P7.M, Z31.S, Z31.S    // ffbf8065
    ZFRINTN P0.M, Z0.D, Z0.D      // 00a0c065
    ZFRINTN P3.M, Z12.D, Z10.D    // 8aadc065
    ZFRINTN P7.M, Z31.D, Z31.D    // ffbfc065

// FRINTP  <Zd>.<T>, <Pg>/M, <Zn>.<T>
    ZFRINTP P0.M, Z0.H, Z0.H      // 00a04165
    ZFRINTP P3.M, Z12.H, Z10.H    // 8aad4165
    ZFRINTP P7.M, Z31.H, Z31.H    // ffbf4165
    ZFRINTP P0.M, Z0.S, Z0.S      // 00a08165
    ZFRINTP P3.M, Z12.S, Z10.S    // 8aad8165
    ZFRINTP P7.M, Z31.S, Z31.S    // ffbf8165
    ZFRINTP P0.M, Z0.D, Z0.D      // 00a0c165
    ZFRINTP P3.M, Z12.D, Z10.D    // 8aadc165
    ZFRINTP P7.M, Z31.D, Z31.D    // ffbfc165

// FRINTX  <Zd>.<T>, <Pg>/M, <Zn>.<T>
    ZFRINTX P0.M, Z0.H, Z0.H      // 00a04665
    ZFRINTX P3.M, Z12.H, Z10.H    // 8aad4665
    ZFRINTX P7.M, Z31.H, Z31.H    // ffbf4665
    ZFRINTX P0.M, Z0.S, Z0.S      // 00a08665
    ZFRINTX P3.M, Z12.S, Z10.S    // 8aad8665
    ZFRINTX P7.M, Z31.S, Z31.S    // ffbf8665
    ZFRINTX P0.M, Z0.D, Z0.D      // 00a0c665
    ZFRINTX P3.M, Z12.D, Z10.D    // 8aadc665
    ZFRINTX P7.M, Z31.D, Z31.D    // ffbfc665

// FRINTZ  <Zd>.<T>, <Pg>/M, <Zn>.<T>
    ZFRINTZ P0.M, Z0.H, Z0.H      // 00a04365
    ZFRINTZ P3.M, Z12.H, Z10.H    // 8aad4365
    ZFRINTZ P7.M, Z31.H, Z31.H    // ffbf4365
    ZFRINTZ P0.M, Z0.S, Z0.S      // 00a08365
    ZFRINTZ P3.M, Z12.S, Z10.S    // 8aad8365
    ZFRINTZ P7.M, Z31.S, Z31.S    // ffbf8365
    ZFRINTZ P0.M, Z0.D, Z0.D      // 00a0c365
    ZFRINTZ P3.M, Z12.D, Z10.D    // 8aadc365
    ZFRINTZ P7.M, Z31.D, Z31.D    // ffbfc365

// FRSQRTE <Zd>.<T>, <Zn>.<T>
    ZFRSQRTE Z0.H, Z0.H      // 00304f65
    ZFRSQRTE Z11.H, Z10.H    // 6a314f65
    ZFRSQRTE Z31.H, Z31.H    // ff334f65
    ZFRSQRTE Z0.S, Z0.S      // 00308f65
    ZFRSQRTE Z11.S, Z10.S    // 6a318f65
    ZFRSQRTE Z31.S, Z31.S    // ff338f65
    ZFRSQRTE Z0.D, Z0.D      // 0030cf65
    ZFRSQRTE Z11.D, Z10.D    // 6a31cf65
    ZFRSQRTE Z31.D, Z31.D    // ff33cf65

// FRSQRTS <Zd>.<T>, <Zn>.<T>, <Zm>.<T>
    ZFRSQRTS Z0.H, Z0.H, Z0.H       // 001c4065
    ZFRSQRTS Z11.H, Z12.H, Z10.H    // 6a1d4c65
    ZFRSQRTS Z31.H, Z31.H, Z31.H    // ff1f5f65
    ZFRSQRTS Z0.S, Z0.S, Z0.S       // 001c8065
    ZFRSQRTS Z11.S, Z12.S, Z10.S    // 6a1d8c65
    ZFRSQRTS Z31.S, Z31.S, Z31.S    // ff1f9f65
    ZFRSQRTS Z0.D, Z0.D, Z0.D       // 001cc065
    ZFRSQRTS Z11.D, Z12.D, Z10.D    // 6a1dcc65
    ZFRSQRTS Z31.D, Z31.D, Z31.D    // ff1fdf65

// FSCALE  <Zdn>.<T>, <Pg>/M, <Zdn>.<T>, <Zm>.<T>
    ZFSCALE P0.M, Z0.H, Z0.H, Z0.H       // 00804965
    ZFSCALE P3.M, Z10.H, Z12.H, Z10.H    // 8a8d4965
    ZFSCALE P7.M, Z31.H, Z31.H, Z31.H    // ff9f4965
    ZFSCALE P0.M, Z0.S, Z0.S, Z0.S       // 00808965
    ZFSCALE P3.M, Z10.S, Z12.S, Z10.S    // 8a8d8965
    ZFSCALE P7.M, Z31.S, Z31.S, Z31.S    // ff9f8965
    ZFSCALE P0.M, Z0.D, Z0.D, Z0.D       // 0080c965
    ZFSCALE P3.M, Z10.D, Z12.D, Z10.D    // 8a8dc965
    ZFSCALE P7.M, Z31.D, Z31.D, Z31.D    // ff9fc965

// FSQRT   <Zd>.<T>, <Pg>/M, <Zn>.<T>
    ZFSQRT P0.M, Z0.H, Z0.H      // 00a04d65
    ZFSQRT P3.M, Z12.H, Z10.H    // 8aad4d65
    ZFSQRT P7.M, Z31.H, Z31.H    // ffbf4d65
    ZFSQRT P0.M, Z0.S, Z0.S      // 00a08d65
    ZFSQRT P3.M, Z12.S, Z10.S    // 8aad8d65
    ZFSQRT P7.M, Z31.S, Z31.S    // ffbf8d65
    ZFSQRT P0.M, Z0.D, Z0.D      // 00a0cd65
    ZFSQRT P3.M, Z12.D, Z10.D    // 8aadcd65
    ZFSQRT P7.M, Z31.D, Z31.D    // ffbfcd65

// FSUB    <Zdn>.<T>, <Pg>/M, <Zdn>.<T>, <const>
    ZFSUB P0.M, Z0.H, $(0.5), Z0.H      // 00805965
    ZFSUB P3.M, Z10.H, $(0.5), Z10.H    // 0a8c5965
    ZFSUB P7.M, Z31.H, $(1.0), Z31.H    // 3f9c5965
    ZFSUB P0.M, Z0.S, $(0.5), Z0.S      // 00809965
    ZFSUB P3.M, Z10.S, $(0.5), Z10.S    // 0a8c9965
    ZFSUB P7.M, Z31.S, $(1.0), Z31.S    // 3f9c9965
    ZFSUB P0.M, Z0.D, $(0.5), Z0.D      // 0080d965
    ZFSUB P3.M, Z10.D, $(0.5), Z10.D    // 0a8cd965
    ZFSUB P7.M, Z31.D, $(1.0), Z31.D    // 3f9cd965

// FSUB    <Zdn>.<T>, <Pg>/M, <Zdn>.<T>, <Zm>.<T>
    ZFSUB P0.M, Z0.H, Z0.H, Z0.H       // 00804165
    ZFSUB P3.M, Z10.H, Z12.H, Z10.H    // 8a8d4165
    ZFSUB P7.M, Z31.H, Z31.H, Z31.H    // ff9f4165
    ZFSUB P0.M, Z0.S, Z0.S, Z0.S       // 00808165
    ZFSUB P3.M, Z10.S, Z12.S, Z10.S    // 8a8d8165
    ZFSUB P7.M, Z31.S, Z31.S, Z31.S    // ff9f8165
    ZFSUB P0.M, Z0.D, Z0.D, Z0.D       // 0080c165
    ZFSUB P3.M, Z10.D, Z12.D, Z10.D    // 8a8dc165
    ZFSUB P7.M, Z31.D, Z31.D, Z31.D    // ff9fc165

// FSUB    <Zd>.<T>, <Zn>.<T>, <Zm>.<T>
    ZFSUB Z0.H, Z0.H, Z0.H       // 00044065
    ZFSUB Z11.H, Z12.H, Z10.H    // 6a054c65
    ZFSUB Z31.H, Z31.H, Z31.H    // ff075f65
    ZFSUB Z0.S, Z0.S, Z0.S       // 00048065
    ZFSUB Z11.S, Z12.S, Z10.S    // 6a058c65
    ZFSUB Z31.S, Z31.S, Z31.S    // ff079f65
    ZFSUB Z0.D, Z0.D, Z0.D       // 0004c065
    ZFSUB Z11.D, Z12.D, Z10.D    // 6a05cc65
    ZFSUB Z31.D, Z31.D, Z31.D    // ff07df65

// FSUBR   <Zdn>.<T>, <Pg>/M, <Zdn>.<T>, <const>
    ZFSUBR P0.M, Z0.H, $(0.5), Z0.H      // 00805b65
    ZFSUBR P3.M, Z10.H, $(0.5), Z10.H    // 0a8c5b65
    ZFSUBR P7.M, Z31.H, $(1.0), Z31.H    // 3f9c5b65
    ZFSUBR P0.M, Z0.S, $(0.5), Z0.S      // 00809b65
    ZFSUBR P3.M, Z10.S, $(0.5), Z10.S    // 0a8c9b65
    ZFSUBR P7.M, Z31.S, $(1.0), Z31.S    // 3f9c9b65
    ZFSUBR P0.M, Z0.D, $(0.5), Z0.D      // 0080db65
    ZFSUBR P3.M, Z10.D, $(0.5), Z10.D    // 0a8cdb65
    ZFSUBR P7.M, Z31.D, $(1.0), Z31.D    // 3f9cdb65

// FSUBR   <Zdn>.<T>, <Pg>/M, <Zdn>.<T>, <Zm>.<T>
    ZFSUBR P0.M, Z0.H, Z0.H, Z0.H       // 00804365
    ZFSUBR P3.M, Z10.H, Z12.H, Z10.H    // 8a8d4365
    ZFSUBR P7.M, Z31.H, Z31.H, Z31.H    // ff9f4365
    ZFSUBR P0.M, Z0.S, Z0.S, Z0.S       // 00808365
    ZFSUBR P3.M, Z10.S, Z12.S, Z10.S    // 8a8d8365
    ZFSUBR P7.M, Z31.S, Z31.S, Z31.S    // ff9f8365
    ZFSUBR P0.M, Z0.D, Z0.D, Z0.D       // 0080c365
    ZFSUBR P3.M, Z10.D, Z12.D, Z10.D    // 8a8dc365
    ZFSUBR P7.M, Z31.D, Z31.D, Z31.D    // ff9fc365

// FTMAD   <Zdn>.<T>, <Zdn>.<T>, <Zm>.<T>, #<imm>
    ZFTMAD Z0.H, Z0.H, $0, Z0.H       // 00805065
    ZFTMAD Z10.H, Z11.H, $2, Z10.H    // 6a815265
    ZFTMAD Z31.H, Z31.H, $7, Z31.H    // ff835765
    ZFTMAD Z0.S, Z0.S, $0, Z0.S       // 00809065
    ZFTMAD Z10.S, Z11.S, $2, Z10.S    // 6a819265
    ZFTMAD Z31.S, Z31.S, $7, Z31.S    // ff839765
    ZFTMAD Z0.D, Z0.D, $0, Z0.D       // 0080d065
    ZFTMAD Z10.D, Z11.D, $2, Z10.D    // 6a81d265
    ZFTMAD Z31.D, Z31.D, $7, Z31.D    // ff83d765

// FTSMUL  <Zd>.<T>, <Zn>.<T>, <Zm>.<T>
    ZFTSMUL Z0.H, Z0.H, Z0.H       // 000c4065
    ZFTSMUL Z11.H, Z12.H, Z10.H    // 6a0d4c65
    ZFTSMUL Z31.H, Z31.H, Z31.H    // ff0f5f65
    ZFTSMUL Z0.S, Z0.S, Z0.S       // 000c8065
    ZFTSMUL Z11.S, Z12.S, Z10.S    // 6a0d8c65
    ZFTSMUL Z31.S, Z31.S, Z31.S    // ff0f9f65
    ZFTSMUL Z0.D, Z0.D, Z0.D       // 000cc065
    ZFTSMUL Z11.D, Z12.D, Z10.D    // 6a0dcc65
    ZFTSMUL Z31.D, Z31.D, Z31.D    // ff0fdf65

// FTSSEL  <Zd>.<T>, <Zn>.<T>, <Zm>.<T>
    ZFTSSEL Z0.H, Z0.H, Z0.H       // 00b06004
    ZFTSSEL Z11.H, Z12.H, Z10.H    // 6ab16c04
    ZFTSSEL Z31.H, Z31.H, Z31.H    // ffb37f04
    ZFTSSEL Z0.S, Z0.S, Z0.S       // 00b0a004
    ZFTSSEL Z11.S, Z12.S, Z10.S    // 6ab1ac04
    ZFTSSEL Z31.S, Z31.S, Z31.S    // ffb3bf04
    ZFTSSEL Z0.D, Z0.D, Z0.D       // 00b0e004
    ZFTSSEL Z11.D, Z12.D, Z10.D    // 6ab1ec04
    ZFTSSEL Z31.D, Z31.D, Z31.D    // ffb3ff04

// INCB    <Xdn>{, <pattern>{, MUL #<imm>}}
    ZINCB POW2, $1, R0      // 00e03004
    ZINCB VL1, $1, R1       // 21e03004
    ZINCB VL2, $2, R2       // 42e03104
    ZINCB VL3, $2, R3       // 63e03104
    ZINCB VL4, $3, R4       // 84e03204
    ZINCB VL5, $3, R5       // a5e03204
    ZINCB VL6, $4, R6       // c6e03304
    ZINCB VL7, $4, R7       // e7e03304
    ZINCB VL8, $5, R8       // 08e13404
    ZINCB VL16, $5, R8      // 28e13404
    ZINCB VL32, $6, R9      // 49e13504
    ZINCB VL64, $6, R10     // 6ae13504
    ZINCB VL128, $7, R11    // 8be13604
    ZINCB VL256, $7, R12    // ace13604
    ZINCB $14, $8, R13      // cde13704
    ZINCB $15, $8, R14      // eee13704
    ZINCB $16, $9, R15      // 0fe23804
    ZINCB $17, $9, R16      // 30e23804
    ZINCB $18, $9, R17      // 51e23804
    ZINCB $19, $10, R17     // 71e23904
    ZINCB $20, $10, R19     // 93e23904
    ZINCB $21, $11, R20     // b4e23a04
    ZINCB $22, $11, R21     // d5e23a04
    ZINCB $23, $12, R22     // f6e23b04
    ZINCB $24, $12, R22     // 16e33b04
    ZINCB $25, $13, R23     // 37e33c04
    ZINCB $26, $13, R24     // 58e33c04
    ZINCB $27, $14, R25     // 79e33d04
    ZINCB $28, $14, R26     // 9ae33d04
    ZINCB MUL4, $15, R27    // bbe33e04
    ZINCB MUL3, $15, R27    // dbe33e04
    ZINCB ALL, $16, R30     // fee33f04

// INCD    <Xdn>{, <pattern>{, MUL #<imm>}}
    ZINCD POW2, $1, R0      // 00e0f004
    ZINCD VL1, $1, R1       // 21e0f004
    ZINCD VL2, $2, R2       // 42e0f104
    ZINCD VL3, $2, R3       // 63e0f104
    ZINCD VL4, $3, R4       // 84e0f204
    ZINCD VL5, $3, R5       // a5e0f204
    ZINCD VL6, $4, R6       // c6e0f304
    ZINCD VL7, $4, R7       // e7e0f304
    ZINCD VL8, $5, R8       // 08e1f404
    ZINCD VL16, $5, R8      // 28e1f404
    ZINCD VL32, $6, R9      // 49e1f504
    ZINCD VL64, $6, R10     // 6ae1f504
    ZINCD VL128, $7, R11    // 8be1f604
    ZINCD VL256, $7, R12    // ace1f604
    ZINCD $14, $8, R13      // cde1f704
    ZINCD $15, $8, R14      // eee1f704
    ZINCD $16, $9, R15      // 0fe2f804
    ZINCD $17, $9, R16      // 30e2f804
    ZINCD $18, $9, R17      // 51e2f804
    ZINCD $19, $10, R17     // 71e2f904
    ZINCD $20, $10, R19     // 93e2f904
    ZINCD $21, $11, R20     // b4e2fa04
    ZINCD $22, $11, R21     // d5e2fa04
    ZINCD $23, $12, R22     // f6e2fb04
    ZINCD $24, $12, R22     // 16e3fb04
    ZINCD $25, $13, R23     // 37e3fc04
    ZINCD $26, $13, R24     // 58e3fc04
    ZINCD $27, $14, R25     // 79e3fd04
    ZINCD $28, $14, R26     // 9ae3fd04
    ZINCD MUL4, $15, R27    // bbe3fe04
    ZINCD MUL3, $15, R27    // dbe3fe04
    ZINCD ALL, $16, R30     // fee3ff04

// INCD    <Zdn>.D{, <pattern>{, MUL #<imm>}}
    ZINCD POW2, $1, Z0.D      // 00c0f004
    ZINCD VL1, $1, Z1.D       // 21c0f004
    ZINCD VL2, $2, Z2.D       // 42c0f104
    ZINCD VL3, $2, Z3.D       // 63c0f104
    ZINCD VL4, $3, Z4.D       // 84c0f204
    ZINCD VL5, $3, Z5.D       // a5c0f204
    ZINCD VL6, $4, Z6.D       // c6c0f304
    ZINCD VL7, $4, Z7.D       // e7c0f304
    ZINCD VL8, $5, Z8.D       // 08c1f404
    ZINCD VL16, $5, Z9.D      // 29c1f404
    ZINCD VL32, $6, Z10.D     // 4ac1f504
    ZINCD VL64, $6, Z11.D     // 6bc1f504
    ZINCD VL128, $7, Z12.D    // 8cc1f604
    ZINCD VL256, $7, Z13.D    // adc1f604
    ZINCD $14, $8, Z14.D      // cec1f704
    ZINCD $15, $8, Z15.D      // efc1f704
    ZINCD $16, $9, Z16.D      // 10c2f804
    ZINCD $17, $9, Z16.D      // 30c2f804
    ZINCD $18, $9, Z17.D      // 51c2f804
    ZINCD $19, $10, Z18.D     // 72c2f904
    ZINCD $20, $10, Z19.D     // 93c2f904
    ZINCD $21, $11, Z20.D     // b4c2fa04
    ZINCD $22, $11, Z21.D     // d5c2fa04
    ZINCD $23, $12, Z22.D     // f6c2fb04
    ZINCD $24, $12, Z23.D     // 17c3fb04
    ZINCD $25, $13, Z24.D     // 38c3fc04
    ZINCD $26, $13, Z25.D     // 59c3fc04
    ZINCD $27, $14, Z26.D     // 7ac3fd04
    ZINCD $28, $14, Z27.D     // 9bc3fd04
    ZINCD MUL4, $15, Z28.D    // bcc3fe04
    ZINCD MUL3, $15, Z29.D    // ddc3fe04
    ZINCD ALL, $16, Z31.D     // ffc3ff04

// INCH    <Xdn>{, <pattern>{, MUL #<imm>}}
    ZINCH POW2, $1, R0      // 00e07004
    ZINCH VL1, $1, R1       // 21e07004
    ZINCH VL2, $2, R2       // 42e07104
    ZINCH VL3, $2, R3       // 63e07104
    ZINCH VL4, $3, R4       // 84e07204
    ZINCH VL5, $3, R5       // a5e07204
    ZINCH VL6, $4, R6       // c6e07304
    ZINCH VL7, $4, R7       // e7e07304
    ZINCH VL8, $5, R8       // 08e17404
    ZINCH VL16, $5, R8      // 28e17404
    ZINCH VL32, $6, R9      // 49e17504
    ZINCH VL64, $6, R10     // 6ae17504
    ZINCH VL128, $7, R11    // 8be17604
    ZINCH VL256, $7, R12    // ace17604
    ZINCH $14, $8, R13      // cde17704
    ZINCH $15, $8, R14      // eee17704
    ZINCH $16, $9, R15      // 0fe27804
    ZINCH $17, $9, R16      // 30e27804
    ZINCH $18, $9, R17      // 51e27804
    ZINCH $19, $10, R17     // 71e27904
    ZINCH $20, $10, R19     // 93e27904
    ZINCH $21, $11, R20     // b4e27a04
    ZINCH $22, $11, R21     // d5e27a04
    ZINCH $23, $12, R22     // f6e27b04
    ZINCH $24, $12, R22     // 16e37b04
    ZINCH $25, $13, R23     // 37e37c04
    ZINCH $26, $13, R24     // 58e37c04
    ZINCH $27, $14, R25     // 79e37d04
    ZINCH $28, $14, R26     // 9ae37d04
    ZINCH MUL4, $15, R27    // bbe37e04
    ZINCH MUL3, $15, R27    // dbe37e04
    ZINCH ALL, $16, R30     // fee37f04

// INCH    <Zdn>.H{, <pattern>{, MUL #<imm>}}
    ZINCH POW2, $1, Z0.H      // 00c07004
    ZINCH VL1, $1, Z1.H       // 21c07004
    ZINCH VL2, $2, Z2.H       // 42c07104
    ZINCH VL3, $2, Z3.H       // 63c07104
    ZINCH VL4, $3, Z4.H       // 84c07204
    ZINCH VL5, $3, Z5.H       // a5c07204
    ZINCH VL6, $4, Z6.H       // c6c07304
    ZINCH VL7, $4, Z7.H       // e7c07304
    ZINCH VL8, $5, Z8.H       // 08c17404
    ZINCH VL16, $5, Z9.H      // 29c17404
    ZINCH VL32, $6, Z10.H     // 4ac17504
    ZINCH VL64, $6, Z11.H     // 6bc17504
    ZINCH VL128, $7, Z12.H    // 8cc17604
    ZINCH VL256, $7, Z13.H    // adc17604
    ZINCH $14, $8, Z14.H      // cec17704
    ZINCH $15, $8, Z15.H      // efc17704
    ZINCH $16, $9, Z16.H      // 10c27804
    ZINCH $17, $9, Z16.H      // 30c27804
    ZINCH $18, $9, Z17.H      // 51c27804
    ZINCH $19, $10, Z18.H     // 72c27904
    ZINCH $20, $10, Z19.H     // 93c27904
    ZINCH $21, $11, Z20.H     // b4c27a04
    ZINCH $22, $11, Z21.H     // d5c27a04
    ZINCH $23, $12, Z22.H     // f6c27b04
    ZINCH $24, $12, Z23.H     // 17c37b04
    ZINCH $25, $13, Z24.H     // 38c37c04
    ZINCH $26, $13, Z25.H     // 59c37c04
    ZINCH $27, $14, Z26.H     // 7ac37d04
    ZINCH $28, $14, Z27.H     // 9bc37d04
    ZINCH MUL4, $15, Z28.H    // bcc37e04
    ZINCH MUL3, $15, Z29.H    // ddc37e04
    ZINCH ALL, $16, Z31.H     // ffc37f04

// INCP    <Xdn>, <Pm>.<T>
    PINCP P0.B, R0      // 00882c25
    PINCP P6.B, R10     // ca882c25
    PINCP P15.B, R30    // fe892c25
    PINCP P0.H, R0      // 00886c25
    PINCP P6.H, R10     // ca886c25
    PINCP P15.H, R30    // fe896c25
    PINCP P0.S, R0      // 0088ac25
    PINCP P6.S, R10     // ca88ac25
    PINCP P15.S, R30    // fe89ac25
    PINCP P0.D, R0      // 0088ec25
    PINCP P6.D, R10     // ca88ec25
    PINCP P15.D, R30    // fe89ec25

// INCP    <Zdn>.<T>, <Pm>.<T>
    ZINCP P0.H, Z0.H      // 00806c25
    ZINCP P6.H, Z10.H     // ca806c25
    ZINCP P15.H, Z31.H    // ff816c25
    ZINCP P0.S, Z0.S      // 0080ac25
    ZINCP P6.S, Z10.S     // ca80ac25
    ZINCP P15.S, Z31.S    // ff81ac25
    ZINCP P0.D, Z0.D      // 0080ec25
    ZINCP P6.D, Z10.D     // ca80ec25
    ZINCP P15.D, Z31.D    // ff81ec25

// INCW    <Xdn>{, <pattern>{, MUL #<imm>}}
    ZINCW POW2, $1, R0      // 00e0b004
    ZINCW VL1, $1, R1       // 21e0b004
    ZINCW VL2, $2, R2       // 42e0b104
    ZINCW VL3, $2, R3       // 63e0b104
    ZINCW VL4, $3, R4       // 84e0b204
    ZINCW VL5, $3, R5       // a5e0b204
    ZINCW VL6, $4, R6       // c6e0b304
    ZINCW VL7, $4, R7       // e7e0b304
    ZINCW VL8, $5, R8       // 08e1b404
    ZINCW VL16, $5, R8      // 28e1b404
    ZINCW VL32, $6, R9      // 49e1b504
    ZINCW VL64, $6, R10     // 6ae1b504
    ZINCW VL128, $7, R11    // 8be1b604
    ZINCW VL256, $7, R12    // ace1b604
    ZINCW $14, $8, R13      // cde1b704
    ZINCW $15, $8, R14      // eee1b704
    ZINCW $16, $9, R15      // 0fe2b804
    ZINCW $17, $9, R16      // 30e2b804
    ZINCW $18, $9, R17      // 51e2b804
    ZINCW $19, $10, R17     // 71e2b904
    ZINCW $20, $10, R19     // 93e2b904
    ZINCW $21, $11, R20     // b4e2ba04
    ZINCW $22, $11, R21     // d5e2ba04
    ZINCW $23, $12, R22     // f6e2bb04
    ZINCW $24, $12, R22     // 16e3bb04
    ZINCW $25, $13, R23     // 37e3bc04
    ZINCW $26, $13, R24     // 58e3bc04
    ZINCW $27, $14, R25     // 79e3bd04
    ZINCW $28, $14, R26     // 9ae3bd04
    ZINCW MUL4, $15, R27    // bbe3be04
    ZINCW MUL3, $15, R27    // dbe3be04
    ZINCW ALL, $16, R30     // fee3bf04

// INCW    <Zdn>.S{, <pattern>{, MUL #<imm>}}
    ZINCW POW2, $1, Z0.S      // 00c0b004
    ZINCW VL1, $1, Z1.S       // 21c0b004
    ZINCW VL2, $2, Z2.S       // 42c0b104
    ZINCW VL3, $2, Z3.S       // 63c0b104
    ZINCW VL4, $3, Z4.S       // 84c0b204
    ZINCW VL5, $3, Z5.S       // a5c0b204
    ZINCW VL6, $4, Z6.S       // c6c0b304
    ZINCW VL7, $4, Z7.S       // e7c0b304
    ZINCW VL8, $5, Z8.S       // 08c1b404
    ZINCW VL16, $5, Z9.S      // 29c1b404
    ZINCW VL32, $6, Z10.S     // 4ac1b504
    ZINCW VL64, $6, Z11.S     // 6bc1b504
    ZINCW VL128, $7, Z12.S    // 8cc1b604
    ZINCW VL256, $7, Z13.S    // adc1b604
    ZINCW $14, $8, Z14.S      // cec1b704
    ZINCW $15, $8, Z15.S      // efc1b704
    ZINCW $16, $9, Z16.S      // 10c2b804
    ZINCW $17, $9, Z16.S      // 30c2b804
    ZINCW $18, $9, Z17.S      // 51c2b804
    ZINCW $19, $10, Z18.S     // 72c2b904
    ZINCW $20, $10, Z19.S     // 93c2b904
    ZINCW $21, $11, Z20.S     // b4c2ba04
    ZINCW $22, $11, Z21.S     // d5c2ba04
    ZINCW $23, $12, Z22.S     // f6c2bb04
    ZINCW $24, $12, Z23.S     // 17c3bb04
    ZINCW $25, $13, Z24.S     // 38c3bc04
    ZINCW $26, $13, Z25.S     // 59c3bc04
    ZINCW $27, $14, Z26.S     // 7ac3bd04
    ZINCW $28, $14, Z27.S     // 9bc3bd04
    ZINCW MUL4, $15, Z28.S    // bcc3be04
    ZINCW MUL3, $15, Z29.S    // ddc3be04
    ZINCW ALL, $16, Z31.S     // ffc3bf04

// INDEX   <Zd>.<T>, #<imm1>, #<imm2>
    ZINDEX $-16, $-16, Z0.B    // 00423004
    ZINDEX $-6, $-6, Z10.B     // 4a433a04
    ZINDEX $15, $15, Z31.B     // ff412f04
    ZINDEX $-16, $-16, Z0.H    // 00427004
    ZINDEX $-6, $-6, Z10.H     // 4a437a04
    ZINDEX $15, $15, Z31.H     // ff416f04
    ZINDEX $-16, $-16, Z0.S    // 0042b004
    ZINDEX $-6, $-6, Z10.S     // 4a43ba04
    ZINDEX $15, $15, Z31.S     // ff41af04
    ZINDEX $-16, $-16, Z0.D    // 0042f004
    ZINDEX $-6, $-6, Z10.D     // 4a43fa04
    ZINDEX $15, $15, Z31.D     // ff41ef04

// INDEX   <Zd>.<T>, #<imm>, <R><m>
    ZINDEX $-16, R0, Z0.B     // 004a2004
    ZINDEX $-6, R12, Z10.B    // 4a4b2c04
    ZINDEX $15, R30, Z31.B    // ff493e04
    ZINDEX $-16, R0, Z0.H     // 004a6004
    ZINDEX $-6, R12, Z10.H    // 4a4b6c04
    ZINDEX $15, R30, Z31.H    // ff497e04
    ZINDEX $-16, R0, Z0.S     // 004aa004
    ZINDEX $-6, R12, Z10.S    // 4a4bac04
    ZINDEX $15, R30, Z31.S    // ff49be04
    ZINDEX $-16, R0, Z0.D     // 004ae004
    ZINDEX $-6, R12, Z10.D    // 4a4bec04
    ZINDEX $15, R30, Z31.D    // ff49fe04

// INDEX   <Zd>.<T>, <R><n>, #<imm>
    ZINDEX R0, $-16, Z0.B     // 00443004
    ZINDEX R11, $-6, Z10.B    // 6a453a04
    ZINDEX R30, $15, Z31.B    // df472f04
    ZINDEX R0, $-16, Z0.H     // 00447004
    ZINDEX R11, $-6, Z10.H    // 6a457a04
    ZINDEX R30, $15, Z31.H    // df476f04
    ZINDEX R0, $-16, Z0.S     // 0044b004
    ZINDEX R11, $-6, Z10.S    // 6a45ba04
    ZINDEX R30, $15, Z31.S    // df47af04
    ZINDEX R0, $-16, Z0.D     // 0044f004
    ZINDEX R11, $-6, Z10.D    // 6a45fa04
    ZINDEX R30, $15, Z31.D    // df47ef04

// INDEX   <Zd>.<T>, <R><n>, <R><m>
    ZINDEX R0, R0, Z0.B       // 004c2004
    ZINDEX R11, R12, Z10.B    // 6a4d2c04
    ZINDEX R30, R30, Z31.B    // df4f3e04
    ZINDEX R0, R0, Z0.H       // 004c6004
    ZINDEX R11, R12, Z10.H    // 6a4d6c04
    ZINDEX R30, R30, Z31.H    // df4f7e04
    ZINDEX R0, R0, Z0.S       // 004ca004
    ZINDEX R11, R12, Z10.S    // 6a4dac04
    ZINDEX R30, R30, Z31.S    // df4fbe04
    ZINDEX R0, R0, Z0.D       // 004ce004
    ZINDEX R11, R12, Z10.D    // 6a4dec04
    ZINDEX R30, R30, Z31.D    // df4ffe04

// INSR    <Zdn>.<T>, <R><m>
    ZINSR R0, Z0.B      // 00382405
    ZINSR R11, Z10.B    // 6a392405
    ZINSR R30, Z31.B    // df3b2405
    ZINSR R0, Z0.H      // 00386405
    ZINSR R11, Z10.H    // 6a396405
    ZINSR R30, Z31.H    // df3b6405
    ZINSR R0, Z0.S      // 0038a405
    ZINSR R11, Z10.S    // 6a39a405
    ZINSR R30, Z31.S    // df3ba405
    ZINSR R0, Z0.D      // 0038e405
    ZINSR R11, Z10.D    // 6a39e405
    ZINSR R30, Z31.D    // df3be405

// INSR    <Zdn>.<T>, <V><m>
    ZINSR V0, Z0.B      // 00383405
    ZINSR V11, Z10.B    // 6a393405
    ZINSR V31, Z31.B    // ff3b3405
    ZINSR V0, Z0.H      // 00387405
    ZINSR V11, Z10.H    // 6a397405
    ZINSR V31, Z31.H    // ff3b7405
    ZINSR V0, Z0.S      // 0038b405
    ZINSR V11, Z10.S    // 6a39b405
    ZINSR V31, Z31.S    // ff3bb405
    ZINSR V0, Z0.D      // 0038f405
    ZINSR V11, Z10.D    // 6a39f405
    ZINSR V31, Z31.D    // ff3bf405

// LASTA   <R><d>, <Pg>, <Zn>.<T>
    ZLASTA P0, Z0.B, R0      // 00a02005
    ZLASTA P3, Z12.B, R10    // 8aad2005
    ZLASTA P7, Z31.B, R30    // febf2005
    ZLASTA P0, Z0.H, R0      // 00a06005
    ZLASTA P3, Z12.H, R10    // 8aad6005
    ZLASTA P7, Z31.H, R30    // febf6005
    ZLASTA P0, Z0.S, R0      // 00a0a005
    ZLASTA P3, Z12.S, R10    // 8aada005
    ZLASTA P7, Z31.S, R30    // febfa005
    ZLASTA P0, Z0.D, R0      // 00a0e005
    ZLASTA P3, Z12.D, R10    // 8aade005
    ZLASTA P7, Z31.D, R30    // febfe005

// LASTA   <V><d>, <Pg>, <Zn>.<T>
    ZLASTA P0, Z0.B, V0      // 00802205
    ZLASTA P3, Z12.B, V10    // 8a8d2205
    ZLASTA P7, Z31.B, V31    // ff9f2205
    ZLASTA P0, Z0.H, V0      // 00806205
    ZLASTA P3, Z12.H, V10    // 8a8d6205
    ZLASTA P7, Z31.H, V31    // ff9f6205
    ZLASTA P0, Z0.S, V0      // 0080a205
    ZLASTA P3, Z12.S, V10    // 8a8da205
    ZLASTA P7, Z31.S, V31    // ff9fa205
    ZLASTA P0, Z0.D, V0      // 0080e205
    ZLASTA P3, Z12.D, V10    // 8a8de205
    ZLASTA P7, Z31.D, V31    // ff9fe205

// LASTB   <R><d>, <Pg>, <Zn>.<T>
    ZLASTB P0, Z0.B, R0      // 00a02105
    ZLASTB P3, Z12.B, R10    // 8aad2105
    ZLASTB P7, Z31.B, R30    // febf2105
    ZLASTB P0, Z0.H, R0      // 00a06105
    ZLASTB P3, Z12.H, R10    // 8aad6105
    ZLASTB P7, Z31.H, R30    // febf6105
    ZLASTB P0, Z0.S, R0      // 00a0a105
    ZLASTB P3, Z12.S, R10    // 8aada105
    ZLASTB P7, Z31.S, R30    // febfa105
    ZLASTB P0, Z0.D, R0      // 00a0e105
    ZLASTB P3, Z12.D, R10    // 8aade105
    ZLASTB P7, Z31.D, R30    // febfe105

// LASTB   <V><d>, <Pg>, <Zn>.<T>
    ZLASTB P0, Z0.B, V0      // 00802305
    ZLASTB P3, Z12.B, V10    // 8a8d2305
    ZLASTB P7, Z31.B, V31    // ff9f2305
    ZLASTB P0, Z0.H, V0      // 00806305
    ZLASTB P3, Z12.H, V10    // 8a8d6305
    ZLASTB P7, Z31.H, V31    // ff9f6305
    ZLASTB P0, Z0.S, V0      // 0080a305
    ZLASTB P3, Z12.S, V10    // 8a8da305
    ZLASTB P7, Z31.S, V31    // ff9fa305
    ZLASTB P0, Z0.D, V0      // 0080e305
    ZLASTB P3, Z12.D, V10    // 8a8de305
    ZLASTB P7, Z31.D, V31    // ff9fe305

// LD1B    { <Zt>.D }, <Pg>/Z, [<Zn>.D{, #<pimm>}]
    ZLD1B P0.Z, (Z0.D), [Z0.D]       // 00c020c4
    ZLD1B P3.Z, 10(Z12.D), [Z10.D]    // 8acd2ac4
    ZLD1B P7.Z, 31(Z31.D), [Z31.D]    // ffdf3fc4

// LD1B    { <Zt>.S }, <Pg>/Z, [<Zn>.S{, #<pimm>}]
    ZLD1B P0.Z, (Z0.S), [Z0.S]       // 00c02084
    ZLD1B P3.Z, 10(Z12.S), [Z10.S]    // 8acd2a84
    ZLD1B P7.Z, 31(Z31.S), [Z31.S]    // ffdf3f84

// LD1B    { <Zt>.H }, <Pg>/Z, [<Xn|SP>{, #<simm>, MUL VL}]
    ZLD1B P0.Z, -8(R0), [Z0.H]      // 00a028a4
    ZLD1B P3.Z, -3(R12), [Z10.H]    // 8aad2da4
    ZLD1B P7.Z, 7(R30), [Z31.H]     // dfbf27a4

// LD1B    { <Zt>.S }, <Pg>/Z, [<Xn|SP>{, #<simm>, MUL VL}]
    ZLD1B P0.Z, -8(R0), [Z0.S]      // 00a048a4
    ZLD1B P3.Z, -3(R12), [Z10.S]    // 8aad4da4
    ZLD1B P7.Z, 7(R30), [Z31.S]     // dfbf47a4

// LD1B    { <Zt>.D }, <Pg>/Z, [<Xn|SP>{, #<simm>, MUL VL}]
    ZLD1B P0.Z, -8(R0), [Z0.D]      // 00a068a4
    ZLD1B P3.Z, -3(R12), [Z10.D]    // 8aad6da4
    ZLD1B P7.Z, 7(R30), [Z31.D]     // dfbf67a4

// LD1B    { <Zt>.B }, <Pg>/Z, [<Xn|SP>{, #<simm>, MUL VL}]
    ZLD1B P0.Z, -8(R0), [Z0.B]      // 00a008a4
    ZLD1B P3.Z, -3(R12), [Z10.B]    // 8aad0da4
    ZLD1B P7.Z, 7(R30), [Z31.B]     // dfbf07a4

// LD1B    { <Zt>.H }, <Pg>/Z, [<Xn|SP>, <Xm>]
    ZLD1B P0.Z, (R0)(R0), [Z0.H]       // 004020a4
    ZLD1B P3.Z, (R12)(R13), [Z10.H]    // 8a4d2da4
    ZLD1B P7.Z, (R30)(R30), [Z31.H]    // df5f3ea4

// LD1B    { <Zt>.S }, <Pg>/Z, [<Xn|SP>, <Xm>]
    ZLD1B P0.Z, (R0)(R0), [Z0.S]       // 004040a4
    ZLD1B P3.Z, (R12)(R13), [Z10.S]    // 8a4d4da4
    ZLD1B P7.Z, (R30)(R30), [Z31.S]    // df5f5ea4

// LD1B    { <Zt>.D }, <Pg>/Z, [<Xn|SP>, <Xm>]
    ZLD1B P0.Z, (R0)(R0), [Z0.D]       // 004060a4
    ZLD1B P3.Z, (R12)(R13), [Z10.D]    // 8a4d6da4
    ZLD1B P7.Z, (R30)(R30), [Z31.D]    // df5f7ea4

// LD1B    { <Zt>.B }, <Pg>/Z, [<Xn|SP>, <Xm>]
    ZLD1B P0.Z, (R0)(R0), [Z0.B]       // 004000a4
    ZLD1B P3.Z, (R12)(R13), [Z10.B]    // 8a4d0da4
    ZLD1B P7.Z, (R30)(R30), [Z31.B]    // df5f1ea4

// LD1B    { <Zt>.D }, <Pg>/Z, [<Xn|SP>, <Zm>.D]
    ZLD1B P0.Z, (R0)(Z0.D), [Z0.D]       // 00c040c4
    ZLD1B P3.Z, (R12)(Z13.D), [Z10.D]    // 8acd4dc4
    ZLD1B P7.Z, (R30)(Z31.D), [Z31.D]    // dfdf5fc4

// LD1B    { <Zt>.D }, <Pg>/Z, [<Xn|SP>, <Zm>.D, <extend>]
    ZLD1B P0.Z, (R0)(Z0.D.UXTW), [Z0.D]       // 004000c4
    ZLD1B P3.Z, (R12)(Z13.D.UXTW), [Z10.D]    // 8a4d0dc4
    ZLD1B P7.Z, (R30)(Z31.D.UXTW), [Z31.D]    // df5f1fc4
    ZLD1B P0.Z, (R0)(Z0.D.SXTW), [Z0.D]       // 004040c4
    ZLD1B P3.Z, (R12)(Z13.D.SXTW), [Z10.D]    // 8a4d4dc4
    ZLD1B P7.Z, (R30)(Z31.D.SXTW), [Z31.D]    // df5f5fc4

// LD1B    { <Zt>.S }, <Pg>/Z, [<Xn|SP>, <Zm>.S, <extend>]
    ZLD1B P0.Z, (R0)(Z0.S.UXTW), [Z0.S]       // 00400084
    ZLD1B P3.Z, (R12)(Z13.S.UXTW), [Z10.S]    // 8a4d0d84
    ZLD1B P7.Z, (R30)(Z31.S.UXTW), [Z31.S]    // df5f1f84
    ZLD1B P0.Z, (R0)(Z0.S.SXTW), [Z0.S]       // 00404084
    ZLD1B P3.Z, (R12)(Z13.S.SXTW), [Z10.S]    // 8a4d4d84
    ZLD1B P7.Z, (R30)(Z31.S.SXTW), [Z31.S]    // df5f5f84

// LD1D    { <Zt>.D }, <Pg>/Z, [<Zn>.D{, #<pimm>}]
    ZLD1D P0.Z, (Z0.D), [Z0.D]        // 00c0a0c5
    ZLD1D P3.Z, 80(Z12.D), [Z10.D]     // 8acdaac5
    ZLD1D P7.Z, 248(Z31.D), [Z31.D]    // ffdfbfc5

// LD1D    { <Zt>.D }, <Pg>/Z, [<Xn|SP>{, #<simm>, MUL VL}]
    ZLD1D P0.Z, -8(R0), [Z0.D]      // 00a0e8a5
    ZLD1D P3.Z, -3(R12), [Z10.D]    // 8aadeda5
    ZLD1D P7.Z, 7(R30), [Z31.D]     // dfbfe7a5

// LD1D    { <Zt>.D }, <Pg>/Z, [<Xn|SP>, <Xm>, LSL #3]
    ZLD1D P0.Z, (R0)(R0<<3), [Z0.D]       // 0040e0a5
    ZLD1D P3.Z, (R12)(R13<<3), [Z10.D]    // 8a4deda5
    ZLD1D P7.Z, (R30)(R30<<3), [Z31.D]    // df5ffea5

// LD1D    { <Zt>.D }, <Pg>/Z, [<Xn|SP>, <Zm>.D, LSL #3]
    ZLD1D P0.Z, (R0)(Z0.D.LSL<<3), [Z0.D]       // 00c0e0c5
    ZLD1D P3.Z, (R12)(Z13.D.LSL<<3), [Z10.D]    // 8acdedc5
    ZLD1D P7.Z, (R30)(Z31.D.LSL<<3), [Z31.D]    // dfdfffc5

// LD1D    { <Zt>.D }, <Pg>/Z, [<Xn|SP>, <Zm>.D]
    ZLD1D P0.Z, (R0)(Z0.D), [Z0.D]       // 00c0c0c5
    ZLD1D P3.Z, (R12)(Z13.D), [Z10.D]    // 8acdcdc5
    ZLD1D P7.Z, (R30)(Z31.D), [Z31.D]    // dfdfdfc5

// LD1D    { <Zt>.D }, <Pg>/Z, [<Xn|SP>, <Zm>.D, <extend> #3]
    ZLD1D P0.Z, (R0)(Z0.D.UXTW<<3), [Z0.D]       // 0040a0c5
    ZLD1D P3.Z, (R12)(Z13.D.UXTW<<3), [Z10.D]    // 8a4dadc5
    ZLD1D P7.Z, (R30)(Z31.D.UXTW<<3), [Z31.D]    // df5fbfc5
    ZLD1D P0.Z, (R0)(Z0.D.SXTW<<3), [Z0.D]       // 0040e0c5
    ZLD1D P3.Z, (R12)(Z13.D.SXTW<<3), [Z10.D]    // 8a4dedc5
    ZLD1D P7.Z, (R30)(Z31.D.SXTW<<3), [Z31.D]    // df5fffc5

// LD1D    { <Zt>.D }, <Pg>/Z, [<Xn|SP>, <Zm>.D, <extend>]
    ZLD1D P0.Z, (R0)(Z0.D.UXTW), [Z0.D]       // 004080c5
    ZLD1D P3.Z, (R12)(Z13.D.UXTW), [Z10.D]    // 8a4d8dc5
    ZLD1D P7.Z, (R30)(Z31.D.UXTW), [Z31.D]    // df5f9fc5
    ZLD1D P0.Z, (R0)(Z0.D.SXTW), [Z0.D]       // 0040c0c5
    ZLD1D P3.Z, (R12)(Z13.D.SXTW), [Z10.D]    // 8a4dcdc5
    ZLD1D P7.Z, (R30)(Z31.D.SXTW), [Z31.D]    // df5fdfc5

// LD1H    { <Zt>.D }, <Pg>/Z, [<Zn>.D{, #<pimm>}]
    ZLD1H P0.Z, (Z0.D), [Z0.D]       // 00c0a0c4
    ZLD1H P3.Z, 20(Z12.D), [Z10.D]    // 8acdaac4
    ZLD1H P7.Z, 62(Z31.D), [Z31.D]    // ffdfbfc4

// LD1H    { <Zt>.S }, <Pg>/Z, [<Zn>.S{, #<pimm>}]
    ZLD1H P0.Z, (Z0.S), [Z0.S]       // 00c0a084
    ZLD1H P3.Z, 20(Z12.S), [Z10.S]    // 8acdaa84
    ZLD1H P7.Z, 62(Z31.S), [Z31.S]    // ffdfbf84

// LD1H    { <Zt>.H }, <Pg>/Z, [<Xn|SP>{, #<simm>, MUL VL}]
    ZLD1H P0.Z, -8(R0), [Z0.H]      // 00a0a8a4
    ZLD1H P3.Z, -3(R12), [Z10.H]    // 8aadada4
    ZLD1H P7.Z, 7(R30), [Z31.H]     // dfbfa7a4

// LD1H    { <Zt>.S }, <Pg>/Z, [<Xn|SP>{, #<simm>, MUL VL}]
    ZLD1H P0.Z, -8(R0), [Z0.S]      // 00a0c8a4
    ZLD1H P3.Z, -3(R12), [Z10.S]    // 8aadcda4
    ZLD1H P7.Z, 7(R30), [Z31.S]     // dfbfc7a4

// LD1H    { <Zt>.D }, <Pg>/Z, [<Xn|SP>{, #<simm>, MUL VL}]
    ZLD1H P0.Z, -8(R0), [Z0.D]      // 00a0e8a4
    ZLD1H P3.Z, -3(R12), [Z10.D]    // 8aadeda4
    ZLD1H P7.Z, 7(R30), [Z31.D]     // dfbfe7a4

// LD1H    { <Zt>.H }, <Pg>/Z, [<Xn|SP>, <Xm>, LSL #1]
    ZLD1H P0.Z, (R0)(R0<<1), [Z0.H]       // 0040a0a4
    ZLD1H P3.Z, (R12)(R13<<1), [Z10.H]    // 8a4dada4
    ZLD1H P7.Z, (R30)(R30<<1), [Z31.H]    // df5fbea4

// LD1H    { <Zt>.S }, <Pg>/Z, [<Xn|SP>, <Xm>, LSL #1]
    ZLD1H P0.Z, (R0)(R0<<1), [Z0.S]       // 0040c0a4
    ZLD1H P3.Z, (R12)(R13<<1), [Z10.S]    // 8a4dcda4
    ZLD1H P7.Z, (R30)(R30<<1), [Z31.S]    // df5fdea4

// LD1H    { <Zt>.D }, <Pg>/Z, [<Xn|SP>, <Xm>, LSL #1]
    ZLD1H P0.Z, (R0)(R0<<1), [Z0.D]       // 0040e0a4
    ZLD1H P3.Z, (R12)(R13<<1), [Z10.D]    // 8a4deda4
    ZLD1H P7.Z, (R30)(R30<<1), [Z31.D]    // df5ffea4

// LD1H    { <Zt>.D }, <Pg>/Z, [<Xn|SP>, <Zm>.D, LSL #1]
    ZLD1H P0.Z, (R0)(Z0.D.LSL<<1), [Z0.D]       // 00c0e0c4
    ZLD1H P3.Z, (R12)(Z13.D.LSL<<1), [Z10.D]    // 8acdedc4
    ZLD1H P7.Z, (R30)(Z31.D.LSL<<1), [Z31.D]    // dfdfffc4

// LD1H    { <Zt>.D }, <Pg>/Z, [<Xn|SP>, <Zm>.D]
    ZLD1H P0.Z, (R0)(Z0.D), [Z0.D]       // 00c0c0c4
    ZLD1H P3.Z, (R12)(Z13.D), [Z10.D]    // 8acdcdc4
    ZLD1H P7.Z, (R30)(Z31.D), [Z31.D]    // dfdfdfc4

// LD1H    { <Zt>.D }, <Pg>/Z, [<Xn|SP>, <Zm>.D, <extend> #1]
    ZLD1H P0.Z, (R0)(Z0.D.UXTW<<1), [Z0.D]       // 0040a0c4
    ZLD1H P3.Z, (R12)(Z13.D.UXTW<<1), [Z10.D]    // 8a4dadc4
    ZLD1H P7.Z, (R30)(Z31.D.UXTW<<1), [Z31.D]    // df5fbfc4
    ZLD1H P0.Z, (R0)(Z0.D.SXTW<<1), [Z0.D]       // 0040e0c4
    ZLD1H P3.Z, (R12)(Z13.D.SXTW<<1), [Z10.D]    // 8a4dedc4
    ZLD1H P7.Z, (R30)(Z31.D.SXTW<<1), [Z31.D]    // df5fffc4

// LD1H    { <Zt>.D }, <Pg>/Z, [<Xn|SP>, <Zm>.D, <extend>]
    ZLD1H P0.Z, (R0)(Z0.D.UXTW), [Z0.D]       // 004080c4
    ZLD1H P3.Z, (R12)(Z13.D.UXTW), [Z10.D]    // 8a4d8dc4
    ZLD1H P7.Z, (R30)(Z31.D.UXTW), [Z31.D]    // df5f9fc4
    ZLD1H P0.Z, (R0)(Z0.D.SXTW), [Z0.D]       // 0040c0c4
    ZLD1H P3.Z, (R12)(Z13.D.SXTW), [Z10.D]    // 8a4dcdc4
    ZLD1H P7.Z, (R30)(Z31.D.SXTW), [Z31.D]    // df5fdfc4

// LD1H    { <Zt>.S }, <Pg>/Z, [<Xn|SP>, <Zm>.S, <extend> #1]
    ZLD1H P0.Z, (R0)(Z0.S.UXTW<<1), [Z0.S]       // 0040a084
    ZLD1H P3.Z, (R12)(Z13.S.UXTW<<1), [Z10.S]    // 8a4dad84
    ZLD1H P7.Z, (R30)(Z31.S.UXTW<<1), [Z31.S]    // df5fbf84
    ZLD1H P0.Z, (R0)(Z0.S.SXTW<<1), [Z0.S]       // 0040e084
    ZLD1H P3.Z, (R12)(Z13.S.SXTW<<1), [Z10.S]    // 8a4ded84
    ZLD1H P7.Z, (R30)(Z31.S.SXTW<<1), [Z31.S]    // df5fff84

// LD1H    { <Zt>.S }, <Pg>/Z, [<Xn|SP>, <Zm>.S, <extend>]
    ZLD1H P0.Z, (R0)(Z0.S.UXTW), [Z0.S]       // 00408084
    ZLD1H P3.Z, (R12)(Z13.S.UXTW), [Z10.S]    // 8a4d8d84
    ZLD1H P7.Z, (R30)(Z31.S.UXTW), [Z31.S]    // df5f9f84
    ZLD1H P0.Z, (R0)(Z0.S.SXTW), [Z0.S]       // 0040c084
    ZLD1H P3.Z, (R12)(Z13.S.SXTW), [Z10.S]    // 8a4dcd84
    ZLD1H P7.Z, (R30)(Z31.S.SXTW), [Z31.S]    // df5fdf84

// LD1RB   { <Zt>.H }, <Pg>/Z, [<Xn|SP>{, #<pimm>}]
    ZLD1RB P0.Z, (R0), [Z0.H]       // 00a04084
    ZLD1RB P3.Z, 21(R12), [Z10.H]    // 8aad5584
    ZLD1RB P7.Z, 63(R30), [Z31.H]    // dfbf7f84

// LD1RB   { <Zt>.S }, <Pg>/Z, [<Xn|SP>{, #<pimm>}]
    ZLD1RB P0.Z, (R0), [Z0.S]       // 00c04084
    ZLD1RB P3.Z, 21(R12), [Z10.S]    // 8acd5584
    ZLD1RB P7.Z, 63(R30), [Z31.S]    // dfdf7f84

// LD1RB   { <Zt>.D }, <Pg>/Z, [<Xn|SP>{, #<pimm>}]
    ZLD1RB P0.Z, (R0), [Z0.D]       // 00e04084
    ZLD1RB P3.Z, 21(R12), [Z10.D]    // 8aed5584
    ZLD1RB P7.Z, 63(R30), [Z31.D]    // dfff7f84

// LD1RB   { <Zt>.B }, <Pg>/Z, [<Xn|SP>{, #<pimm>}]
    ZLD1RB P0.Z, (R0), [Z0.B]       // 00804084
    ZLD1RB P3.Z, 21(R12), [Z10.B]    // 8a8d5584
    ZLD1RB P7.Z, 63(R30), [Z31.B]    // df9f7f84

// LD1RD   { <Zt>.D }, <Pg>/Z, [<Xn|SP>{, #<pimm>}]
    ZLD1RD P0.Z, (R0), [Z0.D]        // 00e0c085
    ZLD1RD P3.Z, 168(R12), [Z10.D]    // 8aedd585
    ZLD1RD P7.Z, 504(R30), [Z31.D]    // dfffff85

// LD1RH   { <Zt>.H }, <Pg>/Z, [<Xn|SP>{, #<pimm>}]
    ZLD1RH P0.Z, (R0), [Z0.H]        // 00a0c084
    ZLD1RH P3.Z, 42(R12), [Z10.H]     // 8aadd584
    ZLD1RH P7.Z, 126(R30), [Z31.H]    // dfbfff84

// LD1RH   { <Zt>.S }, <Pg>/Z, [<Xn|SP>{, #<pimm>}]
    ZLD1RH P0.Z, (R0), [Z0.S]        // 00c0c084
    ZLD1RH P3.Z, 42(R12), [Z10.S]     // 8acdd584
    ZLD1RH P7.Z, 126(R30), [Z31.S]    // dfdfff84

// LD1RH   { <Zt>.D }, <Pg>/Z, [<Xn|SP>{, #<pimm>}]
    ZLD1RH P0.Z, (R0), [Z0.D]        // 00e0c084
    ZLD1RH P3.Z, 42(R12), [Z10.D]     // 8aedd584
    ZLD1RH P7.Z, 126(R30), [Z31.D]    // dfffff84

// LD1ROB  { <Zt>.B }, <Pg>/Z, [<Xn|SP>{, #<simm>}]
    ZLD1ROB P0.Z, -256(R0), [Z0.B]     // 002028a4
    ZLD1ROB P3.Z, -96(R12), [Z10.B]    // 8a2d2da4
    ZLD1ROB P7.Z, 224(R30), [Z31.B]    // df3f27a4

// LD1ROB  { <Zt>.B }, <Pg>/Z, [<Xn|SP>, <Xm>]
    ZLD1ROB P0.Z, (R0)(R0), [Z0.B]       // 000020a4
    ZLD1ROB P3.Z, (R12)(R13), [Z10.B]    // 8a0d2da4
    ZLD1ROB P7.Z, (R30)(R30), [Z31.B]    // df1f3ea4

// LD1ROD  { <Zt>.D }, <Pg>/Z, [<Xn|SP>{, #<simm>}]
    ZLD1ROD P0.Z, -256(R0), [Z0.D]     // 0020a8a5
    ZLD1ROD P3.Z, -96(R12), [Z10.D]    // 8a2dada5
    ZLD1ROD P7.Z, 224(R30), [Z31.D]    // df3fa7a5

// LD1ROD  { <Zt>.D }, <Pg>/Z, [<Xn|SP>, <Xm>, LSL #3]
    ZLD1ROD P0.Z, (R0)(R0<<3), [Z0.D]       // 0000a0a5
    ZLD1ROD P3.Z, (R12)(R13<<3), [Z10.D]    // 8a0dada5
    ZLD1ROD P7.Z, (R30)(R30<<3), [Z31.D]    // df1fbea5

// LD1ROH  { <Zt>.H }, <Pg>/Z, [<Xn|SP>{, #<simm>}]
    ZLD1ROH P0.Z, -256(R0), [Z0.H]     // 0020a8a4
    ZLD1ROH P3.Z, -96(R12), [Z10.H]    // 8a2dada4
    ZLD1ROH P7.Z, 224(R30), [Z31.H]    // df3fa7a4

// LD1ROH  { <Zt>.H }, <Pg>/Z, [<Xn|SP>, <Xm>, LSL #1]
    ZLD1ROH P0.Z, (R0)(R0<<1), [Z0.H]       // 0000a0a4
    ZLD1ROH P3.Z, (R12)(R13<<1), [Z10.H]    // 8a0dada4
    ZLD1ROH P7.Z, (R30)(R30<<1), [Z31.H]    // df1fbea4

// LD1ROW  { <Zt>.S }, <Pg>/Z, [<Xn|SP>{, #<simm>}]
    ZLD1ROW P0.Z, -256(R0), [Z0.S]     // 002028a5
    ZLD1ROW P3.Z, -96(R12), [Z10.S]    // 8a2d2da5
    ZLD1ROW P7.Z, 224(R30), [Z31.S]    // df3f27a5

// LD1ROW  { <Zt>.S }, <Pg>/Z, [<Xn|SP>, <Xm>, LSL #2]
    ZLD1ROW P0.Z, (R0)(R0<<2), [Z0.S]       // 000020a5
    ZLD1ROW P3.Z, (R12)(R13<<2), [Z10.S]    // 8a0d2da5
    ZLD1ROW P7.Z, (R30)(R30<<2), [Z31.S]    // df1f3ea5

// LD1RQB  { <Zt>.B }, <Pg>/Z, [<Xn|SP>{, #<simm>}]
    ZLD1RQB P0.Z, -128(R0), [Z0.B]     // 002008a4
    ZLD1RQB P3.Z, -48(R12), [Z10.B]    // 8a2d0da4
    ZLD1RQB P7.Z, 112(R30), [Z31.B]    // df3f07a4

// LD1RQB  { <Zt>.B }, <Pg>/Z, [<Xn|SP>, <Xm>]
    ZLD1RQB P0.Z, (R0)(R0), [Z0.B]       // 000000a4
    ZLD1RQB P3.Z, (R12)(R13), [Z10.B]    // 8a0d0da4
    ZLD1RQB P7.Z, (R30)(R30), [Z31.B]    // df1f1ea4

// LD1RQD  { <Zt>.D }, <Pg>/Z, [<Xn|SP>{, #<simm>}]
    ZLD1RQD P0.Z, -128(R0), [Z0.D]     // 002088a5
    ZLD1RQD P3.Z, -48(R12), [Z10.D]    // 8a2d8da5
    ZLD1RQD P7.Z, 112(R30), [Z31.D]    // df3f87a5

// LD1RQD  { <Zt>.D }, <Pg>/Z, [<Xn|SP>, <Xm>, LSL #3]
    ZLD1RQD P0.Z, (R0)(R0<<3), [Z0.D]       // 000080a5
    ZLD1RQD P3.Z, (R12)(R13<<3), [Z10.D]    // 8a0d8da5
    ZLD1RQD P7.Z, (R30)(R30<<3), [Z31.D]    // df1f9ea5

// LD1RQH  { <Zt>.H }, <Pg>/Z, [<Xn|SP>{, #<simm>}]
    ZLD1RQH P0.Z, -128(R0), [Z0.H]     // 002088a4
    ZLD1RQH P3.Z, -48(R12), [Z10.H]    // 8a2d8da4
    ZLD1RQH P7.Z, 112(R30), [Z31.H]    // df3f87a4

// LD1RQH  { <Zt>.H }, <Pg>/Z, [<Xn|SP>, <Xm>, LSL #1]
    ZLD1RQH P0.Z, (R0)(R0<<1), [Z0.H]       // 000080a4
    ZLD1RQH P3.Z, (R12)(R13<<1), [Z10.H]    // 8a0d8da4
    ZLD1RQH P7.Z, (R30)(R30<<1), [Z31.H]    // df1f9ea4

// LD1RQW  { <Zt>.S }, <Pg>/Z, [<Xn|SP>{, #<simm>}]
    ZLD1RQW P0.Z, -128(R0), [Z0.S]     // 002008a5
    ZLD1RQW P3.Z, -48(R12), [Z10.S]    // 8a2d0da5
    ZLD1RQW P7.Z, 112(R30), [Z31.S]    // df3f07a5

// LD1RQW  { <Zt>.S }, <Pg>/Z, [<Xn|SP>, <Xm>, LSL #2]
    ZLD1RQW P0.Z, (R0)(R0<<2), [Z0.S]       // 000000a5
    ZLD1RQW P3.Z, (R12)(R13<<2), [Z10.S]    // 8a0d0da5
    ZLD1RQW P7.Z, (R30)(R30<<2), [Z31.S]    // df1f1ea5

// LD1RSB  { <Zt>.H }, <Pg>/Z, [<Xn|SP>{, #<pimm>}]
    ZLD1RSB P0.Z, (R0), [Z0.H]       // 00c0c085
    ZLD1RSB P3.Z, 21(R12), [Z10.H]    // 8acdd585
    ZLD1RSB P7.Z, 63(R30), [Z31.H]    // dfdfff85

// LD1RSB  { <Zt>.S }, <Pg>/Z, [<Xn|SP>{, #<pimm>}]
    ZLD1RSB P0.Z, (R0), [Z0.S]       // 00a0c085
    ZLD1RSB P3.Z, 21(R12), [Z10.S]    // 8aadd585
    ZLD1RSB P7.Z, 63(R30), [Z31.S]    // dfbfff85

// LD1RSB  { <Zt>.D }, <Pg>/Z, [<Xn|SP>{, #<pimm>}]
    ZLD1RSB P0.Z, (R0), [Z0.D]       // 0080c085
    ZLD1RSB P3.Z, 21(R12), [Z10.D]    // 8a8dd585
    ZLD1RSB P7.Z, 63(R30), [Z31.D]    // df9fff85

// LD1RSH  { <Zt>.S }, <Pg>/Z, [<Xn|SP>{, #<pimm>}]
    ZLD1RSH P0.Z, (R0), [Z0.S]        // 00a04085
    ZLD1RSH P3.Z, 42(R12), [Z10.S]     // 8aad5585
    ZLD1RSH P7.Z, 126(R30), [Z31.S]    // dfbf7f85

// LD1RSH  { <Zt>.D }, <Pg>/Z, [<Xn|SP>{, #<pimm>}]
    ZLD1RSH P0.Z, (R0), [Z0.D]        // 00804085
    ZLD1RSH P3.Z, 42(R12), [Z10.D]     // 8a8d5585
    ZLD1RSH P7.Z, 126(R30), [Z31.D]    // df9f7f85

// LD1RSW  { <Zt>.D }, <Pg>/Z, [<Xn|SP>{, #<pimm>}]
    ZLD1RSW P0.Z, (R0), [Z0.D]        // 0080c084
    ZLD1RSW P3.Z, 84(R12), [Z10.D]     // 8a8dd584
    ZLD1RSW P7.Z, 252(R30), [Z31.D]    // df9fff84

// LD1RW   { <Zt>.S }, <Pg>/Z, [<Xn|SP>{, #<pimm>}]
    ZLD1RW P0.Z, (R0), [Z0.S]        // 00c04085
    ZLD1RW P3.Z, 84(R12), [Z10.S]     // 8acd5585
    ZLD1RW P7.Z, 252(R30), [Z31.S]    // dfdf7f85

// LD1RW   { <Zt>.D }, <Pg>/Z, [<Xn|SP>{, #<pimm>}]
    ZLD1RW P0.Z, (R0), [Z0.D]        // 00e04085
    ZLD1RW P3.Z, 84(R12), [Z10.D]     // 8aed5585
    ZLD1RW P7.Z, 252(R30), [Z31.D]    // dfff7f85

// LD1SB   { <Zt>.D }, <Pg>/Z, [<Zn>.D{, #<pimm>}]
    ZLD1SB P0.Z, (Z0.D), [Z0.D]       // 008020c4
    ZLD1SB P3.Z, 10(Z12.D), [Z10.D]    // 8a8d2ac4
    ZLD1SB P7.Z, 31(Z31.D), [Z31.D]    // ff9f3fc4

// LD1SB   { <Zt>.S }, <Pg>/Z, [<Zn>.S{, #<pimm>}]
    ZLD1SB P0.Z, (Z0.S), [Z0.S]       // 00802084
    ZLD1SB P3.Z, 10(Z12.S), [Z10.S]    // 8a8d2a84
    ZLD1SB P7.Z, 31(Z31.S), [Z31.S]    // ff9f3f84

// LD1SB   { <Zt>.H }, <Pg>/Z, [<Xn|SP>{, #<simm>, MUL VL}]
    ZLD1SB P0.Z, -8(R0), [Z0.H]      // 00a0c8a5
    ZLD1SB P3.Z, -3(R12), [Z10.H]    // 8aadcda5
    ZLD1SB P7.Z, 7(R30), [Z31.H]     // dfbfc7a5

// LD1SB   { <Zt>.S }, <Pg>/Z, [<Xn|SP>{, #<simm>, MUL VL}]
    ZLD1SB P0.Z, -8(R0), [Z0.S]      // 00a0a8a5
    ZLD1SB P3.Z, -3(R12), [Z10.S]    // 8aadada5
    ZLD1SB P7.Z, 7(R30), [Z31.S]     // dfbfa7a5

// LD1SB   { <Zt>.D }, <Pg>/Z, [<Xn|SP>{, #<simm>, MUL VL}]
    ZLD1SB P0.Z, -8(R0), [Z0.D]      // 00a088a5
    ZLD1SB P3.Z, -3(R12), [Z10.D]    // 8aad8da5
    ZLD1SB P7.Z, 7(R30), [Z31.D]     // dfbf87a5

// LD1SB   { <Zt>.H }, <Pg>/Z, [<Xn|SP>, <Xm>]
    ZLD1SB P0.Z, (R0)(R0), [Z0.H]       // 0040c0a5
    ZLD1SB P3.Z, (R12)(R13), [Z10.H]    // 8a4dcda5
    ZLD1SB P7.Z, (R30)(R30), [Z31.H]    // df5fdea5

// LD1SB   { <Zt>.S }, <Pg>/Z, [<Xn|SP>, <Xm>]
    ZLD1SB P0.Z, (R0)(R0), [Z0.S]       // 0040a0a5
    ZLD1SB P3.Z, (R12)(R13), [Z10.S]    // 8a4dada5
    ZLD1SB P7.Z, (R30)(R30), [Z31.S]    // df5fbea5

// LD1SB   { <Zt>.D }, <Pg>/Z, [<Xn|SP>, <Xm>]
    ZLD1SB P0.Z, (R0)(R0), [Z0.D]       // 004080a5
    ZLD1SB P3.Z, (R12)(R13), [Z10.D]    // 8a4d8da5
    ZLD1SB P7.Z, (R30)(R30), [Z31.D]    // df5f9ea5

// LD1SB   { <Zt>.D }, <Pg>/Z, [<Xn|SP>, <Zm>.D]
    ZLD1SB P0.Z, (R0)(Z0.D), [Z0.D]       // 008040c4
    ZLD1SB P3.Z, (R12)(Z13.D), [Z10.D]    // 8a8d4dc4
    ZLD1SB P7.Z, (R30)(Z31.D), [Z31.D]    // df9f5fc4

// LD1SB   { <Zt>.D }, <Pg>/Z, [<Xn|SP>, <Zm>.D, <extend>]
    ZLD1SB P0.Z, (R0)(Z0.D.UXTW), [Z0.D]       // 000000c4
    ZLD1SB P3.Z, (R12)(Z13.D.UXTW), [Z10.D]    // 8a0d0dc4
    ZLD1SB P7.Z, (R30)(Z31.D.UXTW), [Z31.D]    // df1f1fc4
    ZLD1SB P0.Z, (R0)(Z0.D.SXTW), [Z0.D]       // 000040c4
    ZLD1SB P3.Z, (R12)(Z13.D.SXTW), [Z10.D]    // 8a0d4dc4
    ZLD1SB P7.Z, (R30)(Z31.D.SXTW), [Z31.D]    // df1f5fc4

// LD1SB   { <Zt>.S }, <Pg>/Z, [<Xn|SP>, <Zm>.S, <extend>]
    ZLD1SB P0.Z, (R0)(Z0.S.UXTW), [Z0.S]       // 00000084
    ZLD1SB P3.Z, (R12)(Z13.S.UXTW), [Z10.S]    // 8a0d0d84
    ZLD1SB P7.Z, (R30)(Z31.S.UXTW), [Z31.S]    // df1f1f84
    ZLD1SB P0.Z, (R0)(Z0.S.SXTW), [Z0.S]       // 00004084
    ZLD1SB P3.Z, (R12)(Z13.S.SXTW), [Z10.S]    // 8a0d4d84
    ZLD1SB P7.Z, (R30)(Z31.S.SXTW), [Z31.S]    // df1f5f84

// LD1SH   { <Zt>.D }, <Pg>/Z, [<Zn>.D{, #<pimm>}]
    ZLD1SH P0.Z, (Z0.D), [Z0.D]       // 0080a0c4
    ZLD1SH P3.Z, 20(Z12.D), [Z10.D]    // 8a8daac4
    ZLD1SH P7.Z, 62(Z31.D), [Z31.D]    // ff9fbfc4

// LD1SH   { <Zt>.S }, <Pg>/Z, [<Zn>.S{, #<pimm>}]
    ZLD1SH P0.Z, (Z0.S), [Z0.S]       // 0080a084
    ZLD1SH P3.Z, 20(Z12.S), [Z10.S]    // 8a8daa84
    ZLD1SH P7.Z, 62(Z31.S), [Z31.S]    // ff9fbf84

// LD1SH   { <Zt>.S }, <Pg>/Z, [<Xn|SP>{, #<simm>, MUL VL}]
    ZLD1SH P0.Z, -8(R0), [Z0.S]      // 00a028a5
    ZLD1SH P3.Z, -3(R12), [Z10.S]    // 8aad2da5
    ZLD1SH P7.Z, 7(R30), [Z31.S]     // dfbf27a5

// LD1SH   { <Zt>.D }, <Pg>/Z, [<Xn|SP>{, #<simm>, MUL VL}]
    ZLD1SH P0.Z, -8(R0), [Z0.D]      // 00a008a5
    ZLD1SH P3.Z, -3(R12), [Z10.D]    // 8aad0da5
    ZLD1SH P7.Z, 7(R30), [Z31.D]     // dfbf07a5

// LD1SH   { <Zt>.S }, <Pg>/Z, [<Xn|SP>, <Xm>, LSL #1]
    ZLD1SH P0.Z, (R0)(R0<<1), [Z0.S]       // 004020a5
    ZLD1SH P3.Z, (R12)(R13<<1), [Z10.S]    // 8a4d2da5
    ZLD1SH P7.Z, (R30)(R30<<1), [Z31.S]    // df5f3ea5

// LD1SH   { <Zt>.D }, <Pg>/Z, [<Xn|SP>, <Xm>, LSL #1]
    ZLD1SH P0.Z, (R0)(R0<<1), [Z0.D]       // 004000a5
    ZLD1SH P3.Z, (R12)(R13<<1), [Z10.D]    // 8a4d0da5
    ZLD1SH P7.Z, (R30)(R30<<1), [Z31.D]    // df5f1ea5

// LD1SH   { <Zt>.D }, <Pg>/Z, [<Xn|SP>, <Zm>.D, LSL #1]
    ZLD1SH P0.Z, (R0)(Z0.D.LSL<<1), [Z0.D]       // 0080e0c4
    ZLD1SH P3.Z, (R12)(Z13.D.LSL<<1), [Z10.D]    // 8a8dedc4
    ZLD1SH P7.Z, (R30)(Z31.D.LSL<<1), [Z31.D]    // df9fffc4

// LD1SH   { <Zt>.D }, <Pg>/Z, [<Xn|SP>, <Zm>.D]
    ZLD1SH P0.Z, (R0)(Z0.D), [Z0.D]       // 0080c0c4
    ZLD1SH P3.Z, (R12)(Z13.D), [Z10.D]    // 8a8dcdc4
    ZLD1SH P7.Z, (R30)(Z31.D), [Z31.D]    // df9fdfc4

// LD1SH   { <Zt>.D }, <Pg>/Z, [<Xn|SP>, <Zm>.D, <extend> #1]
    ZLD1SH P0.Z, (R0)(Z0.D.UXTW<<1), [Z0.D]       // 0000a0c4
    ZLD1SH P3.Z, (R12)(Z13.D.UXTW<<1), [Z10.D]    // 8a0dadc4
    ZLD1SH P7.Z, (R30)(Z31.D.UXTW<<1), [Z31.D]    // df1fbfc4
    ZLD1SH P0.Z, (R0)(Z0.D.SXTW<<1), [Z0.D]       // 0000e0c4
    ZLD1SH P3.Z, (R12)(Z13.D.SXTW<<1), [Z10.D]    // 8a0dedc4
    ZLD1SH P7.Z, (R30)(Z31.D.SXTW<<1), [Z31.D]    // df1fffc4

// LD1SH   { <Zt>.D }, <Pg>/Z, [<Xn|SP>, <Zm>.D, <extend>]
    ZLD1SH P0.Z, (R0)(Z0.D.UXTW), [Z0.D]       // 000080c4
    ZLD1SH P3.Z, (R12)(Z13.D.UXTW), [Z10.D]    // 8a0d8dc4
    ZLD1SH P7.Z, (R30)(Z31.D.UXTW), [Z31.D]    // df1f9fc4
    ZLD1SH P0.Z, (R0)(Z0.D.SXTW), [Z0.D]       // 0000c0c4
    ZLD1SH P3.Z, (R12)(Z13.D.SXTW), [Z10.D]    // 8a0dcdc4
    ZLD1SH P7.Z, (R30)(Z31.D.SXTW), [Z31.D]    // df1fdfc4

// LD1SH   { <Zt>.S }, <Pg>/Z, [<Xn|SP>, <Zm>.S, <extend> #1]
    ZLD1SH P0.Z, (R0)(Z0.S.UXTW<<1), [Z0.S]       // 0000a084
    ZLD1SH P3.Z, (R12)(Z13.S.UXTW<<1), [Z10.S]    // 8a0dad84
    ZLD1SH P7.Z, (R30)(Z31.S.UXTW<<1), [Z31.S]    // df1fbf84
    ZLD1SH P0.Z, (R0)(Z0.S.SXTW<<1), [Z0.S]       // 0000e084
    ZLD1SH P3.Z, (R12)(Z13.S.SXTW<<1), [Z10.S]    // 8a0ded84
    ZLD1SH P7.Z, (R30)(Z31.S.SXTW<<1), [Z31.S]    // df1fff84

// LD1SH   { <Zt>.S }, <Pg>/Z, [<Xn|SP>, <Zm>.S, <extend>]
    ZLD1SH P0.Z, (R0)(Z0.S.UXTW), [Z0.S]       // 00008084
    ZLD1SH P3.Z, (R12)(Z13.S.UXTW), [Z10.S]    // 8a0d8d84
    ZLD1SH P7.Z, (R30)(Z31.S.UXTW), [Z31.S]    // df1f9f84
    ZLD1SH P0.Z, (R0)(Z0.S.SXTW), [Z0.S]       // 0000c084
    ZLD1SH P3.Z, (R12)(Z13.S.SXTW), [Z10.S]    // 8a0dcd84
    ZLD1SH P7.Z, (R30)(Z31.S.SXTW), [Z31.S]    // df1fdf84

// LD1SW   { <Zt>.D }, <Pg>/Z, [<Zn>.D{, #<pimm>}]
    ZLD1SW P0.Z, (Z0.D), [Z0.D]        // 008020c5
    ZLD1SW P3.Z, 40(Z12.D), [Z10.D]     // 8a8d2ac5
    ZLD1SW P7.Z, 124(Z31.D), [Z31.D]    // ff9f3fc5

// LD1SW   { <Zt>.D }, <Pg>/Z, [<Xn|SP>{, #<simm>, MUL VL}]
    ZLD1SW P0.Z, -8(R0), [Z0.D]      // 00a088a4
    ZLD1SW P3.Z, -3(R12), [Z10.D]    // 8aad8da4
    ZLD1SW P7.Z, 7(R30), [Z31.D]     // dfbf87a4

// LD1SW   { <Zt>.D }, <Pg>/Z, [<Xn|SP>, <Xm>, LSL #2]
    ZLD1SW P0.Z, (R0)(R0<<2), [Z0.D]       // 004080a4
    ZLD1SW P3.Z, (R12)(R13<<2), [Z10.D]    // 8a4d8da4
    ZLD1SW P7.Z, (R30)(R30<<2), [Z31.D]    // df5f9ea4

// LD1SW   { <Zt>.D }, <Pg>/Z, [<Xn|SP>, <Zm>.D, LSL #2]
    ZLD1SW P0.Z, (R0)(Z0.D.LSL<<2), [Z0.D]       // 008060c5
    ZLD1SW P3.Z, (R12)(Z13.D.LSL<<2), [Z10.D]    // 8a8d6dc5
    ZLD1SW P7.Z, (R30)(Z31.D.LSL<<2), [Z31.D]    // df9f7fc5

// LD1SW   { <Zt>.D }, <Pg>/Z, [<Xn|SP>, <Zm>.D]
    ZLD1SW P0.Z, (R0)(Z0.D), [Z0.D]       // 008040c5
    ZLD1SW P3.Z, (R12)(Z13.D), [Z10.D]    // 8a8d4dc5
    ZLD1SW P7.Z, (R30)(Z31.D), [Z31.D]    // df9f5fc5

// LD1SW   { <Zt>.D }, <Pg>/Z, [<Xn|SP>, <Zm>.D, <extend> #2]
    ZLD1SW P0.Z, (R0)(Z0.D.UXTW<<2), [Z0.D]       // 000020c5
    ZLD1SW P3.Z, (R12)(Z13.D.UXTW<<2), [Z10.D]    // 8a0d2dc5
    ZLD1SW P7.Z, (R30)(Z31.D.UXTW<<2), [Z31.D]    // df1f3fc5
    ZLD1SW P0.Z, (R0)(Z0.D.SXTW<<2), [Z0.D]       // 000060c5
    ZLD1SW P3.Z, (R12)(Z13.D.SXTW<<2), [Z10.D]    // 8a0d6dc5
    ZLD1SW P7.Z, (R30)(Z31.D.SXTW<<2), [Z31.D]    // df1f7fc5

// LD1SW   { <Zt>.D }, <Pg>/Z, [<Xn|SP>, <Zm>.D, <extend>]
    ZLD1SW P0.Z, (R0)(Z0.D.UXTW), [Z0.D]       // 000000c5
    ZLD1SW P3.Z, (R12)(Z13.D.UXTW), [Z10.D]    // 8a0d0dc5
    ZLD1SW P7.Z, (R30)(Z31.D.UXTW), [Z31.D]    // df1f1fc5
    ZLD1SW P0.Z, (R0)(Z0.D.SXTW), [Z0.D]       // 000040c5
    ZLD1SW P3.Z, (R12)(Z13.D.SXTW), [Z10.D]    // 8a0d4dc5
    ZLD1SW P7.Z, (R30)(Z31.D.SXTW), [Z31.D]    // df1f5fc5

// LD1W    { <Zt>.D }, <Pg>/Z, [<Zn>.D{, #<pimm>}]
    ZLD1W P0.Z, (Z0.D), [Z0.D]        // 00c020c5
    ZLD1W P3.Z, 40(Z12.D), [Z10.D]     // 8acd2ac5
    ZLD1W P7.Z, 124(Z31.D), [Z31.D]    // ffdf3fc5

// LD1W    { <Zt>.S }, <Pg>/Z, [<Zn>.S{, #<pimm>}]
    ZLD1W P0.Z, (Z0.S), [Z0.S]        // 00c02085
    ZLD1W P3.Z, 40(Z12.S), [Z10.S]     // 8acd2a85
    ZLD1W P7.Z, 124(Z31.S), [Z31.S]    // ffdf3f85

// LD1W    { <Zt>.S }, <Pg>/Z, [<Xn|SP>{, #<simm>, MUL VL}]
    ZLD1W P0.Z, -8(R0), [Z0.S]      // 00a048a5
    ZLD1W P3.Z, -3(R12), [Z10.S]    // 8aad4da5
    ZLD1W P7.Z, 7(R30), [Z31.S]     // dfbf47a5

// LD1W    { <Zt>.D }, <Pg>/Z, [<Xn|SP>{, #<simm>, MUL VL}]
    ZLD1W P0.Z, -8(R0), [Z0.D]      // 00a068a5
    ZLD1W P3.Z, -3(R12), [Z10.D]    // 8aad6da5
    ZLD1W P7.Z, 7(R30), [Z31.D]     // dfbf67a5

// LD1W    { <Zt>.S }, <Pg>/Z, [<Xn|SP>, <Xm>, LSL #2]
    ZLD1W P0.Z, (R0)(R0<<2), [Z0.S]       // 004040a5
    ZLD1W P3.Z, (R12)(R13<<2), [Z10.S]    // 8a4d4da5
    ZLD1W P7.Z, (R30)(R30<<2), [Z31.S]    // df5f5ea5

// LD1W    { <Zt>.D }, <Pg>/Z, [<Xn|SP>, <Xm>, LSL #2]
    ZLD1W P0.Z, (R0)(R0<<2), [Z0.D]       // 004060a5
    ZLD1W P3.Z, (R12)(R13<<2), [Z10.D]    // 8a4d6da5
    ZLD1W P7.Z, (R30)(R30<<2), [Z31.D]    // df5f7ea5

// LD1W    { <Zt>.D }, <Pg>/Z, [<Xn|SP>, <Zm>.D, LSL #2]
    ZLD1W P0.Z, (R0)(Z0.D.LSL<<2), [Z0.D]       // 00c060c5
    ZLD1W P3.Z, (R12)(Z13.D.LSL<<2), [Z10.D]    // 8acd6dc5
    ZLD1W P7.Z, (R30)(Z31.D.LSL<<2), [Z31.D]    // dfdf7fc5

// LD1W    { <Zt>.D }, <Pg>/Z, [<Xn|SP>, <Zm>.D]
    ZLD1W P0.Z, (R0)(Z0.D), [Z0.D]       // 00c040c5
    ZLD1W P3.Z, (R12)(Z13.D), [Z10.D]    // 8acd4dc5
    ZLD1W P7.Z, (R30)(Z31.D), [Z31.D]    // dfdf5fc5

// LD1W    { <Zt>.D }, <Pg>/Z, [<Xn|SP>, <Zm>.D, <extend> #2]
    ZLD1W P0.Z, (R0)(Z0.D.UXTW<<2), [Z0.D]       // 004020c5
    ZLD1W P3.Z, (R12)(Z13.D.UXTW<<2), [Z10.D]    // 8a4d2dc5
    ZLD1W P7.Z, (R30)(Z31.D.UXTW<<2), [Z31.D]    // df5f3fc5
    ZLD1W P0.Z, (R0)(Z0.D.SXTW<<2), [Z0.D]       // 004060c5
    ZLD1W P3.Z, (R12)(Z13.D.SXTW<<2), [Z10.D]    // 8a4d6dc5
    ZLD1W P7.Z, (R30)(Z31.D.SXTW<<2), [Z31.D]    // df5f7fc5

// LD1W    { <Zt>.D }, <Pg>/Z, [<Xn|SP>, <Zm>.D, <extend>]
    ZLD1W P0.Z, (R0)(Z0.D.UXTW), [Z0.D]       // 004000c5
    ZLD1W P3.Z, (R12)(Z13.D.UXTW), [Z10.D]    // 8a4d0dc5
    ZLD1W P7.Z, (R30)(Z31.D.UXTW), [Z31.D]    // df5f1fc5
    ZLD1W P0.Z, (R0)(Z0.D.SXTW), [Z0.D]       // 004040c5
    ZLD1W P3.Z, (R12)(Z13.D.SXTW), [Z10.D]    // 8a4d4dc5
    ZLD1W P7.Z, (R30)(Z31.D.SXTW), [Z31.D]    // df5f5fc5

// LD1W    { <Zt>.S }, <Pg>/Z, [<Xn|SP>, <Zm>.S, <extend> #2]
    ZLD1W P0.Z, (R0)(Z0.S.UXTW<<2), [Z0.S]       // 00402085
    ZLD1W P3.Z, (R12)(Z13.S.UXTW<<2), [Z10.S]    // 8a4d2d85
    ZLD1W P7.Z, (R30)(Z31.S.UXTW<<2), [Z31.S]    // df5f3f85
    ZLD1W P0.Z, (R0)(Z0.S.SXTW<<2), [Z0.S]       // 00406085
    ZLD1W P3.Z, (R12)(Z13.S.SXTW<<2), [Z10.S]    // 8a4d6d85
    ZLD1W P7.Z, (R30)(Z31.S.SXTW<<2), [Z31.S]    // df5f7f85

// LD1W    { <Zt>.S }, <Pg>/Z, [<Xn|SP>, <Zm>.S, <extend>]
    ZLD1W P0.Z, (R0)(Z0.S.UXTW), [Z0.S]       // 00400085
    ZLD1W P3.Z, (R12)(Z13.S.UXTW), [Z10.S]    // 8a4d0d85
    ZLD1W P7.Z, (R30)(Z31.S.UXTW), [Z31.S]    // df5f1f85
    ZLD1W P0.Z, (R0)(Z0.S.SXTW), [Z0.S]       // 00404085
    ZLD1W P3.Z, (R12)(Z13.S.SXTW), [Z10.S]    // 8a4d4d85
    ZLD1W P7.Z, (R30)(Z31.S.SXTW), [Z31.S]    // df5f5f85

// LD2B    { <Zt1>.B, <Zt2>.B }, <Pg>/Z, [<Xn|SP>{, #<simm>, MUL VL}]
    ZLD2B P0.Z, -16(R0), [Z0.B, Z1.B]      // 00e028a4
    ZLD2B P3.Z, -6(R12), [Z10.B, Z11.B]    // 8aed2da4
    ZLD2B P7.Z, 14(R30), [Z31.B, Z0.B]     // dfff27a4

// LD2B    { <Zt1>.B, <Zt2>.B }, <Pg>/Z, [<Xn|SP>, <Xm>]
    ZLD2B P0.Z, (R0)(R0), [Z0.B, Z1.B]        // 00c020a4
    ZLD2B P3.Z, (R12)(R13), [Z10.B, Z11.B]    // 8acd2da4
    ZLD2B P7.Z, (R30)(R30), [Z31.B, Z0.B]     // dfdf3ea4

// LD2D    { <Zt1>.D, <Zt2>.D }, <Pg>/Z, [<Xn|SP>{, #<simm>, MUL VL}]
    ZLD2D P0.Z, -16(R0), [Z0.D, Z1.D]      // 00e0a8a5
    ZLD2D P3.Z, -6(R12), [Z10.D, Z11.D]    // 8aedada5
    ZLD2D P7.Z, 14(R30), [Z31.D, Z0.D]     // dfffa7a5

// LD2D    { <Zt1>.D, <Zt2>.D }, <Pg>/Z, [<Xn|SP>, <Xm>, LSL #3]
    ZLD2D P0.Z, (R0)(R0<<3), [Z0.D, Z1.D]        // 00c0a0a5
    ZLD2D P3.Z, (R12)(R13<<3), [Z10.D, Z11.D]    // 8acdada5
    ZLD2D P7.Z, (R30)(R30<<3), [Z31.D, Z0.D]     // dfdfbea5

// LD2H    { <Zt1>.H, <Zt2>.H }, <Pg>/Z, [<Xn|SP>{, #<simm>, MUL VL}]
    ZLD2H P0.Z, -16(R0), [Z0.H, Z1.H]      // 00e0a8a4
    ZLD2H P3.Z, -6(R12), [Z10.H, Z11.H]    // 8aedada4
    ZLD2H P7.Z, 14(R30), [Z31.H, Z0.H]     // dfffa7a4

// LD2H    { <Zt1>.H, <Zt2>.H }, <Pg>/Z, [<Xn|SP>, <Xm>, LSL #1]
    ZLD2H P0.Z, (R0)(R0<<1), [Z0.H, Z1.H]        // 00c0a0a4
    ZLD2H P3.Z, (R12)(R13<<1), [Z10.H, Z11.H]    // 8acdada4
    ZLD2H P7.Z, (R30)(R30<<1), [Z31.H, Z0.H]     // dfdfbea4

// LD2W    { <Zt1>.S, <Zt2>.S }, <Pg>/Z, [<Xn|SP>{, #<simm>, MUL VL}]
    ZLD2W P0.Z, -16(R0), [Z0.S, Z1.S]      // 00e028a5
    ZLD2W P3.Z, -6(R12), [Z10.S, Z11.S]    // 8aed2da5
    ZLD2W P7.Z, 14(R30), [Z31.S, Z0.S]     // dfff27a5

// LD2W    { <Zt1>.S, <Zt2>.S }, <Pg>/Z, [<Xn|SP>, <Xm>, LSL #2]
    ZLD2W P0.Z, (R0)(R0<<2), [Z0.S, Z1.S]        // 00c020a5
    ZLD2W P3.Z, (R12)(R13<<2), [Z10.S, Z11.S]    // 8acd2da5
    ZLD2W P7.Z, (R30)(R30<<2), [Z31.S, Z0.S]     // dfdf3ea5

// LD3B    { <Zt1>.B, <Zt2>.B, <Zt3>.B }, <Pg>/Z, [<Xn|SP>{, #<simm>, MUL VL}]
    ZLD3B P0.Z, -24(R0), [Z0.B, Z1.B, Z2.B]       // 00e048a4
    ZLD3B P3.Z, -9(R12), [Z10.B, Z11.B, Z12.B]    // 8aed4da4
    ZLD3B P7.Z, 21(R30), [Z31.B, Z0.B, Z1.B]      // dfff47a4

// LD3B    { <Zt1>.B, <Zt2>.B, <Zt3>.B }, <Pg>/Z, [<Xn|SP>, <Xm>]
    ZLD3B P0.Z, (R0)(R0), [Z0.B, Z1.B, Z2.B]         // 00c040a4
    ZLD3B P3.Z, (R12)(R13), [Z10.B, Z11.B, Z12.B]    // 8acd4da4
    ZLD3B P7.Z, (R30)(R30), [Z31.B, Z0.B, Z1.B]      // dfdf5ea4

// LD3D    { <Zt1>.D, <Zt2>.D, <Zt3>.D }, <Pg>/Z, [<Xn|SP>{, #<simm>, MUL VL}]
    ZLD3D P0.Z, -24(R0), [Z0.D, Z1.D, Z2.D]       // 00e0c8a5
    ZLD3D P3.Z, -9(R12), [Z10.D, Z11.D, Z12.D]    // 8aedcda5
    ZLD3D P7.Z, 21(R30), [Z31.D, Z0.D, Z1.D]      // dfffc7a5

// LD3D    { <Zt1>.D, <Zt2>.D, <Zt3>.D }, <Pg>/Z, [<Xn|SP>, <Xm>, LSL #3]
    ZLD3D P0.Z, (R0)(R0<<3), [Z0.D, Z1.D, Z2.D]         // 00c0c0a5
    ZLD3D P3.Z, (R12)(R13<<3), [Z10.D, Z11.D, Z12.D]    // 8acdcda5
    ZLD3D P7.Z, (R30)(R30<<3), [Z31.D, Z0.D, Z1.D]      // dfdfdea5

// LD3H    { <Zt1>.H, <Zt2>.H, <Zt3>.H }, <Pg>/Z, [<Xn|SP>{, #<simm>, MUL VL}]
    ZLD3H P0.Z, -24(R0), [Z0.H, Z1.H, Z2.H]       // 00e0c8a4
    ZLD3H P3.Z, -9(R12), [Z10.H, Z11.H, Z12.H]    // 8aedcda4
    ZLD3H P7.Z, 21(R30), [Z31.H, Z0.H, Z1.H]      // dfffc7a4

// LD3H    { <Zt1>.H, <Zt2>.H, <Zt3>.H }, <Pg>/Z, [<Xn|SP>, <Xm>, LSL #1]
    ZLD3H P0.Z, (R0)(R0<<1), [Z0.H, Z1.H, Z2.H]         // 00c0c0a4
    ZLD3H P3.Z, (R12)(R13<<1), [Z10.H, Z11.H, Z12.H]    // 8acdcda4
    ZLD3H P7.Z, (R30)(R30<<1), [Z31.H, Z0.H, Z1.H]      // dfdfdea4

// LD3W    { <Zt1>.S, <Zt2>.S, <Zt3>.S }, <Pg>/Z, [<Xn|SP>{, #<simm>, MUL VL}]
    ZLD3W P0.Z, -24(R0), [Z0.S, Z1.S, Z2.S]       // 00e048a5
    ZLD3W P3.Z, -9(R12), [Z10.S, Z11.S, Z12.S]    // 8aed4da5
    ZLD3W P7.Z, 21(R30), [Z31.S, Z0.S, Z1.S]      // dfff47a5

// LD3W    { <Zt1>.S, <Zt2>.S, <Zt3>.S }, <Pg>/Z, [<Xn|SP>, <Xm>, LSL #2]
    ZLD3W P0.Z, (R0)(R0<<2), [Z0.S, Z1.S, Z2.S]         // 00c040a5
    ZLD3W P3.Z, (R12)(R13<<2), [Z10.S, Z11.S, Z12.S]    // 8acd4da5
    ZLD3W P7.Z, (R30)(R30<<2), [Z31.S, Z0.S, Z1.S]      // dfdf5ea5

// LD4B    { <Zt1>.B, <Zt2>.B, <Zt3>.B, <Zt4>.B }, <Pg>/Z, [<Xn|SP>{, #<simm>, MUL VL}]
    ZLD4B P0.Z, -32(R0), [Z0.B, Z1.B, Z2.B, Z3.B]         // 00e068a4
    ZLD4B P3.Z, -12(R12), [Z10.B, Z11.B, Z12.B, Z13.B]    // 8aed6da4
    ZLD4B P7.Z, 28(R30), [Z31.B, Z0.B, Z1.B, Z2.B]        // dfff67a4

// LD4B    { <Zt1>.B, <Zt2>.B, <Zt3>.B, <Zt4>.B }, <Pg>/Z, [<Xn|SP>, <Xm>]
    ZLD4B P0.Z, (R0)(R0), [Z0.B, Z1.B, Z2.B, Z3.B]          // 00c060a4
    ZLD4B P3.Z, (R12)(R13), [Z10.B, Z11.B, Z12.B, Z13.B]    // 8acd6da4
    ZLD4B P7.Z, (R30)(R30), [Z31.B, Z0.B, Z1.B, Z2.B]       // dfdf7ea4

// LD4D    { <Zt1>.D, <Zt2>.D, <Zt3>.D, <Zt4>.D }, <Pg>/Z, [<Xn|SP>{, #<simm>, MUL VL}]
    ZLD4D P0.Z, -32(R0), [Z0.D, Z1.D, Z2.D, Z3.D]         // 00e0e8a5
    ZLD4D P3.Z, -12(R12), [Z10.D, Z11.D, Z12.D, Z13.D]    // 8aededa5
    ZLD4D P7.Z, 28(R30), [Z31.D, Z0.D, Z1.D, Z2.D]        // dfffe7a5

// LD4D    { <Zt1>.D, <Zt2>.D, <Zt3>.D, <Zt4>.D }, <Pg>/Z, [<Xn|SP>, <Xm>, LSL #3]
    ZLD4D P0.Z, (R0)(R0<<3), [Z0.D, Z1.D, Z2.D, Z3.D]          // 00c0e0a5
    ZLD4D P3.Z, (R12)(R13<<3), [Z10.D, Z11.D, Z12.D, Z13.D]    // 8acdeda5
    ZLD4D P7.Z, (R30)(R30<<3), [Z31.D, Z0.D, Z1.D, Z2.D]       // dfdffea5

// LD4H    { <Zt1>.H, <Zt2>.H, <Zt3>.H, <Zt4>.H }, <Pg>/Z, [<Xn|SP>{, #<simm>, MUL VL}]
    ZLD4H P0.Z, -32(R0), [Z0.H, Z1.H, Z2.H, Z3.H]         // 00e0e8a4
    ZLD4H P3.Z, -12(R12), [Z10.H, Z11.H, Z12.H, Z13.H]    // 8aededa4
    ZLD4H P7.Z, 28(R30), [Z31.H, Z0.H, Z1.H, Z2.H]        // dfffe7a4

// LD4H    { <Zt1>.H, <Zt2>.H, <Zt3>.H, <Zt4>.H }, <Pg>/Z, [<Xn|SP>, <Xm>, LSL #1]
    ZLD4H P0.Z, (R0)(R0<<1), [Z0.H, Z1.H, Z2.H, Z3.H]          // 00c0e0a4
    ZLD4H P3.Z, (R12)(R13<<1), [Z10.H, Z11.H, Z12.H, Z13.H]    // 8acdeda4
    ZLD4H P7.Z, (R30)(R30<<1), [Z31.H, Z0.H, Z1.H, Z2.H]       // dfdffea4

// LD4W    { <Zt1>.S, <Zt2>.S, <Zt3>.S, <Zt4>.S }, <Pg>/Z, [<Xn|SP>{, #<simm>, MUL VL}]
    ZLD4W P0.Z, -32(R0), [Z0.S, Z1.S, Z2.S, Z3.S]         // 00e068a5
    ZLD4W P3.Z, -12(R12), [Z10.S, Z11.S, Z12.S, Z13.S]    // 8aed6da5
    ZLD4W P7.Z, 28(R30), [Z31.S, Z0.S, Z1.S, Z2.S]        // dfff67a5

// LD4W    { <Zt1>.S, <Zt2>.S, <Zt3>.S, <Zt4>.S }, <Pg>/Z, [<Xn|SP>, <Xm>, LSL #2]
    ZLD4W P0.Z, (R0)(R0<<2), [Z0.S, Z1.S, Z2.S, Z3.S]          // 00c060a5
    ZLD4W P3.Z, (R12)(R13<<2), [Z10.S, Z11.S, Z12.S, Z13.S]    // 8acd6da5
    ZLD4W P7.Z, (R30)(R30<<2), [Z31.S, Z0.S, Z1.S, Z2.S]       // dfdf7ea5

// LDFF1B  { <Zt>.D }, <Pg>/Z, [<Zn>.D{, #<pimm>}]
    ZLDFF1B P0.Z, (Z0.D), [Z0.D]       // 00e020c4
    ZLDFF1B P3.Z, 10(Z12.D), [Z10.D]    // 8aed2ac4
    ZLDFF1B P7.Z, 31(Z31.D), [Z31.D]    // ffff3fc4

// LDFF1B  { <Zt>.S }, <Pg>/Z, [<Zn>.S{, #<pimm>}]
    ZLDFF1B P0.Z, (Z0.S), [Z0.S]       // 00e02084
    ZLDFF1B P3.Z, 10(Z12.S), [Z10.S]    // 8aed2a84
    ZLDFF1B P7.Z, 31(Z31.S), [Z31.S]    // ffff3f84

// LDFF1B  { <Zt>.H }, <Pg>/Z, [<Xn|SP>{, <Xm>}]
    ZLDFF1B P0.Z, (R0)(R0), [Z0.H]       // 006020a4
    ZLDFF1B P3.Z, (R12)(R13), [Z10.H]    // 8a6d2da4
    ZLDFF1B P7.Z, (R30)(R30), [Z31.H]    // df7f3ea4

// LDFF1B  { <Zt>.S }, <Pg>/Z, [<Xn|SP>{, <Xm>}]
    ZLDFF1B P0.Z, (R0)(R0), [Z0.S]       // 006040a4
    ZLDFF1B P3.Z, (R12)(R13), [Z10.S]    // 8a6d4da4
    ZLDFF1B P7.Z, (R30)(R30), [Z31.S]    // df7f5ea4

// LDFF1B  { <Zt>.D }, <Pg>/Z, [<Xn|SP>{, <Xm>}]
    ZLDFF1B P0.Z, (R0)(R0), [Z0.D]       // 006060a4
    ZLDFF1B P3.Z, (R12)(R13), [Z10.D]    // 8a6d6da4
    ZLDFF1B P7.Z, (R30)(R30), [Z31.D]    // df7f7ea4

// LDFF1B  { <Zt>.B }, <Pg>/Z, [<Xn|SP>{, <Xm>}]
    ZLDFF1B P0.Z, (R0)(R0), [Z0.B]       // 006000a4
    ZLDFF1B P3.Z, (R12)(R13), [Z10.B]    // 8a6d0da4
    ZLDFF1B P7.Z, (R30)(R30), [Z31.B]    // df7f1ea4

// LDFF1B  { <Zt>.D }, <Pg>/Z, [<Xn|SP>, <Zm>.D]
    ZLDFF1B P0.Z, (R0)(Z0.D), [Z0.D]       // 00e040c4
    ZLDFF1B P3.Z, (R12)(Z13.D), [Z10.D]    // 8aed4dc4
    ZLDFF1B P7.Z, (R30)(Z31.D), [Z31.D]    // dfff5fc4

// LDFF1B  { <Zt>.D }, <Pg>/Z, [<Xn|SP>, <Zm>.D, <extend>]
    ZLDFF1B P0.Z, (R0)(Z0.D.UXTW), [Z0.D]       // 006000c4
    ZLDFF1B P3.Z, (R12)(Z13.D.UXTW), [Z10.D]    // 8a6d0dc4
    ZLDFF1B P7.Z, (R30)(Z31.D.UXTW), [Z31.D]    // df7f1fc4
    ZLDFF1B P0.Z, (R0)(Z0.D.SXTW), [Z0.D]       // 006040c4
    ZLDFF1B P3.Z, (R12)(Z13.D.SXTW), [Z10.D]    // 8a6d4dc4
    ZLDFF1B P7.Z, (R30)(Z31.D.SXTW), [Z31.D]    // df7f5fc4

// LDFF1B  { <Zt>.S }, <Pg>/Z, [<Xn|SP>, <Zm>.S, <extend>]
    ZLDFF1B P0.Z, (R0)(Z0.S.UXTW), [Z0.S]       // 00600084
    ZLDFF1B P3.Z, (R12)(Z13.S.UXTW), [Z10.S]    // 8a6d0d84
    ZLDFF1B P7.Z, (R30)(Z31.S.UXTW), [Z31.S]    // df7f1f84
    ZLDFF1B P0.Z, (R0)(Z0.S.SXTW), [Z0.S]       // 00604084
    ZLDFF1B P3.Z, (R12)(Z13.S.SXTW), [Z10.S]    // 8a6d4d84
    ZLDFF1B P7.Z, (R30)(Z31.S.SXTW), [Z31.S]    // df7f5f84

// LDFF1D  { <Zt>.D }, <Pg>/Z, [<Zn>.D{, #<pimm>}]
    ZLDFF1D P0.Z, (Z0.D), [Z0.D]        // 00e0a0c5
    ZLDFF1D P3.Z, 80(Z12.D), [Z10.D]     // 8aedaac5
    ZLDFF1D P7.Z, 248(Z31.D), [Z31.D]    // ffffbfc5

// LDFF1D  { <Zt>.D }, <Pg>/Z, [<Xn|SP>{, <Xm>, LSL #3}]
    ZLDFF1D P0.Z, (R0)(R0<<3), [Z0.D]       // 0060e0a5
    ZLDFF1D P3.Z, (R12)(R13<<3), [Z10.D]    // 8a6deda5
    ZLDFF1D P7.Z, (R30)(R30<<3), [Z31.D]    // df7ffea5

// LDFF1D  { <Zt>.D }, <Pg>/Z, [<Xn|SP>, <Zm>.D, LSL #3]
    ZLDFF1D P0.Z, (R0)(Z0.D.LSL<<3), [Z0.D]       // 00e0e0c5
    ZLDFF1D P3.Z, (R12)(Z13.D.LSL<<3), [Z10.D]    // 8aededc5
    ZLDFF1D P7.Z, (R30)(Z31.D.LSL<<3), [Z31.D]    // dfffffc5

// LDFF1D  { <Zt>.D }, <Pg>/Z, [<Xn|SP>, <Zm>.D]
    ZLDFF1D P0.Z, (R0)(Z0.D), [Z0.D]       // 00e0c0c5
    ZLDFF1D P3.Z, (R12)(Z13.D), [Z10.D]    // 8aedcdc5
    ZLDFF1D P7.Z, (R30)(Z31.D), [Z31.D]    // dfffdfc5

// LDFF1D  { <Zt>.D }, <Pg>/Z, [<Xn|SP>, <Zm>.D, <extend> #3]
    ZLDFF1D P0.Z, (R0)(Z0.D.UXTW<<3), [Z0.D]       // 0060a0c5
    ZLDFF1D P3.Z, (R12)(Z13.D.UXTW<<3), [Z10.D]    // 8a6dadc5
    ZLDFF1D P7.Z, (R30)(Z31.D.UXTW<<3), [Z31.D]    // df7fbfc5
    ZLDFF1D P0.Z, (R0)(Z0.D.SXTW<<3), [Z0.D]       // 0060e0c5
    ZLDFF1D P3.Z, (R12)(Z13.D.SXTW<<3), [Z10.D]    // 8a6dedc5
    ZLDFF1D P7.Z, (R30)(Z31.D.SXTW<<3), [Z31.D]    // df7fffc5

// LDFF1D  { <Zt>.D }, <Pg>/Z, [<Xn|SP>, <Zm>.D, <extend>]
    ZLDFF1D P0.Z, (R0)(Z0.D.UXTW), [Z0.D]       // 006080c5
    ZLDFF1D P3.Z, (R12)(Z13.D.UXTW), [Z10.D]    // 8a6d8dc5
    ZLDFF1D P7.Z, (R30)(Z31.D.UXTW), [Z31.D]    // df7f9fc5
    ZLDFF1D P0.Z, (R0)(Z0.D.SXTW), [Z0.D]       // 0060c0c5
    ZLDFF1D P3.Z, (R12)(Z13.D.SXTW), [Z10.D]    // 8a6dcdc5
    ZLDFF1D P7.Z, (R30)(Z31.D.SXTW), [Z31.D]    // df7fdfc5

// LDFF1H  { <Zt>.D }, <Pg>/Z, [<Zn>.D{, #<pimm>}]
    ZLDFF1H P0.Z, (Z0.D), [Z0.D]       // 00e0a0c4
    ZLDFF1H P3.Z, 20(Z12.D), [Z10.D]    // 8aedaac4
    ZLDFF1H P7.Z, 62(Z31.D), [Z31.D]    // ffffbfc4

// LDFF1H  { <Zt>.S }, <Pg>/Z, [<Zn>.S{, #<pimm>}]
    ZLDFF1H P0.Z, (Z0.S), [Z0.S]       // 00e0a084
    ZLDFF1H P3.Z, 20(Z12.S), [Z10.S]    // 8aedaa84
    ZLDFF1H P7.Z, 62(Z31.S), [Z31.S]    // ffffbf84

// LDFF1H  { <Zt>.H }, <Pg>/Z, [<Xn|SP>{, <Xm>, LSL #1}]
    ZLDFF1H P0.Z, (R0)(R0<<1), [Z0.H]       // 0060a0a4
    ZLDFF1H P3.Z, (R12)(R13<<1), [Z10.H]    // 8a6dada4
    ZLDFF1H P7.Z, (R30)(R30<<1), [Z31.H]    // df7fbea4

// LDFF1H  { <Zt>.S }, <Pg>/Z, [<Xn|SP>{, <Xm>, LSL #1}]
    ZLDFF1H P0.Z, (R0)(R0<<1), [Z0.S]       // 0060c0a4
    ZLDFF1H P3.Z, (R12)(R13<<1), [Z10.S]    // 8a6dcda4
    ZLDFF1H P7.Z, (R30)(R30<<1), [Z31.S]    // df7fdea4

// LDFF1H  { <Zt>.D }, <Pg>/Z, [<Xn|SP>{, <Xm>, LSL #1}]
    ZLDFF1H P0.Z, (R0)(R0<<1), [Z0.D]       // 0060e0a4
    ZLDFF1H P3.Z, (R12)(R13<<1), [Z10.D]    // 8a6deda4
    ZLDFF1H P7.Z, (R30)(R30<<1), [Z31.D]    // df7ffea4

// LDFF1H  { <Zt>.D }, <Pg>/Z, [<Xn|SP>, <Zm>.D, LSL #1]
    ZLDFF1H P0.Z, (R0)(Z0.D.LSL<<1), [Z0.D]       // 00e0e0c4
    ZLDFF1H P3.Z, (R12)(Z13.D.LSL<<1), [Z10.D]    // 8aededc4
    ZLDFF1H P7.Z, (R30)(Z31.D.LSL<<1), [Z31.D]    // dfffffc4

// LDFF1H  { <Zt>.D }, <Pg>/Z, [<Xn|SP>, <Zm>.D]
    ZLDFF1H P0.Z, (R0)(Z0.D), [Z0.D]       // 00e0c0c4
    ZLDFF1H P3.Z, (R12)(Z13.D), [Z10.D]    // 8aedcdc4
    ZLDFF1H P7.Z, (R30)(Z31.D), [Z31.D]    // dfffdfc4

// LDFF1H  { <Zt>.D }, <Pg>/Z, [<Xn|SP>, <Zm>.D, <extend> #1]
    ZLDFF1H P0.Z, (R0)(Z0.D.UXTW<<1), [Z0.D]       // 0060a0c4
    ZLDFF1H P3.Z, (R12)(Z13.D.UXTW<<1), [Z10.D]    // 8a6dadc4
    ZLDFF1H P7.Z, (R30)(Z31.D.UXTW<<1), [Z31.D]    // df7fbfc4
    ZLDFF1H P0.Z, (R0)(Z0.D.SXTW<<1), [Z0.D]       // 0060e0c4
    ZLDFF1H P3.Z, (R12)(Z13.D.SXTW<<1), [Z10.D]    // 8a6dedc4
    ZLDFF1H P7.Z, (R30)(Z31.D.SXTW<<1), [Z31.D]    // df7fffc4

// LDFF1H  { <Zt>.D }, <Pg>/Z, [<Xn|SP>, <Zm>.D, <extend>]
    ZLDFF1H P0.Z, (R0)(Z0.D.UXTW), [Z0.D]       // 006080c4
    ZLDFF1H P3.Z, (R12)(Z13.D.UXTW), [Z10.D]    // 8a6d8dc4
    ZLDFF1H P7.Z, (R30)(Z31.D.UXTW), [Z31.D]    // df7f9fc4
    ZLDFF1H P0.Z, (R0)(Z0.D.SXTW), [Z0.D]       // 0060c0c4
    ZLDFF1H P3.Z, (R12)(Z13.D.SXTW), [Z10.D]    // 8a6dcdc4
    ZLDFF1H P7.Z, (R30)(Z31.D.SXTW), [Z31.D]    // df7fdfc4

// LDFF1H  { <Zt>.S }, <Pg>/Z, [<Xn|SP>, <Zm>.S, <extend> #1]
    ZLDFF1H P0.Z, (R0)(Z0.S.UXTW<<1), [Z0.S]       // 0060a084
    ZLDFF1H P3.Z, (R12)(Z13.S.UXTW<<1), [Z10.S]    // 8a6dad84
    ZLDFF1H P7.Z, (R30)(Z31.S.UXTW<<1), [Z31.S]    // df7fbf84
    ZLDFF1H P0.Z, (R0)(Z0.S.SXTW<<1), [Z0.S]       // 0060e084
    ZLDFF1H P3.Z, (R12)(Z13.S.SXTW<<1), [Z10.S]    // 8a6ded84
    ZLDFF1H P7.Z, (R30)(Z31.S.SXTW<<1), [Z31.S]    // df7fff84

// LDFF1H  { <Zt>.S }, <Pg>/Z, [<Xn|SP>, <Zm>.S, <extend>]
    ZLDFF1H P0.Z, (R0)(Z0.S.UXTW), [Z0.S]       // 00608084
    ZLDFF1H P3.Z, (R12)(Z13.S.UXTW), [Z10.S]    // 8a6d8d84
    ZLDFF1H P7.Z, (R30)(Z31.S.UXTW), [Z31.S]    // df7f9f84
    ZLDFF1H P0.Z, (R0)(Z0.S.SXTW), [Z0.S]       // 0060c084
    ZLDFF1H P3.Z, (R12)(Z13.S.SXTW), [Z10.S]    // 8a6dcd84
    ZLDFF1H P7.Z, (R30)(Z31.S.SXTW), [Z31.S]    // df7fdf84

// LDFF1SB { <Zt>.D }, <Pg>/Z, [<Zn>.D{, #<pimm>}]
    ZLDFF1SB P0.Z, (Z0.D), [Z0.D]       // 00a020c4
    ZLDFF1SB P3.Z, 10(Z12.D), [Z10.D]    // 8aad2ac4
    ZLDFF1SB P7.Z, 31(Z31.D), [Z31.D]    // ffbf3fc4

// LDFF1SB { <Zt>.S }, <Pg>/Z, [<Zn>.S{, #<pimm>}]
    ZLDFF1SB P0.Z, (Z0.S), [Z0.S]       // 00a02084
    ZLDFF1SB P3.Z, 10(Z12.S), [Z10.S]    // 8aad2a84
    ZLDFF1SB P7.Z, 31(Z31.S), [Z31.S]    // ffbf3f84

// LDFF1SB { <Zt>.H }, <Pg>/Z, [<Xn|SP>{, <Xm>}]
    ZLDFF1SB P0.Z, (R0)(R0), [Z0.H]       // 0060c0a5
    ZLDFF1SB P3.Z, (R12)(R13), [Z10.H]    // 8a6dcda5
    ZLDFF1SB P7.Z, (R30)(R30), [Z31.H]    // df7fdea5

// LDFF1SB { <Zt>.S }, <Pg>/Z, [<Xn|SP>{, <Xm>}]
    ZLDFF1SB P0.Z, (R0)(R0), [Z0.S]       // 0060a0a5
    ZLDFF1SB P3.Z, (R12)(R13), [Z10.S]    // 8a6dada5
    ZLDFF1SB P7.Z, (R30)(R30), [Z31.S]    // df7fbea5

// LDFF1SB { <Zt>.D }, <Pg>/Z, [<Xn|SP>{, <Xm>}]
    ZLDFF1SB P0.Z, (R0)(R0), [Z0.D]       // 006080a5
    ZLDFF1SB P3.Z, (R12)(R13), [Z10.D]    // 8a6d8da5
    ZLDFF1SB P7.Z, (R30)(R30), [Z31.D]    // df7f9ea5

// LDFF1SB { <Zt>.D }, <Pg>/Z, [<Xn|SP>, <Zm>.D]
    ZLDFF1SB P0.Z, (R0)(Z0.D), [Z0.D]       // 00a040c4
    ZLDFF1SB P3.Z, (R12)(Z13.D), [Z10.D]    // 8aad4dc4
    ZLDFF1SB P7.Z, (R30)(Z31.D), [Z31.D]    // dfbf5fc4

// LDFF1SB { <Zt>.D }, <Pg>/Z, [<Xn|SP>, <Zm>.D, <extend>]
    ZLDFF1SB P0.Z, (R0)(Z0.D.UXTW), [Z0.D]       // 002000c4
    ZLDFF1SB P3.Z, (R12)(Z13.D.UXTW), [Z10.D]    // 8a2d0dc4
    ZLDFF1SB P7.Z, (R30)(Z31.D.UXTW), [Z31.D]    // df3f1fc4
    ZLDFF1SB P0.Z, (R0)(Z0.D.SXTW), [Z0.D]       // 002040c4
    ZLDFF1SB P3.Z, (R12)(Z13.D.SXTW), [Z10.D]    // 8a2d4dc4
    ZLDFF1SB P7.Z, (R30)(Z31.D.SXTW), [Z31.D]    // df3f5fc4

// LDFF1SB { <Zt>.S }, <Pg>/Z, [<Xn|SP>, <Zm>.S, <extend>]
    ZLDFF1SB P0.Z, (R0)(Z0.S.UXTW), [Z0.S]       // 00200084
    ZLDFF1SB P3.Z, (R12)(Z13.S.UXTW), [Z10.S]    // 8a2d0d84
    ZLDFF1SB P7.Z, (R30)(Z31.S.UXTW), [Z31.S]    // df3f1f84
    ZLDFF1SB P0.Z, (R0)(Z0.S.SXTW), [Z0.S]       // 00204084
    ZLDFF1SB P3.Z, (R12)(Z13.S.SXTW), [Z10.S]    // 8a2d4d84
    ZLDFF1SB P7.Z, (R30)(Z31.S.SXTW), [Z31.S]    // df3f5f84

// LDFF1SH { <Zt>.D }, <Pg>/Z, [<Zn>.D{, #<pimm>}]
    ZLDFF1SH P0.Z, (Z0.D), [Z0.D]       // 00a0a0c4
    ZLDFF1SH P3.Z, 20(Z12.D), [Z10.D]    // 8aadaac4
    ZLDFF1SH P7.Z, 62(Z31.D), [Z31.D]    // ffbfbfc4

// LDFF1SH { <Zt>.S }, <Pg>/Z, [<Zn>.S{, #<pimm>}]
    ZLDFF1SH P0.Z, (Z0.S), [Z0.S]       // 00a0a084
    ZLDFF1SH P3.Z, 20(Z12.S), [Z10.S]    // 8aadaa84
    ZLDFF1SH P7.Z, 62(Z31.S), [Z31.S]    // ffbfbf84

// LDFF1SH { <Zt>.S }, <Pg>/Z, [<Xn|SP>{, <Xm>, LSL #1}]
    ZLDFF1SH P0.Z, (R0)(R0<<1), [Z0.S]       // 006020a5
    ZLDFF1SH P3.Z, (R12)(R13<<1), [Z10.S]    // 8a6d2da5
    ZLDFF1SH P7.Z, (R30)(R30<<1), [Z31.S]    // df7f3ea5

// LDFF1SH { <Zt>.D }, <Pg>/Z, [<Xn|SP>{, <Xm>, LSL #1}]
    ZLDFF1SH P0.Z, (R0)(R0<<1), [Z0.D]       // 006000a5
    ZLDFF1SH P3.Z, (R12)(R13<<1), [Z10.D]    // 8a6d0da5
    ZLDFF1SH P7.Z, (R30)(R30<<1), [Z31.D]    // df7f1ea5

// LDFF1SH { <Zt>.D }, <Pg>/Z, [<Xn|SP>, <Zm>.D, LSL #1]
    ZLDFF1SH P0.Z, (R0)(Z0.D.LSL<<1), [Z0.D]       // 00a0e0c4
    ZLDFF1SH P3.Z, (R12)(Z13.D.LSL<<1), [Z10.D]    // 8aadedc4
    ZLDFF1SH P7.Z, (R30)(Z31.D.LSL<<1), [Z31.D]    // dfbfffc4

// LDFF1SH { <Zt>.D }, <Pg>/Z, [<Xn|SP>, <Zm>.D]
    ZLDFF1SH P0.Z, (R0)(Z0.D), [Z0.D]       // 00a0c0c4
    ZLDFF1SH P3.Z, (R12)(Z13.D), [Z10.D]    // 8aadcdc4
    ZLDFF1SH P7.Z, (R30)(Z31.D), [Z31.D]    // dfbfdfc4

// LDFF1SH { <Zt>.D }, <Pg>/Z, [<Xn|SP>, <Zm>.D, <extend> #1]
    ZLDFF1SH P0.Z, (R0)(Z0.D.UXTW<<1), [Z0.D]       // 0020a0c4
    ZLDFF1SH P3.Z, (R12)(Z13.D.UXTW<<1), [Z10.D]    // 8a2dadc4
    ZLDFF1SH P7.Z, (R30)(Z31.D.UXTW<<1), [Z31.D]    // df3fbfc4
    ZLDFF1SH P0.Z, (R0)(Z0.D.SXTW<<1), [Z0.D]       // 0020e0c4
    ZLDFF1SH P3.Z, (R12)(Z13.D.SXTW<<1), [Z10.D]    // 8a2dedc4
    ZLDFF1SH P7.Z, (R30)(Z31.D.SXTW<<1), [Z31.D]    // df3fffc4

// LDFF1SH { <Zt>.D }, <Pg>/Z, [<Xn|SP>, <Zm>.D, <extend>]
    ZLDFF1SH P0.Z, (R0)(Z0.D.UXTW), [Z0.D]       // 002080c4
    ZLDFF1SH P3.Z, (R12)(Z13.D.UXTW), [Z10.D]    // 8a2d8dc4
    ZLDFF1SH P7.Z, (R30)(Z31.D.UXTW), [Z31.D]    // df3f9fc4
    ZLDFF1SH P0.Z, (R0)(Z0.D.SXTW), [Z0.D]       // 0020c0c4
    ZLDFF1SH P3.Z, (R12)(Z13.D.SXTW), [Z10.D]    // 8a2dcdc4
    ZLDFF1SH P7.Z, (R30)(Z31.D.SXTW), [Z31.D]    // df3fdfc4

// LDFF1SH { <Zt>.S }, <Pg>/Z, [<Xn|SP>, <Zm>.S, <extend> #1]
    ZLDFF1SH P0.Z, (R0)(Z0.S.UXTW<<1), [Z0.S]       // 0020a084
    ZLDFF1SH P3.Z, (R12)(Z13.S.UXTW<<1), [Z10.S]    // 8a2dad84
    ZLDFF1SH P7.Z, (R30)(Z31.S.UXTW<<1), [Z31.S]    // df3fbf84
    ZLDFF1SH P0.Z, (R0)(Z0.S.SXTW<<1), [Z0.S]       // 0020e084
    ZLDFF1SH P3.Z, (R12)(Z13.S.SXTW<<1), [Z10.S]    // 8a2ded84
    ZLDFF1SH P7.Z, (R30)(Z31.S.SXTW<<1), [Z31.S]    // df3fff84

// LDFF1SH { <Zt>.S }, <Pg>/Z, [<Xn|SP>, <Zm>.S, <extend>]
    ZLDFF1SH P0.Z, (R0)(Z0.S.UXTW), [Z0.S]       // 00208084
    ZLDFF1SH P3.Z, (R12)(Z13.S.UXTW), [Z10.S]    // 8a2d8d84
    ZLDFF1SH P7.Z, (R30)(Z31.S.UXTW), [Z31.S]    // df3f9f84
    ZLDFF1SH P0.Z, (R0)(Z0.S.SXTW), [Z0.S]       // 0020c084
    ZLDFF1SH P3.Z, (R12)(Z13.S.SXTW), [Z10.S]    // 8a2dcd84
    ZLDFF1SH P7.Z, (R30)(Z31.S.SXTW), [Z31.S]    // df3fdf84

// LDFF1SW { <Zt>.D }, <Pg>/Z, [<Zn>.D{, #<pimm>}]
    ZLDFF1SW P0.Z, (Z0.D), [Z0.D]        // 00a020c5
    ZLDFF1SW P3.Z, 40(Z12.D), [Z10.D]     // 8aad2ac5
    ZLDFF1SW P7.Z, 124(Z31.D), [Z31.D]    // ffbf3fc5

// LDFF1SW { <Zt>.D }, <Pg>/Z, [<Xn|SP>{, <Xm>, LSL #2}]
    ZLDFF1SW P0.Z, (R0)(R0<<2), [Z0.D]       // 006080a4
    ZLDFF1SW P3.Z, (R12)(R13<<2), [Z10.D]    // 8a6d8da4
    ZLDFF1SW P7.Z, (R30)(R30<<2), [Z31.D]    // df7f9ea4

// LDFF1SW { <Zt>.D }, <Pg>/Z, [<Xn|SP>, <Zm>.D, LSL #2]
    ZLDFF1SW P0.Z, (R0)(Z0.D.LSL<<2), [Z0.D]       // 00a060c5
    ZLDFF1SW P3.Z, (R12)(Z13.D.LSL<<2), [Z10.D]    // 8aad6dc5
    ZLDFF1SW P7.Z, (R30)(Z31.D.LSL<<2), [Z31.D]    // dfbf7fc5

// LDFF1SW { <Zt>.D }, <Pg>/Z, [<Xn|SP>, <Zm>.D]
    ZLDFF1SW P0.Z, (R0)(Z0.D), [Z0.D]       // 00a040c5
    ZLDFF1SW P3.Z, (R12)(Z13.D), [Z10.D]    // 8aad4dc5
    ZLDFF1SW P7.Z, (R30)(Z31.D), [Z31.D]    // dfbf5fc5

// LDFF1SW { <Zt>.D }, <Pg>/Z, [<Xn|SP>, <Zm>.D, <extend> #2]
    ZLDFF1SW P0.Z, (R0)(Z0.D.UXTW<<2), [Z0.D]       // 002020c5
    ZLDFF1SW P3.Z, (R12)(Z13.D.UXTW<<2), [Z10.D]    // 8a2d2dc5
    ZLDFF1SW P7.Z, (R30)(Z31.D.UXTW<<2), [Z31.D]    // df3f3fc5
    ZLDFF1SW P0.Z, (R0)(Z0.D.SXTW<<2), [Z0.D]       // 002060c5
    ZLDFF1SW P3.Z, (R12)(Z13.D.SXTW<<2), [Z10.D]    // 8a2d6dc5
    ZLDFF1SW P7.Z, (R30)(Z31.D.SXTW<<2), [Z31.D]    // df3f7fc5

// LDFF1SW { <Zt>.D }, <Pg>/Z, [<Xn|SP>, <Zm>.D, <extend>]
    ZLDFF1SW P0.Z, (R0)(Z0.D.UXTW), [Z0.D]       // 002000c5
    ZLDFF1SW P3.Z, (R12)(Z13.D.UXTW), [Z10.D]    // 8a2d0dc5
    ZLDFF1SW P7.Z, (R30)(Z31.D.UXTW), [Z31.D]    // df3f1fc5
    ZLDFF1SW P0.Z, (R0)(Z0.D.SXTW), [Z0.D]       // 002040c5
    ZLDFF1SW P3.Z, (R12)(Z13.D.SXTW), [Z10.D]    // 8a2d4dc5
    ZLDFF1SW P7.Z, (R30)(Z31.D.SXTW), [Z31.D]    // df3f5fc5

// LDFF1W  { <Zt>.D }, <Pg>/Z, [<Zn>.D{, #<pimm>}]
    ZLDFF1W P0.Z, (Z0.D), [Z0.D]        // 00e020c5
    ZLDFF1W P3.Z, 40(Z12.D), [Z10.D]     // 8aed2ac5
    ZLDFF1W P7.Z, 124(Z31.D), [Z31.D]    // ffff3fc5

// LDFF1W  { <Zt>.S }, <Pg>/Z, [<Zn>.S{, #<pimm>}]
    ZLDFF1W P0.Z, (Z0.S), [Z0.S]        // 00e02085
    ZLDFF1W P3.Z, 40(Z12.S), [Z10.S]     // 8aed2a85
    ZLDFF1W P7.Z, 124(Z31.S), [Z31.S]    // ffff3f85

// LDFF1W  { <Zt>.S }, <Pg>/Z, [<Xn|SP>{, <Xm>, LSL #2}]
    ZLDFF1W P0.Z, (R0)(R0<<2), [Z0.S]       // 006040a5
    ZLDFF1W P3.Z, (R12)(R13<<2), [Z10.S]    // 8a6d4da5
    ZLDFF1W P7.Z, (R30)(R30<<2), [Z31.S]    // df7f5ea5

// LDFF1W  { <Zt>.D }, <Pg>/Z, [<Xn|SP>{, <Xm>, LSL #2}]
    ZLDFF1W P0.Z, (R0)(R0<<2), [Z0.D]       // 006060a5
    ZLDFF1W P3.Z, (R12)(R13<<2), [Z10.D]    // 8a6d6da5
    ZLDFF1W P7.Z, (R30)(R30<<2), [Z31.D]    // df7f7ea5

// LDFF1W  { <Zt>.D }, <Pg>/Z, [<Xn|SP>, <Zm>.D, LSL #2]
    ZLDFF1W P0.Z, (R0)(Z0.D.LSL<<2), [Z0.D]       // 00e060c5
    ZLDFF1W P3.Z, (R12)(Z13.D.LSL<<2), [Z10.D]    // 8aed6dc5
    ZLDFF1W P7.Z, (R30)(Z31.D.LSL<<2), [Z31.D]    // dfff7fc5

// LDFF1W  { <Zt>.D }, <Pg>/Z, [<Xn|SP>, <Zm>.D]
    ZLDFF1W P0.Z, (R0)(Z0.D), [Z0.D]       // 00e040c5
    ZLDFF1W P3.Z, (R12)(Z13.D), [Z10.D]    // 8aed4dc5
    ZLDFF1W P7.Z, (R30)(Z31.D), [Z31.D]    // dfff5fc5

// LDFF1W  { <Zt>.D }, <Pg>/Z, [<Xn|SP>, <Zm>.D, <extend> #2]
    ZLDFF1W P0.Z, (R0)(Z0.D.UXTW<<2), [Z0.D]       // 006020c5
    ZLDFF1W P3.Z, (R12)(Z13.D.UXTW<<2), [Z10.D]    // 8a6d2dc5
    ZLDFF1W P7.Z, (R30)(Z31.D.UXTW<<2), [Z31.D]    // df7f3fc5
    ZLDFF1W P0.Z, (R0)(Z0.D.SXTW<<2), [Z0.D]       // 006060c5
    ZLDFF1W P3.Z, (R12)(Z13.D.SXTW<<2), [Z10.D]    // 8a6d6dc5
    ZLDFF1W P7.Z, (R30)(Z31.D.SXTW<<2), [Z31.D]    // df7f7fc5

// LDFF1W  { <Zt>.D }, <Pg>/Z, [<Xn|SP>, <Zm>.D, <extend>]
    ZLDFF1W P0.Z, (R0)(Z0.D.UXTW), [Z0.D]       // 006000c5
    ZLDFF1W P3.Z, (R12)(Z13.D.UXTW), [Z10.D]    // 8a6d0dc5
    ZLDFF1W P7.Z, (R30)(Z31.D.UXTW), [Z31.D]    // df7f1fc5
    ZLDFF1W P0.Z, (R0)(Z0.D.SXTW), [Z0.D]       // 006040c5
    ZLDFF1W P3.Z, (R12)(Z13.D.SXTW), [Z10.D]    // 8a6d4dc5
    ZLDFF1W P7.Z, (R30)(Z31.D.SXTW), [Z31.D]    // df7f5fc5

// LDFF1W  { <Zt>.S }, <Pg>/Z, [<Xn|SP>, <Zm>.S, <extend> #2]
    ZLDFF1W P0.Z, (R0)(Z0.S.UXTW<<2), [Z0.S]       // 00602085
    ZLDFF1W P3.Z, (R12)(Z13.S.UXTW<<2), [Z10.S]    // 8a6d2d85
    ZLDFF1W P7.Z, (R30)(Z31.S.UXTW<<2), [Z31.S]    // df7f3f85
    ZLDFF1W P0.Z, (R0)(Z0.S.SXTW<<2), [Z0.S]       // 00606085
    ZLDFF1W P3.Z, (R12)(Z13.S.SXTW<<2), [Z10.S]    // 8a6d6d85
    ZLDFF1W P7.Z, (R30)(Z31.S.SXTW<<2), [Z31.S]    // df7f7f85

// LDFF1W  { <Zt>.S }, <Pg>/Z, [<Xn|SP>, <Zm>.S, <extend>]
    ZLDFF1W P0.Z, (R0)(Z0.S.UXTW), [Z0.S]       // 00600085
    ZLDFF1W P3.Z, (R12)(Z13.S.UXTW), [Z10.S]    // 8a6d0d85
    ZLDFF1W P7.Z, (R30)(Z31.S.UXTW), [Z31.S]    // df7f1f85
    ZLDFF1W P0.Z, (R0)(Z0.S.SXTW), [Z0.S]       // 00604085
    ZLDFF1W P3.Z, (R12)(Z13.S.SXTW), [Z10.S]    // 8a6d4d85
    ZLDFF1W P7.Z, (R30)(Z31.S.SXTW), [Z31.S]    // df7f5f85

// LDNF1B  { <Zt>.H }, <Pg>/Z, [<Xn|SP>{, #<simm>, MUL VL}]
    ZLDNF1B P0.Z, -8(R0), [Z0.H]      // 00a038a4
    ZLDNF1B P3.Z, -3(R12), [Z10.H]    // 8aad3da4
    ZLDNF1B P7.Z, 7(R30), [Z31.H]     // dfbf37a4

// LDNF1B  { <Zt>.S }, <Pg>/Z, [<Xn|SP>{, #<simm>, MUL VL}]
    ZLDNF1B P0.Z, -8(R0), [Z0.S]      // 00a058a4
    ZLDNF1B P3.Z, -3(R12), [Z10.S]    // 8aad5da4
    ZLDNF1B P7.Z, 7(R30), [Z31.S]     // dfbf57a4

// LDNF1B  { <Zt>.D }, <Pg>/Z, [<Xn|SP>{, #<simm>, MUL VL}]
    ZLDNF1B P0.Z, -8(R0), [Z0.D]      // 00a078a4
    ZLDNF1B P3.Z, -3(R12), [Z10.D]    // 8aad7da4
    ZLDNF1B P7.Z, 7(R30), [Z31.D]     // dfbf77a4

// LDNF1B  { <Zt>.B }, <Pg>/Z, [<Xn|SP>{, #<simm>, MUL VL}]
    ZLDNF1B P0.Z, -8(R0), [Z0.B]      // 00a018a4
    ZLDNF1B P3.Z, -3(R12), [Z10.B]    // 8aad1da4
    ZLDNF1B P7.Z, 7(R30), [Z31.B]     // dfbf17a4

// LDNF1D  { <Zt>.D }, <Pg>/Z, [<Xn|SP>{, #<simm>, MUL VL}]
    ZLDNF1D P0.Z, -8(R0), [Z0.D]      // 00a0f8a5
    ZLDNF1D P3.Z, -3(R12), [Z10.D]    // 8aadfda5
    ZLDNF1D P7.Z, 7(R30), [Z31.D]     // dfbff7a5

// LDNF1H  { <Zt>.H }, <Pg>/Z, [<Xn|SP>{, #<simm>, MUL VL}]
    ZLDNF1H P0.Z, -8(R0), [Z0.H]      // 00a0b8a4
    ZLDNF1H P3.Z, -3(R12), [Z10.H]    // 8aadbda4
    ZLDNF1H P7.Z, 7(R30), [Z31.H]     // dfbfb7a4

// LDNF1H  { <Zt>.S }, <Pg>/Z, [<Xn|SP>{, #<simm>, MUL VL}]
    ZLDNF1H P0.Z, -8(R0), [Z0.S]      // 00a0d8a4
    ZLDNF1H P3.Z, -3(R12), [Z10.S]    // 8aaddda4
    ZLDNF1H P7.Z, 7(R30), [Z31.S]     // dfbfd7a4

// LDNF1H  { <Zt>.D }, <Pg>/Z, [<Xn|SP>{, #<simm>, MUL VL}]
    ZLDNF1H P0.Z, -8(R0), [Z0.D]      // 00a0f8a4
    ZLDNF1H P3.Z, -3(R12), [Z10.D]    // 8aadfda4
    ZLDNF1H P7.Z, 7(R30), [Z31.D]     // dfbff7a4

// LDNF1SB { <Zt>.H }, <Pg>/Z, [<Xn|SP>{, #<simm>, MUL VL}]
    ZLDNF1SB P0.Z, -8(R0), [Z0.H]      // 00a0d8a5
    ZLDNF1SB P3.Z, -3(R12), [Z10.H]    // 8aaddda5
    ZLDNF1SB P7.Z, 7(R30), [Z31.H]     // dfbfd7a5

// LDNF1SB { <Zt>.S }, <Pg>/Z, [<Xn|SP>{, #<simm>, MUL VL}]
    ZLDNF1SB P0.Z, -8(R0), [Z0.S]      // 00a0b8a5
    ZLDNF1SB P3.Z, -3(R12), [Z10.S]    // 8aadbda5
    ZLDNF1SB P7.Z, 7(R30), [Z31.S]     // dfbfb7a5

// LDNF1SB { <Zt>.D }, <Pg>/Z, [<Xn|SP>{, #<simm>, MUL VL}]
    ZLDNF1SB P0.Z, -8(R0), [Z0.D]      // 00a098a5
    ZLDNF1SB P3.Z, -3(R12), [Z10.D]    // 8aad9da5
    ZLDNF1SB P7.Z, 7(R30), [Z31.D]     // dfbf97a5

// LDNF1SH { <Zt>.S }, <Pg>/Z, [<Xn|SP>{, #<simm>, MUL VL}]
    ZLDNF1SH P0.Z, -8(R0), [Z0.S]      // 00a038a5
    ZLDNF1SH P3.Z, -3(R12), [Z10.S]    // 8aad3da5
    ZLDNF1SH P7.Z, 7(R30), [Z31.S]     // dfbf37a5

// LDNF1SH { <Zt>.D }, <Pg>/Z, [<Xn|SP>{, #<simm>, MUL VL}]
    ZLDNF1SH P0.Z, -8(R0), [Z0.D]      // 00a018a5
    ZLDNF1SH P3.Z, -3(R12), [Z10.D]    // 8aad1da5
    ZLDNF1SH P7.Z, 7(R30), [Z31.D]     // dfbf17a5

// LDNF1SW { <Zt>.D }, <Pg>/Z, [<Xn|SP>{, #<simm>, MUL VL}]
    ZLDNF1SW P0.Z, -8(R0), [Z0.D]      // 00a098a4
    ZLDNF1SW P3.Z, -3(R12), [Z10.D]    // 8aad9da4
    ZLDNF1SW P7.Z, 7(R30), [Z31.D]     // dfbf97a4

// LDNF1W  { <Zt>.S }, <Pg>/Z, [<Xn|SP>{, #<simm>, MUL VL}]
    ZLDNF1W P0.Z, -8(R0), [Z0.S]      // 00a058a5
    ZLDNF1W P3.Z, -3(R12), [Z10.S]    // 8aad5da5
    ZLDNF1W P7.Z, 7(R30), [Z31.S]     // dfbf57a5

// LDNF1W  { <Zt>.D }, <Pg>/Z, [<Xn|SP>{, #<simm>, MUL VL}]
    ZLDNF1W P0.Z, -8(R0), [Z0.D]      // 00a078a5
    ZLDNF1W P3.Z, -3(R12), [Z10.D]    // 8aad7da5
    ZLDNF1W P7.Z, 7(R30), [Z31.D]     // dfbf77a5

// LDNT1B  { <Zt>.B }, <Pg>/Z, [<Xn|SP>{, #<simm>, MUL VL}]
    ZLDNT1B P0.Z, -8(R0), [Z0.B]      // 00e008a4
    ZLDNT1B P3.Z, -3(R12), [Z10.B]    // 8aed0da4
    ZLDNT1B P7.Z, 7(R30), [Z31.B]     // dfff07a4

// LDNT1B  { <Zt>.B }, <Pg>/Z, [<Xn|SP>, <Xm>]
    ZLDNT1B P0.Z, (R0)(R0), [Z0.B]       // 00c000a4
    ZLDNT1B P3.Z, (R12)(R13), [Z10.B]    // 8acd0da4
    ZLDNT1B P7.Z, (R30)(R30), [Z31.B]    // dfdf1ea4

// LDNT1D  { <Zt>.D }, <Pg>/Z, [<Xn|SP>{, #<simm>, MUL VL}]
    ZLDNT1D P0.Z, -8(R0), [Z0.D]      // 00e088a5
    ZLDNT1D P3.Z, -3(R12), [Z10.D]    // 8aed8da5
    ZLDNT1D P7.Z, 7(R30), [Z31.D]     // dfff87a5

// LDNT1D  { <Zt>.D }, <Pg>/Z, [<Xn|SP>, <Xm>, LSL #3]
    ZLDNT1D P0.Z, (R0)(R0<<3), [Z0.D]       // 00c080a5
    ZLDNT1D P3.Z, (R12)(R13<<3), [Z10.D]    // 8acd8da5
    ZLDNT1D P7.Z, (R30)(R30<<3), [Z31.D]    // dfdf9ea5

// LDNT1H  { <Zt>.H }, <Pg>/Z, [<Xn|SP>{, #<simm>, MUL VL}]
    ZLDNT1H P0.Z, -8(R0), [Z0.H]      // 00e088a4
    ZLDNT1H P3.Z, -3(R12), [Z10.H]    // 8aed8da4
    ZLDNT1H P7.Z, 7(R30), [Z31.H]     // dfff87a4

// LDNT1H  { <Zt>.H }, <Pg>/Z, [<Xn|SP>, <Xm>, LSL #1]
    ZLDNT1H P0.Z, (R0)(R0<<1), [Z0.H]       // 00c080a4
    ZLDNT1H P3.Z, (R12)(R13<<1), [Z10.H]    // 8acd8da4
    ZLDNT1H P7.Z, (R30)(R30<<1), [Z31.H]    // dfdf9ea4

// LDNT1W  { <Zt>.S }, <Pg>/Z, [<Xn|SP>{, #<simm>, MUL VL}]
    ZLDNT1W P0.Z, -8(R0), [Z0.S]      // 00e008a5
    ZLDNT1W P3.Z, -3(R12), [Z10.S]    // 8aed0da5
    ZLDNT1W P7.Z, 7(R30), [Z31.S]     // dfff07a5

// LDNT1W  { <Zt>.S }, <Pg>/Z, [<Xn|SP>, <Xm>, LSL #2]
    ZLDNT1W P0.Z, (R0)(R0<<2), [Z0.S]       // 00c000a5
    ZLDNT1W P3.Z, (R12)(R13<<2), [Z10.S]    // 8acd0da5
    ZLDNT1W P7.Z, (R30)(R30<<2), [Z31.S]    // dfdf1ea5

// LDR     <Pt>, [<Xn|SP>{, #<simm>, MUL VL}]
    PLDR -256(R0), P0     // 0000a085
    PLDR -86(R11), P5     // 6509b585
    PLDR 255(R30), P15    // cf1f9f85

// LDR     <Zt>, [<Xn|SP>{, #<simm>, MUL VL}]
    ZLDR -256(R0), Z0     // 0040a085
    ZLDR -86(R11), Z10    // 6a49b585
    ZLDR 255(R30), Z31    // df5f9f85

// LSL     <Zdn>.<T>, <Pg>/M, <Zdn>.<T>, #<const>
    ZLSL P0.M, Z0.B, $0, Z0.B       // 00810304
    ZLSL P3.M, Z10.B, $2, Z10.B     // 4a8d0304
    ZLSL P7.M, Z31.B, $7, Z31.B     // ff9d0304
    ZLSL P0.M, Z0.H, $0, Z0.H       // 00820304
    ZLSL P3.M, Z10.H, $5, Z10.H     // aa8e0304
    ZLSL P7.M, Z31.H, $15, Z31.H    // ff9f0304
    ZLSL P0.M, Z0.S, $0, Z0.S       // 00804304
    ZLSL P3.M, Z10.S, $10, Z10.S    // 4a8d4304
    ZLSL P7.M, Z31.S, $31, Z31.S    // ff9f4304
    ZLSL P0.M, Z0.D, $0, Z0.D       // 00808304
    ZLSL P3.M, Z10.D, $21, Z10.D    // aa8e8304
    ZLSL P7.M, Z31.D, $63, Z31.D    // ff9fc304

// LSL     <Zdn>.<T>, <Pg>/M, <Zdn>.<T>, <Zm>.D
    ZLSL P0.M, Z0.B, Z0.D, Z0.B       // 00801b04
    ZLSL P3.M, Z10.B, Z12.D, Z10.B    // 8a8d1b04
    ZLSL P7.M, Z31.B, Z31.D, Z31.B    // ff9f1b04
    ZLSL P0.M, Z0.H, Z0.D, Z0.H       // 00805b04
    ZLSL P3.M, Z10.H, Z12.D, Z10.H    // 8a8d5b04
    ZLSL P7.M, Z31.H, Z31.D, Z31.H    // ff9f5b04
    ZLSL P0.M, Z0.S, Z0.D, Z0.S       // 00809b04
    ZLSL P3.M, Z10.S, Z12.D, Z10.S    // 8a8d9b04
    ZLSL P7.M, Z31.S, Z31.D, Z31.S    // ff9f9b04

// LSL     <Zdn>.<T>, <Pg>/M, <Zdn>.<T>, <Zm>.<T>
    ZLSL P0.M, Z0.B, Z0.B, Z0.B       // 00801304
    ZLSL P3.M, Z10.B, Z12.B, Z10.B    // 8a8d1304
    ZLSL P7.M, Z31.B, Z31.B, Z31.B    // ff9f1304
    ZLSL P0.M, Z0.H, Z0.H, Z0.H       // 00805304
    ZLSL P3.M, Z10.H, Z12.H, Z10.H    // 8a8d5304
    ZLSL P7.M, Z31.H, Z31.H, Z31.H    // ff9f5304
    ZLSL P0.M, Z0.S, Z0.S, Z0.S       // 00809304
    ZLSL P3.M, Z10.S, Z12.S, Z10.S    // 8a8d9304
    ZLSL P7.M, Z31.S, Z31.S, Z31.S    // ff9f9304
    ZLSL P0.M, Z0.D, Z0.D, Z0.D       // 0080d304
    ZLSL P3.M, Z10.D, Z12.D, Z10.D    // 8a8dd304
    ZLSL P7.M, Z31.D, Z31.D, Z31.D    // ff9fd304

// LSL     <Zd>.<T>, <Zn>.<T>, #<const>
    ZLSL Z0.B, $0, Z0.B       // 009c2804
    ZLSL Z11.B, $2, Z10.B     // 6a9d2a04
    ZLSL Z31.B, $7, Z31.B     // ff9f2f04
    ZLSL Z0.H, $0, Z0.H       // 009c3004
    ZLSL Z11.H, $5, Z10.H     // 6a9d3504
    ZLSL Z31.H, $15, Z31.H    // ff9f3f04
    ZLSL Z0.S, $0, Z0.S       // 009c6004
    ZLSL Z11.S, $10, Z10.S    // 6a9d6a04
    ZLSL Z31.S, $31, Z31.S    // ff9f7f04
    ZLSL Z0.D, $0, Z0.D       // 009ca004
    ZLSL Z11.D, $21, Z10.D    // 6a9db504
    ZLSL Z31.D, $63, Z31.D    // ff9fff04

// LSL     <Zd>.<T>, <Zn>.<T>, <Zm>.D
    ZLSL Z0.B, Z0.D, Z0.B       // 008c2004
    ZLSL Z11.B, Z12.D, Z10.B    // 6a8d2c04
    ZLSL Z31.B, Z31.D, Z31.B    // ff8f3f04
    ZLSL Z0.H, Z0.D, Z0.H       // 008c6004
    ZLSL Z11.H, Z12.D, Z10.H    // 6a8d6c04
    ZLSL Z31.H, Z31.D, Z31.H    // ff8f7f04
    ZLSL Z0.S, Z0.D, Z0.S       // 008ca004
    ZLSL Z11.S, Z12.D, Z10.S    // 6a8dac04
    ZLSL Z31.S, Z31.D, Z31.S    // ff8fbf04

// LSLR    <Zdn>.<T>, <Pg>/M, <Zdn>.<T>, <Zm>.<T>
    ZLSLR P0.M, Z0.B, Z0.B, Z0.B       // 00801704
    ZLSLR P3.M, Z10.B, Z12.B, Z10.B    // 8a8d1704
    ZLSLR P7.M, Z31.B, Z31.B, Z31.B    // ff9f1704
    ZLSLR P0.M, Z0.H, Z0.H, Z0.H       // 00805704
    ZLSLR P3.M, Z10.H, Z12.H, Z10.H    // 8a8d5704
    ZLSLR P7.M, Z31.H, Z31.H, Z31.H    // ff9f5704
    ZLSLR P0.M, Z0.S, Z0.S, Z0.S       // 00809704
    ZLSLR P3.M, Z10.S, Z12.S, Z10.S    // 8a8d9704
    ZLSLR P7.M, Z31.S, Z31.S, Z31.S    // ff9f9704
    ZLSLR P0.M, Z0.D, Z0.D, Z0.D       // 0080d704
    ZLSLR P3.M, Z10.D, Z12.D, Z10.D    // 8a8dd704
    ZLSLR P7.M, Z31.D, Z31.D, Z31.D    // ff9fd704

// LSR     <Zdn>.<T>, <Pg>/M, <Zdn>.<T>, #<const>
    ZLSR P0.M, Z0.B, $1, Z0.B       // e0810104
    ZLSR P3.M, Z10.B, $3, Z10.B     // aa8d0104
    ZLSR P7.M, Z31.B, $8, Z31.B     // 1f9d0104
    ZLSR P0.M, Z0.H, $1, Z0.H       // e0830104
    ZLSR P3.M, Z10.H, $6, Z10.H     // 4a8f0104
    ZLSR P7.M, Z31.H, $16, Z31.H    // 1f9e0104
    ZLSR P0.M, Z0.S, $1, Z0.S       // e0834104
    ZLSR P3.M, Z10.S, $11, Z10.S    // aa8e4104
    ZLSR P7.M, Z31.S, $32, Z31.S    // 1f9c4104
    ZLSR P0.M, Z0.D, $1, Z0.D       // e083c104
    ZLSR P3.M, Z10.D, $22, Z10.D    // 4a8dc104
    ZLSR P7.M, Z31.D, $64, Z31.D    // 1f9c8104

// LSR     <Zdn>.<T>, <Pg>/M, <Zdn>.<T>, <Zm>.D
    ZLSR P0.M, Z0.B, Z0.D, Z0.B       // 00801904
    ZLSR P3.M, Z10.B, Z12.D, Z10.B    // 8a8d1904
    ZLSR P7.M, Z31.B, Z31.D, Z31.B    // ff9f1904
    ZLSR P0.M, Z0.H, Z0.D, Z0.H       // 00805904
    ZLSR P3.M, Z10.H, Z12.D, Z10.H    // 8a8d5904
    ZLSR P7.M, Z31.H, Z31.D, Z31.H    // ff9f5904
    ZLSR P0.M, Z0.S, Z0.D, Z0.S       // 00809904
    ZLSR P3.M, Z10.S, Z12.D, Z10.S    // 8a8d9904
    ZLSR P7.M, Z31.S, Z31.D, Z31.S    // ff9f9904

// LSR     <Zdn>.<T>, <Pg>/M, <Zdn>.<T>, <Zm>.<T>
    ZLSR P0.M, Z0.B, Z0.B, Z0.B       // 00801104
    ZLSR P3.M, Z10.B, Z12.B, Z10.B    // 8a8d1104
    ZLSR P7.M, Z31.B, Z31.B, Z31.B    // ff9f1104
    ZLSR P0.M, Z0.H, Z0.H, Z0.H       // 00805104
    ZLSR P3.M, Z10.H, Z12.H, Z10.H    // 8a8d5104
    ZLSR P7.M, Z31.H, Z31.H, Z31.H    // ff9f5104
    ZLSR P0.M, Z0.S, Z0.S, Z0.S       // 00809104
    ZLSR P3.M, Z10.S, Z12.S, Z10.S    // 8a8d9104
    ZLSR P7.M, Z31.S, Z31.S, Z31.S    // ff9f9104
    ZLSR P0.M, Z0.D, Z0.D, Z0.D       // 0080d104
    ZLSR P3.M, Z10.D, Z12.D, Z10.D    // 8a8dd104
    ZLSR P7.M, Z31.D, Z31.D, Z31.D    // ff9fd104

// LSR     <Zd>.<T>, <Zn>.<T>, #<const>
    ZLSR Z0.B, $1, Z0.B       // 00942f04
    ZLSR Z11.B, $3, Z10.B     // 6a952d04
    ZLSR Z31.B, $8, Z31.B     // ff972804
    ZLSR Z0.H, $1, Z0.H       // 00943f04
    ZLSR Z11.H, $6, Z10.H     // 6a953a04
    ZLSR Z31.H, $16, Z31.H    // ff973004
    ZLSR Z0.S, $1, Z0.S       // 00947f04
    ZLSR Z11.S, $11, Z10.S    // 6a957504
    ZLSR Z31.S, $32, Z31.S    // ff976004
    ZLSR Z0.D, $1, Z0.D       // 0094ff04
    ZLSR Z11.D, $22, Z10.D    // 6a95ea04
    ZLSR Z31.D, $64, Z31.D    // ff97a004

// LSR     <Zd>.<T>, <Zn>.<T>, <Zm>.D
    ZLSR Z0.B, Z0.D, Z0.B       // 00842004
    ZLSR Z11.B, Z12.D, Z10.B    // 6a852c04
    ZLSR Z31.B, Z31.D, Z31.B    // ff873f04
    ZLSR Z0.H, Z0.D, Z0.H       // 00846004
    ZLSR Z11.H, Z12.D, Z10.H    // 6a856c04
    ZLSR Z31.H, Z31.D, Z31.H    // ff877f04
    ZLSR Z0.S, Z0.D, Z0.S       // 0084a004
    ZLSR Z11.S, Z12.D, Z10.S    // 6a85ac04
    ZLSR Z31.S, Z31.D, Z31.S    // ff87bf04

// LSRR    <Zdn>.<T>, <Pg>/M, <Zdn>.<T>, <Zm>.<T>
    ZLSRR P0.M, Z0.B, Z0.B, Z0.B       // 00801504
    ZLSRR P3.M, Z10.B, Z12.B, Z10.B    // 8a8d1504
    ZLSRR P7.M, Z31.B, Z31.B, Z31.B    // ff9f1504
    ZLSRR P0.M, Z0.H, Z0.H, Z0.H       // 00805504
    ZLSRR P3.M, Z10.H, Z12.H, Z10.H    // 8a8d5504
    ZLSRR P7.M, Z31.H, Z31.H, Z31.H    // ff9f5504
    ZLSRR P0.M, Z0.S, Z0.S, Z0.S       // 00809504
    ZLSRR P3.M, Z10.S, Z12.S, Z10.S    // 8a8d9504
    ZLSRR P7.M, Z31.S, Z31.S, Z31.S    // ff9f9504
    ZLSRR P0.M, Z0.D, Z0.D, Z0.D       // 0080d504
    ZLSRR P3.M, Z10.D, Z12.D, Z10.D    // 8a8dd504
    ZLSRR P7.M, Z31.D, Z31.D, Z31.D    // ff9fd504

// MAD     <Zdn>.<T>, <Pg>/M, <Zm>.<T>, <Za>.<T>
    ZMAD P0.M, Z0.B, Z0.B, Z0.B       // 00c00004
    ZMAD P3.M, Z12.B, Z13.B, Z10.B    // aacd0c04
    ZMAD P7.M, Z31.B, Z31.B, Z31.B    // ffdf1f04
    ZMAD P0.M, Z0.H, Z0.H, Z0.H       // 00c04004
    ZMAD P3.M, Z12.H, Z13.H, Z10.H    // aacd4c04
    ZMAD P7.M, Z31.H, Z31.H, Z31.H    // ffdf5f04
    ZMAD P0.M, Z0.S, Z0.S, Z0.S       // 00c08004
    ZMAD P3.M, Z12.S, Z13.S, Z10.S    // aacd8c04
    ZMAD P7.M, Z31.S, Z31.S, Z31.S    // ffdf9f04
    ZMAD P0.M, Z0.D, Z0.D, Z0.D       // 00c0c004
    ZMAD P3.M, Z12.D, Z13.D, Z10.D    // aacdcc04
    ZMAD P7.M, Z31.D, Z31.D, Z31.D    // ffdfdf04

// MLA     <Zda>.<T>, <Pg>/M, <Zn>.<T>, <Zm>.<T>
    ZMLA P0.M, Z0.B, Z0.B, Z0.B       // 00400004
    ZMLA P3.M, Z12.B, Z13.B, Z10.B    // 8a4d0d04
    ZMLA P7.M, Z31.B, Z31.B, Z31.B    // ff5f1f04
    ZMLA P0.M, Z0.H, Z0.H, Z0.H       // 00404004
    ZMLA P3.M, Z12.H, Z13.H, Z10.H    // 8a4d4d04
    ZMLA P7.M, Z31.H, Z31.H, Z31.H    // ff5f5f04
    ZMLA P0.M, Z0.S, Z0.S, Z0.S       // 00408004
    ZMLA P3.M, Z12.S, Z13.S, Z10.S    // 8a4d8d04
    ZMLA P7.M, Z31.S, Z31.S, Z31.S    // ff5f9f04
    ZMLA P0.M, Z0.D, Z0.D, Z0.D       // 0040c004
    ZMLA P3.M, Z12.D, Z13.D, Z10.D    // 8a4dcd04
    ZMLA P7.M, Z31.D, Z31.D, Z31.D    // ff5fdf04

// MLS     <Zda>.<T>, <Pg>/M, <Zn>.<T>, <Zm>.<T>
    ZMLS P0.M, Z0.B, Z0.B, Z0.B       // 00600004
    ZMLS P3.M, Z12.B, Z13.B, Z10.B    // 8a6d0d04
    ZMLS P7.M, Z31.B, Z31.B, Z31.B    // ff7f1f04
    ZMLS P0.M, Z0.H, Z0.H, Z0.H       // 00604004
    ZMLS P3.M, Z12.H, Z13.H, Z10.H    // 8a6d4d04
    ZMLS P7.M, Z31.H, Z31.H, Z31.H    // ff7f5f04
    ZMLS P0.M, Z0.S, Z0.S, Z0.S       // 00608004
    ZMLS P3.M, Z12.S, Z13.S, Z10.S    // 8a6d8d04
    ZMLS P7.M, Z31.S, Z31.S, Z31.S    // ff7f9f04
    ZMLS P0.M, Z0.D, Z0.D, Z0.D       // 0060c004
    ZMLS P3.M, Z12.D, Z13.D, Z10.D    // 8a6dcd04
    ZMLS P7.M, Z31.D, Z31.D, Z31.D    // ff7fdf04

// MOVPRFX <Zd>.<T>, <Pg>/<ZM>, <Zn>.<T>
    ZMOVPRFX P0.Z, Z0.B, Z0.B      // 00201004
    ZMOVPRFX P3.Z, Z12.B, Z10.B    // 8a2d1004
    ZMOVPRFX P7.Z, Z31.B, Z31.B    // ff3f1004
    ZMOVPRFX P0.M, Z0.B, Z0.B      // 00201104
    ZMOVPRFX P3.M, Z12.B, Z10.B    // 8a2d1104
    ZMOVPRFX P7.M, Z31.B, Z31.B    // ff3f1104
    ZMOVPRFX P0.Z, Z0.H, Z0.H      // 00205004
    ZMOVPRFX P3.Z, Z12.H, Z10.H    // 8a2d5004
    ZMOVPRFX P7.Z, Z31.H, Z31.H    // ff3f5004
    ZMOVPRFX P0.M, Z0.H, Z0.H      // 00205104
    ZMOVPRFX P3.M, Z12.H, Z10.H    // 8a2d5104
    ZMOVPRFX P7.M, Z31.H, Z31.H    // ff3f5104
    ZMOVPRFX P0.Z, Z0.S, Z0.S      // 00209004
    ZMOVPRFX P3.Z, Z12.S, Z10.S    // 8a2d9004
    ZMOVPRFX P7.Z, Z31.S, Z31.S    // ff3f9004
    ZMOVPRFX P0.M, Z0.S, Z0.S      // 00209104
    ZMOVPRFX P3.M, Z12.S, Z10.S    // 8a2d9104
    ZMOVPRFX P7.M, Z31.S, Z31.S    // ff3f9104
    ZMOVPRFX P0.Z, Z0.D, Z0.D      // 0020d004
    ZMOVPRFX P3.Z, Z12.D, Z10.D    // 8a2dd004
    ZMOVPRFX P7.Z, Z31.D, Z31.D    // ff3fd004
    ZMOVPRFX P0.M, Z0.D, Z0.D      // 0020d104
    ZMOVPRFX P3.M, Z12.D, Z10.D    // 8a2dd104
    ZMOVPRFX P7.M, Z31.D, Z31.D    // ff3fd104

// MOVPRFX <Zd>, <Zn>
    ZMOVPRFX Z0, Z0      // 00bc2004
    ZMOVPRFX Z11, Z10    // 6abd2004
    ZMOVPRFX Z31, Z31    // ffbf2004

// MSB     <Zdn>.<T>, <Pg>/M, <Zm>.<T>, <Za>.<T>
    ZMSB P0.M, Z0.B, Z0.B, Z0.B       // 00e00004
    ZMSB P3.M, Z12.B, Z13.B, Z10.B    // aaed0c04
    ZMSB P7.M, Z31.B, Z31.B, Z31.B    // ffff1f04
    ZMSB P0.M, Z0.H, Z0.H, Z0.H       // 00e04004
    ZMSB P3.M, Z12.H, Z13.H, Z10.H    // aaed4c04
    ZMSB P7.M, Z31.H, Z31.H, Z31.H    // ffff5f04
    ZMSB P0.M, Z0.S, Z0.S, Z0.S       // 00e08004
    ZMSB P3.M, Z12.S, Z13.S, Z10.S    // aaed8c04
    ZMSB P7.M, Z31.S, Z31.S, Z31.S    // ffff9f04
    ZMSB P0.M, Z0.D, Z0.D, Z0.D       // 00e0c004
    ZMSB P3.M, Z12.D, Z13.D, Z10.D    // aaedcc04
    ZMSB P7.M, Z31.D, Z31.D, Z31.D    // ffffdf04

// MUL     <Zdn>.<T>, <Pg>/M, <Zdn>.<T>, <Zm>.<T>
    ZMUL P0.M, Z0.B, Z0.B, Z0.B       // 00001004
    ZMUL P3.M, Z10.B, Z12.B, Z10.B    // 8a0d1004
    ZMUL P7.M, Z31.B, Z31.B, Z31.B    // ff1f1004
    ZMUL P0.M, Z0.H, Z0.H, Z0.H       // 00005004
    ZMUL P3.M, Z10.H, Z12.H, Z10.H    // 8a0d5004
    ZMUL P7.M, Z31.H, Z31.H, Z31.H    // ff1f5004
    ZMUL P0.M, Z0.S, Z0.S, Z0.S       // 00009004
    ZMUL P3.M, Z10.S, Z12.S, Z10.S    // 8a0d9004
    ZMUL P7.M, Z31.S, Z31.S, Z31.S    // ff1f9004
    ZMUL P0.M, Z0.D, Z0.D, Z0.D       // 0000d004
    ZMUL P3.M, Z10.D, Z12.D, Z10.D    // 8a0dd004
    ZMUL P7.M, Z31.D, Z31.D, Z31.D    // ff1fd004

// MUL     <Zdn>.<T>, <Zdn>.<T>, #<imm>
    ZMUL Z0.B, $-128, Z0.B     // 00d03025
    ZMUL Z10.B, $-43, Z10.B    // aada3025
    ZMUL Z31.B, $127, Z31.B    // ffcf3025
    ZMUL Z0.H, $-128, Z0.H     // 00d07025
    ZMUL Z10.H, $-43, Z10.H    // aada7025
    ZMUL Z31.H, $127, Z31.H    // ffcf7025
    ZMUL Z0.S, $-128, Z0.S     // 00d0b025
    ZMUL Z10.S, $-43, Z10.S    // aadab025
    ZMUL Z31.S, $127, Z31.S    // ffcfb025
    ZMUL Z0.D, $-128, Z0.D     // 00d0f025
    ZMUL Z10.D, $-43, Z10.D    // aadaf025
    ZMUL Z31.D, $127, Z31.D    // ffcff025

// NAND    <Pd>.B, <Pg>/Z, <Pn>.B, <Pm>.B
    PNAND P0.Z, P0.B, P0.B, P0.B        // 10428025
    PNAND P6.Z, P7.B, P8.B, P5.B        // f55a8825
    PNAND P15.Z, P15.B, P15.B, P15.B    // ff7f8f25

// NANDS   <Pd>.B, <Pg>/Z, <Pn>.B, <Pm>.B
    PNANDS P0.Z, P0.B, P0.B, P0.B        // 1042c025
    PNANDS P6.Z, P7.B, P8.B, P5.B        // f55ac825
    PNANDS P15.Z, P15.B, P15.B, P15.B    // ff7fcf25

// NEG     <Zd>.<T>, <Pg>/M, <Zn>.<T>
    ZNEG P0.M, Z0.B, Z0.B      // 00a01704
    ZNEG P3.M, Z12.B, Z10.B    // 8aad1704
    ZNEG P7.M, Z31.B, Z31.B    // ffbf1704
    ZNEG P0.M, Z0.H, Z0.H      // 00a05704
    ZNEG P3.M, Z12.H, Z10.H    // 8aad5704
    ZNEG P7.M, Z31.H, Z31.H    // ffbf5704
    ZNEG P0.M, Z0.S, Z0.S      // 00a09704
    ZNEG P3.M, Z12.S, Z10.S    // 8aad9704
    ZNEG P7.M, Z31.S, Z31.S    // ffbf9704
    ZNEG P0.M, Z0.D, Z0.D      // 00a0d704
    ZNEG P3.M, Z12.D, Z10.D    // 8aadd704
    ZNEG P7.M, Z31.D, Z31.D    // ffbfd704

// NOR     <Pd>.B, <Pg>/Z, <Pn>.B, <Pm>.B
    PNOR P0.Z, P0.B, P0.B, P0.B        // 00428025
    PNOR P6.Z, P7.B, P8.B, P5.B        // e55a8825
    PNOR P15.Z, P15.B, P15.B, P15.B    // ef7f8f25

// NORS    <Pd>.B, <Pg>/Z, <Pn>.B, <Pm>.B
    PNORS P0.Z, P0.B, P0.B, P0.B        // 0042c025
    PNORS P6.Z, P7.B, P8.B, P5.B        // e55ac825
    PNORS P15.Z, P15.B, P15.B, P15.B    // ef7fcf25

// NOT     <Zd>.<T>, <Pg>/M, <Zn>.<T>
    ZNOT P0.M, Z0.B, Z0.B      // 00a01e04
    ZNOT P3.M, Z12.B, Z10.B    // 8aad1e04
    ZNOT P7.M, Z31.B, Z31.B    // ffbf1e04
    ZNOT P0.M, Z0.H, Z0.H      // 00a05e04
    ZNOT P3.M, Z12.H, Z10.H    // 8aad5e04
    ZNOT P7.M, Z31.H, Z31.H    // ffbf5e04
    ZNOT P0.M, Z0.S, Z0.S      // 00a09e04
    ZNOT P3.M, Z12.S, Z10.S    // 8aad9e04
    ZNOT P7.M, Z31.S, Z31.S    // ffbf9e04
    ZNOT P0.M, Z0.D, Z0.D      // 00a0de04
    ZNOT P3.M, Z12.D, Z10.D    // 8aadde04
    ZNOT P7.M, Z31.D, Z31.D    // ffbfde04

// ORN     <Pd>.B, <Pg>/Z, <Pn>.B, <Pm>.B
    ORN P0.Z, P0.B, P0.B, P0.B        // 10408025
    ORN P6.Z, P7.B, P8.B, P5.B        // f5588825
    ORN P15.Z, P15.B, P15.B, P15.B    // ff7d8f25

// ORNS    <Pd>.B, <Pg>/Z, <Pn>.B, <Pm>.B
    ORNS P0.Z, P0.B, P0.B, P0.B        // 1040c025
    ORNS P6.Z, P7.B, P8.B, P5.B        // f558c825
    ORNS P15.Z, P15.B, P15.B, P15.B    // ff7dcf25

// ORR     <Pd>.B, <Pg>/Z, <Pn>.B, <Pm>.B
    PORR P0.Z, P0.B, P0.B, P0.B        // 00408025
    PORR P6.Z, P7.B, P8.B, P5.B        // e5588825
    PORR P15.Z, P15.B, P15.B, P15.B    // ef7d8f25

// ORR     <Zdn>.<T>, <Pg>/M, <Zdn>.<T>, <Zm>.<T>
    ZORR P0.M, Z0.B, Z0.B, Z0.B       // 00001804
    ZORR P3.M, Z10.B, Z12.B, Z10.B    // 8a0d1804
    ZORR P7.M, Z31.B, Z31.B, Z31.B    // ff1f1804
    ZORR P0.M, Z0.H, Z0.H, Z0.H       // 00005804
    ZORR P3.M, Z10.H, Z12.H, Z10.H    // 8a0d5804
    ZORR P7.M, Z31.H, Z31.H, Z31.H    // ff1f5804
    ZORR P0.M, Z0.S, Z0.S, Z0.S       // 00009804
    ZORR P3.M, Z10.S, Z12.S, Z10.S    // 8a0d9804
    ZORR P7.M, Z31.S, Z31.S, Z31.S    // ff1f9804
    ZORR P0.M, Z0.D, Z0.D, Z0.D       // 0000d804
    ZORR P3.M, Z10.D, Z12.D, Z10.D    // 8a0dd804
    ZORR P7.M, Z31.D, Z31.D, Z31.D    // ff1fd804

// ORR     <Zdn>.D, <Zdn>.D, #<const>
    ZORR Z0.S, $1, Z0.S                   // 00000005
    ZORR Z10.S, $4192256, Z10.S           // 4aa90005
    ZORR Z0.H, $1, Z0.H                   // 00040005
    ZORR Z10.H, $63489, Z10.H             // aa2c0005
    ZORR Z0.B, $1, Z0.B                   // 00060005
    ZORR Z10.B, $56, Z10.B                // 4a2e0005
    ZORR Z0.B, $17, Z0.B                  // 00070005
    ZORR Z10.B, $153, Z10.B               // 2a0f0005
    ZORR Z0.D, $4294967297, Z0.D          // 00000005
    ZORR Z10.D, $-8787503089663, Z10.D    // aaaa0005

// ORR     <Zd>.D, <Zn>.D, <Zm>.D
    ZORR Z0.D, Z0.D, Z0.D       // 00306004
    ZORR Z11.D, Z12.D, Z10.D    // 6a316c04
    ZORR Z31.D, Z31.D, Z31.D    // ff337f04

// ORRS    <Pd>.B, <Pg>/Z, <Pn>.B, <Pm>.B
    PORRS P0.Z, P0.B, P0.B, P0.B        // 0040c025
    PORRS P6.Z, P7.B, P8.B, P5.B        // e558c825
    PORRS P15.Z, P15.B, P15.B, P15.B    // ef7dcf25

// ORV     <V><d>, <Pg>, <Zn>.<T>
    ZORV P0, Z0.B, V0      // 00201804
    ZORV P3, Z12.B, V10    // 8a2d1804
    ZORV P7, Z31.B, V31    // ff3f1804
    ZORV P0, Z0.H, V0      // 00205804
    ZORV P3, Z12.H, V10    // 8a2d5804
    ZORV P7, Z31.H, V31    // ff3f5804
    ZORV P0, Z0.S, V0      // 00209804
    ZORV P3, Z12.S, V10    // 8a2d9804
    ZORV P7, Z31.S, V31    // ff3f9804
    ZORV P0, Z0.D, V0      // 0020d804
    ZORV P3, Z12.D, V10    // 8a2dd804
    ZORV P7, Z31.D, V31    // ff3fd804

// PFALSE  <Pd>.B
    PFALSE P0.B     // 00e41825
    PFALSE P5.B     // 05e41825
    PFALSE P15.B    // 0fe41825

// PFIRST  <Pdn>.B, <Pg>, <Pdn>.B
    PFIRST P0, P0.B, P0.B       // 00c05825
    PFIRST P6, P5.B, P5.B       // c5c05825
    PFIRST P15, P15.B, P15.B    // efc15825

// PNEXT   <Pdn>.<T>, <Pv>, <Pdn>.<T>
    PNEXT P0, P0.B, P0.B       // 00c41925
    PNEXT P6, P5.B, P5.B       // c5c41925
    PNEXT P15, P15.B, P15.B    // efc51925
    PNEXT P0, P0.H, P0.H       // 00c45925
    PNEXT P6, P5.H, P5.H       // c5c45925
    PNEXT P15, P15.H, P15.H    // efc55925
    PNEXT P0, P0.S, P0.S       // 00c49925
    PNEXT P6, P5.S, P5.S       // c5c49925
    PNEXT P15, P15.S, P15.S    // efc59925
    PNEXT P0, P0.D, P0.D       // 00c4d925
    PNEXT P6, P5.D, P5.D       // c5c4d925
    PNEXT P15, P15.D, P15.D    // efc5d925

// PRFB    <prfop>, <Pg>, [<Zn>.D{, #<imm>}]
    ZPRFB PLDL1KEEP, P0, (Z0.D)      // 00e000c4
    ZPRFB PLDL1STRM, P1, 2(Z4.D)      // 81e402c4
    ZPRFB PLDL2KEEP, P2, 4(Z6.D)      // c2e804c4
    ZPRFB PLDL2STRM, P2, 6(Z8.D)      // 03e906c4
    ZPRFB PLDL3KEEP, P3, 8(Z10.D)     // 44ed08c4
    ZPRFB PLDL3STRM, P3, 10(Z12.D)    // 85ed0ac4
    ZPRFB $6, P4, 12(Z14.D)           // c6f10cc4
    ZPRFB $7, P4, 14(Z16.D)           // 07f20ec4
    ZPRFB PSTL1KEEP, P5, 16(Z18.D)    // 48f610c4
    ZPRFB PSTL1STRM, P5, 17(Z19.D)    // 69f611c4
    ZPRFB PSTL2KEEP, P5, 19(Z21.D)    // aaf613c4
    ZPRFB PSTL2STRM, P6, 21(Z23.D)    // ebfa15c4
    ZPRFB PSTL3KEEP, P6, 23(Z25.D)    // 2cfb17c4
    ZPRFB PSTL3STRM, P7, 25(Z27.D)    // 6dff19c4
    ZPRFB $14, P7, 27(Z29.D)          // aeff1bc4
    ZPRFB $15, P7, 31(Z31.D)          // efff1fc4

// PRFB    <prfop>, <Pg>, [<Zn>.S{, #<imm>}]
    ZPRFB PLDL1KEEP, P0, (Z0.S)      // 00e00084
    ZPRFB PLDL1STRM, P1, 2(Z4.S)      // 81e40284
    ZPRFB PLDL2KEEP, P2, 4(Z6.S)      // c2e80484
    ZPRFB PLDL2STRM, P2, 6(Z8.S)      // 03e90684
    ZPRFB PLDL3KEEP, P3, 8(Z10.S)     // 44ed0884
    ZPRFB PLDL3STRM, P3, 10(Z12.S)    // 85ed0a84
    ZPRFB $6, P4, 12(Z14.S)           // c6f10c84
    ZPRFB $7, P4, 14(Z16.S)           // 07f20e84
    ZPRFB PSTL1KEEP, P5, 16(Z18.S)    // 48f61084
    ZPRFB PSTL1STRM, P5, 17(Z19.S)    // 69f61184
    ZPRFB PSTL2KEEP, P5, 19(Z21.S)    // aaf61384
    ZPRFB PSTL2STRM, P6, 21(Z23.S)    // ebfa1584
    ZPRFB PSTL3KEEP, P6, 23(Z25.S)    // 2cfb1784
    ZPRFB PSTL3STRM, P7, 25(Z27.S)    // 6dff1984
    ZPRFB $14, P7, 27(Z29.S)          // aeff1b84
    ZPRFB $15, P7, 31(Z31.S)          // efff1f84

// PRFB    <prfop>, <Pg>, [<Xn|SP>{, #<imm>, MUL VL}]
    ZPRFB PLDL1KEEP, P0, -32(R0)     // 0000e085
    ZPRFB PLDL1STRM, P1, -28(R4)     // 8104e485
    ZPRFB PLDL2KEEP, P2, -24(R6)     // c208e885
    ZPRFB PLDL2STRM, P2, -20(R8)     // 0309ec85
    ZPRFB PLDL3KEEP, P3, -16(R10)    // 440df085
    ZPRFB PLDL3STRM, P3, -12(R11)    // 650df485
    ZPRFB $6, P4, -8(R13)            // a611f885
    ZPRFB $7, P4, -4(R15)            // e711fc85
    ZPRFB PSTL1KEEP, P5, (R17)      // 2816c085
    ZPRFB PSTL1STRM, P5, 3(R19)      // 6916c385
    ZPRFB PSTL2KEEP, P5, 7(R21)      // aa16c785
    ZPRFB PSTL2STRM, P6, 11(R23)     // eb1acb85
    ZPRFB PSTL3KEEP, P6, 15(R24)     // 0c1bcf85
    ZPRFB PSTL3STRM, P7, 19(R26)     // 4d1fd385
    ZPRFB $14, P7, 23(R27)           // 6e1fd785
    ZPRFB $15, P7, 31(R30)           // cf1fdf85

// PRFB    <prfop>, <Pg>, [<Xn|SP>, <Xm>]
    ZPRFB PLDL1KEEP, P0, (R0)(R0)      // 00c00084
    ZPRFB PLDL1STRM, P1, (R4)(R5)      // 81c40584
    ZPRFB PLDL2KEEP, P2, (R6)(R7)      // c2c80784
    ZPRFB PLDL2STRM, P2, (R8)(R9)      // 03c90984
    ZPRFB PLDL3KEEP, P3, (R10)(R11)    // 44cd0b84
    ZPRFB PLDL3STRM, P3, (R11)(R12)    // 65cd0c84
    ZPRFB $6, P4, (R13)(R14)           // a6d10e84
    ZPRFB $7, P4, (R15)(R16)           // e7d11084
    ZPRFB PSTL1KEEP, P5, (R17)(R17)    // 28d61184
    ZPRFB PSTL1STRM, P5, (R19)(R20)    // 69d61484
    ZPRFB PSTL2KEEP, P5, (R21)(R22)    // aad61684
    ZPRFB PSTL2STRM, P6, (R23)(R24)    // ebda1884
    ZPRFB PSTL3KEEP, P6, (R24)(R25)    // 0cdb1984
    ZPRFB PSTL3STRM, P7, (R26)(R27)    // 4ddf1b84
    ZPRFB $14, P7, (R27)(R29)          // 6edf1d84
    ZPRFB $15, P7, (R30)(R30)          // cfdf1e84

// PRFB    <prfop>, <Pg>, [<Xn|SP>, <Zm>.D]
    ZPRFB PLDL1KEEP, P0, (R0)(Z0.D)      // 008060c4
    ZPRFB PLDL1STRM, P1, (R4)(Z5.D)      // 818465c4
    ZPRFB PLDL2KEEP, P2, (R6)(Z7.D)      // c28867c4
    ZPRFB PLDL2STRM, P2, (R8)(Z9.D)      // 038969c4
    ZPRFB PLDL3KEEP, P3, (R10)(Z11.D)    // 448d6bc4
    ZPRFB PLDL3STRM, P3, (R11)(Z13.D)    // 658d6dc4
    ZPRFB $6, P4, (R13)(Z15.D)           // a6916fc4
    ZPRFB $7, P4, (R15)(Z17.D)           // e79171c4
    ZPRFB PSTL1KEEP, P5, (R17)(Z19.D)    // 289673c4
    ZPRFB PSTL1STRM, P5, (R19)(Z20.D)    // 699674c4
    ZPRFB PSTL2KEEP, P5, (R21)(Z22.D)    // aa9676c4
    ZPRFB PSTL2STRM, P6, (R23)(Z24.D)    // eb9a78c4
    ZPRFB PSTL3KEEP, P6, (R24)(Z26.D)    // 0c9b7ac4
    ZPRFB PSTL3STRM, P7, (R26)(Z28.D)    // 4d9f7cc4
    ZPRFB $14, P7, (R27)(Z30.D)          // 6e9f7ec4
    ZPRFB $15, P7, (R30)(Z31.D)          // cf9f7fc4

// PRFB    <prfop>, <Pg>, [<Xn|SP>, <Zm>.D, <extend>]
    ZPRFB PLDL1KEEP, P0, (R0)(Z0.D.UXTW)      // 000020c4
    ZPRFB PLDL1STRM, P1, (R4)(Z5.D.UXTW)      // 810425c4
    ZPRFB PLDL2KEEP, P2, (R6)(Z7.D.UXTW)      // c20827c4
    ZPRFB PLDL2STRM, P2, (R8)(Z9.D.UXTW)      // 030929c4
    ZPRFB PLDL3KEEP, P3, (R10)(Z11.D.UXTW)    // 440d2bc4
    ZPRFB PLDL3STRM, P3, (R11)(Z13.D.UXTW)    // 650d2dc4
    ZPRFB $6, P4, (R13)(Z15.D.UXTW)           // a6112fc4
    ZPRFB $7, P4, (R15)(Z17.D.UXTW)           // e71131c4
    ZPRFB PSTL1KEEP, P5, (R17)(Z19.D.UXTW)    // 281633c4
    ZPRFB PSTL1STRM, P5, (R19)(Z20.D.UXTW)    // 691634c4
    ZPRFB PSTL2KEEP, P5, (R21)(Z22.D.UXTW)    // aa1636c4
    ZPRFB PSTL2STRM, P6, (R23)(Z24.D.UXTW)    // eb1a38c4
    ZPRFB PSTL3KEEP, P6, (R24)(Z26.D.UXTW)    // 0c1b3ac4
    ZPRFB PSTL3STRM, P7, (R26)(Z28.D.UXTW)    // 4d1f3cc4
    ZPRFB $14, P7, (R27)(Z30.D.UXTW)          // 6e1f3ec4
    ZPRFB $15, P7, (R30)(Z31.D.UXTW)          // cf1f3fc4
    ZPRFB PLDL1KEEP, P0, (R0)(Z0.D.SXTW)      // 000060c4
    ZPRFB PLDL1STRM, P1, (R4)(Z5.D.SXTW)      // 810465c4
    ZPRFB PLDL2KEEP, P2, (R6)(Z7.D.SXTW)      // c20867c4
    ZPRFB PLDL2STRM, P2, (R8)(Z9.D.SXTW)      // 030969c4
    ZPRFB PLDL3KEEP, P3, (R10)(Z11.D.SXTW)    // 440d6bc4
    ZPRFB PLDL3STRM, P3, (R11)(Z13.D.SXTW)    // 650d6dc4
    ZPRFB $6, P4, (R13)(Z15.D.SXTW)           // a6116fc4
    ZPRFB $7, P4, (R15)(Z17.D.SXTW)           // e71171c4
    ZPRFB PSTL1KEEP, P5, (R17)(Z19.D.SXTW)    // 281673c4
    ZPRFB PSTL1STRM, P5, (R19)(Z20.D.SXTW)    // 691674c4
    ZPRFB PSTL2KEEP, P5, (R21)(Z22.D.SXTW)    // aa1676c4
    ZPRFB PSTL2STRM, P6, (R23)(Z24.D.SXTW)    // eb1a78c4
    ZPRFB PSTL3KEEP, P6, (R24)(Z26.D.SXTW)    // 0c1b7ac4
    ZPRFB PSTL3STRM, P7, (R26)(Z28.D.SXTW)    // 4d1f7cc4
    ZPRFB $14, P7, (R27)(Z30.D.SXTW)          // 6e1f7ec4
    ZPRFB $15, P7, (R30)(Z31.D.SXTW)          // cf1f7fc4

// PRFB    <prfop>, <Pg>, [<Xn|SP>, <Zm>.S, <extend>]
    ZPRFB PLDL1KEEP, P0, (R0)(Z0.S.UXTW)      // 00002084
    ZPRFB PLDL1STRM, P1, (R4)(Z5.S.UXTW)      // 81042584
    ZPRFB PLDL2KEEP, P2, (R6)(Z7.S.UXTW)      // c2082784
    ZPRFB PLDL2STRM, P2, (R8)(Z9.S.UXTW)      // 03092984
    ZPRFB PLDL3KEEP, P3, (R10)(Z11.S.UXTW)    // 440d2b84
    ZPRFB PLDL3STRM, P3, (R11)(Z13.S.UXTW)    // 650d2d84
    ZPRFB $6, P4, (R13)(Z15.S.UXTW)           // a6112f84
    ZPRFB $7, P4, (R15)(Z17.S.UXTW)           // e7113184
    ZPRFB PSTL1KEEP, P5, (R17)(Z19.S.UXTW)    // 28163384
    ZPRFB PSTL1STRM, P5, (R19)(Z20.S.UXTW)    // 69163484
    ZPRFB PSTL2KEEP, P5, (R21)(Z22.S.UXTW)    // aa163684
    ZPRFB PSTL2STRM, P6, (R23)(Z24.S.UXTW)    // eb1a3884
    ZPRFB PSTL3KEEP, P6, (R24)(Z26.S.UXTW)    // 0c1b3a84
    ZPRFB PSTL3STRM, P7, (R26)(Z28.S.UXTW)    // 4d1f3c84
    ZPRFB $14, P7, (R27)(Z30.S.UXTW)          // 6e1f3e84
    ZPRFB $15, P7, (R30)(Z31.S.UXTW)          // cf1f3f84
    ZPRFB PLDL1KEEP, P0, (R0)(Z0.S.SXTW)      // 00006084
    ZPRFB PLDL1STRM, P1, (R4)(Z5.S.SXTW)      // 81046584
    ZPRFB PLDL2KEEP, P2, (R6)(Z7.S.SXTW)      // c2086784
    ZPRFB PLDL2STRM, P2, (R8)(Z9.S.SXTW)      // 03096984
    ZPRFB PLDL3KEEP, P3, (R10)(Z11.S.SXTW)    // 440d6b84
    ZPRFB PLDL3STRM, P3, (R11)(Z13.S.SXTW)    // 650d6d84
    ZPRFB $6, P4, (R13)(Z15.S.SXTW)           // a6116f84
    ZPRFB $7, P4, (R15)(Z17.S.SXTW)           // e7117184
    ZPRFB PSTL1KEEP, P5, (R17)(Z19.S.SXTW)    // 28167384
    ZPRFB PSTL1STRM, P5, (R19)(Z20.S.SXTW)    // 69167484
    ZPRFB PSTL2KEEP, P5, (R21)(Z22.S.SXTW)    // aa167684
    ZPRFB PSTL2STRM, P6, (R23)(Z24.S.SXTW)    // eb1a7884
    ZPRFB PSTL3KEEP, P6, (R24)(Z26.S.SXTW)    // 0c1b7a84
    ZPRFB PSTL3STRM, P7, (R26)(Z28.S.SXTW)    // 4d1f7c84
    ZPRFB $14, P7, (R27)(Z30.S.SXTW)          // 6e1f7e84
    ZPRFB $15, P7, (R30)(Z31.S.SXTW)          // cf1f7f84

// PRFD    <prfop>, <Pg>, [<Zn>.D{, #<imm>}]
    ZPRFD PLDL1KEEP, P0, (Z0.D)       // 00e080c5
    ZPRFD PLDL1STRM, P1, 16(Z4.D)      // 81e482c5
    ZPRFD PLDL2KEEP, P2, 32(Z6.D)      // c2e884c5
    ZPRFD PLDL2STRM, P2, 48(Z8.D)      // 03e986c5
    ZPRFD PLDL3KEEP, P3, 64(Z10.D)     // 44ed88c5
    ZPRFD PLDL3STRM, P3, 80(Z12.D)     // 85ed8ac5
    ZPRFD $6, P4, 96(Z14.D)            // c6f18cc5
    ZPRFD $7, P4, 112(Z16.D)           // 07f28ec5
    ZPRFD PSTL1KEEP, P5, 128(Z18.D)    // 48f690c5
    ZPRFD PSTL1STRM, P5, 136(Z19.D)    // 69f691c5
    ZPRFD PSTL2KEEP, P5, 152(Z21.D)    // aaf693c5
    ZPRFD PSTL2STRM, P6, 168(Z23.D)    // ebfa95c5
    ZPRFD PSTL3KEEP, P6, 184(Z25.D)    // 2cfb97c5
    ZPRFD PSTL3STRM, P7, 200(Z27.D)    // 6dff99c5
    ZPRFD $14, P7, 216(Z29.D)          // aeff9bc5
    ZPRFD $15, P7, 248(Z31.D)          // efff9fc5

// PRFD    <prfop>, <Pg>, [<Zn>.S{, #<imm>}]
    ZPRFD PLDL1KEEP, P0, (Z0.S)       // 00e08085
    ZPRFD PLDL1STRM, P1, 16(Z4.S)      // 81e48285
    ZPRFD PLDL2KEEP, P2, 32(Z6.S)      // c2e88485
    ZPRFD PLDL2STRM, P2, 48(Z8.S)      // 03e98685
    ZPRFD PLDL3KEEP, P3, 64(Z10.S)     // 44ed8885
    ZPRFD PLDL3STRM, P3, 80(Z12.S)     // 85ed8a85
    ZPRFD $6, P4, 96(Z14.S)            // c6f18c85
    ZPRFD $7, P4, 112(Z16.S)           // 07f28e85
    ZPRFD PSTL1KEEP, P5, 128(Z18.S)    // 48f69085
    ZPRFD PSTL1STRM, P5, 136(Z19.S)    // 69f69185
    ZPRFD PSTL2KEEP, P5, 152(Z21.S)    // aaf69385
    ZPRFD PSTL2STRM, P6, 168(Z23.S)    // ebfa9585
    ZPRFD PSTL3KEEP, P6, 184(Z25.S)    // 2cfb9785
    ZPRFD PSTL3STRM, P7, 200(Z27.S)    // 6dff9985
    ZPRFD $14, P7, 216(Z29.S)          // aeff9b85
    ZPRFD $15, P7, 248(Z31.S)          // efff9f85

// PRFD    <prfop>, <Pg>, [<Xn|SP>{, #<imm>, MUL VL}]
    ZPRFD PLDL1KEEP, P0, -32(R0)     // 0060e085
    ZPRFD PLDL1STRM, P1, -28(R4)     // 8164e485
    ZPRFD PLDL2KEEP, P2, -24(R6)     // c268e885
    ZPRFD PLDL2STRM, P2, -20(R8)     // 0369ec85
    ZPRFD PLDL3KEEP, P3, -16(R10)    // 446df085
    ZPRFD PLDL3STRM, P3, -12(R11)    // 656df485
    ZPRFD $6, P4, -8(R13)            // a671f885
    ZPRFD $7, P4, -4(R15)            // e771fc85
    ZPRFD PSTL1KEEP, P5, (R17)      // 2876c085
    ZPRFD PSTL1STRM, P5, 3(R19)      // 6976c385
    ZPRFD PSTL2KEEP, P5, 7(R21)      // aa76c785
    ZPRFD PSTL2STRM, P6, 11(R23)     // eb7acb85
    ZPRFD PSTL3KEEP, P6, 15(R24)     // 0c7bcf85
    ZPRFD PSTL3STRM, P7, 19(R26)     // 4d7fd385
    ZPRFD $14, P7, 23(R27)           // 6e7fd785
    ZPRFD $15, P7, 31(R30)           // cf7fdf85

// PRFD    <prfop>, <Pg>, [<Xn|SP>, <Xm>, LSL #3]
    ZPRFD PLDL1KEEP, P0, (R0)(R0<<3)      // 00c08085
    ZPRFD PLDL1STRM, P1, (R4)(R5<<3)      // 81c48585
    ZPRFD PLDL2KEEP, P2, (R6)(R7<<3)      // c2c88785
    ZPRFD PLDL2STRM, P2, (R8)(R9<<3)      // 03c98985
    ZPRFD PLDL3KEEP, P3, (R10)(R11<<3)    // 44cd8b85
    ZPRFD PLDL3STRM, P3, (R11)(R12<<3)    // 65cd8c85
    ZPRFD $6, P4, (R13)(R14<<3)           // a6d18e85
    ZPRFD $7, P4, (R15)(R16<<3)           // e7d19085
    ZPRFD PSTL1KEEP, P5, (R17)(R17<<3)    // 28d69185
    ZPRFD PSTL1STRM, P5, (R19)(R20<<3)    // 69d69485
    ZPRFD PSTL2KEEP, P5, (R21)(R22<<3)    // aad69685
    ZPRFD PSTL2STRM, P6, (R23)(R24<<3)    // ebda9885
    ZPRFD PSTL3KEEP, P6, (R24)(R25<<3)    // 0cdb9985
    ZPRFD PSTL3STRM, P7, (R26)(R27<<3)    // 4ddf9b85
    ZPRFD $14, P7, (R27)(R29<<3)          // 6edf9d85
    ZPRFD $15, P7, (R30)(R30<<3)          // cfdf9e85

// PRFD    <prfop>, <Pg>, [<Xn|SP>, <Zm>.D, LSL #3]
    ZPRFD PLDL1KEEP, P0, (R0)(Z0.D.LSL<<3)      // 00e060c4
    ZPRFD PLDL1STRM, P1, (R4)(Z5.D.LSL<<3)      // 81e465c4
    ZPRFD PLDL2KEEP, P2, (R6)(Z7.D.LSL<<3)      // c2e867c4
    ZPRFD PLDL2STRM, P2, (R8)(Z9.D.LSL<<3)      // 03e969c4
    ZPRFD PLDL3KEEP, P3, (R10)(Z11.D.LSL<<3)    // 44ed6bc4
    ZPRFD PLDL3STRM, P3, (R11)(Z13.D.LSL<<3)    // 65ed6dc4
    ZPRFD $6, P4, (R13)(Z15.D.LSL<<3)           // a6f16fc4
    ZPRFD $7, P4, (R15)(Z17.D.LSL<<3)           // e7f171c4
    ZPRFD PSTL1KEEP, P5, (R17)(Z19.D.LSL<<3)    // 28f673c4
    ZPRFD PSTL1STRM, P5, (R19)(Z20.D.LSL<<3)    // 69f674c4
    ZPRFD PSTL2KEEP, P5, (R21)(Z22.D.LSL<<3)    // aaf676c4
    ZPRFD PSTL2STRM, P6, (R23)(Z24.D.LSL<<3)    // ebfa78c4
    ZPRFD PSTL3KEEP, P6, (R24)(Z26.D.LSL<<3)    // 0cfb7ac4
    ZPRFD PSTL3STRM, P7, (R26)(Z28.D.LSL<<3)    // 4dff7cc4
    ZPRFD $14, P7, (R27)(Z30.D.LSL<<3)          // 6eff7ec4
    ZPRFD $15, P7, (R30)(Z31.D.LSL<<3)          // cfff7fc4

// PRFD    <prfop>, <Pg>, [<Xn|SP>, <Zm>.D, <extend> #3]
    ZPRFD PLDL1KEEP, P0, (R0)(Z0.D.UXTW<<3)      // 006020c4
    ZPRFD PLDL1STRM, P1, (R4)(Z5.D.UXTW<<3)      // 816425c4
    ZPRFD PLDL2KEEP, P2, (R6)(Z7.D.UXTW<<3)      // c26827c4
    ZPRFD PLDL2STRM, P2, (R8)(Z9.D.UXTW<<3)      // 036929c4
    ZPRFD PLDL3KEEP, P3, (R10)(Z11.D.UXTW<<3)    // 446d2bc4
    ZPRFD PLDL3STRM, P3, (R11)(Z13.D.UXTW<<3)    // 656d2dc4
    ZPRFD $6, P4, (R13)(Z15.D.UXTW<<3)           // a6712fc4
    ZPRFD $7, P4, (R15)(Z17.D.UXTW<<3)           // e77131c4
    ZPRFD PSTL1KEEP, P5, (R17)(Z19.D.UXTW<<3)    // 287633c4
    ZPRFD PSTL1STRM, P5, (R19)(Z20.D.UXTW<<3)    // 697634c4
    ZPRFD PSTL2KEEP, P5, (R21)(Z22.D.UXTW<<3)    // aa7636c4
    ZPRFD PSTL2STRM, P6, (R23)(Z24.D.UXTW<<3)    // eb7a38c4
    ZPRFD PSTL3KEEP, P6, (R24)(Z26.D.UXTW<<3)    // 0c7b3ac4
    ZPRFD PSTL3STRM, P7, (R26)(Z28.D.UXTW<<3)    // 4d7f3cc4
    ZPRFD $14, P7, (R27)(Z30.D.UXTW<<3)          // 6e7f3ec4
    ZPRFD $15, P7, (R30)(Z31.D.UXTW<<3)          // cf7f3fc4
    ZPRFD PLDL1KEEP, P0, (R0)(Z0.D.SXTW<<3)      // 006060c4
    ZPRFD PLDL1STRM, P1, (R4)(Z5.D.SXTW<<3)      // 816465c4
    ZPRFD PLDL2KEEP, P2, (R6)(Z7.D.SXTW<<3)      // c26867c4
    ZPRFD PLDL2STRM, P2, (R8)(Z9.D.SXTW<<3)      // 036969c4
    ZPRFD PLDL3KEEP, P3, (R10)(Z11.D.SXTW<<3)    // 446d6bc4
    ZPRFD PLDL3STRM, P3, (R11)(Z13.D.SXTW<<3)    // 656d6dc4
    ZPRFD $6, P4, (R13)(Z15.D.SXTW<<3)           // a6716fc4
    ZPRFD $7, P4, (R15)(Z17.D.SXTW<<3)           // e77171c4
    ZPRFD PSTL1KEEP, P5, (R17)(Z19.D.SXTW<<3)    // 287673c4
    ZPRFD PSTL1STRM, P5, (R19)(Z20.D.SXTW<<3)    // 697674c4
    ZPRFD PSTL2KEEP, P5, (R21)(Z22.D.SXTW<<3)    // aa7676c4
    ZPRFD PSTL2STRM, P6, (R23)(Z24.D.SXTW<<3)    // eb7a78c4
    ZPRFD PSTL3KEEP, P6, (R24)(Z26.D.SXTW<<3)    // 0c7b7ac4
    ZPRFD PSTL3STRM, P7, (R26)(Z28.D.SXTW<<3)    // 4d7f7cc4
    ZPRFD $14, P7, (R27)(Z30.D.SXTW<<3)          // 6e7f7ec4
    ZPRFD $15, P7, (R30)(Z31.D.SXTW<<3)          // cf7f7fc4

// PRFD    <prfop>, <Pg>, [<Xn|SP>, <Zm>.S, <extend> #3]
    ZPRFD PLDL1KEEP, P0, (R0)(Z0.S.UXTW<<3)      // 00602084
    ZPRFD PLDL1STRM, P1, (R4)(Z5.S.UXTW<<3)      // 81642584
    ZPRFD PLDL2KEEP, P2, (R6)(Z7.S.UXTW<<3)      // c2682784
    ZPRFD PLDL2STRM, P2, (R8)(Z9.S.UXTW<<3)      // 03692984
    ZPRFD PLDL3KEEP, P3, (R10)(Z11.S.UXTW<<3)    // 446d2b84
    ZPRFD PLDL3STRM, P3, (R11)(Z13.S.UXTW<<3)    // 656d2d84
    ZPRFD $6, P4, (R13)(Z15.S.UXTW<<3)           // a6712f84
    ZPRFD $7, P4, (R15)(Z17.S.UXTW<<3)           // e7713184
    ZPRFD PSTL1KEEP, P5, (R17)(Z19.S.UXTW<<3)    // 28763384
    ZPRFD PSTL1STRM, P5, (R19)(Z20.S.UXTW<<3)    // 69763484
    ZPRFD PSTL2KEEP, P5, (R21)(Z22.S.UXTW<<3)    // aa763684
    ZPRFD PSTL2STRM, P6, (R23)(Z24.S.UXTW<<3)    // eb7a3884
    ZPRFD PSTL3KEEP, P6, (R24)(Z26.S.UXTW<<3)    // 0c7b3a84
    ZPRFD PSTL3STRM, P7, (R26)(Z28.S.UXTW<<3)    // 4d7f3c84
    ZPRFD $14, P7, (R27)(Z30.S.UXTW<<3)          // 6e7f3e84
    ZPRFD $15, P7, (R30)(Z31.S.UXTW<<3)          // cf7f3f84
    ZPRFD PLDL1KEEP, P0, (R0)(Z0.S.SXTW<<3)      // 00606084
    ZPRFD PLDL1STRM, P1, (R4)(Z5.S.SXTW<<3)      // 81646584
    ZPRFD PLDL2KEEP, P2, (R6)(Z7.S.SXTW<<3)      // c2686784
    ZPRFD PLDL2STRM, P2, (R8)(Z9.S.SXTW<<3)      // 03696984
    ZPRFD PLDL3KEEP, P3, (R10)(Z11.S.SXTW<<3)    // 446d6b84
    ZPRFD PLDL3STRM, P3, (R11)(Z13.S.SXTW<<3)    // 656d6d84
    ZPRFD $6, P4, (R13)(Z15.S.SXTW<<3)           // a6716f84
    ZPRFD $7, P4, (R15)(Z17.S.SXTW<<3)           // e7717184
    ZPRFD PSTL1KEEP, P5, (R17)(Z19.S.SXTW<<3)    // 28767384
    ZPRFD PSTL1STRM, P5, (R19)(Z20.S.SXTW<<3)    // 69767484
    ZPRFD PSTL2KEEP, P5, (R21)(Z22.S.SXTW<<3)    // aa767684
    ZPRFD PSTL2STRM, P6, (R23)(Z24.S.SXTW<<3)    // eb7a7884
    ZPRFD PSTL3KEEP, P6, (R24)(Z26.S.SXTW<<3)    // 0c7b7a84
    ZPRFD PSTL3STRM, P7, (R26)(Z28.S.SXTW<<3)    // 4d7f7c84
    ZPRFD $14, P7, (R27)(Z30.S.SXTW<<3)          // 6e7f7e84
    ZPRFD $15, P7, (R30)(Z31.S.SXTW<<3)          // cf7f7f84

// PRFH    <prfop>, <Pg>, [<Zn>.D{, #<imm>}]
    ZPRFH PLDL1KEEP, P0, (Z0.D)      // 00e080c4
    ZPRFH PLDL1STRM, P1, 4(Z4.D)      // 81e482c4
    ZPRFH PLDL2KEEP, P2, 8(Z6.D)      // c2e884c4
    ZPRFH PLDL2STRM, P2, 12(Z8.D)     // 03e986c4
    ZPRFH PLDL3KEEP, P3, 16(Z10.D)    // 44ed88c4
    ZPRFH PLDL3STRM, P3, 20(Z12.D)    // 85ed8ac4
    ZPRFH $6, P4, 24(Z14.D)           // c6f18cc4
    ZPRFH $7, P4, 28(Z16.D)           // 07f28ec4
    ZPRFH PSTL1KEEP, P5, 32(Z18.D)    // 48f690c4
    ZPRFH PSTL1STRM, P5, 34(Z19.D)    // 69f691c4
    ZPRFH PSTL2KEEP, P5, 38(Z21.D)    // aaf693c4
    ZPRFH PSTL2STRM, P6, 42(Z23.D)    // ebfa95c4
    ZPRFH PSTL3KEEP, P6, 46(Z25.D)    // 2cfb97c4
    ZPRFH PSTL3STRM, P7, 50(Z27.D)    // 6dff99c4
    ZPRFH $14, P7, 54(Z29.D)          // aeff9bc4
    ZPRFH $15, P7, 62(Z31.D)          // efff9fc4

// PRFH    <prfop>, <Pg>, [<Zn>.S{, #<imm>}]
    ZPRFH PLDL1KEEP, P0, (Z0.S)      // 00e08084
    ZPRFH PLDL1STRM, P1, 4(Z4.S)      // 81e48284
    ZPRFH PLDL2KEEP, P2, 8(Z6.S)      // c2e88484
    ZPRFH PLDL2STRM, P2, 12(Z8.S)     // 03e98684
    ZPRFH PLDL3KEEP, P3, 16(Z10.S)    // 44ed8884
    ZPRFH PLDL3STRM, P3, 20(Z12.S)    // 85ed8a84
    ZPRFH $6, P4, 24(Z14.S)           // c6f18c84
    ZPRFH $7, P4, 28(Z16.S)           // 07f28e84
    ZPRFH PSTL1KEEP, P5, 32(Z18.S)    // 48f69084
    ZPRFH PSTL1STRM, P5, 34(Z19.S)    // 69f69184
    ZPRFH PSTL2KEEP, P5, 38(Z21.S)    // aaf69384
    ZPRFH PSTL2STRM, P6, 42(Z23.S)    // ebfa9584
    ZPRFH PSTL3KEEP, P6, 46(Z25.S)    // 2cfb9784
    ZPRFH PSTL3STRM, P7, 50(Z27.S)    // 6dff9984
    ZPRFH $14, P7, 54(Z29.S)          // aeff9b84
    ZPRFH $15, P7, 62(Z31.S)          // efff9f84

// PRFH    <prfop>, <Pg>, [<Xn|SP>{, #<imm>, MUL VL}]
    ZPRFH PLDL1KEEP, P0, -32(R0)     // 0020e085
    ZPRFH PLDL1STRM, P1, -28(R4)     // 8124e485
    ZPRFH PLDL2KEEP, P2, -24(R6)     // c228e885
    ZPRFH PLDL2STRM, P2, -20(R8)     // 0329ec85
    ZPRFH PLDL3KEEP, P3, -16(R10)    // 442df085
    ZPRFH PLDL3STRM, P3, -12(R11)    // 652df485
    ZPRFH $6, P4, -8(R13)            // a631f885
    ZPRFH $7, P4, -4(R15)            // e731fc85
    ZPRFH PSTL1KEEP, P5, (R17)      // 2836c085
    ZPRFH PSTL1STRM, P5, 3(R19)      // 6936c385
    ZPRFH PSTL2KEEP, P5, 7(R21)      // aa36c785
    ZPRFH PSTL2STRM, P6, 11(R23)     // eb3acb85
    ZPRFH PSTL3KEEP, P6, 15(R24)     // 0c3bcf85
    ZPRFH PSTL3STRM, P7, 19(R26)     // 4d3fd385
    ZPRFH $14, P7, 23(R27)           // 6e3fd785
    ZPRFH $15, P7, 31(R30)           // cf3fdf85

// PRFH    <prfop>, <Pg>, [<Xn|SP>, <Xm>, LSL #1]
    ZPRFH PLDL1KEEP, P0, (R0)(R0<<1)      // 00c08084
    ZPRFH PLDL1STRM, P1, (R4)(R5<<1)      // 81c48584
    ZPRFH PLDL2KEEP, P2, (R6)(R7<<1)      // c2c88784
    ZPRFH PLDL2STRM, P2, (R8)(R9<<1)      // 03c98984
    ZPRFH PLDL3KEEP, P3, (R10)(R11<<1)    // 44cd8b84
    ZPRFH PLDL3STRM, P3, (R11)(R12<<1)    // 65cd8c84
    ZPRFH $6, P4, (R13)(R14<<1)           // a6d18e84
    ZPRFH $7, P4, (R15)(R16<<1)           // e7d19084
    ZPRFH PSTL1KEEP, P5, (R17)(R17<<1)    // 28d69184
    ZPRFH PSTL1STRM, P5, (R19)(R20<<1)    // 69d69484
    ZPRFH PSTL2KEEP, P5, (R21)(R22<<1)    // aad69684
    ZPRFH PSTL2STRM, P6, (R23)(R24<<1)    // ebda9884
    ZPRFH PSTL3KEEP, P6, (R24)(R25<<1)    // 0cdb9984
    ZPRFH PSTL3STRM, P7, (R26)(R27<<1)    // 4ddf9b84
    ZPRFH $14, P7, (R27)(R29<<1)          // 6edf9d84
    ZPRFH $15, P7, (R30)(R30<<1)          // cfdf9e84

// PRFH    <prfop>, <Pg>, [<Xn|SP>, <Zm>.D, LSL #1]
    ZPRFH PLDL1KEEP, P0, (R0)(Z0.D.LSL<<1)      // 00a060c4
    ZPRFH PLDL1STRM, P1, (R4)(Z5.D.LSL<<1)      // 81a465c4
    ZPRFH PLDL2KEEP, P2, (R6)(Z7.D.LSL<<1)      // c2a867c4
    ZPRFH PLDL2STRM, P2, (R8)(Z9.D.LSL<<1)      // 03a969c4
    ZPRFH PLDL3KEEP, P3, (R10)(Z11.D.LSL<<1)    // 44ad6bc4
    ZPRFH PLDL3STRM, P3, (R11)(Z13.D.LSL<<1)    // 65ad6dc4
    ZPRFH $6, P4, (R13)(Z15.D.LSL<<1)           // a6b16fc4
    ZPRFH $7, P4, (R15)(Z17.D.LSL<<1)           // e7b171c4
    ZPRFH PSTL1KEEP, P5, (R17)(Z19.D.LSL<<1)    // 28b673c4
    ZPRFH PSTL1STRM, P5, (R19)(Z20.D.LSL<<1)    // 69b674c4
    ZPRFH PSTL2KEEP, P5, (R21)(Z22.D.LSL<<1)    // aab676c4
    ZPRFH PSTL2STRM, P6, (R23)(Z24.D.LSL<<1)    // ebba78c4
    ZPRFH PSTL3KEEP, P6, (R24)(Z26.D.LSL<<1)    // 0cbb7ac4
    ZPRFH PSTL3STRM, P7, (R26)(Z28.D.LSL<<1)    // 4dbf7cc4
    ZPRFH $14, P7, (R27)(Z30.D.LSL<<1)          // 6ebf7ec4
    ZPRFH $15, P7, (R30)(Z31.D.LSL<<1)          // cfbf7fc4

// PRFH    <prfop>, <Pg>, [<Xn|SP>, <Zm>.D, <extend> #1]
    ZPRFH PLDL1KEEP, P0, (R0)(Z0.D.UXTW<<1)      // 002020c4
    ZPRFH PLDL1STRM, P1, (R4)(Z5.D.UXTW<<1)      // 812425c4
    ZPRFH PLDL2KEEP, P2, (R6)(Z7.D.UXTW<<1)      // c22827c4
    ZPRFH PLDL2STRM, P2, (R8)(Z9.D.UXTW<<1)      // 032929c4
    ZPRFH PLDL3KEEP, P3, (R10)(Z11.D.UXTW<<1)    // 442d2bc4
    ZPRFH PLDL3STRM, P3, (R11)(Z13.D.UXTW<<1)    // 652d2dc4
    ZPRFH $6, P4, (R13)(Z15.D.UXTW<<1)           // a6312fc4
    ZPRFH $7, P4, (R15)(Z17.D.UXTW<<1)           // e73131c4
    ZPRFH PSTL1KEEP, P5, (R17)(Z19.D.UXTW<<1)    // 283633c4
    ZPRFH PSTL1STRM, P5, (R19)(Z20.D.UXTW<<1)    // 693634c4
    ZPRFH PSTL2KEEP, P5, (R21)(Z22.D.UXTW<<1)    // aa3636c4
    ZPRFH PSTL2STRM, P6, (R23)(Z24.D.UXTW<<1)    // eb3a38c4
    ZPRFH PSTL3KEEP, P6, (R24)(Z26.D.UXTW<<1)    // 0c3b3ac4
    ZPRFH PSTL3STRM, P7, (R26)(Z28.D.UXTW<<1)    // 4d3f3cc4
    ZPRFH $14, P7, (R27)(Z30.D.UXTW<<1)          // 6e3f3ec4
    ZPRFH $15, P7, (R30)(Z31.D.UXTW<<1)          // cf3f3fc4
    ZPRFH PLDL1KEEP, P0, (R0)(Z0.D.SXTW<<1)      // 002060c4
    ZPRFH PLDL1STRM, P1, (R4)(Z5.D.SXTW<<1)      // 812465c4
    ZPRFH PLDL2KEEP, P2, (R6)(Z7.D.SXTW<<1)      // c22867c4
    ZPRFH PLDL2STRM, P2, (R8)(Z9.D.SXTW<<1)      // 032969c4
    ZPRFH PLDL3KEEP, P3, (R10)(Z11.D.SXTW<<1)    // 442d6bc4
    ZPRFH PLDL3STRM, P3, (R11)(Z13.D.SXTW<<1)    // 652d6dc4
    ZPRFH $6, P4, (R13)(Z15.D.SXTW<<1)           // a6316fc4
    ZPRFH $7, P4, (R15)(Z17.D.SXTW<<1)           // e73171c4
    ZPRFH PSTL1KEEP, P5, (R17)(Z19.D.SXTW<<1)    // 283673c4
    ZPRFH PSTL1STRM, P5, (R19)(Z20.D.SXTW<<1)    // 693674c4
    ZPRFH PSTL2KEEP, P5, (R21)(Z22.D.SXTW<<1)    // aa3676c4
    ZPRFH PSTL2STRM, P6, (R23)(Z24.D.SXTW<<1)    // eb3a78c4
    ZPRFH PSTL3KEEP, P6, (R24)(Z26.D.SXTW<<1)    // 0c3b7ac4
    ZPRFH PSTL3STRM, P7, (R26)(Z28.D.SXTW<<1)    // 4d3f7cc4
    ZPRFH $14, P7, (R27)(Z30.D.SXTW<<1)          // 6e3f7ec4
    ZPRFH $15, P7, (R30)(Z31.D.SXTW<<1)          // cf3f7fc4

// PRFH    <prfop>, <Pg>, [<Xn|SP>, <Zm>.S, <extend> #1]
    ZPRFH PLDL1KEEP, P0, (R0)(Z0.S.UXTW<<1)      // 00202084
    ZPRFH PLDL1STRM, P1, (R4)(Z5.S.UXTW<<1)      // 81242584
    ZPRFH PLDL2KEEP, P2, (R6)(Z7.S.UXTW<<1)      // c2282784
    ZPRFH PLDL2STRM, P2, (R8)(Z9.S.UXTW<<1)      // 03292984
    ZPRFH PLDL3KEEP, P3, (R10)(Z11.S.UXTW<<1)    // 442d2b84
    ZPRFH PLDL3STRM, P3, (R11)(Z13.S.UXTW<<1)    // 652d2d84
    ZPRFH $6, P4, (R13)(Z15.S.UXTW<<1)           // a6312f84
    ZPRFH $7, P4, (R15)(Z17.S.UXTW<<1)           // e7313184
    ZPRFH PSTL1KEEP, P5, (R17)(Z19.S.UXTW<<1)    // 28363384
    ZPRFH PSTL1STRM, P5, (R19)(Z20.S.UXTW<<1)    // 69363484
    ZPRFH PSTL2KEEP, P5, (R21)(Z22.S.UXTW<<1)    // aa363684
    ZPRFH PSTL2STRM, P6, (R23)(Z24.S.UXTW<<1)    // eb3a3884
    ZPRFH PSTL3KEEP, P6, (R24)(Z26.S.UXTW<<1)    // 0c3b3a84
    ZPRFH PSTL3STRM, P7, (R26)(Z28.S.UXTW<<1)    // 4d3f3c84
    ZPRFH $14, P7, (R27)(Z30.S.UXTW<<1)          // 6e3f3e84
    ZPRFH $15, P7, (R30)(Z31.S.UXTW<<1)          // cf3f3f84
    ZPRFH PLDL1KEEP, P0, (R0)(Z0.S.SXTW<<1)      // 00206084
    ZPRFH PLDL1STRM, P1, (R4)(Z5.S.SXTW<<1)      // 81246584
    ZPRFH PLDL2KEEP, P2, (R6)(Z7.S.SXTW<<1)      // c2286784
    ZPRFH PLDL2STRM, P2, (R8)(Z9.S.SXTW<<1)      // 03296984
    ZPRFH PLDL3KEEP, P3, (R10)(Z11.S.SXTW<<1)    // 442d6b84
    ZPRFH PLDL3STRM, P3, (R11)(Z13.S.SXTW<<1)    // 652d6d84
    ZPRFH $6, P4, (R13)(Z15.S.SXTW<<1)           // a6316f84
    ZPRFH $7, P4, (R15)(Z17.S.SXTW<<1)           // e7317184
    ZPRFH PSTL1KEEP, P5, (R17)(Z19.S.SXTW<<1)    // 28367384
    ZPRFH PSTL1STRM, P5, (R19)(Z20.S.SXTW<<1)    // 69367484
    ZPRFH PSTL2KEEP, P5, (R21)(Z22.S.SXTW<<1)    // aa367684
    ZPRFH PSTL2STRM, P6, (R23)(Z24.S.SXTW<<1)    // eb3a7884
    ZPRFH PSTL3KEEP, P6, (R24)(Z26.S.SXTW<<1)    // 0c3b7a84
    ZPRFH PSTL3STRM, P7, (R26)(Z28.S.SXTW<<1)    // 4d3f7c84
    ZPRFH $14, P7, (R27)(Z30.S.SXTW<<1)          // 6e3f7e84
    ZPRFH $15, P7, (R30)(Z31.S.SXTW<<1)          // cf3f7f84

// PRFW    <prfop>, <Pg>, [<Zn>.D{, #<imm>}]
    ZPRFW PLDL1KEEP, P0, (Z0.D)       // 00e000c5
    ZPRFW PLDL1STRM, P1, 8(Z4.D)       // 81e402c5
    ZPRFW PLDL2KEEP, P2, 16(Z6.D)      // c2e804c5
    ZPRFW PLDL2STRM, P2, 24(Z8.D)      // 03e906c5
    ZPRFW PLDL3KEEP, P3, 32(Z10.D)     // 44ed08c5
    ZPRFW PLDL3STRM, P3, 40(Z12.D)     // 85ed0ac5
    ZPRFW $6, P4, 48(Z14.D)            // c6f10cc5
    ZPRFW $7, P4, 56(Z16.D)            // 07f20ec5
    ZPRFW PSTL1KEEP, P5, 64(Z18.D)     // 48f610c5
    ZPRFW PSTL1STRM, P5, 68(Z19.D)     // 69f611c5
    ZPRFW PSTL2KEEP, P5, 76(Z21.D)     // aaf613c5
    ZPRFW PSTL2STRM, P6, 84(Z23.D)     // ebfa15c5
    ZPRFW PSTL3KEEP, P6, 92(Z25.D)     // 2cfb17c5
    ZPRFW PSTL3STRM, P7, 100(Z27.D)    // 6dff19c5
    ZPRFW $14, P7, 108(Z29.D)          // aeff1bc5
    ZPRFW $15, P7, 124(Z31.D)          // efff1fc5

// PRFW    <prfop>, <Pg>, [<Zn>.S{, #<imm>}]
    ZPRFW PLDL1KEEP, P0, (Z0.S)       // 00e00085
    ZPRFW PLDL1STRM, P1, 8(Z4.S)       // 81e40285
    ZPRFW PLDL2KEEP, P2, 16(Z6.S)      // c2e80485
    ZPRFW PLDL2STRM, P2, 24(Z8.S)      // 03e90685
    ZPRFW PLDL3KEEP, P3, 32(Z10.S)     // 44ed0885
    ZPRFW PLDL3STRM, P3, 40(Z12.S)     // 85ed0a85
    ZPRFW $6, P4, 48(Z14.S)            // c6f10c85
    ZPRFW $7, P4, 56(Z16.S)            // 07f20e85
    ZPRFW PSTL1KEEP, P5, 64(Z18.S)     // 48f61085
    ZPRFW PSTL1STRM, P5, 68(Z19.S)     // 69f61185
    ZPRFW PSTL2KEEP, P5, 76(Z21.S)     // aaf61385
    ZPRFW PSTL2STRM, P6, 84(Z23.S)     // ebfa1585
    ZPRFW PSTL3KEEP, P6, 92(Z25.S)     // 2cfb1785
    ZPRFW PSTL3STRM, P7, 100(Z27.S)    // 6dff1985
    ZPRFW $14, P7, 108(Z29.S)          // aeff1b85
    ZPRFW $15, P7, 124(Z31.S)          // efff1f85

// PRFW    <prfop>, <Pg>, [<Xn|SP>{, #<imm>, MUL VL}]
    ZPRFW PLDL1KEEP, P0, -32(R0)     // 0040e085
    ZPRFW PLDL1STRM, P1, -28(R4)     // 8144e485
    ZPRFW PLDL2KEEP, P2, -24(R6)     // c248e885
    ZPRFW PLDL2STRM, P2, -20(R8)     // 0349ec85
    ZPRFW PLDL3KEEP, P3, -16(R10)    // 444df085
    ZPRFW PLDL3STRM, P3, -12(R11)    // 654df485
    ZPRFW $6, P4, -8(R13)            // a651f885
    ZPRFW $7, P4, -4(R15)            // e751fc85
    ZPRFW PSTL1KEEP, P5, (R17)      // 2856c085
    ZPRFW PSTL1STRM, P5, 3(R19)      // 6956c385
    ZPRFW PSTL2KEEP, P5, 7(R21)      // aa56c785
    ZPRFW PSTL2STRM, P6, 11(R23)     // eb5acb85
    ZPRFW PSTL3KEEP, P6, 15(R24)     // 0c5bcf85
    ZPRFW PSTL3STRM, P7, 19(R26)     // 4d5fd385
    ZPRFW $14, P7, 23(R27)           // 6e5fd785
    ZPRFW $15, P7, 31(R30)           // cf5fdf85

// PRFW    <prfop>, <Pg>, [<Xn|SP>, <Xm>, LSL #2]
    ZPRFW PLDL1KEEP, P0, (R0)(R0<<2)      // 00c00085
    ZPRFW PLDL1STRM, P1, (R4)(R5<<2)      // 81c40585
    ZPRFW PLDL2KEEP, P2, (R6)(R7<<2)      // c2c80785
    ZPRFW PLDL2STRM, P2, (R8)(R9<<2)      // 03c90985
    ZPRFW PLDL3KEEP, P3, (R10)(R11<<2)    // 44cd0b85
    ZPRFW PLDL3STRM, P3, (R11)(R12<<2)    // 65cd0c85
    ZPRFW $6, P4, (R13)(R14<<2)           // a6d10e85
    ZPRFW $7, P4, (R15)(R16<<2)           // e7d11085
    ZPRFW PSTL1KEEP, P5, (R17)(R17<<2)    // 28d61185
    ZPRFW PSTL1STRM, P5, (R19)(R20<<2)    // 69d61485
    ZPRFW PSTL2KEEP, P5, (R21)(R22<<2)    // aad61685
    ZPRFW PSTL2STRM, P6, (R23)(R24<<2)    // ebda1885
    ZPRFW PSTL3KEEP, P6, (R24)(R25<<2)    // 0cdb1985
    ZPRFW PSTL3STRM, P7, (R26)(R27<<2)    // 4ddf1b85
    ZPRFW $14, P7, (R27)(R29<<2)          // 6edf1d85
    ZPRFW $15, P7, (R30)(R30<<2)          // cfdf1e85

// PRFW    <prfop>, <Pg>, [<Xn|SP>, <Zm>.D, LSL #2]
    ZPRFW PLDL1KEEP, P0, (R0)(Z0.D.LSL<<2)      // 00c060c4
    ZPRFW PLDL1STRM, P1, (R4)(Z5.D.LSL<<2)      // 81c465c4
    ZPRFW PLDL2KEEP, P2, (R6)(Z7.D.LSL<<2)      // c2c867c4
    ZPRFW PLDL2STRM, P2, (R8)(Z9.D.LSL<<2)      // 03c969c4
    ZPRFW PLDL3KEEP, P3, (R10)(Z11.D.LSL<<2)    // 44cd6bc4
    ZPRFW PLDL3STRM, P3, (R11)(Z13.D.LSL<<2)    // 65cd6dc4
    ZPRFW $6, P4, (R13)(Z15.D.LSL<<2)           // a6d16fc4
    ZPRFW $7, P4, (R15)(Z17.D.LSL<<2)           // e7d171c4
    ZPRFW PSTL1KEEP, P5, (R17)(Z19.D.LSL<<2)    // 28d673c4
    ZPRFW PSTL1STRM, P5, (R19)(Z20.D.LSL<<2)    // 69d674c4
    ZPRFW PSTL2KEEP, P5, (R21)(Z22.D.LSL<<2)    // aad676c4
    ZPRFW PSTL2STRM, P6, (R23)(Z24.D.LSL<<2)    // ebda78c4
    ZPRFW PSTL3KEEP, P6, (R24)(Z26.D.LSL<<2)    // 0cdb7ac4
    ZPRFW PSTL3STRM, P7, (R26)(Z28.D.LSL<<2)    // 4ddf7cc4
    ZPRFW $14, P7, (R27)(Z30.D.LSL<<2)          // 6edf7ec4
    ZPRFW $15, P7, (R30)(Z31.D.LSL<<2)          // cfdf7fc4

// PRFW    <prfop>, <Pg>, [<Xn|SP>, <Zm>.D, <extend> #2]
    ZPRFW PLDL1KEEP, P0, (R0)(Z0.D.UXTW<<2)      // 004020c4
    ZPRFW PLDL1STRM, P1, (R4)(Z5.D.UXTW<<2)      // 814425c4
    ZPRFW PLDL2KEEP, P2, (R6)(Z7.D.UXTW<<2)      // c24827c4
    ZPRFW PLDL2STRM, P2, (R8)(Z9.D.UXTW<<2)      // 034929c4
    ZPRFW PLDL3KEEP, P3, (R10)(Z11.D.UXTW<<2)    // 444d2bc4
    ZPRFW PLDL3STRM, P3, (R11)(Z13.D.UXTW<<2)    // 654d2dc4
    ZPRFW $6, P4, (R13)(Z15.D.UXTW<<2)           // a6512fc4
    ZPRFW $7, P4, (R15)(Z17.D.UXTW<<2)           // e75131c4
    ZPRFW PSTL1KEEP, P5, (R17)(Z19.D.UXTW<<2)    // 285633c4
    ZPRFW PSTL1STRM, P5, (R19)(Z20.D.UXTW<<2)    // 695634c4
    ZPRFW PSTL2KEEP, P5, (R21)(Z22.D.UXTW<<2)    // aa5636c4
    ZPRFW PSTL2STRM, P6, (R23)(Z24.D.UXTW<<2)    // eb5a38c4
    ZPRFW PSTL3KEEP, P6, (R24)(Z26.D.UXTW<<2)    // 0c5b3ac4
    ZPRFW PSTL3STRM, P7, (R26)(Z28.D.UXTW<<2)    // 4d5f3cc4
    ZPRFW $14, P7, (R27)(Z30.D.UXTW<<2)          // 6e5f3ec4
    ZPRFW $15, P7, (R30)(Z31.D.UXTW<<2)          // cf5f3fc4
    ZPRFW PLDL1KEEP, P0, (R0)(Z0.D.SXTW<<2)      // 004060c4
    ZPRFW PLDL1STRM, P1, (R4)(Z5.D.SXTW<<2)      // 814465c4
    ZPRFW PLDL2KEEP, P2, (R6)(Z7.D.SXTW<<2)      // c24867c4
    ZPRFW PLDL2STRM, P2, (R8)(Z9.D.SXTW<<2)      // 034969c4
    ZPRFW PLDL3KEEP, P3, (R10)(Z11.D.SXTW<<2)    // 444d6bc4
    ZPRFW PLDL3STRM, P3, (R11)(Z13.D.SXTW<<2)    // 654d6dc4
    ZPRFW $6, P4, (R13)(Z15.D.SXTW<<2)           // a6516fc4
    ZPRFW $7, P4, (R15)(Z17.D.SXTW<<2)           // e75171c4
    ZPRFW PSTL1KEEP, P5, (R17)(Z19.D.SXTW<<2)    // 285673c4
    ZPRFW PSTL1STRM, P5, (R19)(Z20.D.SXTW<<2)    // 695674c4
    ZPRFW PSTL2KEEP, P5, (R21)(Z22.D.SXTW<<2)    // aa5676c4
    ZPRFW PSTL2STRM, P6, (R23)(Z24.D.SXTW<<2)    // eb5a78c4
    ZPRFW PSTL3KEEP, P6, (R24)(Z26.D.SXTW<<2)    // 0c5b7ac4
    ZPRFW PSTL3STRM, P7, (R26)(Z28.D.SXTW<<2)    // 4d5f7cc4
    ZPRFW $14, P7, (R27)(Z30.D.SXTW<<2)          // 6e5f7ec4
    ZPRFW $15, P7, (R30)(Z31.D.SXTW<<2)          // cf5f7fc4

// PRFW    <prfop>, <Pg>, [<Xn|SP>, <Zm>.S, <extend> #2]
    ZPRFW PLDL1KEEP, P0, (R0)(Z0.S.UXTW<<2)      // 00402084
    ZPRFW PLDL1STRM, P1, (R4)(Z5.S.UXTW<<2)      // 81442584
    ZPRFW PLDL2KEEP, P2, (R6)(Z7.S.UXTW<<2)      // c2482784
    ZPRFW PLDL2STRM, P2, (R8)(Z9.S.UXTW<<2)      // 03492984
    ZPRFW PLDL3KEEP, P3, (R10)(Z11.S.UXTW<<2)    // 444d2b84
    ZPRFW PLDL3STRM, P3, (R11)(Z13.S.UXTW<<2)    // 654d2d84
    ZPRFW $6, P4, (R13)(Z15.S.UXTW<<2)           // a6512f84
    ZPRFW $7, P4, (R15)(Z17.S.UXTW<<2)           // e7513184
    ZPRFW PSTL1KEEP, P5, (R17)(Z19.S.UXTW<<2)    // 28563384
    ZPRFW PSTL1STRM, P5, (R19)(Z20.S.UXTW<<2)    // 69563484
    ZPRFW PSTL2KEEP, P5, (R21)(Z22.S.UXTW<<2)    // aa563684
    ZPRFW PSTL2STRM, P6, (R23)(Z24.S.UXTW<<2)    // eb5a3884
    ZPRFW PSTL3KEEP, P6, (R24)(Z26.S.UXTW<<2)    // 0c5b3a84
    ZPRFW PSTL3STRM, P7, (R26)(Z28.S.UXTW<<2)    // 4d5f3c84
    ZPRFW $14, P7, (R27)(Z30.S.UXTW<<2)          // 6e5f3e84
    ZPRFW $15, P7, (R30)(Z31.S.UXTW<<2)          // cf5f3f84
    ZPRFW PLDL1KEEP, P0, (R0)(Z0.S.SXTW<<2)      // 00406084
    ZPRFW PLDL1STRM, P1, (R4)(Z5.S.SXTW<<2)      // 81446584
    ZPRFW PLDL2KEEP, P2, (R6)(Z7.S.SXTW<<2)      // c2486784
    ZPRFW PLDL2STRM, P2, (R8)(Z9.S.SXTW<<2)      // 03496984
    ZPRFW PLDL3KEEP, P3, (R10)(Z11.S.SXTW<<2)    // 444d6b84
    ZPRFW PLDL3STRM, P3, (R11)(Z13.S.SXTW<<2)    // 654d6d84
    ZPRFW $6, P4, (R13)(Z15.S.SXTW<<2)           // a6516f84
    ZPRFW $7, P4, (R15)(Z17.S.SXTW<<2)           // e7517184
    ZPRFW PSTL1KEEP, P5, (R17)(Z19.S.SXTW<<2)    // 28567384
    ZPRFW PSTL1STRM, P5, (R19)(Z20.S.SXTW<<2)    // 69567484
    ZPRFW PSTL2KEEP, P5, (R21)(Z22.S.SXTW<<2)    // aa567684
    ZPRFW PSTL2STRM, P6, (R23)(Z24.S.SXTW<<2)    // eb5a7884
    ZPRFW PSTL3KEEP, P6, (R24)(Z26.S.SXTW<<2)    // 0c5b7a84
    ZPRFW PSTL3STRM, P7, (R26)(Z28.S.SXTW<<2)    // 4d5f7c84
    ZPRFW $14, P7, (R27)(Z30.S.SXTW<<2)          // 6e5f7e84
    ZPRFW $15, P7, (R30)(Z31.S.SXTW<<2)          // cf5f7f84

// PTEST   <Pg>, <Pn>.B
    PTEST P0, P0.B      // 00c05025
    PTEST P5, P6.B      // c0d45025
    PTEST P15, P15.B    // e0fd5025

// PTRUE   <Pd>.<T>{, <pattern>}
    PTRUE POW2, P0.B     // 00e01825
    PTRUE VL1, P0.B      // 20e01825
    PTRUE VL2, P1.B      // 41e01825
    PTRUE VL3, P1.B      // 61e01825
    PTRUE VL4, P2.B      // 82e01825
    PTRUE VL5, P2.B      // a2e01825
    PTRUE VL6, P3.B      // c3e01825
    PTRUE VL7, P3.B      // e3e01825
    PTRUE VL8, P4.B      // 04e11825
    PTRUE VL16, P4.B     // 24e11825
    PTRUE VL32, P5.B     // 45e11825
    PTRUE VL64, P5.B     // 65e11825
    PTRUE VL128, P6.B    // 86e11825
    PTRUE VL256, P6.B    // a6e11825
    PTRUE $14, P7.B      // c7e11825
    PTRUE $15, P7.B      // e7e11825
    PTRUE $16, P8.B      // 08e21825
    PTRUE $17, P8.B      // 28e21825
    PTRUE $18, P8.B      // 48e21825
    PTRUE $19, P9.B      // 69e21825
    PTRUE $20, P9.B      // 89e21825
    PTRUE $21, P10.B     // aae21825
    PTRUE $22, P10.B     // cae21825
    PTRUE $23, P11.B     // ebe21825
    PTRUE $24, P11.B     // 0be31825
    PTRUE $25, P12.B     // 2ce31825
    PTRUE $26, P12.B     // 4ce31825
    PTRUE $27, P13.B     // 6de31825
    PTRUE $28, P13.B     // 8de31825
    PTRUE MUL4, P14.B    // aee31825
    PTRUE MUL3, P14.B    // cee31825
    PTRUE ALL, P15.B     // efe31825
    PTRUE POW2, P0.H     // 00e05825
    PTRUE VL1, P0.H      // 20e05825
    PTRUE VL2, P1.H      // 41e05825
    PTRUE VL3, P1.H      // 61e05825
    PTRUE VL4, P2.H      // 82e05825
    PTRUE VL5, P2.H      // a2e05825
    PTRUE VL6, P3.H      // c3e05825
    PTRUE VL7, P3.H      // e3e05825
    PTRUE VL8, P4.H      // 04e15825
    PTRUE VL16, P4.H     // 24e15825
    PTRUE VL32, P5.H     // 45e15825
    PTRUE VL64, P5.H     // 65e15825
    PTRUE VL128, P6.H    // 86e15825
    PTRUE VL256, P6.H    // a6e15825
    PTRUE $14, P7.H      // c7e15825
    PTRUE $15, P7.H      // e7e15825
    PTRUE $16, P8.H      // 08e25825
    PTRUE $17, P8.H      // 28e25825
    PTRUE $18, P8.H      // 48e25825
    PTRUE $19, P9.H      // 69e25825
    PTRUE $20, P9.H      // 89e25825
    PTRUE $21, P10.H     // aae25825
    PTRUE $22, P10.H     // cae25825
    PTRUE $23, P11.H     // ebe25825
    PTRUE $24, P11.H     // 0be35825
    PTRUE $25, P12.H     // 2ce35825
    PTRUE $26, P12.H     // 4ce35825
    PTRUE $27, P13.H     // 6de35825
    PTRUE $28, P13.H     // 8de35825
    PTRUE MUL4, P14.H    // aee35825
    PTRUE MUL3, P14.H    // cee35825
    PTRUE ALL, P15.H     // efe35825
    PTRUE POW2, P0.S     // 00e09825
    PTRUE VL1, P0.S      // 20e09825
    PTRUE VL2, P1.S      // 41e09825
    PTRUE VL3, P1.S      // 61e09825
    PTRUE VL4, P2.S      // 82e09825
    PTRUE VL5, P2.S      // a2e09825
    PTRUE VL6, P3.S      // c3e09825
    PTRUE VL7, P3.S      // e3e09825
    PTRUE VL8, P4.S      // 04e19825
    PTRUE VL16, P4.S     // 24e19825
    PTRUE VL32, P5.S     // 45e19825
    PTRUE VL64, P5.S     // 65e19825
    PTRUE VL128, P6.S    // 86e19825
    PTRUE VL256, P6.S    // a6e19825
    PTRUE $14, P7.S      // c7e19825
    PTRUE $15, P7.S      // e7e19825
    PTRUE $16, P8.S      // 08e29825
    PTRUE $17, P8.S      // 28e29825
    PTRUE $18, P8.S      // 48e29825
    PTRUE $19, P9.S      // 69e29825
    PTRUE $20, P9.S      // 89e29825
    PTRUE $21, P10.S     // aae29825
    PTRUE $22, P10.S     // cae29825
    PTRUE $23, P11.S     // ebe29825
    PTRUE $24, P11.S     // 0be39825
    PTRUE $25, P12.S     // 2ce39825
    PTRUE $26, P12.S     // 4ce39825
    PTRUE $27, P13.S     // 6de39825
    PTRUE $28, P13.S     // 8de39825
    PTRUE MUL4, P14.S    // aee39825
    PTRUE MUL3, P14.S    // cee39825
    PTRUE ALL, P15.S     // efe39825
    PTRUE POW2, P0.D     // 00e0d825
    PTRUE VL1, P0.D      // 20e0d825
    PTRUE VL2, P1.D      // 41e0d825
    PTRUE VL3, P1.D      // 61e0d825
    PTRUE VL4, P2.D      // 82e0d825
    PTRUE VL5, P2.D      // a2e0d825
    PTRUE VL6, P3.D      // c3e0d825
    PTRUE VL7, P3.D      // e3e0d825
    PTRUE VL8, P4.D      // 04e1d825
    PTRUE VL16, P4.D     // 24e1d825
    PTRUE VL32, P5.D     // 45e1d825
    PTRUE VL64, P5.D     // 65e1d825
    PTRUE VL128, P6.D    // 86e1d825
    PTRUE VL256, P6.D    // a6e1d825
    PTRUE $14, P7.D      // c7e1d825
    PTRUE $15, P7.D      // e7e1d825
    PTRUE $16, P8.D      // 08e2d825
    PTRUE $17, P8.D      // 28e2d825
    PTRUE $18, P8.D      // 48e2d825
    PTRUE $19, P9.D      // 69e2d825
    PTRUE $20, P9.D      // 89e2d825
    PTRUE $21, P10.D     // aae2d825
    PTRUE $22, P10.D     // cae2d825
    PTRUE $23, P11.D     // ebe2d825
    PTRUE $24, P11.D     // 0be3d825
    PTRUE $25, P12.D     // 2ce3d825
    PTRUE $26, P12.D     // 4ce3d825
    PTRUE $27, P13.D     // 6de3d825
    PTRUE $28, P13.D     // 8de3d825
    PTRUE MUL4, P14.D    // aee3d825
    PTRUE MUL3, P14.D    // cee3d825
    PTRUE ALL, P15.D     // efe3d825

// PTRUES  <Pd>.<T>{, <pattern>}
    PTRUES POW2, P0.B     // 00e01925
    PTRUES VL1, P0.B      // 20e01925
    PTRUES VL2, P1.B      // 41e01925
    PTRUES VL3, P1.B      // 61e01925
    PTRUES VL4, P2.B      // 82e01925
    PTRUES VL5, P2.B      // a2e01925
    PTRUES VL6, P3.B      // c3e01925
    PTRUES VL7, P3.B      // e3e01925
    PTRUES VL8, P4.B      // 04e11925
    PTRUES VL16, P4.B     // 24e11925
    PTRUES VL32, P5.B     // 45e11925
    PTRUES VL64, P5.B     // 65e11925
    PTRUES VL128, P6.B    // 86e11925
    PTRUES VL256, P6.B    // a6e11925
    PTRUES $14, P7.B      // c7e11925
    PTRUES $15, P7.B      // e7e11925
    PTRUES $16, P8.B      // 08e21925
    PTRUES $17, P8.B      // 28e21925
    PTRUES $18, P8.B      // 48e21925
    PTRUES $19, P9.B      // 69e21925
    PTRUES $20, P9.B      // 89e21925
    PTRUES $21, P10.B     // aae21925
    PTRUES $22, P10.B     // cae21925
    PTRUES $23, P11.B     // ebe21925
    PTRUES $24, P11.B     // 0be31925
    PTRUES $25, P12.B     // 2ce31925
    PTRUES $26, P12.B     // 4ce31925
    PTRUES $27, P13.B     // 6de31925
    PTRUES $28, P13.B     // 8de31925
    PTRUES MUL4, P14.B    // aee31925
    PTRUES MUL3, P14.B    // cee31925
    PTRUES ALL, P15.B     // efe31925
    PTRUES POW2, P0.H     // 00e05925
    PTRUES VL1, P0.H      // 20e05925
    PTRUES VL2, P1.H      // 41e05925
    PTRUES VL3, P1.H      // 61e05925
    PTRUES VL4, P2.H      // 82e05925
    PTRUES VL5, P2.H      // a2e05925
    PTRUES VL6, P3.H      // c3e05925
    PTRUES VL7, P3.H      // e3e05925
    PTRUES VL8, P4.H      // 04e15925
    PTRUES VL16, P4.H     // 24e15925
    PTRUES VL32, P5.H     // 45e15925
    PTRUES VL64, P5.H     // 65e15925
    PTRUES VL128, P6.H    // 86e15925
    PTRUES VL256, P6.H    // a6e15925
    PTRUES $14, P7.H      // c7e15925
    PTRUES $15, P7.H      // e7e15925
    PTRUES $16, P8.H      // 08e25925
    PTRUES $17, P8.H      // 28e25925
    PTRUES $18, P8.H      // 48e25925
    PTRUES $19, P9.H      // 69e25925
    PTRUES $20, P9.H      // 89e25925
    PTRUES $21, P10.H     // aae25925
    PTRUES $22, P10.H     // cae25925
    PTRUES $23, P11.H     // ebe25925
    PTRUES $24, P11.H     // 0be35925
    PTRUES $25, P12.H     // 2ce35925
    PTRUES $26, P12.H     // 4ce35925
    PTRUES $27, P13.H     // 6de35925
    PTRUES $28, P13.H     // 8de35925
    PTRUES MUL4, P14.H    // aee35925
    PTRUES MUL3, P14.H    // cee35925
    PTRUES ALL, P15.H     // efe35925
    PTRUES POW2, P0.S     // 00e09925
    PTRUES VL1, P0.S      // 20e09925
    PTRUES VL2, P1.S      // 41e09925
    PTRUES VL3, P1.S      // 61e09925
    PTRUES VL4, P2.S      // 82e09925
    PTRUES VL5, P2.S      // a2e09925
    PTRUES VL6, P3.S      // c3e09925
    PTRUES VL7, P3.S      // e3e09925
    PTRUES VL8, P4.S      // 04e19925
    PTRUES VL16, P4.S     // 24e19925
    PTRUES VL32, P5.S     // 45e19925
    PTRUES VL64, P5.S     // 65e19925
    PTRUES VL128, P6.S    // 86e19925
    PTRUES VL256, P6.S    // a6e19925
    PTRUES $14, P7.S      // c7e19925
    PTRUES $15, P7.S      // e7e19925
    PTRUES $16, P8.S      // 08e29925
    PTRUES $17, P8.S      // 28e29925
    PTRUES $18, P8.S      // 48e29925
    PTRUES $19, P9.S      // 69e29925
    PTRUES $20, P9.S      // 89e29925
    PTRUES $21, P10.S     // aae29925
    PTRUES $22, P10.S     // cae29925
    PTRUES $23, P11.S     // ebe29925
    PTRUES $24, P11.S     // 0be39925
    PTRUES $25, P12.S     // 2ce39925
    PTRUES $26, P12.S     // 4ce39925
    PTRUES $27, P13.S     // 6de39925
    PTRUES $28, P13.S     // 8de39925
    PTRUES MUL4, P14.S    // aee39925
    PTRUES MUL3, P14.S    // cee39925
    PTRUES ALL, P15.S     // efe39925
    PTRUES POW2, P0.D     // 00e0d925
    PTRUES VL1, P0.D      // 20e0d925
    PTRUES VL2, P1.D      // 41e0d925
    PTRUES VL3, P1.D      // 61e0d925
    PTRUES VL4, P2.D      // 82e0d925
    PTRUES VL5, P2.D      // a2e0d925
    PTRUES VL6, P3.D      // c3e0d925
    PTRUES VL7, P3.D      // e3e0d925
    PTRUES VL8, P4.D      // 04e1d925
    PTRUES VL16, P4.D     // 24e1d925
    PTRUES VL32, P5.D     // 45e1d925
    PTRUES VL64, P5.D     // 65e1d925
    PTRUES VL128, P6.D    // 86e1d925
    PTRUES VL256, P6.D    // a6e1d925
    PTRUES $14, P7.D      // c7e1d925
    PTRUES $15, P7.D      // e7e1d925
    PTRUES $16, P8.D      // 08e2d925
    PTRUES $17, P8.D      // 28e2d925
    PTRUES $18, P8.D      // 48e2d925
    PTRUES $19, P9.D      // 69e2d925
    PTRUES $20, P9.D      // 89e2d925
    PTRUES $21, P10.D     // aae2d925
    PTRUES $22, P10.D     // cae2d925
    PTRUES $23, P11.D     // ebe2d925
    PTRUES $24, P11.D     // 0be3d925
    PTRUES $25, P12.D     // 2ce3d925
    PTRUES $26, P12.D     // 4ce3d925
    PTRUES $27, P13.D     // 6de3d925
    PTRUES $28, P13.D     // 8de3d925
    PTRUES MUL4, P14.D    // aee3d925
    PTRUES MUL3, P14.D    // cee3d925
    PTRUES ALL, P15.D     // efe3d925

// PUNPKHI <Pd>.H, <Pn>.B
    PUNPKHI P0.B, P0.H      // 00403105
    PUNPKHI P6.B, P5.H      // c5403105
    PUNPKHI P15.B, P15.H    // ef413105

// PUNPKLO <Pd>.H, <Pn>.B
    PUNPKLO P0.B, P0.H      // 00403005
    PUNPKLO P6.B, P5.H      // c5403005
    PUNPKLO P15.B, P15.H    // ef413005

// RBIT    <Zd>.<T>, <Pg>/M, <Zn>.<T>
    ZRBIT P0.M, Z0.B, Z0.B      // 00802705
    ZRBIT P3.M, Z12.B, Z10.B    // 8a8d2705
    ZRBIT P7.M, Z31.B, Z31.B    // ff9f2705
    ZRBIT P0.M, Z0.H, Z0.H      // 00806705
    ZRBIT P3.M, Z12.H, Z10.H    // 8a8d6705
    ZRBIT P7.M, Z31.H, Z31.H    // ff9f6705
    ZRBIT P0.M, Z0.S, Z0.S      // 0080a705
    ZRBIT P3.M, Z12.S, Z10.S    // 8a8da705
    ZRBIT P7.M, Z31.S, Z31.S    // ff9fa705
    ZRBIT P0.M, Z0.D, Z0.D      // 0080e705
    ZRBIT P3.M, Z12.D, Z10.D    // 8a8de705
    ZRBIT P7.M, Z31.D, Z31.D    // ff9fe705

// RDFFR   <Pd>.B
    PRDFFR P0.B     // 00f01925
    PRDFFR P5.B     // 05f01925
    PRDFFR P15.B    // 0ff01925

// RDFFR   <Pd>.B, <Pg>/Z
    PRDFFR P0.Z, P0.B      // 00f01825
    PRDFFR P6.Z, P5.B      // c5f01825
    PRDFFR P15.Z, P15.B    // eff11825

// RDFFRS  <Pd>.B, <Pg>/Z
    PRDFFRS P0.Z, P0.B      // 00f05825
    PRDFFRS P6.Z, P5.B      // c5f05825
    PRDFFRS P15.Z, P15.B    // eff15825

// RDVL    <Xd>, #<imm>
    ZRDVL $-32, R0     // 0054bf04
    ZRDVL $-11, R10    // aa56bf04
    ZRDVL $31, R30     // fe53bf04

// REV     <Pd>.<T>, <Pn>.<T>
    PREV P0.B, P0.B      // 00403405
    PREV P6.B, P5.B      // c5403405
    PREV P15.B, P15.B    // ef413405
    PREV P0.H, P0.H      // 00407405
    PREV P6.H, P5.H      // c5407405
    PREV P15.H, P15.H    // ef417405
    PREV P0.S, P0.S      // 0040b405
    PREV P6.S, P5.S      // c540b405
    PREV P15.S, P15.S    // ef41b405
    PREV P0.D, P0.D      // 0040f405
    PREV P6.D, P5.D      // c540f405
    PREV P15.D, P15.D    // ef41f405

// REV     <Zd>.<T>, <Zn>.<T>
    ZREV Z0.B, Z0.B      // 00383805
    ZREV Z11.B, Z10.B    // 6a393805
    ZREV Z31.B, Z31.B    // ff3b3805
    ZREV Z0.H, Z0.H      // 00387805
    ZREV Z11.H, Z10.H    // 6a397805
    ZREV Z31.H, Z31.H    // ff3b7805
    ZREV Z0.S, Z0.S      // 0038b805
    ZREV Z11.S, Z10.S    // 6a39b805
    ZREV Z31.S, Z31.S    // ff3bb805
    ZREV Z0.D, Z0.D      // 0038f805
    ZREV Z11.D, Z10.D    // 6a39f805
    ZREV Z31.D, Z31.D    // ff3bf805

// REVB    <Zd>.<T>, <Pg>/M, <Zn>.<T>
    ZREVB P0.M, Z0.H, Z0.H      // 00806405
    ZREVB P3.M, Z12.H, Z10.H    // 8a8d6405
    ZREVB P7.M, Z31.H, Z31.H    // ff9f6405
    ZREVB P0.M, Z0.S, Z0.S      // 0080a405
    ZREVB P3.M, Z12.S, Z10.S    // 8a8da405
    ZREVB P7.M, Z31.S, Z31.S    // ff9fa405
    ZREVB P0.M, Z0.D, Z0.D      // 0080e405
    ZREVB P3.M, Z12.D, Z10.D    // 8a8de405
    ZREVB P7.M, Z31.D, Z31.D    // ff9fe405

// REVH    <Zd>.<T>, <Pg>/M, <Zn>.<T>
    ZREVH P0.M, Z0.S, Z0.S      // 0080a505
    ZREVH P3.M, Z12.S, Z10.S    // 8a8da505
    ZREVH P7.M, Z31.S, Z31.S    // ff9fa505
    ZREVH P0.M, Z0.D, Z0.D      // 0080e505
    ZREVH P3.M, Z12.D, Z10.D    // 8a8de505
    ZREVH P7.M, Z31.D, Z31.D    // ff9fe505

// REVW    <Zd>.D, <Pg>/M, <Zn>.D
    ZREVW P0.M, Z0.D, Z0.D      // 0080e605
    ZREVW P3.M, Z12.D, Z10.D    // 8a8de605
    ZREVW P7.M, Z31.D, Z31.D    // ff9fe605

// SABD    <Zdn>.<T>, <Pg>/M, <Zdn>.<T>, <Zm>.<T>
    ZSABD P0.M, Z0.B, Z0.B, Z0.B       // 00000c04
    ZSABD P3.M, Z10.B, Z12.B, Z10.B    // 8a0d0c04
    ZSABD P7.M, Z31.B, Z31.B, Z31.B    // ff1f0c04
    ZSABD P0.M, Z0.H, Z0.H, Z0.H       // 00004c04
    ZSABD P3.M, Z10.H, Z12.H, Z10.H    // 8a0d4c04
    ZSABD P7.M, Z31.H, Z31.H, Z31.H    // ff1f4c04
    ZSABD P0.M, Z0.S, Z0.S, Z0.S       // 00008c04
    ZSABD P3.M, Z10.S, Z12.S, Z10.S    // 8a0d8c04
    ZSABD P7.M, Z31.S, Z31.S, Z31.S    // ff1f8c04
    ZSABD P0.M, Z0.D, Z0.D, Z0.D       // 0000cc04
    ZSABD P3.M, Z10.D, Z12.D, Z10.D    // 8a0dcc04
    ZSABD P7.M, Z31.D, Z31.D, Z31.D    // ff1fcc04

// SADDV   <Dd>, <Pg>, <Zn>.<T>
    ZSADDV P0, Z0.B, V0      // 00200004
    ZSADDV P3, Z12.B, V10    // 8a2d0004
    ZSADDV P7, Z31.B, V31    // ff3f0004
    ZSADDV P0, Z0.H, V0      // 00204004
    ZSADDV P3, Z12.H, V10    // 8a2d4004
    ZSADDV P7, Z31.H, V31    // ff3f4004
    ZSADDV P0, Z0.S, V0      // 00208004
    ZSADDV P3, Z12.S, V10    // 8a2d8004
    ZSADDV P7, Z31.S, V31    // ff3f8004

// SCVTF   <Zd>.H, <Pg>/M, <Zn>.H
    ZSCVTF P0.M, Z0.H, Z0.H      // 00a05265
    ZSCVTF P3.M, Z12.H, Z10.H    // 8aad5265
    ZSCVTF P7.M, Z31.H, Z31.H    // ffbf5265

// SCVTF   <Zd>.D, <Pg>/M, <Zn>.S
    ZSCVTF P0.M, Z0.S, Z0.D      // 00a0d065
    ZSCVTF P3.M, Z12.S, Z10.D    // 8aadd065
    ZSCVTF P7.M, Z31.S, Z31.D    // ffbfd065

// SCVTF   <Zd>.H, <Pg>/M, <Zn>.S
    ZSCVTF P0.M, Z0.S, Z0.H      // 00a05465
    ZSCVTF P3.M, Z12.S, Z10.H    // 8aad5465
    ZSCVTF P7.M, Z31.S, Z31.H    // ffbf5465

// SCVTF   <Zd>.S, <Pg>/M, <Zn>.S
    ZSCVTF P0.M, Z0.S, Z0.S      // 00a09465
    ZSCVTF P3.M, Z12.S, Z10.S    // 8aad9465
    ZSCVTF P7.M, Z31.S, Z31.S    // ffbf9465

// SCVTF   <Zd>.D, <Pg>/M, <Zn>.D
    ZSCVTF P0.M, Z0.D, Z0.D      // 00a0d665
    ZSCVTF P3.M, Z12.D, Z10.D    // 8aadd665
    ZSCVTF P7.M, Z31.D, Z31.D    // ffbfd665

// SCVTF   <Zd>.H, <Pg>/M, <Zn>.D
    ZSCVTF P0.M, Z0.D, Z0.H      // 00a05665
    ZSCVTF P3.M, Z12.D, Z10.H    // 8aad5665
    ZSCVTF P7.M, Z31.D, Z31.H    // ffbf5665

// SCVTF   <Zd>.S, <Pg>/M, <Zn>.D
    ZSCVTF P0.M, Z0.D, Z0.S      // 00a0d465
    ZSCVTF P3.M, Z12.D, Z10.S    // 8aadd465
    ZSCVTF P7.M, Z31.D, Z31.S    // ffbfd465

// SDIV    <Zdn>.<T>, <Pg>/M, <Zdn>.<T>, <Zm>.<T>
    ZSDIV P0.M, Z0.S, Z0.S, Z0.S       // 00009404
    ZSDIV P3.M, Z10.S, Z12.S, Z10.S    // 8a0d9404
    ZSDIV P7.M, Z31.S, Z31.S, Z31.S    // ff1f9404
    ZSDIV P0.M, Z0.D, Z0.D, Z0.D       // 0000d404
    ZSDIV P3.M, Z10.D, Z12.D, Z10.D    // 8a0dd404
    ZSDIV P7.M, Z31.D, Z31.D, Z31.D    // ff1fd404

// SDIVR   <Zdn>.<T>, <Pg>/M, <Zdn>.<T>, <Zm>.<T>
    ZSDIVR P0.M, Z0.S, Z0.S, Z0.S       // 00009604
    ZSDIVR P3.M, Z10.S, Z12.S, Z10.S    // 8a0d9604
    ZSDIVR P7.M, Z31.S, Z31.S, Z31.S    // ff1f9604
    ZSDIVR P0.M, Z0.D, Z0.D, Z0.D       // 0000d604
    ZSDIVR P3.M, Z10.D, Z12.D, Z10.D    // 8a0dd604
    ZSDIVR P7.M, Z31.D, Z31.D, Z31.D    // ff1fd604

// SDOT    <Zda>.<T>, <Zn>.<Tb>, <Zm>.<Tb>
    ZSDOT Z0.B, Z0.B, Z0.S       // 00008044
    ZSDOT Z11.B, Z12.B, Z10.S    // 6a018c44
    ZSDOT Z31.B, Z31.B, Z31.S    // ff039f44
    ZSDOT Z0.H, Z0.H, Z0.D       // 0000c044
    ZSDOT Z11.H, Z12.H, Z10.D    // 6a01cc44
    ZSDOT Z31.H, Z31.H, Z31.D    // ff03df44

// SDOT    <Zda>.D, <Zn>.H, <Zm>.H[<imm>]
    ZSDOT Z0.H, Z0.H[0], Z0.D       // 0000e044
    ZSDOT Z11.H, Z7.H[0], Z10.D     // 6a01e744
    ZSDOT Z31.H, Z15.H[1], Z31.D    // ff03ff44

// SDOT    <Zda>.S, <Zn>.B, <Zm>.B[<imm>]
    ZSDOT Z0.B, Z0.B[0], Z0.S      // 0000a044
    ZSDOT Z11.B, Z4.B[1], Z10.S    // 6a01ac44
    ZSDOT Z31.B, Z7.B[3], Z31.S    // ff03bf44

// SEL     <Pd>.B, <Pg>, <Pn>.B, <Pm>.B
    PSEL P0, P0.B, P0.B, P0.B        // 10420025
    PSEL P6, P7.B, P8.B, P5.B        // f55a0825
    PSEL P15, P15.B, P15.B, P15.B    // ff7f0f25

// SEL     <Zd>.<T>, <Pv>, <Zn>.<T>, <Zm>.<T>
    ZSEL P0, Z0.B, Z0.B, Z0.B        // 00c02005
    ZSEL P6, Z12.B, Z13.B, Z10.B     // 8ad92d05
    ZSEL P15, Z31.B, Z31.B, Z31.B    // ffff3f05
    ZSEL P0, Z0.H, Z0.H, Z0.H        // 00c06005
    ZSEL P6, Z12.H, Z13.H, Z10.H     // 8ad96d05
    ZSEL P15, Z31.H, Z31.H, Z31.H    // ffff7f05
    ZSEL P0, Z0.S, Z0.S, Z0.S        // 00c0a005
    ZSEL P6, Z12.S, Z13.S, Z10.S     // 8ad9ad05
    ZSEL P15, Z31.S, Z31.S, Z31.S    // ffffbf05
    ZSEL P0, Z0.D, Z0.D, Z0.D        // 00c0e005
    ZSEL P6, Z12.D, Z13.D, Z10.D     // 8ad9ed05
    ZSEL P15, Z31.D, Z31.D, Z31.D    // ffffff05

// SETFFR  
    ZSETFFR    // 00902c25
    ZSETFFR    // 00902c25
    ZSETFFR    // 00902c25

// SMAX    <Zdn>.<T>, <Pg>/M, <Zdn>.<T>, <Zm>.<T>
    ZSMAX P0.M, Z0.B, Z0.B, Z0.B       // 00000804
    ZSMAX P3.M, Z10.B, Z12.B, Z10.B    // 8a0d0804
    ZSMAX P7.M, Z31.B, Z31.B, Z31.B    // ff1f0804
    ZSMAX P0.M, Z0.H, Z0.H, Z0.H       // 00004804
    ZSMAX P3.M, Z10.H, Z12.H, Z10.H    // 8a0d4804
    ZSMAX P7.M, Z31.H, Z31.H, Z31.H    // ff1f4804
    ZSMAX P0.M, Z0.S, Z0.S, Z0.S       // 00008804
    ZSMAX P3.M, Z10.S, Z12.S, Z10.S    // 8a0d8804
    ZSMAX P7.M, Z31.S, Z31.S, Z31.S    // ff1f8804
    ZSMAX P0.M, Z0.D, Z0.D, Z0.D       // 0000c804
    ZSMAX P3.M, Z10.D, Z12.D, Z10.D    // 8a0dc804
    ZSMAX P7.M, Z31.D, Z31.D, Z31.D    // ff1fc804

// SMAX    <Zdn>.<T>, <Zdn>.<T>, #<imm>
    ZSMAX Z0.B, $-128, Z0.B     // 00d02825
    ZSMAX Z10.B, $-43, Z10.B    // aada2825
    ZSMAX Z31.B, $127, Z31.B    // ffcf2825
    ZSMAX Z0.H, $-128, Z0.H     // 00d06825
    ZSMAX Z10.H, $-43, Z10.H    // aada6825
    ZSMAX Z31.H, $127, Z31.H    // ffcf6825
    ZSMAX Z0.S, $-128, Z0.S     // 00d0a825
    ZSMAX Z10.S, $-43, Z10.S    // aadaa825
    ZSMAX Z31.S, $127, Z31.S    // ffcfa825
    ZSMAX Z0.D, $-128, Z0.D     // 00d0e825
    ZSMAX Z10.D, $-43, Z10.D    // aadae825
    ZSMAX Z31.D, $127, Z31.D    // ffcfe825

// SMAXV   <V><d>, <Pg>, <Zn>.<T>
    ZSMAXV P0, Z0.B, V0      // 00200804
    ZSMAXV P3, Z12.B, V10    // 8a2d0804
    ZSMAXV P7, Z31.B, V31    // ff3f0804
    ZSMAXV P0, Z0.H, V0      // 00204804
    ZSMAXV P3, Z12.H, V10    // 8a2d4804
    ZSMAXV P7, Z31.H, V31    // ff3f4804
    ZSMAXV P0, Z0.S, V0      // 00208804
    ZSMAXV P3, Z12.S, V10    // 8a2d8804
    ZSMAXV P7, Z31.S, V31    // ff3f8804
    ZSMAXV P0, Z0.D, V0      // 0020c804
    ZSMAXV P3, Z12.D, V10    // 8a2dc804
    ZSMAXV P7, Z31.D, V31    // ff3fc804

// SMIN    <Zdn>.<T>, <Pg>/M, <Zdn>.<T>, <Zm>.<T>
    ZSMIN P0.M, Z0.B, Z0.B, Z0.B       // 00000a04
    ZSMIN P3.M, Z10.B, Z12.B, Z10.B    // 8a0d0a04
    ZSMIN P7.M, Z31.B, Z31.B, Z31.B    // ff1f0a04
    ZSMIN P0.M, Z0.H, Z0.H, Z0.H       // 00004a04
    ZSMIN P3.M, Z10.H, Z12.H, Z10.H    // 8a0d4a04
    ZSMIN P7.M, Z31.H, Z31.H, Z31.H    // ff1f4a04
    ZSMIN P0.M, Z0.S, Z0.S, Z0.S       // 00008a04
    ZSMIN P3.M, Z10.S, Z12.S, Z10.S    // 8a0d8a04
    ZSMIN P7.M, Z31.S, Z31.S, Z31.S    // ff1f8a04
    ZSMIN P0.M, Z0.D, Z0.D, Z0.D       // 0000ca04
    ZSMIN P3.M, Z10.D, Z12.D, Z10.D    // 8a0dca04
    ZSMIN P7.M, Z31.D, Z31.D, Z31.D    // ff1fca04

// SMIN    <Zdn>.<T>, <Zdn>.<T>, #<imm>
    ZSMIN Z0.B, $-128, Z0.B     // 00d02a25
    ZSMIN Z10.B, $-43, Z10.B    // aada2a25
    ZSMIN Z31.B, $127, Z31.B    // ffcf2a25
    ZSMIN Z0.H, $-128, Z0.H     // 00d06a25
    ZSMIN Z10.H, $-43, Z10.H    // aada6a25
    ZSMIN Z31.H, $127, Z31.H    // ffcf6a25
    ZSMIN Z0.S, $-128, Z0.S     // 00d0aa25
    ZSMIN Z10.S, $-43, Z10.S    // aadaaa25
    ZSMIN Z31.S, $127, Z31.S    // ffcfaa25
    ZSMIN Z0.D, $-128, Z0.D     // 00d0ea25
    ZSMIN Z10.D, $-43, Z10.D    // aadaea25
    ZSMIN Z31.D, $127, Z31.D    // ffcfea25

// SMINV   <V><d>, <Pg>, <Zn>.<T>
    ZSMINV P0, Z0.B, V0      // 00200a04
    ZSMINV P3, Z12.B, V10    // 8a2d0a04
    ZSMINV P7, Z31.B, V31    // ff3f0a04
    ZSMINV P0, Z0.H, V0      // 00204a04
    ZSMINV P3, Z12.H, V10    // 8a2d4a04
    ZSMINV P7, Z31.H, V31    // ff3f4a04
    ZSMINV P0, Z0.S, V0      // 00208a04
    ZSMINV P3, Z12.S, V10    // 8a2d8a04
    ZSMINV P7, Z31.S, V31    // ff3f8a04
    ZSMINV P0, Z0.D, V0      // 0020ca04
    ZSMINV P3, Z12.D, V10    // 8a2dca04
    ZSMINV P7, Z31.D, V31    // ff3fca04

// SMMLA   <Zda>.S, <Zn>.B, <Zm>.B
    ZSMMLA Z0.B, Z0.B, Z0.S       // 00980045
    ZSMMLA Z11.B, Z12.B, Z10.S    // 6a990c45
    ZSMMLA Z31.B, Z31.B, Z31.S    // ff9b1f45

// SMULH   <Zdn>.<T>, <Pg>/M, <Zdn>.<T>, <Zm>.<T>
    ZSMULH P0.M, Z0.B, Z0.B, Z0.B       // 00001204
    ZSMULH P3.M, Z10.B, Z12.B, Z10.B    // 8a0d1204
    ZSMULH P7.M, Z31.B, Z31.B, Z31.B    // ff1f1204
    ZSMULH P0.M, Z0.H, Z0.H, Z0.H       // 00005204
    ZSMULH P3.M, Z10.H, Z12.H, Z10.H    // 8a0d5204
    ZSMULH P7.M, Z31.H, Z31.H, Z31.H    // ff1f5204
    ZSMULH P0.M, Z0.S, Z0.S, Z0.S       // 00009204
    ZSMULH P3.M, Z10.S, Z12.S, Z10.S    // 8a0d9204
    ZSMULH P7.M, Z31.S, Z31.S, Z31.S    // ff1f9204
    ZSMULH P0.M, Z0.D, Z0.D, Z0.D       // 0000d204
    ZSMULH P3.M, Z10.D, Z12.D, Z10.D    // 8a0dd204
    ZSMULH P7.M, Z31.D, Z31.D, Z31.D    // ff1fd204

// SPLICE  <Zdn>.<T>, <Pv>, <Zdn>.<T>, <Zm>.<T>
    ZSPLICE P0, Z0.B, Z0.B, Z0.B       // 00802c05
    ZSPLICE P3, Z10.B, Z12.B, Z10.B    // 8a8d2c05
    ZSPLICE P7, Z31.B, Z31.B, Z31.B    // ff9f2c05
    ZSPLICE P0, Z0.H, Z0.H, Z0.H       // 00806c05
    ZSPLICE P3, Z10.H, Z12.H, Z10.H    // 8a8d6c05
    ZSPLICE P7, Z31.H, Z31.H, Z31.H    // ff9f6c05
    ZSPLICE P0, Z0.S, Z0.S, Z0.S       // 0080ac05
    ZSPLICE P3, Z10.S, Z12.S, Z10.S    // 8a8dac05
    ZSPLICE P7, Z31.S, Z31.S, Z31.S    // ff9fac05
    ZSPLICE P0, Z0.D, Z0.D, Z0.D       // 0080ec05
    ZSPLICE P3, Z10.D, Z12.D, Z10.D    // 8a8dec05
    ZSPLICE P7, Z31.D, Z31.D, Z31.D    // ff9fec05

// SQADD   <Zdn>.<T>, <Zdn>.<T>, #<imm>, <shift>
    ZSQADD Z0.B, $0, $0, Z0.B        // 00c02425
    ZSQADD Z10.B, $85, $0, Z10.B     // aaca2425
    ZSQADD Z31.B, $255, $0, Z31.B    // ffdf2425
    ZSQADD Z0.H, $0, $8, Z0.H        // 00e06425
    ZSQADD Z10.H, $85, $8, Z10.H     // aaea6425
    ZSQADD Z31.H, $255, $0, Z31.H    // ffdf6425
    ZSQADD Z0.S, $0, $8, Z0.S        // 00e0a425
    ZSQADD Z10.S, $85, $8, Z10.S     // aaeaa425
    ZSQADD Z31.S, $255, $0, Z31.S    // ffdfa425
    ZSQADD Z0.D, $0, $8, Z0.D        // 00e0e425
    ZSQADD Z10.D, $85, $8, Z10.D     // aaeae425
    ZSQADD Z31.D, $255, $0, Z31.D    // ffdfe425

// SQADD   <Zd>.<T>, <Zn>.<T>, <Zm>.<T>
    ZSQADD Z0.B, Z0.B, Z0.B       // 00102004
    ZSQADD Z11.B, Z12.B, Z10.B    // 6a112c04
    ZSQADD Z31.B, Z31.B, Z31.B    // ff133f04
    ZSQADD Z0.H, Z0.H, Z0.H       // 00106004
    ZSQADD Z11.H, Z12.H, Z10.H    // 6a116c04
    ZSQADD Z31.H, Z31.H, Z31.H    // ff137f04
    ZSQADD Z0.S, Z0.S, Z0.S       // 0010a004
    ZSQADD Z11.S, Z12.S, Z10.S    // 6a11ac04
    ZSQADD Z31.S, Z31.S, Z31.S    // ff13bf04
    ZSQADD Z0.D, Z0.D, Z0.D       // 0010e004
    ZSQADD Z11.D, Z12.D, Z10.D    // 6a11ec04
    ZSQADD Z31.D, Z31.D, Z31.D    // ff13ff04

// SQDECB  <Xdn>, <Wdn>{, <pattern>{, MUL #<imm>}}
    ZSQDECBW R0, POW2, $1, R0       // 00f82004
    ZSQDECBW R1, VL1, $1, R1        // 21f82004
    ZSQDECBW R2, VL2, $2, R2        // 42f82104
    ZSQDECBW R3, VL3, $2, R3        // 63f82104
    ZSQDECBW R4, VL4, $3, R4        // 84f82204
    ZSQDECBW R5, VL5, $3, R5        // a5f82204
    ZSQDECBW R6, VL6, $4, R6        // c6f82304
    ZSQDECBW R7, VL7, $4, R7        // e7f82304
    ZSQDECBW R8, VL8, $5, R8        // 08f92404
    ZSQDECBW R8, VL16, $5, R8       // 28f92404
    ZSQDECBW R9, VL32, $6, R9       // 49f92504
    ZSQDECBW R10, VL64, $6, R10     // 6af92504
    ZSQDECBW R11, VL128, $7, R11    // 8bf92604
    ZSQDECBW R12, VL256, $7, R12    // acf92604
    ZSQDECBW R13, $14, $8, R13      // cdf92704
    ZSQDECBW R14, $15, $8, R14      // eef92704
    ZSQDECBW R15, $16, $9, R15      // 0ffa2804
    ZSQDECBW R16, $17, $9, R16      // 30fa2804
    ZSQDECBW R17, $18, $9, R17      // 51fa2804
    ZSQDECBW R17, $19, $10, R17     // 71fa2904
    ZSQDECBW R19, $20, $10, R19     // 93fa2904
    ZSQDECBW R20, $21, $11, R20     // b4fa2a04
    ZSQDECBW R21, $22, $11, R21     // d5fa2a04
    ZSQDECBW R22, $23, $12, R22     // f6fa2b04
    ZSQDECBW R22, $24, $12, R22     // 16fb2b04
    ZSQDECBW R23, $25, $13, R23     // 37fb2c04
    ZSQDECBW R24, $26, $13, R24     // 58fb2c04
    ZSQDECBW R25, $27, $14, R25     // 79fb2d04
    ZSQDECBW R26, $28, $14, R26     // 9afb2d04
    ZSQDECBW R27, MUL4, $15, R27    // bbfb2e04
    ZSQDECBW R27, MUL3, $15, R27    // dbfb2e04
    ZSQDECBW R30, ALL, $16, R30     // fefb2f04

// SQDECB  <Xdn>{, <pattern>{, MUL #<imm>}}
    ZSQDECB POW2, $1, R0      // 00f83004
    ZSQDECB VL1, $1, R1       // 21f83004
    ZSQDECB VL2, $2, R2       // 42f83104
    ZSQDECB VL3, $2, R3       // 63f83104
    ZSQDECB VL4, $3, R4       // 84f83204
    ZSQDECB VL5, $3, R5       // a5f83204
    ZSQDECB VL6, $4, R6       // c6f83304
    ZSQDECB VL7, $4, R7       // e7f83304
    ZSQDECB VL8, $5, R8       // 08f93404
    ZSQDECB VL16, $5, R8      // 28f93404
    ZSQDECB VL32, $6, R9      // 49f93504
    ZSQDECB VL64, $6, R10     // 6af93504
    ZSQDECB VL128, $7, R11    // 8bf93604
    ZSQDECB VL256, $7, R12    // acf93604
    ZSQDECB $14, $8, R13      // cdf93704
    ZSQDECB $15, $8, R14      // eef93704
    ZSQDECB $16, $9, R15      // 0ffa3804
    ZSQDECB $17, $9, R16      // 30fa3804
    ZSQDECB $18, $9, R17      // 51fa3804
    ZSQDECB $19, $10, R17     // 71fa3904
    ZSQDECB $20, $10, R19     // 93fa3904
    ZSQDECB $21, $11, R20     // b4fa3a04
    ZSQDECB $22, $11, R21     // d5fa3a04
    ZSQDECB $23, $12, R22     // f6fa3b04
    ZSQDECB $24, $12, R22     // 16fb3b04
    ZSQDECB $25, $13, R23     // 37fb3c04
    ZSQDECB $26, $13, R24     // 58fb3c04
    ZSQDECB $27, $14, R25     // 79fb3d04
    ZSQDECB $28, $14, R26     // 9afb3d04
    ZSQDECB MUL4, $15, R27    // bbfb3e04
    ZSQDECB MUL3, $15, R27    // dbfb3e04
    ZSQDECB ALL, $16, R30     // fefb3f04

// SQDECD  <Xdn>, <Wdn>{, <pattern>{, MUL #<imm>}}
    ZSQDECDW R0, POW2, $1, R0       // 00f8e004
    ZSQDECDW R1, VL1, $1, R1        // 21f8e004
    ZSQDECDW R2, VL2, $2, R2        // 42f8e104
    ZSQDECDW R3, VL3, $2, R3        // 63f8e104
    ZSQDECDW R4, VL4, $3, R4        // 84f8e204
    ZSQDECDW R5, VL5, $3, R5        // a5f8e204
    ZSQDECDW R6, VL6, $4, R6        // c6f8e304
    ZSQDECDW R7, VL7, $4, R7        // e7f8e304
    ZSQDECDW R8, VL8, $5, R8        // 08f9e404
    ZSQDECDW R8, VL16, $5, R8       // 28f9e404
    ZSQDECDW R9, VL32, $6, R9       // 49f9e504
    ZSQDECDW R10, VL64, $6, R10     // 6af9e504
    ZSQDECDW R11, VL128, $7, R11    // 8bf9e604
    ZSQDECDW R12, VL256, $7, R12    // acf9e604
    ZSQDECDW R13, $14, $8, R13      // cdf9e704
    ZSQDECDW R14, $15, $8, R14      // eef9e704
    ZSQDECDW R15, $16, $9, R15      // 0ffae804
    ZSQDECDW R16, $17, $9, R16      // 30fae804
    ZSQDECDW R17, $18, $9, R17      // 51fae804
    ZSQDECDW R17, $19, $10, R17     // 71fae904
    ZSQDECDW R19, $20, $10, R19     // 93fae904
    ZSQDECDW R20, $21, $11, R20     // b4faea04
    ZSQDECDW R21, $22, $11, R21     // d5faea04
    ZSQDECDW R22, $23, $12, R22     // f6faeb04
    ZSQDECDW R22, $24, $12, R22     // 16fbeb04
    ZSQDECDW R23, $25, $13, R23     // 37fbec04
    ZSQDECDW R24, $26, $13, R24     // 58fbec04
    ZSQDECDW R25, $27, $14, R25     // 79fbed04
    ZSQDECDW R26, $28, $14, R26     // 9afbed04
    ZSQDECDW R27, MUL4, $15, R27    // bbfbee04
    ZSQDECDW R27, MUL3, $15, R27    // dbfbee04
    ZSQDECDW R30, ALL, $16, R30     // fefbef04

// SQDECD  <Xdn>{, <pattern>{, MUL #<imm>}}
    ZSQDECD POW2, $1, R0      // 00f8f004
    ZSQDECD VL1, $1, R1       // 21f8f004
    ZSQDECD VL2, $2, R2       // 42f8f104
    ZSQDECD VL3, $2, R3       // 63f8f104
    ZSQDECD VL4, $3, R4       // 84f8f204
    ZSQDECD VL5, $3, R5       // a5f8f204
    ZSQDECD VL6, $4, R6       // c6f8f304
    ZSQDECD VL7, $4, R7       // e7f8f304
    ZSQDECD VL8, $5, R8       // 08f9f404
    ZSQDECD VL16, $5, R8      // 28f9f404
    ZSQDECD VL32, $6, R9      // 49f9f504
    ZSQDECD VL64, $6, R10     // 6af9f504
    ZSQDECD VL128, $7, R11    // 8bf9f604
    ZSQDECD VL256, $7, R12    // acf9f604
    ZSQDECD $14, $8, R13      // cdf9f704
    ZSQDECD $15, $8, R14      // eef9f704
    ZSQDECD $16, $9, R15      // 0ffaf804
    ZSQDECD $17, $9, R16      // 30faf804
    ZSQDECD $18, $9, R17      // 51faf804
    ZSQDECD $19, $10, R17     // 71faf904
    ZSQDECD $20, $10, R19     // 93faf904
    ZSQDECD $21, $11, R20     // b4fafa04
    ZSQDECD $22, $11, R21     // d5fafa04
    ZSQDECD $23, $12, R22     // f6fafb04
    ZSQDECD $24, $12, R22     // 16fbfb04
    ZSQDECD $25, $13, R23     // 37fbfc04
    ZSQDECD $26, $13, R24     // 58fbfc04
    ZSQDECD $27, $14, R25     // 79fbfd04
    ZSQDECD $28, $14, R26     // 9afbfd04
    ZSQDECD MUL4, $15, R27    // bbfbfe04
    ZSQDECD MUL3, $15, R27    // dbfbfe04
    ZSQDECD ALL, $16, R30     // fefbff04

// SQDECD  <Zdn>.D{, <pattern>{, MUL #<imm>}}
    ZSQDECD POW2, $1, Z0.D      // 00c8e004
    ZSQDECD VL1, $1, Z1.D       // 21c8e004
    ZSQDECD VL2, $2, Z2.D       // 42c8e104
    ZSQDECD VL3, $2, Z3.D       // 63c8e104
    ZSQDECD VL4, $3, Z4.D       // 84c8e204
    ZSQDECD VL5, $3, Z5.D       // a5c8e204
    ZSQDECD VL6, $4, Z6.D       // c6c8e304
    ZSQDECD VL7, $4, Z7.D       // e7c8e304
    ZSQDECD VL8, $5, Z8.D       // 08c9e404
    ZSQDECD VL16, $5, Z9.D      // 29c9e404
    ZSQDECD VL32, $6, Z10.D     // 4ac9e504
    ZSQDECD VL64, $6, Z11.D     // 6bc9e504
    ZSQDECD VL128, $7, Z12.D    // 8cc9e604
    ZSQDECD VL256, $7, Z13.D    // adc9e604
    ZSQDECD $14, $8, Z14.D      // cec9e704
    ZSQDECD $15, $8, Z15.D      // efc9e704
    ZSQDECD $16, $9, Z16.D      // 10cae804
    ZSQDECD $17, $9, Z16.D      // 30cae804
    ZSQDECD $18, $9, Z17.D      // 51cae804
    ZSQDECD $19, $10, Z18.D     // 72cae904
    ZSQDECD $20, $10, Z19.D     // 93cae904
    ZSQDECD $21, $11, Z20.D     // b4caea04
    ZSQDECD $22, $11, Z21.D     // d5caea04
    ZSQDECD $23, $12, Z22.D     // f6caeb04
    ZSQDECD $24, $12, Z23.D     // 17cbeb04
    ZSQDECD $25, $13, Z24.D     // 38cbec04
    ZSQDECD $26, $13, Z25.D     // 59cbec04
    ZSQDECD $27, $14, Z26.D     // 7acbed04
    ZSQDECD $28, $14, Z27.D     // 9bcbed04
    ZSQDECD MUL4, $15, Z28.D    // bccbee04
    ZSQDECD MUL3, $15, Z29.D    // ddcbee04
    ZSQDECD ALL, $16, Z31.D     // ffcbef04

// SQDECH  <Xdn>, <Wdn>{, <pattern>{, MUL #<imm>}}
    ZSQDECHW R0, POW2, $1, R0       // 00f86004
    ZSQDECHW R1, VL1, $1, R1        // 21f86004
    ZSQDECHW R2, VL2, $2, R2        // 42f86104
    ZSQDECHW R3, VL3, $2, R3        // 63f86104
    ZSQDECHW R4, VL4, $3, R4        // 84f86204
    ZSQDECHW R5, VL5, $3, R5        // a5f86204
    ZSQDECHW R6, VL6, $4, R6        // c6f86304
    ZSQDECHW R7, VL7, $4, R7        // e7f86304
    ZSQDECHW R8, VL8, $5, R8        // 08f96404
    ZSQDECHW R8, VL16, $5, R8       // 28f96404
    ZSQDECHW R9, VL32, $6, R9       // 49f96504
    ZSQDECHW R10, VL64, $6, R10     // 6af96504
    ZSQDECHW R11, VL128, $7, R11    // 8bf96604
    ZSQDECHW R12, VL256, $7, R12    // acf96604
    ZSQDECHW R13, $14, $8, R13      // cdf96704
    ZSQDECHW R14, $15, $8, R14      // eef96704
    ZSQDECHW R15, $16, $9, R15      // 0ffa6804
    ZSQDECHW R16, $17, $9, R16      // 30fa6804
    ZSQDECHW R17, $18, $9, R17      // 51fa6804
    ZSQDECHW R17, $19, $10, R17     // 71fa6904
    ZSQDECHW R19, $20, $10, R19     // 93fa6904
    ZSQDECHW R20, $21, $11, R20     // b4fa6a04
    ZSQDECHW R21, $22, $11, R21     // d5fa6a04
    ZSQDECHW R22, $23, $12, R22     // f6fa6b04
    ZSQDECHW R22, $24, $12, R22     // 16fb6b04
    ZSQDECHW R23, $25, $13, R23     // 37fb6c04
    ZSQDECHW R24, $26, $13, R24     // 58fb6c04
    ZSQDECHW R25, $27, $14, R25     // 79fb6d04
    ZSQDECHW R26, $28, $14, R26     // 9afb6d04
    ZSQDECHW R27, MUL4, $15, R27    // bbfb6e04
    ZSQDECHW R27, MUL3, $15, R27    // dbfb6e04
    ZSQDECHW R30, ALL, $16, R30     // fefb6f04

// SQDECH  <Xdn>{, <pattern>{, MUL #<imm>}}
    ZSQDECH POW2, $1, R0      // 00f87004
    ZSQDECH VL1, $1, R1       // 21f87004
    ZSQDECH VL2, $2, R2       // 42f87104
    ZSQDECH VL3, $2, R3       // 63f87104
    ZSQDECH VL4, $3, R4       // 84f87204
    ZSQDECH VL5, $3, R5       // a5f87204
    ZSQDECH VL6, $4, R6       // c6f87304
    ZSQDECH VL7, $4, R7       // e7f87304
    ZSQDECH VL8, $5, R8       // 08f97404
    ZSQDECH VL16, $5, R8      // 28f97404
    ZSQDECH VL32, $6, R9      // 49f97504
    ZSQDECH VL64, $6, R10     // 6af97504
    ZSQDECH VL128, $7, R11    // 8bf97604
    ZSQDECH VL256, $7, R12    // acf97604
    ZSQDECH $14, $8, R13      // cdf97704
    ZSQDECH $15, $8, R14      // eef97704
    ZSQDECH $16, $9, R15      // 0ffa7804
    ZSQDECH $17, $9, R16      // 30fa7804
    ZSQDECH $18, $9, R17      // 51fa7804
    ZSQDECH $19, $10, R17     // 71fa7904
    ZSQDECH $20, $10, R19     // 93fa7904
    ZSQDECH $21, $11, R20     // b4fa7a04
    ZSQDECH $22, $11, R21     // d5fa7a04
    ZSQDECH $23, $12, R22     // f6fa7b04
    ZSQDECH $24, $12, R22     // 16fb7b04
    ZSQDECH $25, $13, R23     // 37fb7c04
    ZSQDECH $26, $13, R24     // 58fb7c04
    ZSQDECH $27, $14, R25     // 79fb7d04
    ZSQDECH $28, $14, R26     // 9afb7d04
    ZSQDECH MUL4, $15, R27    // bbfb7e04
    ZSQDECH MUL3, $15, R27    // dbfb7e04
    ZSQDECH ALL, $16, R30     // fefb7f04

// SQDECH  <Zdn>.H{, <pattern>{, MUL #<imm>}}
    ZSQDECH POW2, $1, Z0.H      // 00c86004
    ZSQDECH VL1, $1, Z1.H       // 21c86004
    ZSQDECH VL2, $2, Z2.H       // 42c86104
    ZSQDECH VL3, $2, Z3.H       // 63c86104
    ZSQDECH VL4, $3, Z4.H       // 84c86204
    ZSQDECH VL5, $3, Z5.H       // a5c86204
    ZSQDECH VL6, $4, Z6.H       // c6c86304
    ZSQDECH VL7, $4, Z7.H       // e7c86304
    ZSQDECH VL8, $5, Z8.H       // 08c96404
    ZSQDECH VL16, $5, Z9.H      // 29c96404
    ZSQDECH VL32, $6, Z10.H     // 4ac96504
    ZSQDECH VL64, $6, Z11.H     // 6bc96504
    ZSQDECH VL128, $7, Z12.H    // 8cc96604
    ZSQDECH VL256, $7, Z13.H    // adc96604
    ZSQDECH $14, $8, Z14.H      // cec96704
    ZSQDECH $15, $8, Z15.H      // efc96704
    ZSQDECH $16, $9, Z16.H      // 10ca6804
    ZSQDECH $17, $9, Z16.H      // 30ca6804
    ZSQDECH $18, $9, Z17.H      // 51ca6804
    ZSQDECH $19, $10, Z18.H     // 72ca6904
    ZSQDECH $20, $10, Z19.H     // 93ca6904
    ZSQDECH $21, $11, Z20.H     // b4ca6a04
    ZSQDECH $22, $11, Z21.H     // d5ca6a04
    ZSQDECH $23, $12, Z22.H     // f6ca6b04
    ZSQDECH $24, $12, Z23.H     // 17cb6b04
    ZSQDECH $25, $13, Z24.H     // 38cb6c04
    ZSQDECH $26, $13, Z25.H     // 59cb6c04
    ZSQDECH $27, $14, Z26.H     // 7acb6d04
    ZSQDECH $28, $14, Z27.H     // 9bcb6d04
    ZSQDECH MUL4, $15, Z28.H    // bccb6e04
    ZSQDECH MUL3, $15, Z29.H    // ddcb6e04
    ZSQDECH ALL, $16, Z31.H     // ffcb6f04

// SQDECP  <Xdn>, <Pm>.<T>, <Wdn>
    PSQDECPW P0.B, R0, R0       // 00882a25
    PSQDECPW P6.B, R10, R10     // ca882a25
    PSQDECPW P15.B, R30, R30    // fe892a25
    PSQDECPW P0.H, R0, R0       // 00886a25
    PSQDECPW P6.H, R10, R10     // ca886a25
    PSQDECPW P15.H, R30, R30    // fe896a25
    PSQDECPW P0.S, R0, R0       // 0088aa25
    PSQDECPW P6.S, R10, R10     // ca88aa25
    PSQDECPW P15.S, R30, R30    // fe89aa25
    PSQDECPW P0.D, R0, R0       // 0088ea25
    PSQDECPW P6.D, R10, R10     // ca88ea25
    PSQDECPW P15.D, R30, R30    // fe89ea25

// SQDECP  <Xdn>, <Pm>.<T>
    PSQDECP P0.B, R0      // 008c2a25
    PSQDECP P6.B, R10     // ca8c2a25
    PSQDECP P15.B, R30    // fe8d2a25
    PSQDECP P0.H, R0      // 008c6a25
    PSQDECP P6.H, R10     // ca8c6a25
    PSQDECP P15.H, R30    // fe8d6a25
    PSQDECP P0.S, R0      // 008caa25
    PSQDECP P6.S, R10     // ca8caa25
    PSQDECP P15.S, R30    // fe8daa25
    PSQDECP P0.D, R0      // 008cea25
    PSQDECP P6.D, R10     // ca8cea25
    PSQDECP P15.D, R30    // fe8dea25

// SQDECP  <Zdn>.<T>, <Pm>.<T>
    ZSQDECP P0.H, Z0.H      // 00806a25
    ZSQDECP P6.H, Z10.H     // ca806a25
    ZSQDECP P15.H, Z31.H    // ff816a25
    ZSQDECP P0.S, Z0.S      // 0080aa25
    ZSQDECP P6.S, Z10.S     // ca80aa25
    ZSQDECP P15.S, Z31.S    // ff81aa25
    ZSQDECP P0.D, Z0.D      // 0080ea25
    ZSQDECP P6.D, Z10.D     // ca80ea25
    ZSQDECP P15.D, Z31.D    // ff81ea25

// SQDECW  <Xdn>, <Wdn>{, <pattern>{, MUL #<imm>}}
    ZSQDECWW R0, POW2, $1, R0       // 00f8a004
    ZSQDECWW R1, VL1, $1, R1        // 21f8a004
    ZSQDECWW R2, VL2, $2, R2        // 42f8a104
    ZSQDECWW R3, VL3, $2, R3        // 63f8a104
    ZSQDECWW R4, VL4, $3, R4        // 84f8a204
    ZSQDECWW R5, VL5, $3, R5        // a5f8a204
    ZSQDECWW R6, VL6, $4, R6        // c6f8a304
    ZSQDECWW R7, VL7, $4, R7        // e7f8a304
    ZSQDECWW R8, VL8, $5, R8        // 08f9a404
    ZSQDECWW R8, VL16, $5, R8       // 28f9a404
    ZSQDECWW R9, VL32, $6, R9       // 49f9a504
    ZSQDECWW R10, VL64, $6, R10     // 6af9a504
    ZSQDECWW R11, VL128, $7, R11    // 8bf9a604
    ZSQDECWW R12, VL256, $7, R12    // acf9a604
    ZSQDECWW R13, $14, $8, R13      // cdf9a704
    ZSQDECWW R14, $15, $8, R14      // eef9a704
    ZSQDECWW R15, $16, $9, R15      // 0ffaa804
    ZSQDECWW R16, $17, $9, R16      // 30faa804
    ZSQDECWW R17, $18, $9, R17      // 51faa804
    ZSQDECWW R17, $19, $10, R17     // 71faa904
    ZSQDECWW R19, $20, $10, R19     // 93faa904
    ZSQDECWW R20, $21, $11, R20     // b4faaa04
    ZSQDECWW R21, $22, $11, R21     // d5faaa04
    ZSQDECWW R22, $23, $12, R22     // f6faab04
    ZSQDECWW R22, $24, $12, R22     // 16fbab04
    ZSQDECWW R23, $25, $13, R23     // 37fbac04
    ZSQDECWW R24, $26, $13, R24     // 58fbac04
    ZSQDECWW R25, $27, $14, R25     // 79fbad04
    ZSQDECWW R26, $28, $14, R26     // 9afbad04
    ZSQDECWW R27, MUL4, $15, R27    // bbfbae04
    ZSQDECWW R27, MUL3, $15, R27    // dbfbae04
    ZSQDECWW R30, ALL, $16, R30     // fefbaf04

// SQDECW  <Xdn>{, <pattern>{, MUL #<imm>}}
    ZSQDECW POW2, $1, R0      // 00f8b004
    ZSQDECW VL1, $1, R1       // 21f8b004
    ZSQDECW VL2, $2, R2       // 42f8b104
    ZSQDECW VL3, $2, R3       // 63f8b104
    ZSQDECW VL4, $3, R4       // 84f8b204
    ZSQDECW VL5, $3, R5       // a5f8b204
    ZSQDECW VL6, $4, R6       // c6f8b304
    ZSQDECW VL7, $4, R7       // e7f8b304
    ZSQDECW VL8, $5, R8       // 08f9b404
    ZSQDECW VL16, $5, R8      // 28f9b404
    ZSQDECW VL32, $6, R9      // 49f9b504
    ZSQDECW VL64, $6, R10     // 6af9b504
    ZSQDECW VL128, $7, R11    // 8bf9b604
    ZSQDECW VL256, $7, R12    // acf9b604
    ZSQDECW $14, $8, R13      // cdf9b704
    ZSQDECW $15, $8, R14      // eef9b704
    ZSQDECW $16, $9, R15      // 0ffab804
    ZSQDECW $17, $9, R16      // 30fab804
    ZSQDECW $18, $9, R17      // 51fab804
    ZSQDECW $19, $10, R17     // 71fab904
    ZSQDECW $20, $10, R19     // 93fab904
    ZSQDECW $21, $11, R20     // b4faba04
    ZSQDECW $22, $11, R21     // d5faba04
    ZSQDECW $23, $12, R22     // f6fabb04
    ZSQDECW $24, $12, R22     // 16fbbb04
    ZSQDECW $25, $13, R23     // 37fbbc04
    ZSQDECW $26, $13, R24     // 58fbbc04
    ZSQDECW $27, $14, R25     // 79fbbd04
    ZSQDECW $28, $14, R26     // 9afbbd04
    ZSQDECW MUL4, $15, R27    // bbfbbe04
    ZSQDECW MUL3, $15, R27    // dbfbbe04
    ZSQDECW ALL, $16, R30     // fefbbf04

// SQDECW  <Zdn>.S{, <pattern>{, MUL #<imm>}}
    ZSQDECW POW2, $1, Z0.S      // 00c8a004
    ZSQDECW VL1, $1, Z1.S       // 21c8a004
    ZSQDECW VL2, $2, Z2.S       // 42c8a104
    ZSQDECW VL3, $2, Z3.S       // 63c8a104
    ZSQDECW VL4, $3, Z4.S       // 84c8a204
    ZSQDECW VL5, $3, Z5.S       // a5c8a204
    ZSQDECW VL6, $4, Z6.S       // c6c8a304
    ZSQDECW VL7, $4, Z7.S       // e7c8a304
    ZSQDECW VL8, $5, Z8.S       // 08c9a404
    ZSQDECW VL16, $5, Z9.S      // 29c9a404
    ZSQDECW VL32, $6, Z10.S     // 4ac9a504
    ZSQDECW VL64, $6, Z11.S     // 6bc9a504
    ZSQDECW VL128, $7, Z12.S    // 8cc9a604
    ZSQDECW VL256, $7, Z13.S    // adc9a604
    ZSQDECW $14, $8, Z14.S      // cec9a704
    ZSQDECW $15, $8, Z15.S      // efc9a704
    ZSQDECW $16, $9, Z16.S      // 10caa804
    ZSQDECW $17, $9, Z16.S      // 30caa804
    ZSQDECW $18, $9, Z17.S      // 51caa804
    ZSQDECW $19, $10, Z18.S     // 72caa904
    ZSQDECW $20, $10, Z19.S     // 93caa904
    ZSQDECW $21, $11, Z20.S     // b4caaa04
    ZSQDECW $22, $11, Z21.S     // d5caaa04
    ZSQDECW $23, $12, Z22.S     // f6caab04
    ZSQDECW $24, $12, Z23.S     // 17cbab04
    ZSQDECW $25, $13, Z24.S     // 38cbac04
    ZSQDECW $26, $13, Z25.S     // 59cbac04
    ZSQDECW $27, $14, Z26.S     // 7acbad04
    ZSQDECW $28, $14, Z27.S     // 9bcbad04
    ZSQDECW MUL4, $15, Z28.S    // bccbae04
    ZSQDECW MUL3, $15, Z29.S    // ddcbae04
    ZSQDECW ALL, $16, Z31.S     // ffcbaf04

// SQINCB  <Xdn>, <Wdn>{, <pattern>{, MUL #<imm>}}
    ZSQINCBW R0, POW2, $1, R0       // 00f02004
    ZSQINCBW R1, VL1, $1, R1        // 21f02004
    ZSQINCBW R2, VL2, $2, R2        // 42f02104
    ZSQINCBW R3, VL3, $2, R3        // 63f02104
    ZSQINCBW R4, VL4, $3, R4        // 84f02204
    ZSQINCBW R5, VL5, $3, R5        // a5f02204
    ZSQINCBW R6, VL6, $4, R6        // c6f02304
    ZSQINCBW R7, VL7, $4, R7        // e7f02304
    ZSQINCBW R8, VL8, $5, R8        // 08f12404
    ZSQINCBW R8, VL16, $5, R8       // 28f12404
    ZSQINCBW R9, VL32, $6, R9       // 49f12504
    ZSQINCBW R10, VL64, $6, R10     // 6af12504
    ZSQINCBW R11, VL128, $7, R11    // 8bf12604
    ZSQINCBW R12, VL256, $7, R12    // acf12604
    ZSQINCBW R13, $14, $8, R13      // cdf12704
    ZSQINCBW R14, $15, $8, R14      // eef12704
    ZSQINCBW R15, $16, $9, R15      // 0ff22804
    ZSQINCBW R16, $17, $9, R16      // 30f22804
    ZSQINCBW R17, $18, $9, R17      // 51f22804
    ZSQINCBW R17, $19, $10, R17     // 71f22904
    ZSQINCBW R19, $20, $10, R19     // 93f22904
    ZSQINCBW R20, $21, $11, R20     // b4f22a04
    ZSQINCBW R21, $22, $11, R21     // d5f22a04
    ZSQINCBW R22, $23, $12, R22     // f6f22b04
    ZSQINCBW R22, $24, $12, R22     // 16f32b04
    ZSQINCBW R23, $25, $13, R23     // 37f32c04
    ZSQINCBW R24, $26, $13, R24     // 58f32c04
    ZSQINCBW R25, $27, $14, R25     // 79f32d04
    ZSQINCBW R26, $28, $14, R26     // 9af32d04
    ZSQINCBW R27, MUL4, $15, R27    // bbf32e04
    ZSQINCBW R27, MUL3, $15, R27    // dbf32e04
    ZSQINCBW R30, ALL, $16, R30     // fef32f04

// SQINCB  <Xdn>{, <pattern>{, MUL #<imm>}}
    ZSQINCB POW2, $1, R0      // 00f03004
    ZSQINCB VL1, $1, R1       // 21f03004
    ZSQINCB VL2, $2, R2       // 42f03104
    ZSQINCB VL3, $2, R3       // 63f03104
    ZSQINCB VL4, $3, R4       // 84f03204
    ZSQINCB VL5, $3, R5       // a5f03204
    ZSQINCB VL6, $4, R6       // c6f03304
    ZSQINCB VL7, $4, R7       // e7f03304
    ZSQINCB VL8, $5, R8       // 08f13404
    ZSQINCB VL16, $5, R8      // 28f13404
    ZSQINCB VL32, $6, R9      // 49f13504
    ZSQINCB VL64, $6, R10     // 6af13504
    ZSQINCB VL128, $7, R11    // 8bf13604
    ZSQINCB VL256, $7, R12    // acf13604
    ZSQINCB $14, $8, R13      // cdf13704
    ZSQINCB $15, $8, R14      // eef13704
    ZSQINCB $16, $9, R15      // 0ff23804
    ZSQINCB $17, $9, R16      // 30f23804
    ZSQINCB $18, $9, R17      // 51f23804
    ZSQINCB $19, $10, R17     // 71f23904
    ZSQINCB $20, $10, R19     // 93f23904
    ZSQINCB $21, $11, R20     // b4f23a04
    ZSQINCB $22, $11, R21     // d5f23a04
    ZSQINCB $23, $12, R22     // f6f23b04
    ZSQINCB $24, $12, R22     // 16f33b04
    ZSQINCB $25, $13, R23     // 37f33c04
    ZSQINCB $26, $13, R24     // 58f33c04
    ZSQINCB $27, $14, R25     // 79f33d04
    ZSQINCB $28, $14, R26     // 9af33d04
    ZSQINCB MUL4, $15, R27    // bbf33e04
    ZSQINCB MUL3, $15, R27    // dbf33e04
    ZSQINCB ALL, $16, R30     // fef33f04

// SQINCD  <Xdn>, <Wdn>{, <pattern>{, MUL #<imm>}}
    ZSQINCDW R0, POW2, $1, R0       // 00f0e004
    ZSQINCDW R1, VL1, $1, R1        // 21f0e004
    ZSQINCDW R2, VL2, $2, R2        // 42f0e104
    ZSQINCDW R3, VL3, $2, R3        // 63f0e104
    ZSQINCDW R4, VL4, $3, R4        // 84f0e204
    ZSQINCDW R5, VL5, $3, R5        // a5f0e204
    ZSQINCDW R6, VL6, $4, R6        // c6f0e304
    ZSQINCDW R7, VL7, $4, R7        // e7f0e304
    ZSQINCDW R8, VL8, $5, R8        // 08f1e404
    ZSQINCDW R8, VL16, $5, R8       // 28f1e404
    ZSQINCDW R9, VL32, $6, R9       // 49f1e504
    ZSQINCDW R10, VL64, $6, R10     // 6af1e504
    ZSQINCDW R11, VL128, $7, R11    // 8bf1e604
    ZSQINCDW R12, VL256, $7, R12    // acf1e604
    ZSQINCDW R13, $14, $8, R13      // cdf1e704
    ZSQINCDW R14, $15, $8, R14      // eef1e704
    ZSQINCDW R15, $16, $9, R15      // 0ff2e804
    ZSQINCDW R16, $17, $9, R16      // 30f2e804
    ZSQINCDW R17, $18, $9, R17      // 51f2e804
    ZSQINCDW R17, $19, $10, R17     // 71f2e904
    ZSQINCDW R19, $20, $10, R19     // 93f2e904
    ZSQINCDW R20, $21, $11, R20     // b4f2ea04
    ZSQINCDW R21, $22, $11, R21     // d5f2ea04
    ZSQINCDW R22, $23, $12, R22     // f6f2eb04
    ZSQINCDW R22, $24, $12, R22     // 16f3eb04
    ZSQINCDW R23, $25, $13, R23     // 37f3ec04
    ZSQINCDW R24, $26, $13, R24     // 58f3ec04
    ZSQINCDW R25, $27, $14, R25     // 79f3ed04
    ZSQINCDW R26, $28, $14, R26     // 9af3ed04
    ZSQINCDW R27, MUL4, $15, R27    // bbf3ee04
    ZSQINCDW R27, MUL3, $15, R27    // dbf3ee04
    ZSQINCDW R30, ALL, $16, R30     // fef3ef04

// SQINCD  <Xdn>{, <pattern>{, MUL #<imm>}}
    ZSQINCD POW2, $1, R0      // 00f0f004
    ZSQINCD VL1, $1, R1       // 21f0f004
    ZSQINCD VL2, $2, R2       // 42f0f104
    ZSQINCD VL3, $2, R3       // 63f0f104
    ZSQINCD VL4, $3, R4       // 84f0f204
    ZSQINCD VL5, $3, R5       // a5f0f204
    ZSQINCD VL6, $4, R6       // c6f0f304
    ZSQINCD VL7, $4, R7       // e7f0f304
    ZSQINCD VL8, $5, R8       // 08f1f404
    ZSQINCD VL16, $5, R8      // 28f1f404
    ZSQINCD VL32, $6, R9      // 49f1f504
    ZSQINCD VL64, $6, R10     // 6af1f504
    ZSQINCD VL128, $7, R11    // 8bf1f604
    ZSQINCD VL256, $7, R12    // acf1f604
    ZSQINCD $14, $8, R13      // cdf1f704
    ZSQINCD $15, $8, R14      // eef1f704
    ZSQINCD $16, $9, R15      // 0ff2f804
    ZSQINCD $17, $9, R16      // 30f2f804
    ZSQINCD $18, $9, R17      // 51f2f804
    ZSQINCD $19, $10, R17     // 71f2f904
    ZSQINCD $20, $10, R19     // 93f2f904
    ZSQINCD $21, $11, R20     // b4f2fa04
    ZSQINCD $22, $11, R21     // d5f2fa04
    ZSQINCD $23, $12, R22     // f6f2fb04
    ZSQINCD $24, $12, R22     // 16f3fb04
    ZSQINCD $25, $13, R23     // 37f3fc04
    ZSQINCD $26, $13, R24     // 58f3fc04
    ZSQINCD $27, $14, R25     // 79f3fd04
    ZSQINCD $28, $14, R26     // 9af3fd04
    ZSQINCD MUL4, $15, R27    // bbf3fe04
    ZSQINCD MUL3, $15, R27    // dbf3fe04
    ZSQINCD ALL, $16, R30     // fef3ff04

// SQINCD  <Zdn>.D{, <pattern>{, MUL #<imm>}}
    ZSQINCD POW2, $1, Z0.D      // 00c0e004
    ZSQINCD VL1, $1, Z1.D       // 21c0e004
    ZSQINCD VL2, $2, Z2.D       // 42c0e104
    ZSQINCD VL3, $2, Z3.D       // 63c0e104
    ZSQINCD VL4, $3, Z4.D       // 84c0e204
    ZSQINCD VL5, $3, Z5.D       // a5c0e204
    ZSQINCD VL6, $4, Z6.D       // c6c0e304
    ZSQINCD VL7, $4, Z7.D       // e7c0e304
    ZSQINCD VL8, $5, Z8.D       // 08c1e404
    ZSQINCD VL16, $5, Z9.D      // 29c1e404
    ZSQINCD VL32, $6, Z10.D     // 4ac1e504
    ZSQINCD VL64, $6, Z11.D     // 6bc1e504
    ZSQINCD VL128, $7, Z12.D    // 8cc1e604
    ZSQINCD VL256, $7, Z13.D    // adc1e604
    ZSQINCD $14, $8, Z14.D      // cec1e704
    ZSQINCD $15, $8, Z15.D      // efc1e704
    ZSQINCD $16, $9, Z16.D      // 10c2e804
    ZSQINCD $17, $9, Z16.D      // 30c2e804
    ZSQINCD $18, $9, Z17.D      // 51c2e804
    ZSQINCD $19, $10, Z18.D     // 72c2e904
    ZSQINCD $20, $10, Z19.D     // 93c2e904
    ZSQINCD $21, $11, Z20.D     // b4c2ea04
    ZSQINCD $22, $11, Z21.D     // d5c2ea04
    ZSQINCD $23, $12, Z22.D     // f6c2eb04
    ZSQINCD $24, $12, Z23.D     // 17c3eb04
    ZSQINCD $25, $13, Z24.D     // 38c3ec04
    ZSQINCD $26, $13, Z25.D     // 59c3ec04
    ZSQINCD $27, $14, Z26.D     // 7ac3ed04
    ZSQINCD $28, $14, Z27.D     // 9bc3ed04
    ZSQINCD MUL4, $15, Z28.D    // bcc3ee04
    ZSQINCD MUL3, $15, Z29.D    // ddc3ee04
    ZSQINCD ALL, $16, Z31.D     // ffc3ef04

// SQINCH  <Xdn>, <Wdn>{, <pattern>{, MUL #<imm>}}
    ZSQINCHW R0, POW2, $1, R0       // 00f06004
    ZSQINCHW R1, VL1, $1, R1        // 21f06004
    ZSQINCHW R2, VL2, $2, R2        // 42f06104
    ZSQINCHW R3, VL3, $2, R3        // 63f06104
    ZSQINCHW R4, VL4, $3, R4        // 84f06204
    ZSQINCHW R5, VL5, $3, R5        // a5f06204
    ZSQINCHW R6, VL6, $4, R6        // c6f06304
    ZSQINCHW R7, VL7, $4, R7        // e7f06304
    ZSQINCHW R8, VL8, $5, R8        // 08f16404
    ZSQINCHW R8, VL16, $5, R8       // 28f16404
    ZSQINCHW R9, VL32, $6, R9       // 49f16504
    ZSQINCHW R10, VL64, $6, R10     // 6af16504
    ZSQINCHW R11, VL128, $7, R11    // 8bf16604
    ZSQINCHW R12, VL256, $7, R12    // acf16604
    ZSQINCHW R13, $14, $8, R13      // cdf16704
    ZSQINCHW R14, $15, $8, R14      // eef16704
    ZSQINCHW R15, $16, $9, R15      // 0ff26804
    ZSQINCHW R16, $17, $9, R16      // 30f26804
    ZSQINCHW R17, $18, $9, R17      // 51f26804
    ZSQINCHW R17, $19, $10, R17     // 71f26904
    ZSQINCHW R19, $20, $10, R19     // 93f26904
    ZSQINCHW R20, $21, $11, R20     // b4f26a04
    ZSQINCHW R21, $22, $11, R21     // d5f26a04
    ZSQINCHW R22, $23, $12, R22     // f6f26b04
    ZSQINCHW R22, $24, $12, R22     // 16f36b04
    ZSQINCHW R23, $25, $13, R23     // 37f36c04
    ZSQINCHW R24, $26, $13, R24     // 58f36c04
    ZSQINCHW R25, $27, $14, R25     // 79f36d04
    ZSQINCHW R26, $28, $14, R26     // 9af36d04
    ZSQINCHW R27, MUL4, $15, R27    // bbf36e04
    ZSQINCHW R27, MUL3, $15, R27    // dbf36e04
    ZSQINCHW R30, ALL, $16, R30     // fef36f04

// SQINCH  <Xdn>{, <pattern>{, MUL #<imm>}}
    ZSQINCH POW2, $1, R0      // 00f07004
    ZSQINCH VL1, $1, R1       // 21f07004
    ZSQINCH VL2, $2, R2       // 42f07104
    ZSQINCH VL3, $2, R3       // 63f07104
    ZSQINCH VL4, $3, R4       // 84f07204
    ZSQINCH VL5, $3, R5       // a5f07204
    ZSQINCH VL6, $4, R6       // c6f07304
    ZSQINCH VL7, $4, R7       // e7f07304
    ZSQINCH VL8, $5, R8       // 08f17404
    ZSQINCH VL16, $5, R8      // 28f17404
    ZSQINCH VL32, $6, R9      // 49f17504
    ZSQINCH VL64, $6, R10     // 6af17504
    ZSQINCH VL128, $7, R11    // 8bf17604
    ZSQINCH VL256, $7, R12    // acf17604
    ZSQINCH $14, $8, R13      // cdf17704
    ZSQINCH $15, $8, R14      // eef17704
    ZSQINCH $16, $9, R15      // 0ff27804
    ZSQINCH $17, $9, R16      // 30f27804
    ZSQINCH $18, $9, R17      // 51f27804
    ZSQINCH $19, $10, R17     // 71f27904
    ZSQINCH $20, $10, R19     // 93f27904
    ZSQINCH $21, $11, R20     // b4f27a04
    ZSQINCH $22, $11, R21     // d5f27a04
    ZSQINCH $23, $12, R22     // f6f27b04
    ZSQINCH $24, $12, R22     // 16f37b04
    ZSQINCH $25, $13, R23     // 37f37c04
    ZSQINCH $26, $13, R24     // 58f37c04
    ZSQINCH $27, $14, R25     // 79f37d04
    ZSQINCH $28, $14, R26     // 9af37d04
    ZSQINCH MUL4, $15, R27    // bbf37e04
    ZSQINCH MUL3, $15, R27    // dbf37e04
    ZSQINCH ALL, $16, R30     // fef37f04

// SQINCH  <Zdn>.H{, <pattern>{, MUL #<imm>}}
    ZSQINCH POW2, $1, Z0.H      // 00c06004
    ZSQINCH VL1, $1, Z1.H       // 21c06004
    ZSQINCH VL2, $2, Z2.H       // 42c06104
    ZSQINCH VL3, $2, Z3.H       // 63c06104
    ZSQINCH VL4, $3, Z4.H       // 84c06204
    ZSQINCH VL5, $3, Z5.H       // a5c06204
    ZSQINCH VL6, $4, Z6.H       // c6c06304
    ZSQINCH VL7, $4, Z7.H       // e7c06304
    ZSQINCH VL8, $5, Z8.H       // 08c16404
    ZSQINCH VL16, $5, Z9.H      // 29c16404
    ZSQINCH VL32, $6, Z10.H     // 4ac16504
    ZSQINCH VL64, $6, Z11.H     // 6bc16504
    ZSQINCH VL128, $7, Z12.H    // 8cc16604
    ZSQINCH VL256, $7, Z13.H    // adc16604
    ZSQINCH $14, $8, Z14.H      // cec16704
    ZSQINCH $15, $8, Z15.H      // efc16704
    ZSQINCH $16, $9, Z16.H      // 10c26804
    ZSQINCH $17, $9, Z16.H      // 30c26804
    ZSQINCH $18, $9, Z17.H      // 51c26804
    ZSQINCH $19, $10, Z18.H     // 72c26904
    ZSQINCH $20, $10, Z19.H     // 93c26904
    ZSQINCH $21, $11, Z20.H     // b4c26a04
    ZSQINCH $22, $11, Z21.H     // d5c26a04
    ZSQINCH $23, $12, Z22.H     // f6c26b04
    ZSQINCH $24, $12, Z23.H     // 17c36b04
    ZSQINCH $25, $13, Z24.H     // 38c36c04
    ZSQINCH $26, $13, Z25.H     // 59c36c04
    ZSQINCH $27, $14, Z26.H     // 7ac36d04
    ZSQINCH $28, $14, Z27.H     // 9bc36d04
    ZSQINCH MUL4, $15, Z28.H    // bcc36e04
    ZSQINCH MUL3, $15, Z29.H    // ddc36e04
    ZSQINCH ALL, $16, Z31.H     // ffc36f04

// SQINCP  <Xdn>, <Pm>.<T>, <Wdn>
    PSQINCPW P0.B, R0, R0       // 00882825
    PSQINCPW P6.B, R10, R10     // ca882825
    PSQINCPW P15.B, R30, R30    // fe892825
    PSQINCPW P0.H, R0, R0       // 00886825
    PSQINCPW P6.H, R10, R10     // ca886825
    PSQINCPW P15.H, R30, R30    // fe896825
    PSQINCPW P0.S, R0, R0       // 0088a825
    PSQINCPW P6.S, R10, R10     // ca88a825
    PSQINCPW P15.S, R30, R30    // fe89a825
    PSQINCPW P0.D, R0, R0       // 0088e825
    PSQINCPW P6.D, R10, R10     // ca88e825
    PSQINCPW P15.D, R30, R30    // fe89e825

// SQINCP  <Xdn>, <Pm>.<T>
    PSQINCP P0.B, R0      // 008c2825
    PSQINCP P6.B, R10     // ca8c2825
    PSQINCP P15.B, R30    // fe8d2825
    PSQINCP P0.H, R0      // 008c6825
    PSQINCP P6.H, R10     // ca8c6825
    PSQINCP P15.H, R30    // fe8d6825
    PSQINCP P0.S, R0      // 008ca825
    PSQINCP P6.S, R10     // ca8ca825
    PSQINCP P15.S, R30    // fe8da825
    PSQINCP P0.D, R0      // 008ce825
    PSQINCP P6.D, R10     // ca8ce825
    PSQINCP P15.D, R30    // fe8de825

// SQINCP  <Zdn>.<T>, <Pm>.<T>
    ZSQINCP P0.H, Z0.H      // 00806825
    ZSQINCP P6.H, Z10.H     // ca806825
    ZSQINCP P15.H, Z31.H    // ff816825
    ZSQINCP P0.S, Z0.S      // 0080a825
    ZSQINCP P6.S, Z10.S     // ca80a825
    ZSQINCP P15.S, Z31.S    // ff81a825
    ZSQINCP P0.D, Z0.D      // 0080e825
    ZSQINCP P6.D, Z10.D     // ca80e825
    ZSQINCP P15.D, Z31.D    // ff81e825

// SQINCW  <Xdn>, <Wdn>{, <pattern>{, MUL #<imm>}}
    ZSQINCWW R0, POW2, $1, R0       // 00f0a004
    ZSQINCWW R1, VL1, $1, R1        // 21f0a004
    ZSQINCWW R2, VL2, $2, R2        // 42f0a104
    ZSQINCWW R3, VL3, $2, R3        // 63f0a104
    ZSQINCWW R4, VL4, $3, R4        // 84f0a204
    ZSQINCWW R5, VL5, $3, R5        // a5f0a204
    ZSQINCWW R6, VL6, $4, R6        // c6f0a304
    ZSQINCWW R7, VL7, $4, R7        // e7f0a304
    ZSQINCWW R8, VL8, $5, R8        // 08f1a404
    ZSQINCWW R8, VL16, $5, R8       // 28f1a404
    ZSQINCWW R9, VL32, $6, R9       // 49f1a504
    ZSQINCWW R10, VL64, $6, R10     // 6af1a504
    ZSQINCWW R11, VL128, $7, R11    // 8bf1a604
    ZSQINCWW R12, VL256, $7, R12    // acf1a604
    ZSQINCWW R13, $14, $8, R13      // cdf1a704
    ZSQINCWW R14, $15, $8, R14      // eef1a704
    ZSQINCWW R15, $16, $9, R15      // 0ff2a804
    ZSQINCWW R16, $17, $9, R16      // 30f2a804
    ZSQINCWW R17, $18, $9, R17      // 51f2a804
    ZSQINCWW R17, $19, $10, R17     // 71f2a904
    ZSQINCWW R19, $20, $10, R19     // 93f2a904
    ZSQINCWW R20, $21, $11, R20     // b4f2aa04
    ZSQINCWW R21, $22, $11, R21     // d5f2aa04
    ZSQINCWW R22, $23, $12, R22     // f6f2ab04
    ZSQINCWW R22, $24, $12, R22     // 16f3ab04
    ZSQINCWW R23, $25, $13, R23     // 37f3ac04
    ZSQINCWW R24, $26, $13, R24     // 58f3ac04
    ZSQINCWW R25, $27, $14, R25     // 79f3ad04
    ZSQINCWW R26, $28, $14, R26     // 9af3ad04
    ZSQINCWW R27, MUL4, $15, R27    // bbf3ae04
    ZSQINCWW R27, MUL3, $15, R27    // dbf3ae04
    ZSQINCWW R30, ALL, $16, R30     // fef3af04

// SQINCW  <Xdn>{, <pattern>{, MUL #<imm>}}
    ZSQINCW POW2, $1, R0      // 00f0b004
    ZSQINCW VL1, $1, R1       // 21f0b004
    ZSQINCW VL2, $2, R2       // 42f0b104
    ZSQINCW VL3, $2, R3       // 63f0b104
    ZSQINCW VL4, $3, R4       // 84f0b204
    ZSQINCW VL5, $3, R5       // a5f0b204
    ZSQINCW VL6, $4, R6       // c6f0b304
    ZSQINCW VL7, $4, R7       // e7f0b304
    ZSQINCW VL8, $5, R8       // 08f1b404
    ZSQINCW VL16, $5, R8      // 28f1b404
    ZSQINCW VL32, $6, R9      // 49f1b504
    ZSQINCW VL64, $6, R10     // 6af1b504
    ZSQINCW VL128, $7, R11    // 8bf1b604
    ZSQINCW VL256, $7, R12    // acf1b604
    ZSQINCW $14, $8, R13      // cdf1b704
    ZSQINCW $15, $8, R14      // eef1b704
    ZSQINCW $16, $9, R15      // 0ff2b804
    ZSQINCW $17, $9, R16      // 30f2b804
    ZSQINCW $18, $9, R17      // 51f2b804
    ZSQINCW $19, $10, R17     // 71f2b904
    ZSQINCW $20, $10, R19     // 93f2b904
    ZSQINCW $21, $11, R20     // b4f2ba04
    ZSQINCW $22, $11, R21     // d5f2ba04
    ZSQINCW $23, $12, R22     // f6f2bb04
    ZSQINCW $24, $12, R22     // 16f3bb04
    ZSQINCW $25, $13, R23     // 37f3bc04
    ZSQINCW $26, $13, R24     // 58f3bc04
    ZSQINCW $27, $14, R25     // 79f3bd04
    ZSQINCW $28, $14, R26     // 9af3bd04
    ZSQINCW MUL4, $15, R27    // bbf3be04
    ZSQINCW MUL3, $15, R27    // dbf3be04
    ZSQINCW ALL, $16, R30     // fef3bf04

// SQINCW  <Zdn>.S{, <pattern>{, MUL #<imm>}}
    ZSQINCW POW2, $1, Z0.S      // 00c0a004
    ZSQINCW VL1, $1, Z1.S       // 21c0a004
    ZSQINCW VL2, $2, Z2.S       // 42c0a104
    ZSQINCW VL3, $2, Z3.S       // 63c0a104
    ZSQINCW VL4, $3, Z4.S       // 84c0a204
    ZSQINCW VL5, $3, Z5.S       // a5c0a204
    ZSQINCW VL6, $4, Z6.S       // c6c0a304
    ZSQINCW VL7, $4, Z7.S       // e7c0a304
    ZSQINCW VL8, $5, Z8.S       // 08c1a404
    ZSQINCW VL16, $5, Z9.S      // 29c1a404
    ZSQINCW VL32, $6, Z10.S     // 4ac1a504
    ZSQINCW VL64, $6, Z11.S     // 6bc1a504
    ZSQINCW VL128, $7, Z12.S    // 8cc1a604
    ZSQINCW VL256, $7, Z13.S    // adc1a604
    ZSQINCW $14, $8, Z14.S      // cec1a704
    ZSQINCW $15, $8, Z15.S      // efc1a704
    ZSQINCW $16, $9, Z16.S      // 10c2a804
    ZSQINCW $17, $9, Z16.S      // 30c2a804
    ZSQINCW $18, $9, Z17.S      // 51c2a804
    ZSQINCW $19, $10, Z18.S     // 72c2a904
    ZSQINCW $20, $10, Z19.S     // 93c2a904
    ZSQINCW $21, $11, Z20.S     // b4c2aa04
    ZSQINCW $22, $11, Z21.S     // d5c2aa04
    ZSQINCW $23, $12, Z22.S     // f6c2ab04
    ZSQINCW $24, $12, Z23.S     // 17c3ab04
    ZSQINCW $25, $13, Z24.S     // 38c3ac04
    ZSQINCW $26, $13, Z25.S     // 59c3ac04
    ZSQINCW $27, $14, Z26.S     // 7ac3ad04
    ZSQINCW $28, $14, Z27.S     // 9bc3ad04
    ZSQINCW MUL4, $15, Z28.S    // bcc3ae04
    ZSQINCW MUL3, $15, Z29.S    // ddc3ae04
    ZSQINCW ALL, $16, Z31.S     // ffc3af04

// SQSUB   <Zdn>.<T>, <Zdn>.<T>, #<imm>, <shift>
    ZSQSUB Z0.B, $0, $0, Z0.B        // 00c02625
    ZSQSUB Z10.B, $85, $0, Z10.B     // aaca2625
    ZSQSUB Z31.B, $255, $0, Z31.B    // ffdf2625
    ZSQSUB Z0.H, $0, $8, Z0.H        // 00e06625
    ZSQSUB Z10.H, $85, $8, Z10.H     // aaea6625
    ZSQSUB Z31.H, $255, $0, Z31.H    // ffdf6625
    ZSQSUB Z0.S, $0, $8, Z0.S        // 00e0a625
    ZSQSUB Z10.S, $85, $8, Z10.S     // aaeaa625
    ZSQSUB Z31.S, $255, $0, Z31.S    // ffdfa625
    ZSQSUB Z0.D, $0, $8, Z0.D        // 00e0e625
    ZSQSUB Z10.D, $85, $8, Z10.D     // aaeae625
    ZSQSUB Z31.D, $255, $0, Z31.D    // ffdfe625

// SQSUB   <Zd>.<T>, <Zn>.<T>, <Zm>.<T>
    ZSQSUB Z0.B, Z0.B, Z0.B       // 00182004
    ZSQSUB Z11.B, Z12.B, Z10.B    // 6a192c04
    ZSQSUB Z31.B, Z31.B, Z31.B    // ff1b3f04
    ZSQSUB Z0.H, Z0.H, Z0.H       // 00186004
    ZSQSUB Z11.H, Z12.H, Z10.H    // 6a196c04
    ZSQSUB Z31.H, Z31.H, Z31.H    // ff1b7f04
    ZSQSUB Z0.S, Z0.S, Z0.S       // 0018a004
    ZSQSUB Z11.S, Z12.S, Z10.S    // 6a19ac04
    ZSQSUB Z31.S, Z31.S, Z31.S    // ff1bbf04
    ZSQSUB Z0.D, Z0.D, Z0.D       // 0018e004
    ZSQSUB Z11.D, Z12.D, Z10.D    // 6a19ec04
    ZSQSUB Z31.D, Z31.D, Z31.D    // ff1bff04

// ST1B    { <Zt>.D }, <Pg>, [<Zn>.D{, #<pimm>}]
    ZST1B [Z0.D], P0, (Z0.D)       // 00a040e4
    ZST1B [Z10.D], P3, 10(Z12.D)    // 8aad4ae4
    ZST1B [Z31.D], P7, 31(Z31.D)    // ffbf5fe4

// ST1B    { <Zt>.S }, <Pg>, [<Zn>.S{, #<pimm>}]
    ZST1B [Z0.S], P0, (Z0.S)       // 00a060e4
    ZST1B [Z10.S], P3, 10(Z12.S)    // 8aad6ae4
    ZST1B [Z31.S], P7, 31(Z31.S)    // ffbf7fe4

// ST1B    { <Zt>.<T> }, <Pg>, [<Xn|SP>{, #<simm>, MUL VL}]
    ZST1B [Z0.B], P0, -8(R0)      // 00e008e4
    ZST1B [Z10.B], P3, -3(R12)    // 8aed0de4
    ZST1B [Z31.B], P7, 7(R30)     // dfff07e4
    ZST1B [Z0.H], P0, -8(R0)      // 00e028e4
    ZST1B [Z10.H], P3, -3(R12)    // 8aed2de4
    ZST1B [Z31.H], P7, 7(R30)     // dfff27e4
    ZST1B [Z0.S], P0, -8(R0)      // 00e048e4
    ZST1B [Z10.S], P3, -3(R12)    // 8aed4de4
    ZST1B [Z31.S], P7, 7(R30)     // dfff47e4
    ZST1B [Z0.D], P0, -8(R0)      // 00e068e4
    ZST1B [Z10.D], P3, -3(R12)    // 8aed6de4
    ZST1B [Z31.D], P7, 7(R30)     // dfff67e4

// ST1B    { <Zt>.<T> }, <Pg>, [<Xn|SP>, <Xm>]
    ZST1B [Z0.B], P0, (R0)(R0)       // 004000e4
    ZST1B [Z10.B], P3, (R12)(R13)    // 8a4d0de4
    ZST1B [Z31.B], P7, (R30)(R30)    // df5f1ee4
    ZST1B [Z0.H], P0, (R0)(R0)       // 004020e4
    ZST1B [Z10.H], P3, (R12)(R13)    // 8a4d2de4
    ZST1B [Z31.H], P7, (R30)(R30)    // df5f3ee4
    ZST1B [Z0.S], P0, (R0)(R0)       // 004040e4
    ZST1B [Z10.S], P3, (R12)(R13)    // 8a4d4de4
    ZST1B [Z31.S], P7, (R30)(R30)    // df5f5ee4
    ZST1B [Z0.D], P0, (R0)(R0)       // 004060e4
    ZST1B [Z10.D], P3, (R12)(R13)    // 8a4d6de4
    ZST1B [Z31.D], P7, (R30)(R30)    // df5f7ee4

// ST1B    { <Zt>.D }, <Pg>, [<Xn|SP>, <Zm>.D]
    ZST1B [Z0.D], P0, (R0)(Z0.D)       // 00a000e4
    ZST1B [Z10.D], P3, (R12)(Z13.D)    // 8aad0de4
    ZST1B [Z31.D], P7, (R30)(Z31.D)    // dfbf1fe4

// ST1B    { <Zt>.D }, <Pg>, [<Xn|SP>, <Zm>.D, <extend>]
    ZST1B [Z0.D], P0, (R0)(Z0.D.UXTW)       // 008000e4
    ZST1B [Z10.D], P3, (R12)(Z13.D.UXTW)    // 8a8d0de4
    ZST1B [Z31.D], P7, (R30)(Z31.D.UXTW)    // df9f1fe4
    ZST1B [Z0.D], P0, (R0)(Z0.D.SXTW)       // 00c000e4
    ZST1B [Z10.D], P3, (R12)(Z13.D.SXTW)    // 8acd0de4
    ZST1B [Z31.D], P7, (R30)(Z31.D.SXTW)    // dfdf1fe4

// ST1B    { <Zt>.S }, <Pg>, [<Xn|SP>, <Zm>.S, <extend>]
    ZST1B [Z0.S], P0, (R0)(Z0.S.UXTW)       // 008040e4
    ZST1B [Z10.S], P3, (R12)(Z13.S.UXTW)    // 8a8d4de4
    ZST1B [Z31.S], P7, (R30)(Z31.S.UXTW)    // df9f5fe4
    ZST1B [Z0.S], P0, (R0)(Z0.S.SXTW)       // 00c040e4
    ZST1B [Z10.S], P3, (R12)(Z13.S.SXTW)    // 8acd4de4
    ZST1B [Z31.S], P7, (R30)(Z31.S.SXTW)    // dfdf5fe4

// ST1D    { <Zt>.D }, <Pg>, [<Zn>.D{, #<pimm>}]
    ZST1D [Z0.D], P0, (Z0.D)        // 00a0c0e5
    ZST1D [Z10.D], P3, 80(Z12.D)     // 8aadcae5
    ZST1D [Z31.D], P7, 248(Z31.D)    // ffbfdfe5

// ST1D    { <Zt>.D }, <Pg>, [<Xn|SP>{, #<simm>, MUL VL}]
    ZST1D [Z0.D], P0, -8(R0)      // 00e0e8e5
    ZST1D [Z10.D], P3, -3(R12)    // 8aedede5
    ZST1D [Z31.D], P7, 7(R30)     // dfffe7e5

// ST1D    { <Zt>.D }, <Pg>, [<Xn|SP>, <Xm>, LSL #3]
    ZST1D [Z0.D], P0, (R0)(R0<<3)       // 0040e0e5
    ZST1D [Z10.D], P3, (R12)(R13<<3)    // 8a4dede5
    ZST1D [Z31.D], P7, (R30)(R30<<3)    // df5ffee5

// ST1D    { <Zt>.D }, <Pg>, [<Xn|SP>, <Zm>.D, LSL #3]
    ZST1D [Z0.D], P0, (R0)(Z0.D.LSL<<3)       // 00a0a0e5
    ZST1D [Z10.D], P3, (R12)(Z13.D.LSL<<3)    // 8aadade5
    ZST1D [Z31.D], P7, (R30)(Z31.D.LSL<<3)    // dfbfbfe5

// ST1D    { <Zt>.D }, <Pg>, [<Xn|SP>, <Zm>.D]
    ZST1D [Z0.D], P0, (R0)(Z0.D)       // 00a080e5
    ZST1D [Z10.D], P3, (R12)(Z13.D)    // 8aad8de5
    ZST1D [Z31.D], P7, (R30)(Z31.D)    // dfbf9fe5

// ST1D    { <Zt>.D }, <Pg>, [<Xn|SP>, <Zm>.D, <extend> #3]
    ZST1D [Z0.D], P0, (R0)(Z0.D.UXTW<<3)       // 0080a0e5
    ZST1D [Z10.D], P3, (R12)(Z13.D.UXTW<<3)    // 8a8dade5
    ZST1D [Z31.D], P7, (R30)(Z31.D.UXTW<<3)    // df9fbfe5
    ZST1D [Z0.D], P0, (R0)(Z0.D.SXTW<<3)       // 00c0a0e5
    ZST1D [Z10.D], P3, (R12)(Z13.D.SXTW<<3)    // 8acdade5
    ZST1D [Z31.D], P7, (R30)(Z31.D.SXTW<<3)    // dfdfbfe5

// ST1D    { <Zt>.D }, <Pg>, [<Xn|SP>, <Zm>.D, <extend>]
    ZST1D [Z0.D], P0, (R0)(Z0.D.UXTW)       // 008080e5
    ZST1D [Z10.D], P3, (R12)(Z13.D.UXTW)    // 8a8d8de5
    ZST1D [Z31.D], P7, (R30)(Z31.D.UXTW)    // df9f9fe5
    ZST1D [Z0.D], P0, (R0)(Z0.D.SXTW)       // 00c080e5
    ZST1D [Z10.D], P3, (R12)(Z13.D.SXTW)    // 8acd8de5
    ZST1D [Z31.D], P7, (R30)(Z31.D.SXTW)    // dfdf9fe5

// ST1H    { <Zt>.D }, <Pg>, [<Zn>.D{, #<pimm>}]
    ZST1H [Z0.D], P0, (Z0.D)       // 00a0c0e4
    ZST1H [Z10.D], P3, 20(Z12.D)    // 8aadcae4
    ZST1H [Z31.D], P7, 62(Z31.D)    // ffbfdfe4

// ST1H    { <Zt>.S }, <Pg>, [<Zn>.S{, #<pimm>}]
    ZST1H [Z0.S], P0, (Z0.S)       // 00a0e0e4
    ZST1H [Z10.S], P3, 20(Z12.S)    // 8aadeae4
    ZST1H [Z31.S], P7, 62(Z31.S)    // ffbfffe4

// ST1H    { <Zt>.<T> }, <Pg>, [<Xn|SP>{, #<simm>, MUL VL}]
    ZST1H [Z0.H], P0, -8(R0)      // 00e0a8e4
    ZST1H [Z10.H], P3, -3(R12)    // 8aedade4
    ZST1H [Z31.H], P7, 7(R30)     // dfffa7e4
    ZST1H [Z0.S], P0, -8(R0)      // 00e0c8e4
    ZST1H [Z10.S], P3, -3(R12)    // 8aedcde4
    ZST1H [Z31.S], P7, 7(R30)     // dfffc7e4
    ZST1H [Z0.D], P0, -8(R0)      // 00e0e8e4
    ZST1H [Z10.D], P3, -3(R12)    // 8aedede4
    ZST1H [Z31.D], P7, 7(R30)     // dfffe7e4

// ST1H    { <Zt>.<T> }, <Pg>, [<Xn|SP>, <Xm>, LSL #1]
    ZST1H [Z0.H], P0, (R0)(R0<<1)       // 0040a0e4
    ZST1H [Z10.H], P3, (R12)(R13<<1)    // 8a4dade4
    ZST1H [Z31.H], P7, (R30)(R30<<1)    // df5fbee4
    ZST1H [Z0.S], P0, (R0)(R0<<1)       // 0040c0e4
    ZST1H [Z10.S], P3, (R12)(R13<<1)    // 8a4dcde4
    ZST1H [Z31.S], P7, (R30)(R30<<1)    // df5fdee4
    ZST1H [Z0.D], P0, (R0)(R0<<1)       // 0040e0e4
    ZST1H [Z10.D], P3, (R12)(R13<<1)    // 8a4dede4
    ZST1H [Z31.D], P7, (R30)(R30<<1)    // df5ffee4

// ST1H    { <Zt>.D }, <Pg>, [<Xn|SP>, <Zm>.D, LSL #1]
    ZST1H [Z0.D], P0, (R0)(Z0.D.LSL<<1)       // 00a0a0e4
    ZST1H [Z10.D], P3, (R12)(Z13.D.LSL<<1)    // 8aadade4
    ZST1H [Z31.D], P7, (R30)(Z31.D.LSL<<1)    // dfbfbfe4

// ST1H    { <Zt>.D }, <Pg>, [<Xn|SP>, <Zm>.D]
    ZST1H [Z0.D], P0, (R0)(Z0.D)       // 00a080e4
    ZST1H [Z10.D], P3, (R12)(Z13.D)    // 8aad8de4
    ZST1H [Z31.D], P7, (R30)(Z31.D)    // dfbf9fe4

// ST1H    { <Zt>.D }, <Pg>, [<Xn|SP>, <Zm>.D, <extend> #1]
    ZST1H [Z0.D], P0, (R0)(Z0.D.UXTW<<1)       // 0080a0e4
    ZST1H [Z10.D], P3, (R12)(Z13.D.UXTW<<1)    // 8a8dade4
    ZST1H [Z31.D], P7, (R30)(Z31.D.UXTW<<1)    // df9fbfe4
    ZST1H [Z0.D], P0, (R0)(Z0.D.SXTW<<1)       // 00c0a0e4
    ZST1H [Z10.D], P3, (R12)(Z13.D.SXTW<<1)    // 8acdade4
    ZST1H [Z31.D], P7, (R30)(Z31.D.SXTW<<1)    // dfdfbfe4

// ST1H    { <Zt>.D }, <Pg>, [<Xn|SP>, <Zm>.D, <extend>]
    ZST1H [Z0.D], P0, (R0)(Z0.D.UXTW)       // 008080e4
    ZST1H [Z10.D], P3, (R12)(Z13.D.UXTW)    // 8a8d8de4
    ZST1H [Z31.D], P7, (R30)(Z31.D.UXTW)    // df9f9fe4
    ZST1H [Z0.D], P0, (R0)(Z0.D.SXTW)       // 00c080e4
    ZST1H [Z10.D], P3, (R12)(Z13.D.SXTW)    // 8acd8de4
    ZST1H [Z31.D], P7, (R30)(Z31.D.SXTW)    // dfdf9fe4

// ST1H    { <Zt>.S }, <Pg>, [<Xn|SP>, <Zm>.S, <extend> #1]
    ZST1H [Z0.S], P0, (R0)(Z0.S.UXTW<<1)       // 0080e0e4
    ZST1H [Z10.S], P3, (R12)(Z13.S.UXTW<<1)    // 8a8dede4
    ZST1H [Z31.S], P7, (R30)(Z31.S.UXTW<<1)    // df9fffe4
    ZST1H [Z0.S], P0, (R0)(Z0.S.SXTW<<1)       // 00c0e0e4
    ZST1H [Z10.S], P3, (R12)(Z13.S.SXTW<<1)    // 8acdede4
    ZST1H [Z31.S], P7, (R30)(Z31.S.SXTW<<1)    // dfdfffe4

// ST1H    { <Zt>.S }, <Pg>, [<Xn|SP>, <Zm>.S, <extend>]
    ZST1H [Z0.S], P0, (R0)(Z0.S.UXTW)       // 0080c0e4
    ZST1H [Z10.S], P3, (R12)(Z13.S.UXTW)    // 8a8dcde4
    ZST1H [Z31.S], P7, (R30)(Z31.S.UXTW)    // df9fdfe4
    ZST1H [Z0.S], P0, (R0)(Z0.S.SXTW)       // 00c0c0e4
    ZST1H [Z10.S], P3, (R12)(Z13.S.SXTW)    // 8acdcde4
    ZST1H [Z31.S], P7, (R30)(Z31.S.SXTW)    // dfdfdfe4

// ST1W    { <Zt>.D }, <Pg>, [<Zn>.D{, #<pimm>}]
    ZST1W [Z0.D], P0, (Z0.D)        // 00a040e5
    ZST1W [Z10.D], P3, 40(Z12.D)     // 8aad4ae5
    ZST1W [Z31.D], P7, 124(Z31.D)    // ffbf5fe5

// ST1W    { <Zt>.S }, <Pg>, [<Zn>.S{, #<pimm>}]
    ZST1W [Z0.S], P0, (Z0.S)        // 00a060e5
    ZST1W [Z10.S], P3, 40(Z12.S)     // 8aad6ae5
    ZST1W [Z31.S], P7, 124(Z31.S)    // ffbf7fe5

// ST1W    { <Zt>.<T> }, <Pg>, [<Xn|SP>{, #<simm>, MUL VL}]
    ZST1W [Z0.S], P0, -8(R0)      // 00e048e5
    ZST1W [Z10.S], P3, -3(R12)    // 8aed4de5
    ZST1W [Z31.S], P7, 7(R30)     // dfff47e5
    ZST1W [Z0.D], P0, -8(R0)      // 00e068e5
    ZST1W [Z10.D], P3, -3(R12)    // 8aed6de5
    ZST1W [Z31.D], P7, 7(R30)     // dfff67e5

// ST1W    { <Zt>.<T> }, <Pg>, [<Xn|SP>, <Xm>, LSL #2]
    ZST1W [Z0.S], P0, (R0)(R0<<2)       // 004040e5
    ZST1W [Z10.S], P3, (R12)(R13<<2)    // 8a4d4de5
    ZST1W [Z31.S], P7, (R30)(R30<<2)    // df5f5ee5
    ZST1W [Z0.D], P0, (R0)(R0<<2)       // 004060e5
    ZST1W [Z10.D], P3, (R12)(R13<<2)    // 8a4d6de5
    ZST1W [Z31.D], P7, (R30)(R30<<2)    // df5f7ee5

// ST1W    { <Zt>.D }, <Pg>, [<Xn|SP>, <Zm>.D, LSL #2]
    ZST1W [Z0.D], P0, (R0)(Z0.D.LSL<<2)       // 00a020e5
    ZST1W [Z10.D], P3, (R12)(Z13.D.LSL<<2)    // 8aad2de5
    ZST1W [Z31.D], P7, (R30)(Z31.D.LSL<<2)    // dfbf3fe5

// ST1W    { <Zt>.D }, <Pg>, [<Xn|SP>, <Zm>.D]
    ZST1W [Z0.D], P0, (R0)(Z0.D)       // 00a000e5
    ZST1W [Z10.D], P3, (R12)(Z13.D)    // 8aad0de5
    ZST1W [Z31.D], P7, (R30)(Z31.D)    // dfbf1fe5

// ST1W    { <Zt>.D }, <Pg>, [<Xn|SP>, <Zm>.D, <extend> #2]
    ZST1W [Z0.D], P0, (R0)(Z0.D.UXTW<<2)       // 008020e5
    ZST1W [Z10.D], P3, (R12)(Z13.D.UXTW<<2)    // 8a8d2de5
    ZST1W [Z31.D], P7, (R30)(Z31.D.UXTW<<2)    // df9f3fe5
    ZST1W [Z0.D], P0, (R0)(Z0.D.SXTW<<2)       // 00c020e5
    ZST1W [Z10.D], P3, (R12)(Z13.D.SXTW<<2)    // 8acd2de5
    ZST1W [Z31.D], P7, (R30)(Z31.D.SXTW<<2)    // dfdf3fe5

// ST1W    { <Zt>.D }, <Pg>, [<Xn|SP>, <Zm>.D, <extend>]
    ZST1W [Z0.D], P0, (R0)(Z0.D.UXTW)       // 008000e5
    ZST1W [Z10.D], P3, (R12)(Z13.D.UXTW)    // 8a8d0de5
    ZST1W [Z31.D], P7, (R30)(Z31.D.UXTW)    // df9f1fe5
    ZST1W [Z0.D], P0, (R0)(Z0.D.SXTW)       // 00c000e5
    ZST1W [Z10.D], P3, (R12)(Z13.D.SXTW)    // 8acd0de5
    ZST1W [Z31.D], P7, (R30)(Z31.D.SXTW)    // dfdf1fe5

// ST1W    { <Zt>.S }, <Pg>, [<Xn|SP>, <Zm>.S, <extend> #2]
    ZST1W [Z0.S], P0, (R0)(Z0.S.UXTW<<2)       // 008060e5
    ZST1W [Z10.S], P3, (R12)(Z13.S.UXTW<<2)    // 8a8d6de5
    ZST1W [Z31.S], P7, (R30)(Z31.S.UXTW<<2)    // df9f7fe5
    ZST1W [Z0.S], P0, (R0)(Z0.S.SXTW<<2)       // 00c060e5
    ZST1W [Z10.S], P3, (R12)(Z13.S.SXTW<<2)    // 8acd6de5
    ZST1W [Z31.S], P7, (R30)(Z31.S.SXTW<<2)    // dfdf7fe5

// ST1W    { <Zt>.S }, <Pg>, [<Xn|SP>, <Zm>.S, <extend>]
    ZST1W [Z0.S], P0, (R0)(Z0.S.UXTW)       // 008040e5
    ZST1W [Z10.S], P3, (R12)(Z13.S.UXTW)    // 8a8d4de5
    ZST1W [Z31.S], P7, (R30)(Z31.S.UXTW)    // df9f5fe5
    ZST1W [Z0.S], P0, (R0)(Z0.S.SXTW)       // 00c040e5
    ZST1W [Z10.S], P3, (R12)(Z13.S.SXTW)    // 8acd4de5
    ZST1W [Z31.S], P7, (R30)(Z31.S.SXTW)    // dfdf5fe5

// ST2B    { <Zt1>.B, <Zt2>.B }, <Pg>, [<Xn|SP>{, #<simm>, MUL VL}]
    ZST2B [Z0.B, Z1.B], P0, -16(R0)      // 00e038e4
    ZST2B [Z10.B, Z11.B], P3, -6(R12)    // 8aed3de4
    ZST2B [Z31.B, Z0.B], P7, 14(R30)     // dfff37e4

// ST2B    { <Zt1>.B, <Zt2>.B }, <Pg>, [<Xn|SP>, <Xm>]
    ZST2B [Z0.B, Z1.B], P0, (R0)(R0)        // 006020e4
    ZST2B [Z10.B, Z11.B], P3, (R12)(R13)    // 8a6d2de4
    ZST2B [Z31.B, Z0.B], P7, (R30)(R30)     // df7f3ee4

// ST2D    { <Zt1>.D, <Zt2>.D }, <Pg>, [<Xn|SP>{, #<simm>, MUL VL}]
    ZST2D [Z0.D, Z1.D], P0, -16(R0)      // 00e0b8e5
    ZST2D [Z10.D, Z11.D], P3, -6(R12)    // 8aedbde5
    ZST2D [Z31.D, Z0.D], P7, 14(R30)     // dfffb7e5

// ST2D    { <Zt1>.D, <Zt2>.D }, <Pg>, [<Xn|SP>, <Xm>, LSL #3]
    ZST2D [Z0.D, Z1.D], P0, (R0)(R0<<3)        // 0060a0e5
    ZST2D [Z10.D, Z11.D], P3, (R12)(R13<<3)    // 8a6dade5
    ZST2D [Z31.D, Z0.D], P7, (R30)(R30<<3)     // df7fbee5

// ST2H    { <Zt1>.H, <Zt2>.H }, <Pg>, [<Xn|SP>{, #<simm>, MUL VL}]
    ZST2H [Z0.H, Z1.H], P0, -16(R0)      // 00e0b8e4
    ZST2H [Z10.H, Z11.H], P3, -6(R12)    // 8aedbde4
    ZST2H [Z31.H, Z0.H], P7, 14(R30)     // dfffb7e4

// ST2H    { <Zt1>.H, <Zt2>.H }, <Pg>, [<Xn|SP>, <Xm>, LSL #1]
    ZST2H [Z0.H, Z1.H], P0, (R0)(R0<<1)        // 0060a0e4
    ZST2H [Z10.H, Z11.H], P3, (R12)(R13<<1)    // 8a6dade4
    ZST2H [Z31.H, Z0.H], P7, (R30)(R30<<1)     // df7fbee4

// ST2W    { <Zt1>.S, <Zt2>.S }, <Pg>, [<Xn|SP>{, #<simm>, MUL VL}]
    ZST2W [Z0.S, Z1.S], P0, -16(R0)      // 00e038e5
    ZST2W [Z10.S, Z11.S], P3, -6(R12)    // 8aed3de5
    ZST2W [Z31.S, Z0.S], P7, 14(R30)     // dfff37e5

// ST2W    { <Zt1>.S, <Zt2>.S }, <Pg>, [<Xn|SP>, <Xm>, LSL #2]
    ZST2W [Z0.S, Z1.S], P0, (R0)(R0<<2)        // 006020e5
    ZST2W [Z10.S, Z11.S], P3, (R12)(R13<<2)    // 8a6d2de5
    ZST2W [Z31.S, Z0.S], P7, (R30)(R30<<2)     // df7f3ee5

// ST3B    { <Zt1>.B, <Zt2>.B, <Zt3>.B }, <Pg>, [<Xn|SP>{, #<simm>, MUL VL}]
    ZST3B [Z0.B, Z1.B, Z2.B], P0, -24(R0)       // 00e058e4
    ZST3B [Z10.B, Z11.B, Z12.B], P3, -9(R12)    // 8aed5de4
    ZST3B [Z31.B, Z0.B, Z1.B], P7, 21(R30)      // dfff57e4

// ST3B    { <Zt1>.B, <Zt2>.B, <Zt3>.B }, <Pg>, [<Xn|SP>, <Xm>]
    ZST3B [Z0.B, Z1.B, Z2.B], P0, (R0)(R0)         // 006040e4
    ZST3B [Z10.B, Z11.B, Z12.B], P3, (R12)(R13)    // 8a6d4de4
    ZST3B [Z31.B, Z0.B, Z1.B], P7, (R30)(R30)      // df7f5ee4

// ST3D    { <Zt1>.D, <Zt2>.D, <Zt3>.D }, <Pg>, [<Xn|SP>{, #<simm>, MUL VL}]
    ZST3D [Z0.D, Z1.D, Z2.D], P0, -24(R0)       // 00e0d8e5
    ZST3D [Z10.D, Z11.D, Z12.D], P3, -9(R12)    // 8aeddde5
    ZST3D [Z31.D, Z0.D, Z1.D], P7, 21(R30)      // dfffd7e5

// ST3D    { <Zt1>.D, <Zt2>.D, <Zt3>.D }, <Pg>, [<Xn|SP>, <Xm>, LSL #3]
    ZST3D [Z0.D, Z1.D, Z2.D], P0, (R0)(R0<<3)         // 0060c0e5
    ZST3D [Z10.D, Z11.D, Z12.D], P3, (R12)(R13<<3)    // 8a6dcde5
    ZST3D [Z31.D, Z0.D, Z1.D], P7, (R30)(R30<<3)      // df7fdee5

// ST3H    { <Zt1>.H, <Zt2>.H, <Zt3>.H }, <Pg>, [<Xn|SP>{, #<simm>, MUL VL}]
    ZST3H [Z0.H, Z1.H, Z2.H], P0, -24(R0)       // 00e0d8e4
    ZST3H [Z10.H, Z11.H, Z12.H], P3, -9(R12)    // 8aeddde4
    ZST3H [Z31.H, Z0.H, Z1.H], P7, 21(R30)      // dfffd7e4

// ST3H    { <Zt1>.H, <Zt2>.H, <Zt3>.H }, <Pg>, [<Xn|SP>, <Xm>, LSL #1]
    ZST3H [Z0.H, Z1.H, Z2.H], P0, (R0)(R0<<1)         // 0060c0e4
    ZST3H [Z10.H, Z11.H, Z12.H], P3, (R12)(R13<<1)    // 8a6dcde4
    ZST3H [Z31.H, Z0.H, Z1.H], P7, (R30)(R30<<1)      // df7fdee4

// ST3W    { <Zt1>.S, <Zt2>.S, <Zt3>.S }, <Pg>, [<Xn|SP>{, #<simm>, MUL VL}]
    ZST3W [Z0.S, Z1.S, Z2.S], P0, -24(R0)       // 00e058e5
    ZST3W [Z10.S, Z11.S, Z12.S], P3, -9(R12)    // 8aed5de5
    ZST3W [Z31.S, Z0.S, Z1.S], P7, 21(R30)      // dfff57e5

// ST3W    { <Zt1>.S, <Zt2>.S, <Zt3>.S }, <Pg>, [<Xn|SP>, <Xm>, LSL #2]
    ZST3W [Z0.S, Z1.S, Z2.S], P0, (R0)(R0<<2)         // 006040e5
    ZST3W [Z10.S, Z11.S, Z12.S], P3, (R12)(R13<<2)    // 8a6d4de5
    ZST3W [Z31.S, Z0.S, Z1.S], P7, (R30)(R30<<2)      // df7f5ee5

// ST4B    { <Zt1>.B, <Zt2>.B, <Zt3>.B, <Zt4>.B }, <Pg>, [<Xn|SP>{, #<simm>, MUL VL}]
    ZST4B [Z0.B, Z1.B, Z2.B, Z3.B], P0, -32(R0)         // 00e078e4
    ZST4B [Z10.B, Z11.B, Z12.B, Z13.B], P3, -12(R12)    // 8aed7de4
    ZST4B [Z31.B, Z0.B, Z1.B, Z2.B], P7, 28(R30)        // dfff77e4

// ST4B    { <Zt1>.B, <Zt2>.B, <Zt3>.B, <Zt4>.B }, <Pg>, [<Xn|SP>, <Xm>]
    ZST4B [Z0.B, Z1.B, Z2.B, Z3.B], P0, (R0)(R0)          // 006060e4
    ZST4B [Z10.B, Z11.B, Z12.B, Z13.B], P3, (R12)(R13)    // 8a6d6de4
    ZST4B [Z31.B, Z0.B, Z1.B, Z2.B], P7, (R30)(R30)       // df7f7ee4

// ST4D    { <Zt1>.D, <Zt2>.D, <Zt3>.D, <Zt4>.D }, <Pg>, [<Xn|SP>{, #<simm>, MUL VL}]
    ZST4D [Z0.D, Z1.D, Z2.D, Z3.D], P0, -32(R0)         // 00e0f8e5
    ZST4D [Z10.D, Z11.D, Z12.D, Z13.D], P3, -12(R12)    // 8aedfde5
    ZST4D [Z31.D, Z0.D, Z1.D, Z2.D], P7, 28(R30)        // dffff7e5

// ST4D    { <Zt1>.D, <Zt2>.D, <Zt3>.D, <Zt4>.D }, <Pg>, [<Xn|SP>, <Xm>, LSL #3]
    ZST4D [Z0.D, Z1.D, Z2.D, Z3.D], P0, (R0)(R0<<3)          // 0060e0e5
    ZST4D [Z10.D, Z11.D, Z12.D, Z13.D], P3, (R12)(R13<<3)    // 8a6dede5
    ZST4D [Z31.D, Z0.D, Z1.D, Z2.D], P7, (R30)(R30<<3)       // df7ffee5

// ST4H    { <Zt1>.H, <Zt2>.H, <Zt3>.H, <Zt4>.H }, <Pg>, [<Xn|SP>{, #<simm>, MUL VL}]
    ZST4H [Z0.H, Z1.H, Z2.H, Z3.H], P0, -32(R0)         // 00e0f8e4
    ZST4H [Z10.H, Z11.H, Z12.H, Z13.H], P3, -12(R12)    // 8aedfde4
    ZST4H [Z31.H, Z0.H, Z1.H, Z2.H], P7, 28(R30)        // dffff7e4

// ST4H    { <Zt1>.H, <Zt2>.H, <Zt3>.H, <Zt4>.H }, <Pg>, [<Xn|SP>, <Xm>, LSL #1]
    ZST4H [Z0.H, Z1.H, Z2.H, Z3.H], P0, (R0)(R0<<1)          // 0060e0e4
    ZST4H [Z10.H, Z11.H, Z12.H, Z13.H], P3, (R12)(R13<<1)    // 8a6dede4
    ZST4H [Z31.H, Z0.H, Z1.H, Z2.H], P7, (R30)(R30<<1)       // df7ffee4

// ST4W    { <Zt1>.S, <Zt2>.S, <Zt3>.S, <Zt4>.S }, <Pg>, [<Xn|SP>{, #<simm>, MUL VL}]
    ZST4W [Z0.S, Z1.S, Z2.S, Z3.S], P0, -32(R0)         // 00e078e5
    ZST4W [Z10.S, Z11.S, Z12.S, Z13.S], P3, -12(R12)    // 8aed7de5
    ZST4W [Z31.S, Z0.S, Z1.S, Z2.S], P7, 28(R30)        // dfff77e5

// ST4W    { <Zt1>.S, <Zt2>.S, <Zt3>.S, <Zt4>.S }, <Pg>, [<Xn|SP>, <Xm>, LSL #2]
    ZST4W [Z0.S, Z1.S, Z2.S, Z3.S], P0, (R0)(R0<<2)          // 006060e5
    ZST4W [Z10.S, Z11.S, Z12.S, Z13.S], P3, (R12)(R13<<2)    // 8a6d6de5
    ZST4W [Z31.S, Z0.S, Z1.S, Z2.S], P7, (R30)(R30<<2)       // df7f7ee5

// STNT1B  { <Zt>.B }, <Pg>, [<Xn|SP>{, #<simm>, MUL VL}]
    ZSTNT1B [Z0.B], P0, -8(R0)      // 00e018e4
    ZSTNT1B [Z10.B], P3, -3(R12)    // 8aed1de4
    ZSTNT1B [Z31.B], P7, 7(R30)     // dfff17e4

// STNT1B  { <Zt>.B }, <Pg>, [<Xn|SP>, <Xm>]
    ZSTNT1B [Z0.B], P0, (R0)(R0)       // 006000e4
    ZSTNT1B [Z10.B], P3, (R12)(R13)    // 8a6d0de4
    ZSTNT1B [Z31.B], P7, (R30)(R30)    // df7f1ee4

// STNT1D  { <Zt>.D }, <Pg>, [<Xn|SP>{, #<simm>, MUL VL}]
    ZSTNT1D [Z0.D], P0, -8(R0)      // 00e098e5
    ZSTNT1D [Z10.D], P3, -3(R12)    // 8aed9de5
    ZSTNT1D [Z31.D], P7, 7(R30)     // dfff97e5

// STNT1D  { <Zt>.D }, <Pg>, [<Xn|SP>, <Xm>, LSL #3]
    ZSTNT1D [Z0.D], P0, (R0)(R0<<3)       // 006080e5
    ZSTNT1D [Z10.D], P3, (R12)(R13<<3)    // 8a6d8de5
    ZSTNT1D [Z31.D], P7, (R30)(R30<<3)    // df7f9ee5

// STNT1H  { <Zt>.H }, <Pg>, [<Xn|SP>{, #<simm>, MUL VL}]
    ZSTNT1H [Z0.H], P0, -8(R0)      // 00e098e4
    ZSTNT1H [Z10.H], P3, -3(R12)    // 8aed9de4
    ZSTNT1H [Z31.H], P7, 7(R30)     // dfff97e4

// STNT1H  { <Zt>.H }, <Pg>, [<Xn|SP>, <Xm>, LSL #1]
    ZSTNT1H [Z0.H], P0, (R0)(R0<<1)       // 006080e4
    ZSTNT1H [Z10.H], P3, (R12)(R13<<1)    // 8a6d8de4
    ZSTNT1H [Z31.H], P7, (R30)(R30<<1)    // df7f9ee4

// STNT1W  { <Zt>.S }, <Pg>, [<Xn|SP>{, #<simm>, MUL VL}]
    ZSTNT1W [Z0.S], P0, -8(R0)      // 00e018e5
    ZSTNT1W [Z10.S], P3, -3(R12)    // 8aed1de5
    ZSTNT1W [Z31.S], P7, 7(R30)     // dfff17e5

// STNT1W  { <Zt>.S }, <Pg>, [<Xn|SP>, <Xm>, LSL #2]
    ZSTNT1W [Z0.S], P0, (R0)(R0<<2)       // 006000e5
    ZSTNT1W [Z10.S], P3, (R12)(R13<<2)    // 8a6d0de5
    ZSTNT1W [Z31.S], P7, (R30)(R30<<2)    // df7f1ee5

// STR     <Pt>, [<Xn|SP>{, #<simm>, MUL VL}]
    PSTR P0, -256(R0)     // 0000a0e5
    PSTR P5, -86(R11)     // 6509b5e5
    PSTR P15, 255(R30)    // cf1f9fe5

// STR     <Zt>, [<Xn|SP>{, #<simm>, MUL VL}]
    ZSTR Z0, -256(R0)     // 0040a0e5
    ZSTR Z10, -86(R11)    // 6a49b5e5
    ZSTR Z31, 255(R30)    // df5f9fe5

// SUB     <Zdn>.<T>, <Pg>/M, <Zdn>.<T>, <Zm>.<T>
    ZSUB P0.M, Z0.B, Z0.B, Z0.B       // 00000104
    ZSUB P3.M, Z10.B, Z12.B, Z10.B    // 8a0d0104
    ZSUB P7.M, Z31.B, Z31.B, Z31.B    // ff1f0104
    ZSUB P0.M, Z0.H, Z0.H, Z0.H       // 00004104
    ZSUB P3.M, Z10.H, Z12.H, Z10.H    // 8a0d4104
    ZSUB P7.M, Z31.H, Z31.H, Z31.H    // ff1f4104
    ZSUB P0.M, Z0.S, Z0.S, Z0.S       // 00008104
    ZSUB P3.M, Z10.S, Z12.S, Z10.S    // 8a0d8104
    ZSUB P7.M, Z31.S, Z31.S, Z31.S    // ff1f8104
    ZSUB P0.M, Z0.D, Z0.D, Z0.D       // 0000c104
    ZSUB P3.M, Z10.D, Z12.D, Z10.D    // 8a0dc104
    ZSUB P7.M, Z31.D, Z31.D, Z31.D    // ff1fc104

// SUB     <Zdn>.<T>, <Zdn>.<T>, #<imm>, <shift>
    ZSUB Z0.B, $0, $0, Z0.B        // 00c02125
    ZSUB Z10.B, $85, $0, Z10.B     // aaca2125
    ZSUB Z31.B, $255, $0, Z31.B    // ffdf2125
    ZSUB Z0.H, $0, $8, Z0.H        // 00e06125
    ZSUB Z10.H, $85, $8, Z10.H     // aaea6125
    ZSUB Z31.H, $255, $0, Z31.H    // ffdf6125
    ZSUB Z0.S, $0, $8, Z0.S        // 00e0a125
    ZSUB Z10.S, $85, $8, Z10.S     // aaeaa125
    ZSUB Z31.S, $255, $0, Z31.S    // ffdfa125
    ZSUB Z0.D, $0, $8, Z0.D        // 00e0e125
    ZSUB Z10.D, $85, $8, Z10.D     // aaeae125
    ZSUB Z31.D, $255, $0, Z31.D    // ffdfe125

// SUB     <Zd>.<T>, <Zn>.<T>, <Zm>.<T>
    ZSUB Z0.B, Z0.B, Z0.B       // 00042004
    ZSUB Z11.B, Z12.B, Z10.B    // 6a052c04
    ZSUB Z31.B, Z31.B, Z31.B    // ff073f04
    ZSUB Z0.H, Z0.H, Z0.H       // 00046004
    ZSUB Z11.H, Z12.H, Z10.H    // 6a056c04
    ZSUB Z31.H, Z31.H, Z31.H    // ff077f04
    ZSUB Z0.S, Z0.S, Z0.S       // 0004a004
    ZSUB Z11.S, Z12.S, Z10.S    // 6a05ac04
    ZSUB Z31.S, Z31.S, Z31.S    // ff07bf04
    ZSUB Z0.D, Z0.D, Z0.D       // 0004e004
    ZSUB Z11.D, Z12.D, Z10.D    // 6a05ec04
    ZSUB Z31.D, Z31.D, Z31.D    // ff07ff04

// SUBR    <Zdn>.<T>, <Pg>/M, <Zdn>.<T>, <Zm>.<T>
    ZSUBR P0.M, Z0.B, Z0.B, Z0.B       // 00000304
    ZSUBR P3.M, Z10.B, Z12.B, Z10.B    // 8a0d0304
    ZSUBR P7.M, Z31.B, Z31.B, Z31.B    // ff1f0304
    ZSUBR P0.M, Z0.H, Z0.H, Z0.H       // 00004304
    ZSUBR P3.M, Z10.H, Z12.H, Z10.H    // 8a0d4304
    ZSUBR P7.M, Z31.H, Z31.H, Z31.H    // ff1f4304
    ZSUBR P0.M, Z0.S, Z0.S, Z0.S       // 00008304
    ZSUBR P3.M, Z10.S, Z12.S, Z10.S    // 8a0d8304
    ZSUBR P7.M, Z31.S, Z31.S, Z31.S    // ff1f8304
    ZSUBR P0.M, Z0.D, Z0.D, Z0.D       // 0000c304
    ZSUBR P3.M, Z10.D, Z12.D, Z10.D    // 8a0dc304
    ZSUBR P7.M, Z31.D, Z31.D, Z31.D    // ff1fc304

// SUBR    <Zdn>.<T>, <Zdn>.<T>, #<imm>, <shift>
    ZSUBR Z0.B, $0, $0, Z0.B        // 00c02325
    ZSUBR Z10.B, $85, $0, Z10.B     // aaca2325
    ZSUBR Z31.B, $255, $0, Z31.B    // ffdf2325
    ZSUBR Z0.H, $0, $8, Z0.H        // 00e06325
    ZSUBR Z10.H, $85, $8, Z10.H     // aaea6325
    ZSUBR Z31.H, $255, $0, Z31.H    // ffdf6325
    ZSUBR Z0.S, $0, $8, Z0.S        // 00e0a325
    ZSUBR Z10.S, $85, $8, Z10.S     // aaeaa325
    ZSUBR Z31.S, $255, $0, Z31.S    // ffdfa325
    ZSUBR Z0.D, $0, $8, Z0.D        // 00e0e325
    ZSUBR Z10.D, $85, $8, Z10.D     // aaeae325
    ZSUBR Z31.D, $255, $0, Z31.D    // ffdfe325

// SUDOT   <Zda>.S, <Zn>.B, <Zm>.B[<imm>]
    ZSUDOT Z0.B, Z0.B[0], Z0.S      // 001ca044
    ZSUDOT Z11.B, Z4.B[1], Z10.S    // 6a1dac44
    ZSUDOT Z31.B, Z7.B[3], Z31.S    // ff1fbf44

// SUNPKHI <Zd>.<T>, <Zn>.<Tb>
    ZSUNPKHI Z0.B, Z0.H      // 00387105
    ZSUNPKHI Z11.B, Z10.H    // 6a397105
    ZSUNPKHI Z31.B, Z31.H    // ff3b7105
    ZSUNPKHI Z0.H, Z0.S      // 0038b105
    ZSUNPKHI Z11.H, Z10.S    // 6a39b105
    ZSUNPKHI Z31.H, Z31.S    // ff3bb105
    ZSUNPKHI Z0.S, Z0.D      // 0038f105
    ZSUNPKHI Z11.S, Z10.D    // 6a39f105
    ZSUNPKHI Z31.S, Z31.D    // ff3bf105

// SUNPKLO <Zd>.<T>, <Zn>.<Tb>
    ZSUNPKLO Z0.B, Z0.H      // 00387005
    ZSUNPKLO Z11.B, Z10.H    // 6a397005
    ZSUNPKLO Z31.B, Z31.H    // ff3b7005
    ZSUNPKLO Z0.H, Z0.S      // 0038b005
    ZSUNPKLO Z11.H, Z10.S    // 6a39b005
    ZSUNPKLO Z31.H, Z31.S    // ff3bb005
    ZSUNPKLO Z0.S, Z0.D      // 0038f005
    ZSUNPKLO Z11.S, Z10.D    // 6a39f005
    ZSUNPKLO Z31.S, Z31.D    // ff3bf005

// SXTB    <Zd>.<T>, <Pg>/M, <Zn>.<T>
    ZSXTB P0.M, Z0.H, Z0.H      // 00a05004
    ZSXTB P3.M, Z12.H, Z10.H    // 8aad5004
    ZSXTB P7.M, Z31.H, Z31.H    // ffbf5004
    ZSXTB P0.M, Z0.S, Z0.S      // 00a09004
    ZSXTB P3.M, Z12.S, Z10.S    // 8aad9004
    ZSXTB P7.M, Z31.S, Z31.S    // ffbf9004
    ZSXTB P0.M, Z0.D, Z0.D      // 00a0d004
    ZSXTB P3.M, Z12.D, Z10.D    // 8aadd004
    ZSXTB P7.M, Z31.D, Z31.D    // ffbfd004

// SXTH    <Zd>.<T>, <Pg>/M, <Zn>.<T>
    ZSXTH P0.M, Z0.S, Z0.S      // 00a09204
    ZSXTH P3.M, Z12.S, Z10.S    // 8aad9204
    ZSXTH P7.M, Z31.S, Z31.S    // ffbf9204
    ZSXTH P0.M, Z0.D, Z0.D      // 00a0d204
    ZSXTH P3.M, Z12.D, Z10.D    // 8aadd204
    ZSXTH P7.M, Z31.D, Z31.D    // ffbfd204

// SXTW    <Zd>.D, <Pg>/M, <Zn>.D
    ZSXTW P0.M, Z0.D, Z0.D      // 00a0d404
    ZSXTW P3.M, Z12.D, Z10.D    // 8aadd404
    ZSXTW P7.M, Z31.D, Z31.D    // ffbfd404

// TBL     <Zd>.<T>, { <Zn>.<T> }, <Zm>.<T>
    ZTBL [Z0.B], Z0.B, Z0.B       // 00302005
    ZTBL [Z11.B], Z12.B, Z10.B    // 6a312c05
    ZTBL [Z31.B], Z31.B, Z31.B    // ff333f05
    ZTBL [Z0.H], Z0.H, Z0.H       // 00306005
    ZTBL [Z11.H], Z12.H, Z10.H    // 6a316c05
    ZTBL [Z31.H], Z31.H, Z31.H    // ff337f05
    ZTBL [Z0.S], Z0.S, Z0.S       // 0030a005
    ZTBL [Z11.S], Z12.S, Z10.S    // 6a31ac05
    ZTBL [Z31.S], Z31.S, Z31.S    // ff33bf05
    ZTBL [Z0.D], Z0.D, Z0.D       // 0030e005
    ZTBL [Z11.D], Z12.D, Z10.D    // 6a31ec05
    ZTBL [Z31.D], Z31.D, Z31.D    // ff33ff05

// TRN1    <Pd>.<T>, <Pn>.<T>, <Pm>.<T>
    PTRN1 P0.B, P0.B, P0.B       // 00502005
    PTRN1 P6.B, P7.B, P5.B       // c5502705
    PTRN1 P15.B, P15.B, P15.B    // ef512f05
    PTRN1 P0.H, P0.H, P0.H       // 00506005
    PTRN1 P6.H, P7.H, P5.H       // c5506705
    PTRN1 P15.H, P15.H, P15.H    // ef516f05
    PTRN1 P0.S, P0.S, P0.S       // 0050a005
    PTRN1 P6.S, P7.S, P5.S       // c550a705
    PTRN1 P15.S, P15.S, P15.S    // ef51af05
    PTRN1 P0.D, P0.D, P0.D       // 0050e005
    PTRN1 P6.D, P7.D, P5.D       // c550e705
    PTRN1 P15.D, P15.D, P15.D    // ef51ef05

// TRN1    <Zd>.Q, <Zn>.Q, <Zm>.Q
    ZTRN1 Z0.Q, Z0.Q, Z0.Q       // 0018a005
    ZTRN1 Z11.Q, Z12.Q, Z10.Q    // 6a19ac05
    ZTRN1 Z31.Q, Z31.Q, Z31.Q    // ff1bbf05

// TRN1    <Zd>.<T>, <Zn>.<T>, <Zm>.<T>
    ZTRN1 Z0.B, Z0.B, Z0.B       // 00702005
    ZTRN1 Z11.B, Z12.B, Z10.B    // 6a712c05
    ZTRN1 Z31.B, Z31.B, Z31.B    // ff733f05
    ZTRN1 Z0.H, Z0.H, Z0.H       // 00706005
    ZTRN1 Z11.H, Z12.H, Z10.H    // 6a716c05
    ZTRN1 Z31.H, Z31.H, Z31.H    // ff737f05
    ZTRN1 Z0.S, Z0.S, Z0.S       // 0070a005
    ZTRN1 Z11.S, Z12.S, Z10.S    // 6a71ac05
    ZTRN1 Z31.S, Z31.S, Z31.S    // ff73bf05
    ZTRN1 Z0.D, Z0.D, Z0.D       // 0070e005
    ZTRN1 Z11.D, Z12.D, Z10.D    // 6a71ec05
    ZTRN1 Z31.D, Z31.D, Z31.D    // ff73ff05

// TRN2    <Pd>.<T>, <Pn>.<T>, <Pm>.<T>
    PTRN2 P0.B, P0.B, P0.B       // 00542005
    PTRN2 P6.B, P7.B, P5.B       // c5542705
    PTRN2 P15.B, P15.B, P15.B    // ef552f05
    PTRN2 P0.H, P0.H, P0.H       // 00546005
    PTRN2 P6.H, P7.H, P5.H       // c5546705
    PTRN2 P15.H, P15.H, P15.H    // ef556f05
    PTRN2 P0.S, P0.S, P0.S       // 0054a005
    PTRN2 P6.S, P7.S, P5.S       // c554a705
    PTRN2 P15.S, P15.S, P15.S    // ef55af05
    PTRN2 P0.D, P0.D, P0.D       // 0054e005
    PTRN2 P6.D, P7.D, P5.D       // c554e705
    PTRN2 P15.D, P15.D, P15.D    // ef55ef05

// TRN2    <Zd>.Q, <Zn>.Q, <Zm>.Q
    ZTRN2 Z0.Q, Z0.Q, Z0.Q       // 001ca005
    ZTRN2 Z11.Q, Z12.Q, Z10.Q    // 6a1dac05
    ZTRN2 Z31.Q, Z31.Q, Z31.Q    // ff1fbf05

// TRN2    <Zd>.<T>, <Zn>.<T>, <Zm>.<T>
    ZTRN2 Z0.B, Z0.B, Z0.B       // 00742005
    ZTRN2 Z11.B, Z12.B, Z10.B    // 6a752c05
    ZTRN2 Z31.B, Z31.B, Z31.B    // ff773f05
    ZTRN2 Z0.H, Z0.H, Z0.H       // 00746005
    ZTRN2 Z11.H, Z12.H, Z10.H    // 6a756c05
    ZTRN2 Z31.H, Z31.H, Z31.H    // ff777f05
    ZTRN2 Z0.S, Z0.S, Z0.S       // 0074a005
    ZTRN2 Z11.S, Z12.S, Z10.S    // 6a75ac05
    ZTRN2 Z31.S, Z31.S, Z31.S    // ff77bf05
    ZTRN2 Z0.D, Z0.D, Z0.D       // 0074e005
    ZTRN2 Z11.D, Z12.D, Z10.D    // 6a75ec05
    ZTRN2 Z31.D, Z31.D, Z31.D    // ff77ff05

// UABD    <Zdn>.<T>, <Pg>/M, <Zdn>.<T>, <Zm>.<T>
    ZUABD P0.M, Z0.B, Z0.B, Z0.B       // 00000d04
    ZUABD P3.M, Z10.B, Z12.B, Z10.B    // 8a0d0d04
    ZUABD P7.M, Z31.B, Z31.B, Z31.B    // ff1f0d04
    ZUABD P0.M, Z0.H, Z0.H, Z0.H       // 00004d04
    ZUABD P3.M, Z10.H, Z12.H, Z10.H    // 8a0d4d04
    ZUABD P7.M, Z31.H, Z31.H, Z31.H    // ff1f4d04
    ZUABD P0.M, Z0.S, Z0.S, Z0.S       // 00008d04
    ZUABD P3.M, Z10.S, Z12.S, Z10.S    // 8a0d8d04
    ZUABD P7.M, Z31.S, Z31.S, Z31.S    // ff1f8d04
    ZUABD P0.M, Z0.D, Z0.D, Z0.D       // 0000cd04
    ZUABD P3.M, Z10.D, Z12.D, Z10.D    // 8a0dcd04
    ZUABD P7.M, Z31.D, Z31.D, Z31.D    // ff1fcd04

// UADDV   <Dd>, <Pg>, <Zn>.<T>
    ZUADDV P0, Z0.B, V0      // 00200104
    ZUADDV P3, Z12.B, V10    // 8a2d0104
    ZUADDV P7, Z31.B, V31    // ff3f0104
    ZUADDV P0, Z0.H, V0      // 00204104
    ZUADDV P3, Z12.H, V10    // 8a2d4104
    ZUADDV P7, Z31.H, V31    // ff3f4104
    ZUADDV P0, Z0.S, V0      // 00208104
    ZUADDV P3, Z12.S, V10    // 8a2d8104
    ZUADDV P7, Z31.S, V31    // ff3f8104
    ZUADDV P0, Z0.D, V0      // 0020c104
    ZUADDV P3, Z12.D, V10    // 8a2dc104
    ZUADDV P7, Z31.D, V31    // ff3fc104

// UCVTF   <Zd>.H, <Pg>/M, <Zn>.H
    ZUCVTF P0.M, Z0.H, Z0.H      // 00a05365
    ZUCVTF P3.M, Z12.H, Z10.H    // 8aad5365
    ZUCVTF P7.M, Z31.H, Z31.H    // ffbf5365

// UCVTF   <Zd>.D, <Pg>/M, <Zn>.S
    ZUCVTF P0.M, Z0.S, Z0.D      // 00a0d165
    ZUCVTF P3.M, Z12.S, Z10.D    // 8aadd165
    ZUCVTF P7.M, Z31.S, Z31.D    // ffbfd165

// UCVTF   <Zd>.H, <Pg>/M, <Zn>.S
    ZUCVTF P0.M, Z0.S, Z0.H      // 00a05565
    ZUCVTF P3.M, Z12.S, Z10.H    // 8aad5565
    ZUCVTF P7.M, Z31.S, Z31.H    // ffbf5565

// UCVTF   <Zd>.S, <Pg>/M, <Zn>.S
    ZUCVTF P0.M, Z0.S, Z0.S      // 00a09565
    ZUCVTF P3.M, Z12.S, Z10.S    // 8aad9565
    ZUCVTF P7.M, Z31.S, Z31.S    // ffbf9565

// UCVTF   <Zd>.D, <Pg>/M, <Zn>.D
    ZUCVTF P0.M, Z0.D, Z0.D      // 00a0d765
    ZUCVTF P3.M, Z12.D, Z10.D    // 8aadd765
    ZUCVTF P7.M, Z31.D, Z31.D    // ffbfd765

// UCVTF   <Zd>.H, <Pg>/M, <Zn>.D
    ZUCVTF P0.M, Z0.D, Z0.H      // 00a05765
    ZUCVTF P3.M, Z12.D, Z10.H    // 8aad5765
    ZUCVTF P7.M, Z31.D, Z31.H    // ffbf5765

// UCVTF   <Zd>.S, <Pg>/M, <Zn>.D
    ZUCVTF P0.M, Z0.D, Z0.S      // 00a0d565
    ZUCVTF P3.M, Z12.D, Z10.S    // 8aadd565
    ZUCVTF P7.M, Z31.D, Z31.S    // ffbfd565

// UDIV    <Zdn>.<T>, <Pg>/M, <Zdn>.<T>, <Zm>.<T>
    ZUDIV P0.M, Z0.S, Z0.S, Z0.S       // 00009504
    ZUDIV P3.M, Z10.S, Z12.S, Z10.S    // 8a0d9504
    ZUDIV P7.M, Z31.S, Z31.S, Z31.S    // ff1f9504
    ZUDIV P0.M, Z0.D, Z0.D, Z0.D       // 0000d504
    ZUDIV P3.M, Z10.D, Z12.D, Z10.D    // 8a0dd504
    ZUDIV P7.M, Z31.D, Z31.D, Z31.D    // ff1fd504

// UDIVR   <Zdn>.<T>, <Pg>/M, <Zdn>.<T>, <Zm>.<T>
    ZUDIVR P0.M, Z0.S, Z0.S, Z0.S       // 00009704
    ZUDIVR P3.M, Z10.S, Z12.S, Z10.S    // 8a0d9704
    ZUDIVR P7.M, Z31.S, Z31.S, Z31.S    // ff1f9704
    ZUDIVR P0.M, Z0.D, Z0.D, Z0.D       // 0000d704
    ZUDIVR P3.M, Z10.D, Z12.D, Z10.D    // 8a0dd704
    ZUDIVR P7.M, Z31.D, Z31.D, Z31.D    // ff1fd704

// UDOT    <Zda>.<T>, <Zn>.<Tb>, <Zm>.<Tb>
    ZUDOT Z0.B, Z0.B, Z0.S       // 00048044
    ZUDOT Z11.B, Z12.B, Z10.S    // 6a058c44
    ZUDOT Z31.B, Z31.B, Z31.S    // ff079f44
    ZUDOT Z0.H, Z0.H, Z0.D       // 0004c044
    ZUDOT Z11.H, Z12.H, Z10.D    // 6a05cc44
    ZUDOT Z31.H, Z31.H, Z31.D    // ff07df44

// UDOT    <Zda>.D, <Zn>.H, <Zm>.H[<imm>]
    ZUDOT Z0.H, Z0.H[0], Z0.D       // 0004e044
    ZUDOT Z11.H, Z7.H[0], Z10.D     // 6a05e744
    ZUDOT Z31.H, Z15.H[1], Z31.D    // ff07ff44

// UDOT    <Zda>.S, <Zn>.B, <Zm>.B[<imm>]
    ZUDOT Z0.B, Z0.B[0], Z0.S      // 0004a044
    ZUDOT Z11.B, Z4.B[1], Z10.S    // 6a05ac44
    ZUDOT Z31.B, Z7.B[3], Z31.S    // ff07bf44

// UMAX    <Zdn>.<T>, <Pg>/M, <Zdn>.<T>, <Zm>.<T>
    ZUMAX P0.M, Z0.B, Z0.B, Z0.B       // 00000904
    ZUMAX P3.M, Z10.B, Z12.B, Z10.B    // 8a0d0904
    ZUMAX P7.M, Z31.B, Z31.B, Z31.B    // ff1f0904
    ZUMAX P0.M, Z0.H, Z0.H, Z0.H       // 00004904
    ZUMAX P3.M, Z10.H, Z12.H, Z10.H    // 8a0d4904
    ZUMAX P7.M, Z31.H, Z31.H, Z31.H    // ff1f4904
    ZUMAX P0.M, Z0.S, Z0.S, Z0.S       // 00008904
    ZUMAX P3.M, Z10.S, Z12.S, Z10.S    // 8a0d8904
    ZUMAX P7.M, Z31.S, Z31.S, Z31.S    // ff1f8904
    ZUMAX P0.M, Z0.D, Z0.D, Z0.D       // 0000c904
    ZUMAX P3.M, Z10.D, Z12.D, Z10.D    // 8a0dc904
    ZUMAX P7.M, Z31.D, Z31.D, Z31.D    // ff1fc904

// UMAX    <Zdn>.<T>, <Zdn>.<T>, #<imm>
    ZUMAX Z0.B, $0, Z0.B        // 00c02925
    ZUMAX Z10.B, $85, Z10.B     // aaca2925
    ZUMAX Z31.B, $255, Z31.B    // ffdf2925
    ZUMAX Z0.H, $0, Z0.H        // 00c06925
    ZUMAX Z10.H, $85, Z10.H     // aaca6925
    ZUMAX Z31.H, $255, Z31.H    // ffdf6925
    ZUMAX Z0.S, $0, Z0.S        // 00c0a925
    ZUMAX Z10.S, $85, Z10.S     // aacaa925
    ZUMAX Z31.S, $255, Z31.S    // ffdfa925
    ZUMAX Z0.D, $0, Z0.D        // 00c0e925
    ZUMAX Z10.D, $85, Z10.D     // aacae925
    ZUMAX Z31.D, $255, Z31.D    // ffdfe925

// UMAXV   <V><d>, <Pg>, <Zn>.<T>
    ZUMAXV P0, Z0.B, V0      // 00200904
    ZUMAXV P3, Z12.B, V10    // 8a2d0904
    ZUMAXV P7, Z31.B, V31    // ff3f0904
    ZUMAXV P0, Z0.H, V0      // 00204904
    ZUMAXV P3, Z12.H, V10    // 8a2d4904
    ZUMAXV P7, Z31.H, V31    // ff3f4904
    ZUMAXV P0, Z0.S, V0      // 00208904
    ZUMAXV P3, Z12.S, V10    // 8a2d8904
    ZUMAXV P7, Z31.S, V31    // ff3f8904
    ZUMAXV P0, Z0.D, V0      // 0020c904
    ZUMAXV P3, Z12.D, V10    // 8a2dc904
    ZUMAXV P7, Z31.D, V31    // ff3fc904

// UMIN    <Zdn>.<T>, <Pg>/M, <Zdn>.<T>, <Zm>.<T>
    ZUMIN P0.M, Z0.B, Z0.B, Z0.B       // 00000b04
    ZUMIN P3.M, Z10.B, Z12.B, Z10.B    // 8a0d0b04
    ZUMIN P7.M, Z31.B, Z31.B, Z31.B    // ff1f0b04
    ZUMIN P0.M, Z0.H, Z0.H, Z0.H       // 00004b04
    ZUMIN P3.M, Z10.H, Z12.H, Z10.H    // 8a0d4b04
    ZUMIN P7.M, Z31.H, Z31.H, Z31.H    // ff1f4b04
    ZUMIN P0.M, Z0.S, Z0.S, Z0.S       // 00008b04
    ZUMIN P3.M, Z10.S, Z12.S, Z10.S    // 8a0d8b04
    ZUMIN P7.M, Z31.S, Z31.S, Z31.S    // ff1f8b04
    ZUMIN P0.M, Z0.D, Z0.D, Z0.D       // 0000cb04
    ZUMIN P3.M, Z10.D, Z12.D, Z10.D    // 8a0dcb04
    ZUMIN P7.M, Z31.D, Z31.D, Z31.D    // ff1fcb04

// UMIN    <Zdn>.<T>, <Zdn>.<T>, #<imm>
    ZUMIN Z0.B, $0, Z0.B        // 00c02b25
    ZUMIN Z10.B, $85, Z10.B     // aaca2b25
    ZUMIN Z31.B, $255, Z31.B    // ffdf2b25
    ZUMIN Z0.H, $0, Z0.H        // 00c06b25
    ZUMIN Z10.H, $85, Z10.H     // aaca6b25
    ZUMIN Z31.H, $255, Z31.H    // ffdf6b25
    ZUMIN Z0.S, $0, Z0.S        // 00c0ab25
    ZUMIN Z10.S, $85, Z10.S     // aacaab25
    ZUMIN Z31.S, $255, Z31.S    // ffdfab25
    ZUMIN Z0.D, $0, Z0.D        // 00c0eb25
    ZUMIN Z10.D, $85, Z10.D     // aacaeb25
    ZUMIN Z31.D, $255, Z31.D    // ffdfeb25

// UMINV   <V><d>, <Pg>, <Zn>.<T>
    ZUMINV P0, Z0.B, V0      // 00200b04
    ZUMINV P3, Z12.B, V10    // 8a2d0b04
    ZUMINV P7, Z31.B, V31    // ff3f0b04
    ZUMINV P0, Z0.H, V0      // 00204b04
    ZUMINV P3, Z12.H, V10    // 8a2d4b04
    ZUMINV P7, Z31.H, V31    // ff3f4b04
    ZUMINV P0, Z0.S, V0      // 00208b04
    ZUMINV P3, Z12.S, V10    // 8a2d8b04
    ZUMINV P7, Z31.S, V31    // ff3f8b04
    ZUMINV P0, Z0.D, V0      // 0020cb04
    ZUMINV P3, Z12.D, V10    // 8a2dcb04
    ZUMINV P7, Z31.D, V31    // ff3fcb04

// UMMLA   <Zda>.S, <Zn>.B, <Zm>.B
    ZUMMLA Z0.B, Z0.B, Z0.S       // 0098c045
    ZUMMLA Z11.B, Z12.B, Z10.S    // 6a99cc45
    ZUMMLA Z31.B, Z31.B, Z31.S    // ff9bdf45

// UMULH   <Zdn>.<T>, <Pg>/M, <Zdn>.<T>, <Zm>.<T>
    ZUMULH P0.M, Z0.B, Z0.B, Z0.B       // 00001304
    ZUMULH P3.M, Z10.B, Z12.B, Z10.B    // 8a0d1304
    ZUMULH P7.M, Z31.B, Z31.B, Z31.B    // ff1f1304
    ZUMULH P0.M, Z0.H, Z0.H, Z0.H       // 00005304
    ZUMULH P3.M, Z10.H, Z12.H, Z10.H    // 8a0d5304
    ZUMULH P7.M, Z31.H, Z31.H, Z31.H    // ff1f5304
    ZUMULH P0.M, Z0.S, Z0.S, Z0.S       // 00009304
    ZUMULH P3.M, Z10.S, Z12.S, Z10.S    // 8a0d9304
    ZUMULH P7.M, Z31.S, Z31.S, Z31.S    // ff1f9304
    ZUMULH P0.M, Z0.D, Z0.D, Z0.D       // 0000d304
    ZUMULH P3.M, Z10.D, Z12.D, Z10.D    // 8a0dd304
    ZUMULH P7.M, Z31.D, Z31.D, Z31.D    // ff1fd304

// UQADD   <Zdn>.<T>, <Zdn>.<T>, #<imm>, <shift>
    ZUQADD Z0.B, $0, $0, Z0.B        // 00c02525
    ZUQADD Z10.B, $85, $0, Z10.B     // aaca2525
    ZUQADD Z31.B, $255, $0, Z31.B    // ffdf2525
    ZUQADD Z0.H, $0, $8, Z0.H        // 00e06525
    ZUQADD Z10.H, $85, $8, Z10.H     // aaea6525
    ZUQADD Z31.H, $255, $0, Z31.H    // ffdf6525
    ZUQADD Z0.S, $0, $8, Z0.S        // 00e0a525
    ZUQADD Z10.S, $85, $8, Z10.S     // aaeaa525
    ZUQADD Z31.S, $255, $0, Z31.S    // ffdfa525
    ZUQADD Z0.D, $0, $8, Z0.D        // 00e0e525
    ZUQADD Z10.D, $85, $8, Z10.D     // aaeae525
    ZUQADD Z31.D, $255, $0, Z31.D    // ffdfe525

// UQADD   <Zd>.<T>, <Zn>.<T>, <Zm>.<T>
    ZUQADD Z0.B, Z0.B, Z0.B       // 00142004
    ZUQADD Z11.B, Z12.B, Z10.B    // 6a152c04
    ZUQADD Z31.B, Z31.B, Z31.B    // ff173f04
    ZUQADD Z0.H, Z0.H, Z0.H       // 00146004
    ZUQADD Z11.H, Z12.H, Z10.H    // 6a156c04
    ZUQADD Z31.H, Z31.H, Z31.H    // ff177f04
    ZUQADD Z0.S, Z0.S, Z0.S       // 0014a004
    ZUQADD Z11.S, Z12.S, Z10.S    // 6a15ac04
    ZUQADD Z31.S, Z31.S, Z31.S    // ff17bf04
    ZUQADD Z0.D, Z0.D, Z0.D       // 0014e004
    ZUQADD Z11.D, Z12.D, Z10.D    // 6a15ec04
    ZUQADD Z31.D, Z31.D, Z31.D    // ff17ff04

// UQDECB  <Wdn>{, <pattern>{, MUL #<imm>}}
    ZUQDECBW POW2, $1, R0      // 00fc2004
    ZUQDECBW VL1, $1, R1       // 21fc2004
    ZUQDECBW VL2, $2, R2       // 42fc2104
    ZUQDECBW VL3, $2, R3       // 63fc2104
    ZUQDECBW VL4, $3, R4       // 84fc2204
    ZUQDECBW VL5, $3, R5       // a5fc2204
    ZUQDECBW VL6, $4, R6       // c6fc2304
    ZUQDECBW VL7, $4, R7       // e7fc2304
    ZUQDECBW VL8, $5, R8       // 08fd2404
    ZUQDECBW VL16, $5, R8      // 28fd2404
    ZUQDECBW VL32, $6, R9      // 49fd2504
    ZUQDECBW VL64, $6, R10     // 6afd2504
    ZUQDECBW VL128, $7, R11    // 8bfd2604
    ZUQDECBW VL256, $7, R12    // acfd2604
    ZUQDECBW $14, $8, R13      // cdfd2704
    ZUQDECBW $15, $8, R14      // eefd2704
    ZUQDECBW $16, $9, R15      // 0ffe2804
    ZUQDECBW $17, $9, R16      // 30fe2804
    ZUQDECBW $18, $9, R17      // 51fe2804
    ZUQDECBW $19, $10, R17     // 71fe2904
    ZUQDECBW $20, $10, R19     // 93fe2904
    ZUQDECBW $21, $11, R20     // b4fe2a04
    ZUQDECBW $22, $11, R21     // d5fe2a04
    ZUQDECBW $23, $12, R22     // f6fe2b04
    ZUQDECBW $24, $12, R22     // 16ff2b04
    ZUQDECBW $25, $13, R23     // 37ff2c04
    ZUQDECBW $26, $13, R24     // 58ff2c04
    ZUQDECBW $27, $14, R25     // 79ff2d04
    ZUQDECBW $28, $14, R26     // 9aff2d04
    ZUQDECBW MUL4, $15, R27    // bbff2e04
    ZUQDECBW MUL3, $15, R27    // dbff2e04
    ZUQDECBW ALL, $16, R30     // feff2f04

// UQDECB  <Xdn>{, <pattern>{, MUL #<imm>}}
    ZUQDECB POW2, $1, R0      // 00fc3004
    ZUQDECB VL1, $1, R1       // 21fc3004
    ZUQDECB VL2, $2, R2       // 42fc3104
    ZUQDECB VL3, $2, R3       // 63fc3104
    ZUQDECB VL4, $3, R4       // 84fc3204
    ZUQDECB VL5, $3, R5       // a5fc3204
    ZUQDECB VL6, $4, R6       // c6fc3304
    ZUQDECB VL7, $4, R7       // e7fc3304
    ZUQDECB VL8, $5, R8       // 08fd3404
    ZUQDECB VL16, $5, R8      // 28fd3404
    ZUQDECB VL32, $6, R9      // 49fd3504
    ZUQDECB VL64, $6, R10     // 6afd3504
    ZUQDECB VL128, $7, R11    // 8bfd3604
    ZUQDECB VL256, $7, R12    // acfd3604
    ZUQDECB $14, $8, R13      // cdfd3704
    ZUQDECB $15, $8, R14      // eefd3704
    ZUQDECB $16, $9, R15      // 0ffe3804
    ZUQDECB $17, $9, R16      // 30fe3804
    ZUQDECB $18, $9, R17      // 51fe3804
    ZUQDECB $19, $10, R17     // 71fe3904
    ZUQDECB $20, $10, R19     // 93fe3904
    ZUQDECB $21, $11, R20     // b4fe3a04
    ZUQDECB $22, $11, R21     // d5fe3a04
    ZUQDECB $23, $12, R22     // f6fe3b04
    ZUQDECB $24, $12, R22     // 16ff3b04
    ZUQDECB $25, $13, R23     // 37ff3c04
    ZUQDECB $26, $13, R24     // 58ff3c04
    ZUQDECB $27, $14, R25     // 79ff3d04
    ZUQDECB $28, $14, R26     // 9aff3d04
    ZUQDECB MUL4, $15, R27    // bbff3e04
    ZUQDECB MUL3, $15, R27    // dbff3e04
    ZUQDECB ALL, $16, R30     // feff3f04

// UQDECD  <Wdn>{, <pattern>{, MUL #<imm>}}
    ZUQDECDW POW2, $1, R0      // 00fce004
    ZUQDECDW VL1, $1, R1       // 21fce004
    ZUQDECDW VL2, $2, R2       // 42fce104
    ZUQDECDW VL3, $2, R3       // 63fce104
    ZUQDECDW VL4, $3, R4       // 84fce204
    ZUQDECDW VL5, $3, R5       // a5fce204
    ZUQDECDW VL6, $4, R6       // c6fce304
    ZUQDECDW VL7, $4, R7       // e7fce304
    ZUQDECDW VL8, $5, R8       // 08fde404
    ZUQDECDW VL16, $5, R8      // 28fde404
    ZUQDECDW VL32, $6, R9      // 49fde504
    ZUQDECDW VL64, $6, R10     // 6afde504
    ZUQDECDW VL128, $7, R11    // 8bfde604
    ZUQDECDW VL256, $7, R12    // acfde604
    ZUQDECDW $14, $8, R13      // cdfde704
    ZUQDECDW $15, $8, R14      // eefde704
    ZUQDECDW $16, $9, R15      // 0ffee804
    ZUQDECDW $17, $9, R16      // 30fee804
    ZUQDECDW $18, $9, R17      // 51fee804
    ZUQDECDW $19, $10, R17     // 71fee904
    ZUQDECDW $20, $10, R19     // 93fee904
    ZUQDECDW $21, $11, R20     // b4feea04
    ZUQDECDW $22, $11, R21     // d5feea04
    ZUQDECDW $23, $12, R22     // f6feeb04
    ZUQDECDW $24, $12, R22     // 16ffeb04
    ZUQDECDW $25, $13, R23     // 37ffec04
    ZUQDECDW $26, $13, R24     // 58ffec04
    ZUQDECDW $27, $14, R25     // 79ffed04
    ZUQDECDW $28, $14, R26     // 9affed04
    ZUQDECDW MUL4, $15, R27    // bbffee04
    ZUQDECDW MUL3, $15, R27    // dbffee04
    ZUQDECDW ALL, $16, R30     // feffef04

// UQDECD  <Xdn>{, <pattern>{, MUL #<imm>}}
    ZUQDECD POW2, $1, R0      // 00fcf004
    ZUQDECD VL1, $1, R1       // 21fcf004
    ZUQDECD VL2, $2, R2       // 42fcf104
    ZUQDECD VL3, $2, R3       // 63fcf104
    ZUQDECD VL4, $3, R4       // 84fcf204
    ZUQDECD VL5, $3, R5       // a5fcf204
    ZUQDECD VL6, $4, R6       // c6fcf304
    ZUQDECD VL7, $4, R7       // e7fcf304
    ZUQDECD VL8, $5, R8       // 08fdf404
    ZUQDECD VL16, $5, R8      // 28fdf404
    ZUQDECD VL32, $6, R9      // 49fdf504
    ZUQDECD VL64, $6, R10     // 6afdf504
    ZUQDECD VL128, $7, R11    // 8bfdf604
    ZUQDECD VL256, $7, R12    // acfdf604
    ZUQDECD $14, $8, R13      // cdfdf704
    ZUQDECD $15, $8, R14      // eefdf704
    ZUQDECD $16, $9, R15      // 0ffef804
    ZUQDECD $17, $9, R16      // 30fef804
    ZUQDECD $18, $9, R17      // 51fef804
    ZUQDECD $19, $10, R17     // 71fef904
    ZUQDECD $20, $10, R19     // 93fef904
    ZUQDECD $21, $11, R20     // b4fefa04
    ZUQDECD $22, $11, R21     // d5fefa04
    ZUQDECD $23, $12, R22     // f6fefb04
    ZUQDECD $24, $12, R22     // 16fffb04
    ZUQDECD $25, $13, R23     // 37fffc04
    ZUQDECD $26, $13, R24     // 58fffc04
    ZUQDECD $27, $14, R25     // 79fffd04
    ZUQDECD $28, $14, R26     // 9afffd04
    ZUQDECD MUL4, $15, R27    // bbfffe04
    ZUQDECD MUL3, $15, R27    // dbfffe04
    ZUQDECD ALL, $16, R30     // feffff04

// UQDECD  <Zdn>.D{, <pattern>{, MUL #<imm>}}
    ZUQDECD POW2, $1, Z0.D      // 00cce004
    ZUQDECD VL1, $1, Z1.D       // 21cce004
    ZUQDECD VL2, $2, Z2.D       // 42cce104
    ZUQDECD VL3, $2, Z3.D       // 63cce104
    ZUQDECD VL4, $3, Z4.D       // 84cce204
    ZUQDECD VL5, $3, Z5.D       // a5cce204
    ZUQDECD VL6, $4, Z6.D       // c6cce304
    ZUQDECD VL7, $4, Z7.D       // e7cce304
    ZUQDECD VL8, $5, Z8.D       // 08cde404
    ZUQDECD VL16, $5, Z9.D      // 29cde404
    ZUQDECD VL32, $6, Z10.D     // 4acde504
    ZUQDECD VL64, $6, Z11.D     // 6bcde504
    ZUQDECD VL128, $7, Z12.D    // 8ccde604
    ZUQDECD VL256, $7, Z13.D    // adcde604
    ZUQDECD $14, $8, Z14.D      // cecde704
    ZUQDECD $15, $8, Z15.D      // efcde704
    ZUQDECD $16, $9, Z16.D      // 10cee804
    ZUQDECD $17, $9, Z16.D      // 30cee804
    ZUQDECD $18, $9, Z17.D      // 51cee804
    ZUQDECD $19, $10, Z18.D     // 72cee904
    ZUQDECD $20, $10, Z19.D     // 93cee904
    ZUQDECD $21, $11, Z20.D     // b4ceea04
    ZUQDECD $22, $11, Z21.D     // d5ceea04
    ZUQDECD $23, $12, Z22.D     // f6ceeb04
    ZUQDECD $24, $12, Z23.D     // 17cfeb04
    ZUQDECD $25, $13, Z24.D     // 38cfec04
    ZUQDECD $26, $13, Z25.D     // 59cfec04
    ZUQDECD $27, $14, Z26.D     // 7acfed04
    ZUQDECD $28, $14, Z27.D     // 9bcfed04
    ZUQDECD MUL4, $15, Z28.D    // bccfee04
    ZUQDECD MUL3, $15, Z29.D    // ddcfee04
    ZUQDECD ALL, $16, Z31.D     // ffcfef04

// UQDECH  <Wdn>{, <pattern>{, MUL #<imm>}}
    ZUQDECHW POW2, $1, R0      // 00fc6004
    ZUQDECHW VL1, $1, R1       // 21fc6004
    ZUQDECHW VL2, $2, R2       // 42fc6104
    ZUQDECHW VL3, $2, R3       // 63fc6104
    ZUQDECHW VL4, $3, R4       // 84fc6204
    ZUQDECHW VL5, $3, R5       // a5fc6204
    ZUQDECHW VL6, $4, R6       // c6fc6304
    ZUQDECHW VL7, $4, R7       // e7fc6304
    ZUQDECHW VL8, $5, R8       // 08fd6404
    ZUQDECHW VL16, $5, R8      // 28fd6404
    ZUQDECHW VL32, $6, R9      // 49fd6504
    ZUQDECHW VL64, $6, R10     // 6afd6504
    ZUQDECHW VL128, $7, R11    // 8bfd6604
    ZUQDECHW VL256, $7, R12    // acfd6604
    ZUQDECHW $14, $8, R13      // cdfd6704
    ZUQDECHW $15, $8, R14      // eefd6704
    ZUQDECHW $16, $9, R15      // 0ffe6804
    ZUQDECHW $17, $9, R16      // 30fe6804
    ZUQDECHW $18, $9, R17      // 51fe6804
    ZUQDECHW $19, $10, R17     // 71fe6904
    ZUQDECHW $20, $10, R19     // 93fe6904
    ZUQDECHW $21, $11, R20     // b4fe6a04
    ZUQDECHW $22, $11, R21     // d5fe6a04
    ZUQDECHW $23, $12, R22     // f6fe6b04
    ZUQDECHW $24, $12, R22     // 16ff6b04
    ZUQDECHW $25, $13, R23     // 37ff6c04
    ZUQDECHW $26, $13, R24     // 58ff6c04
    ZUQDECHW $27, $14, R25     // 79ff6d04
    ZUQDECHW $28, $14, R26     // 9aff6d04
    ZUQDECHW MUL4, $15, R27    // bbff6e04
    ZUQDECHW MUL3, $15, R27    // dbff6e04
    ZUQDECHW ALL, $16, R30     // feff6f04

// UQDECH  <Xdn>{, <pattern>{, MUL #<imm>}}
    ZUQDECH POW2, $1, R0      // 00fc7004
    ZUQDECH VL1, $1, R1       // 21fc7004
    ZUQDECH VL2, $2, R2       // 42fc7104
    ZUQDECH VL3, $2, R3       // 63fc7104
    ZUQDECH VL4, $3, R4       // 84fc7204
    ZUQDECH VL5, $3, R5       // a5fc7204
    ZUQDECH VL6, $4, R6       // c6fc7304
    ZUQDECH VL7, $4, R7       // e7fc7304
    ZUQDECH VL8, $5, R8       // 08fd7404
    ZUQDECH VL16, $5, R8      // 28fd7404
    ZUQDECH VL32, $6, R9      // 49fd7504
    ZUQDECH VL64, $6, R10     // 6afd7504
    ZUQDECH VL128, $7, R11    // 8bfd7604
    ZUQDECH VL256, $7, R12    // acfd7604
    ZUQDECH $14, $8, R13      // cdfd7704
    ZUQDECH $15, $8, R14      // eefd7704
    ZUQDECH $16, $9, R15      // 0ffe7804
    ZUQDECH $17, $9, R16      // 30fe7804
    ZUQDECH $18, $9, R17      // 51fe7804
    ZUQDECH $19, $10, R17     // 71fe7904
    ZUQDECH $20, $10, R19     // 93fe7904
    ZUQDECH $21, $11, R20     // b4fe7a04
    ZUQDECH $22, $11, R21     // d5fe7a04
    ZUQDECH $23, $12, R22     // f6fe7b04
    ZUQDECH $24, $12, R22     // 16ff7b04
    ZUQDECH $25, $13, R23     // 37ff7c04
    ZUQDECH $26, $13, R24     // 58ff7c04
    ZUQDECH $27, $14, R25     // 79ff7d04
    ZUQDECH $28, $14, R26     // 9aff7d04
    ZUQDECH MUL4, $15, R27    // bbff7e04
    ZUQDECH MUL3, $15, R27    // dbff7e04
    ZUQDECH ALL, $16, R30     // feff7f04

// UQDECH  <Zdn>.H{, <pattern>{, MUL #<imm>}}
    ZUQDECH POW2, $1, Z0.H      // 00cc6004
    ZUQDECH VL1, $1, Z1.H       // 21cc6004
    ZUQDECH VL2, $2, Z2.H       // 42cc6104
    ZUQDECH VL3, $2, Z3.H       // 63cc6104
    ZUQDECH VL4, $3, Z4.H       // 84cc6204
    ZUQDECH VL5, $3, Z5.H       // a5cc6204
    ZUQDECH VL6, $4, Z6.H       // c6cc6304
    ZUQDECH VL7, $4, Z7.H       // e7cc6304
    ZUQDECH VL8, $5, Z8.H       // 08cd6404
    ZUQDECH VL16, $5, Z9.H      // 29cd6404
    ZUQDECH VL32, $6, Z10.H     // 4acd6504
    ZUQDECH VL64, $6, Z11.H     // 6bcd6504
    ZUQDECH VL128, $7, Z12.H    // 8ccd6604
    ZUQDECH VL256, $7, Z13.H    // adcd6604
    ZUQDECH $14, $8, Z14.H      // cecd6704
    ZUQDECH $15, $8, Z15.H      // efcd6704
    ZUQDECH $16, $9, Z16.H      // 10ce6804
    ZUQDECH $17, $9, Z16.H      // 30ce6804
    ZUQDECH $18, $9, Z17.H      // 51ce6804
    ZUQDECH $19, $10, Z18.H     // 72ce6904
    ZUQDECH $20, $10, Z19.H     // 93ce6904
    ZUQDECH $21, $11, Z20.H     // b4ce6a04
    ZUQDECH $22, $11, Z21.H     // d5ce6a04
    ZUQDECH $23, $12, Z22.H     // f6ce6b04
    ZUQDECH $24, $12, Z23.H     // 17cf6b04
    ZUQDECH $25, $13, Z24.H     // 38cf6c04
    ZUQDECH $26, $13, Z25.H     // 59cf6c04
    ZUQDECH $27, $14, Z26.H     // 7acf6d04
    ZUQDECH $28, $14, Z27.H     // 9bcf6d04
    ZUQDECH MUL4, $15, Z28.H    // bccf6e04
    ZUQDECH MUL3, $15, Z29.H    // ddcf6e04
    ZUQDECH ALL, $16, Z31.H     // ffcf6f04

// UQDECP  <Wdn>, <Pm>.<T>
    PUQDECPW P0.B, R0      // 00882b25
    PUQDECPW P6.B, R10     // ca882b25
    PUQDECPW P15.B, R30    // fe892b25
    PUQDECPW P0.H, R0      // 00886b25
    PUQDECPW P6.H, R10     // ca886b25
    PUQDECPW P15.H, R30    // fe896b25
    PUQDECPW P0.S, R0      // 0088ab25
    PUQDECPW P6.S, R10     // ca88ab25
    PUQDECPW P15.S, R30    // fe89ab25
    PUQDECPW P0.D, R0      // 0088eb25
    PUQDECPW P6.D, R10     // ca88eb25
    PUQDECPW P15.D, R30    // fe89eb25

// UQDECP  <Xdn>, <Pm>.<T>
    PUQDECP P0.B, R0      // 008c2b25
    PUQDECP P6.B, R10     // ca8c2b25
    PUQDECP P15.B, R30    // fe8d2b25
    PUQDECP P0.H, R0      // 008c6b25
    PUQDECP P6.H, R10     // ca8c6b25
    PUQDECP P15.H, R30    // fe8d6b25
    PUQDECP P0.S, R0      // 008cab25
    PUQDECP P6.S, R10     // ca8cab25
    PUQDECP P15.S, R30    // fe8dab25
    PUQDECP P0.D, R0      // 008ceb25
    PUQDECP P6.D, R10     // ca8ceb25
    PUQDECP P15.D, R30    // fe8deb25

// UQDECP  <Zdn>.<T>, <Pm>.<T>
    ZUQDECP P0.H, Z0.H      // 00806b25
    ZUQDECP P6.H, Z10.H     // ca806b25
    ZUQDECP P15.H, Z31.H    // ff816b25
    ZUQDECP P0.S, Z0.S      // 0080ab25
    ZUQDECP P6.S, Z10.S     // ca80ab25
    ZUQDECP P15.S, Z31.S    // ff81ab25
    ZUQDECP P0.D, Z0.D      // 0080eb25
    ZUQDECP P6.D, Z10.D     // ca80eb25
    ZUQDECP P15.D, Z31.D    // ff81eb25

// UQDECW  <Wdn>{, <pattern>{, MUL #<imm>}}
    ZUQDECWW POW2, $1, R0      // 00fca004
    ZUQDECWW VL1, $1, R1       // 21fca004
    ZUQDECWW VL2, $2, R2       // 42fca104
    ZUQDECWW VL3, $2, R3       // 63fca104
    ZUQDECWW VL4, $3, R4       // 84fca204
    ZUQDECWW VL5, $3, R5       // a5fca204
    ZUQDECWW VL6, $4, R6       // c6fca304
    ZUQDECWW VL7, $4, R7       // e7fca304
    ZUQDECWW VL8, $5, R8       // 08fda404
    ZUQDECWW VL16, $5, R8      // 28fda404
    ZUQDECWW VL32, $6, R9      // 49fda504
    ZUQDECWW VL64, $6, R10     // 6afda504
    ZUQDECWW VL128, $7, R11    // 8bfda604
    ZUQDECWW VL256, $7, R12    // acfda604
    ZUQDECWW $14, $8, R13      // cdfda704
    ZUQDECWW $15, $8, R14      // eefda704
    ZUQDECWW $16, $9, R15      // 0ffea804
    ZUQDECWW $17, $9, R16      // 30fea804
    ZUQDECWW $18, $9, R17      // 51fea804
    ZUQDECWW $19, $10, R17     // 71fea904
    ZUQDECWW $20, $10, R19     // 93fea904
    ZUQDECWW $21, $11, R20     // b4feaa04
    ZUQDECWW $22, $11, R21     // d5feaa04
    ZUQDECWW $23, $12, R22     // f6feab04
    ZUQDECWW $24, $12, R22     // 16ffab04
    ZUQDECWW $25, $13, R23     // 37ffac04
    ZUQDECWW $26, $13, R24     // 58ffac04
    ZUQDECWW $27, $14, R25     // 79ffad04
    ZUQDECWW $28, $14, R26     // 9affad04
    ZUQDECWW MUL4, $15, R27    // bbffae04
    ZUQDECWW MUL3, $15, R27    // dbffae04
    ZUQDECWW ALL, $16, R30     // feffaf04

// UQDECW  <Xdn>{, <pattern>{, MUL #<imm>}}
    ZUQDECW POW2, $1, R0      // 00fcb004
    ZUQDECW VL1, $1, R1       // 21fcb004
    ZUQDECW VL2, $2, R2       // 42fcb104
    ZUQDECW VL3, $2, R3       // 63fcb104
    ZUQDECW VL4, $3, R4       // 84fcb204
    ZUQDECW VL5, $3, R5       // a5fcb204
    ZUQDECW VL6, $4, R6       // c6fcb304
    ZUQDECW VL7, $4, R7       // e7fcb304
    ZUQDECW VL8, $5, R8       // 08fdb404
    ZUQDECW VL16, $5, R8      // 28fdb404
    ZUQDECW VL32, $6, R9      // 49fdb504
    ZUQDECW VL64, $6, R10     // 6afdb504
    ZUQDECW VL128, $7, R11    // 8bfdb604
    ZUQDECW VL256, $7, R12    // acfdb604
    ZUQDECW $14, $8, R13      // cdfdb704
    ZUQDECW $15, $8, R14      // eefdb704
    ZUQDECW $16, $9, R15      // 0ffeb804
    ZUQDECW $17, $9, R16      // 30feb804
    ZUQDECW $18, $9, R17      // 51feb804
    ZUQDECW $19, $10, R17     // 71feb904
    ZUQDECW $20, $10, R19     // 93feb904
    ZUQDECW $21, $11, R20     // b4feba04
    ZUQDECW $22, $11, R21     // d5feba04
    ZUQDECW $23, $12, R22     // f6febb04
    ZUQDECW $24, $12, R22     // 16ffbb04
    ZUQDECW $25, $13, R23     // 37ffbc04
    ZUQDECW $26, $13, R24     // 58ffbc04
    ZUQDECW $27, $14, R25     // 79ffbd04
    ZUQDECW $28, $14, R26     // 9affbd04
    ZUQDECW MUL4, $15, R27    // bbffbe04
    ZUQDECW MUL3, $15, R27    // dbffbe04
    ZUQDECW ALL, $16, R30     // feffbf04

// UQDECW  <Zdn>.S{, <pattern>{, MUL #<imm>}}
    ZUQDECW POW2, $1, Z0.S      // 00cca004
    ZUQDECW VL1, $1, Z1.S       // 21cca004
    ZUQDECW VL2, $2, Z2.S       // 42cca104
    ZUQDECW VL3, $2, Z3.S       // 63cca104
    ZUQDECW VL4, $3, Z4.S       // 84cca204
    ZUQDECW VL5, $3, Z5.S       // a5cca204
    ZUQDECW VL6, $4, Z6.S       // c6cca304
    ZUQDECW VL7, $4, Z7.S       // e7cca304
    ZUQDECW VL8, $5, Z8.S       // 08cda404
    ZUQDECW VL16, $5, Z9.S      // 29cda404
    ZUQDECW VL32, $6, Z10.S     // 4acda504
    ZUQDECW VL64, $6, Z11.S     // 6bcda504
    ZUQDECW VL128, $7, Z12.S    // 8ccda604
    ZUQDECW VL256, $7, Z13.S    // adcda604
    ZUQDECW $14, $8, Z14.S      // cecda704
    ZUQDECW $15, $8, Z15.S      // efcda704
    ZUQDECW $16, $9, Z16.S      // 10cea804
    ZUQDECW $17, $9, Z16.S      // 30cea804
    ZUQDECW $18, $9, Z17.S      // 51cea804
    ZUQDECW $19, $10, Z18.S     // 72cea904
    ZUQDECW $20, $10, Z19.S     // 93cea904
    ZUQDECW $21, $11, Z20.S     // b4ceaa04
    ZUQDECW $22, $11, Z21.S     // d5ceaa04
    ZUQDECW $23, $12, Z22.S     // f6ceab04
    ZUQDECW $24, $12, Z23.S     // 17cfab04
    ZUQDECW $25, $13, Z24.S     // 38cfac04
    ZUQDECW $26, $13, Z25.S     // 59cfac04
    ZUQDECW $27, $14, Z26.S     // 7acfad04
    ZUQDECW $28, $14, Z27.S     // 9bcfad04
    ZUQDECW MUL4, $15, Z28.S    // bccfae04
    ZUQDECW MUL3, $15, Z29.S    // ddcfae04
    ZUQDECW ALL, $16, Z31.S     // ffcfaf04

// UQINCB  <Wdn>{, <pattern>{, MUL #<imm>}}
    ZUQINCBW POW2, $1, R0      // 00f42004
    ZUQINCBW VL1, $1, R1       // 21f42004
    ZUQINCBW VL2, $2, R2       // 42f42104
    ZUQINCBW VL3, $2, R3       // 63f42104
    ZUQINCBW VL4, $3, R4       // 84f42204
    ZUQINCBW VL5, $3, R5       // a5f42204
    ZUQINCBW VL6, $4, R6       // c6f42304
    ZUQINCBW VL7, $4, R7       // e7f42304
    ZUQINCBW VL8, $5, R8       // 08f52404
    ZUQINCBW VL16, $5, R8      // 28f52404
    ZUQINCBW VL32, $6, R9      // 49f52504
    ZUQINCBW VL64, $6, R10     // 6af52504
    ZUQINCBW VL128, $7, R11    // 8bf52604
    ZUQINCBW VL256, $7, R12    // acf52604
    ZUQINCBW $14, $8, R13      // cdf52704
    ZUQINCBW $15, $8, R14      // eef52704
    ZUQINCBW $16, $9, R15      // 0ff62804
    ZUQINCBW $17, $9, R16      // 30f62804
    ZUQINCBW $18, $9, R17      // 51f62804
    ZUQINCBW $19, $10, R17     // 71f62904
    ZUQINCBW $20, $10, R19     // 93f62904
    ZUQINCBW $21, $11, R20     // b4f62a04
    ZUQINCBW $22, $11, R21     // d5f62a04
    ZUQINCBW $23, $12, R22     // f6f62b04
    ZUQINCBW $24, $12, R22     // 16f72b04
    ZUQINCBW $25, $13, R23     // 37f72c04
    ZUQINCBW $26, $13, R24     // 58f72c04
    ZUQINCBW $27, $14, R25     // 79f72d04
    ZUQINCBW $28, $14, R26     // 9af72d04
    ZUQINCBW MUL4, $15, R27    // bbf72e04
    ZUQINCBW MUL3, $15, R27    // dbf72e04
    ZUQINCBW ALL, $16, R30     // fef72f04

// UQINCB  <Xdn>{, <pattern>{, MUL #<imm>}}
    ZUQINCB POW2, $1, R0      // 00f43004
    ZUQINCB VL1, $1, R1       // 21f43004
    ZUQINCB VL2, $2, R2       // 42f43104
    ZUQINCB VL3, $2, R3       // 63f43104
    ZUQINCB VL4, $3, R4       // 84f43204
    ZUQINCB VL5, $3, R5       // a5f43204
    ZUQINCB VL6, $4, R6       // c6f43304
    ZUQINCB VL7, $4, R7       // e7f43304
    ZUQINCB VL8, $5, R8       // 08f53404
    ZUQINCB VL16, $5, R8      // 28f53404
    ZUQINCB VL32, $6, R9      // 49f53504
    ZUQINCB VL64, $6, R10     // 6af53504
    ZUQINCB VL128, $7, R11    // 8bf53604
    ZUQINCB VL256, $7, R12    // acf53604
    ZUQINCB $14, $8, R13      // cdf53704
    ZUQINCB $15, $8, R14      // eef53704
    ZUQINCB $16, $9, R15      // 0ff63804
    ZUQINCB $17, $9, R16      // 30f63804
    ZUQINCB $18, $9, R17      // 51f63804
    ZUQINCB $19, $10, R17     // 71f63904
    ZUQINCB $20, $10, R19     // 93f63904
    ZUQINCB $21, $11, R20     // b4f63a04
    ZUQINCB $22, $11, R21     // d5f63a04
    ZUQINCB $23, $12, R22     // f6f63b04
    ZUQINCB $24, $12, R22     // 16f73b04
    ZUQINCB $25, $13, R23     // 37f73c04
    ZUQINCB $26, $13, R24     // 58f73c04
    ZUQINCB $27, $14, R25     // 79f73d04
    ZUQINCB $28, $14, R26     // 9af73d04
    ZUQINCB MUL4, $15, R27    // bbf73e04
    ZUQINCB MUL3, $15, R27    // dbf73e04
    ZUQINCB ALL, $16, R30     // fef73f04

// UQINCD  <Wdn>{, <pattern>{, MUL #<imm>}}
    ZUQINCDW POW2, $1, R0      // 00f4e004
    ZUQINCDW VL1, $1, R1       // 21f4e004
    ZUQINCDW VL2, $2, R2       // 42f4e104
    ZUQINCDW VL3, $2, R3       // 63f4e104
    ZUQINCDW VL4, $3, R4       // 84f4e204
    ZUQINCDW VL5, $3, R5       // a5f4e204
    ZUQINCDW VL6, $4, R6       // c6f4e304
    ZUQINCDW VL7, $4, R7       // e7f4e304
    ZUQINCDW VL8, $5, R8       // 08f5e404
    ZUQINCDW VL16, $5, R8      // 28f5e404
    ZUQINCDW VL32, $6, R9      // 49f5e504
    ZUQINCDW VL64, $6, R10     // 6af5e504
    ZUQINCDW VL128, $7, R11    // 8bf5e604
    ZUQINCDW VL256, $7, R12    // acf5e604
    ZUQINCDW $14, $8, R13      // cdf5e704
    ZUQINCDW $15, $8, R14      // eef5e704
    ZUQINCDW $16, $9, R15      // 0ff6e804
    ZUQINCDW $17, $9, R16      // 30f6e804
    ZUQINCDW $18, $9, R17      // 51f6e804
    ZUQINCDW $19, $10, R17     // 71f6e904
    ZUQINCDW $20, $10, R19     // 93f6e904
    ZUQINCDW $21, $11, R20     // b4f6ea04
    ZUQINCDW $22, $11, R21     // d5f6ea04
    ZUQINCDW $23, $12, R22     // f6f6eb04
    ZUQINCDW $24, $12, R22     // 16f7eb04
    ZUQINCDW $25, $13, R23     // 37f7ec04
    ZUQINCDW $26, $13, R24     // 58f7ec04
    ZUQINCDW $27, $14, R25     // 79f7ed04
    ZUQINCDW $28, $14, R26     // 9af7ed04
    ZUQINCDW MUL4, $15, R27    // bbf7ee04
    ZUQINCDW MUL3, $15, R27    // dbf7ee04
    ZUQINCDW ALL, $16, R30     // fef7ef04

// UQINCD  <Xdn>{, <pattern>{, MUL #<imm>}}
    ZUQINCD POW2, $1, R0      // 00f4f004
    ZUQINCD VL1, $1, R1       // 21f4f004
    ZUQINCD VL2, $2, R2       // 42f4f104
    ZUQINCD VL3, $2, R3       // 63f4f104
    ZUQINCD VL4, $3, R4       // 84f4f204
    ZUQINCD VL5, $3, R5       // a5f4f204
    ZUQINCD VL6, $4, R6       // c6f4f304
    ZUQINCD VL7, $4, R7       // e7f4f304
    ZUQINCD VL8, $5, R8       // 08f5f404
    ZUQINCD VL16, $5, R8      // 28f5f404
    ZUQINCD VL32, $6, R9      // 49f5f504
    ZUQINCD VL64, $6, R10     // 6af5f504
    ZUQINCD VL128, $7, R11    // 8bf5f604
    ZUQINCD VL256, $7, R12    // acf5f604
    ZUQINCD $14, $8, R13      // cdf5f704
    ZUQINCD $15, $8, R14      // eef5f704
    ZUQINCD $16, $9, R15      // 0ff6f804
    ZUQINCD $17, $9, R16      // 30f6f804
    ZUQINCD $18, $9, R17      // 51f6f804
    ZUQINCD $19, $10, R17     // 71f6f904
    ZUQINCD $20, $10, R19     // 93f6f904
    ZUQINCD $21, $11, R20     // b4f6fa04
    ZUQINCD $22, $11, R21     // d5f6fa04
    ZUQINCD $23, $12, R22     // f6f6fb04
    ZUQINCD $24, $12, R22     // 16f7fb04
    ZUQINCD $25, $13, R23     // 37f7fc04
    ZUQINCD $26, $13, R24     // 58f7fc04
    ZUQINCD $27, $14, R25     // 79f7fd04
    ZUQINCD $28, $14, R26     // 9af7fd04
    ZUQINCD MUL4, $15, R27    // bbf7fe04
    ZUQINCD MUL3, $15, R27    // dbf7fe04
    ZUQINCD ALL, $16, R30     // fef7ff04

// UQINCD  <Zdn>.D{, <pattern>{, MUL #<imm>}}
    ZUQINCD POW2, $1, Z0.D      // 00c4e004
    ZUQINCD VL1, $1, Z1.D       // 21c4e004
    ZUQINCD VL2, $2, Z2.D       // 42c4e104
    ZUQINCD VL3, $2, Z3.D       // 63c4e104
    ZUQINCD VL4, $3, Z4.D       // 84c4e204
    ZUQINCD VL5, $3, Z5.D       // a5c4e204
    ZUQINCD VL6, $4, Z6.D       // c6c4e304
    ZUQINCD VL7, $4, Z7.D       // e7c4e304
    ZUQINCD VL8, $5, Z8.D       // 08c5e404
    ZUQINCD VL16, $5, Z9.D      // 29c5e404
    ZUQINCD VL32, $6, Z10.D     // 4ac5e504
    ZUQINCD VL64, $6, Z11.D     // 6bc5e504
    ZUQINCD VL128, $7, Z12.D    // 8cc5e604
    ZUQINCD VL256, $7, Z13.D    // adc5e604
    ZUQINCD $14, $8, Z14.D      // cec5e704
    ZUQINCD $15, $8, Z15.D      // efc5e704
    ZUQINCD $16, $9, Z16.D      // 10c6e804
    ZUQINCD $17, $9, Z16.D      // 30c6e804
    ZUQINCD $18, $9, Z17.D      // 51c6e804
    ZUQINCD $19, $10, Z18.D     // 72c6e904
    ZUQINCD $20, $10, Z19.D     // 93c6e904
    ZUQINCD $21, $11, Z20.D     // b4c6ea04
    ZUQINCD $22, $11, Z21.D     // d5c6ea04
    ZUQINCD $23, $12, Z22.D     // f6c6eb04
    ZUQINCD $24, $12, Z23.D     // 17c7eb04
    ZUQINCD $25, $13, Z24.D     // 38c7ec04
    ZUQINCD $26, $13, Z25.D     // 59c7ec04
    ZUQINCD $27, $14, Z26.D     // 7ac7ed04
    ZUQINCD $28, $14, Z27.D     // 9bc7ed04
    ZUQINCD MUL4, $15, Z28.D    // bcc7ee04
    ZUQINCD MUL3, $15, Z29.D    // ddc7ee04
    ZUQINCD ALL, $16, Z31.D     // ffc7ef04

// UQINCH  <Wdn>{, <pattern>{, MUL #<imm>}}
    ZUQINCHW POW2, $1, R0      // 00f46004
    ZUQINCHW VL1, $1, R1       // 21f46004
    ZUQINCHW VL2, $2, R2       // 42f46104
    ZUQINCHW VL3, $2, R3       // 63f46104
    ZUQINCHW VL4, $3, R4       // 84f46204
    ZUQINCHW VL5, $3, R5       // a5f46204
    ZUQINCHW VL6, $4, R6       // c6f46304
    ZUQINCHW VL7, $4, R7       // e7f46304
    ZUQINCHW VL8, $5, R8       // 08f56404
    ZUQINCHW VL16, $5, R8      // 28f56404
    ZUQINCHW VL32, $6, R9      // 49f56504
    ZUQINCHW VL64, $6, R10     // 6af56504
    ZUQINCHW VL128, $7, R11    // 8bf56604
    ZUQINCHW VL256, $7, R12    // acf56604
    ZUQINCHW $14, $8, R13      // cdf56704
    ZUQINCHW $15, $8, R14      // eef56704
    ZUQINCHW $16, $9, R15      // 0ff66804
    ZUQINCHW $17, $9, R16      // 30f66804
    ZUQINCHW $18, $9, R17      // 51f66804
    ZUQINCHW $19, $10, R17     // 71f66904
    ZUQINCHW $20, $10, R19     // 93f66904
    ZUQINCHW $21, $11, R20     // b4f66a04
    ZUQINCHW $22, $11, R21     // d5f66a04
    ZUQINCHW $23, $12, R22     // f6f66b04
    ZUQINCHW $24, $12, R22     // 16f76b04
    ZUQINCHW $25, $13, R23     // 37f76c04
    ZUQINCHW $26, $13, R24     // 58f76c04
    ZUQINCHW $27, $14, R25     // 79f76d04
    ZUQINCHW $28, $14, R26     // 9af76d04
    ZUQINCHW MUL4, $15, R27    // bbf76e04
    ZUQINCHW MUL3, $15, R27    // dbf76e04
    ZUQINCHW ALL, $16, R30     // fef76f04

// UQINCH  <Xdn>{, <pattern>{, MUL #<imm>}}
    ZUQINCH POW2, $1, R0      // 00f47004
    ZUQINCH VL1, $1, R1       // 21f47004
    ZUQINCH VL2, $2, R2       // 42f47104
    ZUQINCH VL3, $2, R3       // 63f47104
    ZUQINCH VL4, $3, R4       // 84f47204
    ZUQINCH VL5, $3, R5       // a5f47204
    ZUQINCH VL6, $4, R6       // c6f47304
    ZUQINCH VL7, $4, R7       // e7f47304
    ZUQINCH VL8, $5, R8       // 08f57404
    ZUQINCH VL16, $5, R8      // 28f57404
    ZUQINCH VL32, $6, R9      // 49f57504
    ZUQINCH VL64, $6, R10     // 6af57504
    ZUQINCH VL128, $7, R11    // 8bf57604
    ZUQINCH VL256, $7, R12    // acf57604
    ZUQINCH $14, $8, R13      // cdf57704
    ZUQINCH $15, $8, R14      // eef57704
    ZUQINCH $16, $9, R15      // 0ff67804
    ZUQINCH $17, $9, R16      // 30f67804
    ZUQINCH $18, $9, R17      // 51f67804
    ZUQINCH $19, $10, R17     // 71f67904
    ZUQINCH $20, $10, R19     // 93f67904
    ZUQINCH $21, $11, R20     // b4f67a04
    ZUQINCH $22, $11, R21     // d5f67a04
    ZUQINCH $23, $12, R22     // f6f67b04
    ZUQINCH $24, $12, R22     // 16f77b04
    ZUQINCH $25, $13, R23     // 37f77c04
    ZUQINCH $26, $13, R24     // 58f77c04
    ZUQINCH $27, $14, R25     // 79f77d04
    ZUQINCH $28, $14, R26     // 9af77d04
    ZUQINCH MUL4, $15, R27    // bbf77e04
    ZUQINCH MUL3, $15, R27    // dbf77e04
    ZUQINCH ALL, $16, R30     // fef77f04

// UQINCH  <Zdn>.H{, <pattern>{, MUL #<imm>}}
    ZUQINCH POW2, $1, Z0.H      // 00c46004
    ZUQINCH VL1, $1, Z1.H       // 21c46004
    ZUQINCH VL2, $2, Z2.H       // 42c46104
    ZUQINCH VL3, $2, Z3.H       // 63c46104
    ZUQINCH VL4, $3, Z4.H       // 84c46204
    ZUQINCH VL5, $3, Z5.H       // a5c46204
    ZUQINCH VL6, $4, Z6.H       // c6c46304
    ZUQINCH VL7, $4, Z7.H       // e7c46304
    ZUQINCH VL8, $5, Z8.H       // 08c56404
    ZUQINCH VL16, $5, Z9.H      // 29c56404
    ZUQINCH VL32, $6, Z10.H     // 4ac56504
    ZUQINCH VL64, $6, Z11.H     // 6bc56504
    ZUQINCH VL128, $7, Z12.H    // 8cc56604
    ZUQINCH VL256, $7, Z13.H    // adc56604
    ZUQINCH $14, $8, Z14.H      // cec56704
    ZUQINCH $15, $8, Z15.H      // efc56704
    ZUQINCH $16, $9, Z16.H      // 10c66804
    ZUQINCH $17, $9, Z16.H      // 30c66804
    ZUQINCH $18, $9, Z17.H      // 51c66804
    ZUQINCH $19, $10, Z18.H     // 72c66904
    ZUQINCH $20, $10, Z19.H     // 93c66904
    ZUQINCH $21, $11, Z20.H     // b4c66a04
    ZUQINCH $22, $11, Z21.H     // d5c66a04
    ZUQINCH $23, $12, Z22.H     // f6c66b04
    ZUQINCH $24, $12, Z23.H     // 17c76b04
    ZUQINCH $25, $13, Z24.H     // 38c76c04
    ZUQINCH $26, $13, Z25.H     // 59c76c04
    ZUQINCH $27, $14, Z26.H     // 7ac76d04
    ZUQINCH $28, $14, Z27.H     // 9bc76d04
    ZUQINCH MUL4, $15, Z28.H    // bcc76e04
    ZUQINCH MUL3, $15, Z29.H    // ddc76e04
    ZUQINCH ALL, $16, Z31.H     // ffc76f04

// UQINCP  <Wdn>, <Pm>.<T>
    PUQINCPW P0.B, R0      // 00882925
    PUQINCPW P6.B, R10     // ca882925
    PUQINCPW P15.B, R30    // fe892925
    PUQINCPW P0.H, R0      // 00886925
    PUQINCPW P6.H, R10     // ca886925
    PUQINCPW P15.H, R30    // fe896925
    PUQINCPW P0.S, R0      // 0088a925
    PUQINCPW P6.S, R10     // ca88a925
    PUQINCPW P15.S, R30    // fe89a925
    PUQINCPW P0.D, R0      // 0088e925
    PUQINCPW P6.D, R10     // ca88e925
    PUQINCPW P15.D, R30    // fe89e925

// UQINCP  <Xdn>, <Pm>.<T>
    PUQINCP P0.B, R0      // 008c2925
    PUQINCP P6.B, R10     // ca8c2925
    PUQINCP P15.B, R30    // fe8d2925
    PUQINCP P0.H, R0      // 008c6925
    PUQINCP P6.H, R10     // ca8c6925
    PUQINCP P15.H, R30    // fe8d6925
    PUQINCP P0.S, R0      // 008ca925
    PUQINCP P6.S, R10     // ca8ca925
    PUQINCP P15.S, R30    // fe8da925
    PUQINCP P0.D, R0      // 008ce925
    PUQINCP P6.D, R10     // ca8ce925
    PUQINCP P15.D, R30    // fe8de925

// UQINCP  <Zdn>.<T>, <Pm>.<T>
    ZUQINCP P0.H, Z0.H      // 00806925
    ZUQINCP P6.H, Z10.H     // ca806925
    ZUQINCP P15.H, Z31.H    // ff816925
    ZUQINCP P0.S, Z0.S      // 0080a925
    ZUQINCP P6.S, Z10.S     // ca80a925
    ZUQINCP P15.S, Z31.S    // ff81a925
    ZUQINCP P0.D, Z0.D      // 0080e925
    ZUQINCP P6.D, Z10.D     // ca80e925
    ZUQINCP P15.D, Z31.D    // ff81e925

// UQINCW  <Wdn>{, <pattern>{, MUL #<imm>}}
    ZUQINCWW POW2, $1, R0      // 00f4a004
    ZUQINCWW VL1, $1, R1       // 21f4a004
    ZUQINCWW VL2, $2, R2       // 42f4a104
    ZUQINCWW VL3, $2, R3       // 63f4a104
    ZUQINCWW VL4, $3, R4       // 84f4a204
    ZUQINCWW VL5, $3, R5       // a5f4a204
    ZUQINCWW VL6, $4, R6       // c6f4a304
    ZUQINCWW VL7, $4, R7       // e7f4a304
    ZUQINCWW VL8, $5, R8       // 08f5a404
    ZUQINCWW VL16, $5, R8      // 28f5a404
    ZUQINCWW VL32, $6, R9      // 49f5a504
    ZUQINCWW VL64, $6, R10     // 6af5a504
    ZUQINCWW VL128, $7, R11    // 8bf5a604
    ZUQINCWW VL256, $7, R12    // acf5a604
    ZUQINCWW $14, $8, R13      // cdf5a704
    ZUQINCWW $15, $8, R14      // eef5a704
    ZUQINCWW $16, $9, R15      // 0ff6a804
    ZUQINCWW $17, $9, R16      // 30f6a804
    ZUQINCWW $18, $9, R17      // 51f6a804
    ZUQINCWW $19, $10, R17     // 71f6a904
    ZUQINCWW $20, $10, R19     // 93f6a904
    ZUQINCWW $21, $11, R20     // b4f6aa04
    ZUQINCWW $22, $11, R21     // d5f6aa04
    ZUQINCWW $23, $12, R22     // f6f6ab04
    ZUQINCWW $24, $12, R22     // 16f7ab04
    ZUQINCWW $25, $13, R23     // 37f7ac04
    ZUQINCWW $26, $13, R24     // 58f7ac04
    ZUQINCWW $27, $14, R25     // 79f7ad04
    ZUQINCWW $28, $14, R26     // 9af7ad04
    ZUQINCWW MUL4, $15, R27    // bbf7ae04
    ZUQINCWW MUL3, $15, R27    // dbf7ae04
    ZUQINCWW ALL, $16, R30     // fef7af04

// UQINCW  <Xdn>{, <pattern>{, MUL #<imm>}}
    ZUQINCW POW2, $1, R0      // 00f4b004
    ZUQINCW VL1, $1, R1       // 21f4b004
    ZUQINCW VL2, $2, R2       // 42f4b104
    ZUQINCW VL3, $2, R3       // 63f4b104
    ZUQINCW VL4, $3, R4       // 84f4b204
    ZUQINCW VL5, $3, R5       // a5f4b204
    ZUQINCW VL6, $4, R6       // c6f4b304
    ZUQINCW VL7, $4, R7       // e7f4b304
    ZUQINCW VL8, $5, R8       // 08f5b404
    ZUQINCW VL16, $5, R8      // 28f5b404
    ZUQINCW VL32, $6, R9      // 49f5b504
    ZUQINCW VL64, $6, R10     // 6af5b504
    ZUQINCW VL128, $7, R11    // 8bf5b604
    ZUQINCW VL256, $7, R12    // acf5b604
    ZUQINCW $14, $8, R13      // cdf5b704
    ZUQINCW $15, $8, R14      // eef5b704
    ZUQINCW $16, $9, R15      // 0ff6b804
    ZUQINCW $17, $9, R16      // 30f6b804
    ZUQINCW $18, $9, R17      // 51f6b804
    ZUQINCW $19, $10, R17     // 71f6b904
    ZUQINCW $20, $10, R19     // 93f6b904
    ZUQINCW $21, $11, R20     // b4f6ba04
    ZUQINCW $22, $11, R21     // d5f6ba04
    ZUQINCW $23, $12, R22     // f6f6bb04
    ZUQINCW $24, $12, R22     // 16f7bb04
    ZUQINCW $25, $13, R23     // 37f7bc04
    ZUQINCW $26, $13, R24     // 58f7bc04
    ZUQINCW $27, $14, R25     // 79f7bd04
    ZUQINCW $28, $14, R26     // 9af7bd04
    ZUQINCW MUL4, $15, R27    // bbf7be04
    ZUQINCW MUL3, $15, R27    // dbf7be04
    ZUQINCW ALL, $16, R30     // fef7bf04

// UQINCW  <Zdn>.S{, <pattern>{, MUL #<imm>}}
    ZUQINCW POW2, $1, Z0.S      // 00c4a004
    ZUQINCW VL1, $1, Z1.S       // 21c4a004
    ZUQINCW VL2, $2, Z2.S       // 42c4a104
    ZUQINCW VL3, $2, Z3.S       // 63c4a104
    ZUQINCW VL4, $3, Z4.S       // 84c4a204
    ZUQINCW VL5, $3, Z5.S       // a5c4a204
    ZUQINCW VL6, $4, Z6.S       // c6c4a304
    ZUQINCW VL7, $4, Z7.S       // e7c4a304
    ZUQINCW VL8, $5, Z8.S       // 08c5a404
    ZUQINCW VL16, $5, Z9.S      // 29c5a404
    ZUQINCW VL32, $6, Z10.S     // 4ac5a504
    ZUQINCW VL64, $6, Z11.S     // 6bc5a504
    ZUQINCW VL128, $7, Z12.S    // 8cc5a604
    ZUQINCW VL256, $7, Z13.S    // adc5a604
    ZUQINCW $14, $8, Z14.S      // cec5a704
    ZUQINCW $15, $8, Z15.S      // efc5a704
    ZUQINCW $16, $9, Z16.S      // 10c6a804
    ZUQINCW $17, $9, Z16.S      // 30c6a804
    ZUQINCW $18, $9, Z17.S      // 51c6a804
    ZUQINCW $19, $10, Z18.S     // 72c6a904
    ZUQINCW $20, $10, Z19.S     // 93c6a904
    ZUQINCW $21, $11, Z20.S     // b4c6aa04
    ZUQINCW $22, $11, Z21.S     // d5c6aa04
    ZUQINCW $23, $12, Z22.S     // f6c6ab04
    ZUQINCW $24, $12, Z23.S     // 17c7ab04
    ZUQINCW $25, $13, Z24.S     // 38c7ac04
    ZUQINCW $26, $13, Z25.S     // 59c7ac04
    ZUQINCW $27, $14, Z26.S     // 7ac7ad04
    ZUQINCW $28, $14, Z27.S     // 9bc7ad04
    ZUQINCW MUL4, $15, Z28.S    // bcc7ae04
    ZUQINCW MUL3, $15, Z29.S    // ddc7ae04
    ZUQINCW ALL, $16, Z31.S     // ffc7af04

// UQSUB   <Zdn>.<T>, <Zdn>.<T>, #<imm>, <shift>
    ZUQSUB Z0.B, $0, $0, Z0.B        // 00c02725
    ZUQSUB Z10.B, $85, $0, Z10.B     // aaca2725
    ZUQSUB Z31.B, $255, $0, Z31.B    // ffdf2725
    ZUQSUB Z0.H, $0, $8, Z0.H        // 00e06725
    ZUQSUB Z10.H, $85, $8, Z10.H     // aaea6725
    ZUQSUB Z31.H, $255, $0, Z31.H    // ffdf6725
    ZUQSUB Z0.S, $0, $8, Z0.S        // 00e0a725
    ZUQSUB Z10.S, $85, $8, Z10.S     // aaeaa725
    ZUQSUB Z31.S, $255, $0, Z31.S    // ffdfa725
    ZUQSUB Z0.D, $0, $8, Z0.D        // 00e0e725
    ZUQSUB Z10.D, $85, $8, Z10.D     // aaeae725
    ZUQSUB Z31.D, $255, $0, Z31.D    // ffdfe725

// UQSUB   <Zd>.<T>, <Zn>.<T>, <Zm>.<T>
    ZUQSUB Z0.B, Z0.B, Z0.B       // 001c2004
    ZUQSUB Z11.B, Z12.B, Z10.B    // 6a1d2c04
    ZUQSUB Z31.B, Z31.B, Z31.B    // ff1f3f04
    ZUQSUB Z0.H, Z0.H, Z0.H       // 001c6004
    ZUQSUB Z11.H, Z12.H, Z10.H    // 6a1d6c04
    ZUQSUB Z31.H, Z31.H, Z31.H    // ff1f7f04
    ZUQSUB Z0.S, Z0.S, Z0.S       // 001ca004
    ZUQSUB Z11.S, Z12.S, Z10.S    // 6a1dac04
    ZUQSUB Z31.S, Z31.S, Z31.S    // ff1fbf04
    ZUQSUB Z0.D, Z0.D, Z0.D       // 001ce004
    ZUQSUB Z11.D, Z12.D, Z10.D    // 6a1dec04
    ZUQSUB Z31.D, Z31.D, Z31.D    // ff1fff04

// USDOT   <Zda>.S, <Zn>.B, <Zm>.B
    ZUSDOT Z0.B, Z0.B, Z0.S       // 00788044
    ZUSDOT Z11.B, Z12.B, Z10.S    // 6a798c44
    ZUSDOT Z31.B, Z31.B, Z31.S    // ff7b9f44

// USDOT   <Zda>.S, <Zn>.B, <Zm>.B[<imm>]
    ZUSDOT Z0.B, Z0.B[0], Z0.S      // 0018a044
    ZUSDOT Z11.B, Z4.B[1], Z10.S    // 6a19ac44
    ZUSDOT Z31.B, Z7.B[3], Z31.S    // ff1bbf44

// USMMLA  <Zda>.S, <Zn>.B, <Zm>.B
    ZUSMMLA Z0.B, Z0.B, Z0.S       // 00988045
    ZUSMMLA Z11.B, Z12.B, Z10.S    // 6a998c45
    ZUSMMLA Z31.B, Z31.B, Z31.S    // ff9b9f45

// UUNPKHI <Zd>.<T>, <Zn>.<Tb>
    ZUUNPKHI Z0.B, Z0.H      // 00387305
    ZUUNPKHI Z11.B, Z10.H    // 6a397305
    ZUUNPKHI Z31.B, Z31.H    // ff3b7305
    ZUUNPKHI Z0.H, Z0.S      // 0038b305
    ZUUNPKHI Z11.H, Z10.S    // 6a39b305
    ZUUNPKHI Z31.H, Z31.S    // ff3bb305
    ZUUNPKHI Z0.S, Z0.D      // 0038f305
    ZUUNPKHI Z11.S, Z10.D    // 6a39f305
    ZUUNPKHI Z31.S, Z31.D    // ff3bf305

// UUNPKLO <Zd>.<T>, <Zn>.<Tb>
    ZUUNPKLO Z0.B, Z0.H      // 00387205
    ZUUNPKLO Z11.B, Z10.H    // 6a397205
    ZUUNPKLO Z31.B, Z31.H    // ff3b7205
    ZUUNPKLO Z0.H, Z0.S      // 0038b205
    ZUUNPKLO Z11.H, Z10.S    // 6a39b205
    ZUUNPKLO Z31.H, Z31.S    // ff3bb205
    ZUUNPKLO Z0.S, Z0.D      // 0038f205
    ZUUNPKLO Z11.S, Z10.D    // 6a39f205
    ZUUNPKLO Z31.S, Z31.D    // ff3bf205

// UXTB    <Zd>.<T>, <Pg>/M, <Zn>.<T>
    ZUXTB P0.M, Z0.H, Z0.H      // 00a05104
    ZUXTB P3.M, Z12.H, Z10.H    // 8aad5104
    ZUXTB P7.M, Z31.H, Z31.H    // ffbf5104
    ZUXTB P0.M, Z0.S, Z0.S      // 00a09104
    ZUXTB P3.M, Z12.S, Z10.S    // 8aad9104
    ZUXTB P7.M, Z31.S, Z31.S    // ffbf9104
    ZUXTB P0.M, Z0.D, Z0.D      // 00a0d104
    ZUXTB P3.M, Z12.D, Z10.D    // 8aadd104
    ZUXTB P7.M, Z31.D, Z31.D    // ffbfd104

// UXTH    <Zd>.<T>, <Pg>/M, <Zn>.<T>
    ZUXTH P0.M, Z0.S, Z0.S      // 00a09304
    ZUXTH P3.M, Z12.S, Z10.S    // 8aad9304
    ZUXTH P7.M, Z31.S, Z31.S    // ffbf9304
    ZUXTH P0.M, Z0.D, Z0.D      // 00a0d304
    ZUXTH P3.M, Z12.D, Z10.D    // 8aadd304
    ZUXTH P7.M, Z31.D, Z31.D    // ffbfd304

// UXTW    <Zd>.D, <Pg>/M, <Zn>.D
    ZUXTW P0.M, Z0.D, Z0.D      // 00a0d504
    ZUXTW P3.M, Z12.D, Z10.D    // 8aadd504
    ZUXTW P7.M, Z31.D, Z31.D    // ffbfd504

// UZP1    <Pd>.<T>, <Pn>.<T>, <Pm>.<T>
    PUZP1 P0.B, P0.B, P0.B       // 00482005
    PUZP1 P6.B, P7.B, P5.B       // c5482705
    PUZP1 P15.B, P15.B, P15.B    // ef492f05
    PUZP1 P0.H, P0.H, P0.H       // 00486005
    PUZP1 P6.H, P7.H, P5.H       // c5486705
    PUZP1 P15.H, P15.H, P15.H    // ef496f05
    PUZP1 P0.S, P0.S, P0.S       // 0048a005
    PUZP1 P6.S, P7.S, P5.S       // c548a705
    PUZP1 P15.S, P15.S, P15.S    // ef49af05
    PUZP1 P0.D, P0.D, P0.D       // 0048e005
    PUZP1 P6.D, P7.D, P5.D       // c548e705
    PUZP1 P15.D, P15.D, P15.D    // ef49ef05

// UZP1    <Zd>.Q, <Zn>.Q, <Zm>.Q
    ZUZP1 Z0.Q, Z0.Q, Z0.Q       // 0008a005
    ZUZP1 Z11.Q, Z12.Q, Z10.Q    // 6a09ac05
    ZUZP1 Z31.Q, Z31.Q, Z31.Q    // ff0bbf05

// UZP1    <Zd>.<T>, <Zn>.<T>, <Zm>.<T>
    ZUZP1 Z0.B, Z0.B, Z0.B       // 00682005
    ZUZP1 Z11.B, Z12.B, Z10.B    // 6a692c05
    ZUZP1 Z31.B, Z31.B, Z31.B    // ff6b3f05
    ZUZP1 Z0.H, Z0.H, Z0.H       // 00686005
    ZUZP1 Z11.H, Z12.H, Z10.H    // 6a696c05
    ZUZP1 Z31.H, Z31.H, Z31.H    // ff6b7f05
    ZUZP1 Z0.S, Z0.S, Z0.S       // 0068a005
    ZUZP1 Z11.S, Z12.S, Z10.S    // 6a69ac05
    ZUZP1 Z31.S, Z31.S, Z31.S    // ff6bbf05
    ZUZP1 Z0.D, Z0.D, Z0.D       // 0068e005
    ZUZP1 Z11.D, Z12.D, Z10.D    // 6a69ec05
    ZUZP1 Z31.D, Z31.D, Z31.D    // ff6bff05

// UZP2    <Pd>.<T>, <Pn>.<T>, <Pm>.<T>
    PUZP2 P0.B, P0.B, P0.B       // 004c2005
    PUZP2 P6.B, P7.B, P5.B       // c54c2705
    PUZP2 P15.B, P15.B, P15.B    // ef4d2f05
    PUZP2 P0.H, P0.H, P0.H       // 004c6005
    PUZP2 P6.H, P7.H, P5.H       // c54c6705
    PUZP2 P15.H, P15.H, P15.H    // ef4d6f05
    PUZP2 P0.S, P0.S, P0.S       // 004ca005
    PUZP2 P6.S, P7.S, P5.S       // c54ca705
    PUZP2 P15.S, P15.S, P15.S    // ef4daf05
    PUZP2 P0.D, P0.D, P0.D       // 004ce005
    PUZP2 P6.D, P7.D, P5.D       // c54ce705
    PUZP2 P15.D, P15.D, P15.D    // ef4def05

// UZP2    <Zd>.Q, <Zn>.Q, <Zm>.Q
    ZUZP2 Z0.Q, Z0.Q, Z0.Q       // 000ca005
    ZUZP2 Z11.Q, Z12.Q, Z10.Q    // 6a0dac05
    ZUZP2 Z31.Q, Z31.Q, Z31.Q    // ff0fbf05

// UZP2    <Zd>.<T>, <Zn>.<T>, <Zm>.<T>
    ZUZP2 Z0.B, Z0.B, Z0.B       // 006c2005
    ZUZP2 Z11.B, Z12.B, Z10.B    // 6a6d2c05
    ZUZP2 Z31.B, Z31.B, Z31.B    // ff6f3f05
    ZUZP2 Z0.H, Z0.H, Z0.H       // 006c6005
    ZUZP2 Z11.H, Z12.H, Z10.H    // 6a6d6c05
    ZUZP2 Z31.H, Z31.H, Z31.H    // ff6f7f05
    ZUZP2 Z0.S, Z0.S, Z0.S       // 006ca005
    ZUZP2 Z11.S, Z12.S, Z10.S    // 6a6dac05
    ZUZP2 Z31.S, Z31.S, Z31.S    // ff6fbf05
    ZUZP2 Z0.D, Z0.D, Z0.D       // 006ce005
    ZUZP2 Z11.D, Z12.D, Z10.D    // 6a6dec05
    ZUZP2 Z31.D, Z31.D, Z31.D    // ff6fff05

// WHILELE <Pd>.<T>, <R><n>, <R><m>
    WHILELEW R0, R0, P0.B       // 10042025
    WHILELEW R11, R12, P5.B     // 75052c25
    WHILELEW R30, R30, P15.B    // df073e25
    WHILELE R0, R0, P0.B        // 10142025
    WHILELE R11, R12, P5.B      // 75152c25
    WHILELE R30, R30, P15.B     // df173e25
    WHILELEW R0, R0, P0.H       // 10046025
    WHILELEW R11, R12, P5.H     // 75056c25
    WHILELEW R30, R30, P15.H    // df077e25
    WHILELE R0, R0, P0.H        // 10146025
    WHILELE R11, R12, P5.H      // 75156c25
    WHILELE R30, R30, P15.H     // df177e25
    WHILELEW R0, R0, P0.S       // 1004a025
    WHILELEW R11, R12, P5.S     // 7505ac25
    WHILELEW R30, R30, P15.S    // df07be25
    WHILELE R0, R0, P0.S        // 1014a025
    WHILELE R11, R12, P5.S      // 7515ac25
    WHILELE R30, R30, P15.S     // df17be25
    WHILELEW R0, R0, P0.D       // 1004e025
    WHILELEW R11, R12, P5.D     // 7505ec25
    WHILELEW R30, R30, P15.D    // df07fe25
    WHILELE R0, R0, P0.D        // 1014e025
    WHILELE R11, R12, P5.D      // 7515ec25
    WHILELE R30, R30, P15.D     // df17fe25

// WHILELO <Pd>.<T>, <R><n>, <R><m>
    WHILELOW R0, R0, P0.B       // 000c2025
    WHILELOW R11, R12, P5.B     // 650d2c25
    WHILELOW R30, R30, P15.B    // cf0f3e25
    WHILELO R0, R0, P0.B        // 001c2025
    WHILELO R11, R12, P5.B      // 651d2c25
    WHILELO R30, R30, P15.B     // cf1f3e25
    WHILELOW R0, R0, P0.H       // 000c6025
    WHILELOW R11, R12, P5.H     // 650d6c25
    WHILELOW R30, R30, P15.H    // cf0f7e25
    WHILELO R0, R0, P0.H        // 001c6025
    WHILELO R11, R12, P5.H      // 651d6c25
    WHILELO R30, R30, P15.H     // cf1f7e25
    WHILELOW R0, R0, P0.S       // 000ca025
    WHILELOW R11, R12, P5.S     // 650dac25
    WHILELOW R30, R30, P15.S    // cf0fbe25
    WHILELO R0, R0, P0.S        // 001ca025
    WHILELO R11, R12, P5.S      // 651dac25
    WHILELO R30, R30, P15.S     // cf1fbe25
    WHILELOW R0, R0, P0.D       // 000ce025
    WHILELOW R11, R12, P5.D     // 650dec25
    WHILELOW R30, R30, P15.D    // cf0ffe25
    WHILELO R0, R0, P0.D        // 001ce025
    WHILELO R11, R12, P5.D      // 651dec25
    WHILELO R30, R30, P15.D     // cf1ffe25

// WHILELS <Pd>.<T>, <R><n>, <R><m>
    WHILELSW R0, R0, P0.B       // 100c2025
    WHILELSW R11, R12, P5.B     // 750d2c25
    WHILELSW R30, R30, P15.B    // df0f3e25
    WHILELS R0, R0, P0.B        // 101c2025
    WHILELS R11, R12, P5.B      // 751d2c25
    WHILELS R30, R30, P15.B     // df1f3e25
    WHILELSW R0, R0, P0.H       // 100c6025
    WHILELSW R11, R12, P5.H     // 750d6c25
    WHILELSW R30, R30, P15.H    // df0f7e25
    WHILELS R0, R0, P0.H        // 101c6025
    WHILELS R11, R12, P5.H      // 751d6c25
    WHILELS R30, R30, P15.H     // df1f7e25
    WHILELSW R0, R0, P0.S       // 100ca025
    WHILELSW R11, R12, P5.S     // 750dac25
    WHILELSW R30, R30, P15.S    // df0fbe25
    WHILELS R0, R0, P0.S        // 101ca025
    WHILELS R11, R12, P5.S      // 751dac25
    WHILELS R30, R30, P15.S     // df1fbe25
    WHILELSW R0, R0, P0.D       // 100ce025
    WHILELSW R11, R12, P5.D     // 750dec25
    WHILELSW R30, R30, P15.D    // df0ffe25
    WHILELS R0, R0, P0.D        // 101ce025
    WHILELS R11, R12, P5.D      // 751dec25
    WHILELS R30, R30, P15.D     // df1ffe25

// WHILELT <Pd>.<T>, <R><n>, <R><m>
    WHILELTW R0, R0, P0.B       // 00042025
    WHILELTW R11, R12, P5.B     // 65052c25
    WHILELTW R30, R30, P15.B    // cf073e25
    WHILELT R0, R0, P0.B        // 00142025
    WHILELT R11, R12, P5.B      // 65152c25
    WHILELT R30, R30, P15.B     // cf173e25
    WHILELTW R0, R0, P0.H       // 00046025
    WHILELTW R11, R12, P5.H     // 65056c25
    WHILELTW R30, R30, P15.H    // cf077e25
    WHILELT R0, R0, P0.H        // 00146025
    WHILELT R11, R12, P5.H      // 65156c25
    WHILELT R30, R30, P15.H     // cf177e25
    WHILELTW R0, R0, P0.S       // 0004a025
    WHILELTW R11, R12, P5.S     // 6505ac25
    WHILELTW R30, R30, P15.S    // cf07be25
    WHILELT R0, R0, P0.S        // 0014a025
    WHILELT R11, R12, P5.S      // 6515ac25
    WHILELT R30, R30, P15.S     // cf17be25
    WHILELTW R0, R0, P0.D       // 0004e025
    WHILELTW R11, R12, P5.D     // 6505ec25
    WHILELTW R30, R30, P15.D    // cf07fe25
    WHILELT R0, R0, P0.D        // 0014e025
    WHILELT R11, R12, P5.D      // 6515ec25
    WHILELT R30, R30, P15.D     // cf17fe25

// WRFFR   <Pn>.B
    PWRFFR P0.B     // 00902825
    PWRFFR P5.B     // a0902825
    PWRFFR P15.B    // e0912825

// ZIP1    <Pd>.<T>, <Pn>.<T>, <Pm>.<T>
    PZIP1 P0.B, P0.B, P0.B       // 00402005
    PZIP1 P6.B, P7.B, P5.B       // c5402705
    PZIP1 P15.B, P15.B, P15.B    // ef412f05
    PZIP1 P0.H, P0.H, P0.H       // 00406005
    PZIP1 P6.H, P7.H, P5.H       // c5406705
    PZIP1 P15.H, P15.H, P15.H    // ef416f05
    PZIP1 P0.S, P0.S, P0.S       // 0040a005
    PZIP1 P6.S, P7.S, P5.S       // c540a705
    PZIP1 P15.S, P15.S, P15.S    // ef41af05
    PZIP1 P0.D, P0.D, P0.D       // 0040e005
    PZIP1 P6.D, P7.D, P5.D       // c540e705
    PZIP1 P15.D, P15.D, P15.D    // ef41ef05

// ZIP1    <Zd>.Q, <Zn>.Q, <Zm>.Q
    ZZIP1 Z0.Q, Z0.Q, Z0.Q       // 0000a005
    ZZIP1 Z11.Q, Z12.Q, Z10.Q    // 6a01ac05
    ZZIP1 Z31.Q, Z31.Q, Z31.Q    // ff03bf05

// ZIP1    <Zd>.<T>, <Zn>.<T>, <Zm>.<T>
    ZZIP1 Z0.B, Z0.B, Z0.B       // 00602005
    ZZIP1 Z11.B, Z12.B, Z10.B    // 6a612c05
    ZZIP1 Z31.B, Z31.B, Z31.B    // ff633f05
    ZZIP1 Z0.H, Z0.H, Z0.H       // 00606005
    ZZIP1 Z11.H, Z12.H, Z10.H    // 6a616c05
    ZZIP1 Z31.H, Z31.H, Z31.H    // ff637f05
    ZZIP1 Z0.S, Z0.S, Z0.S       // 0060a005
    ZZIP1 Z11.S, Z12.S, Z10.S    // 6a61ac05
    ZZIP1 Z31.S, Z31.S, Z31.S    // ff63bf05
    ZZIP1 Z0.D, Z0.D, Z0.D       // 0060e005
    ZZIP1 Z11.D, Z12.D, Z10.D    // 6a61ec05
    ZZIP1 Z31.D, Z31.D, Z31.D    // ff63ff05

// ZIP2    <Pd>.<T>, <Pn>.<T>, <Pm>.<T>
    PZIP2 P0.B, P0.B, P0.B       // 00442005
    PZIP2 P6.B, P7.B, P5.B       // c5442705
    PZIP2 P15.B, P15.B, P15.B    // ef452f05
    PZIP2 P0.H, P0.H, P0.H       // 00446005
    PZIP2 P6.H, P7.H, P5.H       // c5446705
    PZIP2 P15.H, P15.H, P15.H    // ef456f05
    PZIP2 P0.S, P0.S, P0.S       // 0044a005
    PZIP2 P6.S, P7.S, P5.S       // c544a705
    PZIP2 P15.S, P15.S, P15.S    // ef45af05
    PZIP2 P0.D, P0.D, P0.D       // 0044e005
    PZIP2 P6.D, P7.D, P5.D       // c544e705
    PZIP2 P15.D, P15.D, P15.D    // ef45ef05

// ZIP2    <Zd>.Q, <Zn>.Q, <Zm>.Q
    ZZIP2 Z0.Q, Z0.Q, Z0.Q       // 0004a005
    ZZIP2 Z11.Q, Z12.Q, Z10.Q    // 6a05ac05
    ZZIP2 Z31.Q, Z31.Q, Z31.Q    // ff07bf05

// ZIP2    <Zd>.<T>, <Zn>.<T>, <Zm>.<T>
    ZZIP2 Z0.B, Z0.B, Z0.B       // 00642005
    ZZIP2 Z11.B, Z12.B, Z10.B    // 6a652c05
    ZZIP2 Z31.B, Z31.B, Z31.B    // ff673f05
    ZZIP2 Z0.H, Z0.H, Z0.H       // 00646005
    ZZIP2 Z11.H, Z12.H, Z10.H    // 6a656c05
    ZZIP2 Z31.H, Z31.H, Z31.H    // ff677f05
    ZZIP2 Z0.S, Z0.S, Z0.S       // 0064a005
    ZZIP2 Z11.S, Z12.S, Z10.S    // 6a65ac05
    ZZIP2 Z31.S, Z31.S, Z31.S    // ff67bf05
    ZZIP2 Z0.D, Z0.D, Z0.D       // 0064e005
    ZZIP2 Z11.D, Z12.D, Z10.D    // 6a65ec05
    ZZIP2 Z31.D, Z31.D, Z31.D    // ff67ff05
    RET

