package main

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"math"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"
)

type giniData struct {
	Gini float64   `json:"gini"`
	Data []float64 `json:"data"`
	Entities [][]string `json:"entities"`
}

type EntityPropCount struct {
	Entity string `json:"-"`
	PropCount int `json:"-"`
}

type HeadVar struct {
	Var []string `json:"vars"`
}

type ItemBindingContent struct {
	Type  string `json:"type"`
	Value string `json:"value"`
}

type Item struct {
	ItemBinding ItemBindingContent `json:"item"`
}

type ResultVar struct {
	Bindings []Item `json:"bindings"`
}

type InstancesResult struct {
	Head   HeadVar   `json:"head"`
	Result ResultVar `json:"results"`
}

type PropertyCountBindingContent struct {
	DataType string `json:"datatype"`
	Type     string `json:"type"`
	Value    string `json:"value"`
}

type PropertyCount struct {
	PropertyCountBinding PropertyCountBindingContent `json:"propertyCount"`
}

type CountResultVar struct {
	Bindings []PropertyCount `json:"bindings"`
}

type CountResult struct {
	Head   HeadVar        `json:"head"`
	Result CountResultVar `json:"results"`
}

func homeLink(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Welcome home!")
}

func findMinAndMax(a []int) (min int, max int) {
	if len(a) == 0 {
		return 0, 0
	}
	min = a[0]
	max = a[0]
	for _, value := range a {
		if value < min {
			min = value
		}
		if value > max {
			max = value
		}
	}
	return min, max
}

