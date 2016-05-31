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
    "fmt"
    "net/http"
    "net/url"
    "errors"
    "strings"
)

type ActionService struct {
    client *Client
}

type Action struct {
    Namespace string `json:"namespace,omitempty"`
    Name      string `json:"name,omitempty"`
    Version   string `json:"version,omitempty"`
    Publish   bool   `json:"publish"`

    Exec *Exec       `json:"exec,omitempty"`
    Annotations      `json:"annotations,omitempty"`
    Parameters       `json:"parameters,omitempty"`
    BindParameters   `json:"-"`
    Limits           `json:"limits,omitempty"`
}

type SentActionPublish struct {
    Namespace string    `json:"-"`
    Version   string    `json:"-"`
    Publish   bool      `json:"publish"`

    Parameters          `json:"parameters,omitempty"`
    BindParameters          `json:"bindarameters,omitempty"`
    Exec    *Exec       `json:"exec,omitempty"`
    Annotations         `json:"annotations,omitempty"`
    Limits              `json:"-"`

    Error   string `json:"error,omitempty"`
    Code    int `json:"code,omitempty"`
}

type BindSentActionPublish struct {
    Namespace string    `json:"-"`
    Version   string    `json:"-"`
    Publish   bool      `json:"publish"`

    BindParameters      `json:"parameters,omitempty"`
    Exec    *Exec       `json:"exec,omitempty"`
    Annotations         `json:"annotations,omitempty"`
    Limits              `json:"-"`

    Error   string `json:"error,omitempty"`
    Code    int `json:"code,omitempty"`
}

type SentActionNoPublish struct {
    Namespace string `json:"-"`
    Version   string `json:"-"`
    Publish   bool   `json:"publish,omitempty"`

    Parameters      `json:"parameters,omitempty"`
    BindParameters      `json:"bindparameters,omitempty"`
    Exec    *Exec   `json:"exec,omitempty"`
    Annotations     `json:"annotations,omitempty"`
    Limits          `json:"-"`

    Error   string `json:"error,omitempty"`
    Code    int `json:"code,omitempty"`
}

type BindSentActionNoPublish struct {
    Namespace string `json:"-"`
    Version   string `json:"-"`
    Publish   bool   `json:"publish,omitempty"`

    BindParameters      `json:"parameters,omitempty"`
    Exec    *Exec   `json:"exec,omitempty"`
    Annotations     `json:"annotations,omitempty"`
    Limits          `json:"-"`

    Error   string `json:"error,omitempty"`
    Code    int `json:"code,omitempty"`
}


type Exec struct {
    Kind  string `json:"kind,omitempty"`
    Code  string `json:"code,omitempty"`
    Image string `json:"image,omitempty"`
    Init  string `json:"init,omitempty"`
    Jar   string `json:"jar,omitempty"`
    Main  string `json:"main,omitempty"`
}

type ActionListOptions struct {
    Limit           int  `url:"limit,omitempty"`
    Skip            int  `url:"skip,omitempty"`
    Docs            bool `url:"docs,omitempty"`
}

////////////////////
// Action Methods //
////////////////////

