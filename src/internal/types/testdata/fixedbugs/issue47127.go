// Copyright 2021 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Embedding of stand-alone type parameters is not permitted.

package p

type (
        _[P any] interface{ *P | []P | chan P | map[string]P }
        _[P any] interface{ P /* ERR term cannot be a type parameter */ }
        _[P any] interface{ ~P /* ERR type in term ~P cannot be a type parameter */ }
        _[P any] interface{ int | P /* ERR term cannot be a type parameter */ }
        _[P any] interface{ int | ~P /* ERR type in term ~P cannot be a type parameter */ }
)

func _[P any]() {
        type (
                _[P any] interface{ *P | []P | chan P | map[string]P }
                _[P any] interface{ P /* ERR term cannot be a type parameter */ }
                _[P any] interface{ ~P /* ERR type in term ~P cannot be a type parameter */ }
                _[P any] interface{ int | P /* ERR term cannot be a type parameter */ }
                _[P any] interface{ int | ~P /* ERR type in term ~P cannot be a type parameter */ }

                _ interface{ *P | []P | chan P | map[string]P }
                _ interface{ P /* ERR term cannot be a type parameter */ }
                _ interface{ ~P /* ERR type in term ~P cannot be a type parameter */ }
                _ interface{ int | P /* ERR term cannot be a type parameter */ }
                _ interface{ int | ~P /* ERR type in term ~P cannot be a type parameter */ }
        )
}

func _[P any, Q interface{ *P | []P | chan P | map[string]P }]() {}
func _[P any, Q interface{ P /* ERR term cannot be a type parameter */ }]() {}
func _[P any, Q interface{ ~P /* ERR type in term ~P cannot be a type parameter */ }]() {}
func _[P any, Q interface{ int | P /* ERR term cannot be a type parameter */ }]() {}
func _[P any, Q interface{ int | ~P /* ERR type in term ~P cannot be a type parameter */ }]() {}
