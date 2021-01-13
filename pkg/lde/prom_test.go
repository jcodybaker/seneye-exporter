package lde

import (
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCollect(t *testing.T) {
	lde := &LDE{
		SUD: SUD{
			Name:      "example",
			ID:        "1234",
			Timestamp: 1610505992,
			Type:      HomeSUD,
			Data: Data{
				Status: SUDStatus{
					Water:       1,
					Temperature: 1,
				},
				Temperature: 21.3,
				PH:          7.0,
				NH3:         0.01,
				Kelvin:      100,
				Lux:         200,
				PAR:         300,
			},
		},
	}
	s := &LDEServer{
		lastLDEs: map[string]*LDE{
			"1234": lde,
		},
	}
	reg := prometheus.NewPedanticRegistry()
	reg.MustRegister(s)
	ts := httptest.NewServer(promhttp.HandlerFor(reg, promhttp.HandlerOpts{}))
	res, err := http.Get(ts.URL)
	require.NoError(t, err)
	b, err := ioutil.ReadAll(res.Body)
	res.Body.Close()
	require.NoError(t, err)
	fmt.Println(string(b))
	assert.Equal(t, `# HELP ammonia PPM Water NH3 free ammonia
# TYPE ammonia gauge
ammonia{id="1234",name="example",sud_type="home"} 0.01 1610505992000
# HELP light_kelvin Kelvin is the numeric Correlated Color Temperature value of the colour temperature in degrees Kelvin.
# TYPE light_kelvin gauge
light_kelvin{id="1234",name="example",sud_type="home"} 100 1610505992000
# HELP light_lux Lux describes the intensity of the light observed in the tank. 
# TYPE light_lux gauge
light_lux{id="1234",name="example",sud_type="home"} 200 1610505992000
# HELP light_par PAR describes the photosynthetic active radiation is a measurement of light power between 400nm and 700nm.
# TYPE light_par gauge
light_par{id="1234",name="example",sud_type="home"} 300 1610505992000
# HELP ph Water pH
# TYPE ph gauge
ph{id="1234",name="example",sud_type="home"} 7 1610505992000
# HELP seneye_status_ammonia Ammonia (NH3) is 0 if the free ammonia is within limits..
# TYPE seneye_status_ammonia gauge
seneye_status_ammonia{id="1234",name="example",sud_type="home"} 0 1610505992000
# HELP seneye_status_kelvin Kelvin is 0 if the Kelvin measurement is within limits.
# TYPE seneye_status_kelvin gauge
seneye_status_kelvin{id="1234",name="example",sud_type="home"} 0 1610505992000
# HELP seneye_status_ph PH is 0 if the pH is within limits.
# TYPE seneye_status_ph gauge
seneye_status_ph{id="1234",name="example",sud_type="home"} 0 1610505992000
# HELP seneye_status_slide Slide is 0 if the slide is correctly installed and unexpired, 1 otherwise.
# TYPE seneye_status_slide gauge
seneye_status_slide{id="1234",name="example",sud_type="home"} 0 1610505992000
# HELP seneye_status_temperature Temperature is 1 if the temperature is within limits.
# TYPE seneye_status_temperature gauge
seneye_status_temperature{id="1234",name="example",sud_type="home"} 1 1610505992000
# HELP seneye_status_water Water is 1 if the SUD is submerged in water, false otherwise.
# TYPE seneye_status_water gauge
seneye_status_water{id="1234",name="example",sud_type="home"} 1 1610505992000
# HELP temperature_celsius Water temperature in celsius
# TYPE temperature_celsius gauge
temperature_celsius{id="1234",name="example",sud_type="home"} 21.3 1610505992000
`, string(b))
}
