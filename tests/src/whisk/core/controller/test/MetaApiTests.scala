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

import java.io.ByteArrayOutputStream
import java.io.PrintStream

import scala.concurrent.Future
import scala.language.postfixOps

import org.junit.runner.RunWith
import org.scalatest.junit.JUnitRunner

import akka.event.Logging.InfoLevel
import spray.http.StatusCodes._
import spray.httpx.SprayJsonSupport._
import spray.httpx.SprayJsonSupport.sprayJsonMarshaller
import spray.httpx.SprayJsonSupport.sprayJsonUnmarshaller
import spray.json._
import spray.json.DefaultJsonProtocol
import spray.json.DefaultJsonProtocol._
import whisk.common.TransactionId
import whisk.core.controller.WhiskMetaApi
import whisk.core.entity._

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
    override lazy val systemId = Subject().toString

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

        val notmeta = WhiskPackage(
            EntityPath(systemId),
            EntityName("notmeta"),
            annotations = Parameters("meta", JsBoolean(false)))
        put(entityStore, notmeta)

        val alsonotmeta = WhiskPackage(
            EntityPath("xyz"),
            EntityName("notmeta"),
            annotations = Parameters("meta", JsBoolean(false)))
        put(entityStore, alsonotmeta)

        val badmeta = WhiskPackage(
            EntityPath(systemId),
            EntityName("badmeta"),
            annotations = Parameters("meta", JsBoolean(true)))
        put(entityStore, badmeta)

        val methods = Seq(Get, Post, Delete)

        methods.map { m =>
            m("/meta") ~> sealRoute(routes(creds)) ~> check {
                status shouldBe NotFound
            }
        }

        val paths = Seq("/meta/doesntexist", "/meta/xyz", "/meta/notmeta", "/meta/badmeta")
        paths.map { p =>
            methods.map { m =>
                m(p) ~> sealRoute(routes(creds)) ~> check {
                    withClue(p) {
                        status shouldBe MethodNotAllowed
                    }
                }
            }
        }
    }

    it should "invoke action for allowed verbs on meta handler" in {
        implicit val tid = transid()

        val heavymeta = WhiskPackage(
            EntityPath(systemId),
            EntityName("heavymeta"),
            annotations = Parameters("meta", JsBoolean(true)) ++
                Parameters("get", JsString("getApi")) ++
                Parameters("post", JsString("createRoute")) ++
                Parameters("delete", JsString("deleteApi")))
        put(entityStore, heavymeta)

        val methods = Seq((Get, "getApi"), (Post, "createRoute"), (Delete, "deleteApi"))
        methods.map {
            case (m, name) =>
                m("/meta/heavymeta?a=b&c=d&namespace=xyz") ~> sealRoute(routes(creds)) ~> check {
                    status should be(OK)
                    val response = responseAs[JsObject]
                    response shouldBe JsObject(
                        "pkg" -> "heavymeta".toJson,
                        "action" -> name.toJson,
                        "content" -> JsObject(
                            "namespace" -> creds.namespace.toJson, //namespace overriden by API
                            "a" -> "b".toJson,
                            "c" -> "d".toJson))
                }
        }
    }

    it should "invoke action for allowed verbs on meta handler with partial mapping" in {
        implicit val tid = transid()

        val heavymeta = WhiskPackage(
            EntityPath(systemId),
            EntityName("heavymeta"),
            annotations = Parameters("meta", JsBoolean(true)) ++
                Parameters("get", JsString("getApi")))
        put(entityStore, heavymeta)

        val methods = Seq((Get, OK), (Post, MethodNotAllowed), (Delete, MethodNotAllowed))
        methods.map {
            case (m, status) =>
                m("/meta/heavymeta?a=b&c=d&namespace=xyz") ~> sealRoute(routes(creds)) ~> check {
                    status should be(status)
                    if (status == OK) {
                        val response = responseAs[JsObject]
                        response shouldBe JsObject(
                            "pkg" -> "heavymeta".toJson,
                            "action" -> "getApi".toJson,
                            "content" -> JsObject(
                                "namespace" -> creds.namespace.toJson, //namespace overriden by API
                                "a" -> "b".toJson,
                                "c" -> "d".toJson))
                    }
                }
        }
    }

    it should "warn if meta package is public" in {
        implicit val tid = transid()
        val stream = new ByteArrayOutputStream
        val printstream = new PrintStream(stream)
        val savedstream = this.outputStream
        this.outputStream = printstream

        val publicmeta = WhiskPackage(
            EntityPath(systemId),
            EntityName("publicmeta"),
            publish = true,
            annotations = Parameters("meta", JsBoolean(true)) ++
                Parameters("get", JsString("getApi")))
        put(entityStore, publicmeta)

        try {
            Get("/meta/publicmeta") ~> sealRoute(routes(creds)) ~> check {
                status should be(OK)
                stream.toString should include regex (s"""[WARN] *.* '${publicmeta.fullyQualifiedName(true)}' is public""")
                stream.reset()
            }
        } finally {
            stream.close()
            printstream.close()
        }
    }

}
