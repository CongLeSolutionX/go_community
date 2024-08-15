// Copyright 2024 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

TEXT svetest(SB),$0
    ZABS P0.M, Z0.B, Z0.B                            // 00a01604
    ZABS P3.M, Z12.B, Z10.B                          // 8aad1604
    ZABS P7.M, Z31.B, Z31.B                          // ffbf1604
    ZABS P0.M, Z0.H, Z0.H                            // 00a05604
    ZABS P3.M, Z12.H, Z10.H                          // 8aad5604
    ZABS P7.M, Z31.H, Z31.H                          // ffbf5604
    ZABS P0.M, Z0.S, Z0.S                            // 00a09604
    ZABS P3.M, Z12.S, Z10.S                          // 8aad9604
    ZABS P7.M, Z31.S, Z31.S                          // ffbf9604
    ZABS P0.M, Z0.D, Z0.D                            // 00a0d604
    ZABS P3.M, Z12.D, Z10.D                          // 8aadd604
    ZABS P7.M, Z31.D, Z31.D                          // ffbfd604

    ZBFDOT Z0.H, Z0.H, Z0.S                          // 00806064
    ZBFDOT Z11.H, Z12.H, Z10.S                       // 6a816c64
    ZBFDOT Z31.H, Z31.H, Z31.S                       // ff837f64
    
    ZBFDOT Z0.H, Z0.H[0], Z0.S                       // 00406064
    ZBFDOT Z11.H, Z4.H[1], Z10.S                     // 6a416c64
    ZBFDOT Z31.H, Z7.H[3], Z31.S                     // ff437f64

    ZBFMLALT Z0.H, Z0.H, Z0.S                        // 0084e064
    ZBFMLALT Z11.H, Z12.H, Z10.S                     // 6a85ec64
    ZBFMLALT Z31.H, Z31.H, Z31.S                     // ff87ff64
    ZBFMLALT Z0.H, Z0.H[0], Z0.S                     // 0044e064
    ZBFMLALT Z11.H, Z4.H[2], Z10.S                   // 6a45ec64
    ZBFMLALT Z31.H, Z7.H[7], Z31.S                   // ff4fff64
    ZBFMLALB Z0.H, Z0.H, Z0.S                        // 0080e064
    ZBFMLALB Z11.H, Z12.H, Z10.S                     // 6a81ec64
    ZBFMLALB Z31.H, Z31.H, Z31.S                     // ff83ff64
    ZBFMLALB Z0.H, Z0.H[0], Z0.S                     // 0040e064
    ZBFMLALB Z11.H, Z4.H[2], Z10.S                   // 6a41ec64
    ZBFMLALB Z31.H, Z7.H[7], Z31.S                   // ff4bff64
    ZBFMMLA Z0.H, Z0.H, Z0.S                         // 00e46064
    ZBFMMLA Z11.H, Z12.H, Z10.S                      // 6ae56c64
    ZBFMMLA Z31.H, Z31.H, Z31.S                      // ffe77f64

    ZCNOT P0.M, Z0.B, Z0.B                           // 00a01b04
    ZCNOT P3.M, Z12.B, Z10.B                         // 8aad1b04
    ZCNOT P7.M, Z31.B, Z31.B                         // ffbf1b04
    ZCNOT P0.M, Z0.H, Z0.H                           // 00a05b04
    ZCNOT P3.M, Z12.H, Z10.H                         // 8aad5b04
    ZCNOT P7.M, Z31.H, Z31.H                         // ffbf5b04
    ZCNOT P0.M, Z0.S, Z0.S                           // 00a09b04
    ZCNOT P3.M, Z12.S, Z10.S                         // 8aad9b04
    ZCNOT P7.M, Z31.S, Z31.S                         // ffbf9b04
    ZCNOT P0.M, Z0.D, Z0.D                           // 00a0db04
    ZCNOT P3.M, Z12.D, Z10.D                         // 8aaddb04
    ZCNOT P7.M, Z31.D, Z31.D                         // ffbfdb04

    ZCNT P0.M, Z0.B, Z0.B                            // 00a01a04
    ZCNT P3.M, Z12.B, Z10.B                          // 8aad1a04
    ZCNT P7.M, Z31.B, Z31.B                          // ffbf1a04
    ZCNT P0.M, Z0.H, Z0.H                            // 00a05a04
    ZCNT P3.M, Z12.H, Z10.H                          // 8aad5a04
    ZCNT P7.M, Z31.H, Z31.H                          // ffbf5a04
    ZCNT P0.M, Z0.S, Z0.S                            // 00a09a04
    ZCNT P3.M, Z12.S, Z10.S                          // 8aad9a04
    ZCNT P7.M, Z31.S, Z31.S                          // ffbf9a04
    ZCNT P0.M, Z0.D, Z0.D                            // 00a0da04
    ZCNT P3.M, Z12.D, Z10.D                          // 8aadda04
    ZCNT P7.M, Z31.D, Z31.D                          // ffbfda04

    ZFABS P0.M, Z0.H, Z0.H                           // 00a05c04
    ZFABS P3.M, Z12.H, Z10.H                         // 8aad5c04
    ZFABS P7.M, Z31.H, Z31.H                         // ffbf5c04
    ZFABS P0.M, Z0.S, Z0.S                           // 00a09c04
    ZFABS P3.M, Z12.S, Z10.S                         // 8aad9c04
    ZFABS P7.M, Z31.S, Z31.S                         // ffbf9c04
    ZFABS P0.M, Z0.D, Z0.D                           // 00a0dc04
    ZFABS P3.M, Z12.D, Z10.D                         // 8aaddc04
    ZFABS P7.M, Z31.D, Z31.D                         // ffbfdc04

    ZFNEG P0.M, Z0.H, Z0.H                           // 00a05d04
    ZFNEG P3.M, Z12.H, Z10.H                         // 8aad5d04
    ZFNEG P7.M, Z31.H, Z31.H                         // ffbf5d04
    ZFNEG P0.M, Z0.S, Z0.S                           // 00a09d04
    ZFNEG P3.M, Z12.S, Z10.S                         // 8aad9d04
    ZFNEG P7.M, Z31.S, Z31.S                         // ffbf9d04
    ZFNEG P0.M, Z0.D, Z0.D                           // 00a0dd04
    ZFNEG P3.M, Z12.D, Z10.D                         // 8aaddd04
    ZFNEG P7.M, Z31.D, Z31.D                         // ffbfdd04

    ZFRINTA P0.M, Z0.H, Z0.H                         // 00a04465
    ZFRINTA P3.M, Z12.H, Z10.H                       // 8aad4465
    ZFRINTA P7.M, Z31.H, Z31.H                       // ffbf4465
    ZFRINTA P0.M, Z0.S, Z0.S                         // 00a08465
    ZFRINTA P3.M, Z12.S, Z10.S                       // 8aad8465
    ZFRINTA P7.M, Z31.S, Z31.S                       // ffbf8465
    ZFRINTA P0.M, Z0.D, Z0.D                         // 00a0c465
    ZFRINTA P3.M, Z12.D, Z10.D                       // 8aadc465
    ZFRINTA P7.M, Z31.D, Z31.D                       // ffbfc465

    ZFRINTI P0.M, Z0.H, Z0.H                         // 00a04765
    ZFRINTI P3.M, Z12.H, Z10.H                       // 8aad4765
    ZFRINTI P7.M, Z31.H, Z31.H                       // ffbf4765
    ZFRINTI P0.M, Z0.S, Z0.S                         // 00a08765
    ZFRINTI P3.M, Z12.S, Z10.S                       // 8aad8765
    ZFRINTI P7.M, Z31.S, Z31.S                       // ffbf8765
    ZFRINTI P0.M, Z0.D, Z0.D                         // 00a0c765
    ZFRINTI P3.M, Z12.D, Z10.D                       // 8aadc765
    ZFRINTI P7.M, Z31.D, Z31.D                       // ffbfc765

    ZFRINTM P0.M, Z0.H, Z0.H                         // 00a04265
    ZFRINTM P3.M, Z12.H, Z10.H                       // 8aad4265
    ZFRINTM P7.M, Z31.H, Z31.H                       // ffbf4265
    ZFRINTM P0.M, Z0.S, Z0.S                         // 00a08265
    ZFRINTM P3.M, Z12.S, Z10.S                       // 8aad8265
    ZFRINTM P7.M, Z31.S, Z31.S                       // ffbf8265
    ZFRINTM P0.M, Z0.D, Z0.D                         // 00a0c265
    ZFRINTM P3.M, Z12.D, Z10.D                       // 8aadc265
    ZFRINTM P7.M, Z31.D, Z31.D                       // ffbfc265

    ZFRINTN P0.M, Z0.H, Z0.H                         // 00a04065
    ZFRINTN P3.M, Z12.H, Z10.H                       // 8aad4065
    ZFRINTN P7.M, Z31.H, Z31.H                       // ffbf4065
    ZFRINTN P0.M, Z0.S, Z0.S                         // 00a08065
    ZFRINTN P3.M, Z12.S, Z10.S                       // 8aad8065
    ZFRINTN P7.M, Z31.S, Z31.S                       // ffbf8065
    ZFRINTN P0.M, Z0.D, Z0.D                         // 00a0c065
    ZFRINTN P3.M, Z12.D, Z10.D                       // 8aadc065
    ZFRINTN P7.M, Z31.D, Z31.D                       // ffbfc065

    ZFRINTP P0.M, Z0.H, Z0.H                         // 00a04165
    ZFRINTP P3.M, Z12.H, Z10.H                       // 8aad4165
    ZFRINTP P7.M, Z31.H, Z31.H                       // ffbf4165
    ZFRINTP P0.M, Z0.S, Z0.S                         // 00a08165
    ZFRINTP P3.M, Z12.S, Z10.S                       // 8aad8165
    ZFRINTP P7.M, Z31.S, Z31.S                       // ffbf8165
    ZFRINTP P0.M, Z0.D, Z0.D                         // 00a0c165
    ZFRINTP P3.M, Z12.D, Z10.D                       // 8aadc165
    ZFRINTP P7.M, Z31.D, Z31.D                       // ffbfc165

    ZFRINTX P0.M, Z0.H, Z0.H                         // 00a04665
    ZFRINTX P3.M, Z12.H, Z10.H                       // 8aad4665
    ZFRINTX P7.M, Z31.H, Z31.H                       // ffbf4665
    ZFRINTX P0.M, Z0.S, Z0.S                         // 00a08665
    ZFRINTX P3.M, Z12.S, Z10.S                       // 8aad8665
    ZFRINTX P7.M, Z31.S, Z31.S                       // ffbf8665
    ZFRINTX P0.M, Z0.D, Z0.D                         // 00a0c665
    ZFRINTX P3.M, Z12.D, Z10.D                       // 8aadc665
    ZFRINTX P7.M, Z31.D, Z31.D                       // ffbfc665

    ZFRINTZ P0.M, Z0.H, Z0.H                         // 00a04365
    ZFRINTZ P3.M, Z12.H, Z10.H                       // 8aad4365
    ZFRINTZ P7.M, Z31.H, Z31.H                       // ffbf4365
    ZFRINTZ P0.M, Z0.S, Z0.S                         // 00a08365
    ZFRINTZ P3.M, Z12.S, Z10.S                       // 8aad8365
    ZFRINTZ P7.M, Z31.S, Z31.S                       // ffbf8365
    ZFRINTZ P0.M, Z0.D, Z0.D                         // 00a0c365
    ZFRINTZ P3.M, Z12.D, Z10.D                       // 8aadc365
    ZFRINTZ P7.M, Z31.D, Z31.D                       // ffbfc365

    ZFSQRT P0.M, Z0.H, Z0.H                          // 00a04d65
    ZFSQRT P3.M, Z12.H, Z10.H                        // 8aad4d65
    ZFSQRT P7.M, Z31.H, Z31.H                        // ffbf4d65
    ZFSQRT P0.M, Z0.S, Z0.S                          // 00a08d65
    ZFSQRT P3.M, Z12.S, Z10.S                        // 8aad8d65
    ZFSQRT P7.M, Z31.S, Z31.S                        // ffbf8d65
    ZFSQRT P0.M, Z0.D, Z0.D                          // 00a0cd65
    ZFSQRT P3.M, Z12.D, Z10.D                        // 8aadcd65
    ZFSQRT P7.M, Z31.D, Z31.D                        // ffbfcd65

    ZNOT P0.M, Z0.B, Z0.B                            // 00a01e04
    ZNOT P3.M, Z12.B, Z10.B                          // 8aad1e04
    ZNOT P7.M, Z31.B, Z31.B                          // ffbf1e04
    ZNOT P0.M, Z0.H, Z0.H                            // 00a05e04
    ZNOT P3.M, Z12.H, Z10.H                          // 8aad5e04
    ZNOT P7.M, Z31.H, Z31.H                          // ffbf5e04
    ZNOT P0.M, Z0.S, Z0.S                            // 00a09e04
    ZNOT P3.M, Z12.S, Z10.S                          // 8aad9e04
    ZNOT P7.M, Z31.S, Z31.S                          // ffbf9e04
    ZNOT P0.M, Z0.D, Z0.D                            // 00a0de04
    ZNOT P3.M, Z12.D, Z10.D                          // 8aadde04
    ZNOT P7.M, Z31.D, Z31.D                          // ffbfde04

    ZREVB P0.M, Z0.H, Z0.H                           // 00806405
    ZREVB P3.M, Z12.H, Z10.H                         // 8a8d6405
    ZREVB P7.M, Z31.H, Z31.H                         // ff9f6405
    ZREVB P0.M, Z0.S, Z0.S                           // 0080a405
    ZREVB P3.M, Z12.S, Z10.S                         // 8a8da405
    ZREVB P7.M, Z31.S, Z31.S                         // ff9fa405
    ZREVB P0.M, Z0.D, Z0.D                           // 0080e405
    ZREVB P3.M, Z12.D, Z10.D                         // 8a8de405
    ZREVB P7.M, Z31.D, Z31.D                         // ff9fe405

    ZREVH P0.M, Z0.S, Z0.S                           // 0080a505
    ZREVH P3.M, Z12.S, Z10.S                         // 8a8da505
    ZREVH P7.M, Z31.S, Z31.S                         // ff9fa505
    ZREVH P0.M, Z0.D, Z0.D                           // 0080e505
    ZREVH P3.M, Z12.D, Z10.D                         // 8a8de505
    ZREVH P7.M, Z31.D, Z31.D                         // ff9fe505

// ADD     <Zdn>.<T>, <Pg>/M, <Zdn>.<T>, <Zm>.<T>
    ZADD P0.M, Z0.B, Z0.B, Z0.B                      // 00000004
    ZADD P3.M, Z10.B, Z12.B, Z10.B                   // 8a0d0004
    ZADD P7.M, Z31.B, Z31.B, Z31.B                   // ff1f0004
    ZADD P0.M, Z0.H, Z0.H, Z0.H                      // 00004004
    ZADD P3.M, Z10.H, Z12.H, Z10.H                   // 8a0d4004
    ZADD P7.M, Z31.H, Z31.H, Z31.H                   // ff1f4004
    ZADD P0.M, Z0.S, Z0.S, Z0.S                      // 00008004
    ZADD P3.M, Z10.S, Z12.S, Z10.S                   // 8a0d8004
    ZADD P7.M, Z31.S, Z31.S, Z31.S                   // ff1f8004
    ZADD P0.M, Z0.D, Z0.D, Z0.D                      // 0000c004
    ZADD P3.M, Z10.D, Z12.D, Z10.D                   // 8a0dc004
    ZADD P7.M, Z31.D, Z31.D, Z31.D                   // ff1fc004

// AND     <Zdn>.<T>, <Pg>/M, <Zdn>.<T>, <Zm>.<T>
    ZAND P0.M, Z0.B, Z0.B, Z0.B                      // 00001a04
    ZAND P3.M, Z10.B, Z12.B, Z10.B                   // 8a0d1a04
    ZAND P7.M, Z31.B, Z31.B, Z31.B                   // ff1f1a04
    ZAND P0.M, Z0.H, Z0.H, Z0.H                      // 00005a04
    ZAND P3.M, Z10.H, Z12.H, Z10.H                   // 8a0d5a04
    ZAND P7.M, Z31.H, Z31.H, Z31.H                   // ff1f5a04
    ZAND P0.M, Z0.S, Z0.S, Z0.S                      // 00009a04
    ZAND P3.M, Z10.S, Z12.S, Z10.S                   // 8a0d9a04
    ZAND P7.M, Z31.S, Z31.S, Z31.S                   // ff1f9a04
    ZAND P0.M, Z0.D, Z0.D, Z0.D                      // 0000da04
    ZAND P3.M, Z10.D, Z12.D, Z10.D                   // 8a0dda04
    ZAND P7.M, Z31.D, Z31.D, Z31.D                   // ff1fda04

// ASR     <Zdn>.<T>, <Pg>/M, <Zdn>.<T>, <Zm>.<T>
    ZASR P0.M, Z0.B, Z0.B, Z0.B                      // 00801004
    ZASR P3.M, Z10.B, Z12.B, Z10.B                   // 8a8d1004
    ZASR P7.M, Z31.B, Z31.B, Z31.B                   // ff9f1004
    ZASR P0.M, Z0.H, Z0.H, Z0.H                      // 00805004
    ZASR P3.M, Z10.H, Z12.H, Z10.H                   // 8a8d5004
    ZASR P7.M, Z31.H, Z31.H, Z31.H                   // ff9f5004
    ZASR P0.M, Z0.S, Z0.S, Z0.S                      // 00809004
    ZASR P3.M, Z10.S, Z12.S, Z10.S                   // 8a8d9004
    ZASR P7.M, Z31.S, Z31.S, Z31.S                   // ff9f9004
    ZASR P0.M, Z0.D, Z0.D, Z0.D                      // 0080d004
    ZASR P3.M, Z10.D, Z12.D, Z10.D                   // 8a8dd004
    ZASR P7.M, Z31.D, Z31.D, Z31.D                   // ff9fd004

// ASRR    <Zdn>.<T>, <Pg>/M, <Zdn>.<T>, <Zm>.<T>
    ZASRR P0.M, Z0.B, Z0.B, Z0.B                     // 00801404
    ZASRR P3.M, Z10.B, Z12.B, Z10.B                  // 8a8d1404
    ZASRR P7.M, Z31.B, Z31.B, Z31.B                  // ff9f1404
    ZASRR P0.M, Z0.H, Z0.H, Z0.H                     // 00805404
    ZASRR P3.M, Z10.H, Z12.H, Z10.H                  // 8a8d5404
    ZASRR P7.M, Z31.H, Z31.H, Z31.H                  // ff9f5404
    ZASRR P0.M, Z0.S, Z0.S, Z0.S                     // 00809404
    ZASRR P3.M, Z10.S, Z12.S, Z10.S                  // 8a8d9404
    ZASRR P7.M, Z31.S, Z31.S, Z31.S                  // ff9f9404
    ZASRR P0.M, Z0.D, Z0.D, Z0.D                     // 0080d404
    ZASRR P3.M, Z10.D, Z12.D, Z10.D                  // 8a8dd404
    ZASRR P7.M, Z31.D, Z31.D, Z31.D                  // ff9fd404

// BIC     <Zdn>.<T>, <Pg>/M, <Zdn>.<T>, <Zm>.<T>
    ZBIC P0.M, Z0.B, Z0.B, Z0.B                      // 00001b04
    ZBIC P3.M, Z10.B, Z12.B, Z10.B                   // 8a0d1b04
    ZBIC P7.M, Z31.B, Z31.B, Z31.B                   // ff1f1b04
    ZBIC P0.M, Z0.H, Z0.H, Z0.H                      // 00005b04
    ZBIC P3.M, Z10.H, Z12.H, Z10.H                   // 8a0d5b04
    ZBIC P7.M, Z31.H, Z31.H, Z31.H                   // ff1f5b04
    ZBIC P0.M, Z0.S, Z0.S, Z0.S                      // 00009b04
    ZBIC P3.M, Z10.S, Z12.S, Z10.S                   // 8a0d9b04
    ZBIC P7.M, Z31.S, Z31.S, Z31.S                   // ff1f9b04
    ZBIC P0.M, Z0.D, Z0.D, Z0.D                      // 0000db04
    ZBIC P3.M, Z10.D, Z12.D, Z10.D                   // 8a0ddb04
    ZBIC P7.M, Z31.D, Z31.D, Z31.D                   // ff1fdb04

// EOR     <Zdn>.<T>, <Pg>/M, <Zdn>.<T>, <Zm>.<T>
    ZEOR P0.M, Z0.B, Z0.B, Z0.B                      // 00001904
    ZEOR P3.M, Z10.B, Z12.B, Z10.B                   // 8a0d1904
    ZEOR P7.M, Z31.B, Z31.B, Z31.B                   // ff1f1904
    ZEOR P0.M, Z0.H, Z0.H, Z0.H                      // 00005904
    ZEOR P3.M, Z10.H, Z12.H, Z10.H                   // 8a0d5904
    ZEOR P7.M, Z31.H, Z31.H, Z31.H                   // ff1f5904
    ZEOR P0.M, Z0.S, Z0.S, Z0.S                      // 00009904
    ZEOR P3.M, Z10.S, Z12.S, Z10.S                   // 8a0d9904
    ZEOR P7.M, Z31.S, Z31.S, Z31.S                   // ff1f9904
    ZEOR P0.M, Z0.D, Z0.D, Z0.D                      // 0000d904
    ZEOR P3.M, Z10.D, Z12.D, Z10.D                   // 8a0dd904
    ZEOR P7.M, Z31.D, Z31.D, Z31.D                   // ff1fd904

// FABD    <Zdn>.<T>, <Pg>/M, <Zdn>.<T>, <Zm>.<T>
    ZFABD P0.M, Z0.H, Z0.H, Z0.H                     // 00804865
    ZFABD P3.M, Z10.H, Z12.H, Z10.H                  // 8a8d4865
    ZFABD P7.M, Z31.H, Z31.H, Z31.H                  // ff9f4865
    ZFABD P0.M, Z0.S, Z0.S, Z0.S                     // 00808865
    ZFABD P3.M, Z10.S, Z12.S, Z10.S                  // 8a8d8865
    ZFABD P7.M, Z31.S, Z31.S, Z31.S                  // ff9f8865
    ZFABD P0.M, Z0.D, Z0.D, Z0.D                     // 0080c865
    ZFABD P3.M, Z10.D, Z12.D, Z10.D                  // 8a8dc865
    ZFABD P7.M, Z31.D, Z31.D, Z31.D                  // ff9fc865

// FADD    <Zdn>.<T>, <Pg>/M, <Zdn>.<T>, <Zm>.<T>
    ZFADD P0.M, Z0.H, Z0.H, Z0.H                     // 00804065
    ZFADD P3.M, Z10.H, Z12.H, Z10.H                  // 8a8d4065
    ZFADD P7.M, Z31.H, Z31.H, Z31.H                  // ff9f4065
    ZFADD P0.M, Z0.S, Z0.S, Z0.S                     // 00808065
    ZFADD P3.M, Z10.S, Z12.S, Z10.S                  // 8a8d8065
    ZFADD P7.M, Z31.S, Z31.S, Z31.S                  // ff9f8065
    ZFADD P0.M, Z0.D, Z0.D, Z0.D                     // 0080c065
    ZFADD P3.M, Z10.D, Z12.D, Z10.D                  // 8a8dc065
    ZFADD P7.M, Z31.D, Z31.D, Z31.D                  // ff9fc065

