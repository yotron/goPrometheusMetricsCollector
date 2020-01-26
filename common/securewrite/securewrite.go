package securewrite

import (
	"encoding/json"
	. "github.com/yotron/goConfigurableLogger"
	"io/ioutil"
	"os"
	"sync"
)

type Persistor struct {
	path     string
	keyValue map[string]interface{}
	mux      sync.Mutex
}

func (persist *Persistor) InitConfig(filePath string) {
	Info.Println("Initiate Config")
	persist.path = filePath
	persist.ReadFile()
}

func (persist *Persistor) ReadFile() {
	Debug.Println("SecuredFileWriting: Read file secured")
	err := json.Unmarshal(readFile(persist.path), &persist.keyValue)
	if err != nil {
		panic("Could not unmarshal file: " + persist.path)
	}
	Info.Println("Read: File Status:", persist.path)
}

func (persist *Persistor) Read(sid string) interface{} {
	Debug.Println(sid, "SecuredRead value", persist.keyValue[sid])
	return persist.keyValue[sid]
}

func (persist *Persistor) Write(sid string, value interface{}) {
	Debug.Println("SecuredFileWriting: Write file secured")
	defer persist.mux.Unlock()
	persist.mux.Lock()
	Debug.Println(sid, "Write: File Status:", persist.path, "Wert:", value)
	persist.keyValue[sid] = value
	persist.writeFile()
}

func readFile(path string) []byte {
	persist, err := ioutil.ReadFile(path)
	if err != nil {
		panic("Could not read file: " + path)
	}
	return persist
}

func (persist *Persistor) writeFile() {
	f, err := os.Create(persist.path)
	val, err := json.Marshal(persist.keyValue)
	if err != nil {
		Error.Println("Error during Mashalling json: ", persist.keyValue, "Error:", err)
	}
	_, err = f.Write(val)
	if err != nil {
		panic("Could not write file: " + persist.path)
	}
	f.Sync()
}
