package lde

import (
	"fmt"
	jwt "github.com/dgrijalva/jwt-go"
)

// SUDType describes the type of Seneye USB Device.
type SUDType int

const (
	// HomeSUD describes the home type Seneye USB Device..
	HomeSUD SUDType = 1

	// PondSUD describes the pond type Seneye USB Device..
	PondSUD SUDType = 2

	// ReefSUD describes the reef type Seneye USB Device..
	ReefSUD SUDType = 3
)

// String translates the SUD type to a string.
func (t SUDType) String() string {
	switch t {
	case HomeSUD:
		return "home"
	case PondSUD:
		return "pong"
	case ReefSUD:
		return "reef"
	default:
		return "unknown"
	}
}

// LDE is the Seneye Local Data Exchange object.
type LDE struct {
	// Version of the LDE protocol.
	Version string `json:"version"`
	// SUD describes the state of Seneye USB Device.
	SUD SUD `json:"SUD"`
}

// Valid implements jwt.Claims so we can make jwt parse the body.
func (l *LDE) Valid() error {
	return nil
}

// SUD describes the state of Seneye USB Device.
type SUD struct {
	// ID describes the serial number of the SUD.
	ID string `json:"id"`
	// Name is the user assigned name of the SUD.
	Name string `json:"name"`
	// Type describes the seneye device model type.
	Type SUDType `json:"type"`
	// Timestamp describes when the sample was taken (UNIX timestamp)
	Timestamp int64 `json:"TS"`
	// Data holds the readings from the SUD.
	Data Data `json:"data"`
}

// Data describes readings from the SUD.
type Data struct {
	// Status describes the condition of the SUD and any alert conditions.
	Status SUDStatus `json:"S"`
	// Temperature describes the water temperature in celsius degrees.
	Temperature float64 `json:"T"`
	// PH describes the water's PH.
	PH float64 `json:"P"`
	// NH3 describes the amount of free ammonia. (See: https://answers.seneye.com/en/water_chemistry/what_is_ammonia_NH3_NH4 )
	NH3 float64 `json:"N"`
	// Kelvin is the numeric Correlated Color Temperature value of the colour temperature in degrees Kelvin.
	// https://www.seneye.com/kelvin
	Kelvin float64 `json:"K"`
	// Lux describes the intensity of the light observed in the tank. ( https://en.wikipedia.org/wiki/Lux )
	Lux float64 `json:"L"`
	// PAR describes the photosynthetic active radiation is a measurement of light power between 400nm and 700nm.
	// ( https://answers.seneye.com/index.php?title=en/Aquarium_help/What_is_PAR_%26_PUR_%3F )
	PAR float64 `json:"A"`
}

// SUDStatus describes the condition of the SUD and any alert conditions.
type SUDStatus struct {
	// Water is 1 if the SUD is submerged in water, 0 otherwise.
	Water int `json:"W"`
	// Temperature is 0 if the temperature is within limits, 1 otherwise.
	Temperature int `json:"T"`
	// PH is 0 if the pH is within limits, 1 otherwise
	PH int `json:"P"`
	// NH3 is 0 if the free ammonia is within limits, 1 otherwise
	NH3 int `json:"N"`
	// Slide is 0 if the slide is correctly installed and unexpired, 1 otherwise.
	Slide int `json:"S"`
	// Kelvin is 0 if the Kelvin measurement is within limits, 1 otherwise.
	Kelvin int `json:"K"`
}

// FromRequestBody parses the LDE body.
func FromRequestBody(requestBody []byte, secrets map[string][]byte) (*LDE, error) {
	lde := &LDE{}
	_, err := jwt.ParseWithClaims(string(fixEncoding(requestBody)), lde, func(token *jwt.Token) (interface{}, error) {
		var secret []byte
		var ok bool
		if secret, ok = secrets[lde.SUD.ID]; !ok {
			// Try to fallback to a default key.
			if secret, ok = secrets[""]; !ok {
				return nil, fmt.Errorf("Unknown Seneye device ID: %q", lde.SUD.ID)
			}
		}
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}
		return secret, nil
	})
	return lde, err
}

// fixEncoding ensure the request is encoded data using base64 encoding with URL and filename safe
// alphabet expected by jwt, instead of the normal base64 encoding.
// https://datatracker.ietf.org/doc/rfc4648/
func fixEncoding(in []byte) []byte {
	for i := range in {
		switch in[i] {
		case '+':
			in[i] = '-'
		case '/':
			in[i] = '_'
		}
	}
	return in
}
