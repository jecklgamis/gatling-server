package gatling.test.example.simulation

import io.gatling.core.Predef.{constantUsersPerSec, global, scenario, _}
import io.gatling.http.Predef.{http, status, _}

import scala.concurrent.duration._
import scala.sys.SystemProperties

class SingleFileExampleSimulation extends Simulation {

  import PerfTestConfig._

  val httpConf = http.baseUrl(baseUrl)
  val getUsers = scenario("Endpoint calls")
    .exec(http("endpoint")
      .get("")
      .check(status.is(200))
    )
  setUp(getUsers.inject(
    constantUsersPerSec(requestPerSecond) during (durationMin minutes))
    .protocols(httpConf))
    .assertions(
      global.responseTime.max.lt(maxResponseTimeMs),
      global.responseTime.mean.lt(meanResponseTimeMs),
      global.successfulRequests.percent.gt(95)
    )

  object PerfTestConfig {

    import SystemPropertiesUtil._

    val baseUrl = getAsStringOrElse("baseUrl", "http://localhost:8080")
    val requestPerSecond = getAsDoubleOrElse("requestPerSecond", 10f)
    val durationMin = getAsDoubleOrElse("durationMin", 1.0)
    val meanResponseTimeMs = getAsIntOrElse("meanResponseTimeMs", 500)
    val maxResponseTimeMs = getAsIntOrElse("maxResponseTimeMs", 1000)
  }

  object SystemPropertiesUtil {
    val systemProperties = new SystemProperties

    def getAsIntOrElse(property: String, default: Int): Int =
      systemProperties.getOrElse(property, default).toString.toInt

    def getAsStringOrElse(property: String, default: String): String =
      systemProperties.getOrElse(property, default)

    def getAsBooleanOrElse(property: String, default: Boolean): Boolean =
      systemProperties.getOrElse(property, default).toString.toBoolean

    def getAsDoubleOrElse(property: String, default: Double): Double =
      systemProperties.getOrElse(property, default).toString.toDouble
  }

}