package weather

// weather is a daemon that will download weather data from accuweather once daily at:
//	12 PM on Sunday and Saturday
//	1 PM on Monday - Friday

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"mariners/db"
	"math"
	"net/http"
	"net/url"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

// accuWeather is a []struct that can store the unmarshalled results of the accuweather current conditions API
/* type accuWeather []struct {
	LocalObservationDateTime string      `json:"LocalObservationDateTime"`
	EpochTime                int         `json:"EpochTime"`
	WeatherText              string      `json:"WeatherText"`
	WeatherIcon              int         `json:"WeatherIcon"`
	HasPrecipitation         bool        `json:"HasPrecipitation"`
	PrecipitationType        interface{} `json:"PrecipitationType"`
	IsDayTime                bool        `json:"IsDayTime"`
	Temperature              struct {
		Metric struct {
			Value    float64 `json:"Value"`
			Unit     string  `json:"Unit"`
			UnitType int     `json:"UnitType"`
		} `json:"Metric"`
		Imperial struct {
			Value    float64 `json:"Value"`
			Unit     string  `json:"Unit"`
			UnitType int     `json:"UnitType"`
		} `json:"Imperial"`
	} `json:"Temperature"`
	RealFeelTemperature struct {
		Metric struct {
			Value    float64 `json:"Value"`
			Unit     string  `json:"Unit"`
			UnitType int     `json:"UnitType"`
			Phrase   string  `json:"Phrase"`
		} `json:"Metric"`
		Imperial struct {
			Value    float64 `json:"Value"`
			Unit     string  `json:"Unit"`
			UnitType int     `json:"UnitType"`
			Phrase   string  `json:"Phrase"`
		} `json:"Imperial"`
	} `json:"RealFeelTemperature"`
	RealFeelTemperatureShade struct {
		Metric struct {
			Value    float64 `json:"Value"`
			Unit     string  `json:"Unit"`
			UnitType int     `json:"UnitType"`
			Phrase   string  `json:"Phrase"`
		} `json:"Metric"`
		Imperial struct {
			Value    float64 `json:"Value"`
			Unit     string  `json:"Unit"`
			UnitType int     `json:"UnitType"`
			Phrase   string  `json:"Phrase"`
		} `json:"Imperial"`
	} `json:"RealFeelTemperatureShade"`
	RelativeHumidity       int `json:"RelativeHumidity"`
	IndoorRelativeHumidity int `json:"IndoorRelativeHumidity"`
	DewPoint               struct {
		Metric struct {
			Value    float64 `json:"Value"`
			Unit     string  `json:"Unit"`
			UnitType int     `json:"UnitType"`
		} `json:"Metric"`
		Imperial struct {
			Value    float64 `json:"Value"`
			Unit     string  `json:"Unit"`
			UnitType int     `json:"UnitType"`
		} `json:"Imperial"`
	} `json:"DewPoint"`
	Wind struct {
		Direction struct {
			Degrees   int    `json:"Degrees"`
			Localized string `json:"Localized"`
			English   string `json:"English"`
		} `json:"Direction"`
		Speed struct {
			Metric struct {
				Value    float64 `json:"Value"`
				Unit     string  `json:"Unit"`
				UnitType int     `json:"UnitType"`
			} `json:"Metric"`
			Imperial struct {
				Value    float64 `json:"Value"`
				Unit     string  `json:"Unit"`
				UnitType int     `json:"UnitType"`
			} `json:"Imperial"`
		} `json:"Speed"`
	} `json:"Wind"`
	WindGust struct {
		Speed struct {
			Metric struct {
				Value    float64 `json:"Value"`
				Unit     string  `json:"Unit"`
				UnitType int     `json:"UnitType"`
			} `json:"Metric"`
			Imperial struct {
				Value    float64 `json:"Value"`
				Unit     string  `json:"Unit"`
				UnitType int     `json:"UnitType"`
			} `json:"Imperial"`
		} `json:"Speed"`
	} `json:"WindGust"`
	UVIndex     int    `json:"UVIndex"`
	UVIndexText string `json:"UVIndexText"`
	Visibility  struct {
		Metric struct {
			Value    float64 `json:"Value"`
			Unit     string  `json:"Unit"`
			UnitType int     `json:"UnitType"`
		} `json:"Metric"`
		Imperial struct {
			Value    float64 `json:"Value"`
			Unit     string  `json:"Unit"`
			UnitType int     `json:"UnitType"`
		} `json:"Imperial"`
	} `json:"Visibility"`
	ObstructionsToVisibility string `json:"ObstructionsToVisibility"`
	CloudCover               int    `json:"CloudCover"`
	Ceiling                  struct {
		Metric struct {
			Value    float64 `json:"Value"`
			Unit     string  `json:"Unit"`
			UnitType int     `json:"UnitType"`
		} `json:"Metric"`
		Imperial struct {
			Value    float64 `json:"Value"`
			Unit     string  `json:"Unit"`
			UnitType int     `json:"UnitType"`
		} `json:"Imperial"`
	} `json:"Ceiling"`
	Pressure struct {
		Metric struct {
			Value    float64 `json:"Value"`
			Unit     string  `json:"Unit"`
			UnitType int     `json:"UnitType"`
		} `json:"Metric"`
		Imperial struct {
			Value    float64 `json:"Value"`
			Unit     string  `json:"Unit"`
			UnitType int     `json:"UnitType"`
		} `json:"Imperial"`
	} `json:"Pressure"`
	PressureTendency struct {
		LocalizedText string `json:"LocalizedText"`
		Code          string `json:"Code"`
	} `json:"PressureTendency"`
	Past24HourTemperatureDeparture struct {
		Metric struct {
			Value    float64 `json:"Value"`
			Unit     string  `json:"Unit"`
			UnitType int     `json:"UnitType"`
		} `json:"Metric"`
		Imperial struct {
			Value    float64 `json:"Value"`
			Unit     string  `json:"Unit"`
			UnitType int     `json:"UnitType"`
		} `json:"Imperial"`
	} `json:"Past24HourTemperatureDeparture"`
	ApparentTemperature struct {
		Metric struct {
			Value    float64 `json:"Value"`
			Unit     string  `json:"Unit"`
			UnitType int     `json:"UnitType"`
		} `json:"Metric"`
		Imperial struct {
			Value    float64 `json:"Value"`
			Unit     string  `json:"Unit"`
			UnitType int     `json:"UnitType"`
		} `json:"Imperial"`
	} `json:"ApparentTemperature"`
	WindChillTemperature struct {
		Metric struct {
			Value    float64 `json:"Value"`
			Unit     string  `json:"Unit"`
			UnitType int     `json:"UnitType"`
		} `json:"Metric"`
		Imperial struct {
			Value    float64 `json:"Value"`
			Unit     string  `json:"Unit"`
			UnitType int     `json:"UnitType"`
		} `json:"Imperial"`
	} `json:"WindChillTemperature"`
	WetBulbTemperature struct {
		Metric struct {
			Value    float64 `json:"Value"`
			Unit     string  `json:"Unit"`
			UnitType int     `json:"UnitType"`
		} `json:"Metric"`
		Imperial struct {
			Value    float64 `json:"Value"`
			Unit     string  `json:"Unit"`
			UnitType int     `json:"UnitType"`
		} `json:"Imperial"`
	} `json:"WetBulbTemperature"`
	Precip1Hr struct {
		Metric struct {
			Value    float64 `json:"Value"`
			Unit     string  `json:"Unit"`
			UnitType int     `json:"UnitType"`
		} `json:"Metric"`
		Imperial struct {
			Value    float64 `json:"Value"`
			Unit     string  `json:"Unit"`
			UnitType int     `json:"UnitType"`
		} `json:"Imperial"`
	} `json:"Precip1hr"`
	PrecipitationSummary struct {
		Precipitation struct {
			Metric struct {
				Value    float64 `json:"Value"`
				Unit     string  `json:"Unit"`
				UnitType int     `json:"UnitType"`
			} `json:"Metric"`
			Imperial struct {
				Value    float64 `json:"Value"`
				Unit     string  `json:"Unit"`
				UnitType int     `json:"UnitType"`
			} `json:"Imperial"`
		} `json:"Precipitation"`
		PastHour struct {
			Metric struct {
				Value    float64 `json:"Value"`
				Unit     string  `json:"Unit"`
				UnitType int     `json:"UnitType"`
			} `json:"Metric"`
			Imperial struct {
				Value    float64 `json:"Value"`
				Unit     string  `json:"Unit"`
				UnitType int     `json:"UnitType"`
			} `json:"Imperial"`
		} `json:"PastHour"`
		Past3Hours struct {
			Metric struct {
				Value    float64 `json:"Value"`
				Unit     string  `json:"Unit"`
				UnitType int     `json:"UnitType"`
			} `json:"Metric"`
			Imperial struct {
				Value    float64 `json:"Value"`
				Unit     string  `json:"Unit"`
				UnitType int     `json:"UnitType"`
			} `json:"Imperial"`
		} `json:"Past3Hours"`
		Past6Hours struct {
			Metric struct {
				Value    float64 `json:"Value"`
				Unit     string  `json:"Unit"`
				UnitType int     `json:"UnitType"`
			} `json:"Metric"`
			Imperial struct {
				Value    float64 `json:"Value"`
				Unit     string  `json:"Unit"`
				UnitType int     `json:"UnitType"`
			} `json:"Imperial"`
		} `json:"Past6Hours"`
		Past9Hours struct {
			Metric struct {
				Value    float64 `json:"Value"`
				Unit     string  `json:"Unit"`
				UnitType int     `json:"UnitType"`
			} `json:"Metric"`
			Imperial struct {
				Value    float64 `json:"Value"`
				Unit     string  `json:"Unit"`
				UnitType int     `json:"UnitType"`
			} `json:"Imperial"`
		} `json:"Past9Hours"`
		Past12Hours struct {
			Metric struct {
				Value    float64 `json:"Value"`
				Unit     string  `json:"Unit"`
				UnitType int     `json:"UnitType"`
			} `json:"Metric"`
			Imperial struct {
				Value    float64 `json:"Value"`
				Unit     string  `json:"Unit"`
				UnitType int     `json:"UnitType"`
			} `json:"Imperial"`
		} `json:"Past12Hours"`
		Past18Hours struct {
			Metric struct {
				Value    float64 `json:"Value"`
				Unit     string  `json:"Unit"`
				UnitType int     `json:"UnitType"`
			} `json:"Metric"`
			Imperial struct {
				Value    float64 `json:"Value"`
				Unit     string  `json:"Unit"`
				UnitType int     `json:"UnitType"`
			} `json:"Imperial"`
		} `json:"Past18Hours"`
		Past24Hours struct {
			Metric struct {
				Value    float64 `json:"Value"`
				Unit     string  `json:"Unit"`
				UnitType int     `json:"UnitType"`
			} `json:"Metric"`
			Imperial struct {
				Value    float64 `json:"Value"`
				Unit     string  `json:"Unit"`
				UnitType int     `json:"UnitType"`
			} `json:"Imperial"`
		} `json:"Past24Hours"`
	} `json:"PrecipitationSummary"`
	TemperatureSummary struct {
		Past6HourRange struct {
			Minimum struct {
				Metric struct {
					Value    float64 `json:"Value"`
					Unit     string  `json:"Unit"`
					UnitType int     `json:"UnitType"`
				} `json:"Metric"`
				Imperial struct {
					Value    float64 `json:"Value"`
					Unit     string  `json:"Unit"`
					UnitType int     `json:"UnitType"`
				} `json:"Imperial"`
			} `json:"Minimum"`
			Maximum struct {
				Metric struct {
					Value    float64 `json:"Value"`
					Unit     string  `json:"Unit"`
					UnitType int     `json:"UnitType"`
				} `json:"Metric"`
				Imperial struct {
					Value    float64 `json:"Value"`
					Unit     string  `json:"Unit"`
					UnitType int     `json:"UnitType"`
				} `json:"Imperial"`
			} `json:"Maximum"`
		} `json:"Past6HourRange"`
		Past12HourRange struct {
			Minimum struct {
				Metric struct {
					Value    float64 `json:"Value"`
					Unit     string  `json:"Unit"`
					UnitType int     `json:"UnitType"`
				} `json:"Metric"`
				Imperial struct {
					Value    float64 `json:"Value"`
					Unit     string  `json:"Unit"`
					UnitType int     `json:"UnitType"`
				} `json:"Imperial"`
			} `json:"Minimum"`
			Maximum struct {
				Metric struct {
					Value    float64 `json:"Value"`
					Unit     string  `json:"Unit"`
					UnitType int     `json:"UnitType"`
				} `json:"Metric"`
				Imperial struct {
					Value    float64 `json:"Value"`
					Unit     string  `json:"Unit"`
					UnitType int     `json:"UnitType"`
				} `json:"Imperial"`
			} `json:"Maximum"`
		} `json:"Past12HourRange"`
		Past24HourRange struct {
			Minimum struct {
				Metric struct {
					Value    float64 `json:"Value"`
					Unit     string  `json:"Unit"`
					UnitType int     `json:"UnitType"`
				} `json:"Metric"`
				Imperial struct {
					Value    float64 `json:"Value"`
					Unit     string  `json:"Unit"`
					UnitType int     `json:"UnitType"`
				} `json:"Imperial"`
			} `json:"Minimum"`
			Maximum struct {
				Metric struct {
					Value    float64 `json:"Value"`
					Unit     string  `json:"Unit"`
					UnitType int     `json:"UnitType"`
				} `json:"Metric"`
				Imperial struct {
					Value    float64 `json:"Value"`
					Unit     string  `json:"Unit"`
					UnitType int     `json:"UnitType"`
				} `json:"Imperial"`
			} `json:"Maximum"`
		} `json:"Past24HourRange"`
	} `json:"TemperatureSummary"`
	MobileLink string `json:"MobileLink"`
	Link       string `json:"Link"`
}*/

