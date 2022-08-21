package rsa

import (
    "testing"
    . "github.com/smartystreets/goconvey/convey"
)

func getPrivateKeyForTest() []byte {
    pri := `-----BEGIN RSA PRIVATE KEY-----
MIICXwIBAAKBgQC525iaO1g+tXx2RNqPksRGUjwK1Pd3cRTsIxt/D/FMx0pli26G
fwMThWQDn7tJ1MvwBqNw+MuYPwQftEsHAkLXF0Kn8dBjT0PpflXbJwLgNM4W9T8L
NN0a5+EEm83EkVEGcOWqSjMNSUNxIIdTUNdihcWFv+zHBIvaYWBxX6h58QIDAQAB
AoGBAIEQE0qXB1KUqNdgPP4ShyXmGTfUZ/yTlFnej0uPAJu2kN0vFBNlw/ccXDWA
CIjesrf9hCYBPzB8Ihr6ElfNpCeVJcKUyVtiiTbn7XcNSXOn8YtMftk+sR1a4MQY
6SUpFiB11wjkqhI88qU3knjNoHj78PAVi/2mw6OY17GOtSehAkEA7lsn9QQXuzu6
rw6dYVviJxdGPcIIGGj4GB4hLW5g6eFpTUY+01p2BpeCjhBI6uteObnRvNQ6ZgwR
SPfstdVhRQJBAMedmTI/mSwH8yG9n1drVpAZGhohQaGqh+JS1Ia+tmMju7ezoeUP
QDj6607ubSnqw7q58oEwRPY+ecRWxjdqor0CQQDCCiN1K2fGXNGVQWiNoadx+1iL
XjII7StLNvv7aCgtPfvjlJQAq1v58c2uqUMzO3jxtXwxJPSFrr1DkdF6FcOhAkEA
hGDppb8zj1W+UZP1Rf4zK+DZxJZldhcnglo4AxwazGh4Jv2D0eppRuBwiKnpzzCX
mQ+T2UTvlvYbvq9lSH75aQJBAK8DKB1+hyk1sQVzVtvC2n3/vplLdxkXayZ7qIRL
kTcwxgUY8RML6ez3fV/732KmgKFg3h9E0YVnsR3F+4j+X98=
-----END RSA PRIVATE KEY-----`

    return []byte(pri)
}

func getPublicKeyForTest() []byte {
    pk := `-----BEGIN PUBLIC KEY-----
MIGfMA0GCSqGSIb3DQEBAQUAA4GNADCBiQKBgQC525iaO1g+tXx2RNqPksRGUjwK
1Pd3cRTsIxt/D/FMx0pli26GfwMThWQDn7tJ1MvwBqNw+MuYPwQftEsHAkLXF0Kn
8dBjT0PpflXbJwLgNM4W9T8LNN0a5+EEm83EkVEGcOWqSjMNSUNxIIdTUNdihcWF
v+zHBIvaYWBxX6h58QIDAQAB
-----END PUBLIC KEY-----`

    return []byte(pk)
}

func TestSignRSA2(t *testing.T) {
    Convey("RSA2", t, func() {
        Convey("SignRSA2", func() {
            src := "hello"
            _, err := SignRSA2([]byte(src), getPrivateKeyForTest())
            So(err, ShouldEqual, nil)
        })

        Convey("VerifyRSA2", func() {
            src := []byte("hello")
            sig, err := SignRSA2(src, getPrivateKeyForTest())
            So(err, ShouldEqual, nil)

            err = VerifyRSA2(src, getPublicKeyForTest(), sig)
            So(err, ShouldEqual, nil)
        })
    })
}
