package service

import (
	"bufio"
	"context"
	"io"
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
	f       io.Reader
	manager model.GeoLocationManager
}

func NewParser(f io.Reader, mn model.GeoLocationManager) ParserService {
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

	outPutChan := make(chan *model.Geolocation, 10000)
	go p.saveToDB(outPutChan)

	var validDataCount int64 = 0
	var inValidDataCount int64 = 0
	var wg sync.WaitGroup
	visitedIP := make(map[string]bool) // will keep track of already visited ip address

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

		c1, c2 := processChunk(buf, positions, outPutChan, visitedIP)
		inValidDataCount += c1
		validDataCount += c2
	}

	close(outPutChan)

	wg.Wait() //wait ExtractAndLoad

	return time.Since(timeThen).Seconds(), inValidDataCount, validDataCount, nil
}

func processChunk(chunk []byte, positions map[int]string, outPutChan chan<- *model.Geolocation, visitedIP map[string]bool) (int64, int64) {
	var invalidCount int64 = 0
	var validCount int64 = 0
	logs := string(chunk)

	logsSlice := strings.Split(logs, "\n")
	n := len(logsSlice)

	for i := 0; i < n; i++ {
		geoloc, valid := isValidLine(positions, logsSlice[i], visitedIP)
		if !valid {
			invalidCount++
		} else {
			outPutChan <- geoloc
			visitedIP[geoloc.IP] = true
			validCount++
		}
	}

	return invalidCount, validCount //return the invalid data count
}

func isValidLine(positions map[int]string, text string, visitedIP map[string]bool) (*model.Geolocation, bool) {
	if len(text) == 0 { //in case there is line gap
		return nil, false
	}
	logSlice := strings.Split(text, ",")
	if len(logSlice) != 7 { //if not valid number of fields
		return nil, false
	}

	geoloc := model.Geolocation{}
	for i, value := range logSlice {
		if len(value) == 0 { //if empty value
			return nil, false
		}
		col := positions[i]
		switch col {
		case common.IP:
			IPValid := common.IsIpv4Regex(value)
			if !IPValid {
				return nil, false
			}
			if ok := visitedIP[value]; ok {
				return nil, false
			}
			geoloc.IP = value
		case common.CountryCode:
			geoloc.CountryCode = value
		case common.Country:
			geoloc.Country = value
		case common.Longitude:
			longitude, err := strconv.ParseFloat(value, 64)
			if err != nil {
				return nil, false
			}
			if longitude > 180 || longitude < -180 { //invalid longitude coordinates
				return nil, false
			}
			geoloc.Longitude = value
		case common.Latitude:
			latitude, err := strconv.ParseFloat(value, 64)
			if err != nil {
				return nil, false
			}
			if latitude > 90 || latitude < -90 { //invalid latitude coordinates
				return nil, false
			}

			geoloc.Latitude = value
		case common.MysteryValue:
			geoloc.MysteryValue = value
		case common.City:
			geoloc.City = value
		default: //if some other columns come in
			return nil, false
		}
	}
	return &geoloc, true
}

func (p *parser) saveToDB(savChan <-chan *model.Geolocation) {
	resultSlice := make([]*model.Geolocation, 0, bulkSize) //8191, coz postgres supports 65535 parameters in bulk
	for data := range savChan {
		resultSlice = append(resultSlice, data)
		if len(resultSlice) == bulkSize {
			var local_data []*model.Geolocation
			local_data = append(local_data, resultSlice...)
			go func() {
				err := p.manager.BulkInsert(context.Background(), local_data)
				if err != nil {
					zlog.Logger().Warn("Error occurred while bulk insert", zlog.ParamsType{"Error": err.Error()})
					return
				}
			}()
			resultSlice = nil
		}
	}
}
