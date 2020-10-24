/*
 * Copyright (c) 2020 - present Kurtosis Technologies LLC.
 * All Rights Reserved.
 */

package access_controller

import (
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"github.com/kurtosis-tech/kurtosis/initializer/access_controller/auth0_authorizer"
	"github.com/kurtosis-tech/kurtosis/initializer/access_controller/auth0_constants"
	"github.com/kurtosis-tech/kurtosis/initializer/access_controller/encrypted_session_cache"
	"github.com/palantir/stacktrace"
	"github.com/sirupsen/logrus"
	"time"
)

const (

	// How long we'll pause after displaying auth warnings, to give users a chance to see it
	authWarningPause = 3 * time.Second

	// For extra security, make sure only the user can read & write the session cache file
	sessionCacheFilePerms = 0600

	// Key in the Headers hashmap of the token that points to the key ID
	keyIdTokenHeaderKey = "kid"

	// Header and footer to attach to base64-encoded key data that we receive from Auth0
	pubKeyHeader = "-----BEGIN CERTIFICATE-----"
	pubKeyFooter = "-----END CERTIFICATE-----"
)

/*
Used for a developer running Kurtosis on their local machine. This will:

1) Check if they have a valid session cached locally that's still valid and, if not
2) Prompt them for their username and password

Args:
	sessionCacheFilepath: Filepath to store the encrypted session cache at

Returns:
	An error if and only if an irrecoverable login error occurred
 */
func RunDeveloperMachineAuthFlow(sessionCacheFilepath string) error {
	/*
	NOTE: As of 2020-10-24, we actually don't strictly *need* to encrypt anything on disk because we hardcode the
	 Auth0 public keys used for verifying tokens so unless the user cracks Auth0 and gets the private key, there's
	 no way for a user to forge a token.

	However, this hardcode-public-keys approach becomes much harder if we start doing private key rotation (which would
	 be good security hygiene) because:
	  a) now our code needs to dynamically discover what public keys it should use and
	  b) Kurtosis needs to work even if the developer is offline
	The offline requirement is the real kicker, because it means we need to write the public keys to the developer's local
	 machine and somehow protect it from tampering. This likely means encrypting the data, which means having an encryption
	 key in the code, which would shift the weakpoint to someone decompiling kurtosis-core and discovering the encryption
	 key there.
	*/
	cache := encrypted_session_cache.NewEncryptedSessionCache(sessionCacheFilepath, sessionCacheFilePerms)

	tokenStr, err := getTokenStr(cache)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting the token string")
	}

	claims, err := parseAndValidateTokenClaims(tokenStr)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred parsing and validating the token claims")
	}

	scope, err := getScopeFromClaimsAndRenewIfNeeded(claims, cache)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred renewing the token or getting the scope from the token's claims")
	}

	if scope != auth0_constants.ExecutionScope {
		return stacktrace.NewError(
			"Kurtosis requires scope '%v' to run but token has scope '%v'; this is most likely due to an expired Kurtosis license",
			auth0_constants.ExecutionScope,
			scope)
	}
	return nil
}

/*
This workflow is for authenticating and authorizing Kurtosis tests running in CI (no device or username).
	See also: https://www.oauth.com/oauth2-servers/access-tokens/client-credentials/
 */
func RunCIAuthFlow(clientId string, clientSecret string) error {
	tokenResponse, err := auth0_authorizer.AuthorizeClientCredentials(clientId, clientSecret)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred authenticating with the client ID & secret")
	}

	claims, err := parseAndValidateTokenClaims(tokenResponse.AccessToken)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred parsing and validating the token claims")
	}

	scope := claims.Scope
	if scope != auth0_constants.ExecutionScope {
		return stacktrace.NewError(
			"Kurtosis requires scope '%v' to run but token has scope '%v'; this is most likely due to an expired Kurtosis license",
			auth0_constants.ExecutionScope,
			scope)
	}
	return nil
}

// ============================== PRIVATE HELPER FUNCTIONS =========================================
/*
Gets the token string, either by reading a valid cache or by prompting the user for their login credentials
 */
func getTokenStr(cache *encrypted_session_cache.EncryptedSessionCache) (string, error) {
	var result string
	session, err := cache.LoadSession()
	if err != nil {
		// We couldn't load any cached session, so the user MUST log in
		logrus.Tracef("The following error occurred loading the session from file: %v", err)
		tokenResponse, err := auth0_authorizer.AuthorizeUserDevice()
		if err != nil {
			return "", stacktrace.Propagate(err, "An error occurred during Auth0 authentication")
		}

		// The user has successfully authenticated, so we're good to go
		newSession := encrypted_session_cache.Session{
			Token:                    tokenResponse.AccessToken,
		}
		if err := cache.SaveSession(newSession); err != nil {
			logrus.Warnf("We received a token from Auth0 but the following error occurred when caching it locally:")
			fmt.Fprintln(logrus.StandardLogger().Out, err)
			logrus.Warn("If this error isn't corrected, you'll need to log into Kurtosis every time you run it")
			time.Sleep(authWarningPause)
		}
		result = tokenResponse.AccessToken
	} else {
		// We were able to load a session
		result = session.Token
	}

	return result, nil
}

