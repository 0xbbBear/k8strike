package evaluate

import (
	"k8strike/pkg/conf"
	"log"
	"os"
	"regexp"
)

func SearchSensitiveEnv() {
	for _, env := range os.Environ() {
		ans, err := regexp.MatchString(conf.SensitiveEnvRegex, env)
		if err != nil {
			log.Println(err)
		} else if ans {
			log.Printf("sensitive env found:\n\t%s", env)
		}
	}
}

func init() {
	RegisterSimpleCheck(CategoryServices, "services.sensitive_env", "Search sensitive environment variables", SearchSensitiveEnv)
}
