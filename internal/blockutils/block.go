package blockutils

// Enum for block names
const (
	WeatherName string = "WEATHER"
	DiskName    string = "DISK"
	PackName    string = "PACKAGES"
	TempName    string = "TEMPERATURE"
	VolName     string = "VOLUME"
	MediaName   string = "MEDIA"
	DateName    string = "DATE"
	TimeName    string = "TIME"
	BatteryName string = "BATTERY"
)

// Block represents a "block" of information in the bar
type Block struct {
	Name        string `json:"name"`
	Border      Color  `json:"border"`
	BorderLeft  int    `json:"border_left"`
	BorderRight int    `json:"border_right"`
	BorderTop   int    `json:"border_top"`
	Urgent      bool   `json:"urgent"`
	FullText    string `json:"full_text"`
}

// Marshalable is an interface allowing things to be marshalled into a byte
// array
type Marshalable interface {
	Marshal() []byte
}