func (s *ActionService) List(packageName string, options *ActionListOptions) ([]Action, *http.Response, error) {
    var route string
    var actions []Action

    if (len(packageName) > 0) {
        packageName = strings.Replace(url.QueryEscape(packageName), "+", " ", -1)
        route = fmt.Sprintf("actions/%s/", packageName)
    } else {
        route = fmt.Sprintf("actions")
    }

    route, err := addRouteOptions(route, options)

    if err != nil {

        if IsDebug() {
            fmt.Printf("ActionService.List: addRouteOptions(%s, %s)\nerror: %s\n", route, options, err)
        }

        errMsg := fmt.Sprintf("Unable to add route options: %s\n", options)
        whiskErr := MakeWskErrorFromWskError(errors.New(errMsg), err, EXITCODE_ERR_GENERAL, DISPLAY_MSG,
            NO_DISPLAY_USAGE)

        return nil, nil, whiskErr
    }

    if IsDebug() {
        fmt.Printf("Action list route with options: %s\n", route)
    }

    req, err := s.client.NewRequest("GET", route, nil)

    if err != nil {

        if IsDebug() {
            fmt.Printf("ActionService.List: s.client.NewRequest(\"GET\", %s, nil) error: %s\n", route, err)
        }

        errMsg := fmt.Sprintf("New request failure: %s\n", err)
        whiskErr := MakeWskErrorFromWskError(errors.New(errMsg), err, EXITCODE_ERR_NETWORK, DISPLAY_MSG,
            NO_DISPLAY_USAGE)

        return nil, nil, whiskErr
    }

    resp, err := s.client.Do(req, &actions)

    if err != nil {

        if IsDebug() {
            fmt.Printf("ActionService.List: s.client.Do(%#v) error: %s\n", req, err)
        }

        errMsg := fmt.Sprintf("Request failure: %s\n", err)
        whiskErr := MakeWskErrorFromWskError(errors.New(errMsg), err, EXITCODE_ERR_NETWORK, DISPLAY_MSG,
            NO_DISPLAY_USAGE)

        return nil, resp, whiskErr
    }

    return actions, resp, err
}

func (s *ActionService) Insert(action *Action, sharedSet bool, overwrite bool) (*Action, *http.Response, error) {
    var sentAction interface{}

    action.Name = strings.Replace(url.QueryEscape(action.Name), "+", " ", -1)
    route := fmt.Sprintf("actions/%s?overwrite=%t", action.Name, overwrite)

    if sharedSet {

        if len(action.BindParameters) > 0 {
            sentAction = BindSentActionPublish{
                BindParameters: action.BindParameters,
                Exec: action.Exec,
                Publish: action.Publish,
                Annotations: action.Annotations,
            }
        } else {
            sentAction = SentActionPublish{
                Parameters: action.Parameters,
                Exec: action.Exec,
                Publish: action.Publish,
                Annotations: action.Annotations,
            }
        }
    } else {

        if len(action.BindParameters) > 0 {
            sentAction = BindSentActionNoPublish{
                BindParameters: action.BindParameters,
                Exec: action.Exec,
                Annotations: action.Annotations,
            }
        } else {
            sentAction = SentActionNoPublish{
                Parameters: action.Parameters,
                Exec: action.Exec,
                Annotations: action.Annotations,
            }
        }
    }

    if s.client.IsDebug() {
        fmt.Printf("Action insert route: %s\n", route)
    }

    req, err := s.client.NewRequest("PUT", route, sentAction)

    if err != nil {

        if IsDebug() {
            fmt.Printf("ActionService.Insert: s.client.NewRequest(\"PUT\", %s, nil) error: %s\n", route, err)
        }

        errMsg := fmt.Sprintf("New request failure: %s\n", err)
        whiskErr := MakeWskErrorFromWskError(errors.New(errMsg), err, EXITCODE_ERR_NETWORK, DISPLAY_MSG,
            NO_DISPLAY_USAGE)

        return nil, nil, whiskErr
    }

    a := new(Action)
    resp, err := s.client.Do(req, &a)

    if err != nil {

        if IsDebug() {
            fmt.Printf("ActionService.Insert: s.client.Do(%#v) error: %s\n", req, err)
        }

        errMsg := fmt.Sprintf("Request failure: %s\n", err)
        whiskErr := MakeWskErrorFromWskError(errors.New(errMsg), err, EXITCODE_ERR_NETWORK, DISPLAY_MSG,
            NO_DISPLAY_USAGE)

        return nil, resp, whiskErr
    }

    return a, resp, nil
}