type weatherapi struct {
	Location struct {
		Name           string  `json:"name"`
		Region         string  `json:"region"`
		Country        string  `json:"country"`
		Lat            float64 `json:"lat"`
		Lon            float64 `json:"lon"`
		TzID           string  `json:"tz_id"`
		LocaltimeEpoch int     `json:"localtime_epoch"`
		Localtime      string  `json:"localtime"`
	} `json:"location"`
	Current struct {
		LastUpdatedEpoch int     `json:"last_updated_epoch"`
		LastUpdated      string  `json:"last_updated"`
		TempC            float64 `json:"temp_c"`
		TempF            float64 `json:"temp_f"`
		IsDay            int     `json:"is_day"`
		Condition        struct {
			Text string `json:"text"`
			Icon string `json:"icon"`
			Code int    `json:"code"`
		} `json:"condition"`
		WindMph    float64 `json:"wind_mph"`
		WindKph    float64 `json:"wind_kph"`
		WindDegree int     `json:"wind_degree"`
		WindDir    string  `json:"wind_dir"`
		PressureMb float64 `json:"pressure_mb"`
		PressureIn float64 `json:"pressure_in"`
		PrecipMm   float64 `json:"precip_mm"`
		PrecipIn   float64 `json:"precip_in"`
		Humidity   int     `json:"humidity"`
		Cloud      int     `json:"cloud"`
		FeelslikeC float64 `json:"feelslike_c"`
		FeelslikeF float64 `json:"feelslike_f"`
		VisKm      float64 `json:"vis_km"`
		VisMiles   float64 `json:"vis_miles"`
		Uv         float64 `json:"uv"`
		GustMph    float64 `json:"gust_mph"`
		GustKph    float64 `json:"gust_kph"`
	} `json:"current"`
}

