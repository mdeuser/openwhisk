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
            werr := whisk.MakeWskError(errors.New(errStr), whisk.EXITCODE_ERR_NETWORK, whisk.DISPLAY_MSG, whisk.NO_DISPLAY_USAGE) //FIXME MWD exitCode
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
        var nsName string
        if len(args) > 0 {
            nsName = args[0]
        }

        namespace, _, err := client.Namespaces.Get(nsName)
        if err != nil {
            if IsDebug() {
                fmt.Printf("namespaceGetCmd: client.Namespaces.Get(%s) error: %s\n", nsName, err)
            }
            errStr := fmt.Sprintf("Unable to obtain namespace entities for namespace '%s': %s", nsName, err)
            werr := whisk.MakeWskError(errors.New(errStr), whisk.EXITCODE_ERR_NETWORK, whisk.DISPLAY_MSG, whisk.NO_DISPLAY_USAGE) //FIXME MWD exitCode
            return werr
        }

        //fmt.Printf("Entities in namespace: %s\n", boldString(namespace.Name))  // Did not work on Windows; so replaced with following two lines
        fmt.Printf("Entities in namespace: ")
        color.New(color.Bold).Printf("%s\n", namespace.Name)
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
