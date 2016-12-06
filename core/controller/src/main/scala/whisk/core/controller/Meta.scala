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

package whisk.core.controller

import scala.concurrent.Future

import akka.actor.ActorSystem
import spray.client.pipelining._
import spray.http._
import spray.http.BasicHttpCredentials
import spray.http.HttpMethods._
import spray.http.StatusCodes._
import spray.http.Uri
import spray.http.Uri.Path
import spray.httpx.SprayJsonSupport._
import spray.json._
import spray.json.DefaultJsonProtocol._
import spray.routing.Directives
import whisk.common.Logging
import whisk.common.TransactionId
import whisk.core.entity._
import scala.concurrent.ExecutionContext

trait WhiskMetaApi extends Directives with Logging {
    services: WhiskServices =>

    /** API path and version for posting activations directly through the host. */
    val apipath: String
    val apiversion: String

    /** An actor system for HTTP requests. */
    protected implicit val actorSystem: ActorSystem

    /** An execution context for futures. */
    protected implicit val executionContext: ExecutionContext

    /** The route prefix e.g., /meta/package-name. */
    protected val routePrefix = pathPrefix("meta")

    /** Allowed verbs. */
    private lazy val allowedOperations = get | delete | post

    /** Allowed packages implementing meta api methods. */
    protected val allowedMetaApis = Seq("routemgmt")

    private val hostPath = Uri(s"http://localhost:${whiskConfig.servicePort}")
    private val systemNamespace = "whisk.system"
    private val baseApiPath = Path(s"/api/$apiversion") / "namespaces" / systemNamespace / "actions"

    private def makeUrl(namespace: String, pkg: String, action: String) = {
        val actionPath = (Path.SingleSlash + pkg) / action
        hostPath.withPath(baseApiPath + actionPath.toString).toString.concat("?blocking=true")
    }

    private lazy val pipeline: Future[HttpRequest => Future[HttpResponse]] = {
        val authStore = WhiskAuthStore.datastore(whiskConfig)
        val keyLookup = WhiskAuth.get(authStore, Subject(systemNamespace), true)(TransactionId.controller)

        keyLookup.map {
            key =>
                val validCredentials = BasicHttpCredentials(key.authkey.uuid(), key.authkey.key())
                addCredentials(validCredentials) ~> sendReceive
        }
    }

    /**
     * Invokes an actions via REST API.
     * This is a stop gap and will be replaced by internal activation POST in the future.
     *
     * @return Future[JsObject] from action result which is either an activation record (less logs)
     * if status is 200 or an activation id if 202.
     */
    protected def invokeAction(requestBody: JsObject, pkg: String, action: String)(implicit transid: TransactionId): Future[JsObject] = {
        val url = makeUrl(systemNamespace, pkg, action)
        pipeline flatMap {
            _(Post(url, requestBody)) map {
                response =>
                    val result = response.entity.asString.parseJson.asJsObject
                    val code = response.status
                    info(this, s"$action status code: $code")
                    result
            }
        }
    }

    /** Extracts the HTTP method and query params. */
    private val requestMethodAndParams = {
        extract(ctx => (ctx.request.method, ctx.request.message.uri.query.toMap))
    }

    def routes(user: Identity)(implicit transid: TransactionId) = {
        (routePrefix & pathPrefix(Segment) & allowedOperations) {
            metaPackage =>
                if (allowedMetaApis.contains(metaPackage)) {
                    requestMethodAndParams {
                        case (method, params) =>
                            val content = params + ("namespace" -> user.namespace())
                            val action = method match {
                                case GET    => "getApi"
                                case POST   => "createRoute"
                                case DELETE => "deleteApi"
                            }
                            complete(OK, invokeAction(content.toJson.asJsObject, metaPackage, action))
                    }
                } else reject
        }
    }

}
