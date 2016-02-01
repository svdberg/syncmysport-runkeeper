package sync

import (
	"fmt"
	"testing"
)

func TestMakeDbStringHerokuCompliant(t *testing.T) {
	herokuDbString := "mysql://b932bxxx43b903:83f507xx@eu-cdbr-west-01.cleardb.com/heroku_e6b0xx3a037c2d6?reconnect=true"
	expectedResult := "b932bxxx43b903:83f507xx@tcp(eu-cdbr-west-01.cleardb.com:3306)/heroku_e6b0xx3a037c2d6"

	result := makeDbStringHerokuCompliant(herokuDbString)
	fmt.Printf("result: %s", result)
	if result != expectedResult {
		t.Error(fmt.Sprintf("%s is not %s", result, expectedResult))
	}
}