/* type weatherapihour struct {
	Location struct {
		Name           string  `json:"name"`
		Region         string  `json:"region"`
		Country        string  `json:"country"`
		Lat            float64 `json:"lat"`
		Lon            float64 `json:"lon"`
		TzID           string  `json:"tz_id"`
		LocaltimeEpoch int     `json:"localtime_epoch"`
		Localtime      string  `json:"localtime"`
	} `json:"location"`
	Current struct {
		LastUpdatedEpoch int     `json:"last_updated_epoch"`
		LastUpdated      string  `json:"last_updated"`
		TempC            int     `json:"temp_c"`
		TempF            float64 `json:"temp_f"`
		IsDay            int     `json:"is_day"`
		Condition        struct {
			Text string `json:"text"`
			Icon string `json:"icon"`
			Code int    `json:"code"`
		} `json:"condition"`
		WindMph    float64 `json:"wind_mph"`
		WindKph    float64 `json:"wind_kph"`
		WindDegree int     `json:"wind_degree"`
		WindDir    string  `json:"wind_dir"`
		PressureMb int     `json:"pressure_mb"`
		PressureIn float64 `json:"pressure_in"`
		PrecipMm   int     `json:"precip_mm"`
		PrecipIn   int     `json:"precip_in"`
		Humidity   int     `json:"humidity"`
		Cloud      int     `json:"cloud"`
		FeelslikeC int     `json:"feelslike_c"`
		FeelslikeF float64 `json:"feelslike_f"`
		VisKm      int     `json:"vis_km"`
		VisMiles   int     `json:"vis_miles"`
		Uv         int     `json:"uv"`
		GustMph    float64 `json:"gust_mph"`
		GustKph    float64 `json:"gust_kph"`
	} `json:"current"`
	Forecast struct {
		Forecastday []struct {
			Date      string `json:"date"`
			DateEpoch int    `json:"date_epoch"`
			Day       struct {
				MaxtempC          int     `json:"maxtemp_c"`
				MaxtempF          float64 `json:"maxtemp_f"`
				MintempC          float64 `json:"mintemp_c"`
				MintempF          int     `json:"mintemp_f"`
				AvgtempC          float64 `json:"avgtemp_c"`
				AvgtempF          float64 `json:"avgtemp_f"`
				MaxwindMph        float64 `json:"maxwind_mph"`
				MaxwindKph        float64 `json:"maxwind_kph"`
				TotalprecipMm     float64 `json:"totalprecip_mm"`
				TotalprecipIn     float64 `json:"totalprecip_in"`
				TotalsnowCm       int     `json:"totalsnow_cm"`
				AvgvisKm          float64 `json:"avgvis_km"`
				AvgvisMiles       int     `json:"avgvis_miles"`
				Avghumidity       int     `json:"avghumidity"`
				DailyWillItRain   int     `json:"daily_will_it_rain"`
				DailyChanceOfRain int     `json:"daily_chance_of_rain"`
				DailyWillItSnow   int     `json:"daily_will_it_snow"`
				DailyChanceOfSnow int     `json:"daily_chance_of_snow"`
				Condition         struct {
					Text string `json:"text"`
					Icon string `json:"icon"`
					Code int    `json:"code"`
				} `json:"condition"`
				Uv int `json:"uv"`
			} `json:"day"`
			Astro struct {
				Sunrise          string `json:"sunrise"`
				Sunset           string `json:"sunset"`
				Moonrise         string `json:"moonrise"`
				Moonset          string `json:"moonset"`
				MoonPhase        string `json:"moon_phase"`
				MoonIllumination string `json:"moon_illumination"`
				IsMoonUp         int    `json:"is_moon_up"`
				IsSunUp          int    `json:"is_sun_up"`
			} `json:"astro"`
			Hour []struct {
				TimeEpoch int     `json:"time_epoch"`
				Time      string  `json:"time"`
				TempC     float64 `json:"temp_c"`
				TempF     float64 `json:"temp_f"`
				IsDay     int     `json:"is_day"`
				Condition struct {
					Text string `json:"text"`
					Icon string `json:"icon"`
					Code int    `json:"code"`
				} `json:"condition"`
				WindMph      float64 `json:"wind_mph"`
				WindKph      float64 `json:"wind_kph"`
				WindDegree   int     `json:"wind_degree"`
				WindDir      string  `json:"wind_dir"`
				PressureMb   int     `json:"pressure_mb"`
				PressureIn   float64 `json:"pressure_in"`
				PrecipMm     int     `json:"precip_mm"`
				PrecipIn     int     `json:"precip_in"`
				Humidity     int     `json:"humidity"`
				Cloud        int     `json:"cloud"`
				FeelslikeC   float64 `json:"feelslike_c"`
				FeelslikeF   float64 `json:"feelslike_f"`
				WindchillC   float64 `json:"windchill_c"`
				WindchillF   float64 `json:"windchill_f"`
				HeatindexC   float64 `json:"heatindex_c"`
				HeatindexF   float64 `json:"heatindex_f"`
				DewpointC    float64 `json:"dewpoint_c"`
				DewpointF    float64 `json:"dewpoint_f"`
				WillItRain   int     `json:"will_it_rain"`
				ChanceOfRain int     `json:"chance_of_rain"`
				WillItSnow   int     `json:"will_it_snow"`
				ChanceOfSnow int     `json:"chance_of_snow"`
				VisKm        int     `json:"vis_km"`
				VisMiles     int     `json:"vis_miles"`
				GustMph      float64 `json:"gust_mph"`
				GustKph      int     `json:"gust_kph"`
				Uv           int     `json:"uv"`
			} `json:"hour"`
		} `json:"forecastday"`
	} `json:"forecast"`
} */

