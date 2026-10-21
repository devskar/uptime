package main

import (
	"encoding/json"
	"fmt"
	"log"
)

func display(healthcheckResult HealthcheckResult) {
	if prettyOutput, err := json.MarshalIndent(healthcheckResult, "", "  "); err == nil {
		fmt.Println(string(prettyOutput))
	} else {
		log.Println(err.Error())
	}
}
