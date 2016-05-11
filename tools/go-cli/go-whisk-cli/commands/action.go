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
    "archive/tar"
    "compress/gzip"
    "encoding/base64"
    "encoding/json"
    "errors"
    "fmt"
    "io"
    "io/ioutil"
    "os"
    "path/filepath"
    "regexp"
    "strings"

    "../../go-whisk/whisk"

    "github.com/fatih/color"
    "github.com/spf13/cobra"
)

//////////////
// Commands //
//////////////

var actionCmd = &cobra.Command{
    Use:   "action",
    Short: "work with actions",
}

/*
usage: wsk action update [-h] [-u AUTH] [--docker] [--copy] [--sequence]
                     [--lib LIB] [--shared [{yes,no}]]
                     [-a ANNOTATION ANNOTATION] [-p PARAM PARAM]
                     [-t TIMEOUT] [-m MEMORY]
                     name [artifact]

positional arguments:
name                  the name of the action
artifact              artifact (e.g., file name) containing action
                    definition

optional arguments:
-h, --help            show this help message and exit
-u AUTH, --auth AUTH  authorization key
--docker              treat artifact as docker image path on dockerhub
--copy                treat artifact as the name of an existing action
--sequence            treat artifact as comma separated sequence of actions
                    to invoke
--lib LIB             add library to artifact (must be a gzipped tar file)
--shared [{yes,no}]   shared action (default: private)
-a ANNOTATION ANNOTATION, --annotation ANNOTATION ANNOTATION
                    annotations
-p PARAM PARAM, --param PARAM PARAM
                    default parameters
-t TIMEOUT, --timeout TIMEOUT
                    the timeout limit in milliseconds when the action will
                    be terminated
-m MEMORY, --memory MEMORY
                    the memory limit in MB of the container that runs the
                    action
*/
var actionCreateCmd = &cobra.Command{
    Use:   "create <name string> <artifact string>",
    Short: "create a new action",
    SilenceUsage:   true,
    SilenceErrors:  true,
    RunE: func(cmd *cobra.Command, args []string) error {
        action, sharedSet, err := parseAction(cmd, args)

        if err != nil {
            return err
        }

        action, _, err = client.Actions.Insert(action, sharedSet, false)

        if err != nil {
            return err
        }

        fmt.Printf("%s created action %s", color.GreenString("ok:"), boldString(action.Name))

        return nil
    },
}

/*
usage: wsk action update [-h] [-u AUTH] [--docker] [--copy] [--sequence]
[--lib LIB] [--shared [{yes,no}]]
[-a ANNOTATION ANNOTATION] [-p PARAM PARAM]
[-t TIMEOUT] [-m MEMORY]
name [artifact]

positional arguments:
name                  the name of the action
artifact              artifact (e.g., file name) containing action
definition

optional arguments:
-h, --help            show this help message and exit
-u AUTH, --auth AUTH  authorization key
--docker              treat artifact as docker image path on dockerhub
--copy                treat artifact as the name of an existing action
--sequence            treat artifact as comma separated sequence of actions
to invoke
--lib LIB             add library to artifact (must be a gzipped tar file)
--shared [{yes,no}]   shared action (default: private)
-a ANNOTATION ANNOTATION, --annotation ANNOTATION ANNOTATION
annotations
-p PARAM PARAM, --param PARAM PARAM
default parameters
-t TIMEOUT, --timeout TIMEOUT
the timeout limit in milliseconds when the action will
be terminated
-m MEMORY, --memory MEMORY
the memory limit in MB of the container that runs the
action
*/
var actionUpdateCmd = &cobra.Command{
    Use:   "update <name string> <artifact string>",
    Short: "update an existing action",
    SilenceUsage:   true,
    SilenceErrors:  true,
    RunE: func(cmd *cobra.Command, args []string) error {
        action, sharedSet, err := parseAction(cmd, args)

        if err != nil {
            return err
        }

        action, _, err = client.Actions.Insert(action, sharedSet, true)

        if err != nil {
            return err
        }

        fmt.Printf("%s updated action %s", color.GreenString("ok:"), boldString(action.Name))

        return nil
    },
}