// FDIV    <Zdn>.<T>, <Pg>/M, <Zdn>.<T>, <Zm>.<T>
    ZFDIV P0.M, Z0.H, Z0.H, Z0.H                     // 00804d65
    ZFDIV P3.M, Z10.H, Z12.H, Z10.H                  // 8a8d4d65
    ZFDIV P7.M, Z31.H, Z31.H, Z31.H                  // ff9f4d65
    ZFDIV P0.M, Z0.S, Z0.S, Z0.S                     // 00808d65
    ZFDIV P3.M, Z10.S, Z12.S, Z10.S                  // 8a8d8d65
    ZFDIV P7.M, Z31.S, Z31.S, Z31.S                  // ff9f8d65
    ZFDIV P0.M, Z0.D, Z0.D, Z0.D                     // 0080cd65
    ZFDIV P3.M, Z10.D, Z12.D, Z10.D                  // 8a8dcd65
    ZFDIV P7.M, Z31.D, Z31.D, Z31.D                  // ff9fcd65

// FDIVR   <Zdn>.<T>, <Pg>/M, <Zdn>.<T>, <Zm>.<T>
    ZFDIVR P0.M, Z0.H, Z0.H, Z0.H                    // 00804c65
    ZFDIVR P3.M, Z10.H, Z12.H, Z10.H                 // 8a8d4c65
    ZFDIVR P7.M, Z31.H, Z31.H, Z31.H                 // ff9f4c65
    ZFDIVR P0.M, Z0.S, Z0.S, Z0.S                    // 00808c65
    ZFDIVR P3.M, Z10.S, Z12.S, Z10.S                 // 8a8d8c65
    ZFDIVR P7.M, Z31.S, Z31.S, Z31.S                 // ff9f8c65
    ZFDIVR P0.M, Z0.D, Z0.D, Z0.D                    // 0080cc65
    ZFDIVR P3.M, Z10.D, Z12.D, Z10.D                 // 8a8dcc65
    ZFDIVR P7.M, Z31.D, Z31.D, Z31.D                 // ff9fcc65

// FMAX    <Zdn>.<T>, <Pg>/M, <Zdn>.<T>, <Zm>.<T>
    ZFMAX P0.M, Z0.H, Z0.H, Z0.H                     // 00804665
    ZFMAX P3.M, Z10.H, Z12.H, Z10.H                  // 8a8d4665
    ZFMAX P7.M, Z31.H, Z31.H, Z31.H                  // ff9f4665
    ZFMAX P0.M, Z0.S, Z0.S, Z0.S                     // 00808665
    ZFMAX P3.M, Z10.S, Z12.S, Z10.S                  // 8a8d8665
    ZFMAX P7.M, Z31.S, Z31.S, Z31.S                  // ff9f8665
    ZFMAX P0.M, Z0.D, Z0.D, Z0.D                     // 0080c665
    ZFMAX P3.M, Z10.D, Z12.D, Z10.D                  // 8a8dc665
    ZFMAX P7.M, Z31.D, Z31.D, Z31.D                  // ff9fc665

// FMAXNM  <Zdn>.<T>, <Pg>/M, <Zdn>.<T>, <Zm>.<T>
    ZFMAXNM P0.M, Z0.H, Z0.H, Z0.H                   // 00804465
    ZFMAXNM P3.M, Z10.H, Z12.H, Z10.H                // 8a8d4465
    ZFMAXNM P7.M, Z31.H, Z31.H, Z31.H                // ff9f4465
    ZFMAXNM P0.M, Z0.S, Z0.S, Z0.S                   // 00808465
    ZFMAXNM P3.M, Z10.S, Z12.S, Z10.S                // 8a8d8465
    ZFMAXNM P7.M, Z31.S, Z31.S, Z31.S                // ff9f8465
    ZFMAXNM P0.M, Z0.D, Z0.D, Z0.D                   // 0080c465
    ZFMAXNM P3.M, Z10.D, Z12.D, Z10.D                // 8a8dc465
    ZFMAXNM P7.M, Z31.D, Z31.D, Z31.D                // ff9fc465

// FMIN    <Zdn>.<T>, <Pg>/M, <Zdn>.<T>, <Zm>.<T>
    ZFMIN P0.M, Z0.H, Z0.H, Z0.H                     // 00804765
    ZFMIN P3.M, Z10.H, Z12.H, Z10.H                  // 8a8d4765
    ZFMIN P7.M, Z31.H, Z31.H, Z31.H                  // ff9f4765
    ZFMIN P0.M, Z0.S, Z0.S, Z0.S                     // 00808765
    ZFMIN P3.M, Z10.S, Z12.S, Z10.S                  // 8a8d8765
    ZFMIN P7.M, Z31.S, Z31.S, Z31.S                  // ff9f8765
    ZFMIN P0.M, Z0.D, Z0.D, Z0.D                     // 0080c765
    ZFMIN P3.M, Z10.D, Z12.D, Z10.D                  // 8a8dc765
    ZFMIN P7.M, Z31.D, Z31.D, Z31.D                  // ff9fc765

// FMINNM  <Zdn>.<T>, <Pg>/M, <Zdn>.<T>, <Zm>.<T>
    ZFMINNM P0.M, Z0.H, Z0.H, Z0.H                   // 00804565
    ZFMINNM P3.M, Z10.H, Z12.H, Z10.H                // 8a8d4565
    ZFMINNM P7.M, Z31.H, Z31.H, Z31.H                // ff9f4565
    ZFMINNM P0.M, Z0.S, Z0.S, Z0.S                   // 00808565
    ZFMINNM P3.M, Z10.S, Z12.S, Z10.S                // 8a8d8565
    ZFMINNM P7.M, Z31.S, Z31.S, Z31.S                // ff9f8565
    ZFMINNM P0.M, Z0.D, Z0.D, Z0.D                   // 0080c565
    ZFMINNM P3.M, Z10.D, Z12.D, Z10.D                // 8a8dc565
    ZFMINNM P7.M, Z31.D, Z31.D, Z31.D                // ff9fc565

// FMUL    <Zdn>.<T>, <Pg>/M, <Zdn>.<T>, <Zm>.<T>
    ZFMUL P0.M, Z0.H, Z0.H, Z0.H                     // 00804265
    ZFMUL P3.M, Z10.H, Z12.H, Z10.H                  // 8a8d4265
    ZFMUL P7.M, Z31.H, Z31.H, Z31.H                  // ff9f4265
    ZFMUL P0.M, Z0.S, Z0.S, Z0.S                     // 00808265
    ZFMUL P3.M, Z10.S, Z12.S, Z10.S                  // 8a8d8265
    ZFMUL P7.M, Z31.S, Z31.S, Z31.S                  // ff9f8265
    ZFMUL P0.M, Z0.D, Z0.D, Z0.D                     // 0080c265
    ZFMUL P3.M, Z10.D, Z12.D, Z10.D                  // 8a8dc265
    ZFMUL P7.M, Z31.D, Z31.D, Z31.D                  // ff9fc265

// FMULX   <Zdn>.<T>, <Pg>/M, <Zdn>.<T>, <Zm>.<T>
    ZFMULX P0.M, Z0.H, Z0.H, Z0.H                    // 00804a65
    ZFMULX P3.M, Z10.H, Z12.H, Z10.H                 // 8a8d4a65
    ZFMULX P7.M, Z31.H, Z31.H, Z31.H                 // ff9f4a65
    ZFMULX P0.M, Z0.S, Z0.S, Z0.S                    // 00808a65
    ZFMULX P3.M, Z10.S, Z12.S, Z10.S                 // 8a8d8a65
    ZFMULX P7.M, Z31.S, Z31.S, Z31.S                 // ff9f8a65
    ZFMULX P0.M, Z0.D, Z0.D, Z0.D                    // 0080ca65
    ZFMULX P3.M, Z10.D, Z12.D, Z10.D                 // 8a8dca65
    ZFMULX P7.M, Z31.D, Z31.D, Z31.D                 // ff9fca65

// FSCALE  <Zdn>.<T>, <Pg>/M, <Zdn>.<T>, <Zm>.<T>
    ZFSCALE P0.M, Z0.H, Z0.H, Z0.H                   // 00804965
    ZFSCALE P3.M, Z10.H, Z12.H, Z10.H                // 8a8d4965
    ZFSCALE P7.M, Z31.H, Z31.H, Z31.H                // ff9f4965
    ZFSCALE P0.M, Z0.S, Z0.S, Z0.S                   // 00808965
    ZFSCALE P3.M, Z10.S, Z12.S, Z10.S                // 8a8d8965
    ZFSCALE P7.M, Z31.S, Z31.S, Z31.S                // ff9f8965
    ZFSCALE P0.M, Z0.D, Z0.D, Z0.D                   // 0080c965
    ZFSCALE P3.M, Z10.D, Z12.D, Z10.D                // 8a8dc965
    ZFSCALE P7.M, Z31.D, Z31.D, Z31.D                // ff9fc965

// FSUB    <Zdn>.<T>, <Pg>/M, <Zdn>.<T>, <Zm>.<T>
    ZFSUB P0.M, Z0.H, Z0.H, Z0.H                     // 00804165
    ZFSUB P3.M, Z10.H, Z12.H, Z10.H                  // 8a8d4165
    ZFSUB P7.M, Z31.H, Z31.H, Z31.H                  // ff9f4165
    ZFSUB P0.M, Z0.S, Z0.S, Z0.S                     // 00808165
    ZFSUB P3.M, Z10.S, Z12.S, Z10.S                  // 8a8d8165
    ZFSUB P7.M, Z31.S, Z31.S, Z31.S                  // ff9f8165
    ZFSUB P0.M, Z0.D, Z0.D, Z0.D                     // 0080c165
    ZFSUB P3.M, Z10.D, Z12.D, Z10.D                  // 8a8dc165
    ZFSUB P7.M, Z31.D, Z31.D, Z31.D                  // ff9fc165

// FSUBR   <Zdn>.<T>, <Pg>/M, <Zdn>.<T>, <Zm>.<T>
    ZFSUBR P0.M, Z0.H, Z0.H, Z0.H                    // 00804365
    ZFSUBR P3.M, Z10.H, Z12.H, Z10.H                 // 8a8d4365
    ZFSUBR P7.M, Z31.H, Z31.H, Z31.H                 // ff9f4365
    ZFSUBR P0.M, Z0.S, Z0.S, Z0.S                    // 00808365
    ZFSUBR P3.M, Z10.S, Z12.S, Z10.S                 // 8a8d8365
    ZFSUBR P7.M, Z31.S, Z31.S, Z31.S                 // ff9f8365
    ZFSUBR P0.M, Z0.D, Z0.D, Z0.D                    // 0080c365
    ZFSUBR P3.M, Z10.D, Z12.D, Z10.D                 // 8a8dc365
    ZFSUBR P7.M, Z31.D, Z31.D, Z31.D                 // ff9fc365

// LSL     <Zdn>.<T>, <Pg>/M, <Zdn>.<T>, <Zm>.<T>
    ZLSL P0.M, Z0.B, Z0.B, Z0.B                      // 00801304
    ZLSL P3.M, Z10.B, Z12.B, Z10.B                   // 8a8d1304
    ZLSL P7.M, Z31.B, Z31.B, Z31.B                   // ff9f1304
    ZLSL P0.M, Z0.H, Z0.H, Z0.H                      // 00805304
    ZLSL P3.M, Z10.H, Z12.H, Z10.H                   // 8a8d5304
    ZLSL P7.M, Z31.H, Z31.H, Z31.H                   // ff9f5304
    ZLSL P0.M, Z0.S, Z0.S, Z0.S                      // 00809304
    ZLSL P3.M, Z10.S, Z12.S, Z10.S                   // 8a8d9304
    ZLSL P7.M, Z31.S, Z31.S, Z31.S                   // ff9f9304
    ZLSL P0.M, Z0.D, Z0.D, Z0.D                      // 0080d304
    ZLSL P3.M, Z10.D, Z12.D, Z10.D                   // 8a8dd304
    ZLSL P7.M, Z31.D, Z31.D, Z31.D                   // ff9fd304

// LSLR    <Zdn>.<T>, <Pg>/M, <Zdn>.<T>, <Zm>.<T>
    ZLSLR P0.M, Z0.B, Z0.B, Z0.B                     // 00801704
    ZLSLR P3.M, Z10.B, Z12.B, Z10.B                  // 8a8d1704
    ZLSLR P7.M, Z31.B, Z31.B, Z31.B                  // ff9f1704
    ZLSLR P0.M, Z0.H, Z0.H, Z0.H                     // 00805704
    ZLSLR P3.M, Z10.H, Z12.H, Z10.H                  // 8a8d5704
    ZLSLR P7.M, Z31.H, Z31.H, Z31.H                  // ff9f5704
    ZLSLR P0.M, Z0.S, Z0.S, Z0.S                     // 00809704
    ZLSLR P3.M, Z10.S, Z12.S, Z10.S                  // 8a8d9704
    ZLSLR P7.M, Z31.S, Z31.S, Z31.S                  // ff9f9704
    ZLSLR P0.M, Z0.D, Z0.D, Z0.D                     // 0080d704
    ZLSLR P3.M, Z10.D, Z12.D, Z10.D                  // 8a8dd704
    ZLSLR P7.M, Z31.D, Z31.D, Z31.D                  // ff9fd704

// LSR     <Zdn>.<T>, <Pg>/M, <Zdn>.<T>, <Zm>.<T>
    ZLSR P0.M, Z0.B, Z0.B, Z0.B                      // 00801104
    ZLSR P3.M, Z10.B, Z12.B, Z10.B                   // 8a8d1104
    ZLSR P7.M, Z31.B, Z31.B, Z31.B                   // ff9f1104
    ZLSR P0.M, Z0.H, Z0.H, Z0.H                      // 00805104
    ZLSR P3.M, Z10.H, Z12.H, Z10.H                   // 8a8d5104
    ZLSR P7.M, Z31.H, Z31.H, Z31.H                   // ff9f5104
    ZLSR P0.M, Z0.S, Z0.S, Z0.S                      // 00809104
    ZLSR P3.M, Z10.S, Z12.S, Z10.S                   // 8a8d9104
    ZLSR P7.M, Z31.S, Z31.S, Z31.S                   // ff9f9104
    ZLSR P0.M, Z0.D, Z0.D, Z0.D                      // 0080d104
    ZLSR P3.M, Z10.D, Z12.D, Z10.D                   // 8a8dd104
    ZLSR P7.M, Z31.D, Z31.D, Z31.D                   // ff9fd104

// LSRR    <Zdn>.<T>, <Pg>/M, <Zdn>.<T>, <Zm>.<T>
    ZLSRR P0.M, Z0.B, Z0.B, Z0.B                     // 00801504
    ZLSRR P3.M, Z10.B, Z12.B, Z10.B                  // 8a8d1504
    ZLSRR P7.M, Z31.B, Z31.B, Z31.B                  // ff9f1504
    ZLSRR P0.M, Z0.H, Z0.H, Z0.H                     // 00805504
    ZLSRR P3.M, Z10.H, Z12.H, Z10.H                  // 8a8d5504
    ZLSRR P7.M, Z31.H, Z31.H, Z31.H                  // ff9f5504
    ZLSRR P0.M, Z0.S, Z0.S, Z0.S                     // 00809504
    ZLSRR P3.M, Z10.S, Z12.S, Z10.S                  // 8a8d9504
    ZLSRR P7.M, Z31.S, Z31.S, Z31.S                  // ff9f9504
    ZLSRR P0.M, Z0.D, Z0.D, Z0.D                     // 0080d504
    ZLSRR P3.M, Z10.D, Z12.D, Z10.D                  // 8a8dd504
    ZLSRR P7.M, Z31.D, Z31.D, Z31.D                  // ff9fd504

// MUL     <Zdn>.<T>, <Pg>/M, <Zdn>.<T>, <Zm>.<T>
    ZMUL P0.M, Z0.B, Z0.B, Z0.B                      // 00001004
    ZMUL P3.M, Z10.B, Z12.B, Z10.B                   // 8a0d1004
    ZMUL P7.M, Z31.B, Z31.B, Z31.B                   // ff1f1004
    ZMUL P0.M, Z0.H, Z0.H, Z0.H                      // 00005004
    ZMUL P3.M, Z10.H, Z12.H, Z10.H                   // 8a0d5004
    ZMUL P7.M, Z31.H, Z31.H, Z31.H                   // ff1f5004
    ZMUL P0.M, Z0.S, Z0.S, Z0.S                      // 00009004
    ZMUL P3.M, Z10.S, Z12.S, Z10.S                   // 8a0d9004
    ZMUL P7.M, Z31.S, Z31.S, Z31.S                   // ff1f9004
    ZMUL P0.M, Z0.D, Z0.D, Z0.D                      // 0000d004
    ZMUL P3.M, Z10.D, Z12.D, Z10.D                   // 8a0dd004
    ZMUL P7.M, Z31.D, Z31.D, Z31.D                   // ff1fd004

// ORR     <Zdn>.<T>, <Pg>/M, <Zdn>.<T>, <Zm>.<T>
    ZORR P0.M, Z0.B, Z0.B, Z0.B                      // 00001804
    ZORR P3.M, Z10.B, Z12.B, Z10.B                   // 8a0d1804
    ZORR P7.M, Z31.B, Z31.B, Z31.B                   // ff1f1804
    ZORR P0.M, Z0.H, Z0.H, Z0.H                      // 00005804
    ZORR P3.M, Z10.H, Z12.H, Z10.H                   // 8a0d5804
    ZORR P7.M, Z31.H, Z31.H, Z31.H                   // ff1f5804
    ZORR P0.M, Z0.S, Z0.S, Z0.S                      // 00009804
    ZORR P3.M, Z10.S, Z12.S, Z10.S                   // 8a0d9804
    ZORR P7.M, Z31.S, Z31.S, Z31.S                   // ff1f9804
    ZORR P0.M, Z0.D, Z0.D, Z0.D                      // 0000d804
    ZORR P3.M, Z10.D, Z12.D, Z10.D                   // 8a0dd804
    ZORR P7.M, Z31.D, Z31.D, Z31.D                   // ff1fd804

// SABD    <Zdn>.<T>, <Pg>/M, <Zdn>.<T>, <Zm>.<T>
    ZSABD P0.M, Z0.B, Z0.B, Z0.B                     // 00000c04
    ZSABD P3.M, Z10.B, Z12.B, Z10.B                  // 8a0d0c04
    ZSABD P7.M, Z31.B, Z31.B, Z31.B                  // ff1f0c04
    ZSABD P0.M, Z0.H, Z0.H, Z0.H                     // 00004c04
    ZSABD P3.M, Z10.H, Z12.H, Z10.H                  // 8a0d4c04
    ZSABD P7.M, Z31.H, Z31.H, Z31.H                  // ff1f4c04
    ZSABD P0.M, Z0.S, Z0.S, Z0.S                     // 00008c04
    ZSABD P3.M, Z10.S, Z12.S, Z10.S                  // 8a0d8c04
    ZSABD P7.M, Z31.S, Z31.S, Z31.S                  // ff1f8c04
    ZSABD P0.M, Z0.D, Z0.D, Z0.D                     // 0000cc04
    ZSABD P3.M, Z10.D, Z12.D, Z10.D                  // 8a0dcc04
    ZSABD P7.M, Z31.D, Z31.D, Z31.D                  // ff1fcc04

// SDIV    <Zdn>.<T>, <Pg>/M, <Zdn>.<T>, <Zm>.<T>
    ZSDIV P0.M, Z0.S, Z0.S, Z0.S                     // 00009404
    ZSDIV P3.M, Z10.S, Z12.S, Z10.S                  // 8a0d9404
    ZSDIV P7.M, Z31.S, Z31.S, Z31.S                  // ff1f9404
    ZSDIV P0.M, Z0.D, Z0.D, Z0.D                     // 0000d404
    ZSDIV P3.M, Z10.D, Z12.D, Z10.D                  // 8a0dd404
    ZSDIV P7.M, Z31.D, Z31.D, Z31.D                  // ff1fd404

// SDIVR   <Zdn>.<T>, <Pg>/M, <Zdn>.<T>, <Zm>.<T>
    ZSDIVR P0.M, Z0.S, Z0.S, Z0.S                    // 00009604
    ZSDIVR P3.M, Z10.S, Z12.S, Z10.S                 // 8a0d9604
    ZSDIVR P7.M, Z31.S, Z31.S, Z31.S                 // ff1f9604
    ZSDIVR P0.M, Z0.D, Z0.D, Z0.D                    // 0000d604
    ZSDIVR P3.M, Z10.D, Z12.D, Z10.D                 // 8a0dd604
    ZSDIVR P7.M, Z31.D, Z31.D, Z31.D                 // ff1fd604

// SMAX    <Zdn>.<T>, <Pg>/M, <Zdn>.<T>, <Zm>.<T>
    ZSMAX P0.M, Z0.B, Z0.B, Z0.B                     // 00000804
    ZSMAX P3.M, Z10.B, Z12.B, Z10.B                  // 8a0d0804
    ZSMAX P7.M, Z31.B, Z31.B, Z31.B                  // ff1f0804
    ZSMAX P0.M, Z0.H, Z0.H, Z0.H                     // 00004804
    ZSMAX P3.M, Z10.H, Z12.H, Z10.H                  // 8a0d4804
    ZSMAX P7.M, Z31.H, Z31.H, Z31.H                  // ff1f4804
    ZSMAX P0.M, Z0.S, Z0.S, Z0.S                     // 00008804
    ZSMAX P3.M, Z10.S, Z12.S, Z10.S                  // 8a0d8804
    ZSMAX P7.M, Z31.S, Z31.S, Z31.S                  // ff1f8804
    ZSMAX P0.M, Z0.D, Z0.D, Z0.D                     // 0000c804
    ZSMAX P3.M, Z10.D, Z12.D, Z10.D                  // 8a0dc804
    ZSMAX P7.M, Z31.D, Z31.D, Z31.D                  // ff1fc804

// SMIN    <Zdn>.<T>, <Pg>/M, <Zdn>.<T>, <Zm>.<T>
    ZSMIN P0.M, Z0.B, Z0.B, Z0.B                     // 00000a04
    ZSMIN P3.M, Z10.B, Z12.B, Z10.B                  // 8a0d0a04
    ZSMIN P7.M, Z31.B, Z31.B, Z31.B                  // ff1f0a04
    ZSMIN P0.M, Z0.H, Z0.H, Z0.H                     // 00004a04
    ZSMIN P3.M, Z10.H, Z12.H, Z10.H                  // 8a0d4a04
    ZSMIN P7.M, Z31.H, Z31.H, Z31.H                  // ff1f4a04
    ZSMIN P0.M, Z0.S, Z0.S, Z0.S                     // 00008a04
    ZSMIN P3.M, Z10.S, Z12.S, Z10.S                  // 8a0d8a04
    ZSMIN P7.M, Z31.S, Z31.S, Z31.S                  // ff1f8a04
    ZSMIN P0.M, Z0.D, Z0.D, Z0.D                     // 0000ca04
    ZSMIN P3.M, Z10.D, Z12.D, Z10.D                  // 8a0dca04
    ZSMIN P7.M, Z31.D, Z31.D, Z31.D                  // ff1fca04

