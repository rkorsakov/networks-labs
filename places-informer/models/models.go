package models

type GraphHopperResponse struct {
	Hits []struct {
		OsmID    int    `json:"osm_id"`
		OsmType  string `json:"osm_type"`
		Country  string `json:"country"`
		OsmKey   string `json:"osm_key"`
		City     string `json:"city,omitempty"`
		OsmValue string `json:"osm_value"`
		Postcode string `json:"postcode,omitempty"`
		Name     string `json:"name"`
		Point    struct {
			Lng float64 `json:"lng"`
			Lat float64 `json:"lat"`
		} `json:"point"`
		Extent      []float64 `json:"extent,omitempty"`
		Street      string    `json:"street,omitempty"`
		Housenumber string    `json:"housenumber,omitempty"`
	} `json:"hits"`
	Took int `json:"took"`
}

type Location struct {
	Name      string
	Country   string
	City      string
	Address   string
	Latitude  float64
	Longitude float64
}

type OpenWeatherResponse struct {
	Coord struct {
		Lon float64 `json:"lon"`
		Lat float64 `json:"lat"`
	} `json:"coord"`
	Weather []struct {
		ID          int    `json:"id"`
		Main        string `json:"main"`
		Description string `json:"description"`
		Icon        string `json:"icon"`
	} `json:"weather"`
	Base string `json:"base"`
	Main struct {
		Temp      float64 `json:"temp"`
		FeelsLike float64 `json:"feels_like"`
		TempMin   float64 `json:"temp_min"`
		TempMax   float64 `json:"temp_max"`
		Pressure  int     `json:"pressure"`
		Humidity  int     `json:"humidity"`
		SeaLevel  int     `json:"sea_level"`
		GrndLevel int     `json:"grnd_level"`
	} `json:"main"`
	Visibility int `json:"visibility"`
	Wind       struct {
		Speed float64 `json:"speed"`
		Deg   int     `json:"deg"`
		Gust  float64 `json:"gust"`
	} `json:"wind"`
	Rain struct {
		OneH float64 `json:"1h"`
	} `json:"rain"`
	Clouds struct {
		All int `json:"all"`
	} `json:"clouds"`
	Dt  int `json:"dt"`
	Sys struct {
		Type    int    `json:"type"`
		ID      int    `json:"id"`
		Country string `json:"country"`
		Sunrise int    `json:"sunrise"`
		Sunset  int    `json:"sunset"`
	} `json:"sys"`
	Timezone int    `json:"timezone"`
	ID       int    `json:"id"`
	Name     string `json:"name"`
	Cod      int    `json:"cod"`
}

type Weather struct {
	Temp      float64
	FeelsLike float64
}

type PlacesResponse struct {
	Features []struct {
		Properties struct {
			Xid   string  `json:"xid"`
			Name  string  `json:"name"`
			Rate  float64 `json:"rate"`
			Kinds string  `json:"kinds"`
		} `json:"properties"`
	} `json:"features"`
}

type PlaceInfo struct {
	ID      string
	Name    string
	Rating  float64
	Kinds   string
	Details *PlaceDetailsResponse
}

type PlaceDetailsResponse struct {
	Kinds   string `json:"kinds"`
	Sources struct {
		Geometry   string   `json:"geometry"`
		Attributes []string `json:"attributes"`
	} `json:"sources"`
	Bbox struct {
		LatMax float64 `json:"lat_max"`
		LatMin float64 `json:"lat_min"`
		LonMax float64 `json:"lon_max"`
		LonMin float64 `json:"lon_min"`
	} `json:"bbox"`
	Point struct {
		Lon float64 `json:"lon"`
		Lat float64 `json:"lat"`
	} `json:"point"`
	Osm       string `json:"osm"`
	Otm       string `json:"otm"`
	Xid       string `json:"xid"`
	Name      string `json:"name"`
	Wikipedia string `json:"wikipedia"`
	Image     string `json:"image"`
	Wikidata  string `json:"wikidata"`
	Rate      string `json:"rate"`
	Info      struct {
		Descr     string `json:"descr"`
		Image     string `json:"image"`
		ImgWidth  int    `json:"img_width"`
		Src       string `json:"src"`
		SrcID     int    `json:"src_id"`
		ImgHeight int    `json:"img_height"`
	} `json:"info"`
}
