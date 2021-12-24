package weather

// weather is a daemon that will download weather data from accuweather once daily at:
//	12 PM on Sunday and Saturday
//	1 PM on Monday - Friday

import (
	"context"
	"database/sql"
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

// Database connection parameters (secrets)

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
	ID            int     `json:"id"`
	Date          string  `json:"date"`
	Temperature   int     `json:"temperature"`
	FeelsLike     int     `json:"feels_like"`
	Precipitation float64 `json:"precipitation"`
	Wind          float64 `json:"wind"`
	WindGust      float64 `json:"wind_gust"`
	WindDirection string  `json:"wind_direction"`
	Humidity      int     `json:"humidity"`
	CloudCover    int     `json:"cloud_cover"`
}

func GetWeather(id int64, w *Weather) error {
	db, err := db.DBConnection()
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	err = getWeather(db, id, w)
	if err != nil {
		return err
	}

	return nil
}

func GetWeatherByDate(d string, w *Weather) error {
	db, err := db.DBConnection()
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	err = getWeatherByDate(db, d, w)
	if err != nil {
		return err
	}

	return nil
}

func getWeather(db *sql.DB, id int64, w *Weather) error {
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
			"FROM weather WHERE idweather=%d",
		id)

	fmt.Printf("\n\nQUERY: \n%s\n\n", query)

	ctx, cancelfunc := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelfunc()
	err := db.QueryRowContext(ctx, query).Scan(
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

	fmt.Printf("%#v", w)

	return nil
}

func getWeatherByDate(db *sql.DB, d string, w *Weather) error {
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

	fmt.Printf("%#v", w)

	return nil
}

func AddWeather() error {
	db, err := db.DBConnection()
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	err = writeWeather(db)
	if err != nil {
		return err
	}

	return nil
}

func getAccuWeather(aw *accuWeather) error {
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

func writeWeather(db *sql.DB) error {
	aw := accuWeather{}
	err := getAccuWeather(&aw)
	if err != nil {
		return err
	}

	query := fmt.Sprintf("INSERT INTO weather VALUES (NULL, \"%s\", %d, %d, %.2f, %.1f, %.1f, \"%s\", %d, %d)",
		aw[0].LocalObservationDateTime,
		int64(math.Round(aw[0].Temperature.Imperial.Value)),
		int64(math.Round(aw[0].RealFeelTemperature.Imperial.Value)),
		aw[0].Precip1Hr.Imperial.Value,
		aw[0].Wind.Speed.Imperial.Value,
		aw[0].WindGust.Speed.Imperial.Value,
		aw[0].Wind.Direction.English,
		aw[0].RelativeHumidity,
		aw[0].CloudCover)

	ctx, cancelfunc := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelfunc()
	res, err := db.ExecContext(ctx, query)
	if err != nil {
		return err
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	fmt.Printf("Rows affected by insert: %d\n", rows)

	return nil
}
