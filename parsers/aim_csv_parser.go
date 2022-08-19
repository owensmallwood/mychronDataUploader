package parsers

import (
	"fmt"
	"go.mongodb.org/mongo-driver/mongo"
	"karting-grafana/database"
	"os"
	"strconv"
	"time"
)

// LapFrame assumes the csv is exported from Race Studio in the AiM CSV format
type LapFrame struct {
	SessionTimestamp string
	Venue            string
	Driver           string
	Lap              int
	Time             float32
	Distance         float32
	Speed            float32
	LatAccel         float32
	LongAccel        float32
	PosAccuracy      float32
	Lat              float32
	Long             float32
}

type SessionMeta struct {
	SessionTimestamp string
	Venue            string
	Headers          []string
	LapMarkers       []string
	Driver           string
	Date             string
	Time             string
}

const (
	TIME_ROW_INDEX     = 0
	VENUE_INDEX        = 1
	RACER_NAME_INDEX   = 3
	SESSION_DATE_INDEX = 6
	SESSION_TIME_INDEX = 7
	LAP_MARKERS_INDEX  = 11
	COLUMNS_INDEX      = 14
	LAP_START_INDEX    = 17
)

func ParseMeta(data [][]string) SessionMeta {
	lapMarkers := deleteEmpty(data[LAP_MARKERS_INDEX][1:])
	headers := data[COLUMNS_INDEX]
	racerName := data[RACER_NAME_INDEX][1]
	sessionDate := data[SESSION_DATE_INDEX][1]
	sessionTime := data[SESSION_TIME_INDEX][1]
	venue := data[VENUE_INDEX][1]
	sessionTimestamp, err := time.Parse("Monday, January 2, 2006, 3:04 PM", sessionDate+", "+sessionTime)
	if err != nil {
		fmt.Println(err)
	}

	meta := SessionMeta{
		SessionTimestamp: sessionTimestamp.String(),
		Venue:            venue,
		Headers:          headers,
		LapMarkers:       lapMarkers,
		Driver:           racerName,
		Date:             sessionDate,
		Time:             sessionTime,
	}

	return meta
}

func CreateLapTimes(data [][]string, meta SessionMeta, db *mongo.Client) {
	type Lap struct {
		SessionTimestamp string
		Venue            string
		Driver           string
		LapNumber        int
		Time             float32
	}

	// Get the raw lap times from the csv data
	rawLapTimes := deleteEmpty(data[LAP_MARKERS_INDEX][1:])
	// Remove the last lap because its never a full lap time
	rawLapTimes = rawLapTimes[:len(rawLapTimes)-1]

	var prevLapTime float32
	prevLapTime = 0.000
	var laps []interface{} // needs to blank interface or mongodb shits itself
	for i, rawLapTime := range rawLapTimes {
		lap := Lap{
			LapNumber:        i + 1,
			Time:             float32(strToFloat(rawLapTime)) - prevLapTime,
			SessionTimestamp: meta.SessionTimestamp,
			Venue:            meta.Venue,
			Driver:           meta.Driver,
		}
		laps = append(laps, lap)
		prevLapTime = float32(strToFloat(rawLapTime))
	}

	if os.Getenv("LOAD_SESSION") == "true" {
		database.InsertMany(db, "lap_times", laps)
	}
}

func CreateLapFrames(data [][]string, meta SessionMeta, db *mongo.Client) {
	nextMarkerIndex := 0
	nextMarker := strToFloat(meta.LapMarkers[nextMarkerIndex])
	currentLap := 1
	var rows []map[string]string
	startTime, _ := time.Parse("2006-01-02 15:04:05.000", "2022-06-19 00:00:00.000")

	var lapTimeBase float32
	lapTimeBase = 0.000

	for _, row := range data {
		// skip blank rows
		row := deleteEmpty(row)
		if len(row) == 0 {
			continue
		}

		rowTime := strToFloat(row[TIME_ROW_INDEX])
		// Are we at a new lap?
		if rowTime >= nextMarker {
			nextMarkerIndex++
			// Is there a next marker?
			if nextMarkerIndex < len(meta.LapMarkers) {
				nextMarker = strToFloat(meta.LapMarkers[nextMarkerIndex])
				currentLap++
				lapTimeBase = float32(strToFloat(meta.LapMarkers[nextMarkerIndex-1])) // set base lap time as the last encountered marker
			}
		}

		// Process the row
		// In addition to all the column headers present, we need to add on lap number, session timestamp, and lap timestamp

		// Add values from each header to the row
		newRow := map[string]string{}
		for i, header := range meta.Headers {
			newRow[header] = row[i]
		}

		// Add lap number to the row
		newRow["Lap"] = strconv.FormatInt(int64(currentLap), 10)

		// Add session timestamp to row
		dur, _ := time.ParseDuration(row[TIME_ROW_INDEX] + "s")
		sessionTime := startTime.Add(dur)
		newRow["Time"] = sessionTime.Format("2006-01-02 15:04:05.000")

		// Add lap timestamp to row
		relativeLapTime := float32(strToFloat(row[TIME_ROW_INDEX])) - lapTimeBase
		lapDur, _ := time.ParseDuration(fmt.Sprintf("%f", relativeLapTime) + "s")
		relativeLapTimestamp := startTime.Add(lapDur)
		newRow["LapTime"] = relativeLapTimestamp.Format("2006-01-02 15:04:05.000")

		rows = append(rows, newRow)
	}

	fmt.Println(len(rows), "csv rows processed")

	if len(rows) > 0 {
		insertLapFramesMongo(rows, meta, db)
	}
}

func insertLapFramesMongo(rows []map[string]string, meta SessionMeta, db *mongo.Client) {
	type Row struct {
		SessionTimestamp string
		Venue            string
		Driver           string
		Date             string
		Lap              int64
		RelativeLap      string
		Speed            float32
		Lat              float64
		Long             float64
		Rpm              float32
	}

	var lapFrames []interface{}
	for _, row := range rows {
		lap, _ := strconv.ParseInt(row["Lap"], 10, 12)

		// aim_format_example.csvThe coords were offset a little bit and didnt line up with the regina track.
		// Had to manually adjust them. Hopefully the offset doesnt change in future mychron data.
		newRow := Row{
			SessionTimestamp: meta.SessionTimestamp,
			Venue:            meta.Venue,
			Driver:           meta.Driver,
			Date:             row["Time"],
			Lap:              lap,
			RelativeLap:      row["LapTime"],
			Speed:            float32(strToFloat(row["GPS Speed"])),
			Lat:              strToFloat(row["GPS Latitude"]) + 0.00003000,
			Long:             strToFloat(row["GPS Longitude"]) - 0.00002000,
			Rpm:              float32(strToFloat(row["RPM"])),
		}

		lapFrames = append(lapFrames, newRow)
	}

	fmt.Println("Sample lap frame to insert:")
	fmt.Printf("%+v\n", lapFrames[0])

	if os.Getenv("LOAD_SESSION") == "true" {
		database.InsertMany(db, "lap_frames", lapFrames)
	}
}
