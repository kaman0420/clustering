package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"net/http"
	"reflect"

	"github.com/gorilla/mux"
)

type site struct {
	ActiveDate int `json:"activeDate"`
	Coverage   struct {
		Raw struct {
			Lat             float64 `json:"lat"`
			Lng             float64 `json:"lng"`
			LocationCovered bool    `json:"locationCovered"`
			Margins         []int   `json:"margins"`
		} `json:"raw"`
		Processed struct {
			Coverage string `json:"coverage"`
		} `json:"processed"`
	} `json:"coverage"`
	Customer struct {
		Name string `json:"name"`
		ABN  string `json:"ABN"`
	} `json:"customer"`
	Environment string `json:"environment"`
	Location    struct {
		Lat float64 `json:"lat"`
		Lng float64 `json:"lng"`
	} `json:"location"`
	NumberOfDevices int `json:"numberOfDevices"`
	Project         struct {
		Cname      string `json:"Cname"`
		Pcode      string `json:"Pcode"`
		SalesLead  string `json:"SalesLead"`
		App        string `json:"App"`
		Cpriority  string `json:"Cpriority"`
		Location   string `json:"Location"`
		BudgetLink int    `json:"BudgetLink"`
		Revenue    int    `json:"Revenue"`
	} `json:"project"`
	Radius   int    `json:"radius"`
	SiteName string `json:"siteName"`
}

type tile struct {
	A struct {
		Lat float64 `json:"lat"`
		Lng float64 `json:"lng"`
	} `json:"A"`
	B struct {
		Lat float64 `json:"lat"`
		Lng float64 `json:"lng"`
	} `json:"B"`
	C struct {
		Lat float64 `json:"lat"`
		Lng float64 `json:"lng"`
	} `json:"C"`
	D struct {
		Lat float64 `json:"lat"`
		Lng float64 `json:"lng"`
	} `json:"D"`
	Site []interface{} `json:"sites"`
}

type point struct {
	X float64
	Y float64
}

type cluster struct {
	CenterPoint point
	Site        []interface{} `json:"sites"`
}

func createSite(w http.ResponseWriter, r *http.Request) {
	var newSite site
	var sites = []site{}

	reqBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		fmt.Fprintf(w, "Kindly enter data at least one data in order to update")
	}
	json.Unmarshal(reqBody, &newSite)
	sites = append(sites, newSite)
	w.WriteHeader(http.StatusCreated)

	json.NewEncoder(w).Encode(newSite)
	fmt.Println("created a site successfully!")
}

func processRequests(w http.ResponseWriter, r *http.Request) {
	var newTile tile
	var tiles = []tile{}
	var clusterSet = []cluster{}

	reqBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		fmt.Fprintf(w, "Kindly enter data at least one data in order to update")
	}
	json.Unmarshal(reqBody, &newTile)
	tiles = append(tiles, newTile)
	w.WriteHeader(http.StatusCreated)

	fmt.Println("")
	fmt.Println("Successfully created a tile with these points which are clustered as:")

	for _, selectedSite := range newTile.Site {

		activeSite := selectedSite.(map[string]interface{})
		// fmt.Printf("%+v\n", extractSitePoint(activeSite))

		if siteIsInExistingCluster(activeSite, clusterSet) == true {
			selectedCluster := whichCluster(activeSite, clusterSet)
			clusterSet = removeCluster(selectedCluster, clusterSet)
			selectedCluster.Site = append(selectedCluster.Site, activeSite)
			clusterSet = append(clusterSet, selectedCluster)
			// fmt.Println("selectedCluster: ")
			// printCluster(selectedCluster)

		} else {
			newCluster := createCluster(activeSite)
			// fmt.Println("New cluster is created:")
			// printCluster(newCluster)
			clusterSet = append(clusterSet, newCluster)
		}
	}
	fmt.Println("Final cluster:")
	fmt.Println(clusterSet)
	json.NewEncoder(w).Encode(clusterSet)
}

func printCluster(c cluster) {
	fmt.Printf("%p: %v\n\n", &c, &c)
}

func removeCluster(c cluster, clusterSet []cluster) []cluster {
	foundIndex := findIndex(c, clusterSet)
	clusterSet = append(clusterSet[:foundIndex], clusterSet[foundIndex+1:]...)
	return clusterSet
}

func findIndex(c cluster, clusterSet []cluster) int {
	var targetIndex int
	for index, selectiveCluster := range clusterSet {
		if selectiveCluster.CenterPoint.X == c.CenterPoint.X && selectiveCluster.CenterPoint.Y == c.CenterPoint.Y {
			targetIndex = index
		}
	}
	return targetIndex
}

func siteIsInExistingCluster(selectedSite map[string]interface{}, clusterSet []cluster) bool {
	isInThisCluster := false

	for _, receivingCluster := range clusterSet {

		extractedPoint := extractSitePoint(selectedSite)

		initialPoint := receivingCluster.CenterPoint

		distance := distance(initialPoint, extractedPoint)

		if compareDistance(distance, 4) == true {
			isInThisCluster = true
		}
	}
	return isInThisCluster
}

func whichCluster(selectedSite map[string]interface{}, clusterSet []cluster) cluster {
	var selectedCluster cluster

	for _, receivingCluster := range clusterSet {

		extractedPoint := extractSitePoint(selectedSite)

		initialPoint := receivingCluster.CenterPoint

		distance := distance(initialPoint, extractedPoint)

		if compareDistance(distance, 4) == true {
			selectedCluster = receivingCluster
		}
	}
	return selectedCluster
}

func createCluster(selectedSite map[string]interface{}) cluster {
	extractedPoint := extractSitePoint(selectedSite)
	newCluster := cluster{
		CenterPoint: point{
			X: extractedPoint.X,
			Y: extractedPoint.Y,
		},
	}
	newCluster.Site = append(newCluster.Site, selectedSite)
	return newCluster
}

func extractSitePoint(selectedSite map[string]interface{}) point {
	valueOfLocation := reflect.ValueOf(selectedSite["location"])
	locationIsInterface := valueOfLocation.Interface().(map[string]interface{})
	point := point{locationIsInterface["lat"].(float64), locationIsInterface["lng"].(float64)}
	return point
}

func distance(p1, p2 point) float64 {
	first := math.Pow(float64(p2.X-p1.X), 2)
	second := math.Pow(float64(p2.Y-p1.Y), 2)
	return math.Sqrt(first + second)
}

func compareDistance(distance float64, radius float64) bool {
	var result bool
	if distance < radius {
		result = true
	} else {
		result = false
	}
	return result
}

func main() {
	router := mux.NewRouter().StrictSlash(true)
	router.HandleFunc("/process", processRequests).Methods("POST")
	log.Fatal(http.ListenAndServe(":8080", router))
}