func parseAndValidateTokenClaims(tokenStr string) (auth0_authorizer.Auth0TokenClaims, error) {
	// This includes validation like expiration date, issuer, etc.
	// See Auth0TokenClaims for more details

	token, err := new(jwt.Parser).ParseWithClaims(
		tokenStr,
		&auth0_authorizer.Auth0TokenClaims{},
		func(token *jwt.Token) (interface{}, error) {
			// IMPORTANT: Validating the algorithm per https://godoc.org/github.com/dgrijalva/jwt-go#example-Parse--Hmac
			if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
				return nil, stacktrace.NewError(
					"Expected token algorithm '%v' but got '%v'",
					jwt.SigningMethodRS256.Name,
					token.Header)
			}

			untypedKeyId, found := token.Header[keyIdTokenHeaderKey]
			if !found {
				return nil, stacktrace.NewError("No key ID key '%v' found in token header", keyIdTokenHeaderKey)
			}
			keyId, ok := untypedKeyId.(string)
			if !ok {
				return nil, stacktrace.NewError("Found key ID, but value was not a string")
			}

			keyBase64, found := auth0_constants.RsaPublicKeyBase64[keyId]
			if !found {
				return nil, stacktrace.NewError("No public RSA key found corresponding to key ID from token '%v'", keyId)
			}
			keyStr := pubKeyHeader + "\n" + keyBase64 + "\n" + pubKeyFooter

			pubKey, err := jwt.ParseRSAPublicKeyFromPEM([]byte(keyStr))
			if err != nil {
				return nil, stacktrace.Propagate(err, "An error occurred parsing the public key base64 for key ID '%v'; this is a code bug", keyId)
			}

			return pubKey, nil
		},
	)

	if err != nil {
		return auth0_authorizer.Auth0TokenClaims{}, stacktrace.Propagate(err, "An error occurred parsing or validating the JWT token")
	}

	claims, ok := token.Claims.(*auth0_authorizer.Auth0TokenClaims)
	if !ok {
		return auth0_authorizer.Auth0TokenClaims{}, stacktrace.NewError("Could not cast token claims to Auth0 token claims object, indicating an invalid token")
	}

	return *claims, nil
}

func getScopeFromClaimsAndRenewIfNeeded(claims auth0_authorizer.Auth0TokenClaims, cache *encrypted_session_cache.EncryptedSessionCache) (string, error) {
	now := time.Now()
	expiration := time.Unix(claims.ExpiresAt, 0)
	if expiration.After(now) {
		return claims.Scope, nil
	}

	// If we've gotten here, it means that the token is beyond the expiration but not beyond the grace period (else
	//  token validation would have failed completely)
	expirationExceededAmount := now.Sub(expiration)
	logrus.Infof("Kurtosis token expired %v ago; attempting to get a new token...", expirationExceededAmount)
	newTokenResponse, err := auth0_authorizer.AuthorizeUserDevice()
	if err != nil {
		logrus.Debugf("Token expiration error: %v", err)
		logrus.Warnf(
			"WARNING: Your Kurtosis token expired %v ago and we couldn't reach Auth0 to get a new one",
			expirationExceededAmount)
		logrus.Warnf(
			"You will have a grace period of %v from expiration to get a connection to Auth0 before Kurtosis stops working",
			claims.GetGracePeriod())
		time.Sleep(authWarningPause)

		// NOTE: If it's annoying for users for Kurtosis to try and hit Auth0 on every run after their token is expired
		//  (say, they have to wait for the connection to time out) then we can add a tracker in the session on the last
		//  time we warned them and only warn them say every 3 hours

		return claims.Scope, nil
	}

	// If we've gotten here, the user's token was expired but we were able to connect and get a new one
	newSession := encrypted_session_cache.Session{
		Token: newTokenResponse.AccessToken,
	}
	if err := cache.SaveSession(newSession); err != nil {
		logrus.Warnf("We received a new token from Auth0 but the following error occurred when caching it locally:")
		fmt.Fprintln(logrus.StandardLogger().Out, err)
		logrus.Warn("If this error isn't corrected, you'll need to log into Kurtosis every time you run it")
		time.Sleep(authWarningPause)
	}
	return newTokenResponse.Scope, nil
}