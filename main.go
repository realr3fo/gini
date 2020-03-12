package main

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
)

type event struct {
	ID          string `json:"ID"`
	Title       string `json:"Title"`
	Description string `json:"Description"`
}

type giniData struct {
	Gini float64 `json:"gini"`
	Data []int `json:"data"`
}

type allEvents []event

var events = allEvents{
	{
		ID:          "1",
		Title:       "Introduction to Golang",
		Description: "Come join us for a chance to learn how golang works and get to eventually try it out",
	},
}

func homeLink(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Welcome home!")
}

func createEvent(w http.ResponseWriter, r *http.Request) {
	var newEvent event
	reqBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		fmt.Fprintf(w, "Kindly enter data with the event title and description only in order to update")
	}

	json.Unmarshal(reqBody, &newEvent)
	events = append(events, newEvent)
	w.WriteHeader(http.StatusCreated)

	json.NewEncoder(w).Encode(newEvent)
}

func getOneEvent(w http.ResponseWriter, r *http.Request) {
	eventID := mux.Vars(r)["id"]

	for _, singleEvent := range events {
		if singleEvent.ID == eventID {
			json.NewEncoder(w).Encode(singleEvent)
		}
	}
}

func getAllEvents(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(events)
}

func updateEvent(w http.ResponseWriter, r *http.Request) {
	eventID := mux.Vars(r)["id"]
	var updatedEvent event

	reqBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		fmt.Fprintf(w, "Kindly enter data with the event title and description only in order to update")
	}
	json.Unmarshal(reqBody, &updatedEvent)

	for i, singleEvent := range events {
		if singleEvent.ID == eventID {
			singleEvent.Title = updatedEvent.Title
			singleEvent.Description = updatedEvent.Description
			events = append(events[:i], singleEvent)
			json.NewEncoder(w).Encode(singleEvent)
		}
	}
}

func deleteEvent(w http.ResponseWriter, r *http.Request) {
	eventID := mux.Vars(r)["id"]

	for i, singleEvent := range events {
		if singleEvent.ID == eventID {
			events = append(events[:i], events[i+1:]...)
			fmt.Fprintf(w, "The event with ID %v has been deleted successfully", eventID)
		}
	}
}

func getGini(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	entities, entityOk := r.URL.Query()["entity"]
	if !entityOk {
		http.Error(w, "Url Param 'entity' is missing", http.StatusInternalServerError)
		return
	}
	entity := entities[0]

	var unbounded = true
	var propertiesArr []string

	properties, propertyOk := r.URL.Query()["properties"]
	if propertyOk {
		unbounded = false
		propertiesArr = strings.Split(properties[0], ",")
	}
	fmt.Println(unbounded)
	fmt.Println(propertiesArr)

	if entity == "Q5" {
		giniArrData := []int{286,641,1078,1527,1981,2530,3325,4236,5163,6248}
		giniCoefficient := 0.23524
		giniResp := &giniData{
			Gini: giniCoefficient,
			Data: giniArrData,
		}
		json.NewEncoder(w).Encode(giniResp)
	} else {
		giniArrData := []int{100,300,600,1000,1500,2100,2800,3600,4500,5500}
		giniCoefficient := 0.12345
		giniResp := &giniData{
			Gini: giniCoefficient,
			Data: giniArrData,
		}
		json.NewEncoder(w).Encode(giniResp)
	}
}

func main() {
	//initEvents()
	router := mux.NewRouter().StrictSlash(true)
	router.HandleFunc("/", homeLink)
	router.HandleFunc("/event", createEvent).Methods("POST")
	router.HandleFunc("/events", getAllEvents).Methods("GET")
	router.HandleFunc("/events/{id}", getOneEvent).Methods("GET")
	router.HandleFunc("/events/{id}", updateEvent).Methods("PUT")
	router.HandleFunc("/events/{id}", deleteEvent).Methods("DELETE")

	router.HandleFunc("/api/gini", getGini).Methods("GET")
	port := os.Getenv("PORT")
	if port == "" {
		port = "8000" //localhost
	}
	fmt.Println(port)

	err := http.ListenAndServe(":"+port, router) //Launch the app, visit localhost:8000/api
	if err != nil {
		fmt.Print(err)
	}
}