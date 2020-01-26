/*
 * To change this template, choose Tools | Templates
 * and open the template in the editor.
 */

package common

import (
	"encoding/json"
	"fmt"
	"github.com/howeyc/gopass"
	. "github.com/yotron/goConfigurableLogger"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strconv"
)

func ReadYAMLFile(path string, type_struct interface{}) []byte {
	err := yaml.Unmarshal(ReadFile(path), &type_struct)
	if err != nil {
		Error.Println("file malformed:", path, "Error:", err)
	}
	byte, _ := yaml.Marshal(type_struct)
	Debug.Println("YML: Type Structure created.")
	return byte
}

func ReadJSONFile(path string, type_struct interface{}) []byte {
	err := json.Unmarshal(ReadFile(path), &type_struct)
	if err != nil {
		Error.Println("File malformed:", path, "Error:", err)
	}
	byte, _ := json.Marshal(type_struct)
	Debug.Println("JSON: Type Structure created.")
	return byte
}

func ReadFile(path string) []byte {
	Debug.Println("File to Read:", path)
	file, err := ioutil.ReadFile(path)
	if err != nil {
		Error.Println("Could not read file:", path)
		panic("Could not read file: " + path)
	}
	return file
}

func HandleErrWithPanic(err error) {
	if err != nil {
		Error.Println(err)
		//panic(err)
	}
}

func WriteFile(filenname string, content []byte) {
	f, _ := os.Create(filenname)
	_, _ = f.Write(content)
	f.Sync()
}

func WriteStringFile(filenname string, content string) {
	f, _ := os.Create(filenname)
	_, _ = f.WriteString(content)
	f.Sync()
}

func GetPassword(conf ConfAuthentication) string {
	pwDefinition := GetDataByCluster(conf.Password)
	var pw string
	var err error
	if filepath.Ext(pwDefinition) == ".yml" {
		var confSecret confSecret
		Debug.Println("Password: Try read from yml-file")
		if err = confSecret.readSecretPasswordFile(pwDefinition); err != nil {
			panic("Could not read password file.")
		}
		if pw, err = confSecret.getPasswordByEnvExportType(); err != nil {
			panic("Could not read passwort from file. Error: " + err.Error())
		}
	} else if pwDefinition == "YOPRO_AUTH_PASSWORD" {
		Debug.Println("Password: Try read from environmnt variable ATLAS_KPI_PASSWORD")
		if pw = os.Getenv("YOPRO_AUTH_PASSWORD"); pw == "" { // from start.sh
			panic("ATLAS_KPI_PASSWORD not set")
		}
	} else if pwDefinition == "MANUALLY" { // for local start without start.sh
		Debug.Println("Password: Try read from manual input")
		pw = string(getInputPassword())
	} else {
		Debug.Println("Password: Try read clear password in config file")
		pw = pwDefinition
	}
	Debug.Println("Password set")
	return pw
}

func getInputPassword() []byte {
	fmt.Printf("Set Password for authentication: ")
	pass, err := gopass.GetPasswd()
	if err != nil {
		Error.Println(err)
	}
	return pass
}

func GetDataByCluster(conf ConfClusterType) string {
	Debug.Println("Cluster:", Cluster)
	Debug.Println("conf file:", conf)
	switch Cluster {
	case "localTest":
		return conf.LocalTest
	case "staging":
		return conf.Staging
	case "production":
		return conf.Production
	}
	Error.Println("No value found:", Cluster)
	return ""
}

func GetDataOfSlice(value reflect.Value, e string) (int, reflect.Value) {
	var strct reflect.Value
	for index, key := range value.MapKeys() {
		Debug.Println("Reflect Interface: ", key.Interface().(string), "Search string:", e)
		if key.Interface().(string) == e {
			strct := value.MapIndex(key)
			Debug.Println("Reflect Interface: ", key.Interface().(string), "Structure:", strct)
			return index, strct
		}
	}
	return -1, strct
}

func GetInt(intTXT string) (int64, error) {
	integer, err := strconv.ParseInt(intTXT, 10, 64)
	if err != nil {
		Error.Println("Could not convert string:", intTXT, "Error:", err)
	}
	return integer, err
}

func GetIntOnly(intTXT string) int64 {
	integer, err := GetInt(intTXT)
	if err != nil {
		panic("Could not convert string.")
	}
	return integer
}

func GetGoCommandValue(metricJSON string, job ConfJobs) string {
	Debug.Println(job.Sid, "Metric JSON:", metricJSON)
	out, err := exec.Command("go", "run", createGoCommand(metricJSON, job), metricJSON).Output()
	if err != nil {
		Error.Println(job.Sid, "Could not run code. Error:", err)
		panic("Could not run code")
	} else {
		Debug.Println(job.Sid, "Output after function:", out)
		return string(out)
	}
}

func createGoCommand(metricJSON string, job ConfJobs) string {
	code := "package main\n" +
		"import (\n" +
		"\"fmt\"\n" +
		getImports(job) +
		")\n" +
		"func main() {\n" +
		"queryEntry := " + strconv.Quote(metricJSON) + "\n" +
		"var queryResult string\n" +
		job.GoCommand + "\n" +
		"fmt.Print(queryResult) }"
	filepath := "tmp/" + job.Sid + ".go"
	WriteFile(filepath, []byte(code))
	Debug.Println(job.Sid, "Code to run:", "go", "run", filepath)
	return filepath
}

func getImports(job ConfJobs) string {
	var imports string
	for _, txt := range job.GoCommandImports {
		imports = imports + strconv.Quote(txt) + "\n"
	}
	return imports
}
