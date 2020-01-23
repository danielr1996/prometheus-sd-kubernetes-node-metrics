package prometheus

import (
	"encoding/json"
	"io/ioutil"
	"os"
)

type TargetsConfig struct{
	Targets []string`json:"targets"`
}

func WriteTargetsConfig(configFile string, targetsConfig []TargetsConfig){
	data, err := json.Marshal(targetsConfig)
	if err != nil{
		panic(err)
	}
	os.MkdirAll("/var/prometheus", os.ModePerm)
	err3 := ioutil.WriteFile(configFile, data, 0644)
	if err3 != nil {
		panic(err3)
	}
}