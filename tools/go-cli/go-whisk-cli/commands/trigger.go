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
    "encoding/json"
    "errors"
    "fmt"
    "strings"

    "../../go-whisk/whisk"

    "github.com/spf13/cobra"
)

// triggerCmd represents the trigger command
var triggerCmd = &cobra.Command{
    Use:   "trigger",
    Short: "work with triggers",
}

var triggerFireCmd = &cobra.Command{
    Use:   "fire <name string> <payload string>",
    Short: "fire trigger event",
    SilenceUsage:   true,
    SilenceErrors:  true,
    RunE: func(cmd *cobra.Command, args []string) error {

        var err error
        var triggerName, payloadArg string
        if len(args) < 1 || len(args) > 2 {
            if IsDebug() {
                fmt.Printf("triggerFireCmd: Invalid number of arguments %d (expected 1 or 2); args: %#v\n", len(args), args)
            }
            errStr := fmt.Sprintf("Invalid number of arguments (%d) provided; 1 or 2 arguments are expected", len(args))
            werr := whisk.MakeWskError(errors.New(errStr), whisk.EXITCODE_ERR_GENERAL, whisk.DISPLAY_MSG, whisk.DISPLAY_USAGE)
            return werr
        }

        triggerName = args[0]

        payload := map[string]interface{}{}

        if len(flags.common.param) > 0 {
            parameters, err := parseParameters(flags.common.param)
            if err != nil {
                if IsDebug() {
                    fmt.Printf("triggerFireCmd: parseParameters(%#v) failed: %s\n", flags.common.param, err)
                }
                errStr := fmt.Sprintf("Invalid parameter argument '%#v': %s", flags.common.param, err)
                werr := whisk.MakeWskErrorFromWskError(errors.New(errStr), err, whisk.EXITCODE_ERR_GENERAL, whisk.DISPLAY_MSG, whisk.DISPLAY_USAGE)
                return werr
            }

            for _, param := range parameters {
                payload[param.Key] = param.Value
            }
        }

        if len(args) == 2 {
            payloadArg = args[1]
            reader := strings.NewReader(payloadArg)
            err = json.NewDecoder(reader).Decode(&payload)
            if err != nil {
                payload["payload"] = payloadArg
            } else {
                fmt.Printf("triggerFireCmd: json Decode of '%s' failed: %s\n", payloadArg, err)
                errStr := fmt.Sprintf("Invalid payload argument '%s': %s", payloadArg, err)
                werr := whisk.MakeWskError(errors.New(errStr), whisk.EXITCODE_ERR_GENERAL, whisk.DISPLAY_MSG, whisk.DISPLAY_USAGE)
                return werr
            }
        }

        _, _, err = client.Triggers.Fire(triggerName, payload)
        if err != nil {
            if IsDebug() {
                fmt.Printf("triggerFireCmd: client.Triggers.Fire(%s, %#v) failed: %s\n", triggerName, payload, err)
            }
            errStr := fmt.Sprintf("Unable to fire trigger '%s': %s", triggerName, err)
            werr := whisk.MakeWskErrorFromWskError(errors.New(errStr), err, whisk.EXITCODE_ERR_GENERAL, whisk.DISPLAY_MSG, whisk.NO_DISPLAY_USAGE)
            return werr
        }

        fmt.Println("ok: fired trigger")
        return nil
    },
}

