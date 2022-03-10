package utils

import (
	"fmt"
	"testing"
)

func TestGetJsonBody(t *testing.T) {

	//a := map[string]string{
	//	"1":"2",
	//}
	//
	//b ,_:= json.Marshal(a)
	//
	//c := GetJsonBody(b)
	//
	//d := c["1"]
	//fmt.Printf("%v\n",d)
	//
	//var h string
	//h = d.(string)
	//
	//fmt.Printf("%+v\n",h)

	e := []string{"1","2","3"}
	var p []string

	for _,s := range e {
		fmt.Println(s)
		p = append(p,s)
	}

	fmt.Printf("%+v\n",p)
}

