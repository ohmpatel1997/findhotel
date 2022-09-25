package service

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"math"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/ohmpatel1997/findhotel/internal/common"
	"github.com/ohmpatel1997/findhotel/internal/model"
	zlog "github.com/ohmpatel1997/findhotel/lib/log"
)

const (
	bulkSize = 8191
)

type ParserService interface {
	ParseAndStore() (float64, int64, int64, error)
}

type parser struct {
	f       *os.File
	manager model.GeoLocationManager
}

func NewParser(f *os.File, mn model.GeoLocationManager) ParserService {
	return &parser{
		f:       f,
		manager: mn,
	}
}

func (p *parser) ParseAndStore() (float64, int64, int64, error) {

	timeThen := time.Now()
	r := bufio.NewReader(p.f)

	firstLine, _, err := r.ReadLine()
	if err != nil {
		return 0, 0, 0, err
	}
	firstLineSlice := strings.Split(string(firstLine), ",")
	positions := make(map[int]string)

	//map the positions
	for i, header := range firstLineSlice {
		positions[i] = header
	}

	var invalidDataCountFromFirstPass int64 = 0
	var invalidDataCountFromSecondPass int64 = 0
	var validDataCount int64 = 0

	outPutChan := make(chan model.Geolocation, 10000)
	savToDbChan := make(chan model.Geolocation, 10000)
	var wg sync.WaitGroup
	var wg2 sync.WaitGroup

	go func() {
		validDataCount, invalidDataCountFromSecondPass = p.extractAndLoad(outPutChan, &wg, savToDbChan, &wg2)
	}()

	go p.SaveToDB(savToDbChan, &wg2)

	//PrintMemUsage()

	for {
		buf := make([]byte, 500*1024)
		n, err := r.Read(buf)
		buf = buf[:n]

		if n == 0 {
			break
		}

		nextUntillNewline, err := r.ReadBytes('\n')

		if err != io.EOF {
			buf = append(buf, nextUntillNewline...)
		}

		invalidDataCountFromFirstPass += processChunk(buf, positions, outPutChan, &wg)
	}

	//PrintMemUsage()
	close(outPutChan)

	wg.Wait()  //wait ExtractAndLoad
	wg2.Wait() //wait until saving data to db

	return time.Since(timeThen).Seconds(), invalidDataCountFromFirstPass + invalidDataCountFromSecondPass, validDataCount, nil
}

func processChunk(chunk []byte, positions map[int]string, outPutChan chan<- model.Geolocation, wg *sync.WaitGroup) int64 {

	var wg2 sync.WaitGroup
	var invalid int64 = 0
	logs := string(chunk)

	logsSlice := strings.Split(logs, "\n")

	chunkSize := 500
	n := len(logsSlice)
	noOfThread := n / chunkSize

	if n%chunkSize != 0 {
		noOfThread++
	}

	for i := 0; i < (noOfThread); i++ {

		wg2.Add(1) //span out locally

		go func(c, j int) {
			for ; c < j; c++ { //first stage of cleaning
				text := logsSlice[c]
				if len(text) == 0 { //in case there is line gap
					continue
				}
				logSlice := strings.Split(text, ",")

				if len(logSlice) != 7 { //if not valid number of fields
					invalid++
					continue
				}

				geoloc := model.Geolocation{}
				invalidData := false
				for i, value := range logSlice {

					if len(value) == 0 { //if empty value
						invalid++
						invalidData = true
						break
					}
					col := positions[i]
					switch col {
					case common.IP:
						geoloc.IP = value
					case common.CountryCode:
						geoloc.CountryCode = value
					case common.Country:
						geoloc.Country = value
					case common.Longitude:
						geoloc.Longitude = value
					case common.Latitude:
						geoloc.Latitude = value
					case common.MysteryValue:
						geoloc.MysteryValue = value
					case common.City:
						geoloc.City = value
					default: //if some other columns come in
						invalidData = true
						invalid++
						break
					}
				}

				if !invalidData {
					wg.Add(1)            //increment counter for data processing
					outPutChan <- geoloc //send to output chan
				}

			}
			wg2.Done() //done processing a chunk

		}(i*chunkSize, int(math.Min(float64((i+1)*chunkSize), float64(len(logsSlice))))) //prevent overflow
	}

	wg2.Wait()
	return invalid //return the invalid data count
}

// extractAndLoad  will extract the data, checks the validity and load it into database
func (p *parser) extractAndLoad(outPutChan <-chan model.Geolocation, wg *sync.WaitGroup, saveToDbChan chan<- model.Geolocation, wg2 *sync.WaitGroup) (int64, int64) {

	visitedIP := make(map[string]bool)          // will keep track of already visited ip address
	visitedCoordinates := make(map[string]bool) // will keep track of already visited coordinates
	var invalidCount int64
	var validCount int64

	for data := range outPutChan { // second stage of cleaning
		local_data := data
		IPValid := common.IsIpv4Regex(local_data.IP)
		if !IPValid {
			invalidCount++
			wg.Done()
			continue
		}

		latitude, err := strconv.ParseFloat(local_data.Latitude, 64)
		if err != nil {
			invalidCount++
			wg.Done()
			continue
		}

		longitude, err := strconv.ParseFloat(local_data.Longitude, 64)
		if err != nil {
			invalidCount++
			wg.Done()
			continue
		}

		if latitude > 90 || latitude < -90 { //invalid latitude coordinates
			invalidCount++
			wg.Done()
			continue
		}

		if longitude > 180 || longitude < -180 { //invalid longitude coordinates
			invalidCount++
			wg.Done()
			continue
		}

		if ok := visitedIP[local_data.IP]; ok {
			invalidCount++
			wg.Done()
			continue
		}

		coordinates := fmt.Sprintf("%s+%s", data.Latitude, data.Longitude)
		if ok := visitedCoordinates[coordinates]; ok {
			invalidCount++
			wg.Done()
			continue
		}

		visitedIP[data.IP] = true
		visitedCoordinates[coordinates] = true
		validCount++

		if validCount%bulkSize == 0 {
			wg2.Add(1) //add count for saving data to db
		}
		saveToDbChan <- local_data //push to save data to db

		wg.Done() //decrement for process done
	}

	close(saveToDbChan) //once all data have been processed, close save to db chan
	return validCount, invalidCount
}

func (p *parser) SaveToDB(savChan <-chan model.Geolocation, wg2 *sync.WaitGroup) {
	resultSlice := make([]*model.Geolocation, 0, bulkSize) //8191, coz postgres supports 65535 parameters in bulk
	for data := range savChan {
		resultSlice = append(resultSlice, &data)
		if len(resultSlice) == bulkSize {
			var local_data []*model.Geolocation
			local_data = append(local_data, resultSlice...)
			go func() {
				defer wg2.Done()
				err := p.manager.BulkInsert(context.Background(), local_data)
				if err != nil {
					zlog.Logger().Warn("Error occurred while bulk insert", zlog.ParamsType{"Error": err.Error()})
					return
				}
			}()

			resultSlice = resultSlice[:0]
		}

	}
}
