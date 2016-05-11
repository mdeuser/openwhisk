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
)

type PackageService struct {
        client *Client
}

type SentPackagePublish struct {
        Namespace string `json:"-"`
        Name      string `json:"-"`
        Version   string `json:"version,omitempty"`
        Publish   bool   `json:"publish"`

        Annotations 	`json:"annotations,omitempty"`
        Parameters  	`json:"parameters,omitempty"`
        Bindings  bool  `json:"binding,omitempty"`
}

type SentPackageNoPublish struct {
        Namespace string `json:"-"`
        Name      string `json:"-"`
        Version   string `json:"version,omitempty"`
        Publish   bool   `json:"publish,omitempty"`

        Annotations 	`json:"annotations,omitempty"`
        Parameters  	`json:"parameters,omitempty"`
        Bindings  bool  `json:"binding,omitempty"`
}

type Package struct {
        Namespace string `json:"namespace,omitempty"`
        Name      string `json:"name,omitempty"`
        Version   string `json:"version,omitempty"`
        Publish   bool   `json:"publish"`

        Annotations 	`json:"annotations,omitempty"`
        Parameters  	`json:"parameters,omitempty"`
        Bindings  bool  `json:"binding,omitempty"`
}

type Binding struct {
        Namespace string `json:"namespace"`
        Name      string `json:"name"`
}

type BindingUpdates struct {
        Added   []Binding `json:"added,omitempty"`
        Updated []Binding `json:"added,omitempty"`
        Deleted []Binding `json:"added,omitempty"`
}

type PackageListOptions struct {
        Public bool `url:"public,omitempty"`
        Limit  int  `url:"limit,omitempty"`
        Skip   int  `url:"skip,omitempty"`
        Since  int  `url:"since,omitempty"`
        Docs   bool `url:"docs,omitempty"`
}

func (s *PackageService) List(options *PackageListOptions) ([]Package, *http.Response, error) {
        route := fmt.Sprintf("packages")
        route, err := addRouteOptions(route, options)
        if err != nil {
                return nil, nil, err
        }

        req, err := s.client.NewRequest("GET", route, nil)
        if err != nil {
                return nil, nil, err
        }

        var packages []Package
        resp, err := s.client.Do(req, &packages)
        if err != nil {
                return nil, resp, err
        }

        return packages, resp, err

}

func (s *PackageService) Get(packageName string) (*Package, *http.Response, error) {
        route := fmt.Sprintf("packages/%s", packageName)

        req, err := s.client.NewRequest("GET", route, nil)
        if err != nil {
                return nil, nil, err
        }

        p := new(Package)
        resp, err := s.client.Do(req, &p)
        if err != nil {
                return nil, resp, err
        }

        return p, resp, nil

}

func (s *PackageService) Insert(x_package *Package, shared bool, overwrite bool) (*Package, *http.Response, error) {
        route := fmt.Sprintf("packages/%s?overwrite=%t", x_package.Name, overwrite)

        var sentPackage interface{}

        if shared {
                sentPackage = SentPackagePublish{
                        Annotations: x_package.Annotations,
                        Parameters: x_package.Parameters,
                        Publish: x_package.Publish,
                }
        } else {
                sentPackage = SentPackageNoPublish{
                        Annotations: x_package.Annotations,
                        Parameters: x_package.Parameters,
                }
        }

        req, err := s.client.NewRequest("PUT", route, sentPackage)
        if err != nil {
                return nil, nil, err
        }

        p := new(Package)
        resp, err := s.client.Do(req, &p)
        if err != nil {
                return nil, resp, err
        }

        return p, resp, nil
}

func (s *PackageService) Delete(packageName string) (*http.Response, error) {
        route := fmt.Sprintf("packages/%s", packageName)

        req, err := s.client.NewRequest("DELETE", route, nil)
        if err != nil {
                return nil, err
        }

        resp, err := s.client.Do(req, nil)
        if err != nil {
                return resp, err
        }

        return resp, nil
}

func (s *PackageService) Refresh() (*BindingUpdates, *http.Response, error) {
        route := "packages/refresh"

        req, err := s.client.NewRequest("POST", route, nil)
        if err != nil {
                return nil, nil, err
        }

        updates := &BindingUpdates{}
        resp, err := s.client.Do(req, updates)
        if err != nil {
                return nil, resp, err
        }

        return updates, resp, nil
}
