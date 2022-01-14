package weather

// weather is a daemon that will download weather data from accuweather once daily at:
//	12 PM on Sunday and Saturday
//	1 PM on Monday - Friday

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"mariners/db"
	"math"
	"net/http"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

// accuWeather is a []struct that can store the unmarshalled results of the accuweather current conditions API
type accuWeather []struct {
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
}

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
	WeatherIcon   int64   `json:"weather_icon"`
	WeatherLink   string  `json:"weather_link"`
}

func (w *Weather) AddWeather() error {
	db, err := db.DBConnection()
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	aw := accuWeather{}
	err = aw.getAccuWeather()
	if err != nil {
		return err
	}

	w.Date = aw[0].LocalObservationDateTime
	w.Temperature = int64(math.Round(aw[0].Temperature.Imperial.Value))
	w.FeelsLike = int64(math.Round(aw[0].RealFeelTemperature.Imperial.Value))
	w.Precipitation = aw[0].Precip1Hr.Imperial.Value
	w.Wind = aw[0].Wind.Speed.Imperial.Value
	w.WindGust = aw[0].WindGust.Speed.Imperial.Value
	w.WindDirection = aw[0].Wind.Direction.English
	w.Humidity = int64(aw[0].RelativeHumidity)
	w.CloudCover = int64(aw[0].CloudCover)
	w.WeatherText = aw[0].WeatherText
	w.WeatherIcon = int64(aw[0].WeatherIcon)
	w.WeatherLink = aw[0].Link

	query := fmt.Sprintf("INSERT INTO weather VALUES (NULL, \"%s\", %d, %d, %.2f, %.1f, %.1f, \"%s\", %d, %d, \"%s\", %d, \"%s\")",
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

	fmt.Printf("\n\nQUERY: \n%s\n\n", query)

	ctx, cancelfunc := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelfunc()
	res, err := db.ExecContext(ctx, query)
	if err != nil {
		fmt.Printf("%s\n", err.Error())
		return err
	}

	w.ID, err = res.LastInsertId()
	if err != nil {
		fmt.Printf("%s\n", err.Error())
		return err
	}

	fmt.Printf("AddWeather: %#v\n", w)
	return nil
}

func (w *Weather) GetWeatherByID() error {
	db, err := db.DBConnection()
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

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
			"cloudcover, "+
			"weather_text, "+
			"weather_icon, "+
			"weather_link "+
			"FROM weather WHERE idweather=%d",
		w.ID)

	fmt.Printf("\n\nQUERY: \n%s\n\n", query)

	ctx, cancelfunc := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelfunc()
	err = db.QueryRowContext(ctx, query).Scan(
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
		&w.WeatherIcon,
		&w.WeatherLink)
	if err != nil {
		return err
	}

	fmt.Printf("GetWeatherByID: %#v\n", w)
	return nil
}

func (w *Weather) GetWeatherByDate(d string) error {
	db, err := db.DBConnection()
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

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
			"weather_link "+
			"FROM weather WHERE "+
			"date>=%s AND date<=%s",
		s, f)

	fmt.Printf("\n\nQUERY: \n%s\n\n", query)

	ctx, cancelfunc := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelfunc()

	err = db.QueryRowContext(ctx, query).Scan(
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

	fmt.Printf("GetWeatherByDate: %#v", w)

	return nil
}

func (aw *accuWeather) getAccuWeather() error {
	resp, err := http.Get("http://dataservice.accuweather.com/currentconditions/v1/332128?apikey=put8mfXawbPRMEXpDunTjZrKWJCw4AeE&details=true")
	if err != nil {
		return err
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	err = json.Unmarshal([]byte(body), aw)
	if err != nil {
		return err
	}

	return nil
}
