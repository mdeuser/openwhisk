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
    "strings"

    "../../go-whisk/whisk"

    "github.com/fatih/color"
    prettyjson "github.com/hokaccha/go-prettyjson"
)

type qualifiedName struct {
    namespace   string
    packageName string
    entityName  string
}

func (qName qualifiedName) String() string {
    output := []string{}

    if len(qName.namespace) > 0 {
        output = append(output, "/", qName.namespace, "/")
    }
    if len(qName.packageName) > 0 {
        output = append(output, qName.packageName, "/")
    }
    output = append(output, qName.entityName)

    return strings.Join(output, "")
}

//
// Parse a (possibly fully qualified) resource name into
// namespace and name components. If the given qualified
// name isNone, then this is a default qualified name
// and it is resolved from properties. If the namespace
// is missing from the qualified name, the namespace is also
// resolved from the property file.
//
// Return a qualifiedName struct
//
// Examples:
//      foo => qName {namespace: "_", entityName: foo}
//      pkg/foo => qName {namespace: "_", entityName: pkg/foo}
//      /ns/foo => qName {namespace: ns, entityName: foo}
//      /ns/pkg/foo => qName {namespace: ns, entityName: pkg/foo}
//
func parseQualifiedName(name string) (qName qualifiedName, err error) {

    // If name has a preceding delimiter (/), it contains a namespace. Otherwise the name does not specify a namespace,
    // so default the namespace to the namespace value set in the properties file; if that is not set, use "_"
    if len(name) > 0 {
        if name[0] == '/' {
            parts := strings.Split(name, "/")
            qName.namespace = parts[1]

            if len(parts) > 2 {
                qName.entityName = strings.Join(parts[2:], "")
            } else {
                qName.entityName = name
            }
        } else {
            qName.entityName = name

            if Properties.Namespace != "" {
                qName.namespace = Properties.Namespace
            } else {
                qName.namespace = "_"
            }
        }
    } else {
        if IsDebug() {
            fmt.Println("parseQualifiedName: Error - empty name string could not be parsed")
        }

        err = whisk.MakeWskError(errors.New("Invalid name format"), whisk.EXITCODE_ERR_GENERAL,
            whisk.DISPLAY_MSG, whisk.DISPLAY_USAGE )
    }

    if IsDebug() {
        fmt.Printf("Package entityName: %s\n", qName.entityName)
        fmt.Printf("Package namespace: %s\n", qName.namespace)
    }

    return qName, err
}

func parseGenericArray(args []string) (whisk.Annotations, error) {
    parsed := make(whisk.Annotations, 0)

    if len(args)%2 != 0 {
        if IsDebug() {
            fmt.Printf("parseKeyValueArray: Number of arguments (%d) must be an even number; args: %#v\n", len(args), args)
        }
        err := whisk.MakeWskError(
            errors.New("key|value arguments must be submitted in comma-separated pairs; keys or values with spaces must be quoted"),
            whisk.EXITCODE_ERR_GENERAL, whisk.DISPLAY_MSG, whisk.DISPLAY_USAGE )
        return parsed, err
    }

    for i := 0; i < len(args); i += 2 {
        parsedItem := make(map[string]interface{}, 0)
        parsedItem["key"] = args[i]
        parsedItem["value"] = args[i + 1]
        parsed = append(parsed, parsedItem)
    }

    return parsed, nil
}


func parseKeyValueArray(args []string) ([]whisk.KeyValue, error) {
    parsed := []whisk.KeyValue{}
    if len(args)%2 != 0 {
        if IsDebug() {
            fmt.Printf("parseKeyValueArray: Number of arguments (%d) must be an even number; args: %#v\n", len(args), args)
        }
        err := whisk.MakeWskError(
            errors.New("key|value arguments must be submitted in comma-separated pairs; keys or values with spaces must be quoted"),
            whisk.EXITCODE_ERR_GENERAL, whisk.DISPLAY_MSG, whisk.DISPLAY_USAGE )
        return parsed, err
    }

    for i := 0; i < len(args); i += 2 {
        keyValue := whisk.KeyValue{
            Key:   args[i],
            Value: args[i+1],
        }
        parsed = append(parsed, keyValue)

    }
    return parsed, nil
}

func parseParameters(args []string) (whisk.Parameters, error) {
    parameters := whisk.Parameters{}
    parsedArgs, err := parseKeyValueArray(args)
    if err != nil {
        return parameters, err
    }
    parameters = whisk.Parameters(parsedArgs)
    return parameters, nil
}

func parseAnnotations(args []string) (whisk.Annotations, error) {
    annotations := whisk.Annotations{}
    //parsedArgs, err := parseKeyValueArray(args)

    parsedArgs, err := parseGenericArray(args)
    if err != nil {
        return annotations, err
    }

    annotations = whisk.Annotations(parsedArgs)

    return annotations, nil
}

