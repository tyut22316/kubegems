// Copyright 2022 The kubegems.io Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package httpsigs

import (
	"crypto/md5"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
)

var signer *Signer

const (
	token       = "fedf4ca02b9235a616ccb11d9540bb42"
	headerToken = "sign-token"
	headerTime  = "sign-time"
)

type Signer struct {
	Token     string
	Duration  int64
	WhiteList []string
}

func init() {
	signer = &Signer{
		Token: token,
		//延长时延误差
		Duration: 60,
	}
}

func GetSigner() *Signer {
	return signer
}

func (s *Signer) AddWhiteList(path string) {
	if s.IsWhiteList(path) {
		return
	}
	s.WhiteList = append(s.WhiteList, path)
}

func (s *Signer) IsWhiteList(path string) bool {
	for idx := range s.WhiteList {
		if s.WhiteList[idx] == path {
			return true
		}
	}
	return false
}

func (s *Signer) Sign(req *http.Request, prefix string) {
	path := strings.TrimPrefix(req.URL.Path, prefix)
	timeStr := strconv.FormatInt(time.Now().Unix(), 10)
	toSignStr := path + timeStr + s.Token
	sign := fmt.Sprintf("%x", md5.Sum([]byte(toSignStr)))
	req.Header.Set(headerToken, sign)
	req.Header.Set(headerTime, timeStr)
}

func (s *Signer) Validate(req *http.Request) error {
	if s.IsWhiteList(req.URL.Path) {
		return nil
	}
	timeStr := req.Header.Get(headerTime)
	token := req.Header.Get(headerToken)
	path := req.URL.Path
	timestamp, err := strconv.ParseInt(timeStr, 10, 64)
	if err != nil {
		return err
	}
	n := time.Now()
	after := n.Add(time.Second * time.Duration(s.Duration)).Unix()
	before := n.Add(time.Second * (-1 * time.Duration(s.Duration))).Unix()
	if timestamp > after || timestamp < before {
		return fmt.Errorf("httpsigs time out, origin: %s, now: %v", timeStr, n)
	}
	toSignStr := path + timeStr + s.Token
	signOut := fmt.Sprintf("%x", md5.Sum([]byte(toSignStr)))
	if signOut != token {
		return fmt.Errorf("invalid http signature, path: %s, tosign: %s, sign: %s", path, toSignStr, signOut)
	}
	return nil
}
