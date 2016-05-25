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
    "errors"
    "fmt"
    "net/http"

    "../../go-whisk/whisk"

    //"github.com/fatih/color"
    "github.com/spf13/cobra"
)

var packageCmd = &cobra.Command{
    Use:   "package",
    Short: "work with packages",
}

/*
bind parameters to the package

Usage:
  wsk package bind <package string> <name string> [flags]

Flags:
  -a, --annotation value   annotations (default [])
  -p, --param value        default parameters (default [])

Global Flags:
      --apihost string      whisk API host
      --apiversion string   whisk API version
  -u, --auth string         authorization key
  -d, --debug               debug level output
  -v, --verbose             verbose output

Request URL
PUT https://openwhisk.ng.bluemix.net/api/v1/namespaces/<namespace>/packages/<bindingname>

payload:
{
  "binding": {
    "namespace": "<pkgnamespace>",
    "name": "<pkgname>"
  },
  "annotations": [
    {"value": "abv1", "key": "ab1"}
  ],
  "parameters": [
    {"value": "pbv1", "key": "pb1"}
  ],
  "publish": false
}

*/
var packageBindCmd = &cobra.Command{
    Use:   "bind <package string> <name string>",
    Short: "bind parameters to a package",
    SilenceUsage:   true,
    SilenceErrors:  true,
    RunE: func(cmd *cobra.Command, args []string) error {
        var err error
        if len(args) != 2 {
            if IsDebug() {
                fmt.Printf("packageBindCmd: Invalid number of arguments %d; args: %#v\n", len(args), args)
            }
            errStr := fmt.Sprintf("Invalid number of arguments (%d) provided; either the package name or the binding name is missing", len(args))
            werr := whisk.MakeWskError(errors.New(errStr), whisk.EXITCODE_ERR_GENERAL, whisk.DISPLAY_MSG, whisk.DISPLAY_USAGE)
            return werr
        }

        packageName := args[0]
        pkgQName, err := parseQualifiedName(packageName)
        if err != nil {
            if IsDebug() {
                fmt.Println("packageBindCmd: parseQualifiedName(%s)\nerror: %s\n", packageName, err)
            }

            errMsg := fmt.Sprintf("Failed to parse qualified name: %s\n",packageName)
            werr := whisk.MakeWskErrorFromWskError(errors.New(errMsg), err, whisk.EXITCODE_ERR_GENERAL,
                whisk.DISPLAY_MSG, whisk.NO_DISPLAY_USAGE)

            return werr
        }

        bindingName := args[1]
        bindQName, err := parseQualifiedName(bindingName)
        if err != nil {
            if IsDebug() {
                fmt.Println("packageBindCmd: parseQualifiedName(%s)\nerror: %s\n", bindingName, err)
            }

            errMsg := fmt.Sprintf("Failed to parse qualified name: %s\n", bindingName)
            werr := whisk.MakeWskErrorFromWskError(errors.New(errMsg), err, whisk.EXITCODE_ERR_GENERAL,
                whisk.DISPLAY_MSG, whisk.NO_DISPLAY_USAGE)

            return werr
        }

        // Convert the binding's list of default parameters from a string into []KeyValue
        // The 1 or more --param arguments have all been combined into a single []string
        // e.g.   --p arg1,arg2 --p arg3,arg4   ->  [arg1, arg2, arg3, arg4]
        if IsDebug() {
            fmt.Printf("packageBindCmd: parsing parameters: %#v\n", flags.common.param)
        }
        parameters, err := parseParameters(flags.common.param)
        if err != nil {
            if IsDebug() {
                fmt.Printf("packageBindCmd: parseParameters(%#v) failed: %s\n", flags.common.param, err)
            }
            errStr := fmt.Sprintf("Invalid parameter argument '%#v': %s", flags.common.param, err)
            werr := whisk.MakeWskErrorFromWskError(errors.New(errStr), err, whisk.EXITCODE_ERR_GENERAL, whisk.DISPLAY_MSG, whisk.DISPLAY_USAGE)
            return werr
        }

        // Convert the binding's list of default annotations from a string into []KeyValue
        // The 1 or more --annotation arguments have all been combined into a single []string
        // e.g.   --a arg1,arg2 --a arg3,arg4   ->  [arg1, arg2, arg3, arg4]
        if IsDebug() {
            fmt.Printf("packageBindCmd: parsing annotations: %#v\n", flags.common.annotation)
        }
        annotations, err := parseAnnotations(flags.common.annotation)
        if err != nil {
            if IsDebug() {
                fmt.Printf("packageBindCmd: parseParameters(%#v) failed: %s\n", flags.common.annotation, err)
            }
            errStr := fmt.Sprintf("Invalid parameter argument '%#v': %s", flags.common.annotation, err)
            werr := whisk.MakeWskErrorFromWskError(errors.New(errStr), err, whisk.EXITCODE_ERR_GENERAL, whisk.DISPLAY_MSG, whisk.DISPLAY_USAGE)
            return werr
        }


        //
        // parsedBindingArg := strings.Split(bindingArg, ":")
        // bindingName := parsedBindingArg[0]
        // var bindingNamespace string
        // if len(parsedBindingArg) == 1 {
        // 	bindingNamespace = client.Config.Namespace
        // } else if len(parsedBindingArg) == 2 {
        // 	bindingNamespace = parsedBindingArg[1]
        // } else {
        // 	err = fmt.Errorf("Invalid binding argument %s", bindingArg)
        // 	fmt.Println(err)
        // 	return
        // }
        //

        binding := whisk.Binding{
            Name:      pkgQName.entityName,
            Namespace: pkgQName.namespace,
        }

        p := &whisk.BindingPackage{
            Name:        bindQName.entityName,
            Annotations: annotations,
            Parameters:  parameters,
            Binding:     binding,
        }

        _,  _, err = client.Packages.Insert(p, false)
        if err != nil {
            if IsDebug() {
                fmt.Printf("packageBindCmd: client.Packages.Insert(%#v, false, false) failed: %s\n", p, err)
            }
            errStr := fmt.Sprintf("Binding creation failed: %s", err)
            werr := whisk.MakeWskErrorFromWskError(errors.New(errStr), err, whisk.EXITCODE_ERR_GENERAL, whisk.DISPLAY_MSG, whisk.NO_DISPLAY_USAGE)
            return werr
        }

        fmt.Printf("ok: created binding %s\n", bindingName)
        return nil
    },
}