// SMULH   <Zdn>.<T>, <Pg>/M, <Zdn>.<T>, <Zm>.<T>
    ZSMULH P0.M, Z0.B, Z0.B, Z0.B                    // 00001204
    ZSMULH P3.M, Z10.B, Z12.B, Z10.B                 // 8a0d1204
    ZSMULH P7.M, Z31.B, Z31.B, Z31.B                 // ff1f1204
    ZSMULH P0.M, Z0.H, Z0.H, Z0.H                    // 00005204
    ZSMULH P3.M, Z10.H, Z12.H, Z10.H                 // 8a0d5204
    ZSMULH P7.M, Z31.H, Z31.H, Z31.H                 // ff1f5204
    ZSMULH P0.M, Z0.S, Z0.S, Z0.S                    // 00009204
    ZSMULH P3.M, Z10.S, Z12.S, Z10.S                 // 8a0d9204
    ZSMULH P7.M, Z31.S, Z31.S, Z31.S                 // ff1f9204
    ZSMULH P0.M, Z0.D, Z0.D, Z0.D                    // 0000d204
    ZSMULH P3.M, Z10.D, Z12.D, Z10.D                 // 8a0dd204
    ZSMULH P7.M, Z31.D, Z31.D, Z31.D                 // ff1fd204

// SUB     <Zdn>.<T>, <Pg>/M, <Zdn>.<T>, <Zm>.<T>
    ZSUB P0.M, Z0.B, Z0.B, Z0.B                      // 00000104
    ZSUB P3.M, Z10.B, Z12.B, Z10.B                   // 8a0d0104
    ZSUB P7.M, Z31.B, Z31.B, Z31.B                   // ff1f0104
    ZSUB P0.M, Z0.H, Z0.H, Z0.H                      // 00004104
    ZSUB P3.M, Z10.H, Z12.H, Z10.H                   // 8a0d4104
    ZSUB P7.M, Z31.H, Z31.H, Z31.H                   // ff1f4104
    ZSUB P0.M, Z0.S, Z0.S, Z0.S                      // 00008104
    ZSUB P3.M, Z10.S, Z12.S, Z10.S                   // 8a0d8104
    ZSUB P7.M, Z31.S, Z31.S, Z31.S                   // ff1f8104
    ZSUB P0.M, Z0.D, Z0.D, Z0.D                      // 0000c104
    ZSUB P3.M, Z10.D, Z12.D, Z10.D                   // 8a0dc104
    ZSUB P7.M, Z31.D, Z31.D, Z31.D                   // ff1fc104

// SUBR    <Zdn>.<T>, <Pg>/M, <Zdn>.<T>, <Zm>.<T>
    ZSUBR P0.M, Z0.B, Z0.B, Z0.B                     // 00000304
    ZSUBR P3.M, Z10.B, Z12.B, Z10.B                  // 8a0d0304
    ZSUBR P7.M, Z31.B, Z31.B, Z31.B                  // ff1f0304
    ZSUBR P0.M, Z0.H, Z0.H, Z0.H                     // 00004304
    ZSUBR P3.M, Z10.H, Z12.H, Z10.H                  // 8a0d4304
    ZSUBR P7.M, Z31.H, Z31.H, Z31.H                  // ff1f4304
    ZSUBR P0.M, Z0.S, Z0.S, Z0.S                     // 00008304
    ZSUBR P3.M, Z10.S, Z12.S, Z10.S                  // 8a0d8304
    ZSUBR P7.M, Z31.S, Z31.S, Z31.S                  // ff1f8304
    ZSUBR P0.M, Z0.D, Z0.D, Z0.D                     // 0000c304
    ZSUBR P3.M, Z10.D, Z12.D, Z10.D                  // 8a0dc304
    ZSUBR P7.M, Z31.D, Z31.D, Z31.D                  // ff1fc304

// UABD    <Zdn>.<T>, <Pg>/M, <Zdn>.<T>, <Zm>.<T>
    ZUABD P0.M, Z0.B, Z0.B, Z0.B                     // 00000d04
    ZUABD P3.M, Z10.B, Z12.B, Z10.B                  // 8a0d0d04
    ZUABD P7.M, Z31.B, Z31.B, Z31.B                  // ff1f0d04
    ZUABD P0.M, Z0.H, Z0.H, Z0.H                     // 00004d04
    ZUABD P3.M, Z10.H, Z12.H, Z10.H                  // 8a0d4d04
    ZUABD P7.M, Z31.H, Z31.H, Z31.H                  // ff1f4d04
    ZUABD P0.M, Z0.S, Z0.S, Z0.S                     // 00008d04
    ZUABD P3.M, Z10.S, Z12.S, Z10.S                  // 8a0d8d04
    ZUABD P7.M, Z31.S, Z31.S, Z31.S                  // ff1f8d04
    ZUABD P0.M, Z0.D, Z0.D, Z0.D                     // 0000cd04
    ZUABD P3.M, Z10.D, Z12.D, Z10.D                  // 8a0dcd04
    ZUABD P7.M, Z31.D, Z31.D, Z31.D                  // ff1fcd04

// UDIV    <Zdn>.<T>, <Pg>/M, <Zdn>.<T>, <Zm>.<T>
    ZUDIV P0.M, Z0.S, Z0.S, Z0.S                     // 00009504
    ZUDIV P3.M, Z10.S, Z12.S, Z10.S                  // 8a0d9504
    ZUDIV P7.M, Z31.S, Z31.S, Z31.S                  // ff1f9504
    ZUDIV P0.M, Z0.D, Z0.D, Z0.D                     // 0000d504
    ZUDIV P3.M, Z10.D, Z12.D, Z10.D                  // 8a0dd504
    ZUDIV P7.M, Z31.D, Z31.D, Z31.D                  // ff1fd504

// UDIVR   <Zdn>.<T>, <Pg>/M, <Zdn>.<T>, <Zm>.<T>
    ZUDIVR P0.M, Z0.S, Z0.S, Z0.S                    // 00009704
    ZUDIVR P3.M, Z10.S, Z12.S, Z10.S                 // 8a0d9704
    ZUDIVR P7.M, Z31.S, Z31.S, Z31.S                 // ff1f9704
    ZUDIVR P0.M, Z0.D, Z0.D, Z0.D                    // 0000d704
    ZUDIVR P3.M, Z10.D, Z12.D, Z10.D                 // 8a0dd704
    ZUDIVR P7.M, Z31.D, Z31.D, Z31.D                 // ff1fd704

// UMAX    <Zdn>.<T>, <Pg>/M, <Zdn>.<T>, <Zm>.<T>
    ZUMAX P0.M, Z0.B, Z0.B, Z0.B                     // 00000904
    ZUMAX P3.M, Z10.B, Z12.B, Z10.B                  // 8a0d0904
    ZUMAX P7.M, Z31.B, Z31.B, Z31.B                  // ff1f0904
    ZUMAX P0.M, Z0.H, Z0.H, Z0.H                     // 00004904
    ZUMAX P3.M, Z10.H, Z12.H, Z10.H                  // 8a0d4904
    ZUMAX P7.M, Z31.H, Z31.H, Z31.H                  // ff1f4904
    ZUMAX P0.M, Z0.S, Z0.S, Z0.S                     // 00008904
    ZUMAX P3.M, Z10.S, Z12.S, Z10.S                  // 8a0d8904
    ZUMAX P7.M, Z31.S, Z31.S, Z31.S                  // ff1f8904
    ZUMAX P0.M, Z0.D, Z0.D, Z0.D                     // 0000c904
    ZUMAX P3.M, Z10.D, Z12.D, Z10.D                  // 8a0dc904
    ZUMAX P7.M, Z31.D, Z31.D, Z31.D                  // ff1fc904

// UMIN    <Zdn>.<T>, <Pg>/M, <Zdn>.<T>, <Zm>.<T>
    ZUMIN P0.M, Z0.B, Z0.B, Z0.B                     // 00000b04
    ZUMIN P3.M, Z10.B, Z12.B, Z10.B                  // 8a0d0b04
    ZUMIN P7.M, Z31.B, Z31.B, Z31.B                  // ff1f0b04
    ZUMIN P0.M, Z0.H, Z0.H, Z0.H                     // 00004b04
    ZUMIN P3.M, Z10.H, Z12.H, Z10.H                  // 8a0d4b04
    ZUMIN P7.M, Z31.H, Z31.H, Z31.H                  // ff1f4b04
    ZUMIN P0.M, Z0.S, Z0.S, Z0.S                     // 00008b04
    ZUMIN P3.M, Z10.S, Z12.S, Z10.S                  // 8a0d8b04
    ZUMIN P7.M, Z31.S, Z31.S, Z31.S                  // ff1f8b04
    ZUMIN P0.M, Z0.D, Z0.D, Z0.D                     // 0000cb04
    ZUMIN P3.M, Z10.D, Z12.D, Z10.D                  // 8a0dcb04
    ZUMIN P7.M, Z31.D, Z31.D, Z31.D                  // ff1fcb04

// UMULH   <Zdn>.<T>, <Pg>/M, <Zdn>.<T>, <Zm>.<T>
    ZUMULH P0.M, Z0.B, Z0.B, Z0.B                    // 00001304
    ZUMULH P3.M, Z10.B, Z12.B, Z10.B                 // 8a0d1304
    ZUMULH P7.M, Z31.B, Z31.B, Z31.B                 // ff1f1304
    ZUMULH P0.M, Z0.H, Z0.H, Z0.H                    // 00005304
    ZUMULH P3.M, Z10.H, Z12.H, Z10.H                 // 8a0d5304
    ZUMULH P7.M, Z31.H, Z31.H, Z31.H                 // ff1f5304
    ZUMULH P0.M, Z0.S, Z0.S, Z0.S                    // 00009304
    ZUMULH P3.M, Z10.S, Z12.S, Z10.S                 // 8a0d9304
    ZUMULH P7.M, Z31.S, Z31.S, Z31.S                 // ff1f9304
    ZUMULH P0.M, Z0.D, Z0.D, Z0.D                    // 0000d304
    ZUMULH P3.M, Z10.D, Z12.D, Z10.D                 // 8a0dd304
    ZUMULH P7.M, Z31.D, Z31.D, Z31.D                 // ff1fd304

// CLS     <Zd>.<T>, <Pg>/M, <Zn>.<T>
    ZCLS P0.M, Z0.B, Z0.B                            // 00a01804
    ZCLS P3.M, Z12.B, Z10.B                          // 8aad1804
    ZCLS P7.M, Z31.B, Z31.B                          // ffbf1804
    ZCLS P0.M, Z0.H, Z0.H                            // 00a05804
    ZCLS P3.M, Z12.H, Z10.H                          // 8aad5804
    ZCLS P7.M, Z31.H, Z31.H                          // ffbf5804
    ZCLS P0.M, Z0.S, Z0.S                            // 00a09804
    ZCLS P3.M, Z12.S, Z10.S                          // 8aad9804
    ZCLS P7.M, Z31.S, Z31.S                          // ffbf9804
    ZCLS P0.M, Z0.D, Z0.D                            // 00a0d804
    ZCLS P3.M, Z12.D, Z10.D                          // 8aadd804
    ZCLS P7.M, Z31.D, Z31.D                          // ffbfd804

// CLZ     <Zd>.<T>, <Pg>/M, <Zn>.<T>
    ZCLZ P0.M, Z0.B, Z0.B                            // 00a01904
    ZCLZ P3.M, Z12.B, Z10.B                          // 8aad1904
    ZCLZ P7.M, Z31.B, Z31.B                          // ffbf1904
    ZCLZ P0.M, Z0.H, Z0.H                            // 00a05904
    ZCLZ P3.M, Z12.H, Z10.H                          // 8aad5904
    ZCLZ P7.M, Z31.H, Z31.H                          // ffbf5904
    ZCLZ P0.M, Z0.S, Z0.S                            // 00a09904
    ZCLZ P3.M, Z12.S, Z10.S                          // 8aad9904
    ZCLZ P7.M, Z31.S, Z31.S                          // ffbf9904
    ZCLZ P0.M, Z0.D, Z0.D                            // 00a0d904
    ZCLZ P3.M, Z12.D, Z10.D                          // 8aadd904
    ZCLZ P7.M, Z31.D, Z31.D                          // ffbfd904

// NEG     <Zd>.<T>, <Pg>/M, <Zn>.<T>
    ZNEG P0.M, Z0.B, Z0.B                            // 00a01704
    ZNEG P3.M, Z12.B, Z10.B                          // 8aad1704
    ZNEG P7.M, Z31.B, Z31.B                          // ffbf1704
    ZNEG P0.M, Z0.H, Z0.H                            // 00a05704
    ZNEG P3.M, Z12.H, Z10.H                          // 8aad5704
    ZNEG P7.M, Z31.H, Z31.H                          // ffbf5704
    ZNEG P0.M, Z0.S, Z0.S                            // 00a09704
    ZNEG P3.M, Z12.S, Z10.S                          // 8aad9704
    ZNEG P7.M, Z31.S, Z31.S                          // ffbf9704
    ZNEG P0.M, Z0.D, Z0.D                            // 00a0d704
    ZNEG P3.M, Z12.D, Z10.D                          // 8aadd704
    ZNEG P7.M, Z31.D, Z31.D                          // ffbfd704

// RBIT    <Zd>.<T>, <Pg>/M, <Zn>.<T>
    ZRBIT P0.M, Z0.B, Z0.B                           // 00802705
    ZRBIT P3.M, Z12.B, Z10.B                         // 8a8d2705
    ZRBIT P7.M, Z31.B, Z31.B                         // ff9f2705
    ZRBIT P0.M, Z0.H, Z0.H                           // 00806705
    ZRBIT P3.M, Z12.H, Z10.H                         // 8a8d6705
    ZRBIT P7.M, Z31.H, Z31.H                         // ff9f6705
    ZRBIT P0.M, Z0.S, Z0.S                           // 0080a705
    ZRBIT P3.M, Z12.S, Z10.S                         // 8a8da705
    ZRBIT P7.M, Z31.S, Z31.S                         // ff9fa705
    ZRBIT P0.M, Z0.D, Z0.D                           // 0080e705
    ZRBIT P3.M, Z12.D, Z10.D                         // 8a8de705
    ZRBIT P7.M, Z31.D, Z31.D                         // ff9fe705

// SXTB    <Zd>.<T>, <Pg>/M, <Zn>.<T>
    ZSXTB P0.M, Z0.H, Z0.H                           // 00a05004
    ZSXTB P3.M, Z12.H, Z10.H                         // 8aad5004
    ZSXTB P7.M, Z31.H, Z31.H                         // ffbf5004
    ZSXTB P0.M, Z0.S, Z0.S                           // 00a09004
    ZSXTB P3.M, Z12.S, Z10.S                         // 8aad9004
    ZSXTB P7.M, Z31.S, Z31.S                         // ffbf9004
    ZSXTB P0.M, Z0.D, Z0.D                           // 00a0d004
    ZSXTB P3.M, Z12.D, Z10.D                         // 8aadd004
    ZSXTB P7.M, Z31.D, Z31.D                         // ffbfd004

// SXTH    <Zd>.<T>, <Pg>/M, <Zn>.<T>
    ZSXTH P0.M, Z0.S, Z0.S                           // 00a09204
    ZSXTH P3.M, Z12.S, Z10.S                         // 8aad9204
    ZSXTH P7.M, Z31.S, Z31.S                         // ffbf9204
    ZSXTH P0.M, Z0.D, Z0.D                           // 00a0d204
    ZSXTH P3.M, Z12.D, Z10.D                         // 8aadd204
    ZSXTH P7.M, Z31.D, Z31.D                         // ffbfd204

// UXTB    <Zd>.<T>, <Pg>/M, <Zn>.<T>
    ZUXTB P0.M, Z0.H, Z0.H                           // 00a05104
    ZUXTB P3.M, Z12.H, Z10.H                         // 8aad5104
    ZUXTB P7.M, Z31.H, Z31.H                         // ffbf5104
    ZUXTB P0.M, Z0.S, Z0.S                           // 00a09104
    ZUXTB P3.M, Z12.S, Z10.S                         // 8aad9104
    ZUXTB P7.M, Z31.S, Z31.S                         // ffbf9104
    ZUXTB P0.M, Z0.D, Z0.D                           // 00a0d104
    ZUXTB P3.M, Z12.D, Z10.D                         // 8aadd104
    ZUXTB P7.M, Z31.D, Z31.D                         // ffbfd104

// UXTH    <Zd>.<T>, <Pg>/M, <Zn>.<T>
    ZUXTH P0.M, Z0.S, Z0.S                           // 00a09304
    ZUXTH P3.M, Z12.S, Z10.S                         // 8aad9304
    ZUXTH P7.M, Z31.S, Z31.S                         // ffbf9304
    ZUXTH P0.M, Z0.D, Z0.D                           // 00a0d304
    ZUXTH P3.M, Z12.D, Z10.D                         // 8aadd304
    ZUXTH P7.M, Z31.D, Z31.D                         // ffbfd304

// DECP    <Xdn>, <Pm>.<T>
    ZDECP P0.B, R0                                   // 00882d25
    ZDECP P6.B, R10                                  // ca882d25
    ZDECP P15.B, R30                                 // fe892d25
    ZDECP P0.H, R0                                   // 00886d25
    ZDECP P6.H, R10                                  // ca886d25
    ZDECP P15.H, R30                                 // fe896d25
    ZDECP P0.S, R0                                   // 0088ad25
    ZDECP P6.S, R10                                  // ca88ad25
    ZDECP P15.S, R30                                 // fe89ad25
    ZDECP P0.D, R0                                   // 0088ed25
    ZDECP P6.D, R10                                  // ca88ed25
    ZDECP P15.D, R30                                 // fe89ed25

// INCP    <Xdn>, <Pm>.<T>
    ZINCP P0.B, R0                                   // 00882c25
    ZINCP P6.B, R10                                  // ca882c25
    ZINCP P15.B, R30                                 // fe892c25
    ZINCP P0.H, R0                                   // 00886c25
    ZINCP P6.H, R10                                  // ca886c25
    ZINCP P15.H, R30                                 // fe896c25
    ZINCP P0.S, R0                                   // 0088ac25
    ZINCP P6.S, R10                                  // ca88ac25
    ZINCP P15.S, R30                                 // fe89ac25
    ZINCP P0.D, R0                                   // 0088ec25
    ZINCP P6.D, R10                                  // ca88ec25
    ZINCP P15.D, R30                                 // fe89ec25

// LDR     <Zt>, [<Xn|SP>{, #<simm>, MUL VL}]
    ZLDR $-256(R0), Z0                               // 0040a085
    ZLDR $(R11), Z10                                  // 6a418085
    ZLDR $255(RSP), Z31                              // ff5f9f85

// STR     <Zt>, [<Xn|SP>{, #<simm>, MUL VL}]
    ZSTR $-256(R0), Z0                               // 0040a0e5
    ZSTR $(R11), Z10                                  // 6a4180e5
    ZSTR $255(RSP), Z31                              // ff5f9fe5

// LDR     <Pt>, [<Xn|SP>{, #<simm>, MUL VL}]
    ZLDR $-256(R0), P0                               // 0000a085
    ZLDR $(R11), P5                                   // 65018085
    ZLDR $255(RSP), P15                              // ef1f9f85

// STR     <Pt>, [<Xn|SP>{, #<simm>, MUL VL}]
    ZSTR $-256(R0), P0                               // 0000a0e5
    ZSTR $(R11), P5                                   // 650180e5
    ZSTR $255(RSP), P15                              // ef1f9fe5

    SETFFR                                          // 00902c25

// SADDV   <Dd>, <Pg>, <Zn>.<T>
    ZSADDV P0, Z0.B, F0                              // 00200004
    ZSADDV P3, Z12.B, F10                            // 8a2d0004
    ZSADDV P7, Z31.B, F31                            // ff3f0004
    ZSADDV P0, Z0.H, F0                              // 00204004
    ZSADDV P3, Z12.H, F10                            // 8a2d4004
    ZSADDV P7, Z31.H, F31                            // ff3f4004
    ZSADDV P0, Z0.S, F0                              // 00208004
    ZSADDV P3, Z12.S, F10                            // 8a2d8004
    ZSADDV P7, Z31.S, F31                            // ff3f8004

// SADDV   <Dd>, <Pg>, <Zn>.<T>
    ZSADDV P0, Z0.B, F0                              // 00200004
    ZSADDV P3, Z12.B, F10                            // 8a2d0004
    ZSADDV P7, Z31.B, F31                            // ff3f0004
    ZSADDV P0, Z0.H, F0                              // 00204004
    ZSADDV P3, Z12.H, F10                            // 8a2d4004
    ZSADDV P7, Z31.H, F31                            // ff3f4004
    ZSADDV P0, Z0.S, F0                              // 00208004
    ZSADDV P3, Z12.S, F10                            // 8a2d8004
    ZSADDV P7, Z31.S, F31                            // ff3f8004

// FCMEQ   <Pd>.<T>, <Pg>/Z, <Zn>.<T>, #0.0
    ZFCMEQ P0.Z, Z0.H, $(0.0), P0.H                  // 00205265
    ZFCMEQ P3.Z, Z12.H, $(0.0), P5.H                 // 852d5265
    ZFCMEQ P7.Z, Z31.H, $(0.0), P15.H                // ef3f5265
    ZFCMEQ P0.Z, Z0.S, $(0.0), P0.S                  // 00209265
    ZFCMEQ P3.Z, Z12.S, $(0.0), P5.S                 // 852d9265
    ZFCMEQ P7.Z, Z31.S, $(0.0), P15.S                // ef3f9265
    ZFCMEQ P0.Z, Z0.D, $(0.0), P0.D                  // 0020d265
    ZFCMEQ P3.Z, Z12.D, $(0.0), P5.D                 // 852dd265
    ZFCMEQ P7.Z, Z31.D, $(0.0), P15.D                // ef3fd265

// FCMGE   <Pd>.<T>, <Pg>/Z, <Zn>.<T>, #0.0
    ZFCMGE P0.Z, Z0.H, $(0.0), P0.H                    // 00205065
    ZFCMGE P3.Z, Z12.H, $(0.0), P5.H                   // 852d5065
    ZFCMGE P7.Z, Z31.H, $(0.0), P15.H                  // ef3f5065
    ZFCMGE P0.Z, Z0.S, $(0.0), P0.S                    // 00209065
    ZFCMGE P3.Z, Z12.S, $(0.0), P5.S                   // 852d9065
    ZFCMGE P7.Z, Z31.S, $(0.0), P15.S                  // ef3f9065
    ZFCMGE P0.Z, Z0.D, $(0.0), P0.D                    // 0020d065
    ZFCMGE P3.Z, Z12.D, $(0.0), P5.D                   // 852dd065
    ZFCMGE P7.Z, Z31.D, $(0.0), P15.D                  // ef3fd065

// FCMGT   <Pd>.<T>, <Pg>/Z, <Zn>.<T>, #0.0
    ZFCMGT P0.Z, Z0.H, $(0.0), P0.H                    // 10205065
    ZFCMGT P3.Z, Z12.H, $(0.0), P5.H                   // 952d5065
    ZFCMGT P7.Z, Z31.H, $(0.0), P15.H                  // ff3f5065
    ZFCMGT P0.Z, Z0.S, $(0.0), P0.S                    // 10209065
    ZFCMGT P3.Z, Z12.S, $(0.0), P5.S                   // 952d9065
    ZFCMGT P7.Z, Z31.S, $(0.0), P15.S                  // ff3f9065
    ZFCMGT P0.Z, Z0.D, $(0.0), P0.D                    // 1020d065
    ZFCMGT P3.Z, Z12.D, $(0.0), P5.D                   // 952dd065
    ZFCMGT P7.Z, Z31.D, $(0.0), P15.D                  // ff3fd065

