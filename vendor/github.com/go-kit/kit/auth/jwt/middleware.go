package jwt

import (
	"errors"

	jwt "github.com/dgrijalva/jwt-go"
	"golang.org/x/net/context"

	"github.com/go-kit/kit/endpoint"
)

type contextKey string

const (
	// JWTTokenContextKey holds the key used to store a JWT Token in the
	// context.
	JWTTokenContextKey contextKey = "JWTToken"
	// JWTClaimsContxtKey holds the key used to store the JWT Claims in the
	// context.
	JWTClaimsContextKey contextKey = "JWTClaims"
)

var (
	// ErrTokenContextMissing denotes a token was not passed into the parsing
	// middleware's context.
	ErrTokenContextMissing = errors.New("token up for parsing was not passed through the context")
	// ErrTokenInvalid denotes a token was not able to be validated.
	ErrTokenInvalid = errors.New("JWT Token was invalid")
	// ErrTokenExpired denotes a token's expire header (exp) has since passed.
	ErrTokenExpired = errors.New("JWT Token is expired")
	// ErrTokenMalformed denotes a token was not formatted as a JWT token.
	ErrTokenMalformed = errors.New("JWT Token is malformed")
	// ErrTokenNotActive denotes a token's not before header (nbf) is in the
	// future.
	ErrTokenNotActive = errors.New("token is not valid yet")
	// ErrUncesptedSigningMethod denotes a token was signed with an unexpected
	// signing method.
	ErrUnexpectedSigningMethod = errors.New("unexpected signing method")
)

type Claims map[string]interface{}

// NewSigner creates a new JWT token generating middleware, specifying key ID,
// signing string, signing method and the claims you would like it to contain.
// Tokens are signed with a Key ID header (kid) which is useful for determining
// the key to use for parsing. Particularly useful for clients.
func NewSigner(kid string, key []byte, method jwt.SigningMethod, claims Claims) endpoint.Middleware {
	return func(next endpoint.Endpoint) endpoint.Endpoint {
		return func(ctx context.Context, request interface{}) (response interface{}, err error) {
			token := jwt.NewWithClaims(method, jwt.MapClaims(claims))
			token.Header["kid"] = kid

			// Sign and get the complete encoded token as a string using the secret
			tokenString, err := token.SignedString(key)
			if err != nil {
				return nil, err
			}
			ctx = context.WithValue(ctx, JWTTokenContextKey, tokenString)

			return next(ctx, request)
		}
	}
}

// NewParser creates a new JWT token parsing middleware, specifying a
// jwt.Keyfunc interface and the signing method. NewParser adds the resulting
// claims to endpoint context or returns error on invalid token. Particularly
// useful for servers.
func NewParser(keyFunc jwt.Keyfunc, method jwt.SigningMethod) endpoint.Middleware {
	return func(next endpoint.Endpoint) endpoint.Endpoint {
		return func(ctx context.Context, request interface{}) (response interface{}, err error) {
			// tokenString is stored in the context from the transport handlers.
			tokenString, ok := ctx.Value(JWTTokenContextKey).(string)
			if !ok {
				return nil, ErrTokenContextMissing
			}

			// Parse takes the token string and a function for looking up the
			// key. The latter is especially useful if you use multiple keys
			// for your application.  The standard is to use 'kid' in the head
			// of the token to identify which key to use, but the parsed token
			// (head and claims) is provided to the callback, providing
			// flexibility.
			token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
				// Don't forget to validate the alg is what you expect:
				if token.Method != method {
					return nil, ErrUnexpectedSigningMethod
				}

				return keyFunc(token)
			})
			if err != nil {
				if e, ok := err.(*jwt.ValidationError); ok && e.Inner != nil {
					if e.Errors&jwt.ValidationErrorMalformed != 0 {
						// Token is malformed
						return nil, ErrTokenMalformed
					} else if e.Errors&jwt.ValidationErrorExpired != 0 {
						// Token is expired
						return nil, ErrTokenExpired
					} else if e.Errors&jwt.ValidationErrorNotValidYet != 0 {
						// Token is not active yet
						return nil, ErrTokenNotActive
					}

					return nil, e.Inner
				}

				return nil, err
			}

			if !token.Valid {
				return nil, ErrTokenInvalid
			}

			if claims, ok := token.Claims.(jwt.MapClaims); ok {
				ctx = context.WithValue(ctx, JWTClaimsContextKey, Claims(claims))
			}

			return next(ctx, request)
		}
	}
}
