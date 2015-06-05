package uaa

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("UAA", func() {
	It("checks for valid token format", func() {
		publicKey := `-----BEGIN PUBLIC KEY-----
MFwwDQYJKoZIhvcNAQEBBQADSwAwSAJBAN+5O6n85LSs/fj46Ht1jNbc5e+3QX+s
uxVPJqICvuV6sIukJXXEzfblneN2GeEVqgeNvglAU9tnm3OIKzlwM5UCAwEAAQ==
-----END PUBLIC KEY-----
`

		decoder, _ := NewUaaTokenDecoder([]byte(publicKey))

		tokenWithoutBearerString := "token"

		_, err := decoder.Decode(tokenWithoutBearerString)

		Expect(err).To(MatchError(ContainSubstring("invalid authentication header: token")))

		tokenWithInvalidBearerString := "notBearer token"

		_, err = decoder.Decode(tokenWithInvalidBearerString)

		Expect(err).To(MatchError(ContainSubstring("invalid authentication header: notBearer token")))
	})

	It("that token has proper number of segments", func() {
		publicKey := `-----BEGIN PUBLIC KEY-----
MFwwDQYJKoZIhvcNAQEBBQADSwAwSAJBAN+5O6n85LSs/fj46Ht1jNbc5e+3QX+s
uxVPJqICvuV6sIukJXXEzfblneN2GeEVqgeNvglAU9tnm3OIKzlwM5UCAwEAAQ==
-----END PUBLIC KEY-----
`

		decoder, _ := NewUaaTokenDecoder([]byte(publicKey))

		tokenWithNotEnough := "bearer header.payload"

		_, err := decoder.Decode(tokenWithNotEnough)

		Expect(err).To(MatchError(ContainSubstring("Not enough or too many segments")))

		tokenWithTooMuch := "bearer header.payload.crypto.tooMuchStuff"

		_, err = decoder.Decode(tokenWithTooMuch)

		Expect(err).To(MatchError(ContainSubstring("Not enough or too many segments")))
	})

	It("verifies payload signature", func() {
		publicKey := `-----BEGIN PUBLIC KEY-----
MFwwDQYJKoZIhvcNAQEBBQADSwAwSAJBAN+5O6n85LSs/fj46Ht1jNbc5e+3QX+s
uxVPJqICvuV6sIukJXXEzfblneN2GeEVqgeNvglAU9tnm3OIKzlwM5UCAwEAAQ==
-----END PUBLIC KEY-----
`

		decoder, _ := NewUaaTokenDecoder([]byte(publicKey))

		tokenWithInvalidSignature := `bearer eyJ0eXAiOiJKV1QiLCJhbGciOiJSUzI1NiJ9.eyJ1c2VyX2lkIjoiYWJjMTIzNCIsImV4cCI6MTM3NDUyMzg2MH0.JF8dTUJp3NaZhfIhYgesKh-HmV9isnJc51eFaqeFuIhJQ73wiyekfgu-5jSoquVRITSL3cIRjD42F8acbCMYHA`
		_, err := decoder.Decode(tokenWithInvalidSignature)

		Expect(err).To(MatchError(ContainSubstring("Signature verification failed")))

		tokenWithValidSignature := `bearer eyJ0eXAiOiJKV1QiLCJhbGciOiJSUzI1NiJ9.eyJ1c2VyX2lkIjoiYWJjMTIzNCIsImV4cCI6MTM3NDUyMzg2MH0.JF8dTUJp3NaZhfIhYgesKh-HmV9isnJc51eFaqeFuIhJQ73wiyekfgu-5jSoquVRITSL3cIRjD42F8WabCMYHA`

		results, err := decoder.Decode(tokenWithValidSignature)

		Expect(err).NotTo(HaveOccurred())

		Expect(results.UserId).To(Equal("abc1234"))
	})

	It("parses the email", func() {
		publicKey := `-----BEGIN PUBLIC KEY-----
MIGfMA0GCSqGSIb3DQEBAQUAA4GNADCBiQKBgQDHFr+KICms+tuT1OXJwhCUmR2d
KVy7psa8xzElSyzqx7oJyfJ1JZyOzToj9T5SfTIq396agbHJWVfYphNahvZ/7uMX
qHxf+ZH9BL1gk9Y6kCnbM5R60gfwjyW1/dQPjOzn9N394zd2FJoFHwdq9Qs0wBug
spULZVNRxq7veq/fzwIDAQAB
-----END PUBLIC KEY-----
`
		decoder, _ := NewUaaTokenDecoder([]byte(publicKey))

		tokenWithValidScopes := `bearer eyJhbGciOiJSUzI1NiJ9.eyJqdGkiOiI2ZmY3NTRhYy0zYTQ2LTQyNzEtOTE3Yi01YTA4ZWMwOTAyMDYiLCJzdWIiOiIyYTlmNDZmMC1mZGNhLTRjNzgtYmUxMy0wNDUzZDU0MmRjNzgiLCJzY29wZSI6WyJjbG91ZF9jb250cm9sbGVyLnJlYWQiLCJjbG91ZF9jb250cm9sbGVyLndyaXRlIiwibG9nZ3JlZ2F0b3IiLCJvcGVuaWQiLCJwYXNzd29yZC53cml0ZSJdLCJjbGllbnRfaWQiOiJjZiIsImNpZCI6ImNmIiwiZ3JhbnRfdHlwZSI6InBhc3N3b3JkIiwidXNlcl9pZCI6IjJhOWY0NmYwLWZkY2EtNGM3OC1iZTEzLTA0NTNkNTQyZGM3OCIsInVzZXJfbmFtZSI6InVzZXIxQGV4YW1wbGUuY29tIiwiZW1haWwiOiJ1c2VyMUBleGFtcGxlLmNvbSIsImlhdCI6MTM3NzUyNjgyNywiZXhwIjoxMzc3NTM0MDI3LCJpc3MiOiJodHRwczovL3VhYS5vYWsuY2YtYXBwLmNvbS9vYXV0aC90b2tlbiIsImF1ZCI6WyJvcGVuaWQiLCJjbG91ZF9jb250cm9sbGVyIiwicGFzc3dvcmQiXX0.Vobn4P7HHGkhaeGhHeS2LWccWQ4HmlhgUiu9JaRlZEMPH6hnrCH8VKKwZQfXObENydgqcs3C85_nT4a94vmtG9dDDSxWZ8juJbfmsftud31j0_s_Y3iV-NekY0EbuH_2MG0DqJc9Xl2aJIbJ1OIX9Dr1e9krtMHmjmia0jErHUU`

		results, err := decoder.Decode(tokenWithValidScopes)

		Expect(err).NotTo(HaveOccurred())

		Expect(results.UserId).To(Equal("2a9f46f0-fdca-4c78-be13-0453d542dc78"))
		Expect(results.Email).To(Equal("user1@example.com"))
	})
})