// FCMLE   <Pd>.<T>, <Pg>/Z, <Zn>.<T>, #0.0
    ZFCMLE P0.Z, Z0.H, $(0.0), P0.H                    // 10205165
    ZFCMLE P3.Z, Z12.H, $(0.0), P5.H                   // 952d5165
    ZFCMLE P7.Z, Z31.H, $(0.0), P15.H                  // ff3f5165
    ZFCMLE P0.Z, Z0.S, $(0.0), P0.S                    // 10209165
    ZFCMLE P3.Z, Z12.S, $(0.0), P5.S                   // 952d9165
    ZFCMLE P7.Z, Z31.S, $(0.0), P15.S                  // ff3f9165
    ZFCMLE P0.Z, Z0.D, $(0.0), P0.D                    // 1020d165
    ZFCMLE P3.Z, Z12.D, $(0.0), P5.D                   // 952dd165
    ZFCMLE P7.Z, Z31.D, $(0.0), P15.D                  // ff3fd165

// FCMLT   <Pd>.<T>, <Pg>/Z, <Zn>.<T>, #0.0
    ZFCMLT P0.Z, Z0.H, $(0.0), P0.H                    // 00205165
    ZFCMLT P3.Z, Z12.H, $(0.0), P5.H                   // 852d5165
    ZFCMLT P7.Z, Z31.H, $(0.0), P15.H                  // ef3f5165
    ZFCMLT P0.Z, Z0.S, $(0.0), P0.S                    // 00209165
    ZFCMLT P3.Z, Z12.S, $(0.0), P5.S                   // 852d9165
    ZFCMLT P7.Z, Z31.S, $(0.0), P15.S                  // ef3f9165
    ZFCMLT P0.Z, Z0.D, $(0.0), P0.D                    // 0020d165
    ZFCMLT P3.Z, Z12.D, $(0.0), P5.D                   // 852dd165
    ZFCMLT P7.Z, Z31.D, $(0.0), P15.D                  // ef3fd165

// FCMNE   <Pd>.<T>, <Pg>/Z, <Zn>.<T>, #0.0
    ZFCMNE P0.Z, Z0.H, $(0.0), P0.H                    // 00205365
    ZFCMNE P3.Z, Z12.H, $(0.0), P5.H                   // 852d5365
    ZFCMNE P7.Z, Z31.H, $(0.0), P15.H                  // ef3f5365
    ZFCMNE P0.Z, Z0.S, $(0.0), P0.S                    // 00209365
    ZFCMNE P3.Z, Z12.S, $(0.0), P5.S                   // 852d9365
    ZFCMNE P7.Z, Z31.S, $(0.0), P15.S                  // ef3f9365
    ZFCMNE P0.Z, Z0.D, $(0.0), P0.D                    // 0020d365
    ZFCMNE P3.Z, Z12.D, $(0.0), P5.D                   // 852dd365
    ZFCMNE P7.Z, Z31.D, $(0.0), P15.D                  // ef3fd365

// CMPEQ   <Pd>.<T>, <Pg>/Z, <Zn>.<T>, #<imm>
    ZCMPEQ P0.Z, Z0.B, $-16, P0.B                  // 00801025
    ZCMPEQ P3.Z, Z12.B, $-6, P5.B                  // 858d1a25
    ZCMPEQ P7.Z, Z31.B, $15, P15.B                  // ef9f0f25
    ZCMPEQ P0.Z, Z0.H, $-16, P0.H                  // 00805025
    ZCMPEQ P3.Z, Z12.H, $-6, P5.H                  // 858d5a25
    ZCMPEQ P7.Z, Z31.H, $15, P15.H                  // ef9f4f25
    ZCMPEQ P0.Z, Z0.S, $-16, P0.S                  // 00809025
    ZCMPEQ P3.Z, Z12.S, $-6, P5.S                  // 858d9a25
    ZCMPEQ P7.Z, Z31.S, $15, P15.S                  // ef9f8f25
    ZCMPEQ P0.Z, Z0.D, $-16, P0.D                  // 0080d025
    ZCMPEQ P3.Z, Z12.D, $-6, P5.D                  // 858dda25
    ZCMPEQ P7.Z, Z31.D, $15, P15.D                  // ef9fcf25

// CMPGE   <Pd>.<T>, <Pg>/Z, <Zn>.<T>, #<imm>
    ZCMPGE P0.Z, Z0.B, $-16, P0.B                  // 00001025
    ZCMPGE P3.Z, Z12.B, $-6, P5.B                  // 850d1a25
    ZCMPGE P7.Z, Z31.B, $15, P15.B                  // ef1f0f25
    ZCMPGE P0.Z, Z0.H, $-16, P0.H                  // 00005025
    ZCMPGE P3.Z, Z12.H, $-6, P5.H                  // 850d5a25
    ZCMPGE P7.Z, Z31.H, $15, P15.H                  // ef1f4f25
    ZCMPGE P0.Z, Z0.S, $-16, P0.S                  // 00009025
    ZCMPGE P3.Z, Z12.S, $-6, P5.S                  // 850d9a25
    ZCMPGE P7.Z, Z31.S, $15, P15.S                  // ef1f8f25
    ZCMPGE P0.Z, Z0.D, $-16, P0.D                  // 0000d025
    ZCMPGE P3.Z, Z12.D, $-6, P5.D                  // 850dda25
    ZCMPGE P7.Z, Z31.D, $15, P15.D                  // ef1fcf25

// CMPGT   <Pd>.<T>, <Pg>/Z, <Zn>.<T>, #<imm>
    ZCMPGT P0.Z, Z0.B, $-16, P0.B                  // 10001025
    ZCMPGT P3.Z, Z12.B, $-6, P5.B                  // 950d1a25
    ZCMPGT P7.Z, Z31.B, $15, P15.B                  // ff1f0f25
    ZCMPGT P0.Z, Z0.H, $-16, P0.H                  // 10005025
    ZCMPGT P3.Z, Z12.H, $-6, P5.H                  // 950d5a25
    ZCMPGT P7.Z, Z31.H, $15, P15.H                  // ff1f4f25
    ZCMPGT P0.Z, Z0.S, $-16, P0.S                  // 10009025
    ZCMPGT P3.Z, Z12.S, $-6, P5.S                  // 950d9a25
    ZCMPGT P7.Z, Z31.S, $15, P15.S                  // ff1f8f25
    ZCMPGT P0.Z, Z0.D, $-16, P0.D                  // 1000d025
    ZCMPGT P3.Z, Z12.D, $-6, P5.D                  // 950dda25
    ZCMPGT P7.Z, Z31.D, $15, P15.D                  // ff1fcf25

// CMPHI   <Pd>.<T>, <Pg>/Z, <Zn>.<T>, #<imm>
    ZCMPHI P0.Z, Z0.B, $0, P0.B                    // 10002024
    ZCMPHI P3.Z, Z12.B, $42, P5.B                  // 958d2a24
    ZCMPHI P7.Z, Z31.B, $127, P15.B                 // ffdf3f24
    ZCMPHI P0.Z, Z0.H, $0, P0.H                    // 10006024
    ZCMPHI P3.Z, Z12.H, $42, P5.H                  // 958d6a24
    ZCMPHI P7.Z, Z31.H, $127, P15.H                 // ffdf7f24
    ZCMPHI P0.Z, Z0.S, $0, P0.S                    // 1000a024
    ZCMPHI P3.Z, Z12.S, $42, P5.S                  // 958daa24
    ZCMPHI P7.Z, Z31.S, $127, P15.S                 // ffdfbf24
    ZCMPHI P0.Z, Z0.D, $0, P0.D                    // 1000e024
    ZCMPHI P3.Z, Z12.D, $42, P5.D                  // 958dea24
    ZCMPHI P7.Z, Z31.D, $127, P15.D                 // ffdfff24

// CMPHS   <Pd>.<T>, <Pg>/Z, <Zn>.<T>, #<imm>
    ZCMPHS P0.Z, Z0.B, $0, P0.B                    // 00002024
    ZCMPHS P3.Z, Z12.B, $42, P5.B                  // 858d2a24
    ZCMPHS P7.Z, Z31.B, $127, P15.B                 // efdf3f24
    ZCMPHS P0.Z, Z0.H, $0, P0.H                    // 00006024
    ZCMPHS P3.Z, Z12.H, $42, P5.H                  // 858d6a24
    ZCMPHS P7.Z, Z31.H, $127, P15.H                 // efdf7f24
    ZCMPHS P0.Z, Z0.S, $0, P0.S                    // 0000a024
    ZCMPHS P3.Z, Z12.S, $42, P5.S                  // 858daa24
    ZCMPHS P7.Z, Z31.S, $127, P15.S                 // efdfbf24
    ZCMPHS P0.Z, Z0.D, $0, P0.D                    // 0000e024
    ZCMPHS P3.Z, Z12.D, $42, P5.D                  // 858dea24
    ZCMPHS P7.Z, Z31.D, $127, P15.D                 // efdfff24

// CMPLE   <Pd>.<T>, <Pg>/Z, <Zn>.<T>, #<imm>
    ZCMPLE P0.Z, Z0.B, $-16, P0.B                  // 10201025
    ZCMPLE P3.Z, Z12.B, $-6, P5.B                  // 952d1a25
    ZCMPLE P7.Z, Z31.B, $15, P15.B                  // ff3f0f25
    ZCMPLE P0.Z, Z0.H, $-16, P0.H                  // 10205025
    ZCMPLE P3.Z, Z12.H, $-6, P5.H                  // 952d5a25
    ZCMPLE P7.Z, Z31.H, $15, P15.H                  // ff3f4f25
    ZCMPLE P0.Z, Z0.S, $-16, P0.S                  // 10209025
    ZCMPLE P3.Z, Z12.S, $-6, P5.S                  // 952d9a25
    ZCMPLE P7.Z, Z31.S, $15, P15.S                  // ff3f8f25
    ZCMPLE P0.Z, Z0.D, $-16, P0.D                  // 1020d025
    ZCMPLE P3.Z, Z12.D, $-6, P5.D                  // 952dda25
    ZCMPLE P7.Z, Z31.D, $15, P15.D                  // ff3fcf25

// CMPLO   <Pd>.<T>, <Pg>/Z, <Zn>.<T>, #<imm>
    ZCMPLO P0.Z, Z0.B, $0, P0.B                    // 00202024
    ZCMPLO P3.Z, Z12.B, $42, P5.B                  // 85ad2a24
    ZCMPLO P7.Z, Z31.B, $127, P15.B                 // efff3f24
    ZCMPLO P0.Z, Z0.H, $0, P0.H                    // 00206024
    ZCMPLO P3.Z, Z12.H, $42, P5.H                  // 85ad6a24
    ZCMPLO P7.Z, Z31.H, $127, P15.H                 // efff7f24
    ZCMPLO P0.Z, Z0.S, $0, P0.S                    // 0020a024
    ZCMPLO P3.Z, Z12.S, $42, P5.S                  // 85adaa24
    ZCMPLO P7.Z, Z31.S, $127, P15.S                 // efffbf24
    ZCMPLO P0.Z, Z0.D, $0, P0.D                    // 0020e024
    ZCMPLO P3.Z, Z12.D, $42, P5.D                  // 85adea24
    ZCMPLO P7.Z, Z31.D, $127, P15.D                 // efffff24

// CMPLS   <Pd>.<T>, <Pg>/Z, <Zn>.<T>, #<imm>
    ZCMPLS P0.Z, Z0.B, $0, P0.B                    // 10202024
    ZCMPLS P3.Z, Z12.B, $42, P5.B                  // 95ad2a24
    ZCMPLS P7.Z, Z31.B, $127, P15.B                 // ffff3f24
    ZCMPLS P0.Z, Z0.H, $0, P0.H                    // 10206024
    ZCMPLS P3.Z, Z12.H, $42, P5.H                  // 95ad6a24
    ZCMPLS P7.Z, Z31.H, $127, P15.H                 // ffff7f24
    ZCMPLS P0.Z, Z0.S, $0, P0.S                    // 1020a024
    ZCMPLS P3.Z, Z12.S, $42, P5.S                  // 95adaa24
    ZCMPLS P7.Z, Z31.S, $127, P15.S                 // ffffbf24
    ZCMPLS P0.Z, Z0.D, $0, P0.D                    // 1020e024
    ZCMPLS P3.Z, Z12.D, $42, P5.D                  // 95adea24
    ZCMPLS P7.Z, Z31.D, $127, P15.D                 // ffffff24

// CMPLT   <Pd>.<T>, <Pg>/Z, <Zn>.<T>, #<imm>
    ZCMPLT P0.Z, Z0.B, $-16, P0.B                  // 00201025
    ZCMPLT P3.Z, Z12.B, $-6, P5.B                  // 852d1a25
    ZCMPLT P7.Z, Z31.B, $15, P15.B                  // ef3f0f25
    ZCMPLT P0.Z, Z0.H, $-16, P0.H                  // 00205025
    ZCMPLT P3.Z, Z12.H, $-6, P5.H                  // 852d5a25
    ZCMPLT P7.Z, Z31.H, $15, P15.H                  // ef3f4f25
    ZCMPLT P0.Z, Z0.S, $-16, P0.S                  // 00209025
    ZCMPLT P3.Z, Z12.S, $-6, P5.S                  // 852d9a25
    ZCMPLT P7.Z, Z31.S, $15, P15.S                  // ef3f8f25
    ZCMPLT P0.Z, Z0.D, $-16, P0.D                  // 0020d025
    ZCMPLT P3.Z, Z12.D, $-6, P5.D                  // 852dda25
    ZCMPLT P7.Z, Z31.D, $15, P15.D                  // ef3fcf25

// CMPNE   <Pd>.<T>, <Pg>/Z, <Zn>.<T>, #<imm>
    ZCMPNE P0.Z, Z0.B, $-16, P0.B                  // 10801025
    ZCMPNE P3.Z, Z12.B, $-6, P5.B                  // 958d1a25
    ZCMPNE P7.Z, Z31.B, $15, P15.B                  // ff9f0f25
    ZCMPNE P0.Z, Z0.H, $-16, P0.H                  // 10805025
    ZCMPNE P3.Z, Z12.H, $-6, P5.H                  // 958d5a25
    ZCMPNE P7.Z, Z31.H, $15, P15.H                  // ff9f4f25
    ZCMPNE P0.Z, Z0.S, $-16, P0.S                  // 10809025
    ZCMPNE P3.Z, Z12.S, $-6, P5.S                  // 958d9a25
    ZCMPNE P7.Z, Z31.S, $15, P15.S                  // ff9f8f25
    ZCMPNE P0.Z, Z0.D, $-16, P0.D                  // 1080d025
    ZCMPNE P3.Z, Z12.D, $-6, P5.D                  // 958dda25
    ZCMPNE P7.Z, Z31.D, $15, P15.D                  // ff9fcf25

// CMPEQ   <Pd>.<T>, <Pg>/Z, <Zn>.<T>, <Zm>.<T>
    ZCMPEQ P0.Z, Z0.B, Z0.B, P0.B                    // 00a00024
    ZCMPEQ P3.Z, Z12.B, Z13.B, P5.B                  // 85ad0d24
    ZCMPEQ P7.Z, Z31.B, Z31.B, P15.B                 // efbf1f24
    ZCMPEQ P0.Z, Z0.H, Z0.H, P0.H                    // 00a04024
    ZCMPEQ P3.Z, Z12.H, Z13.H, P5.H                  // 85ad4d24
    ZCMPEQ P7.Z, Z31.H, Z31.H, P15.H                 // efbf5f24
    ZCMPEQ P0.Z, Z0.S, Z0.S, P0.S                    // 00a08024
    ZCMPEQ P3.Z, Z12.S, Z13.S, P5.S                  // 85ad8d24
    ZCMPEQ P7.Z, Z31.S, Z31.S, P15.S                 // efbf9f24
    ZCMPEQ P0.Z, Z0.D, Z0.D, P0.D                    // 00a0c024
    ZCMPEQ P3.Z, Z12.D, Z13.D, P5.D                  // 85adcd24
    ZCMPEQ P7.Z, Z31.D, Z31.D, P15.D                 // efbfdf24

// CMPGE   <Pd>.<T>, <Pg>/Z, <Zn>.<T>, <Zm>.<T>
    ZCMPGE P0.Z, Z0.B, Z0.B, P0.B                    // 00800024
    ZCMPGE P3.Z, Z12.B, Z13.B, P5.B                  // 858d0d24
    ZCMPGE P7.Z, Z31.B, Z31.B, P15.B                 // ef9f1f24
    ZCMPGE P0.Z, Z0.H, Z0.H, P0.H                    // 00804024
    ZCMPGE P3.Z, Z12.H, Z13.H, P5.H                  // 858d4d24
    ZCMPGE P7.Z, Z31.H, Z31.H, P15.H                 // ef9f5f24
    ZCMPGE P0.Z, Z0.S, Z0.S, P0.S                    // 00808024
    ZCMPGE P3.Z, Z12.S, Z13.S, P5.S                  // 858d8d24
    ZCMPGE P7.Z, Z31.S, Z31.S, P15.S                 // ef9f9f24
    ZCMPGE P0.Z, Z0.D, Z0.D, P0.D                    // 0080c024
    ZCMPGE P3.Z, Z12.D, Z13.D, P5.D                  // 858dcd24
    ZCMPGE P7.Z, Z31.D, Z31.D, P15.D                 // ef9fdf24

// CMPGT   <Pd>.<T>, <Pg>/Z, <Zn>.<T>, <Zm>.<T>
    ZCMPGT P0.Z, Z0.B, Z0.B, P0.B                    // 10800024
    ZCMPGT P3.Z, Z12.B, Z13.B, P5.B                  // 958d0d24
    ZCMPGT P7.Z, Z31.B, Z31.B, P15.B                 // ff9f1f24
    ZCMPGT P0.Z, Z0.H, Z0.H, P0.H                    // 10804024
    ZCMPGT P3.Z, Z12.H, Z13.H, P5.H                  // 958d4d24
    ZCMPGT P7.Z, Z31.H, Z31.H, P15.H                 // ff9f5f24
    ZCMPGT P0.Z, Z0.S, Z0.S, P0.S                    // 10808024
    ZCMPGT P3.Z, Z12.S, Z13.S, P5.S                  // 958d8d24
    ZCMPGT P7.Z, Z31.S, Z31.S, P15.S                 // ff9f9f24
    ZCMPGT P0.Z, Z0.D, Z0.D, P0.D                    // 1080c024
    ZCMPGT P3.Z, Z12.D, Z13.D, P5.D                  // 958dcd24
    ZCMPGT P7.Z, Z31.D, Z31.D, P15.D                 // ff9fdf24

// CMPHI   <Pd>.<T>, <Pg>/Z, <Zn>.<T>, <Zm>.<T>
    ZCMPHI P0.Z, Z0.B, Z0.B, P0.B                    // 10000024
    ZCMPHI P3.Z, Z12.B, Z13.B, P5.B                  // 950d0d24
    ZCMPHI P7.Z, Z31.B, Z31.B, P15.B                 // ff1f1f24
    ZCMPHI P0.Z, Z0.H, Z0.H, P0.H                    // 10004024
    ZCMPHI P3.Z, Z12.H, Z13.H, P5.H                  // 950d4d24
    ZCMPHI P7.Z, Z31.H, Z31.H, P15.H                 // ff1f5f24
    ZCMPHI P0.Z, Z0.S, Z0.S, P0.S                    // 10008024
    ZCMPHI P3.Z, Z12.S, Z13.S, P5.S                  // 950d8d24
    ZCMPHI P7.Z, Z31.S, Z31.S, P15.S                 // ff1f9f24
    ZCMPHI P0.Z, Z0.D, Z0.D, P0.D                    // 1000c024
    ZCMPHI P3.Z, Z12.D, Z13.D, P5.D                  // 950dcd24
    ZCMPHI P7.Z, Z31.D, Z31.D, P15.D                 // ff1fdf24

// CMPHS   <Pd>.<T>, <Pg>/Z, <Zn>.<T>, <Zm>.<T>
    ZCMPHS P0.Z, Z0.B, Z0.B, P0.B                    // 00000024
    ZCMPHS P3.Z, Z12.B, Z13.B, P5.B                  // 850d0d24
    ZCMPHS P7.Z, Z31.B, Z31.B, P15.B                 // ef1f1f24
    ZCMPHS P0.Z, Z0.H, Z0.H, P0.H                    // 00004024
    ZCMPHS P3.Z, Z12.H, Z13.H, P5.H                  // 850d4d24
    ZCMPHS P7.Z, Z31.H, Z31.H, P15.H                 // ef1f5f24
    ZCMPHS P0.Z, Z0.S, Z0.S, P0.S                    // 00008024
    ZCMPHS P3.Z, Z12.S, Z13.S, P5.S                  // 850d8d24
    ZCMPHS P7.Z, Z31.S, Z31.S, P15.S                 // ef1f9f24
    ZCMPHS P0.Z, Z0.D, Z0.D, P0.D                    // 0000c024
    ZCMPHS P3.Z, Z12.D, Z13.D, P5.D                  // 850dcd24
    ZCMPHS P7.Z, Z31.D, Z31.D, P15.D                 // ef1fdf24

// CMPNE   <Pd>.<T>, <Pg>/Z, <Zn>.<T>, <Zm>.<T>
    ZCMPNE P0.Z, Z0.B, Z0.B, P0.B                    // 10a00024
    ZCMPNE P3.Z, Z12.B, Z13.B, P5.B                  // 95ad0d24
    ZCMPNE P7.Z, Z31.B, Z31.B, P15.B                 // ffbf1f24
    ZCMPNE P0.Z, Z0.H, Z0.H, P0.H                    // 10a04024
    ZCMPNE P3.Z, Z12.H, Z13.H, P5.H                  // 95ad4d24
    ZCMPNE P7.Z, Z31.H, Z31.H, P15.H                 // ffbf5f24
    ZCMPNE P0.Z, Z0.S, Z0.S, P0.S                    // 10a08024
    ZCMPNE P3.Z, Z12.S, Z13.S, P5.S                  // 95ad8d24
    ZCMPNE P7.Z, Z31.S, Z31.S, P15.S                 // ffbf9f24
    ZCMPNE P0.Z, Z0.D, Z0.D, P0.D                    // 10a0c024
    ZCMPNE P3.Z, Z12.D, Z13.D, P5.D                  // 95adcd24
    ZCMPNE P7.Z, Z31.D, Z31.D, P15.D                 // ffbfdf24