var packageCreateCmd = &cobra.Command{
    Use:   "create <name string>",
    Short: "create a new package",

    Run: func(cmd *cobra.Command, args []string) {
        var err error
        var shared, sharedSet bool

        if len(args) != 1 {
            err = errors.New("Invalid argument")
            fmt.Println(err)
            return
        }

        qName, err := parseQualifiedName(args[0])
        if err != nil {
            fmt.Printf("error: %s", err)
            return
        }

        client.Namespace = qName.namespace

        if (flags.common.shared == "yes") {
            shared = true
            sharedSet = true
        } else if (flags.common.shared == "no") {
            shared = false
            sharedSet = true
        } else {
            sharedSet = false
        }

        if IsDebug() {
            fmt.Printf("packageCreateCmd: raw parameters: %#v\n", flags.common.param)
        }
        parameters, err := parseParameters(flags.common.param)
        if err != nil {
            fmt.Println(err)
            return
        }

        annotations, err := parseAnnotations(flags.common.annotation)
        if err != nil {
            fmt.Println(err)
            return
        }

        var p whisk.PackageInterface
        if sharedSet {
            p = &whisk.SentPackagePublish{
                Name:        qName.entityName,
                Namespace:   qName.namespace,
                Publish:     shared,
                Annotations: annotations,
                Parameters:  parameters,
            }
        } else {
            p = &whisk.SentPackageNoPublish{
                Name:        qName.entityName,
                Namespace:   qName.namespace,
                Publish:     shared,
                Annotations: annotations,
                Parameters:  parameters,
            }
        }

        p, _, err = client.Packages.Insert(p, false)
        if err != nil {
            fmt.Println(err)
            return
        }

        //fmt.Printf("%s created package %s\n", color.GreenString("ok:"), boldString(qName.entityName))
        fmt.Printf("ok: created package %s\n", qName.entityName)
    },
}

