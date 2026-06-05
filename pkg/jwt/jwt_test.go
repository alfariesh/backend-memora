package jwt_test

import (
	"testing"
	"time"

	"github.com/alfariesh/backend-memora/pkg/jwt"
	jwtlib "github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestJWT_GenerateAndParse(t *testing.T) {
	t.Parallel()

	j := jwt.New("test-secret", time.Hour)

	token, err := j.GenerateToken("user-123")
	require.NoError(t, err)
	assert.NotEmpty(t, token)

	userID, err := j.ParseToken(token)
	require.NoError(t, err)
	assert.Equal(t, "user-123", userID)
}

func TestJWT_ParseToken_Invalid(t *testing.T) {
	t.Parallel()

	j := jwt.New("test-secret", time.Hour)

	_, err := j.ParseToken("invalid-token")
	require.Error(t, err)
}

func TestJWT_ParseToken_WrongSecret(t *testing.T) {
	t.Parallel()

	j1 := jwt.New("secret-1", time.Hour)
	j2 := jwt.New("secret-2", time.Hour)

	token, err := j1.GenerateToken("user-123")
	require.NoError(t, err)

	_, err = j2.ParseToken(token)
	require.Error(t, err)
}

func TestJWT_ParseToken_Expired(t *testing.T) {
	t.Parallel()

	j := jwt.New("test-secret", -time.Hour)

	token, err := j.GenerateToken("user-123")
	require.NoError(t, err)

	_, err = j.ParseToken(token)
	require.Error(t, err)
}

func TestJWT_GenerateToken_EmptySubject(t *testing.T) {
	t.Parallel()

	j := jwt.New("test-secret", time.Hour)

	token, err := j.GenerateToken("")

	require.ErrorIs(t, err, jwt.ErrInvalidTokenClaims)
	assert.Empty(t, token)
}

func TestJWT_ParseToken_StrictClaims(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		claims jwtlib.MapClaims
		method jwtlib.SigningMethod
	}{
		{
			name:   "wrong algorithm",
			claims: validAccessClaims(),
			method: jwtlib.SigningMethodHS512,
		},
		{
			name: "missing exp",
			claims: mutateClaims(validAccessClaims(), func(claims jwtlib.MapClaims) {
				delete(claims, "exp")
			}),
			method: jwtlib.SigningMethodHS256,
		},
		{
			name: "wrong issuer",
			claims: mutateClaims(validAccessClaims(), func(claims jwtlib.MapClaims) {
				claims["iss"] = "other-service"
			}),
			method: jwtlib.SigningMethodHS256,
		},
		{
			name: "wrong audience",
			claims: mutateClaims(validAccessClaims(), func(claims jwtlib.MapClaims) {
				claims["aud"] = "other-client"
			}),
			method: jwtlib.SigningMethodHS256,
		},
		{
			name: "wrong token type",
			claims: mutateClaims(validAccessClaims(), func(claims jwtlib.MapClaims) {
				claims["typ"] = "refresh"
			}),
			method: jwtlib.SigningMethodHS256,
		},
		{
			name: "empty subject",
			claims: mutateClaims(validAccessClaims(), func(claims jwtlib.MapClaims) {
				claims["sub"] = ""
			}),
			method: jwtlib.SigningMethodHS256,
		},
		{
			name: "missing issued at",
			claims: mutateClaims(validAccessClaims(), func(claims jwtlib.MapClaims) {
				delete(claims, "iat")
			}),
			method: jwtlib.SigningMethodHS256,
		},
		{
			name: "missing not before",
			claims: mutateClaims(validAccessClaims(), func(claims jwtlib.MapClaims) {
				delete(claims, "nbf")
			}),
			method: jwtlib.SigningMethodHS256,
		},
	}

	for _, tc := range tests {
		localTc := tc

		t.Run(localTc.name, func(t *testing.T) {
			t.Parallel()

			j := jwt.New("test-secret", time.Hour)
			token := signedToken(t, localTc.method, localTc.claims)

			_, err := j.ParseToken(token)

			require.Error(t, err)
		})
	}
}

func validAccessClaims() jwtlib.MapClaims {
	now := time.Now().UTC()

	return jwtlib.MapClaims{
		"sub": "user-123",
		"iss": "backend-memora",
		"aud": "memora-clients",
		"iat": now.Unix(),
		"nbf": now.Unix(),
		"exp": now.Add(time.Hour).Unix(),
		"typ": "access",
	}
}

func mutateClaims(claims jwtlib.MapClaims, mutate func(jwtlib.MapClaims)) jwtlib.MapClaims {
	copied := make(jwtlib.MapClaims, len(claims))
	for key, value := range claims {
		copied[key] = value
	}

	mutate(copied)

	return copied
}

func signedToken(t *testing.T, method jwtlib.SigningMethod, claims jwtlib.MapClaims) string {
	t.Helper()

	token, err := jwtlib.NewWithClaims(method, claims).SignedString([]byte("test-secret"))
	require.NoError(t, err)

	return token
}