// FACGE   <Pd>.<T>, <Pg>/Z, <Zn>.<T>, <Zm>.<T>
    ZFACGE P0.Z, Z0.H, Z0.H, P0.H                    // 10c04065
    ZFACGE P3.Z, Z12.H, Z13.H, P5.H                  // 95cd4d65
    ZFACGE P7.Z, Z31.H, Z31.H, P15.H                 // ffdf5f65
    ZFACGE P0.Z, Z0.S, Z0.S, P0.S                    // 10c08065
    ZFACGE P3.Z, Z12.S, Z13.S, P5.S                  // 95cd8d65
    ZFACGE P7.Z, Z31.S, Z31.S, P15.S                 // ffdf9f65
    ZFACGE P0.Z, Z0.D, Z0.D, P0.D                    // 10c0c065
    ZFACGE P3.Z, Z12.D, Z13.D, P5.D                  // 95cdcd65
    ZFACGE P7.Z, Z31.D, Z31.D, P15.D                 // ffdfdf65

// FACGT   <Pd>.<T>, <Pg>/Z, <Zn>.<T>, <Zm>.<T>
    ZFACGT P0.Z, Z0.H, Z0.H, P0.H                    // 10e04065
    ZFACGT P3.Z, Z12.H, Z13.H, P5.H                  // 95ed4d65
    ZFACGT P7.Z, Z31.H, Z31.H, P15.H                 // ffff5f65
    ZFACGT P0.Z, Z0.S, Z0.S, P0.S                    // 10e08065
    ZFACGT P3.Z, Z12.S, Z13.S, P5.S                  // 95ed8d65
    ZFACGT P7.Z, Z31.S, Z31.S, P15.S                 // ffff9f65
    ZFACGT P0.Z, Z0.D, Z0.D, P0.D                    // 10e0c065
    ZFACGT P3.Z, Z12.D, Z13.D, P5.D                  // 95edcd65
    ZFACGT P7.Z, Z31.D, Z31.D, P15.D                 // ffffdf65

// FCMUO   <Pd>.<T>, <Pg>/Z, <Zn>.<T>, <Zm>.<T>
    ZFCMUO P0.Z, Z0.H, Z0.H, P0.H                    // 00c04065
    ZFCMUO P3.Z, Z12.H, Z13.H, P5.H                  // 85cd4d65
    ZFCMUO P7.Z, Z31.H, Z31.H, P15.H                 // efdf5f65
    ZFCMUO P0.Z, Z0.S, Z0.S, P0.S                    // 00c08065
    ZFCMUO P3.Z, Z12.S, Z13.S, P5.S                  // 85cd8d65
    ZFCMUO P7.Z, Z31.S, Z31.S, P15.S                 // efdf9f65
    ZFCMUO P0.Z, Z0.D, Z0.D, P0.D                    // 00c0c065
    ZFCMUO P3.Z, Z12.D, Z13.D, P5.D                  // 85cdcd65
    ZFCMUO P7.Z, Z31.D, Z31.D, P15.D                 // efdfdf65

// FCMEQ   <Pd>.<T>, <Pg>/Z, <Zn>.<T>, <Zm>.<T>
    ZFCMEQ P0.Z, Z0.H, Z0.H, P0.H                    // 00604065
    ZFCMEQ P3.Z, Z12.H, Z13.H, P5.H                  // 856d4d65
    ZFCMEQ P7.Z, Z31.H, Z31.H, P15.H                 // ef7f5f65
    ZFCMEQ P0.Z, Z0.S, Z0.S, P0.S                    // 00608065
    ZFCMEQ P3.Z, Z12.S, Z13.S, P5.S                  // 856d8d65
    ZFCMEQ P7.Z, Z31.S, Z31.S, P15.S                 // ef7f9f65
    ZFCMEQ P0.Z, Z0.D, Z0.D, P0.D                    // 0060c065
    ZFCMEQ P3.Z, Z12.D, Z13.D, P5.D                  // 856dcd65
    ZFCMEQ P7.Z, Z31.D, Z31.D, P15.D                 // ef7fdf65

// FCMGE   <Pd>.<T>, <Pg>/Z, <Zn>.<T>, <Zm>.<T>
    ZFCMGE P0.Z, Z0.H, Z0.H, P0.H                    // 00404065
    ZFCMGE P3.Z, Z12.H, Z13.H, P5.H                  // 854d4d65
    ZFCMGE P7.Z, Z31.H, Z31.H, P15.H                 // ef5f5f65
    ZFCMGE P0.Z, Z0.S, Z0.S, P0.S                    // 00408065
    ZFCMGE P3.Z, Z12.S, Z13.S, P5.S                  // 854d8d65
    ZFCMGE P7.Z, Z31.S, Z31.S, P15.S                 // ef5f9f65
    ZFCMGE P0.Z, Z0.D, Z0.D, P0.D                    // 0040c065
    ZFCMGE P3.Z, Z12.D, Z13.D, P5.D                  // 854dcd65
    ZFCMGE P7.Z, Z31.D, Z31.D, P15.D                 // ef5fdf65

// FCMGT   <Pd>.<T>, <Pg>/Z, <Zn>.<T>, <Zm>.<T>
    ZFCMGT P0.Z, Z0.H, Z0.H, P0.H                    // 10404065
    ZFCMGT P3.Z, Z12.H, Z13.H, P5.H                  // 954d4d65
    ZFCMGT P7.Z, Z31.H, Z31.H, P15.H                 // ff5f5f65
    ZFCMGT P0.Z, Z0.S, Z0.S, P0.S                    // 10408065
    ZFCMGT P3.Z, Z12.S, Z13.S, P5.S                  // 954d8d65
    ZFCMGT P7.Z, Z31.S, Z31.S, P15.S                 // ff5f9f65
    ZFCMGT P0.Z, Z0.D, Z0.D, P0.D                    // 1040c065
    ZFCMGT P3.Z, Z12.D, Z13.D, P5.D                  // 954dcd65
    ZFCMGT P7.Z, Z31.D, Z31.D, P15.D                 // ff5fdf65

// FCMNE   <Pd>.<T>, <Pg>/Z, <Zn>.<T>, <Zm>.<T>
    ZFCMNE P0.Z, Z0.H, Z0.H, P0.H                    // 10604065
    ZFCMNE P3.Z, Z12.H, Z13.H, P5.H                  // 956d4d65
    ZFCMNE P7.Z, Z31.H, Z31.H, P15.H                 // ff7f5f65
    ZFCMNE P0.Z, Z0.S, Z0.S, P0.S                    // 10608065
    ZFCMNE P3.Z, Z12.S, Z13.S, P5.S                  // 956d8d65
    ZFCMNE P7.Z, Z31.S, Z31.S, P15.S                 // ff7f9f65
    ZFCMNE P0.Z, Z0.D, Z0.D, P0.D                    // 1060c065
    ZFCMNE P3.Z, Z12.D, Z13.D, P5.D                  // 956dcd65
    ZFCMNE P7.Z, Z31.D, Z31.D, P15.D                 // ff7fdf65

// FCMUO   <Pd>.<T>, <Pg>/Z, <Zn>.<T>, <Zm>.<T>
    ZFCMUO P0.Z, Z0.H, Z0.H, P0.H                    // 00c04065
    ZFCMUO P3.Z, Z12.H, Z13.H, P5.H                  // 85cd4d65
    ZFCMUO P7.Z, Z31.H, Z31.H, P15.H                 // efdf5f65
    ZFCMUO P0.Z, Z0.S, Z0.S, P0.S                    // 00c08065
    ZFCMUO P3.Z, Z12.S, Z13.S, P5.S                  // 85cd8d65
    ZFCMUO P7.Z, Z31.S, Z31.S, P15.S                 // efdf9f65
    ZFCMUO P0.Z, Z0.D, Z0.D, P0.D                    // 00c0c065
    ZFCMUO P3.Z, Z12.D, Z13.D, P5.D                  // 85cdcd65
    ZFCMUO P7.Z, Z31.D, Z31.D, P15.D                 // efdfdf65

// CMPEQ   <Pd>.<T>, <Pg>/Z, <Zn>.<T>, <Zm>.D
    ZCMPEQ P0.Z, Z0.B, Z0.D, P0.B                    // 00200024
    ZCMPEQ P3.Z, Z12.B, Z13.D, P5.B                  // 852d0d24
    ZCMPEQ P7.Z, Z31.B, Z31.D, P15.B                 // ef3f1f24
    ZCMPEQ P0.Z, Z0.H, Z0.D, P0.H                    // 00204024
    ZCMPEQ P3.Z, Z12.H, Z13.D, P5.H                  // 852d4d24
    ZCMPEQ P7.Z, Z31.H, Z31.D, P15.H                 // ef3f5f24
    ZCMPEQ P0.Z, Z0.S, Z0.D, P0.S                    // 00208024
    ZCMPEQ P3.Z, Z12.S, Z13.D, P5.S                  // 852d8d24
    ZCMPEQ P7.Z, Z31.S, Z31.D, P15.S                 // ef3f9f24

// CMPGE   <Pd>.<T>, <Pg>/Z, <Zn>.<T>, <Zm>.D
    ZCMPGE P0.Z, Z0.B, Z0.D, P0.B                    // 00400024
    ZCMPGE P3.Z, Z12.B, Z13.D, P5.B                  // 854d0d24
    ZCMPGE P7.Z, Z31.B, Z31.D, P15.B                 // ef5f1f24
    ZCMPGE P0.Z, Z0.H, Z0.D, P0.H                    // 00404024
    ZCMPGE P3.Z, Z12.H, Z13.D, P5.H                  // 854d4d24
    ZCMPGE P7.Z, Z31.H, Z31.D, P15.H                 // ef5f5f24
    ZCMPGE P0.Z, Z0.S, Z0.D, P0.S                    // 00408024
    ZCMPGE P3.Z, Z12.S, Z13.D, P5.S                  // 854d8d24
    ZCMPGE P7.Z, Z31.S, Z31.D, P15.S                 // ef5f9f24

// CMPGT   <Pd>.<T>, <Pg>/Z, <Zn>.<T>, <Zm>.D
    ZCMPGT P0.Z, Z0.B, Z0.D, P0.B                    // 10400024
    ZCMPGT P3.Z, Z12.B, Z13.D, P5.B                  // 954d0d24
    ZCMPGT P7.Z, Z31.B, Z31.D, P15.B                 // ff5f1f24
    ZCMPGT P0.Z, Z0.H, Z0.D, P0.H                    // 10404024
    ZCMPGT P3.Z, Z12.H, Z13.D, P5.H                  // 954d4d24
    ZCMPGT P7.Z, Z31.H, Z31.D, P15.H                 // ff5f5f24
    ZCMPGT P0.Z, Z0.S, Z0.D, P0.S                    // 10408024
    ZCMPGT P3.Z, Z12.S, Z13.D, P5.S                  // 954d8d24
    ZCMPGT P7.Z, Z31.S, Z31.D, P15.S                 // ff5f9f24

// CMPHI   <Pd>.<T>, <Pg>/Z, <Zn>.<T>, <Zm>.D
    ZCMPHI P0.Z, Z0.B, Z0.D, P0.B                    // 10c00024
    ZCMPHI P3.Z, Z12.B, Z13.D, P5.B                  // 95cd0d24
    ZCMPHI P7.Z, Z31.B, Z31.D, P15.B                 // ffdf1f24
    ZCMPHI P0.Z, Z0.H, Z0.D, P0.H                    // 10c04024
    ZCMPHI P3.Z, Z12.H, Z13.D, P5.H                  // 95cd4d24
    ZCMPHI P7.Z, Z31.H, Z31.D, P15.H                 // ffdf5f24
    ZCMPHI P0.Z, Z0.S, Z0.D, P0.S                    // 10c08024
    ZCMPHI P3.Z, Z12.S, Z13.D, P5.S                  // 95cd8d24
    ZCMPHI P7.Z, Z31.S, Z31.D, P15.S                 // ffdf9f24

// CMPHS   <Pd>.<T>, <Pg>/Z, <Zn>.<T>, <Zm>.D
    ZCMPHS P0.Z, Z0.B, Z0.D, P0.B                    // 00c00024
    ZCMPHS P3.Z, Z12.B, Z13.D, P5.B                  // 85cd0d24
    ZCMPHS P7.Z, Z31.B, Z31.D, P15.B                 // efdf1f24
    ZCMPHS P0.Z, Z0.H, Z0.D, P0.H                    // 00c04024
    ZCMPHS P3.Z, Z12.H, Z13.D, P5.H                  // 85cd4d24
    ZCMPHS P7.Z, Z31.H, Z31.D, P15.H                 // efdf5f24
    ZCMPHS P0.Z, Z0.S, Z0.D, P0.S                    // 00c08024
    ZCMPHS P3.Z, Z12.S, Z13.D, P5.S                  // 85cd8d24
    ZCMPHS P7.Z, Z31.S, Z31.D, P15.S                 // efdf9f24

// CMPLE   <Pd>.<T>, <Pg>/Z, <Zn>.<T>, <Zm>.D
    ZCMPLE P0.Z, Z0.B, Z0.D, P0.B                    // 10600024
    ZCMPLE P3.Z, Z12.B, Z13.D, P5.B                  // 956d0d24
    ZCMPLE P7.Z, Z31.B, Z31.D, P15.B                 // ff7f1f24
    ZCMPLE P0.Z, Z0.H, Z0.D, P0.H                    // 10604024
    ZCMPLE P3.Z, Z12.H, Z13.D, P5.H                  // 956d4d24
    ZCMPLE P7.Z, Z31.H, Z31.D, P15.H                 // ff7f5f24
    ZCMPLE P0.Z, Z0.S, Z0.D, P0.S                    // 10608024
    ZCMPLE P3.Z, Z12.S, Z13.D, P5.S                  // 956d8d24
    ZCMPLE P7.Z, Z31.S, Z31.D, P15.S                 // ff7f9f24

// CMPLO   <Pd>.<T>, <Pg>/Z, <Zn>.<T>, <Zm>.D
    ZCMPLO P0.Z, Z0.B, Z0.D, P0.B                    // 00e00024
    ZCMPLO P3.Z, Z12.B, Z13.D, P5.B                  // 85ed0d24
    ZCMPLO P7.Z, Z31.B, Z31.D, P15.B                 // efff1f24
    ZCMPLO P0.Z, Z0.H, Z0.D, P0.H                    // 00e04024
    ZCMPLO P3.Z, Z12.H, Z13.D, P5.H                  // 85ed4d24
    ZCMPLO P7.Z, Z31.H, Z31.D, P15.H                 // efff5f24
    ZCMPLO P0.Z, Z0.S, Z0.D, P0.S                    // 00e08024
    ZCMPLO P3.Z, Z12.S, Z13.D, P5.S                  // 85ed8d24
    ZCMPLO P7.Z, Z31.S, Z31.D, P15.S                 // efff9f24

// CMPLS   <Pd>.<T>, <Pg>/Z, <Zn>.<T>, <Zm>.D
    ZCMPLS P0.Z, Z0.B, Z0.D, P0.B                    // 10e00024
    ZCMPLS P3.Z, Z12.B, Z13.D, P5.B                  // 95ed0d24
    ZCMPLS P7.Z, Z31.B, Z31.D, P15.B                 // ffff1f24
    ZCMPLS P0.Z, Z0.H, Z0.D, P0.H                    // 10e04024
    ZCMPLS P3.Z, Z12.H, Z13.D, P5.H                  // 95ed4d24
    ZCMPLS P7.Z, Z31.H, Z31.D, P15.H                 // ffff5f24
    ZCMPLS P0.Z, Z0.S, Z0.D, P0.S                    // 10e08024
    ZCMPLS P3.Z, Z12.S, Z13.D, P5.S                  // 95ed8d24
    ZCMPLS P7.Z, Z31.S, Z31.D, P15.S                 // ffff9f24

// CMPLT   <Pd>.<T>, <Pg>/Z, <Zn>.<T>, <Zm>.D
    ZCMPLT P0.Z, Z0.B, Z0.D, P0.B                    // 00600024
    ZCMPLT P3.Z, Z12.B, Z13.D, P5.B                  // 856d0d24
    ZCMPLT P7.Z, Z31.B, Z31.D, P15.B                 // ef7f1f24
    ZCMPLT P0.Z, Z0.H, Z0.D, P0.H                    // 00604024
    ZCMPLT P3.Z, Z12.H, Z13.D, P5.H                  // 856d4d24
    ZCMPLT P7.Z, Z31.H, Z31.D, P15.H                 // ef7f5f24
    ZCMPLT P0.Z, Z0.S, Z0.D, P0.S                    // 00608024
    ZCMPLT P3.Z, Z12.S, Z13.D, P5.S                  // 856d8d24
    ZCMPLT P7.Z, Z31.S, Z31.D, P15.S                 // ef7f9f24

// CMPNE   <Pd>.<T>, <Pg>/Z, <Zn>.<T>, <Zm>.D
    ZCMPNE P0.Z, Z0.B, Z0.D, P0.B                    // 10200024
    ZCMPNE P3.Z, Z12.B, Z13.D, P5.B                  // 952d0d24
    ZCMPNE P7.Z, Z31.B, Z31.D, P15.B                 // ff3f1f24
    ZCMPNE P0.Z, Z0.H, Z0.D, P0.H                    // 10204024
    ZCMPNE P3.Z, Z12.H, Z13.D, P5.H                  // 952d4d24
    ZCMPNE P7.Z, Z31.H, Z31.D, P15.H                 // ff3f5f24
    ZCMPNE P0.Z, Z0.S, Z0.D, P0.S                    // 10208024
    ZCMPNE P3.Z, Z12.S, Z13.D, P5.S                  // 952d8d24
    ZCMPNE P7.Z, Z31.S, Z31.D, P15.S                 // ff3f9f24

// REV     <Pd>.<T>, <Pn>.<T>
    ZREV P0.B, P0.B                                  // 00403405
    ZREV P6.B, P5.B                                  // c5403405
    ZREV P15.B, P15.B                                // ef413405
    ZREV P0.H, P0.H                                  // 00407405
    ZREV P6.H, P5.H                                  // c5407405
    ZREV P15.H, P15.H                                // ef417405
    ZREV P0.S, P0.S                                  // 0040b405
    ZREV P6.S, P5.S                                  // c540b405
    ZREV P15.S, P15.S                                // ef41b405
    ZREV P0.D, P0.D                                  // 0040f405
    ZREV P6.D, P5.D                                  // c540f405
    ZREV P15.D, P15.D                                // ef41f405

// TRN1    <Pd>.<T>, <Pn>.<T>, <Pm>.<T>
    PTRN1 P0.B, P0.B, P0.B                           // 00502005
    PTRN1 P6.B, P7.B, P5.B                           // c5502705
    PTRN1 P15.B, P15.B, P15.B                        // ef512f05
    PTRN1 P0.H, P0.H, P0.H                           // 00506005
    PTRN1 P6.H, P7.H, P5.H                           // c5506705
    PTRN1 P15.H, P15.H, P15.H                        // ef516f05
    PTRN1 P0.S, P0.S, P0.S                           // 0050a005
    PTRN1 P6.S, P7.S, P5.S                           // c550a705
    PTRN1 P15.S, P15.S, P15.S                        // ef51af05
    PTRN1 P0.D, P0.D, P0.D                           // 0050e005
    PTRN1 P6.D, P7.D, P5.D                           // c550e705
    PTRN1 P15.D, P15.D, P15.D                        // ef51ef05

// TRN2    <Pd>.<T>, <Pn>.<T>, <Pm>.<T>
    PTRN2 P0.B, P0.B, P0.B                           // 00542005
    PTRN2 P6.B, P7.B, P5.B                           // c5542705
    PTRN2 P15.B, P15.B, P15.B                        // ef552f05
    PTRN2 P0.H, P0.H, P0.H                           // 00546005
    PTRN2 P6.H, P7.H, P5.H                           // c5546705
    PTRN2 P15.H, P15.H, P15.H                        // ef556f05
    PTRN2 P0.S, P0.S, P0.S                           // 0054a005
    PTRN2 P6.S, P7.S, P5.S                           // c554a705
    PTRN2 P15.S, P15.S, P15.S                        // ef55af05
    PTRN2 P0.D, P0.D, P0.D                           // 0054e005
    PTRN2 P6.D, P7.D, P5.D                           // c554e705
    PTRN2 P15.D, P15.D, P15.D                        // ef55ef05

// UZP1    <Pd>.<T>, <Pn>.<T>, <Pm>.<T>
    PUZP1 P0.B, P0.B, P0.B                           // 00482005
    PUZP1 P6.B, P7.B, P5.B                           // c5482705
    PUZP1 P15.B, P15.B, P15.B                        // ef492f05
    PUZP1 P0.H, P0.H, P0.H                           // 00486005
    PUZP1 P6.H, P7.H, P5.H                           // c5486705
    PUZP1 P15.H, P15.H, P15.H                        // ef496f05
    PUZP1 P0.S, P0.S, P0.S                           // 0048a005
    PUZP1 P6.S, P7.S, P5.S                           // c548a705
    PUZP1 P15.S, P15.S, P15.S                        // ef49af05
    PUZP1 P0.D, P0.D, P0.D                           // 0048e005
    PUZP1 P6.D, P7.D, P5.D                           // c548e705
    PUZP1 P15.D, P15.D, P15.D                        // ef49ef05

// UZP2    <Pd>.<T>, <Pn>.<T>, <Pm>.<T>
    PUZP2 P0.B, P0.B, P0.B                           // 004c2005
    PUZP2 P6.B, P7.B, P5.B                           // c54c2705
    PUZP2 P15.B, P15.B, P15.B                        // ef4d2f05
    PUZP2 P0.H, P0.H, P0.H                           // 004c6005
    PUZP2 P6.H, P7.H, P5.H                           // c54c6705
    PUZP2 P15.H, P15.H, P15.H                        // ef4d6f05
    PUZP2 P0.S, P0.S, P0.S                           // 004ca005
    PUZP2 P6.S, P7.S, P5.S                           // c54ca705
    PUZP2 P15.S, P15.S, P15.S                        // ef4daf05
    PUZP2 P0.D, P0.D, P0.D                           // 004ce005
    PUZP2 P6.D, P7.D, P5.D                           // c54ce705
    PUZP2 P15.D, P15.D, P15.D                        // ef4def05

// ZIP1    <Pd>.<T>, <Pn>.<T>, <Pm>.<T>
    PZIP1 P0.B, P0.B, P0.B                           // 00402005
    PZIP1 P6.B, P7.B, P5.B                           // c5402705
    PZIP1 P15.B, P15.B, P15.B                        // ef412f05
    PZIP1 P0.H, P0.H, P0.H                           // 00406005
    PZIP1 P6.H, P7.H, P5.H                           // c5406705
    PZIP1 P15.H, P15.H, P15.H                        // ef416f05
    PZIP1 P0.S, P0.S, P0.S                           // 0040a005
    PZIP1 P6.S, P7.S, P5.S                           // c540a705
    PZIP1 P15.S, P15.S, P15.S                        // ef41af05
    PZIP1 P0.D, P0.D, P0.D                           // 0040e005
    PZIP1 P6.D, P7.D, P5.D                           // c540e705
    PZIP1 P15.D, P15.D, P15.D                        // ef41ef05

// ZIP2    <Pd>.<T>, <Pn>.<T>, <Pm>.<T>
    PZIP2 P0.B, P0.B, P0.B                           // 00442005
    PZIP2 P6.B, P7.B, P5.B                           // c5442705
    PZIP2 P15.B, P15.B, P15.B                        // ef452f05
    PZIP2 P0.H, P0.H, P0.H                           // 00446005
    PZIP2 P6.H, P7.H, P5.H                           // c5446705
    PZIP2 P15.H, P15.H, P15.H                        // ef456f05
    PZIP2 P0.S, P0.S, P0.S                           // 0044a005
    PZIP2 P6.S, P7.S, P5.S                           // c544a705
    PZIP2 P15.S, P15.S, P15.S                        // ef45af05
    PZIP2 P0.D, P0.D, P0.D                           // 0044e005
    PZIP2 P6.D, P7.D, P5.D                           // c544e705
    PZIP2 P15.D, P15.D, P15.D                        // ef45ef05

