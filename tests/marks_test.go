package delta_test

import (
	"bytes"
	"encoding/json"
	"github.com/GeoNet/delta/meta"
	"io/ioutil"
	"testing"
	"time"
)

type markFeatures struct {
	Features []markFeature
}

type markFeature struct {
	Properties markProperties
	Geometry   geometry
}

type geometry struct {
	Coordinates []float64
}

type markProperties struct {
	Code       string    `json:"code"`
	Network    string    `json:"network"`
	Restricted bool      `json:"restricted"`
	Name       string    `json:"name"`
	Datum      string    `json:"datum"`
	Elevation  float64   `json:"elevation"`
	Start      time.Time `json:"start"`
	End        time.Time `json:"end"`
}

type GeoJsonFeatureCollection struct {
	Type     string    `json:"type"`
	Features []Feature `json:"features"`
}

type Feature struct {
	Type       string          `json:"type"`
	Geometry   FeatureGeometry `json:"geometry"`
	Properties markProperties  `json:"properties"`
}

type FeatureGeometry struct {
	Type        string    `json:"type"`
	Coordinates []float64 `json:"coordinates"`
}

func TestMarks(t *testing.T) {

	var marks meta.MarkList
	t.Log("Load network marks file")
	{
		if err := meta.LoadList("../network/marks.csv", &marks); err != nil {
			t.Fatal(err)
		}
	}

	for i := 0; i < len(marks); i++ {
		for j := i + 1; j < len(marks); j++ {
			if marks[i].Code == marks[j].Code {
				t.Errorf("mark duplication: " + marks[i].Code)
			}
		}
	}

}

// TestMarksGeoJSON creates a GeoJSON file of Mark locations.
func TestMarksGeoJSON(t *testing.T) {
	var networks meta.NetworkList

	if err := meta.LoadList("../network/networks.csv", &networks); err != nil {
		t.Error(err)
	}

	var net = make(map[string]meta.Network)

	for _, v := range networks {
		net[v.Code] = v
	}

	var marks meta.MarkList
	if err := meta.LoadList("../network/marks.csv", &marks); err != nil {
		t.Error(err)
	}

	if len(marks) == 0 {
		t.Error("zero length mark list.")
	}

	var b bytes.Buffer
	allFeatures := make([]Feature, 0)

	for _, v := range marks {
		feature := Feature{Type: "Feature"}
		featureGeo := FeatureGeometry{Type: "Point"}
		siteProp := markProperties{}
		featureGeo.Coordinates = make([]float64, 2)
		featureGeo.Coordinates[0] = v.Longitude
		featureGeo.Coordinates[1] = v.Latitude
		siteProp.Code = v.Code
		siteProp.Name = v.Name
		siteProp.Start = v.Start.UTC()
		siteProp.End = v.End.UTC()
		siteProp.Datum = v.Datum

		n := net[v.Network]
		if n.External != "" {
			siteProp.Network = n.External
			siteProp.Restricted = n.Restricted
		} else {
			siteProp.Network = v.Network
			siteProp.Restricted = false
		}
		feature.Geometry = featureGeo
		feature.Properties = siteProp
		allFeatures = append(allFeatures, feature)
	}

	outputJson := GeoJsonFeatureCollection{
		Type:     "FeatureCollection",
		Features: allFeatures,
	}

	jsonBytes, err := json.MarshalIndent(outputJson, "", "\t")
	if err != nil {
		t.Fatal(err)
	}

	b.Write(jsonBytes)

	if len(allFeatures) != len(marks) {
		t.Errorf("len marks not the same as geojson features expected %d got %d", len(marks), len(allFeatures))
	}

	var found bool
	for _, v := range allFeatures {
		if v.Properties.Code == "TAUP" {
			found = true
			if v.Properties.Network != "LI" {
				t.Errorf("expected network LI got %s", v.Properties.Network)
			}
			if v.Properties.Restricted {
				t.Error("TAUP should not be restricted")
			}
			//	TODO start end elevation location.
		}
	}

	if !found {
		t.Error("expected to find mark TAUP")
	}

	if err := ioutil.WriteFile("../data/gnss/marks.geojson", b.Bytes(), 0644); err != nil {
		t.Error(err)
	}
}

// TestStationsGeoJSON creates a GeoJSON file of stations locations.
func TestStationsGeoJSON(t *testing.T) {
	var networks meta.NetworkList

	if err := meta.LoadList("../network/networks.csv", &networks); err != nil {
		t.Error(err)
	}

	var net = make(map[string]meta.Network)

	for _, v := range networks {
		net[v.Code] = v
	}

	var stations meta.StationList
	if err := meta.LoadList("../network/stations.csv", &stations); err != nil {
		t.Error(err)
	}

	if len(stations) == 0 {
		t.Error("zero length station list.")
	}

	var b bytes.Buffer
	allFeatures := make([]Feature, 0)

	for _, v := range stations {
		feature := Feature{Type: "Feature"}
		featureGeo := FeatureGeometry{Type: "Point"}
		siteProp := markProperties{}
		featureGeo.Coordinates = make([]float64, 2)
		featureGeo.Coordinates[0] = v.Longitude
		featureGeo.Coordinates[1] = v.Latitude
		siteProp.Code = v.Code
		siteProp.Name = v.Name
		siteProp.Start = v.Start.UTC()
		siteProp.End = v.End.UTC()
		siteProp.Datum = v.Datum

		n := net[v.Network]
		if n.External != "" {
			siteProp.Network = n.External
			siteProp.Restricted = n.Restricted
		} else {
			siteProp.Network = v.Network
			siteProp.Restricted = false
		}
		feature.Geometry = featureGeo
		feature.Properties = siteProp
		allFeatures = append(allFeatures, feature)
	}

	outputJson := GeoJsonFeatureCollection{
		Type:     "FeatureCollection",
		Features: allFeatures,
	}

	jsonBytes, err := json.MarshalIndent(outputJson, "", "\t")
	if err != nil {
		t.Fatal(err)
	}

	b.Write(jsonBytes)

	if len(allFeatures) != len(stations) {
		t.Errorf("len stations not the same as geojson features expected %d got %d", len(stations), len(allFeatures))
	}

	if err := ioutil.WriteFile("../data/gnss/stations.geojson", b.Bytes(), 0644); err != nil {
		t.Error(err)
	}
}