var triggerCreateCmd = &cobra.Command{
    Use:   "create",
    Short: "create new trigger",
    SilenceUsage:   true,
    SilenceErrors:  true,
    RunE: func(cmd *cobra.Command, args []string) error {

        var err error
        if len(args) != 1 {
            if IsDebug() {
                fmt.Printf("triggerCreateCmd: Invalid number of arguments %d; args: %#v\n", len(args), args)
            }
            errStr := fmt.Sprintf("Invalid number of arguments (%d) provided; exactly one argument is expected", len(args))
            werr := whisk.MakeWskError(errors.New(errStr), whisk.EXITCODE_ERR_GENERAL, whisk.DISPLAY_MSG, whisk.DISPLAY_USAGE)
            return werr
        }

        qName, err := parseQualifiedName(args[0])

        if err != nil {

            if IsDebug() {
                fmt.Println("actionInvokeCmd: parseQualifiedName(%s)\nerror: %s\n", args[0], err)
            }

            errMsg := fmt.Sprintf("Failed to parse qualified name: %s\n", args[0])
            whiskErr := whisk.MakeWskErrorFromWskError(errors.New(errMsg), err, whisk.EXITCODE_ERR_GENERAL,
                whisk.DISPLAY_MSG, whisk.NO_DISPLAY_USAGE)

            return whiskErr
        }

        client.Namespace = qName.namespace

        // Convert the trigger's list of default parameters from a string into []KeyValue
        // The 1 or more --param arguments have all been combined into a single []string
        // e.g.   --p arg1,arg2 --p arg3,arg4   ->  [arg1, arg2, arg3, arg4]
        if IsDebug() {
            fmt.Printf("triggerCreateCmd: parsing parameters: %#v\n", flags.common.param)
        }
        parameters, err := parseParameters(flags.common.param)
        if err != nil {
            if IsDebug() {
                fmt.Printf("triggerCreateCmd: parseParameters(%#v) failed: %s\n", flags.common.param, err)
            }
            errStr := fmt.Sprintf("Invalid parameter argument '%#v': %s", flags.common.param, err)
            werr := whisk.MakeWskErrorFromWskError(errors.New(errStr), err, whisk.EXITCODE_ERR_GENERAL, whisk.DISPLAY_MSG, whisk.DISPLAY_USAGE)
            return werr
        }

        if IsDebug() {
            fmt.Printf("triggerCreateCmd: parsing annotations: %#v\n", flags.common.annotation)
        }
        annotations, err := parseAnnotations(flags.common.annotation)
        if err != nil {
            if IsDebug() {
                fmt.Printf("triggerCreateCmd: parseAnnotations(%#v) failed: %s\n", flags.common.annotation, err)
            }
            errStr := fmt.Sprintf("Invalid annotations argument value '%#v': %s", flags.common.annotation, err)
            werr := whisk.MakeWskErrorFromWskError(errors.New(errStr), err, whisk.EXITCODE_ERR_GENERAL, whisk.DISPLAY_MSG, whisk.DISPLAY_USAGE)
            return werr
        }

        trigger := &whisk.Trigger{
            Name:        qName.entityName,
            Parameters:  parameters,
            Annotations: annotations,
        }

        retTrigger, _, err := client.Triggers.Insert(trigger, false)
        if err != nil {
            if IsDebug() {
                fmt.Printf("triggerCreateCmd: client.Triggers.Insert(%+v,false) failed: %s\n", trigger, err)
            }
            errStr := fmt.Sprintf("Unable to create trigger '%s': %s", trigger.Name, err)
            werr := whisk.MakeWskErrorFromWskError(errors.New(errStr), err, whisk.EXITCODE_ERR_GENERAL, whisk.DISPLAY_MSG, whisk.NO_DISPLAY_USAGE)
            return werr
        }

        fmt.Println("ok: created trigger")
        printJSON(retTrigger)
        return nil
    },
}

var triggerUpdateCmd = &cobra.Command{
    Use:   "update",
    Short: "update existing trigger",
    SilenceUsage:   true,
    SilenceErrors:  true,
    RunE: func(cmd *cobra.Command, args []string) error {

        var err error
        if len(args) != 1 {
            if IsDebug() {
                fmt.Printf("triggerUpdateCmd: Invalid number of arguments %d; args: %#v\n", len(args), args)
            }
            errStr := fmt.Sprintf("Invalid number of arguments (%d) provided; exactly one argument is expected", len(args))
            werr := whisk.MakeWskError(errors.New(errStr), whisk.EXITCODE_ERR_GENERAL, whisk.DISPLAY_MSG, whisk.DISPLAY_USAGE)
            return werr
        }

        triggerName := args[0]

        // Convert the trigger's list of default parameters from a string into []KeyValue
        // The 1 or more --param arguments have all been combined into a single []string
        // e.g.   --p arg1,arg2 --p arg3,arg4   ->  [arg1, arg2, arg3, arg4]
        if IsDebug() {
            fmt.Printf("triggerUpdateCmd: parsing parameters: %#v\n", flags.common.param)
        }
        parameters, err := parseParameters(flags.common.param)
        if err != nil {
            if IsDebug() {
                fmt.Printf("triggerUpdateCmd: parseParameters(%#v) failed: %s\n", flags.common.param, err)
            }
            errStr := fmt.Sprintf("Invalid parameter argument '%#v': %s", flags.common.param, err)
            werr := whisk.MakeWskErrorFromWskError(errors.New(errStr), err, whisk.EXITCODE_ERR_GENERAL, whisk.DISPLAY_MSG, whisk.DISPLAY_USAGE)
            return werr
        }

        if IsDebug() {
            fmt.Printf("triggerUpdateCmd: parsing annotations: %#v\n", flags.common.annotation)
        }
        annotations, err := parseAnnotations(flags.common.annotation)
        if err != nil {
            if IsDebug() {
                fmt.Printf("triggerUpdateCmd: parseAnnotations(%#v) failed: %s\n", flags.common.annotation, err)
            }
            errStr := fmt.Sprintf("Invalid annotations argument value '%#v': %s", flags.common.annotation, err)
            werr := whisk.MakeWskErrorFromWskError(errors.New(errStr), err, whisk.EXITCODE_ERR_GENERAL, whisk.DISPLAY_MSG, whisk.DISPLAY_USAGE)
            return werr
        }

        trigger := &whisk.Trigger{
            Name:        triggerName,
            Parameters:  parameters,
            Annotations: annotations,
        }

        retTrigger, _, err := client.Triggers.Insert(trigger, true)
        if err != nil {
            if IsDebug() {
                fmt.Printf("triggerUpdateCmd: client.Triggers.Insert(%+v,true) failed: %s\n", trigger, err)
            }
            errStr := fmt.Sprintf("Unable to update trigger '%s': %s", trigger.Name, err)
            werr := whisk.MakeWskErrorFromWskError(errors.New(errStr), err, whisk.EXITCODE_ERR_GENERAL, whisk.DISPLAY_MSG, whisk.NO_DISPLAY_USAGE)
            return werr
        }

        fmt.Println("ok: updated trigger")
        printJSON(retTrigger)
        return nil
    },
}

