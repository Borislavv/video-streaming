package tokenizer

import (
	"context"
	"github.com/Borislavv/video-streaming/internal/domain/agg"
	"github.com/Borislavv/video-streaming/internal/domain/errtype"
	"github.com/Borislavv/video-streaming/internal/domain/logger/interface"
	repositoryinterface "github.com/Borislavv/video-streaming/internal/domain/repository/interface"
	"github.com/Borislavv/video-streaming/internal/domain/service/di/interface"
	"github.com/Borislavv/video-streaming/internal/domain/vo"
	"github.com/golang-jwt/jwt/v5"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"strings"
	"time"
)

type JwtService struct {
	ctx                     context.Context
	logger                  loggerinterface.Logger
	blockedTokenRepository  repositoryinterface.BlockedToken
	jwtTokenAcceptedIssuers []string
	jwtSecretSalt           []byte
	jwtTokenIssuer          string
	jwtTokenEncryptAlgo     string
	jwtTokenExpiresAfter    int64
}

func NewJwtService(serviceContainer diinterface.ServiceContainer) (*JwtService, error) {
	loggerService, err := serviceContainer.GetLoggerService()
	if err != nil {
		return nil, err
	}

	ctx, err := serviceContainer.GetCtx()
	if err != nil {
		return nil, loggerService.LogPropagate(err)
	}

	blockedTokenRepository, err := serviceContainer.GetBlockedTokenRepository()
	if err != nil {
		return nil, loggerService.LogPropagate(err)
	}

	cfg, err := serviceContainer.GetConfig()
	if err != nil {
		return nil, loggerService.LogPropagate(err)
	}

	return &JwtService{
		ctx:                     ctx,
		logger:                  loggerService,
		blockedTokenRepository:  blockedTokenRepository,
		jwtTokenAcceptedIssuers: strings.Split(cfg.JwtTokenAcceptedIssuers, ","),
		jwtSecretSalt:           []byte(cfg.JwtSecretSalt),
		jwtTokenIssuer:          cfg.JwtTokenIssuer,
		jwtTokenEncryptAlgo:     cfg.JwtTokenEncryptAlgo,
		jwtTokenExpiresAfter:    cfg.JwtTokenExpiresAfter,
	}, nil
}

// New will generate a new JWT.
func (s *JwtService) New(user *agg.User) (token string, err error) {
	tkn := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": user.ID.Value.Hex(),
		"iss": s.jwtTokenIssuer,
		"exp": &jwt.NumericDate{Time: time.Now().Add(time.Second * time.Duration(s.jwtTokenExpiresAfter))},
	})

	if token, err = tkn.SignedString(s.jwtSecretSalt); err != nil {
		return "", s.logger.LogPropagate(err)
	} else {
		return token, nil
	}
}

// Verify will decode the token and return a user ID or error, if it was occurred.
func (s *JwtService) Verify(token string) (userID vo.ID, err error) {
	parsedToken, err := jwt.Parse(token, func(decodedToken *jwt.Token) (interface{}, error) {
		if decodedToken.Header["alg"] != s.jwtTokenEncryptAlgo {
			// user must be banned here because the algo wasn't matched
			return nil, errtype.NewTokenAlgoWasNotMatchedInternalError(token)
		}
		// cast to the configured givenToken signature type (stored in `s.jwtTokenEncryptAlgo`)
		if _, success := decodedToken.Method.(*jwt.SigningMethodHMAC); !success {
			return nil, errtype.NewTokenUnexpectedSigningMethodInternalError(token, decodedToken.Header["alg"])
		}
		// jwtSecretSalt is a string containing your secret, but you need pass the []byte
		return s.jwtSecretSalt, nil
	})
	if err != nil {
		// parsing givenToken error occurred
		s.logger.Log(err)
		// return a token invalid error
		return vo.ID{}, s.logger.LogPropagate(errtype.NewAccessTokenIsInvalidError())
	}

	// checking that token is not blocked
	found, err := s.blockedTokenRepository.Has(s.ctx, parsedToken.Raw)
	if err != nil {
		return vo.ID{}, s.logger.LogPropagate(err)
	}
	if found {
		return vo.ID{}, s.logger.LogPropagate(errtype.NewAccessTokenWasBlockedError())
	}

	// extracting claims of the givenToken payload
	if claims, success := parsedToken.Claims.(jwt.MapClaims); success && parsedToken.Valid {
		if err = s.isValidIssuer(token, claims); err != nil {
			// the issuer is not valid, log it
			s.logger.Log(err)
			// return a token invalid error
			return vo.ID{}, s.logger.LogPropagate(errtype.NewAccessTokenIsInvalidError())
		}

		userID, err = s.getUserID(claims)
		if err != nil {
			return vo.ID{}, s.logger.LogPropagate(err)
		}

		return userID, nil
	} else {
		// error occurred while extracting claims from givenToken or givenToken is not valid
		s.logger.Log(errtype.NewTokenInvalidInternalError(token))
		// return a token invalid error
		return vo.ID{}, s.logger.LogPropagate(errtype.NewAccessTokenIsInvalidError())
	}
}

