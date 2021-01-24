package gatling.test.example.simulation

import io.gatling.core.Predef.{constantUsersPerSec, global, scenario, _}
import io.gatling.http.Predef.{http, status, _}

import scala.concurrent.duration._
import scala.sys.SystemProperties

class SingleFileExampleSimulation extends Simulation {
  val httpConf = http.baseUrl("http://localhost:8080")
  val getUsers = scenario("Root end point calls")
    .exec(http("root end point")
      .get("")
      .check(status.is(200))
    )
  setUp(getUsers.inject(
    constantUsersPerSec(10) during (0.10 minutes))
    .protocols(httpConf))
    .assertions(
      global.responseTime.max.lt(5000),
      global.responseTime.mean.lt(10000),
      global.successfulRequests.percent.gt(95)
    )
}


