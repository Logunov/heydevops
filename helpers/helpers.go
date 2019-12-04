/*
Copyright © 2019 Ilya V. Logounov <ilya@logounov.com>

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program. If not, see <http://www.gnu.org/licenses/>.
*/

package helpers

import (
	"github.com/sirupsen/logrus"
	"time"
)

var log *logrus.Logger

func SetLogger(logger *logrus.Logger) {
	log = logger
}

func Elapsed(what string) func() {
	start := time.Now()
	return func() {
		log.Infof("%s took %v", what, time.Since(start))
	}
}

func CheckTrace(error error) {
	if error != nil {
		log.Trace("ERROR: " + error.Error())
	}
}

func CheckDebug(error error) {
	if error != nil {
		log.Debug("DEBUG: " + error.Error())
	}
}

func CheckError(error error) {
	if error != nil {
		log.Error("ERROR: " + error.Error())
	}
}

func CheckPanic(error error) {
	if error != nil {
		log.Panic("PANIC: " + error.Error())
	}
}
