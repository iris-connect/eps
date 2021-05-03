// IRIS Endpoint-Server (EPS)
// Copyright (C) 2021-2021 The IRIS Endpoint-Server Authors (see AUTHORS.md)
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as
// published by the Free Software Foundation, either version 3 of the
// License, or (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

package jsonrpc

import (
	"fmt"
	"github.com/iris-gateway/eps"
	"github.com/iris-gateway/eps/http"
	"regexp"
	"strings"
)

func uniques(list []string) []string {
	us := make([]string, 0)
	found := make(map[string]bool)
	for _, s := range list {
		s = strings.ToLower(s)
		if _, ok := found[s]; ok {
			continue
		}
		us = append(us, s)
	}
	return us
}

func Cors(settings *CorsSettings, defaultRoute bool) http.Handler {

	if settings == nil {
		return func(c *http.Context) {

		}
	}

	allowedHostPatterns := make([]*regexp.Regexp, len(settings.AllowedHosts))

	for i, allowedHost := range settings.AllowedHosts {
		if pattern, err := regexp.Compile(allowedHost); err != nil {
			// this should not happen...
			panic(err)
		} else {
			allowedHostPatterns[i] = pattern
		}
	}

	decorator := func(c *http.Context) {

		eps.Log.Debugf("Checking cors...")

		allAllowedHeaders := strings.Join(
			uniques(append([]string{c.Request.Header.Get("Access-Control-Request-Headers")},
				settings.AllowedHeaders...)), ", ")

		origin := c.Request.Header.Get("Origin")
		found := false
		for _, pattern := range allowedHostPatterns {
			if pattern.MatchString(origin) {
				c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
				found = true
				break
			}
		}

		if found {
			c.Writer.Header().Set("Access-Control-Max-Age", fmt.Sprintf("%d", 60))
			c.Writer.Header().Set("Access-Control-Allow-Headers", allAllowedHeaders)
			c.Writer.Header().Set("Access-Control-Allow-Methods", strings.Join(settings.AllowedMethods, ", "))

			// for OPTIONS calls we set the status code explicitly
			if c.Request.Method == "OPTIONS" {
				c.AbortWithStatus(200)
				return
			}

		}

		if defaultRoute {
			c.JSON(404, http.H{"message": "route not found"})
			return
		}

	}

	return decorator
}

func CorsFromEverywhere(settings *CorsSettings) http.Handler {

	if settings == nil {
		return func(c *http.Context) {

		}
	}

	decorator := func(c *http.Context) {

		origin := c.Request.Header.Get("Origin")
		c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
		c.Writer.Header().Set("Access-Control-Max-Age", fmt.Sprintf("%d", 60))
		c.Writer.Header().Set("Access-Control-Allow-Headers", c.Request.Header.Get("Access-Control-Request-Headers"))
		c.Writer.Header().Set("Access-Control-Allow-Methods", strings.Join([]string{"POST", "GET"}, ", "))

		// for OPTIONS calls we set the status code explicitly
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(200)
			return
		}

	}

	return decorator
}
