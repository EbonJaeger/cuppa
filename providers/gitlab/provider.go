//
// Copyright 2016-2018 Bryan T. Meyers <bmeyers@datadrake.com>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//

package gitlab

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"regexp"
	"strings"

	"github.com/DataDrake/cuppa/results"
)

const (
	// SourceFormat is the format string for GitLab release tarballs
	SourceFormat = "https://gitlab.com/%s/-/archive/%s/%s.tar.gz"

	// TagsEndpoint is the API endpoint URL for GitLab project tags
	TagsEndpoint = "https://gitlab.com/api/v4/projects/%s/repository/tags"
)

// SourceRegex is the regex for GitLab sources
var SourceRegex = regexp.MustCompile("gitlab.com/([^/]+/[^/.]+)")

// VersionRegex is used to parse GitLab version numbers
var VersionRegex = regexp.MustCompile("(?:\\d+\\.)*\\d+\\w*")

// Provider is the upstream provider interface for GitLab
type Provider struct{}

// Latest finds the newest release for a GitLab package
func (c Provider) Latest(name string) (r *results.Result, s results.Status) {
	rs, s := c.Releases(name)
	if s != results.OK {
		return
	}
	r = rs.Last()
	return
}

// Match checks to see if this provider can handle this kind of query
func (c Provider) Match(query string) string {
	sm := SourceRegex.FindStringSubmatch(query)
	if len(sm) != 2 {
		return ""
	}
	return sm[1]
}

// Name gives the name of this provider
func (c Provider) Name() string {
	return "GitLab"
}

// Releases finds all matching releases for a GitLab package
func (c Provider) Releases(name string) (rs *results.ResultSet, s results.Status) {
	// Query the API
	encoded := strings.Replace(name, "/", "%2f", 1)
	resp, err := http.Get(fmt.Sprintf(TagsEndpoint, encoded))
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		s = results.Unavailable
		return
	}

	defer resp.Body.Close()
	// Translate Status Code
	switch resp.StatusCode {
	case 200:
		s = results.OK
	case 404:
		s = results.NotFound
	default:
		s = results.Unavailable
	}

	// Fail if not OK
	if s != results.OK {
		return
	}

	dec := json.NewDecoder(resp.Body)
	keys := make([]Tag, 0)
	err = dec.Decode(&keys)
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		s = results.Unavailable
		return
	}

	tags := &Tags{keys}
	rs = tags.Convert(name)
	return
}