// WHILELE <Pd>.<T>, <R><n>, <R><m>
    WHILELEW R0, R0, P0.B                           // 10042025
    WHILELEW R11, R12, P5.B                         // 75052c25
    WHILELEW R30, R30, P15.B                        // df073e25
    WHILELE  R0, R0, P0.B                           // 10142025
    WHILELE  R11, R12, P5.B                         // 75152c25
    WHILELE  R30, R30, P15.B                        // df173e25
    WHILELEW R0, R0, P0.H                           // 10046025
    WHILELEW R11, R12, P5.H                         // 75056c25
    WHILELEW R30, R30, P15.H                        // df077e25
    WHILELE  R0, R0, P0.H                           // 10146025
    WHILELE  R11, R12, P5.H                         // 75156c25
    WHILELE  R30, R30, P15.H                        // df177e25
    WHILELEW R0, R0, P0.S                           // 1004a025
    WHILELEW R11, R12, P5.S                         // 7505ac25
    WHILELEW R30, R30, P15.S                        // df07be25
    WHILELE  R0, R0, P0.S                           // 1014a025
    WHILELE  R11, R12, P5.S                         // 7515ac25
    WHILELE  R30, R30, P15.S                        // df17be25
    WHILELEW R0, R0, P0.D                           // 1004e025
    WHILELEW R11, R12, P5.D                         // 7505ec25
    WHILELEW R30, R30, P15.D                        // df07fe25
    WHILELE R0, R0, P0.D                            // 1014e025
    WHILELE R11, R12, P5.D                          // 7515ec25
    WHILELE R30, R30, P15.D                         // df17fe25

// WHILELO <Pd>.<T>, <R><n>, <R><m>
    WHILELOW R0, R0, P0.B                           // 000c2025
    WHILELOW R11, R12, P5.B                         // 650d2c25
    WHILELOW R30, R30, P15.B                        // cf0f3e25
    WHILELO R0, R0, P0.B                            // 001c2025
    WHILELO R11, R12, P5.B                          // 651d2c25
    WHILELO R30, R30, P15.B                         // cf1f3e25
    WHILELOW R0, R0, P0.H                           // 000c6025
    WHILELOW R11, R12, P5.H                         // 650d6c25
    WHILELOW R30, R30, P15.H                        // cf0f7e25
    WHILELO  R0, R0, P0.H                           // 001c6025
    WHILELO  R11, R12, P5.H                         // 651d6c25
    WHILELO  R30, R30, P15.H                        // cf1f7e25
    WHILELOW R0, R0, P0.S                           // 000ca025
    WHILELOW R11, R12, P5.S                         // 650dac25
    WHILELOW R30, R30, P15.S                        // cf0fbe25
    WHILELO  R0, R0, P0.S                           // 001ca025
    WHILELO  R11, R12, P5.S                         // 651dac25
    WHILELO  R30, R30, P15.S                        // cf1fbe25
    WHILELOW R0, R0, P0.D                           // 000ce025
    WHILELOW R11, R12, P5.D                         // 650dec25
    WHILELOW R30, R30, P15.D                        // cf0ffe25
    WHILELO R0, R0, P0.D                            // 001ce025
    WHILELO R11, R12, P5.D                          // 651dec25
    WHILELO R30, R30, P15.D                         // cf1ffe25

// WHILELS <Pd>.<T>, <R><n>, <R><m>
    WHILELSW R0, R0, P0.B                           // 100c2025
    WHILELSW R11, R12, P5.B                         // 750d2c25
    WHILELSW R30, R30, P15.B                        // df0f3e25
    WHILELS  R0, R0, P0.B                           // 101c2025
    WHILELS  R11, R12, P5.B                         // 751d2c25
    WHILELS  R30, R30, P15.B                        // df1f3e25
    WHILELSW R0, R0, P0.H                           // 100c6025
    WHILELSW R11, R12, P5.H                         // 750d6c25
    WHILELSW R30, R30, P15.H                        // df0f7e25
    WHILELS  R0, R0, P0.H                           // 101c6025
    WHILELS  R11, R12, P5.H                         // 751d6c25
    WHILELS  R30, R30, P15.H                        // df1f7e25
    WHILELSW R0, R0, P0.S                           // 100ca025
    WHILELSW R11, R12, P5.S                         // 750dac25
    WHILELSW R30, R30, P15.S                        // df0fbe25
    WHILELS  R0, R0, P0.S                           // 101ca025
    WHILELS  R11, R12, P5.S                         // 751dac25
    WHILELS  R30, R30, P15.S                        // df1fbe25
    WHILELSW R0, R0, P0.D                           // 100ce025
    WHILELSW R11, R12, P5.D                         // 750dec25
    WHILELSW R30, R30, P15.D                        // df0ffe25
    WHILELS R0, R0, P0.D                            // 101ce025
    WHILELS R11, R12, P5.D                          // 751dec25
    WHILELS R30, R30, P15.D                         // df1ffe25

// WHILELT <Pd>.<T>, <R><n>, <R><m>
    WHILELTW R0, R0, P0.B                           // 00042025
    WHILELTW R11, R12, P5.B                         // 65052c25
    WHILELTW R30, R30, P15.B                        // cf073e25
    WHILELT  R0, R0, P0.B                           // 00142025
    WHILELT  R11, R12, P5.B                         // 65152c25
    WHILELT  R30, R30, P15.B                        // cf173e25
    WHILELTW R0, R0, P0.H                           // 00046025
    WHILELTW R11, R12, P5.H                         // 65056c25
    WHILELTW R30, R30, P15.H                        // cf077e25
    WHILELT  R0, R0, P0.H                           // 00146025
    WHILELT  R11, R12, P5.H                         // 65156c25
    WHILELT  R30, R30, P15.H                        // cf177e25
    WHILELTW R0, R0, P0.S                           // 0004a025
    WHILELTW R11, R12, P5.S                         // 6505ac25
    WHILELTW R30, R30, P15.S                        // cf07be25
    WHILELT  R0, R0, P0.S                           // 0014a025
    WHILELT  R11, R12, P5.S                         // 6515ac25
    WHILELT  R30, R30, P15.S                        // cf17be25
    WHILELTW R0, R0, P0.D                           // 0004e025
    WHILELTW R11, R12, P5.D                         // 6505ec25
    WHILELTW R30, R30, P15.D                        // cf07fe25
    WHILELT R0, R0, P0.D                            // 0014e025
    WHILELT R11, R12, P5.D                          // 6515ec25
    WHILELT R30, R30, P15.D                         // cf17fe25

// PFALSE  <Pd>.B
    PFALSE P0.B                                     // 00e41825
    PFALSE P5.B                                     // 05e41825
    PFALSE P15.B                                    // 0fe41825

// RDFFR   <Pd>.B
    RDFFR P0.B                                     // 00f01925
    RDFFR P5.B                                     // 05f01925
    RDFFR P15.B                                    // 0ff01925

// ANDS    <Pd>.B, <Pg>/Z, <Pn>.B, <Pm>.B
    PANDS P0.Z, P0.B, P0.B, P0.B                     // 00404025
    PANDS P6.Z, P7.B, P8.B, P5.B                     // e5584825
    PANDS P15.Z, P15.B, P15.B, P15.B                 // ef7d4f25

// BIC     <Pd>.B, <Pg>/Z, <Pn>.B, <Pm>.B
    PBIC P0.Z, P0.B, P0.B, P0.B                      // 10400025
    PBIC P6.Z, P7.B, P8.B, P5.B                      // f5580825
    PBIC P15.Z, P15.B, P15.B, P15.B                  // ff7d0f25

// BICS    <Pd>.B, <Pg>/Z, <Pn>.B, <Pm>.B
    PBICS P0.Z, P0.B, P0.B, P0.B                     // 10404025
    PBICS P6.Z, P7.B, P8.B, P5.B                     // f5584825
    PBICS P15.Z, P15.B, P15.B, P15.B                 // ff7d4f25

// BRKPA   <Pd>.B, <Pg>/Z, <Pn>.B, <Pm>.B
    BRKPA P0.Z, P0.B, P0.B, P0.B                    // 00c00025
    BRKPA P6.Z, P7.B, P8.B, P5.B                    // e5d80825
    BRKPA P15.Z, P15.B, P15.B, P15.B                // effd0f25

// BRKPAS  <Pd>.B, <Pg>/Z, <Pn>.B, <Pm>.B
    BRKPAS P0.Z, P0.B, P0.B, P0.B                   // 00c04025
    BRKPAS P6.Z, P7.B, P8.B, P5.B                   // e5d84825
    BRKPAS P15.Z, P15.B, P15.B, P15.B               // effd4f25

// BRKPB   <Pd>.B, <Pg>/Z, <Pn>.B, <Pm>.B
    BRKPB P0.Z, P0.B, P0.B, P0.B                    // 10c00025
    BRKPB P6.Z, P7.B, P8.B, P5.B                    // f5d80825
    BRKPB P15.Z, P15.B, P15.B, P15.B                // fffd0f25

// BRKPBS  <Pd>.B, <Pg>/Z, <Pn>.B, <Pm>.B
    BRKPBS P0.Z, P0.B, P0.B, P0.B                   // 10c04025
    BRKPBS P6.Z, P7.B, P8.B, P5.B                   // f5d84825
    BRKPBS P15.Z, P15.B, P15.B, P15.B               // fffd4f25

// EOR     <Pd>.B, <Pg>/Z, <Pn>.B, <Pm>.B
    PEOR P0.Z, P0.B, P0.B, P0.B                      // 00420025
    PEOR P6.Z, P7.B, P8.B, P5.B                      // e55a0825
    PEOR P15.Z, P15.B, P15.B, P15.B                  // ef7f0f25

// EORS    <Pd>.B, <Pg>/Z, <Pn>.B, <Pm>.B
    PEORS P0.Z, P0.B, P0.B, P0.B                     // 00424025
    PEORS P6.Z, P7.B, P8.B, P5.B                     // e55a4825
    PEORS P15.Z, P15.B, P15.B, P15.B                 // ef7f4f25

// NAND    <Pd>.B, <Pg>/Z, <Pn>.B, <Pm>.B
    PNAND P0.Z, P0.B, P0.B, P0.B                     // 10428025
    PNAND P6.Z, P7.B, P8.B, P5.B                     // f55a8825
    PNAND P15.Z, P15.B, P15.B, P15.B                 // ff7f8f25

// NANDS   <Pd>.B, <Pg>/Z, <Pn>.B, <Pm>.B
    PNANDS P0.Z, P0.B, P0.B, P0.B                    // 1042c025
    PNANDS P6.Z, P7.B, P8.B, P5.B                    // f55ac825
    PNANDS P15.Z, P15.B, P15.B, P15.B                // ff7fcf25

// NOR     <Pd>.B, <Pg>/Z, <Pn>.B, <Pm>.B
    PNOR P0.Z, P0.B, P0.B, P0.B                      // 00428025
    PNOR P6.Z, P7.B, P8.B, P5.B                      // e55a8825
    PNOR P15.Z, P15.B, P15.B, P15.B                  // ef7f8f25

// NORS    <Pd>.B, <Pg>/Z, <Pn>.B, <Pm>.B
    PNORS P0.Z, P0.B, P0.B, P0.B                     // 0042c025
    PNORS P6.Z, P7.B, P8.B, P5.B                     // e55ac825
    PNORS P15.Z, P15.B, P15.B, P15.B                 // ef7fcf25

// ORN     <Pd>.B, <Pg>/Z, <Pn>.B, <Pm>.B
    ORN P0.Z, P0.B, P0.B, P0.B                      // 10408025
    ORN P6.Z, P7.B, P8.B, P5.B                      // f5588825
    ORN P15.Z, P15.B, P15.B, P15.B                  // ff7d8f25

// ORNS    <Pd>.B, <Pg>/Z, <Pn>.B, <Pm>.B
    ORNS P0.Z, P0.B, P0.B, P0.B                     // 1040c025
    ORNS P6.Z, P7.B, P8.B, P5.B                     // f558c825
    ORNS P15.Z, P15.B, P15.B, P15.B                 // ff7dcf25

// ORR     <Pd>.B, <Pg>/Z, <Pn>.B, <Pm>.B
    PORR P0.Z, P0.B, P0.B, P0.B                      // 00408025
    PORR P6.Z, P7.B, P8.B, P5.B                      // e5588825
    PORR P15.Z, P15.B, P15.B, P15.B                  // ef7d8f25

// ORRS    <Pd>.B, <Pg>/Z, <Pn>.B, <Pm>.B
    PORRS P0.Z, P0.B, P0.B, P0.B                     // 0040c025
    PORRS P6.Z, P7.B, P8.B, P5.B                     // e558c825
    PORRS P15.Z, P15.B, P15.B, P15.B                 // ef7dcf25

// SEL     <Pd>.B, <Pg>, <Pn>.B, <Pm>.B
    PSEL P0, P0.B, P0.B, P0.B                        // 10420025
    PSEL P6, P7.B, P8.B, P5.B                        // f55a0825
    PSEL P15, P15.B, P15.B, P15.B                    // ff7f0f25

// RDFFR   <Pd>.B, <Pg>/Z
    RDFFR P0.Z, P0.B                                // 00f01825
    RDFFR P6.Z, P5.B                                // c5f01825
    RDFFR P15.Z, P15.B                              // eff11825

// RDFFRS  <Pd>.B, <Pg>/Z
    RDFFRS P0.Z, P0.B                               // 00f05825
    RDFFRS P6.Z, P5.B                               // c5f05825
    RDFFRS P15.Z, P15.B                             // eff15825

// PUNPKHI <Pd>.H, <Pn>.B
    PUNPKHI P0.B, P0.H                              // 00403105
    PUNPKHI P6.B, P5.H                              // c5403105
    PUNPKHI P15.B, P15.H                            // ef413105

// PUNPKLO <Pd>.H, <Pn>.B
    PUNPKLO P0.B, P0.H                              // 00403005
    PUNPKLO P6.B, P5.H                              // c5403005
    PUNPKLO P15.B, P15.H                            // ef413005

// BRKA    <Pd>.B, <Pg>/<ZM>, <Pn>.B
    BRKA P0.Z, P0.B, P0.B                           // 00401025
    BRKA P6.Z, P7.B, P5.B                           // e5581025
    BRKA P15.Z, P15.B, P15.B                        // ef7d1025
    BRKA P0.M, P0.B, P0.B                           // 10401025
    BRKA P6.M, P7.B, P5.B                           // f5581025
    BRKA P15.M, P15.B, P15.B                        // ff7d1025

// BRKB    <Pd>.B, <Pg>/<ZM>, <Pn>.B
    BRKB P0.Z, P0.B, P0.B                           // 00409025
    BRKB P6.Z, P7.B, P5.B                           // e5589025
    BRKB P15.Z, P15.B, P15.B                        // ef7d9025
    BRKB P0.M, P0.B, P0.B                           // 10409025
    BRKB P6.M, P7.B, P5.B                           // f5589025
    BRKB P15.M, P15.B, P15.B                        // ff7d9025

// BRKAS   <Pd>.B, <Pg>/Z, <Pn>.B
    BRKAS P0.Z, P0.B, P0.B                          // 00405025
    BRKAS P6.Z, P7.B, P5.B                          // e5585025
    BRKAS P15.Z, P15.B, P15.B                       // ef7d5025

// BRKBS   <Pd>.B, <Pg>/Z, <Pn>.B
    BRKBS P0.Z, P0.B, P0.B                          // 0040d025
    BRKBS P6.Z, P7.B, P5.B                          // e558d025
    BRKBS P15.Z, P15.B, P15.B                       // ef7dd025

// BRKN    <Pdm>.B, <Pg>/Z, <Pn>.B, <Pdm>.B
    BRKN P0.Z, P0.B, P0.B, P0.B                     // 00401825
    BRKN P6.Z, P7.B, P5.B, P5.B                     // e5581825
    BRKN P15.Z, P15.B, P15.B, P15.B                 // ef7d1825

// BRKNS   <Pdm>.B, <Pg>/Z, <Pn>.B, <Pdm>.B
    BRKNS P0.Z, P0.B, P0.B, P0.B                    // 00405825
    BRKNS P6.Z, P7.B, P5.B, P5.B                    // e5585825
    BRKNS P15.Z, P15.B, P15.B, P15.B                // ef7d5825

// PNEXT   <Pdn>.<T>, <Pv>, <Pdn>.<T>
    PNEXT P0, P0.B, P0.B                            // 00c41925
    PNEXT P6, P5.B, P5.B                            // c5c41925
    PNEXT P15, P15.B, P15.B                         // efc51925
    PNEXT P0, P0.H, P0.H                            // 00c45925
    PNEXT P6, P5.H, P5.H                            // c5c45925
    PNEXT P15, P15.H, P15.H                         // efc55925
    PNEXT P0, P0.S, P0.S                            // 00c49925
    PNEXT P6, P5.S, P5.S                            // c5c49925
    PNEXT P15, P15.S, P15.S                         // efc59925
    PNEXT P0, P0.D, P0.D                            // 00c4d925
    PNEXT P6, P5.D, P5.D                            // c5c4d925
    PNEXT P15, P15.D, P15.D                         // efc5d925

// PFIRST  <Pdn>.B, <Pg>, <Pdn>.B
    PFIRST P0, P0.B, P0.B                           // 00c05825
    PFIRST P6, P5.B, P5.B                           // c5c05825
    PFIRST P15, P15.B, P15.B                        // efc15825

// PTEST   <Pg>, <Pn>.B
    PTEST P0.B, P0                                  // 00c05025
    PTEST P6.B, P5                                  // c0d45025
    PTEST P15.B, P15                                // e0fd5025

// WRFFR   <Pn>.B
    WRFFR P0.B                                      // 00902825
    WRFFR P5.B                                      // a0902825
    WRFFR P15.B                                     // e0912825

// LASTA   <R><d>, <Pg>, <Zn>.<T>
    ZLASTA P0, Z0.B, R0                              // 00a02005
    ZLASTA P3, Z12.B, R10                            // 8aad2005
    ZLASTA P7, Z31.B, R30                            // febf2005
    ZLASTA P0, Z0.H, R0                              // 00a06005
    ZLASTA P3, Z12.H, R10                            // 8aad6005
    ZLASTA P7, Z31.H, R30                            // febf6005
    ZLASTA P0, Z0.S, R0                              // 00a0a005
    ZLASTA P3, Z12.S, R10                            // 8aada005
    ZLASTA P7, Z31.S, R30                            // febfa005
    ZLASTA P0, Z0.D, R0                              // 00a0e005
    ZLASTA P3, Z12.D, R10                            // 8aade005
    ZLASTA P7, Z31.D, R30                            // febfe005

// LASTB   <R><d>, <Pg>, <Zn>.<T>
    ZLASTB P0, Z0.B, R0                              // 00a02105
    ZLASTB P3, Z12.B, R10                            // 8aad2105
    ZLASTB P7, Z31.B, R30                            // febf2105
    ZLASTB P0, Z0.H, R0                              // 00a06105
    ZLASTB P3, Z12.H, R10                            // 8aad6105
    ZLASTB P7, Z31.H, R30                            // febf6105
    ZLASTB P0, Z0.S, R0                              // 00a0a105
    ZLASTB P3, Z12.S, R10                            // 8aada105
    ZLASTB P7, Z31.S, R30                            // febfa105
    ZLASTB P0, Z0.D, R0                              // 00a0e105
    ZLASTB P3, Z12.D, R10                            // 8aade105
    ZLASTB P7, Z31.D, R30                            // febfe105

// CLASTA  <R><dn>, <Pg>, <R><dn>, <Zm>.<T>
    ZCLASTA P0, R0, Z0.B, R0                         // 00a03005
    ZCLASTA P3, R10, Z12.B, R10                      // 8aad3005
    ZCLASTA P7, R30, Z31.B, R30                      // febf3005
    ZCLASTA P0, R0, Z0.H, R0                         // 00a07005
    ZCLASTA P3, R10, Z12.H, R10                      // 8aad7005
    ZCLASTA P7, R30, Z31.H, R30                      // febf7005
    ZCLASTA P0, R0, Z0.S, R0                         // 00a0b005
    ZCLASTA P3, R10, Z12.S, R10                      // 8aadb005
    ZCLASTA P7, R30, Z31.S, R30                      // febfb005
    ZCLASTA P0, R0, Z0.D, R0                         // 00a0f005
    ZCLASTA P3, R10, Z12.D, R10                      // 8aadf005
    ZCLASTA P7, R30, Z31.D, R30                      // febff005

// CLASTB  <R><dn>, <Pg>, <R><dn>, <Zm>.<T>
    ZCLASTB P0, R0, Z0.B, R0                         // 00a03105
    ZCLASTB P3, R10, Z12.B, R10                      // 8aad3105
    ZCLASTB P7, R30, Z31.B, R30                      // febf3105
    ZCLASTB P0, R0, Z0.H, R0                         // 00a07105
    ZCLASTB P3, R10, Z12.H, R10                      // 8aad7105
    ZCLASTB P7, R30, Z31.H, R30                      // febf7105
    ZCLASTB P0, R0, Z0.S, R0                         // 00a0b105
    ZCLASTB P3, R10, Z12.S, R10                      // 8aadb105
    ZCLASTB P7, R30, Z31.S, R30                      // febfb105
    ZCLASTB P0, R0, Z0.D, R0                         // 00a0f105
    ZCLASTB P3, R10, Z12.D, R10                      // 8aadf105
    ZCLASTB P7, R30, Z31.D, R30                      // febff105

// CTERMEQ <R><n>, <R><m>
    CTERMEQW R0, R0                                  // 0020a025
    CTERMEQW R10, R11                                // 4021ab25
    CTERMEQW R30, R30                                // c023be25
    CTERMEQ R0, R0                                   // 0020e025
    CTERMEQ R10, R11                                 // 4021eb25
    CTERMEQ R30, R30                                 // c023fe25

// CTERMNE <R><n>, <R><m>
    CTERMNEW R0, R0                                  // 1020a025
    CTERMNEW R10, R11                                // 5021ab25
    CTERMNEW R30, R30                                // d023be25
    CTERMNE R0, R0                                   // 1020e025
    CTERMNE R10, R11                                 // 5021eb25
    CTERMNE R30, R30                                 // d023fe25

// ANDV    <V><d>, <Pg>, <Zn>.<T>
    ZANDV P0, Z0.B, V0                               // 00201a04
    ZANDV P3, Z12.B, V10                             // 8a2d1a04
    ZANDV P7, Z31.B, V31                             // ff3f1a04
    ZANDV P0, Z0.H, V0                               // 00205a04
    ZANDV P3, Z12.H, V10                             // 8a2d5a04
    ZANDV P7, Z31.H, V31                             // ff3f5a04
    ZANDV P0, Z0.S, V0                               // 00209a04
    ZANDV P3, Z12.S, V10                             // 8a2d9a04
    ZANDV P7, Z31.S, V31                             // ff3f9a04
    ZANDV P0, Z0.D, V0                               // 0020da04
    ZANDV P3, Z12.D, V10                             // 8a2dda04
    ZANDV P7, Z31.D, V31                             // ff3fda04

