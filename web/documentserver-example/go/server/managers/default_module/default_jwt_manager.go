/**
 *
 * (c) Copyright Ascensio System SIA 2021
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

package default_managers

import (
	"errors"
	"strings"

	"github.com/ONLYOFFICE/document-server-integration/config"
	"github.com/ONLYOFFICE/document-server-integration/server/managers"
	"github.com/ONLYOFFICE/document-server-integration/server/models"
	"github.com/golang-jwt/jwt"
	"github.com/mitchellh/mapstructure"
	"go.uber.org/zap"
)

type DefaultJwtManager struct {
	config config.ApplicationConfig
	logger *zap.SugaredLogger
}

func NewDefaultJwtManager(config config.ApplicationConfig, logger *zap.SugaredLogger) managers.JwtManager {
	return &DefaultJwtManager{
		config,
		logger,
	}
}

func (jm *DefaultJwtManager) JwtSign(payload jwt.Claims, key []byte) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, payload)
	ss, err := token.SignedString(key)

	if err != nil {
		return "", errors.New("jwt could not create a signed string with the given key")
	}
	return ss, nil
}

func (jm *DefaultJwtManager) JwtDecode(jwtString string, key []byte) (jwt.MapClaims, error) {
	if jwtString == "" {
		return nil, errors.New("jwt string is empty")
	}

	token, err := jwt.Parse(jwtString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected JWT signing method")
		}

		return key, nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return claims, nil
	} else {
		return nil, errors.New("jwt token is not valid")
	}
}

func (jm *DefaultJwtManager) ParseCallback(body *models.Callback, token_header string) error {
	secret := strings.TrimSpace(jm.config.JwtSecret)
	if secret != "" && jm.config.JwtEnabled {
		var decodedCallback jwt.MapClaims
		var jwtDecodingErr error

		if token_header == "" {
			decodedCallback, jwtDecodingErr = jm.JwtDecode(body.Token, []byte(secret))
		} else {
			decodedCallback, jwtDecodingErr = jm.JwtDecode(strings.Split(token_header, " ")[1], []byte(secret))
		}

		if jwtDecodingErr != nil {
			return errors.New("could not process JWT in callback body")
		}

		err := mapstructure.Decode(decodedCallback, &body)

		if err != nil {
			return errors.New("could not populate callback body with decoded JWT")
		}
	}
	return nil
}