// Block will mark the token as blocked into the storage.
func (s *JwtService) Block(token string, reason string) error {
	userID, err := s.parseUserID(token)
	if err != nil {
		// block token anyway and log message,
		// because the token may be is invalid but must be blocked.
		// in this case a user will be undetermined
		s.logger.Log(err)
	}

	if err = s.blockedTokenRepository.Insert(s.ctx, agg.NewBlockedToken(token, reason, userID)); err != nil {
		return s.logger.LogPropagate(err)
	}
	return nil
}

func (s *JwtService) parseUserID(token string) (userID vo.ID, err error) {
	parsedToken, err := jwt.Parse(token, func(decodedToken *jwt.Token) (interface{}, error) {
		if decodedToken.Header["alg"] != s.jwtTokenEncryptAlgo {
			// user must be banned here because the algo wasn't matched
			return nil, errtype.NewTokenAlgoWasNotMatchedInternalError(token)
		}
		// cast to the configured givenToken signature type (stored in `s.jwtTokenEncryptAlgo`)
		if _, success := decodedToken.Method.(*jwt.SigningMethodHMAC); !success {
			return nil, errtype.NewTokenUnexpectedSigningMethodInternalError(token, decodedToken.Header["alg"])
		}
		// jwtSecretSalt is a string containing your secret, but you need pass the []byte
		return s.jwtSecretSalt, nil
	})
	if err != nil {
		// parsing givenToken error occurred
		s.logger.Log(err)
		// return a token invalid error
		return vo.ID{}, s.logger.LogPropagate(errtype.NewAccessTokenIsInvalidError())
	}

	if claims, success := parsedToken.Claims.(jwt.MapClaims); success {
		if userID, err = s.getUserID(claims); err != nil {
			return vo.ID{}, s.logger.LogPropagate(err)
		} else {
			return userID, nil
		}
	} else {
		return vo.ID{}, s.logger.LogPropagate(errtype.NewAccessTokenIsInvalidError())
	}
}

func (s *JwtService) isValidIssuer(token string, claims jwt.Claims) error {
	// extracting the token issuer
	iss, err := claims.GetIssuer()
	if err != nil {
		return s.logger.LogPropagate(err)
	}
	// checking that token issuer is valid
	if iss != s.jwtTokenIssuer && !(func() (isIssuerWasMatched bool) {
		for _, acceptedIssuer := range s.jwtTokenAcceptedIssuers {
			if iss == acceptedIssuer {
				return true
			}
		}
		return false
	}()) {
		return errtype.NewTokenIssuerWasNotMatchedInternalError(token)
	}

	return nil
}

func (s *JwtService) getUserID(claims jwt.Claims) (userID vo.ID, err error) {
	// extracting subject (hexID of user) from the claims
	hexID, err := claims.GetSubject()
	if err != nil {
		return vo.ID{}, s.logger.LogPropagate(err)
	}
	// creating a user object ID from hex
	oID, err := primitive.ObjectIDFromHex(hexID)
	if err != nil {
		return vo.ID{}, s.logger.LogPropagate(err)
	}
	// returning a success response
	return vo.ID{Value: oID}, nil
}
