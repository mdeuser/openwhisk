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

type WskError struct {
    RootErr         error
    ExitCode        int
    DisplayMsg      bool    // Error message should be displayed to console
    MsgDisplayed    bool    // Error message has already been displayed
    DisplayUsage    bool    // When true, the CLI usage should be displayed before exiting
    ErrMsg          string
}

func (e WskError) Error() string {
    return e.RootErr.Error()
}

