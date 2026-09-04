package gatling

import "github.com/jecklgamis/gatling-server/pkg/env"

func SomeGatling() *Gatling {
	return &Gatling{ScriptsDir: env.GetOrElse("GATLING_TEST_SCRIPTS_DIR", "../../scripts")}
}
