package auth

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"gopkg.in/square/go-jose.v2/jwt"

	"github.com/lgcshy/protocol/utils"
)

func TestAccessToken(t *testing.T) {
	t.Parallel()

	t.Run("keys must be set", func(t *testing.T) {
		token := NewAccessToken("", "")
		_, err := token.ToJWT()
		require.Equal(t, ErrKeysMissing, err)
	})

	t.Run("generates a decode-able key", func(t *testing.T) {
		apiKey, secret := apiKeypair()
		videoGrant := &VideoGrant{RoomJoin: true, Room: "myroom"}
		at := NewAccessToken(apiKey, secret).
			AddGrant(videoGrant).
			SetValidFor(time.Minute * 5).
			SetIdentity("user")
		value, err := at.ToJWT()
		//fmt.Println(raw)
		require.NoError(t, err)

		require.Len(t, strings.Split(value, "."), 3)

		// ensure it's a valid JWT
		token, err := jwt.ParseSigned(value)
		require.NoError(t, err)

		decodedGrant := ClaimGrants{}
		err = token.UnsafeClaimsWithoutVerification(&decodedGrant)
		require.NoError(t, err)

		require.EqualValues(t, videoGrant, decodedGrant.Video)
	})

	t.Run("default validity should be more than a minute", func(t *testing.T) {
		apiKey, secret := apiKeypair()
		videoGrant := &VideoGrant{RoomJoin: true, Room: "myroom"}
		at := NewAccessToken(apiKey, secret).
			AddGrant(videoGrant)
		value, err := at.ToJWT()
		token, err := jwt.ParseSigned(value)

		claim := jwt.Claims{}
		decodedGrant := ClaimGrants{}
		err = token.UnsafeClaimsWithoutVerification(&claim, &decodedGrant)
		require.NoError(t, err)
		require.EqualValues(t, videoGrant, decodedGrant.Video)

		// default validity
		require.True(t, claim.Expiry.Time().Sub(claim.IssuedAt.Time()) > time.Minute)
	})

	t.Run("backwards compatible with jti identity tokens", func(t *testing.T) {
		apiKey, secret := apiKeypair()
		videoGrant := &VideoGrant{RoomJoin: true, Room: "myroom"}
		at := NewAccessToken(apiKey, secret).
			AddGrant(videoGrant).
			SetValidFor(time.Minute * 5).
			SetIdentity("user")
		value, err := at.toJWTOld()
		//fmt.Println(raw)
		require.NoError(t, err)

		verifier, err := ParseAPIToken(value)
		require.NoError(t, err)

		grants, err := verifier.Verify(secret)
		require.NoError(t, err)
		require.Equal(t, "user", grants.Identity)
	})
}

func apiKeypair() (string, string) {
	return utils.NewGuid(utils.APIKeyPrefix), utils.RandomSecret()
}