// EORV    <V><d>, <Pg>, <Zn>.<T>
    ZEORV P0, Z0.B, V0                               // 00201904
    ZEORV P3, Z12.B, V10                             // 8a2d1904
    ZEORV P7, Z31.B, V31                             // ff3f1904
    ZEORV P0, Z0.H, V0                               // 00205904
    ZEORV P3, Z12.H, V10                             // 8a2d5904
    ZEORV P7, Z31.H, V31                             // ff3f5904
    ZEORV P0, Z0.S, V0                               // 00209904
    ZEORV P3, Z12.S, V10                             // 8a2d9904
    ZEORV P7, Z31.S, V31                             // ff3f9904
    ZEORV P0, Z0.D, V0                               // 0020d904
    ZEORV P3, Z12.D, V10                             // 8a2dd904
    ZEORV P7, Z31.D, V31                             // ff3fd904

// FADDV   <V><d>, <Pg>, <Zn>.<T>
    ZFADDV P0, Z0.H, F0                              // 00204065
    ZFADDV P3, Z12.H, F10                            // 8a2d4065
    ZFADDV P7, Z31.H, F31                            // ff3f4065
    ZFADDV P0, Z0.S, F0                              // 00208065
    ZFADDV P3, Z12.S, F10                            // 8a2d8065
    ZFADDV P7, Z31.S, F31                            // ff3f8065
    ZFADDV P0, Z0.D, F0                              // 0020c065
    ZFADDV P3, Z12.D, F10                            // 8a2dc065
    ZFADDV P7, Z31.D, F31                            // ff3fc065

// FMAXNMV <V><d>, <Pg>, <Zn>.<T>
    ZFMAXNMV P0, Z0.H, F0                            // 00204465
    ZFMAXNMV P3, Z12.H, F10                          // 8a2d4465
    ZFMAXNMV P7, Z31.H, F31                          // ff3f4465
    ZFMAXNMV P0, Z0.S, F0                            // 00208465
    ZFMAXNMV P3, Z12.S, F10                          // 8a2d8465
    ZFMAXNMV P7, Z31.S, F31                          // ff3f8465
    ZFMAXNMV P0, Z0.D, F0                            // 0020c465
    ZFMAXNMV P3, Z12.D, F10                          // 8a2dc465
    ZFMAXNMV P7, Z31.D, F31                          // ff3fc465

// FMAXV   <V><d>, <Pg>, <Zn>.<T>
    ZFMAXV P0, Z0.H, F0                              // 00204665
    ZFMAXV P3, Z12.H, F10                            // 8a2d4665
    ZFMAXV P7, Z31.H, F31                            // ff3f4665
    ZFMAXV P0, Z0.S, F0                              // 00208665
    ZFMAXV P3, Z12.S, F10                            // 8a2d8665
    ZFMAXV P7, Z31.S, F31                            // ff3f8665
    ZFMAXV P0, Z0.D, F0                              // 0020c665
    ZFMAXV P3, Z12.D, F10                            // 8a2dc665
    ZFMAXV P7, Z31.D, F31                            // ff3fc665

// FMINNMV <V><d>, <Pg>, <Zn>.<T>
    ZFMINNMV P0, Z0.H, F0                            // 00204565
    ZFMINNMV P3, Z12.H, F10                          // 8a2d4565
    ZFMINNMV P7, Z31.H, F31                          // ff3f4565
    ZFMINNMV P0, Z0.S, F0                            // 00208565
    ZFMINNMV P3, Z12.S, F10                          // 8a2d8565
    ZFMINNMV P7, Z31.S, F31                          // ff3f8565
    ZFMINNMV P0, Z0.D, F0                            // 0020c565
    ZFMINNMV P3, Z12.D, F10                          // 8a2dc565
    ZFMINNMV P7, Z31.D, F31                          // ff3fc565

// FMINV   <V><d>, <Pg>, <Zn>.<T>
    ZFMINV P0, Z0.H, F0                              // 00204765
    ZFMINV P3, Z12.H, F10                            // 8a2d4765
    ZFMINV P7, Z31.H, F31                            // ff3f4765
    ZFMINV P0, Z0.S, F0                              // 00208765
    ZFMINV P3, Z12.S, F10                            // 8a2d8765
    ZFMINV P7, Z31.S, F31                            // ff3f8765
    ZFMINV P0, Z0.D, F0                              // 0020c765
    ZFMINV P3, Z12.D, F10                            // 8a2dc765
    ZFMINV P7, Z31.D, F31                            // ff3fc765

// LASTA   <V><d>, <Pg>, <Zn>.<T>
    ZLASTA P0, Z0.B, V0                              // 00802205
    ZLASTA P3, Z12.B, V10                            // 8a8d2205
    ZLASTA P7, Z31.B, V31                            // ff9f2205
    ZLASTA P0, Z0.H, V0                              // 00806205
    ZLASTA P3, Z12.H, V10                            // 8a8d6205
    ZLASTA P7, Z31.H, V31                            // ff9f6205
    ZLASTA P0, Z0.S, V0                              // 0080a205
    ZLASTA P3, Z12.S, V10                            // 8a8da205
    ZLASTA P7, Z31.S, V31                            // ff9fa205
    ZLASTA P0, Z0.D, V0                              // 0080e205
    ZLASTA P3, Z12.D, V10                            // 8a8de205
    ZLASTA P7, Z31.D, V31                            // ff9fe205

// LASTB   <V><d>, <Pg>, <Zn>.<T>
    ZLASTB P0, Z0.B, V0                              // 00802305
    ZLASTB P3, Z12.B, V10                            // 8a8d2305
    ZLASTB P7, Z31.B, V31                            // ff9f2305
    ZLASTB P0, Z0.H, V0                              // 00806305
    ZLASTB P3, Z12.H, V10                            // 8a8d6305
    ZLASTB P7, Z31.H, V31                            // ff9f6305
    ZLASTB P0, Z0.S, V0                              // 0080a305
    ZLASTB P3, Z12.S, V10                            // 8a8da305
    ZLASTB P7, Z31.S, V31                            // ff9fa305
    ZLASTB P0, Z0.D, V0                              // 0080e305
    ZLASTB P3, Z12.D, V10                            // 8a8de305
    ZLASTB P7, Z31.D, V31                            // ff9fe305

// ORV     <V><d>, <Pg>, <Zn>.<T>
    ZORV P0, Z0.B, V0                                // 00201804
    ZORV P3, Z12.B, V10                              // 8a2d1804
    ZORV P7, Z31.B, V31                              // ff3f1804
    ZORV P0, Z0.H, V0                                // 00205804
    ZORV P3, Z12.H, V10                              // 8a2d5804
    ZORV P7, Z31.H, V31                              // ff3f5804
    ZORV P0, Z0.S, V0                                // 00209804
    ZORV P3, Z12.S, V10                              // 8a2d9804
    ZORV P7, Z31.S, V31                              // ff3f9804
    ZORV P0, Z0.D, V0                                // 0020d804
    ZORV P3, Z12.D, V10                              // 8a2dd804
    ZORV P7, Z31.D, V31                              // ff3fd804

// SMAXV   <V><d>, <Pg>, <Zn>.<T>
    ZSMAXV P0, Z0.B, V0                              // 00200804
    ZSMAXV P3, Z12.B, V10                            // 8a2d0804
    ZSMAXV P7, Z31.B, V31                            // ff3f0804
    ZSMAXV P0, Z0.H, V0                              // 00204804
    ZSMAXV P3, Z12.H, V10                            // 8a2d4804
    ZSMAXV P7, Z31.H, V31                            // ff3f4804
    ZSMAXV P0, Z0.S, V0                              // 00208804
    ZSMAXV P3, Z12.S, V10                            // 8a2d8804
    ZSMAXV P7, Z31.S, V31                            // ff3f8804
    ZSMAXV P0, Z0.D, V0                              // 0020c804
    ZSMAXV P3, Z12.D, V10                            // 8a2dc804
    ZSMAXV P7, Z31.D, V31                            // ff3fc804

// SMINV   <V><d>, <Pg>, <Zn>.<T>
    ZSMINV P0, Z0.B, V0                              // 00200a04
    ZSMINV P3, Z12.B, V10                            // 8a2d0a04
    ZSMINV P7, Z31.B, V31                            // ff3f0a04
    ZSMINV P0, Z0.H, V0                              // 00204a04
    ZSMINV P3, Z12.H, V10                            // 8a2d4a04
    ZSMINV P7, Z31.H, V31                            // ff3f4a04
    ZSMINV P0, Z0.S, V0                              // 00208a04
    ZSMINV P3, Z12.S, V10                            // 8a2d8a04
    ZSMINV P7, Z31.S, V31                            // ff3f8a04
    ZSMINV P0, Z0.D, V0                              // 0020ca04
    ZSMINV P3, Z12.D, V10                            // 8a2dca04
    ZSMINV P7, Z31.D, V31                            // ff3fca04

// UMAXV   <V><d>, <Pg>, <Zn>.<T>
    ZUMAXV P0, Z0.B, V0                              // 00200904
    ZUMAXV P3, Z12.B, V10                            // 8a2d0904
    ZUMAXV P7, Z31.B, V31                            // ff3f0904
    ZUMAXV P0, Z0.H, V0                              // 00204904
    ZUMAXV P3, Z12.H, V10                            // 8a2d4904
    ZUMAXV P7, Z31.H, V31                            // ff3f4904
    ZUMAXV P0, Z0.S, V0                              // 00208904
    ZUMAXV P3, Z12.S, V10                            // 8a2d8904
    ZUMAXV P7, Z31.S, V31                            // ff3f8904
    ZUMAXV P0, Z0.D, V0                              // 0020c904
    ZUMAXV P3, Z12.D, V10                            // 8a2dc904
    ZUMAXV P7, Z31.D, V31                            // ff3fc904

// UMINV   <V><d>, <Pg>, <Zn>.<T>
    ZUMINV P0, Z0.B, V0                              // 00200b04
    ZUMINV P3, Z12.B, V10                            // 8a2d0b04
    ZUMINV P7, Z31.B, V31                            // ff3f0b04
    ZUMINV P0, Z0.H, V0                              // 00204b04
    ZUMINV P3, Z12.H, V10                            // 8a2d4b04
    ZUMINV P7, Z31.H, V31                            // ff3f4b04
    ZUMINV P0, Z0.S, V0                              // 00208b04
    ZUMINV P3, Z12.S, V10                            // 8a2d8b04
    ZUMINV P7, Z31.S, V31                            // ff3f8b04
    ZUMINV P0, Z0.D, V0                              // 0020cb04
    ZUMINV P3, Z12.D, V10                            // 8a2dcb04
    ZUMINV P7, Z31.D, V31                            // ff3fcb04

// ADD     <Zd>.<T>, <Zn>.<T>, <Zm>.<T>
    ZADD Z0.B, Z0.B, Z0.B                            // 00002004
    ZADD Z11.B, Z12.B, Z10.B                         // 6a012c04
    ZADD Z31.B, Z31.B, Z31.B                         // ff033f04
    ZADD Z0.H, Z0.H, Z0.H                            // 00006004
    ZADD Z11.H, Z12.H, Z10.H                         // 6a016c04
    ZADD Z31.H, Z31.H, Z31.H                         // ff037f04
    ZADD Z0.S, Z0.S, Z0.S                            // 0000a004
    ZADD Z11.S, Z12.S, Z10.S                         // 6a01ac04
    ZADD Z31.S, Z31.S, Z31.S                         // ff03bf04
    ZADD Z0.D, Z0.D, Z0.D                            // 0000e004
    ZADD Z11.D, Z12.D, Z10.D                         // 6a01ec04
    ZADD Z31.D, Z31.D, Z31.D                         // ff03ff04

// FADD    <Zd>.<T>, <Zn>.<T>, <Zm>.<T>
    ZFADD Z0.H, Z0.H, Z0.H                           // 00004065
    ZFADD Z11.H, Z12.H, Z10.H                        // 6a014c65
    ZFADD Z31.H, Z31.H, Z31.H                        // ff035f65
    ZFADD Z0.S, Z0.S, Z0.S                           // 00008065
    ZFADD Z11.S, Z12.S, Z10.S                        // 6a018c65
    ZFADD Z31.S, Z31.S, Z31.S                        // ff039f65
    ZFADD Z0.D, Z0.D, Z0.D                           // 0000c065
    ZFADD Z11.D, Z12.D, Z10.D                        // 6a01cc65
    ZFADD Z31.D, Z31.D, Z31.D                        // ff03df65

// FMUL    <Zd>.<T>, <Zn>.<T>, <Zm>.<T>
    ZFMUL Z0.H, Z0.H, Z0.H                           // 00084065
    ZFMUL Z11.H, Z12.H, Z10.H                        // 6a094c65
    ZFMUL Z31.H, Z31.H, Z31.H                        // ff0b5f65
    ZFMUL Z0.S, Z0.S, Z0.S                           // 00088065
    ZFMUL Z11.S, Z12.S, Z10.S                        // 6a098c65
    ZFMUL Z31.S, Z31.S, Z31.S                        // ff0b9f65
    ZFMUL Z0.D, Z0.D, Z0.D                           // 0008c065
    ZFMUL Z11.D, Z12.D, Z10.D                        // 6a09cc65
    ZFMUL Z31.D, Z31.D, Z31.D                        // ff0bdf65

// FRECPS  <Zd>.<T>, <Zn>.<T>, <Zm>.<T>
    ZFRECPS Z0.H, Z0.H, Z0.H                         // 00184065
    ZFRECPS Z11.H, Z12.H, Z10.H                      // 6a194c65
    ZFRECPS Z31.H, Z31.H, Z31.H                      // ff1b5f65
    ZFRECPS Z0.S, Z0.S, Z0.S                         // 00188065
    ZFRECPS Z11.S, Z12.S, Z10.S                      // 6a198c65
    ZFRECPS Z31.S, Z31.S, Z31.S                      // ff1b9f65
    ZFRECPS Z0.D, Z0.D, Z0.D                         // 0018c065
    ZFRECPS Z11.D, Z12.D, Z10.D                      // 6a19cc65
    ZFRECPS Z31.D, Z31.D, Z31.D                      // ff1bdf65

// FRSQRTS <Zd>.<T>, <Zn>.<T>, <Zm>.<T>
    ZFRSQRTS Z0.H, Z0.H, Z0.H                        // 001c4065
    ZFRSQRTS Z11.H, Z12.H, Z10.H                     // 6a1d4c65
    ZFRSQRTS Z31.H, Z31.H, Z31.H                     // ff1f5f65
    ZFRSQRTS Z0.S, Z0.S, Z0.S                        // 001c8065
    ZFRSQRTS Z11.S, Z12.S, Z10.S                     // 6a1d8c65
    ZFRSQRTS Z31.S, Z31.S, Z31.S                     // ff1f9f65
    ZFRSQRTS Z0.D, Z0.D, Z0.D                        // 001cc065
    ZFRSQRTS Z11.D, Z12.D, Z10.D                     // 6a1dcc65
    ZFRSQRTS Z31.D, Z31.D, Z31.D                     // ff1fdf65

// FSUB    <Zd>.<T>, <Zn>.<T>, <Zm>.<T>
    ZFSUB Z0.H, Z0.H, Z0.H                           // 00044065
    ZFSUB Z11.H, Z12.H, Z10.H                        // 6a054c65
    ZFSUB Z31.H, Z31.H, Z31.H                        // ff075f65
    ZFSUB Z0.S, Z0.S, Z0.S                           // 00048065
    ZFSUB Z11.S, Z12.S, Z10.S                        // 6a058c65
    ZFSUB Z31.S, Z31.S, Z31.S                        // ff079f65
    ZFSUB Z0.D, Z0.D, Z0.D                           // 0004c065
    ZFSUB Z11.D, Z12.D, Z10.D                        // 6a05cc65
    ZFSUB Z31.D, Z31.D, Z31.D                        // ff07df65

// FTSMUL  <Zd>.<T>, <Zn>.<T>, <Zm>.<T>
    ZFTSMUL Z0.H, Z0.H, Z0.H                         // 000c4065
    ZFTSMUL Z11.H, Z12.H, Z10.H                      // 6a0d4c65
    ZFTSMUL Z31.H, Z31.H, Z31.H                      // ff0f5f65
    ZFTSMUL Z0.S, Z0.S, Z0.S                         // 000c8065
    ZFTSMUL Z11.S, Z12.S, Z10.S                      // 6a0d8c65
    ZFTSMUL Z31.S, Z31.S, Z31.S                      // ff0f9f65
    ZFTSMUL Z0.D, Z0.D, Z0.D                         // 000cc065
    ZFTSMUL Z11.D, Z12.D, Z10.D                      // 6a0dcc65
    ZFTSMUL Z31.D, Z31.D, Z31.D                      // ff0fdf65

// FTSSEL  <Zd>.<T>, <Zn>.<T>, <Zm>.<T>
    ZFTSSEL Z0.H, Z0.H, Z0.H                         // 00b06004
    ZFTSSEL Z11.H, Z12.H, Z10.H                      // 6ab16c04
    ZFTSSEL Z31.H, Z31.H, Z31.H                      // ffb37f04
    ZFTSSEL Z0.S, Z0.S, Z0.S                         // 00b0a004
    ZFTSSEL Z11.S, Z12.S, Z10.S                      // 6ab1ac04
    ZFTSSEL Z31.S, Z31.S, Z31.S                      // ffb3bf04
    ZFTSSEL Z0.D, Z0.D, Z0.D                         // 00b0e004
    ZFTSSEL Z11.D, Z12.D, Z10.D                      // 6ab1ec04
    ZFTSSEL Z31.D, Z31.D, Z31.D                      // ffb3ff04

// SQADD   <Zd>.<T>, <Zn>.<T>, <Zm>.<T>
    ZSQADD Z0.B, Z0.B, Z0.B                          // 00102004
    ZSQADD Z11.B, Z12.B, Z10.B                       // 6a112c04
    ZSQADD Z31.B, Z31.B, Z31.B                       // ff133f04
    ZSQADD Z0.H, Z0.H, Z0.H                          // 00106004
    ZSQADD Z11.H, Z12.H, Z10.H                       // 6a116c04
    ZSQADD Z31.H, Z31.H, Z31.H                       // ff137f04
    ZSQADD Z0.S, Z0.S, Z0.S                          // 0010a004
    ZSQADD Z11.S, Z12.S, Z10.S                       // 6a11ac04
    ZSQADD Z31.S, Z31.S, Z31.S                       // ff13bf04
    ZSQADD Z0.D, Z0.D, Z0.D                          // 0010e004
    ZSQADD Z11.D, Z12.D, Z10.D                       // 6a11ec04
    ZSQADD Z31.D, Z31.D, Z31.D                       // ff13ff04

// SQSUB   <Zd>.<T>, <Zn>.<T>, <Zm>.<T>
    ZSQSUB Z0.B, Z0.B, Z0.B                          // 00182004
    ZSQSUB Z11.B, Z12.B, Z10.B                       // 6a192c04
    ZSQSUB Z31.B, Z31.B, Z31.B                       // ff1b3f04
    ZSQSUB Z0.H, Z0.H, Z0.H                          // 00186004
    ZSQSUB Z11.H, Z12.H, Z10.H                       // 6a196c04
    ZSQSUB Z31.H, Z31.H, Z31.H                       // ff1b7f04
    ZSQSUB Z0.S, Z0.S, Z0.S                          // 0018a004
    ZSQSUB Z11.S, Z12.S, Z10.S                       // 6a19ac04
    ZSQSUB Z31.S, Z31.S, Z31.S                       // ff1bbf04
    ZSQSUB Z0.D, Z0.D, Z0.D                          // 0018e004
    ZSQSUB Z11.D, Z12.D, Z10.D                       // 6a19ec04
    ZSQSUB Z31.D, Z31.D, Z31.D                       // ff1bff04

// SUB     <Zd>.<T>, <Zn>.<T>, <Zm>.<T>
    ZSUB Z0.B, Z0.B, Z0.B                            // 00042004
    ZSUB Z11.B, Z12.B, Z10.B                         // 6a052c04
    ZSUB Z31.B, Z31.B, Z31.B                         // ff073f04
    ZSUB Z0.H, Z0.H, Z0.H                            // 00046004
    ZSUB Z11.H, Z12.H, Z10.H                         // 6a056c04
    ZSUB Z31.H, Z31.H, Z31.H                         // ff077f04
    ZSUB Z0.S, Z0.S, Z0.S                            // 0004a004
    ZSUB Z11.S, Z12.S, Z10.S                         // 6a05ac04
    ZSUB Z31.S, Z31.S, Z31.S                         // ff07bf04
    ZSUB Z0.D, Z0.D, Z0.D                            // 0004e004
    ZSUB Z11.D, Z12.D, Z10.D                         // 6a05ec04
    ZSUB Z31.D, Z31.D, Z31.D                         // ff07ff04

// TRN1    <Zd>.<T>, <Zn>.<T>, <Zm>.<T>
    ZTRN1 Z0.B, Z0.B, Z0.B                           // 00702005
    ZTRN1 Z11.B, Z12.B, Z10.B                        // 6a712c05
    ZTRN1 Z31.B, Z31.B, Z31.B                        // ff733f05
    ZTRN1 Z0.H, Z0.H, Z0.H                           // 00706005
    ZTRN1 Z11.H, Z12.H, Z10.H                        // 6a716c05
    ZTRN1 Z31.H, Z31.H, Z31.H                        // ff737f05
    ZTRN1 Z0.S, Z0.S, Z0.S                           // 0070a005
    ZTRN1 Z11.S, Z12.S, Z10.S                        // 6a71ac05
    ZTRN1 Z31.S, Z31.S, Z31.S                        // ff73bf05
    ZTRN1 Z0.D, Z0.D, Z0.D                           // 0070e005
    ZTRN1 Z11.D, Z12.D, Z10.D                        // 6a71ec05
    ZTRN1 Z31.D, Z31.D, Z31.D                        // ff73ff05

// TRN2    <Zd>.<T>, <Zn>.<T>, <Zm>.<T>
    ZTRN2 Z0.B, Z0.B, Z0.B                           // 00742005
    ZTRN2 Z11.B, Z12.B, Z10.B                        // 6a752c05
    ZTRN2 Z31.B, Z31.B, Z31.B                        // ff773f05
    ZTRN2 Z0.H, Z0.H, Z0.H                           // 00746005
    ZTRN2 Z11.H, Z12.H, Z10.H                        // 6a756c05
    ZTRN2 Z31.H, Z31.H, Z31.H                        // ff777f05
    ZTRN2 Z0.S, Z0.S, Z0.S                           // 0074a005
    ZTRN2 Z11.S, Z12.S, Z10.S                        // 6a75ac05
    ZTRN2 Z31.S, Z31.S, Z31.S                        // ff77bf05
    ZTRN2 Z0.D, Z0.D, Z0.D                           // 0074e005
    ZTRN2 Z11.D, Z12.D, Z10.D                        // 6a75ec05
    ZTRN2 Z31.D, Z31.D, Z31.D                        // ff77ff05

