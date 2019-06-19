// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package internal

import "strings"

// LocalhostCert is a PEM-encoded TLS cert with SAN IPs
// "127.0.0.1" and "[::1]", expiring at Jan 29 16:00:00 2084 GMT.
// generated from src/crypto/tls:
// go run generate_cert.go --rsa-bits 1024 --host "127.0.0.1,::1,example.com,*.test.example" --ca --start-date "Jan 1 00:00:00 1970" --duration=1000000h
var LocalhostCert = []byte(`-----BEGIN CERTIFICATE-----
MIICIzCCAYygAwIBAgIQFd0/NIjcQteXI/5IUTAkdzANBgkqhkiG9w0BAQsFADAS
MRAwDgYDVQQKEwdBY21lIENvMCAXDTcwMDEwMTAwMDAwMFoYDzIwODQwMTI5MTYw
MDAwWjASMRAwDgYDVQQKEwdBY21lIENvMIGfMA0GCSqGSIb3DQEBAQUAA4GNADCB
iQKBgQCxECpq9t32ivJJeQK+WJQUYuClYIWOyWFZICIdp4bzsWuR+97zWQTOmvla
ogiO4Zx9of+FJADg+zXuX5ZoPqF6IOBdA6Yu1iSKcYTBhe7BzoCSxAzwOMatTwEU
X/XRtAyeMtlZ880SNL7/FKzB6Bj2Tdd3GL5+p7aEAkPVoPK+OQIDAQABo3gwdjAO
BgNVHQ8BAf8EBAMCAqQwEwYDVR0lBAwwCgYIKwYBBQUHAwEwDwYDVR0TAQH/BAUw
AwEB/zA+BgNVHREENzA1ggtleGFtcGxlLmNvbYIOKi50ZXN0LmV4YW1wbGWHBH8A
AAGHEAAAAAAAAAAAAAAAAAAAAAEwDQYJKoZIhvcNAQELBQADgYEAPOKyZaR/5y8N
XDbqnqTChTzjuf5XhECFZJqc+Qxwmhv0cw6rPg3/yTH/XJQHMB9NjUx2OiaXeLAp
HlRCz7mfAeYFvhqa2E7sTZhlsxIrlnoJfRzX/e8Y+gB4usB1I9s8onVDo93ASJAa
8m0IAUV6khBDEBHOeOmoMZZ4SfmlaMA=
-----END CERTIFICATE-----`)

// LocalhostKey is the private key for LocalhostCert.
var LocalhostKey = []byte(testingKey(`-----BEGIN RSA TESTING KEY-----
MIICeAIBADANBgkqhkiG9w0BAQEFAASCAmIwggJeAgEAAoGBALEQKmr23faK8kl5
Ar5YlBRi4KVghY7JYVkgIh2nhvOxa5H73vNZBM6a+VqiCI7hnH2h/4UkAOD7Ne5f
lmg+oXog4F0Dpi7WJIpxhMGF7sHOgJLEDPA4xq1PARRf9dG0DJ4y2VnzzRI0vv8U
rMHoGPZN13cYvn6ntoQCQ9Wg8r45AgMBAAECgYEAiAbPT5WQOFPOhzt8LBeIy5Ca
3PImKOf07a+OBhIpzxXCekhxo6oD02Wjo0lQIdSZkLrfvr1GH5FegV7WAgf7rLwG
2IMdV5Gi5nmAxiXEZmMmg5VIq3k2Bo3mh0HajQXRak8LwMHjGyz+P7jmQJ+3A0Iw
OaQXaLWPuT7PNbLOuTECQQDKqbjyJemAL5Wr9ZVTq3whnQa/4yydWdHooxXyfUVE
XLK/l9VFaqqix2rGUOrReI241qP96G6LUD5wg8gDsC0LAkEA36msIfxxmPBri9J+
oq5FndqqXLxcGGCDRhWk5VYkQsmwGazIxBYRqPuktAVFo17XtIfKmgw9j/sNtQPH
EgAkSwJASY+JXft84dZj0WL2rMQV7m18wbHHw+WSV4q6sKXfuoybJQgLlYK+JQ+V
Jh7A3P+REYJ3S/ZOCas6vsRWmWAdOQJBAJcurm6un+6cHGX/25+FIwOHif5zR+Em
Y7Wc7cAjURFgGUvRkkeRD1DlADi7E45Rjoa1/wnP2lEXhvVjX01YkS8CQQCkC4ql
FpDcjbho/JCWOeMFm7JFjWNcHw8WjkZysuADUCqM937AubwPPiYC4t8g3QKtCihU
5Oy2F/u9vtPuEaU1
-----END RSA TESTING KEY-----`))

func testingKey(s string) string { return strings.ReplaceAll(s, "TESTING KEY", "PRIVATE KEY") }