var actionInvokeCmd = &cobra.Command{
    Use:   "invoke <name string> <payload string>",
    Short: "invoke action",
    SilenceUsage:   true,
    SilenceErrors:  true,
    RunE: func(cmd *cobra.Command, args []string) error {
        var err error
        var payloadArg string

        if len(args) < 1 || len(args) > 2 {
            err = errors.New("Invalid argument list")
            err := whisk.MakeWskError(err, whisk.EXITCODE_ERR_GENERAL, whisk.DISPLAY_MSG, whisk.NO_DISPLAY_USAGE)
            return err
        }

        qName, err := parseQualifiedName(args[0])
        if err != nil {
            err := whisk.MakeWskError(err, whisk.EXITCODE_ERR_GENERAL, whisk.DISPLAY_MSG, whisk.NO_DISPLAY_USAGE)
            return err
        }

        client.Namespace = qName.namespace

        payload := map[string]interface{}{}

        if len(flags.common.param) > 0 {
            parameters, err := parseParameters(flags.common.param)
            if err != nil {
                err := whisk.MakeWskError(err, whisk.EXITCODE_ERR_GENERAL, whisk.DISPLAY_MSG, whisk.NO_DISPLAY_USAGE)
                return err
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
            }
        }

        activation, _, err := client.Actions.Invoke(qName.entityName, payload, flags.common.blocking)
        if err != nil {
            return err
        }

        if flags.common.blocking && flags.action.result {
            printJSON(activation.Response.Result)
        } else if flags.common.blocking {
            fmt.Printf("%s invoked %s with id %s\n", color.GreenString("ok:"), boldString(qName.entityName), boldString(activation.ActivationID))
            boldPrintf("response:\n")
            printJSON(activation.Response)
        } else {
            fmt.Printf("%s invoked %s with id %s\n", color.GreenString("ok:"), boldString(qName.entityName), boldString(activation.ActivationID))
        }

        return nil
    },
}

var actionGetCmd = &cobra.Command{
    Use:   "get <name string>",
    Short: "get action",
    SilenceUsage:   true,
    SilenceErrors:  true,
    RunE: func(cmd *cobra.Command, args []string) error {
        var err error

        if len(args) != 1 {
            err = errors.New("Invalid argument")
            err := whisk.MakeWskError(err, whisk.EXITCODE_ERR_GENERAL, whisk.DISPLAY_MSG, whisk.NO_DISPLAY_USAGE)
            return err
        }

        qName, err := parseQualifiedName(args[0])
        if err != nil {
            err := whisk.MakeWskError(err, whisk.EXITCODE_ERR_GENERAL, whisk.DISPLAY_MSG, whisk.NO_DISPLAY_USAGE)
            return err
        }

        client.Namespace = qName.namespace

        action, _, err := client.Actions.Get(qName.entityName)
        if err != nil {
            err := whisk.MakeWskError(err, whisk.EXITCODE_ERR_GENERAL, whisk.DISPLAY_MSG, whisk.NO_DISPLAY_USAGE)
            return err
        }
        // print out response

        if flags.common.summary {
            fmt.Printf("%s /%s/%s\n", boldString("action"), action.Namespace, action.Name)
        } else {
            fmt.Printf("%s got action %s\n", color.GreenString("ok:"), boldString(qName.entityName))
            printJSON(action)
        }

        return nil
    },
}

var actionDeleteCmd = &cobra.Command{
    Use:   "delete <name string>",
    Short: "delete action",
    SilenceUsage:   true,
    SilenceErrors:  true,
    RunE: func(cmd *cobra.Command, args []string) error {
        qName, err := parseQualifiedName(args[0])
        if err != nil {
            err := whisk.MakeWskError(err, whisk.EXITCODE_ERR_GENERAL, whisk.DISPLAY_MSG, whisk.NO_DISPLAY_USAGE)
            return err
        }

        client.Namespace = qName.namespace

        _, err = client.Actions.Delete(qName.entityName)

        if err != nil {
            return err
        }

        // print out response
        fmt.Printf("%s deleted action %s\n", color.GreenString("ok:"), boldString(qName.entityName))

        return nil
    },
}