// UQADD   <Zd>.<T>, <Zn>.<T>, <Zm>.<T>
    ZUQADD Z0.B, Z0.B, Z0.B                          // 00142004
    ZUQADD Z11.B, Z12.B, Z10.B                       // 6a152c04
    ZUQADD Z31.B, Z31.B, Z31.B                       // ff173f04
    ZUQADD Z0.H, Z0.H, Z0.H                          // 00146004
    ZUQADD Z11.H, Z12.H, Z10.H                       // 6a156c04
    ZUQADD Z31.H, Z31.H, Z31.H                       // ff177f04
    ZUQADD Z0.S, Z0.S, Z0.S                          // 0014a004
    ZUQADD Z11.S, Z12.S, Z10.S                       // 6a15ac04
    ZUQADD Z31.S, Z31.S, Z31.S                       // ff17bf04
    ZUQADD Z0.D, Z0.D, Z0.D                          // 0014e004
    ZUQADD Z11.D, Z12.D, Z10.D                       // 6a15ec04
    ZUQADD Z31.D, Z31.D, Z31.D                       // ff17ff04

// UQSUB   <Zd>.<T>, <Zn>.<T>, <Zm>.<T>
    ZUQSUB Z0.B, Z0.B, Z0.B                          // 001c2004
    ZUQSUB Z11.B, Z12.B, Z10.B                       // 6a1d2c04
    ZUQSUB Z31.B, Z31.B, Z31.B                       // ff1f3f04
    ZUQSUB Z0.H, Z0.H, Z0.H                          // 001c6004
    ZUQSUB Z11.H, Z12.H, Z10.H                       // 6a1d6c04
    ZUQSUB Z31.H, Z31.H, Z31.H                       // ff1f7f04
    ZUQSUB Z0.S, Z0.S, Z0.S                          // 001ca004
    ZUQSUB Z11.S, Z12.S, Z10.S                       // 6a1dac04
    ZUQSUB Z31.S, Z31.S, Z31.S                       // ff1fbf04
    ZUQSUB Z0.D, Z0.D, Z0.D                          // 001ce004
    ZUQSUB Z11.D, Z12.D, Z10.D                       // 6a1dec04
    ZUQSUB Z31.D, Z31.D, Z31.D                       // ff1fff04

// UZP1    <Zd>.<T>, <Zn>.<T>, <Zm>.<T>
    ZUZP1 Z0.B, Z0.B, Z0.B                           // 00682005
    ZUZP1 Z11.B, Z12.B, Z10.B                        // 6a692c05
    ZUZP1 Z31.B, Z31.B, Z31.B                        // ff6b3f05
    ZUZP1 Z0.H, Z0.H, Z0.H                           // 00686005
    ZUZP1 Z11.H, Z12.H, Z10.H                        // 6a696c05
    ZUZP1 Z31.H, Z31.H, Z31.H                        // ff6b7f05
    ZUZP1 Z0.S, Z0.S, Z0.S                           // 0068a005
    ZUZP1 Z11.S, Z12.S, Z10.S                        // 6a69ac05
    ZUZP1 Z31.S, Z31.S, Z31.S                        // ff6bbf05
    ZUZP1 Z0.D, Z0.D, Z0.D                           // 0068e005
    ZUZP1 Z11.D, Z12.D, Z10.D                        // 6a69ec05
    ZUZP1 Z31.D, Z31.D, Z31.D                        // ff6bff05

// UZP2    <Zd>.<T>, <Zn>.<T>, <Zm>.<T>
    ZUZP2 Z0.B, Z0.B, Z0.B                           // 006c2005
    ZUZP2 Z11.B, Z12.B, Z10.B                        // 6a6d2c05
    ZUZP2 Z31.B, Z31.B, Z31.B                        // ff6f3f05
    ZUZP2 Z0.H, Z0.H, Z0.H                           // 006c6005
    ZUZP2 Z11.H, Z12.H, Z10.H                        // 6a6d6c05
    ZUZP2 Z31.H, Z31.H, Z31.H                        // ff6f7f05
    ZUZP2 Z0.S, Z0.S, Z0.S                           // 006ca005
    ZUZP2 Z11.S, Z12.S, Z10.S                        // 6a6dac05
    ZUZP2 Z31.S, Z31.S, Z31.S                        // ff6fbf05
    ZUZP2 Z0.D, Z0.D, Z0.D                           // 006ce005
    ZUZP2 Z11.D, Z12.D, Z10.D                        // 6a6dec05
    ZUZP2 Z31.D, Z31.D, Z31.D                        // ff6fff05

// ZIP1    <Zd>.<T>, <Zn>.<T>, <Zm>.<T>
    ZZIP1 Z0.B, Z0.B, Z0.B                           // 00602005
    ZZIP1 Z11.B, Z12.B, Z10.B                        // 6a612c05
    ZZIP1 Z31.B, Z31.B, Z31.B                        // ff633f05
    ZZIP1 Z0.H, Z0.H, Z0.H                           // 00606005
    ZZIP1 Z11.H, Z12.H, Z10.H                        // 6a616c05
    ZZIP1 Z31.H, Z31.H, Z31.H                        // ff637f05
    ZZIP1 Z0.S, Z0.S, Z0.S                           // 0060a005
    ZZIP1 Z11.S, Z12.S, Z10.S                        // 6a61ac05
    ZZIP1 Z31.S, Z31.S, Z31.S                        // ff63bf05
    ZZIP1 Z0.D, Z0.D, Z0.D                           // 0060e005
    ZZIP1 Z11.D, Z12.D, Z10.D                        // 6a61ec05
    ZZIP1 Z31.D, Z31.D, Z31.D                        // ff63ff05

// ZIP2    <Zd>.<T>, <Zn>.<T>, <Zm>.<T>
    ZZIP2 Z0.B, Z0.B, Z0.B                           // 00642005
    ZZIP2 Z11.B, Z12.B, Z10.B                        // 6a652c05
    ZZIP2 Z31.B, Z31.B, Z31.B                        // ff673f05
    ZZIP2 Z0.H, Z0.H, Z0.H                           // 00646005
    ZZIP2 Z11.H, Z12.H, Z10.H                        // 6a656c05
    ZZIP2 Z31.H, Z31.H, Z31.H                        // ff677f05
    ZZIP2 Z0.S, Z0.S, Z0.S                           // 0064a005
    ZZIP2 Z11.S, Z12.S, Z10.S                        // 6a65ac05
    ZZIP2 Z31.S, Z31.S, Z31.S                        // ff67bf05
    ZZIP2 Z0.D, Z0.D, Z0.D                           // 0064e005
    ZZIP2 Z11.D, Z12.D, Z10.D                        // 6a65ec05
    ZZIP2 Z31.D, Z31.D, Z31.D                        // ff67ff05

// CLASTA  <V><dn>, <Pg>, <V><dn>, <Zm>.<T>
    ZCLASTA P0, V0, Z0.B, V0                         // 00802a05
    ZCLASTA P3, V10, Z12.B, V10                      // 8a8d2a05
    ZCLASTA P7, V31, Z31.B, V31                      // ff9f2a05
    ZCLASTA P0, V0, Z0.H, V0                         // 00806a05
    ZCLASTA P3, V10, Z12.H, V10                      // 8a8d6a05
    ZCLASTA P7, V31, Z31.H, V31                      // ff9f6a05
    ZCLASTA P0, V0, Z0.S, V0                         // 0080aa05
    ZCLASTA P3, V10, Z12.S, V10                      // 8a8daa05
    ZCLASTA P7, V31, Z31.S, V31                      // ff9faa05
    ZCLASTA P0, V0, Z0.D, V0                         // 0080ea05
    ZCLASTA P3, V10, Z12.D, V10                      // 8a8dea05
    ZCLASTA P7, V31, Z31.D, V31                      // ff9fea05

// CLASTB  <V><dn>, <Pg>, <V><dn>, <Zm>.<T>
    ZCLASTB P0, V0, Z0.B, V0                         // 00802b05
    ZCLASTB P3, V10, Z12.B, V10                      // 8a8d2b05
    ZCLASTB P7, V31, Z31.B, V31                      // ff9f2b05
    ZCLASTB P0, V0, Z0.H, V0                         // 00806b05
    ZCLASTB P3, V10, Z12.H, V10                      // 8a8d6b05
    ZCLASTB P7, V31, Z31.H, V31                      // ff9f6b05
    ZCLASTB P0, V0, Z0.S, V0                         // 0080ab05
    ZCLASTB P3, V10, Z12.S, V10                      // 8a8dab05
    ZCLASTB P7, V31, Z31.S, V31                      // ff9fab05
    ZCLASTB P0, V0, Z0.D, V0                         // 0080eb05
    ZCLASTB P3, V10, Z12.D, V10                      // 8a8deb05
    ZCLASTB P7, V31, Z31.D, V31                      // ff9feb05

// FADDA   <V><dn>, <Pg>, <V><dn>, <Zm>.<T>
    ZFADDA P0, F0, Z0.H, F0                          // 00205865
    ZFADDA P3, F10, Z12.H, F10                       // 8a2d5865
    ZFADDA P7, F31, Z31.H, F31                       // ff3f5865
    ZFADDA P0, F0, Z0.S, F0                          // 00209865
    ZFADDA P3, F10, Z12.S, F10                       // 8a2d9865
    ZFADDA P7, F31, Z31.S, F31                       // ff3f9865
    ZFADDA P0, F0, Z0.D, F0                          // 0020d865
    ZFADDA P3, F10, Z12.D, F10                       // 8a2dd865
    ZFADDA P7, F31, Z31.D, F31                       // ff3fd865

// AND     <Zd>.D, <Zn>.D, <Zm>.D
    ZAND Z0.D, Z0.D, Z0.D                            // 00302004
    ZAND Z11.D, Z12.D, Z10.D                         // 6a312c04
    ZAND Z31.D, Z31.D, Z31.D                         // ff333f04

// BIC     <Zd>.D, <Zn>.D, <Zm>.D
    ZBIC Z0.D, Z0.D, Z0.D                            // 0030e004
    ZBIC Z11.D, Z12.D, Z10.D                         // 6a31ec04
    ZBIC Z31.D, Z31.D, Z31.D                         // ff33ff04

// EOR     <Zd>.D, <Zn>.D, <Zm>.D
    ZEOR Z0.D, Z0.D, Z0.D                            // 0030a004
    ZEOR Z11.D, Z12.D, Z10.D                         // 6a31ac04
    ZEOR Z31.D, Z31.D, Z31.D                         // ff33bf04

// ORR     <Zd>.D, <Zn>.D, <Zm>.D
    ZORR Z0.D, Z0.D, Z0.D                            // 00306004
    ZORR Z11.D, Z12.D, Z10.D                         // 6a316c04
    ZORR Z31.D, Z31.D, Z31.D                         // ff337f04

// FCVTZS  <Zd>.D, <Pg>/M, <Zn>.D
    ZFCVTZS P0.M, Z0.D, Z0.D                         // 00a0de65
    ZFCVTZS P3.M, Z12.D, Z10.D                       // 8aadde65
    ZFCVTZS P7.M, Z31.D, Z31.D                       // ffbfde65

// FCVTZU  <Zd>.D, <Pg>/M, <Zn>.D
    ZFCVTZU P0.M, Z0.D, Z0.D                         // 00a0df65
    ZFCVTZU P3.M, Z12.D, Z10.D                       // 8aaddf65
    ZFCVTZU P7.M, Z31.D, Z31.D                       // ffbfdf65

// REVW    <Zd>.D, <Pg>/M, <Zn>.D
    ZREVW P0.M, Z0.D, Z0.D                           // 0080e605
    ZREVW P3.M, Z12.D, Z10.D                         // 8a8de605
    ZREVW P7.M, Z31.D, Z31.D                         // ff9fe605

// SCVTF   <Zd>.D, <Pg>/M, <Zn>.D
    ZSCVTF P0.M, Z0.D, Z0.D                          // 00a0d665
    ZSCVTF P3.M, Z12.D, Z10.D                        // 8aadd665
    ZSCVTF P7.M, Z31.D, Z31.D                        // ffbfd665

// SXTW    <Zd>.D, <Pg>/M, <Zn>.D
    ZSXTW P0.M, Z0.D, Z0.D                           // 00a0d404
    ZSXTW P3.M, Z12.D, Z10.D                         // 8aadd404
    ZSXTW P7.M, Z31.D, Z31.D                         // ffbfd404

// UCVTF   <Zd>.D, <Pg>/M, <Zn>.D
    ZUCVTF P0.M, Z0.D, Z0.D                          // 00a0d765
    ZUCVTF P3.M, Z12.D, Z10.D                        // 8aadd765
    ZUCVTF P7.M, Z31.D, Z31.D                        // ffbfd765

// UXTW    <Zd>.D, <Pg>/M, <Zn>.D
    ZUXTW P0.M, Z0.D, Z0.D                           // 00a0d504
    ZUXTW P3.M, Z12.D, Z10.D                         // 8aadd504
    ZUXTW P7.M, Z31.D, Z31.D                         // ffbfd504

// FCVT    <Zd>.D, <Pg>/M, <Zn>.H
    ZFCVT P0.M, Z0.H, Z0.D                           // 00a0c965
    ZFCVT P3.M, Z12.H, Z10.D                         // 8aadc965
    ZFCVT P7.M, Z31.H, Z31.D                         // ffbfc965

// FCVTZS  <Zd>.D, <Pg>/M, <Zn>.H
    ZFCVTZS P0.M, Z0.H, Z0.D                         // 00a05e65
    ZFCVTZS P3.M, Z12.H, Z10.D                       // 8aad5e65
    ZFCVTZS P7.M, Z31.H, Z31.D                       // ffbf5e65

// FCVTZU  <Zd>.D, <Pg>/M, <Zn>.H
    ZFCVTZU P0.M, Z0.H, Z0.D                         // 00a05f65
    ZFCVTZU P3.M, Z12.H, Z10.D                       // 8aad5f65
    ZFCVTZU P7.M, Z31.H, Z31.D                       // ffbf5f65

// FCVT    <Zd>.D, <Pg>/M, <Zn>.S
    ZFCVT P0.M, Z0.S, Z0.D                           // 00a0cb65
    ZFCVT P3.M, Z12.S, Z10.D                         // 8aadcb65
    ZFCVT P7.M, Z31.S, Z31.D                         // ffbfcb65

// FCVTZS  <Zd>.D, <Pg>/M, <Zn>.S
    ZFCVTZS P0.M, Z0.S, Z0.D                         // 00a0dc65
    ZFCVTZS P3.M, Z12.S, Z10.D                       // 8aaddc65
    ZFCVTZS P7.M, Z31.S, Z31.D                       // ffbfdc65

// FCVTZU  <Zd>.D, <Pg>/M, <Zn>.S
    ZFCVTZU P0.M, Z0.S, Z0.D                         // 00a0dd65
    ZFCVTZU P3.M, Z12.S, Z10.D                       // 8aaddd65
    ZFCVTZU P7.M, Z31.S, Z31.D                       // ffbfdd65

// SCVTF   <Zd>.D, <Pg>/M, <Zn>.S
    ZSCVTF P0.M, Z0.S, Z0.D                          // 00a0d065
    ZSCVTF P3.M, Z12.S, Z10.D                        // 8aadd065
    ZSCVTF P7.M, Z31.S, Z31.D                        // ffbfd065

// UCVTF   <Zd>.D, <Pg>/M, <Zn>.S
    ZUCVTF P0.M, Z0.S, Z0.D                          // 00a0d165
    ZUCVTF P3.M, Z12.S, Z10.D                        // 8aadd165
    ZUCVTF P7.M, Z31.S, Z31.D                        // ffbfd165

// FCVTZS  <Zd>.H, <Pg>/M, <Zn>.H
    ZFCVTZS P0.M, Z0.H, Z0.H                         // 00a05a65
    ZFCVTZS P3.M, Z12.H, Z10.H                       // 8aad5a65
    ZFCVTZS P7.M, Z31.H, Z31.H                       // ffbf5a65

// FCVTZU  <Zd>.H, <Pg>/M, <Zn>.H
    ZFCVTZU P0.M, Z0.H, Z0.H                         // 00a05b65
    ZFCVTZU P3.M, Z12.H, Z10.H                       // 8aad5b65
    ZFCVTZU P7.M, Z31.H, Z31.H                       // ffbf5b65

// SCVTF   <Zd>.H, <Pg>/M, <Zn>.H
    ZSCVTF P0.M, Z0.H, Z0.H                          // 00a05265
    ZSCVTF P3.M, Z12.H, Z10.H                        // 8aad5265
    ZSCVTF P7.M, Z31.H, Z31.H                        // ffbf5265

// UCVTF   <Zd>.H, <Pg>/M, <Zn>.H
    ZUCVTF P0.M, Z0.H, Z0.H                          // 00a05365
    ZUCVTF P3.M, Z12.H, Z10.H                        // 8aad5365
    ZUCVTF P7.M, Z31.H, Z31.H                        // ffbf5365

// FCVT    <Zd>.H, <Pg>/M, <Zn>.D
    ZFCVT P0.M, Z0.D, Z0.H                           // 00a0c865
    ZFCVT P3.M, Z12.D, Z10.H                         // 8aadc865
    ZFCVT P7.M, Z31.D, Z31.H                         // ffbfc865

// SCVTF   <Zd>.H, <Pg>/M, <Zn>.D
    ZSCVTF P0.M, Z0.D, Z0.H                          // 00a05665
    ZSCVTF P3.M, Z12.D, Z10.H                        // 8aad5665
    ZSCVTF P7.M, Z31.D, Z31.H                        // ffbf5665

// UCVTF   <Zd>.H, <Pg>/M, <Zn>.D
    ZUCVTF P0.M, Z0.D, Z0.H                          // 00a05765
    ZUCVTF P3.M, Z12.D, Z10.H                        // 8aad5765
    ZUCVTF P7.M, Z31.D, Z31.H                        // ffbf5765

// FCVT    <Zd>.S, <Pg>/M, <Zn>.D
    ZFCVT P0.M, Z0.D, Z0.S                           // 00a0ca65
    ZFCVT P3.M, Z12.D, Z10.S                         // 8aadca65
    ZFCVT P7.M, Z31.D, Z31.S                         // ffbfca65

// FCVTZS  <Zd>.S, <Pg>/M, <Zn>.D
    ZFCVTZS P0.M, Z0.D, Z0.S                         // 00a0d865
    ZFCVTZS P3.M, Z12.D, Z10.S                       // 8aadd865
    ZFCVTZS P7.M, Z31.D, Z31.S                       // ffbfd865

// FCVTZU  <Zd>.S, <Pg>/M, <Zn>.D
    ZFCVTZU P0.M, Z0.D, Z0.S                         // 00a0d965
    ZFCVTZU P3.M, Z12.D, Z10.S                       // 8aadd965
    ZFCVTZU P7.M, Z31.D, Z31.S                       // ffbfd965

// SCVTF   <Zd>.S, <Pg>/M, <Zn>.D
    ZSCVTF P0.M, Z0.D, Z0.S                          // 00a0d465
    ZSCVTF P3.M, Z12.D, Z10.S                        // 8aadd465
    ZSCVTF P7.M, Z31.D, Z31.S                        // ffbfd465

// UCVTF   <Zd>.S, <Pg>/M, <Zn>.D
    ZUCVTF P0.M, Z0.D, Z0.S                          // 00a0d565
    ZUCVTF P3.M, Z12.D, Z10.S                        // 8aadd565
    ZUCVTF P7.M, Z31.D, Z31.S                        // ffbfd565

// FCVT    <Zd>.S, <Pg>/M, <Zn>.H
    ZFCVT P0.M, Z0.H, Z0.S                           // 00a08965
    ZFCVT P3.M, Z12.H, Z10.S                         // 8aad8965
    ZFCVT P7.M, Z31.H, Z31.S                         // ffbf8965

// FCVTZS  <Zd>.S, <Pg>/M, <Zn>.H
    ZFCVTZS P0.M, Z0.H, Z0.S                         // 00a05c65
    ZFCVTZS P3.M, Z12.H, Z10.S                       // 8aad5c65
    ZFCVTZS P7.M, Z31.H, Z31.S                       // ffbf5c65

// FCVTZU  <Zd>.S, <Pg>/M, <Zn>.H
    ZFCVTZU P0.M, Z0.H, Z0.S                         // 00a05d65
    ZFCVTZU P3.M, Z12.H, Z10.S                       // 8aad5d65
    ZFCVTZU P7.M, Z31.H, Z31.S                       // ffbf5d65

// FCVTZS  <Zd>.S, <Pg>/M, <Zn>.S
    ZFCVTZS P0.M, Z0.S, Z0.S                         // 00a09c65
    ZFCVTZS P3.M, Z12.S, Z10.S                       // 8aad9c65
    ZFCVTZS P7.M, Z31.S, Z31.S                       // ffbf9c65

// FCVTZU  <Zd>.S, <Pg>/M, <Zn>.S
    ZFCVTZU P0.M, Z0.S, Z0.S                         // 00a09d65
    ZFCVTZU P3.M, Z12.S, Z10.S                       // 8aad9d65
    ZFCVTZU P7.M, Z31.S, Z31.S                       // ffbf9d65

// SCVTF   <Zd>.S, <Pg>/M, <Zn>.S
    ZSCVTF P0.M, Z0.S, Z0.S                          // 00a09465
    ZSCVTF P3.M, Z12.S, Z10.S                        // 8aad9465
    ZSCVTF P7.M, Z31.S, Z31.S                        // ffbf9465

// UCVTF   <Zd>.S, <Pg>/M, <Zn>.S
    ZUCVTF P0.M, Z0.S, Z0.S                          // 00a09565
    ZUCVTF P3.M, Z12.S, Z10.S                        // 8aad9565
    ZUCVTF P7.M, Z31.S, Z31.S                        // ffbf9565

// BFCVT   <Zd>.H, <Pg>/M, <Zn>.S
    ZBFCVT P0.M, Z0.S, Z0.H                          // 00a08a65
    ZBFCVT P3.M, Z12.S, Z10.H                        // 8aad8a65
    ZBFCVT P7.M, Z31.S, Z31.H                        // ffbf8a65

// BFCVTNT <Zd>.H, <Pg>/M, <Zn>.S
    ZBFCVTNT P0.M, Z0.S, Z0.H                        // 00a08a64
    ZBFCVTNT P3.M, Z12.S, Z10.H                      // 8aad8a64
    ZBFCVTNT P7.M, Z31.S, Z31.H                      // ffbf8a64

// FCVT    <Zd>.H, <Pg>/M, <Zn>.S
    ZFCVT P0.M, Z0.S, Z0.H                           // 00a08865
    ZFCVT P3.M, Z12.S, Z10.H                         // 8aad8865
    ZFCVT P7.M, Z31.S, Z31.H                         // ffbf8865

// SCVTF   <Zd>.H, <Pg>/M, <Zn>.S
    ZSCVTF P0.M, Z0.S, Z0.H                          // 00a05465
    ZSCVTF P3.M, Z12.S, Z10.H                        // 8aad5465
    ZSCVTF P7.M, Z31.S, Z31.H                        // ffbf5465

// UCVTF   <Zd>.H, <Pg>/M, <Zn>.S
    ZUCVTF P0.M, Z0.S, Z0.H                          // 00a05565
    ZUCVTF P3.M, Z12.S, Z10.H                        // 8aad5565
    ZUCVTF P7.M, Z31.S, Z31.H                        // ffbf5565

    RET