var triggerGetCmd = &cobra.Command{
    Use:   "get <name string>",
    Short: "get trigger",
    SilenceUsage:   true,
    SilenceErrors:  true,
    RunE: func(cmd *cobra.Command, args []string) error {
        var err error
        if len(args) != 1 {
            if IsDebug() {
                fmt.Printf("triggerGetCmd: Invalid number of arguments %d; args: %#v\n", len(args), args)
            }
            errStr := fmt.Sprintf("Invalid number of arguments (%d) provided; exactly one argument is expected", len(args))
            werr := whisk.MakeWskError(errors.New(errStr), whisk.EXITCODE_ERR_GENERAL, whisk.DISPLAY_MSG, whisk.DISPLAY_USAGE)
            return werr
        }

        triggerName := args[0]

        retTrigger, _, err := client.Triggers.Get(triggerName)
        if err != nil {
            if IsDebug() {
                fmt.Printf("triggerGetCmd: client.Triggers.Get(%s) failed: %s\n", triggerName, err)
            }
            errStr := fmt.Sprintf("Unable to get trigger '%s': %s", triggerName, err)
            werr := whisk.MakeWskErrorFromWskError(errors.New(errStr), err, whisk.EXITCODE_ERR_GENERAL, whisk.DISPLAY_MSG, whisk.NO_DISPLAY_USAGE)
            return werr
        }
        fmt.Println("ok: got trigger ", triggerName)
        printJSON(retTrigger)
        return nil
    },
}

var triggerDeleteCmd = &cobra.Command{
    Use:   "delete <name string>",
    Short: "delete trigger",
    SilenceUsage:   true,
    SilenceErrors:  true,
    RunE: func(cmd *cobra.Command, args []string) error {
        var err error
        if len(args) != 1 {
            if IsDebug() {
                fmt.Printf("triggerDeleteCmd: Invalid number of arguments %d; args: %#v\n", len(args), args)
            }
            errStr := fmt.Sprintf("Invalid number of arguments (%d) provided; exactly one argument is expected", len(args))
            werr := whisk.MakeWskError(errors.New(errStr), whisk.EXITCODE_ERR_GENERAL, whisk.DISPLAY_MSG, whisk.DISPLAY_USAGE)
            return werr
        }

        qName, err := parseQualifiedName(args[0])

        if err != nil {

            if IsDebug() {
                fmt.Println("actionInvokeCmd: parseQualifiedName(%s)\nerror: %s\n", args[0], err)
            }

            errMsg := fmt.Sprintf("Failed to parse qualified name: %s\n", args[0])
            whiskErr := whisk.MakeWskErrorFromWskError(errors.New(errMsg), err, whisk.EXITCODE_ERR_GENERAL,
                whisk.DISPLAY_MSG, whisk.NO_DISPLAY_USAGE)

            return whiskErr
        }

        client.Namespace = qName.namespace

        _, err = client.Triggers.Delete(qName.entityName)
        if err != nil {
            if IsDebug() {
                fmt.Printf("triggerDeleteCmd: client.Triggers.Delete(%s) failed: %s\n", qName.entityName, err)
            }
            errStr := fmt.Sprintf("Unable to delete trigger '%s': %s", qName.entityName, err)
            werr := whisk.MakeWskErrorFromWskError(errors.New(errStr), err, whisk.EXITCODE_ERR_GENERAL, whisk.DISPLAY_MSG, whisk.NO_DISPLAY_USAGE)
            return werr
        }
        fmt.Println("ok: deleted trigger ", qName.entityName)
        return nil
    },
}

