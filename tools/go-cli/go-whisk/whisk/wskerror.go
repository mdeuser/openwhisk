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

const EXITCODE_ERR_GENERAL      int = 1
const EXITCODE_ERR_NETWORK      int = 2
const EXITCODE_ERR_HTTP_RESP    int = 3

const DISPLAY_MSG bool      = true
const NO_DISPLAY_MSG bool   = false
const DISPLAY_USAGE bool    = true
const NO_DISPLAY_USAGE bool = false

type WskError struct {
    RootErr         error
    ExitCode        int
    DisplayMsg      bool    // Error message should be displayed to console
    MsgDisplayed    bool    // Error message has already been displayed
    DisplayUsage    bool    // When true, the CLI usage should be displayed before exiting
}

func (e WskError) Error() string {
    return e.RootErr.Error()
}

func MakeWskError (e error, ec int, flags ...bool ) (we *WskError) {
    we = &WskError{RootErr: e, ExitCode: ec, DisplayMsg: false, DisplayUsage: false, MsgDisplayed: false}
    if len(flags) > 0 { we.DisplayMsg = flags[0] }
    if len(flags) > 1 { we.DisplayUsage = flags[1] }
    if len(flags) > 2 { we.MsgDisplayed = flags[2] }
    return we
}
