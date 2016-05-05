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

package main

import (
    "fmt"
    "os"

    "../go-whisk/whisk"
    "../go-whisk-cli/commands"
    "reflect"
)

func main() {
    defer func() {
        if r := recover(); r != nil {
            fmt.Println(r)
            fmt.Println("Application exited unexpectedly")
        }
    }()

    if err := commands.Execute(); err != nil {
        if commands.IsDebug() {
            fmt.Println("Main: err type: ", reflect.TypeOf(err))
        }

        werr, isWskError := err.(whisk.WskError)  // Is the err a WskError?
        if isWskError {
            if commands.IsDebug() {
                fmt.Println("Main: got a whisk.WskError error")
            }
            os.Exit(werr.ExitCode)
        } else {
            rsperr, isRespError := err.(*whisk.ErrorResponse)
            if isRespError {
                if commands.IsDebug() {
                    fmt.Print("Main: got a whisk.ErrorResponse: code = ", rsperr.Response.StatusCode)
                }
                os.Exit(rsperr.Response.StatusCode - 256);
            } else {
                if commands.IsDebug() {
                    fmt.Println("Main: got some other error")
                }
                os.Exit(1);
            }
        }

        return
    }
}
