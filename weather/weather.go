package main

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
	"os"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

// Database connection parameters (secrets)

// Weather is a []struct that can store the unmarshalled results of the accuweather current conditions API
type Weather []struct {
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

func getWeather(weather *Weather) error {
	resp, err := http.Get("http://dataservice.accuweather.com/currentconditions/v1/332128?apikey=put8mfXawbPRMEXpDunTjZrKWJCw4AeE&details=true")
	if err != nil {
		return err
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	err = json.Unmarshal([]byte(body), weather)
	if err != nil {
		return err
	}

	return nil
}

func dsn(dbName string) string {
	username := os.Getenv("dbuser")
	password := os.Getenv("dbpassword")
	hostname := os.Getenv("dbhost")

	return fmt.Sprintf("%s:%s@tcp(%s)/%s", username, password, hostname, dbName)
}

func dbConnection() (*sql.DB, error) {
	dbname := os.Getenv("dbname")

	db, err := sql.Open("mysql", dsn(dbname))
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(20)
	db.SetMaxIdleConns(20)
	db.SetConnMaxLifetime(time.Minute * 5)

	ctx, cancelfunc := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelfunc()
	err = db.PingContext(ctx)
	if err != nil {
		return nil, err
	}

	return db, nil
}

func writeWeather(db *sql.DB, weather Weather) error {
	query := fmt.Sprintf("INSERT INTO weather VALUES (NULL, \"%s\", %d, %d, %.2f, %.1f, %.1f, \"%s\", %d, %d)",
		weather[0].LocalObservationDateTime,
		int64(math.Round(weather[0].Temperature.Imperial.Value)),
		int64(math.Round(weather[0].RealFeelTemperature.Imperial.Value)),
		weather[0].Precip1Hr.Imperial.Value,
		weather[0].Wind.Speed.Imperial.Value,
		weather[0].WindGust.Speed.Imperial.Value,
		weather[0].Wind.Direction.English,
		weather[0].RelativeHumidity,
		weather[0].CloudCover)

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

func main() {
	weather := Weather{}
	err := getWeather(&weather)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Retrieved weather fo date %s\n", weather[0].LocalObservationDateTime)

	db, err := db.DBConnection()
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	fmt.Printf("Successfully connected to database\n")

	err = writeWeather(db, weather)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Weather inserted to database\n")
}
