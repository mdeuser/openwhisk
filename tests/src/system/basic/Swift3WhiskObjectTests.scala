/*
 * Copyright 2015-2016 IBM Corporation
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package system.basic

import scala.concurrent.duration.DurationInt
import scala.language.postfixOps

import org.junit.runner.RunWith
import org.scalatest.junit.JUnitRunner

import common.TestHelpers
import common.TestUtils
import common.Wsk
import common.WskProps
import common.WskTestHelpers
import spray.json.DefaultJsonProtocol.StringJsonFormat
import spray.json.pimpAny

@RunWith(classOf[JUnitRunner])
class Swift3WhiskObjectTests
    extends TestHelpers
    with WskTestHelpers {

    implicit val wskprops = WskProps()
    val wsk = new Wsk()

    behavior of "Swift 3 Whisk backend API"

    it should "allow Swift actions to invoke other actions" in withAssetCleaner(wskprops) {
        (wp, assetHelper) =>
            // use CLI to create action from dat/actions/invokeAction.swift
            val file = TestUtils.getCatalogFilename("/samples/invoke.swift")
            val actionName = "invokeAction"
            assetHelper.withCleaner(wsk.action, actionName) {
                (action, _) => action.create(name = actionName, artifact = Some(file), kind = Some("swift:3"))
            }

            // invoke the action
            val run = wsk.action.invoke(actionName, Map("key0" -> "value0"))
            withActivation(wsk.activation, run, initialWait = 5 seconds, totalWait = 60 seconds) {
                activation =>
                    val logs = activation.fields("logs").toString

                    logs should include("It is now")
                    logs should not include ("Could not parse date of of the response.")
                    logs should not include ("Could not invoke date action.")
            }
    }

    it should "allow Swift actions to trigger events" in withAssetCleaner(wskprops) {
        (wp, assetHelper) =>
            // create a trigger
            val triggerName = s"TestTrigger ${System.currentTimeMillis()}"
            assetHelper.withCleaner(wsk.trigger, triggerName) {
                (trigger, _) => trigger.create(triggerName)
            }

            // create an action that fires the trigger
            val file = TestUtils.getCatalogFilename("/samples/trigger.swift")
            val actionName = "ActionThatTriggers"
            assetHelper.withCleaner(wsk.action, actionName) {
                (action, _) => action.create(name = actionName, artifact = Some(file), kind = Some("swift:3"))
            }

            // invoke the action
            val run = wsk.action.invoke(actionName, Map("triggerName" -> triggerName))
            withActivation(wsk.activation, run, initialWait = 5 seconds, totalWait = 60 seconds) {
                activation =>
                    activation.fields("logs").toString should include(s"Tigger Name: $triggerName")

                    // wait for trigger activation
                    val triggerActivations = wsk.activation.pollFor(1, Some(triggerName), retries = 20)
                    withClue(s"trigger activations for $triggerName:") {
                        triggerActivations.length should be(1)
                    }
            }
    }

}