type Weather struct {
	ID            int64   `json:"id"`
	Date          string  `json:"date"`
	Temperature   int64   `json:"temperature"`
	FeelsLike     int64   `json:"feels_like"`
	Precipitation float64 `json:"precipitation"`
	Wind          float64 `json:"wind"`
	WindGust      float64 `json:"wind_gust"`
	WindDirection string  `json:"wind_direction"`
	Humidity      int64   `json:"humidity"`
	CloudCover    int64   `json:"cloud_cover"`
	WeatherText   string  `json:"weather_text"`
	WeatherIcon   string  `json:"weather_icon"`
	WeatherLink   string  `json:"weather_link"`
}

func (w *Weather) AddWeather() error {
	wa := weatherapi{}
	err := wa.getWeatherAPI()
	if err != nil {
		return err
	}

	w.Date = wa.Location.Localtime
	w.Temperature = int64(math.Round(wa.Current.TempF))
	w.FeelsLike = int64(math.Round(wa.Current.FeelslikeF))
	w.Wind = wa.Current.WindMph
	w.WindGust = wa.Current.GustMph
	w.WindDirection = wa.Current.WindDir
	w.Humidity = int64(wa.Current.Humidity)
	w.CloudCover = int64(wa.Current.Cloud)
	w.WeatherText = wa.Current.Condition.Text
	u, err := url.Parse("https:" + wa.Current.Condition.Icon)
	if err != nil {
		return err
	}
	w.WeatherIcon = "static/img" + u.Path
	w.WeatherLink = "https://weather.com/"

	query := fmt.Sprintf("INSERT INTO weather VALUES (NULL, \"%s\", %d, %d, %.2f, %.1f, %.1f, \"%s\", %d, %d, \"%s\", \"%s\", \"%s\")",
		w.Date,
		w.Temperature,
		w.FeelsLike,
		w.Precipitation,
		w.Wind,
		w.WindGust,
		w.WindDirection,
		w.Humidity,
		w.CloudCover,
		w.WeatherText,
		w.WeatherIcon,
		w.WeatherLink)

	ctx, cancelfunc := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelfunc()
	res, err := db.Con.ExecContext(ctx, query)
	if err != nil {
		return err
	}

	w.ID, err = res.LastInsertId()
	if err != nil {
		return err
	}

	return nil
}

