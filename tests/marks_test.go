package delta_test

import (
	"testing"

	"github.com/GeoNet/delta/meta"
	"bytes"
	"fmt"
	"time"
	"io/ioutil"
	"encoding/json"
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
	Code             string
	Network          string
	Restricted       bool
	Start            time.Time
	End              time.Time
	Elevation float64
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
	l := len(marks) - 1

	b.WriteString(`{"type": "FeatureCollection","features": [`)

	for i, v := range marks {
		b.WriteString(`{"type":"Feature","geometry":{"type": "Point","coordinates": [`)
		b.WriteString(fmt.Sprintf("%f,%f", v.Longitude, v.Latitude))
		b.WriteString(`]},"properties":{`)
		b.WriteString(fmt.Sprintf("\"code\":\"%s\",", v.Code))
		n := net[v.Network]
		if n.External != "" {
			b.WriteString(fmt.Sprintf("\"network\":\"%s\",", n.External))
			b.WriteString(fmt.Sprintf("\"restricted\":%t,", n.Restricted))
		} else {
			b.WriteString(fmt.Sprintf("\"network\":\"%s\",", v.Network))
			b.WriteString(`"restricted":false`)
		}
		b.WriteString(fmt.Sprintf("\"name\":\"%s\",", v.Name))
		b.WriteString(fmt.Sprintf("\"datum\":\"%s\",", v.Datum))
		b.WriteString(fmt.Sprintf("\"elevation\":%f,", v.Elevation))
		b.WriteString(fmt.Sprintf("\"start\":\"%s\",", v.Start.UTC().Format(time.RFC3339)))
		b.WriteString(fmt.Sprintf("\"end\":\"%s\"", v.End.UTC().Format(time.RFC3339)))
		b.WriteString(`}}`)

		if i < l {
			b.WriteString(`,`)
		}
	}

	b.WriteString(`]}`)

	var mj markFeatures

	if err :=json.Unmarshal(b.Bytes(), &mj); err != nil {
		t.Error(err)
	}

	if len(mj.Features) != len(marks) {
		t.Errorf("len marks not the same as geojson features expected %d got %d", len(marks), len(mj.Features))
	}

	var found bool
	for _, v := range mj.Features {
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