var boldString = color.New(color.Bold).SprintFunc()
var boldPrintf = color.New(color.Bold).PrintfFunc()

func printList(collection interface{}) {
    switch collection := collection.(type) {
    case []whisk.Action:
        printActionList(collection)
    case []whisk.Trigger:
        printTriggerList(collection)
    case []whisk.Package:
        printPackageList(collection)
    case []whisk.Rule:
        printRuleList(collection)
    case []whisk.Namespace:
        printNamespaceList(collection)
    case []whisk.Activation:
        printActivationList(collection)
    }
}

func printFullList(collection interface{}) {
    switch collection := collection.(type) {
    case []whisk.Action:

    case []whisk.Trigger:

    case []whisk.Package:

    case []whisk.Rule:

    case []whisk.Namespace:

    case []whisk.Activation:
        printFullActivationList(collection)
    }
}

func printActionList(actions []whisk.Action) {
    boldPrintf("actions\n")
    for _, action := range actions {
        publishState := "private"
        if action.Publish {
            publishState = "shared"
        }
        fmt.Printf("%-70s%s\n", fmt.Sprintf("/%s/%s", action.Namespace, action.Name), publishState)
    }
}

func printTriggerList(triggers []whisk.Trigger) {
    boldPrintf("triggers\n")
    for _, trigger := range triggers {
        publishState := "private"
        if trigger.Publish {
            publishState = "shared"
        }
        fmt.Printf("%-70s%s\n", fmt.Sprintf("/%s/%s", trigger.Namespace, trigger.Name), publishState)
    }
}

func printPackageList(packages []whisk.Package) {
    boldPrintf("packages\n")
    for _, xPackage := range packages {
        publishState := "private"
        if xPackage.Publish {
            publishState = "shared"
        }
        fmt.Printf("%-70s%s\n", fmt.Sprintf("/%s/%s", xPackage.Namespace, xPackage.Name), publishState)
    }
}

func printRuleList(rules []whisk.Rule) {
    boldPrintf("rules\n")
    for _, rule := range rules {
        publishState := "private"
        if rule.Publish {
            publishState = "shared"
        }
        fmt.Printf("%-70s%s\n", fmt.Sprintf("/%s/%s", rule.Namespace, rule.Name), publishState)
    }
}

func printNamespaceList(namespaces []whisk.Namespace) {
    boldPrintf("namespaces\n")
    for _, namespace := range namespaces {
        fmt.Printf("%s\n", namespace.Name)
    }
}

func printActivationList(activations []whisk.Activation) {
    boldPrintf("activations\n")
    for _, activation := range activations {
        fmt.Printf("%s%20s\n", activation.ActivationID, activation.Name)
    }
}

func printFullActivationList(activations []whisk.Activation) {
    boldPrintf("activations\n")
    for _, activation := range activations {
        printJsonNoColor(activation)
    }
}

//
//
//
// func parseParameters(jsonStr string) (whisk.Parameters, error) {
// 	parameters := whisk.Parameters{}
// 	if len(jsonStr) == 0 {
// 		return parameters, nil
// 	}
// 	reader := strings.NewReader(jsonStr)
// 	err := json.NewDecoder(reader).Decode(&parameters)
// 	if err != nil {
// 		return nil, err
// 	}
// 	return parameters, nil
// }
//
// func parseAnnotations(jsonStr string) (whisk.Annotations, error) {
// 	annotations := whisk.Annotations{}
// 	if len(jsonStr) == 0 {
// 		return annotations, nil
// 	}
// 	reader := strings.NewReader(jsonStr)
// 	err := json.NewDecoder(reader).Decode(&annotations)
// 	if err != nil {
// 		return nil, err
// 	}
// 	return annotations, nil
// }

func logoText() string {

    logo := `

__          ___     _     _
\ \        / / |   (_)   | |
 \ \  /\  / /| |__  _ ___| | __
  \ \/  \/ / | '_ \| / __| |/ /
   \  /\  /  | | | | \__ \   <
    \/  \/   |_| |_|_|___/_|\_\

                        `

    return logo
}

func printJSON(v interface{}) {
    output, _ := prettyjson.Marshal(v)
    fmt.Println(string(output))
}

// Same as printJSON, but with coloring disabled.
func printJsonNoColor(v interface{}) {
    jsonFormatter := prettyjson.NewFormatter()
    jsonFormatter.DisabledColor = true
    output, err := jsonFormatter.Marshal(v)
    if err != nil && IsDebug() {
        fmt.Printf("printJsonNoColor: Marshal() failure: %s\n", err)
    }
    fmt.Println(string(output))
}