var actionListCmd = &cobra.Command{
    Use:   "list <namespace string>",
    Short: "list all actions",
    SilenceUsage:   true,
    SilenceErrors:  true,
    RunE: func(cmd *cobra.Command, args []string) error {
        var err error
        qName := qualifiedName{}
        if len(args) == 1 {
            qName, err = parseQualifiedName(args[0])
            if err != nil {
                err := whisk.MakeWskError(err, whisk.EXITCODE_ERR_GENERAL, whisk.DISPLAY_MSG, whisk.NO_DISPLAY_USAGE)
                return err
            }
            ns := qName.namespace
            if len(ns) == 0 {
                err = errors.New("No valid namespace detected.  Make sure that namespace argument is preceded by a \"/\"")
                err := whisk.MakeWskError(err, whisk.EXITCODE_ERR_GENERAL, whisk.DISPLAY_MSG, whisk.NO_DISPLAY_USAGE)
                return err
            }

            client.Namespace = ns

            if pkg := qName.packageName; len(pkg) > 0 {
                // todo :: scope call to package
            }
        }

        options := &whisk.ActionListOptions{
            Skip:  flags.common.skip,
            Limit: flags.common.limit,
        }

        actions, _, err := client.Actions.List(options)
        if err != nil {
            return err
        }

        printList(actions)

        return nil
    },
}

/*
usage: wsk action update [-h] [-u AUTH] [--docker] [--copy] [--sequence]
[--lib LIB] [--shared [{yes,no}]]
[-a ANNOTATION ANNOTATION] [-p PARAM PARAM]
[-t TIMEOUT] [-m MEMORY]
name [artifact]
*/
func parseAction(cmd *cobra.Command, args []string) (*whisk.Action, bool, error) {
    var err error
    var shared, sharedSet bool
    var actionName, artifact string

    if (IsDebug()) {
        fmt.Printf("Action arguments: %s\n", args)
    }

    if len(args) < 1 {
        err = errors.New("Invalid argument list")
        return nil, sharedSet, err
    }

    actionName = args[0]

    qName := qualifiedName{}

    qName, err = parseQualifiedName(args[0])
    if err != nil {
        return nil, sharedSet, err
    }

    client.Namespace = qName.namespace

    if len(args) == 2 {
        artifact = args[1]
    }

    if flags.action.shared == "yes" {
        shared = true
        sharedSet = true
    } else if flags.action.shared == "no" {
        shared = false
        sharedSet = true
    } else {
        sharedSet = false
    }

    parameters, err := parseParameters(flags.common.param)
    if err != nil {
        return nil, sharedSet, err
    }

    annotations, err := parseAnnotations(flags.common.annotation)
    if err != nil {
        return nil, sharedSet, err
    }

    // TODO: exclude limits if none set
    /*limits := whisk.Limits{
            Timeout: flags.action.timeout,
            Memory:  flags.action.memory,
    }*/

    action := new(whisk.Action)


    if flags.action.docker {
        action.Exec = new(whisk.Exec)
        action.Exec.Image = artifact
    } else if flags.action.copy {
        existingAction, _, err := client.Actions.Get(actionName)
        if err != nil {
            return nil, sharedSet, err
        }

        action.Exec = existingAction.Exec
    } else if flags.action.sequence {
        currentNamespace := client.Config.Namespace
        client.Config.Namespace = "whisk.system"
        pipeAction, _, err := client.Actions.Get("system/pipe")
        if err != nil {
            return nil, sharedSet, err
        }
        action.Exec = pipeAction.Exec
        client.Config.Namespace = currentNamespace
    } else if artifact != "" {
        stat, err := os.Stat(artifact)
        if err != nil {
            // file does not exist
            return nil, sharedSet, err
        }

        file, err := ioutil.ReadFile(artifact)
        if err != nil {
            return nil, sharedSet, err
        }
        if action.Exec == nil {
            action.Exec = new(whisk.Exec)
        }

        action.Exec.Code = string(file)

        if matched, _ := regexp.MatchString(".swift$", stat.Name()); matched {
            action.Exec.Kind = "swift"
        } else {
            action.Exec.Kind = "nodejs"
        }
    }

    if flags.action.lib != "" {
        file, err := os.Open(flags.action.lib)
        if err != nil {
            return nil, sharedSet, err
        }

        var r io.Reader
        switch ext := filepath.Ext(file.Name()); ext {
        case "tar":
            r = tar.NewReader(file)
        case "gzip":
            r, err = gzip.NewReader(file)
        default:
            err = fmt.Errorf("Unrecognized file compression %s", ext)
        }
        if err != nil {
            return nil, sharedSet, err
        }
        lib, err := ioutil.ReadAll(r)
        if err != nil {
            return nil, sharedSet, err
        }

        if action.Exec == nil {
            action.Exec = new(whisk.Exec)
        }

        action.Exec.Init = base64.StdEncoding.EncodeToString(lib)
    }

    action.Name = qName.entityName
    action.Namespace = qName.namespace
    action.Publish = shared
    action.Annotations = annotations
    action.Parameters = parameters
    //action.Limits = limits

    if IsDebug() {
        fmt.Printf("Parsed action struct: %+v\n", action)
    }

    return action, sharedSet, nil
}

