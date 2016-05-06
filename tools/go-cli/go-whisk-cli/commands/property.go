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
        "bufio"
        "fmt"
        "os"
        "strings"

        "github.com/mitchellh/go-homedir"
        "github.com/spf13/cobra"
        "net"
)

var Properties struct {
        Auth       string
        APIHost    string
        APIVersion string
        APIBuild   string
        APIBuildNo string
        CLIVersion string
        Namespace  string
        PropsFile  string
}

const DefaultAuth       string = ""
const DefaultAPIHost    string = "openwhisk.ng.bluemix.net"
const DefaultAPIVersion string = "v1"
const DefaultAPIBuild   string = ""
const DefaultAPIBuildNo string = ""
const DefaultCLIVersion string = ""
const DefaultNamespace  string = "_"
const DefaultPropsFile  string = "~/.wskprops"

var propertyCmd = &cobra.Command{
        Use:   "property",
        Short: "work with whisk properties",
}

var propertySetCmd = &cobra.Command{
        Use:   "set",
        Short: "set property",
        Run: func(cmd *cobra.Command, args []string) {
                // get current props
                props, err := readProps(Properties.PropsFile)
                if err != nil {
                        fmt.Println(err)
                        return
                }

                // read in each flag, update if necessary

                if auth := flags.global.auth; len(auth) > 0 {
                        props["AUTH"] = auth
                        fmt.Println("ok: whisk auth set")
                }

                if apiHost := flags.property.apihostSet; len(apiHost) > 0 {
                        props["APIHOST"] = apiHost
                        fmt.Println("ok: whisk API host set to ", apiHost)
                }

                if apiVersion := flags.property.apiversionSet; len(apiVersion) > 0 {
                        props["APIVERSION"] = apiVersion
                        fmt.Println("ok: whisk API version set to ", apiVersion)
                }

                if namespace := flags.property.namespaceSet; len(namespace) > 0 {

                        namespaces, _, err := client.Namespaces.List()
                        if err != nil {
                                fmt.Println(err)
                                return
                        }

                        var validNamespace bool
                        for _, ns := range namespaces {
                                if ns.Name == namespace {
                                        validNamespace = true
                                }
                        }

                        if !validNamespace {
                                err = fmt.Errorf("Invalid namespace %s", namespace)
                                fmt.Println(err)
                                return
                        }

                        props["NAMESPACE"] = namespace
                        fmt.Println("ok: whisk namespace set to ", namespace)
                }

                err = writeProps(Properties.PropsFile, props)
                if err != nil {
                        fmt.Println(err)
                        return
                }

        },
}

var propertyUnsetCmd = &cobra.Command{
        Use:   "unset",
        Short: "unset property",
        Run: func(cmd *cobra.Command, args []string) {
                props, err := readProps(Properties.PropsFile)
                if err != nil {
                        fmt.Println(err)
                        return
                }

                // read in each flag, update if necessary

                if flags.property.auth {
                        delete(props, "AUTH")
                        fmt.Print("ok: whisk auth deleted")
                        if len(DefaultAuth) > 0 {
                                fmt.Printf("; the default value of '%s' will be used.\n", DefaultAuth)
                        } else {
                                fmt.Println("; no default value will be used.")
                        }
                }

                if flags.property.namespace {
                        delete(props, "NAMESPACE")
                        fmt.Print("ok: whisk namespace deleted")
                        if len(DefaultNamespace) > 0 {
                                fmt.Printf("; the default value of '%s' will be used.\n", DefaultNamespace)
                        } else {
                                fmt.Println("; there is no default value that can be used.")
                        }
                }

                if flags.property.apihost {
                        delete(props, "APIHOST")
                        fmt.Print("whisk API host deleted; using default")
                        if len(DefaultAPIHost) > 0 {
                                fmt.Printf("; the default value of '%s' will be used.\n", DefaultAPIHost)
                        } else {
                                fmt.Println("; there is no default value that can be used.")
                        }
                }

                if flags.property.apiversion {
                        delete(props, "APIVERSION")
                        fmt.Print("ok: whisk API version deleted")
                        if len(DefaultAPIVersion) > 0 {
                                fmt.Printf("; the default value of '%s' will be used.\n", DefaultAPIVersion)
                        } else {
                                fmt.Println("; there is no default value that can be used.")
                        }
                }

                err = writeProps(Properties.PropsFile, props)
                if err != nil {
                        fmt.Println(err)
                        return
                }
                loadProperties()
        },
}