func (s *ActionService) Get(actionName string) (*Action, *http.Response, error) {

    actionName = strings.Replace(url.QueryEscape(actionName), "+", " ", -1)
    route := fmt.Sprintf("actions/%s", actionName)

    req, err := s.client.NewRequest("GET", route, nil)
    if err != nil {

        if IsDebug() {
            fmt.Printf("ActionService.Get: s.client.NewRequest(\"GET\", %s, nil) error: %s\n", route, err)
        }

        errMsg := fmt.Sprintf("New request failure: %s\n", err)
        whiskErr := MakeWskErrorFromWskError(errors.New(errMsg), err, EXITCODE_ERR_NETWORK, DISPLAY_MSG,
            NO_DISPLAY_USAGE)

        return nil, nil, whiskErr
    }

    a := new(Action)
    resp, err := s.client.Do(req, &a)
    if err != nil {

        if IsDebug() {
            fmt.Printf("ActionService.Get: s.client.Do(%#v) error: %s\n", req, err)
        }

        errMsg := fmt.Sprintf("Request failure: %s\n", err)
        whiskErr := MakeWskErrorFromWskError(errors.New(errMsg), err, EXITCODE_ERR_NETWORK, DISPLAY_MSG,
            NO_DISPLAY_USAGE)

        return nil, resp, whiskErr
    }

    return a, resp, nil
}

func (s *ActionService) Delete(actionName string) (*http.Response, error) {

    actionName = strings.Replace(url.QueryEscape(actionName), "+", " ", -1)
    route := fmt.Sprintf("actions/%s", actionName)

    if s.client.IsDebug() {
        fmt.Printf("HTTP route: %s\n", route)
    }

    req, err := s.client.NewRequest("DELETE", route, nil)

    if err != nil {
        if IsDebug() {
            fmt.Printf("ActionService.Delete: s.client.NewRequest(\"DELETE\", %s, nil) error: %s\n", route, err)
        }

        errMsg := fmt.Sprintf("New request failure: %s\n", err)
        whiskErr := MakeWskErrorFromWskError(errors.New(errMsg), err, EXITCODE_ERR_NETWORK, DISPLAY_MSG,
            NO_DISPLAY_USAGE)

        return nil, whiskErr
    }

    a := new(SentActionNoPublish)
    resp, err := s.client.Do(req, a)

    if err != nil {

        if IsDebug() {
            fmt.Printf("ActionService.Delete: s.client.Do(%#v) error: %s\n", req, err)
        }

        errMsg := fmt.Sprintf("Request failure: %s\n", err)
        whiskErr := MakeWskErrorFromWskError(errors.New(errMsg), err, EXITCODE_ERR_NETWORK, DISPLAY_MSG,
            NO_DISPLAY_USAGE)

        return resp, whiskErr
    }

    return resp, nil
}

func (s *ActionService) Invoke(actionName string, payload map[string]interface{}, blocking bool) (*Activation, *http.Response, error) {

    actionName = strings.Replace(url.QueryEscape(actionName), "+", " ", -1)
    route := fmt.Sprintf("actions/%s?blocking=%t", actionName, blocking)

    if s.client.IsDebug() {
        fmt.Printf("HTTP route: %s\n", route)
    }

    req, err := s.client.NewRequest("POST", route, payload)

    if err != nil {

        if IsDebug() {
            fmt.Printf("ActionService.Invoke: s.client.NewRequest(\"POST\", %s, nil) error: %s\n", route, err)
        }

        errMsg := fmt.Sprintf("New request failure: %s\n", err)
        whiskErr := MakeWskErrorFromWskError(errors.New(errMsg), err, EXITCODE_ERR_NETWORK, DISPLAY_MSG,
            NO_DISPLAY_USAGE)

        return nil, nil, whiskErr
    }

    a := new(Activation)
    resp, err := s.client.Do(req, &a)

    if err != nil {

        if IsDebug() {
            fmt.Printf("ActionService.Invoke: s.client.Do(%#v) error: %s\n", req, err)
        }

        errMsg := fmt.Sprintf("Request failure: %s\n", err)
        whiskErr := MakeWskErrorFromWskError(errors.New(errMsg), err, EXITCODE_ERR_NETWORK, DISPLAY_MSG,
            NO_DISPLAY_USAGE)

        return nil, resp, whiskErr
    }

    return a, resp, nil
}