func allCombination(set []string) (subsets [][]string) {
	length := uint(len(set))

	// Go through all possible combinations of objects
	// from 1 (only first object in subset) to 2^length (all objects in subset)
	for subsetBits := 1; subsetBits < (1 << length); subsetBits++ {
		var subset []string

		for object := uint(0); object < length; object++ {
			// checks if object is contained in subset
			// by checking if bit 'object' is set in subsetBits
			if (subsetBits>>object)&1 == 1 {
				// add object to subset
				subset = append(subset, set[object])
			}
		}
		// add subset to subsets
		subsets = append(subsets, subset)
	}
	return subsets
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
	if propertyOk && properties[0] != "[]"  && len(properties[0]) != 0 {
		unbounded = false
		trimmedString := strings.Trim(properties[0], "[]")
		propertiesArr = strings.Split(trimmedString, ",")
	}

	if unbounded {
		wikiDataQueryURL := fmt.Sprintf("https://query.wikidata.org/sparql?query=select%%3Fitem%%7B%%3Fitem%%20wdt" +
			"%%3AP31%%20wd%%3A%s%%7DLIMIT%%20500&format=json", entity)
		response, err := http.Get(wikiDataQueryURL)
		if err != nil {
			http.Error(w, "Error while query WikiData", http.StatusInternalServerError)
			return
		}
		decoder := json.NewDecoder(response.Body)
		var result InstancesResult
		err = decoder.Decode(&result)
		if err != nil {
			http.Error(w, "Error while decoding", http.StatusInternalServerError)
			return
		}
		var resultEntities []string
		for _, elem := range result.Result.Bindings {
			splitElem := strings.Split(elem.ItemBinding.Value, "/")
			entityID := splitElem[len(splitElem)-1]
			resultEntities = append(resultEntities, entityID)
		}

		var propertyCountData []EntityPropCount
		for _, elem := range resultEntities {
			wikiDataCountURL := fmt.Sprintf("https://query.wikidata.org/sparql?query=SELECT%%20(COUNT("+
				"DISTINCT(%%3Fp))%%20AS%%20%%3FpropertyCount)%%20%%7Bwd%%3A%s%%20%%3Fp%%20%%3Fo%%20.%%20FILTER("+
				"STRSTARTS(STR(%%3Fp)%%2C%%22http%%3A%%2F%%2Fwww.wikidata.org%%2Fprop%%2Fdirect%%2F%%22))"+
				"%%7D&format=json", elem)
			countResponse, err := http.Get(wikiDataCountURL)
			if err != nil {
				http.Error(w, "Error while query count WikiData", http.StatusInternalServerError)
				return
			}
			decoder := json.NewDecoder(countResponse.Body)
			var result CountResult
			err = decoder.Decode(&result)
			if err != nil {
				http.Error(w, "Error while decoding", http.StatusInternalServerError)
				return
			}
			strCount := result.Result.Bindings[0].PropertyCountBinding.Value
			intCount, _ := strconv.Atoi(strCount)
			elemPropCount := EntityPropCount{
				Entity:    elem,
				PropCount: intCount,
			}
			propertyCountData = append(propertyCountData, elemPropCount)
		}
		sort.Slice(propertyCountData, func(i, j int) bool {
			return propertyCountData[i].PropCount < propertyCountData[j].PropCount
		})
		n := len(propertyCountData)

		sum := 0
		for _, elem := range propertyCountData {
			sum += elem.PropCount
		}

		calculateTopSum := 0
		for idx, elem := range propertyCountData {
			calculateTopSum += (n + 1 - (idx + 1)) * elem.PropCount
		}

		rightBelowGiniCoef := n * sum
		rightTopGiniCoef := 2 * calculateTopSum
		rightGiniCoef := float64(rightTopGiniCoef) / float64(rightBelowGiniCoef)
		leftGiniCoef := float64(n+1) / float64(n)
		giniCoef := leftGiniCoef - rightGiniCoef
		chunkSize := float64(n) / float64(10)
		chunkSize = math.Ceil(chunkSize)


		var chunkedArray []int
		var groupedEntities [][]string
		cumSum := 0
		if chunkSize > 1 {
			var tmpGroupedEntities []string
			for idx, elem := range propertyCountData {
				if idx != 0 && math.Mod(float64(idx), chunkSize) == 0 {
					groupedEntities = append(groupedEntities, tmpGroupedEntities)
					tmpGroupedEntities = []string{}
					chunkedArray = append(chunkedArray, cumSum)
				}
				cumSum += elem.PropCount
				tmpGroupedEntities = append(tmpGroupedEntities, elem.Entity)
				if idx == len(propertyCountData)-1 {
					chunkedArray = append(chunkedArray, cumSum)
					groupedEntities = append(groupedEntities, tmpGroupedEntities)
				}
			}
		} else {
			for _, elem := range propertyCountData {
				groupedEntities = append(groupedEntities, []string{elem.Entity})
				cumSum += elem.PropCount
				chunkedArray = append(chunkedArray, cumSum)
			}
		}

		var normalizedCountData []float64
		min, max := findMinAndMax(chunkedArray)
		for _, elem := range chunkedArray {
			normalizedElem := (float64(elem) - float64(min)) / (float64(max) - float64(min))
			normalizedCountData = append(normalizedCountData, normalizedElem)
		}


		entitiesData := groupedEntities
		giniArrData := normalizedCountData
		giniCoefficient := giniCoef
		giniResp := &giniData{
			Gini: giniCoefficient,
			Data: giniArrData,
			Entities: entitiesData,
		}
		json.NewEncoder(w).Encode(giniResp)
	} else {
		propLength := len(propertiesArr)
		//var entitySet []map[string]int

		setComb := map[int][][]string{}
		combRes := allCombination(propertiesArr)
		for _, elem := range combRes {
			currLen := len(elem)
			_, ok := setComb[currLen]
			if ok {
				setComb[currLen] = append(setComb[currLen], elem)
			} else {
				setComb[currLen] = [][]string{elem}
			}
		}

		for propNum := propLength; propNum >= 0; propNum-- {
			arrCombs := setComb[propNum]
			strElem := fmt.Sprintf("https://query.wikidata.org/sparql?query=select%%20DISTINCT%%20%%3Fitem%%20%%7B%%7B%%3Fitem%%20wdt%%3AP31%%20wd%%3A%s%%3B%%7D", entity)
			for _, elem := range arrCombs {
				strElem += fmt.Sprintf("%%20UNION%%20")
				num := 0
				for _, insideElem := range elem {
					strElem += fmt.Sprintf("%%7B%%3Fitem%%20wdt%%3A%s%%20%%3F%d%%7D", insideElem, num)
					num += 1
				}
			}

			strElem += fmt.Sprintf("%%7D&format=json")
			fmt.Println(strElem)
			response, err := http.Get(strElem)
			if err != nil {
				http.Error(w, "Error while query WikiData", http.StatusInternalServerError)
				return
			}
			decoder := json.NewDecoder(response.Body)
			var result InstancesResult
			err = decoder.Decode(&result)
			if err != nil {
				http.Error(w, "Error while decoding", http.StatusInternalServerError)
				return
			}
			var resultEntities []string
			for _, elem := range result.Result.Bindings {
				splitElem := strings.Split(elem.ItemBinding.Value, "/")
				entityID := splitElem[len(splitElem)-1]
				resultEntities = append(resultEntities, entityID)
			}


		}
	}
}

func main() {
	//initEvents()
	router := mux.NewRouter().StrictSlash(true)
	router.HandleFunc("/", homeLink)

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