var triggerListCmd = &cobra.Command{
    Use:   "list <namespace string>",
    Short: "list all triggers",
    SilenceUsage:   true,
    SilenceErrors:  true,
    RunE: func(cmd *cobra.Command, args []string) error {
        var err error
        qName := qualifiedName{}
        if len(args) == 1 {
            qName, err = parseQualifiedName(args[0])
            if err != nil {
                if IsDebug() {
                    fmt.Printf("triggerListCmd: parseQualifiedName(%#v) error: %s\n", args[0], err)
                }
                errStr := fmt.Sprintf("'%s' is not a valid qualified name: %s", args[0], err)
                werr := whisk.MakeWskErrorFromWskError(errors.New(errStr), err, whisk.EXITCODE_ERR_GENERAL, whisk.DISPLAY_MSG, whisk.DISPLAY_USAGE)
                return werr
            }
            ns := qName.namespace
            if len(ns) == 0 {
                if IsDebug() {
                    fmt.Printf("triggerListCmd: Namespace is missing from '%s'\n", args[0])
                }
                errStr := fmt.Sprintf("No valid namespace detected.  Run 'wsk property set --namespace' or ensure the name argument is preceded by a \"/\"")
                werr := whisk.MakeWskError(errors.New(errStr), whisk.EXITCODE_ERR_GENERAL, whisk.DISPLAY_MSG, whisk.DISPLAY_USAGE)
                return werr
            }

            client.Namespace = ns
            if IsDebug() {
                fmt.Printf("triggerListCmd: Using namespace '%s' from argument '%s''\n", ns, args[0])
            }

            if pkg := qName.packageName; len(pkg) > 0 {
                // todo :: scope call to package
            }
        }

        options := &whisk.TriggerListOptions{
            Skip:  flags.common.skip,
            Limit: flags.common.limit,
        }
        triggers, _, err := client.Triggers.List(options)
        if err != nil {
            if IsDebug() {
                fmt.Printf("triggerListCmd: client.Triggers.List(%#v) for namespace '%s' failed: %s\n", options, client.Namespace, err)
            }
            errStr := fmt.Sprintf("Unable to obtain the trigger list for namespace '%s': %s", client.Namespace, err)
            werr := whisk.MakeWskErrorFromWskError(errors.New(errStr), err, whisk.EXITCODE_ERR_GENERAL, whisk.DISPLAY_MSG, whisk.NO_DISPLAY_USAGE)
            return werr
        }
        fmt.Println(triggers)
        printJSON(triggers)
        return nil
    },
}

func init() {

    triggerCreateCmd.Flags().StringSliceVarP(&flags.common.annotation, "annotation", "a", []string{}, "annotations")
    triggerCreateCmd.Flags().StringSliceVarP(&flags.common.param, "param", "p", []string{}, "default parameters")
    triggerCreateCmd.Flags().StringVar(&flags.common.shared, "shared", "", "shared action (yes = shared, no[default] = private)")

    triggerUpdateCmd.Flags().StringSliceVarP(&flags.common.annotation, "annotation", "a", []string{}, "annotations")
    triggerUpdateCmd.Flags().StringSliceVarP(&flags.common.param, "param", "p", []string{}, "default parameters")
    triggerUpdateCmd.Flags().StringVar(&flags.common.shared, "shared", "", "shared action (yes = shared, no[default] = private)")

    triggerFireCmd.Flags().StringSliceVarP(&flags.common.param, "param", "p", []string{}, "default parameters")

    triggerListCmd.Flags().IntVarP(&flags.common.skip, "skip", "s", 0, "skip this many entities from the head of the collection")
    triggerListCmd.Flags().IntVarP(&flags.common.limit, "limit", "l", 0, "only return this many entities from the collection")

    triggerCmd.AddCommand(
        triggerFireCmd,
        triggerCreateCmd,
        triggerUpdateCmd,
        triggerGetCmd,
        triggerDeleteCmd,
        triggerListCmd,
    )

}