/*
usage: wsk package update [-h] [-u AUTH] [-a ANNOTATION ANNOTATION]
                          [-p PARAM PARAM] [--shared [{yes,no}]]
                          name

positional arguments:
  name                  the name of the package

optional arguments:
  -h, --help            show this help message and exit
  -u AUTH, --auth AUTH  authorization key
  -a ANNOTATION ANNOTATION, --annotation ANNOTATION ANNOTATION
                        annotations
  -p PARAM PARAM, --param PARAM PARAM
                        default parameters
  --shared [{yes,no}]   shared action (default: private)

  UPDATE:
        If --shared is present, published is true. Otherwise, published is false.

PUT https://172.17.0.1/api/v1/namespaces/_/packages/slack?overwrite=true: 400  []
    https://172.17.0.1/api/v1/namespaces/_/packages/slack?overwrite=true

Request URL

https://raw.githubusercontent.com/api/v1/namespaces/_/packages/slack?overwrite=true

payload:
        {"name":"slack","publish":true,"annotations":[],"parameters":[],"binding":false}


 */
var packageUpdateCmd = &cobra.Command{
    Use:   "update <name string>",
    Short: "update an existing package",

    Run: func(cmd *cobra.Command, args []string) {
        var err error
        var shared, sharedSet bool

        if len(args) < 1 {
            err = errors.New("Invalid argument")
            fmt.Println(err)
            return
        }

        qName, err := parseQualifiedName(args[0])
        if err != nil {
            fmt.Printf("error: %s", err)
            return
        }

        client.Namespace = qName.namespace

        if (flags.common.shared == "yes") {
            shared = true
            sharedSet = true
        } else if (flags.common.shared == "no") {
            shared = false
            sharedSet = true
        } else {
            sharedSet = false
        }

        parameters, err := parseParameters(flags.common.param)
        if err != nil {
            fmt.Println(err)
            return
        }

        annotations, err := parseAnnotations(flags.common.annotation)
        if err != nil {
            fmt.Println(err)
            return
        }

        var p whisk.PackageInterface
        if sharedSet {
            p = &whisk.SentPackagePublish{
                Name:        qName.entityName,
                Namespace:   qName.namespace,
                Publish:     shared,
                Annotations: annotations,
                Parameters:  parameters,
            }
        } else {
            p = &whisk.SentPackageNoPublish{
                Name:        qName.entityName,
                Namespace:   qName.namespace,
                Publish:     shared,
                Annotations: annotations,
                Parameters:  parameters,
            }
        }

        p, _, err = client.Packages.Insert(p, true)
        if err != nil {
            fmt.Println(err)
            return
        }

        //fmt.Printf("%s updated package %s\n", color.GreenString("ok:"), boldString(qName.entityName))
        fmt.Printf("ok: updated package %s\n",qName.entityName)
    },
}

var packageGetCmd = &cobra.Command{
    Use:   "get <name string>",
    Short: "get package",

    Run: func(cmd *cobra.Command, args []string) {
        var err error
        if len(args) != 1 {
            err = errors.New("Invalid argument")
            fmt.Println(err)
            return
        }

        qName, err := parseQualifiedName(args[0])
        if err != nil {
            fmt.Printf("error: %s", err)
            return
        }

        client.Namespace = qName.namespace

        xPackage, _, err := client.Packages.Get(qName.entityName)
        if err != nil {
            fmt.Println(err)
            return
        }

        if flags.common.summary {
            fmt.Printf("%s /%s/%s\n", boldString("package"), xPackage.Namespace, xPackage.Name)
        } else {
            //fmt.Printf("%s got package %s\n", color.GreenString("ok:"), boldString(qName.entityName))
            fmt.Printf("ok: got package %s\n", qName.entityName)
            //printJSON(xPackage)
            printJsonNoColor(xPackage)
        }
    },
}

var packageDeleteCmd = &cobra.Command{
    Use:   "delete <name string>",
    Short: "delete package",

    Run: func(cmd *cobra.Command, args []string) {
        var err error
        if len(args) != 1 {
            err = errors.New("Invalid argument")
            fmt.Println(err)
            return
        }

        qName, err := parseQualifiedName(args[0])
        if err != nil {
            fmt.Printf("error: %s", err)
            return
        }

        client.Namespace = qName.namespace

        _, err = client.Packages.Delete(qName.entityName)
        if err != nil {
            fmt.Println(err)
            return
        }

        //fmt.Printf("%s deleted package %s\n", color.GreenString("ok:"), boldString(qName.entityName))
        fmt.Printf("ok: deleted package %s\n", qName.entityName)
    },
}

