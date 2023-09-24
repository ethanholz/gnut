package main

import (
	"errors"
	"flag"
	"html/template"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	nut "github.com/robbiet480/go.nut"
)

var (
	summaries map[string]UPSSummary
	nutHost   string
)

type UPSSummary struct {
	Name          string
	Status        string
	BatteryCharge int64
	Description   string
	Variables     []nut.Variable
}

func IndexHandler(client *nut.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := generateMap(client)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}

		t := template.Must(template.ParseFiles("base.html.tmpl", "index.html.tmpl"))
		t.ExecuteTemplate(w, "base", summaries)
	}
}

func UPSHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	name := vars["ups"]
	if _, ok := summaries[name]; !ok {
		http.Redirect(w, r, "/", http.StatusSeeOther)
	}
	summary := summaries[name]
	t := template.Must(template.ParseFiles("base.html.tmpl", "item.html.tmpl"))
	t.ExecuteTemplate(w, "base", summary)
}

func generateMap(client *nut.Client) error {
	upsList, listErr := client.GetUPSList()
	if listErr != nil {
		return listErr
	}
	if len(upsList) == 0 {
		return errors.New("no ups found")
	}

	for _, ups := range upsList {
		summary := UPSSummary{}
		summary.Name = ups.Name
		summary.Description = ups.Description
		summary.Variables = ups.Variables
		for _, upsVar := range ups.Variables {
			switch upsVar.Name {
			case "ups.status":
				summary.Status = upsVar.Value.(string)
			case "battery.charge":
				summary.BatteryCharge = upsVar.Value.(int64)
			}
		}
		summaries[ups.Name] = summary
	}
	return nil
}

// This example connects to NUT, authenticates and returns the first UPS listed.
func main() {
	nutHost := flag.String("client", os.Getenv("NUT_HOST"), "NUT client")
	key := flag.String("key", os.Getenv("NUT_KEY"), "the NUT server password/key")
	flag.Parse()
	if *nutHost == "" {
		log.Fatal("NUT_HOST not set")
	}
	if *key == "" {
		log.Fatal("NUT_KEY not set")
	}
	client, connectErr := nut.Connect(*nutHost)
	if connectErr != nil {
		log.Fatal(connectErr)
	}
	_, authenticationError := client.Authenticate("admin", *key)
	if authenticationError != nil {
		log.Println("Authentication error")
		log.Fatal(authenticationError)
	}
	summaries = make(map[string]UPSSummary)
	err := generateMap(&client)
	if err != nil {
		panic(err)
	}
	r := mux.NewRouter()
	r.HandleFunc("/", IndexHandler(&client))
	r.HandleFunc("/{ups}", UPSHandler)
	err = http.ListenAndServe(":8080", r)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
