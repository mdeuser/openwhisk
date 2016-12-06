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

package whisk.core.controller.test

import java.time.Instant

import scala.language.postfixOps

import org.junit.runner.RunWith
import org.scalatest.junit.JUnitRunner
import akka.event.Logging.InfoLevel

import spray.http.StatusCodes._
import spray.httpx.SprayJsonSupport._
import spray.httpx.SprayJsonSupport.sprayJsonMarshaller
import spray.httpx.SprayJsonSupport.sprayJsonUnmarshaller
import spray.json.DefaultJsonProtocol
import spray.json.DefaultJsonProtocol._
import spray.json._
import whisk.core.controller.WhiskMetaApi
import whisk.core.entity._
import whisk.http.ErrorResponse
import whisk.http.Messages
import whisk.common.TransactionId
import scala.concurrent.Future

/**
 * Tests Meta API.
 *
 * Unit tests of the controller service as a standalone component.
 * These tests exercise a fresh instance of the service object in memory -- these
 * tests do NOT communication with a whisk deployment.
 *
 *
 * @Idioglossia
 * "using Specification DSL to write unit tests, as in should, must, not, be"
 * "using Specs2RouteTest DSL to chain HTTP requests for unit testing, as in ~>"
 */
@RunWith(classOf[JUnitRunner])
class MetaApiTests extends ControllerTestCommon with WhiskMetaApi {

    override val apipath = "api"
    override val apiversion = "v1"

    /** Meta API tests */
    behavior of "Meta API"

    val creds = WhiskAuth(Subject(), AuthKey()).toIdentity
    val namespace = EntityPath(creds.subject())
    setVerbosity(InfoLevel)

    override protected def invokeAction(requestBody: JsObject, pkg: String, action: String)(
        implicit transid: TransactionId): Future[JsObject] = {
        Future.successful(JsObject(
            "pkg" -> pkg.toJson,
            "action" -> action.toJson,
            "content" -> requestBody))
    }

    it should "reject access to unknown package" in {
        implicit val tid = transid()

        val methods = Seq(Get, Post, Delete)
        val paths = Seq("/meta", "/meta/xyz")

        paths.map { p =>
            methods.map { m =>
                m(p) ~> sealRoute(routes(creds)) ~> check {
                    withClue(p) {
                        status shouldBe NotFound
                    }
                }
            }
        }

        Put(s"/meta/routemgmt") ~> sealRoute(routes(creds)) ~> check {
            status shouldBe MethodNotAllowed
        }
    }

    it should "invoke routemgmt action for allowed verbs" in {
        implicit val tid = transid()

        val methods = Seq((Get, "getApi"), (Post, "createRoute"), (Delete, "deleteApi"))

        methods.map {
            case (m, name) =>
                m("/meta/routemgmt?a=b&c=d&namespace=xyz") ~> sealRoute(routes(creds)) ~> check {
                    status should be(OK)
                    val response = responseAs[JsObject]
                    response shouldBe JsObject(
                        "pkg" -> "routemgmt".toJson,
                        "action" -> name.toJson,
                        "content" -> JsObject(
                            "namespace" -> creds.namespace.toJson, //namespace overriden by API
                            "a" -> "b".toJson,
                            "c" -> "d".toJson))
                }
        }
    }

}