var propertyGetCmd = &cobra.Command{
        Use:   "get",
        Short: "get property",
        Run: func(cmd *cobra.Command, args []string) {

                if flags.property.all || flags.property.auth {
                        fmt.Println("whisk auth\t\t", Properties.Auth)
                }

                if flags.property.all || flags.property.apihost {
                        fmt.Println("whisk API host\t\t", Properties.APIHost)
                }

                if flags.property.all || flags.property.apiversion {
                        fmt.Println("whisk API version\t", Properties.APIVersion)
                }

                if flags.property.all|| flags.property.cliversion {
                        fmt.Println("whisk CLI version\t", Properties.CLIVersion)
                }

                if flags.property.all || flags.property.namespace {
                        fmt.Println("whisk namespace\t\t", Properties.Namespace)
                }

                if flags.property.all || flags.property.apibuild || flags.property.apibuildno {
                        info, _, err := client.Info.Get()
                        if err != nil {
                                fmt.Println(err)
                        } else {
                                if flags.property.all || flags.property.apibuild {
                                        fmt.Println("whisk API build\t\t", info.Build)
                                }
                                if flags.property.all || flags.property.apibuildno {
                                        fmt.Println("whisk API build number\t", info.BuildNo)
                                }
                        }
                }

        },
}

func init() {
        propertyCmd.AddCommand(
                propertySetCmd,
                propertyUnsetCmd,
                propertyGetCmd,
        )

        // need to set property flags as booleans instead of strings... perhaps with boolApihost...
        propertyGetCmd.Flags().BoolVar(&flags.property.auth, "auth", false, "authorization key")
        propertyGetCmd.Flags().BoolVar(&flags.property.apihost, "apihost", false, "whisk API host")
        propertyGetCmd.Flags().BoolVar(&flags.property.apiversion, "apiversion", false, "whisk API version")
        propertyGetCmd.Flags().BoolVar(&flags.property.apibuild, "apibuild", false, "whisk API build version")
        propertyGetCmd.Flags().BoolVar(&flags.property.apibuildno, "apibuildno", false, "whisk API build number")
        propertyGetCmd.Flags().BoolVar(&flags.property.cliversion, "cliversion", false, "whisk CLI version")
        propertyGetCmd.Flags().BoolVar(&flags.property.namespace, "namespace", false, "authorization key")
        propertyGetCmd.Flags().BoolVar(&flags.property.all, "all", false, "all properties")

        propertySetCmd.Flags().StringVarP(&flags.global.auth, "auth", "u", "", "authorization key")
        propertySetCmd.Flags().StringVar(&flags.property.apihostSet, "apihost", "", "whisk API host")
        propertySetCmd.Flags().StringVar(&flags.property.apiversionSet, "apiversion", "", "whisk API version")
        propertySetCmd.Flags().StringVar(&flags.property.namespaceSet, "namespace", "", "whisk namespace")

        propertyUnsetCmd.Flags().BoolVar(&flags.property.auth, "auth", false, "authorization key")
        propertyUnsetCmd.Flags().BoolVar(&flags.property.apihost, "apihost", false, "whisk API host")
        propertyUnsetCmd.Flags().BoolVar(&flags.property.apiversion, "apiversion", false, "whisk API version")
        propertyUnsetCmd.Flags().BoolVar(&flags.property.namespace, "namespace", false, "whisk namespace")

}

func setDefaultProperties() {
        Properties.Auth = DefaultAuth
        Properties.Namespace = DefaultNamespace
        Properties.APIHost = DefaultAPIHost
        Properties.APIBuild = DefaultAPIBuild
        Properties.APIBuildNo = DefaultAPIBuildNo
        Properties.APIVersion = DefaultAPIVersion
        Properties.CLIVersion = DefaultCLIVersion
        Properties.PropsFile = DefaultPropsFile
}