var packageListCmd = &cobra.Command{
    Use:   "list <namespace string>",
    Short: "list all packages",

    Run: func(cmd *cobra.Command, args []string) {
        var err error
        var shared bool

        qName := qualifiedName{}
        if len(args) == 1 {
            qName, err = parseQualifiedName(args[0])
            if err != nil {
                fmt.Printf("error: %s", err)
                return
            }
            ns := qName.namespace
            if len(ns) == 0 {
                err = errors.New("No valid namespace detected.  Make sure that namespace argument is preceded by a \"/\"")
                fmt.Printf("error: %s\n", err)
                return
            }

            client.Namespace = ns
        }

        if (flags.common.shared == "yes") {
            shared = true
        } else  {
            shared = false
        }

        options := &whisk.PackageListOptions{
            Skip:   flags.common.skip,
            Limit:  flags.common.limit,
            Public: shared,
            Docs:   flags.common.full,
        }

        packages, _, err := client.Packages.List(options)
        if err != nil {
            fmt.Println(err)
            return
        }

        printList(packages)
    },
}

var packageRefreshCmd = &cobra.Command{
    Use:   "refresh <namespace string>",
    Short: "refresh package bindings",

    Run: func(cmd *cobra.Command, args []string) {
        var err error

        if len(args) == 1 {
            namespace := args[0]
            currentNamespace := client.Config.Namespace
            client.Config.Namespace = namespace
            defer func() {
                client.Config.Namespace = currentNamespace
            }()
        }

        updates, resp, err := client.Packages.Refresh()
        if err != nil {
            fmt.Println(err)
            return
        }

        switch resp.StatusCode {
        case http.StatusOK:
            fmt.Printf("\n%s refreshed successfully\n", client.Config.Namespace)

            if len(updates.Added) > 0 {
                fmt.Println("created bindings:")
                printJSON(updates.Added)
            } else {
                fmt.Println("no bindings created")
            }

            if len(updates.Updated) > 0 {
                fmt.Println("updated bindings:")
                printJSON(updates.Updated)
            } else {
                fmt.Println("no bindings updated")
            }

            if len(updates.Deleted) > 0 {
                fmt.Println("deleted bindings:")
                printJSON(updates.Deleted)
            } else {
                fmt.Println("no bindings deleted")
            }

        case http.StatusNotImplemented:
            fmt.Println("error: This feature is not implemented in the targeted deployment")
            return
        default:
            fmt.Println("error: ", resp.Status)
            return
        }

    },
}

func init() {

    packageCreateCmd.Flags().StringSliceVarP(&flags.common.annotation, "annotation", "a", []string{}, "annotations")
    packageCreateCmd.Flags().StringSliceVarP(&flags.common.param, "param", "p", []string{}, "default parameters")
    packageCreateCmd.Flags().StringVarP(&flags.xPackage.serviceGUID, "service_guid", "s", "", "a unique identifier of the service")
    packageCreateCmd.Flags().StringVar(&flags.common.shared, "shared", "" , "shared action (default: private)")

    packageUpdateCmd.Flags().StringSliceVarP(&flags.common.annotation, "annotation", "a", []string{}, "annotations")
    packageUpdateCmd.Flags().StringSliceVarP(&flags.common.param, "param", "p", []string{}, "default parameters")
    packageUpdateCmd.Flags().StringVarP(&flags.xPackage.serviceGUID, "service_guid", "s", "", "a unique identifier of the service")
    packageUpdateCmd.Flags().StringVar(&flags.common.shared, "shared", "", "shared action (default: private)")

    packageGetCmd.Flags().BoolVarP(&flags.common.summary, "summary", "s", false, "summarize entity details")

    packageBindCmd.Flags().StringSliceVarP(&flags.common.annotation, "annotation", "a", []string{}, "annotations")
    packageBindCmd.Flags().StringSliceVarP(&flags.common.param, "param", "p", []string{}, "default parameters")

    packageListCmd.Flags().StringVar(&flags.common.shared, "shared", "", "include publicly shared entities in the result")
    packageListCmd.Flags().IntVarP(&flags.common.skip, "skip", "s", 0, "skip this many entities from the head of the collection")
    packageListCmd.Flags().IntVarP(&flags.common.limit, "limit", "l", 0, "only return this many entities from the collection")
    packageListCmd.Flags().BoolVar(&flags.common.full, "full", false, "include full entity description")

    packageCmd.AddCommand(
        packageBindCmd,
        packageCreateCmd,
        packageUpdateCmd,
        packageGetCmd,
        packageDeleteCmd,
        packageListCmd,
        packageRefreshCmd,
    )
}