func (w *Weather) GetWeatherByID() error {
	query := fmt.Sprintf(
		"SELECT "+
			"idweather, "+
			"weather_date, "+
			"temperature, "+
			"feels_like, "+
			"precipitation, "+
			"wind, "+
			"wind_gust, "+
			"wind_direction, "+
			"humidity, "+
			"cloudcover, "+
			"weather_text, "+
			"weather_icon "+
			"FROM weather WHERE idweather=%d",
		w.ID)

	ctx, cancelfunc := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelfunc()
	err := db.Con.QueryRowContext(ctx, query).Scan(
		&w.ID,
		&w.Date,
		&w.Temperature,
		&w.FeelsLike,
		&w.Precipitation,
		&w.Wind,
		&w.WindGust,
		&w.WindDirection,
		&w.Humidity,
		&w.CloudCover,
		&w.WeatherText,
		&w.WeatherIcon)
	if err != nil {
		return err
	}

	return nil
}

func (w *Weather) GetWeatherByDate(d string) error {
	t, err := time.Parse("2006-01-02", d)
	if err != nil {
		return err
	}
	s := fmt.Sprintf("\"%d-%d-%d 00:00:00\"", t.Year(), t.Month(), t.Day())
	f := fmt.Sprintf("\"%d-%d-%d 23:59:59\"", t.Year(), t.Month(), t.Day())
	query := fmt.Sprintf(
		"SELECT "+
			"idweather, "+
			"date, "+
			"temperature, "+
			"feels_like, "+
			"precipitation, "+
			"wind, "+
			"wind_gust, "+
			"wind_direction, "+
			"humidity, "+
			"cloudcover "+
			"weather_text "+
			"weather_icon "+
			"FROM weather WHERE "+
			"date>=%s AND date<=%s",
		s, f)

	ctx, cancelfunc := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelfunc()

	err = db.Con.QueryRowContext(ctx, query).Scan(
		&w.ID,
		&w.Date,
		&w.Temperature,
		&w.FeelsLike,
		&w.Precipitation,
		&w.Wind,
		&w.WindGust,
		&w.WindDirection,
		&w.Humidity,
		&w.CloudCover)
	if err != nil {
		return err
	}

	return nil
}

/* func (aw *accuWeather) getAccuWeather() error {
	resp, err := http.Get("http://dataservice.accuweather.com/currentconditions/v1/332128?apikey=put8mfXawbPRMEXpDunTjZrKWJCw4AeE&details=true")
	if err != nil {
		return err
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	fmt.Printf("%s\n", body)

	err = json.Unmarshal([]byte(body), aw)
	if err != nil {
		return err
	}

	return nil
} */

func (wa *weatherapi) getWeatherAPI() error {
	resp, err := http.Get("http://api.weatherapi.com/v1/current.json?key=a59ece49937045878a8175453230605&q=37.57,-122.28")
	if err != nil {
		return err
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	err = json.Unmarshal([]byte(body), wa)
	if err != nil {
		return err
	}

	return nil
}