///////////
// Flags //
///////////

func init() {
    actionCreateCmd.Flags().BoolVar(&flags.action.docker, "docker", false, "treat artifact as docker image path on dockerhub")
    actionCreateCmd.Flags().BoolVar(&flags.action.copy, "copy", false, "treat artifact as the name of an existing action")
    actionCreateCmd.Flags().BoolVar(&flags.action.sequence, "sequence", false, "treat artifact as comma separated sequence of actions to invoke")
    actionCreateCmd.Flags().StringVar(&flags.action.shared, "shared", "", "shared action (default: private)")
    actionCreateCmd.Flags().StringVar(&flags.action.lib, "lib", "", "add library to artifact (must be a gzipped tar file)")
    actionCreateCmd.Flags().StringVar(&flags.action.xPackage, "package", "", "package")
    actionCreateCmd.Flags().IntVarP(&flags.action.timeout, "timeout", "t", 0, "the timeout limit in miliseconds when the action will be terminated")
    actionCreateCmd.Flags().IntVarP(&flags.action.memory, "memory", "m", 0, "the memory limit in MB of the container that runs the action")
    actionCreateCmd.Flags().StringSliceVarP(&flags.common.annotation, "annotation", "a", []string{}, "annotations")
    actionCreateCmd.Flags().StringSliceVarP(&flags.common.param, "param", "p", []string{}, "default parameters")

    actionUpdateCmd.Flags().BoolVar(&flags.action.docker, "docker", false, "treat artifact as docker image path on dockerhub")
    actionUpdateCmd.Flags().BoolVar(&flags.action.copy, "copy", false, "treat artifact as the name of an existing action")
    actionUpdateCmd.Flags().BoolVar(&flags.action.sequence, "sequence", false, "treat artifact as comma separated sequence of actions to invoke")
    actionUpdateCmd.Flags().StringVar(&flags.action.shared, "shared", "", "shared action (default: private)")
    actionUpdateCmd.Flags().StringVar(&flags.action.lib, "lib", "", "add library to artifact (must be a gzipped tar file)")
    actionUpdateCmd.Flags().StringVar(&flags.action.xPackage, "package", "", "package")
    actionUpdateCmd.Flags().IntVarP(&flags.action.timeout, "timeout", "t", 0, "the timeout limit in miliseconds when the action will be terminated")
    actionUpdateCmd.Flags().IntVarP(&flags.action.memory, "memory", "m", 0, "the memory limit in MB of the container that runs the action")
    actionUpdateCmd.Flags().StringSliceVarP(&flags.common.annotation, "annotation", "a", []string{}, "annotations")
    actionUpdateCmd.Flags().StringSliceVarP(&flags.common.param, "param", "p", []string{}, "default parameters")

    actionInvokeCmd.Flags().StringSliceVarP(&flags.common.param, "param", "p", []string{}, "parameters")
    actionInvokeCmd.Flags().BoolVarP(&flags.common.blocking, "blocking", "b", false, "blocking invoke")
    actionInvokeCmd.Flags().BoolVarP(&flags.action.result, "result", "r", false, "show only activation result if a blocking activation (unless there is a failure)")

    actionGetCmd.Flags().BoolVarP(&flags.common.summary, "summary", "s", false, "summarize entity details")

    actionListCmd.Flags().IntVarP(&flags.common.skip, "skip", "s", 0, "skip this many entitites from the head of the collection")
    actionListCmd.Flags().IntVarP(&flags.common.limit, "limit", "l", 30, "only return this many entities from the collection")
    actionListCmd.Flags().BoolVar(&flags.common.full, "full", false, "include full entity description")

    actionCmd.AddCommand(
        actionCreateCmd,
        actionUpdateCmd,
        actionInvokeCmd,
        actionGetCmd,
        actionDeleteCmd,
        actionListCmd,
    )
}
