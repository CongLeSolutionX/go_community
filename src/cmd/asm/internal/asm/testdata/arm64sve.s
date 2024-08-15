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

    RET
