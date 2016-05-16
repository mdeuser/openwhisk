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

package commands

import (
    "fmt"
    "errors"

    "github.com/spf13/cobra"
    "github.com/fatih/color"

    "../../go-whisk/whisk"
)

// namespaceCmd represents the namespace command
var namespaceCmd = &cobra.Command{
    Use:   "namespace",
    Short: "work with namespaces",
}

var namespaceListCmd = &cobra.Command{
    Use:   "list",
    Short: "list available namespaces",
    SilenceUsage:   true,
    SilenceErrors:  true,
    RunE: func(cmd *cobra.Command, args []string) error {
        // add "TYPE" --> public / private

        namespaces, _, err := client.Namespaces.List()
        if err != nil {
            if IsDebug() {
                fmt.Printf("namespaceListCmd: client.Namespaces.List() error: %s\n", err)
            }
            errStr := fmt.Sprintf("Unable to obtain list of available namspaces: %s", err)
            werr := whisk.MakeWskErrorFromWskError(errors.New(errStr), err, whisk.EXITCODE_ERR_NETWORK, whisk.DISPLAY_MSG, whisk.NO_DISPLAY_USAGE)
            return werr
        }
        printList(namespaces)
        return nil
    },
}

var namespaceGetCmd = &cobra.Command{
    Use:   "get <namespace string>",
    Short: "get triggers, actions, and rules in the registry for a namespace",
    SilenceUsage:   true,
    SilenceErrors:  true,
    RunE: func(cmd *cobra.Command, args []string) error {
        var qName qualifiedName

        if len(args) < 1 {

            if IsDebug() {
                fmt.Printf("namespaceGetCmd: Invalid argument list: %s\n", cmd, args)
            }

            errMsg := "Invalid argument list.\n"
            whiskErr := whisk.MakeWskError(errors.New(errMsg), whisk.EXITCODE_ERR_GENERAL,
                whisk.DISPLAY_MSG, whisk.NO_DISPLAY_USAGE)

            return whiskErr
        }

        qName, err := parseQualifiedName(args[0])

        if err != nil {

            if IsDebug() {
                fmt.Println("namespaceGetCmd: parseQualifiedName(%s)\nerror: %s\n", args[0], err)
            }

            errMsg := fmt.Sprintf("Failed to parse qualified name: %s\n", args[0])
            whiskErr := whisk.MakeWskErrorFromWskError(errors.New(errMsg), err, whisk.EXITCODE_ERR_GENERAL,
                whisk.DISPLAY_MSG, whisk.NO_DISPLAY_USAGE)

            return whiskErr
        }

        namespace, _, err := client.Namespaces.Get(qName.packageName)
        if err != nil {
            if IsDebug() {
                fmt.Printf("namespaceGetCmd: client.Namespaces.Get(%s) error: %s\n", qName.namespace, err)
            }
            errStr := fmt.Sprintf("Unable to obtain namespace entities for namespace '%s': %s", qName.namespace, err)
            werr := whisk.MakeWskErrorFromWskError(errors.New(errStr), err, whisk.EXITCODE_ERR_NETWORK, whisk.DISPLAY_MSG, whisk.NO_DISPLAY_USAGE)
            return werr
        }

        //fmt.Printf("Entities in namespace: %s\n", boldString(namespace.Name))  // Did not work on Windows; so replaced with following two lines
        fmt.Printf("Entities in namespace: ")

        if (qName.namespace != "_") {
            color.New(color.Bold).Printf("%s\n", namespace.Name)
        } else {
            color.New(color.Bold).Printf("default\n")
        }

        printList(namespace.Contents.Packages)
        printList(namespace.Contents.Actions)
        printList(namespace.Contents.Triggers)
        printList(namespace.Contents.Rules)

        return nil
    },
}

// listCmd is a shortcut for "wsk namespace get _"
var listCmd = &cobra.Command{
    Use:   "list <namespace string>",
    Short: "list triggers, actions, and rules in the registry for a namespace",
    SilenceUsage:   true,
    SilenceErrors:  true,
    RunE:   namespaceGetCmd.RunE,
}

func init() {
    namespaceCmd.AddCommand(
        namespaceListCmd,
        namespaceGetCmd,
    )
}