func getPropertiesFilePath() (propsFilePath string, err error) {
        // Environment variable overrides the default properties file path
        if propsFilePath := os.Getenv("WSK_CONFIG_FILE"); len(propsFilePath) > 0 {
                if IsDebug() {
                        fmt.Println("Using property file from WSK_CONFIG_FILE environment variable: ", propsFilePath)
                }
                return propsFilePath, nil
        } else {
                propsFilePath, err = homedir.Expand(Properties.PropsFile)
                if IsDebug() {
                        fmt.Println("Using property file home dir: ", propsFilePath)
                }
                return propsFilePath, err
        }
}

func loadProperties() error {
        var err error

        setDefaultProperties()

        Properties.PropsFile, err = getPropertiesFilePath()
        if err != nil {
                return err
        }

        props, err := readProps(Properties.PropsFile)
        if err != nil {
                return err
        }

        if authToken, hasProp := props["AUTH"]; hasProp {
                Properties.Auth = authToken
        }

        if authToken := os.Getenv("WHISK_AUTH"); len(authToken) > 0 {
                Properties.Auth = authToken
        }

        if apiVersion, hasProp := props["APIVERSION"]; hasProp {
                Properties.APIVersion = apiVersion
        }

        if apiVersion := os.Getenv("WHISK_APIVERSION"); len(apiVersion) > 0 {
                Properties.APIVersion = apiVersion
        }

        if apiHost, hasProp := props["APIHOST"]; hasProp {
                Properties.APIHost = apiHost
        }

        if apiHost := os.Getenv("WHISK_APIHOST"); len(apiHost) > 0 {
                Properties.APIHost = apiHost
        }

        if namespace, hasProp := props["NAMESPACE"]; hasProp {
                Properties.Namespace = namespace
        }

        if namespace := os.Getenv("WHISK_NAMESPACE"); len(namespace) > 0 {
                Properties.Namespace = namespace
        }

        return nil
}

func parseConfigFlags(cmd *cobra.Command, args []string) {

        if auth := flags.global.auth; len(auth) > 0 {
                Properties.Auth = auth
                client.Config.AuthToken = auth
        }

        if namespace := flags.property.namespaceSet; len(namespace) > 0 {
                Properties.Namespace = namespace
                client.Config.Namespace = namespace
        }

        if apiVersion := flags.global.apiversion; len(apiVersion) > 0 {
                Properties.APIVersion = apiVersion
                client.Config.Version = apiVersion
        }

        if apiHost := flags.global.apihost; len(apiHost) > 0 {
                Properties.APIHost = apiHost
                parsedIP := net.ParseIP(apiHost)

                if parsedIP != nil {
                        client.Config.Host = apiHost
                } else {
                        fmt.Println("Invalid IP address.")
                        os.Exit(-1)
                }
        }

        if IsVerbose() {
                client.Config.Verbose = flags.global.verbose
        }
}

func readProps(path string) (map[string]string, error) {

        props := map[string]string{}

        file, err := os.Open(path)
        if err != nil {
                // If file does not exist, just return props
                return props, nil
        }
        defer file.Close()

        lines := []string{}
        scanner := bufio.NewScanner(file)
        for scanner.Scan() {
                lines = append(lines, scanner.Text())
        }

        props = map[string]string{}
        for _, line := range lines {
                kv := strings.Split(line, "=")
                if len(kv) != 2 {
                        // Invalid format; skip
                        continue
                }
                props[kv[0]] = kv[1]
        }

        return props, nil

}

func writeProps(path string, props map[string]string) error {

        file, err := os.Create(path)
        if err != nil {
                return err
        }
        defer file.Close()

        writer := bufio.NewWriter(file)
        defer writer.Flush()
        for key, value := range props {
                line := fmt.Sprintf("%s=%s", strings.ToUpper(key), value)
                _, err = fmt.Fprintln(writer, line)
                if err != nil {
                        return err
                }
        }
        return nil
}
