/*
Copyright 2015-2016 IBM Corporation

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
 */

package whisk

import (
    "net/http"
    "net/url"
    "fmt"
    "errors"
)

type Info struct {
    Whisk   string `json:"whisk,omitempty"`
    Version string `json:"version,omitempty"`
    Build   string `json:"build,omitempty"`
    BuildNo string `json:"buildno,omitempty"`
}

type InfoService struct {
    client *Client
}

func (s *InfoService) Get() (*Info, *http.Response, error) {
    // make a request to c.BaseURL / v1

    ref, err := url.Parse(s.client.Config.Version)
    if err != nil {
        if IsDebug() {
            fmt.Printf("InfoService.Get: url.Parse error - URL '%s'; err '%s'\n", s.client.Config.Version, err)
        }
        errStr := fmt.Sprintf("Unable to URL parse '%s'; error: %s", s.client.Config.Version, err)
        werr := MakeWskError(errors.New(errStr), EXITCODE_ERR_GENERAL, DISPLAY_MSG, NO_DISPLAY_USAGE)
        return nil, nil, werr
    }

    u := s.client.BaseURL.ResolveReference(ref)

    req, err := http.NewRequest("GET", u.String(), nil)
    if err != nil {
        if IsDebug() {
            fmt.Printf("InfoService.Get: http.NewRequest error - URL GET '%s'; err '%s'\n", u.String(), err)
        }
        errStr := fmt.Sprintf("Unable to create HTTP request for GET '%s'; error: %s", u.String(), err)
        werr := MakeWskError(errors.New(errStr), EXITCODE_ERR_GENERAL, DISPLAY_MSG, NO_DISPLAY_USAGE)
        return nil, nil, werr
    }

    if IsDebug() {
        fmt.Printf("InfoService.Get: Sending HTTP URL '%s'; req %#v\n", req.URL.String(), req)
    }

    info := new(Info)
    resp, err := s.client.Do(req, &info)
    if err != nil {
        if IsDebug() {
            fmt.Printf("InfoService.Get: s.client.Do() error - HTTP req %s; error '%s'\n", req.URL.String(), err)
        }
        errStr := fmt.Sprintf("HTTP GET request failure '%s'; error %s", req.URL.String(), err)
        werr := MakeWskError(errors.New(errStr), EXITCODE_ERR_NETWORK, DISPLAY_MSG, NO_DISPLAY_USAGE)
        return nil, nil, werr
    }

    return info, resp, nil
}
