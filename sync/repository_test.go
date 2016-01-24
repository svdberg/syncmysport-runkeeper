package sync

import (
	"fmt"
	"testing"
)

func TestMakeDbStringHerokuCompliant(t *testing.T) {
	herokuDbString := "mysql://b932b77e43b903:83f50752@eu-cdbr-west-01.cleardb.com/heroku_e6b0083a037c2d6?reconnect=true"
	expectedResult := "mysql://b932b77e43b903:83f50752@tcp(eu-cdbr-west-01.cleardb.com:3306)/heroku_e6b0083a037c2d6?reconnect=true"

	result := MakeDbStringHerokuCompliant(herokuDbString)
	fmt.Printf("result: %s", result)
	if result != expectedResult {
		t.Error(fmt.Sprintf("%s is not %s", result, expectedResult))
	}
}
