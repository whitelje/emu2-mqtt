/*
 *  XML Structures and adapters for EMU2
 *
 *  Copyright Jake Whiteley
 *  SPDX-License-Identifier: Apache-2.0
 *
 *  Licensed under the Apache License, Version 2.0 (the "License"); you may
 *  not use this file except in compliance with the License.
 *  You may obtain a copy of the License at
 *
 *  http://www.apache.org/licenses/LICENSE-2.0
 *
 *  Unless required by applicable law or agreed to in writing, software
 *  distributed under the License is distributed on an "AS IS" BASIS, WITHOUT
 *  WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 *  See the License for the specific language governing permissions and
 *  limitations under the License.
 */
package main

import (
	"encoding/xml"
	"github.com/shopspring/decimal"
	"math"
	"strconv"
	"time"
)

type XMLProfileData struct {
	XMLName                 xml.Name `xml:"ProfileData"`
	DeviceMacId             []byte
	MeterMacId              []byte
	EndTime                 string
	Status                  string
	ProfileIntervalPeriod   string
	NumberOfPointsDelivered string
	IntervalData            []string
}

// CommandNoArgs
//Commands without arguments:
//initialize
//restart
//factory_reset
//get_connection_status
//get_device_info
//get_meter_list
//get_network_info
type CommandNoArgs struct {
	Name string
}

// CommandGetSchedule
//get_sechedule
type CommandGetSchedule struct {
	Name       string
	MeterMacId []byte
	Event      string
}

// CommandSetSchedule
// set_schedule
type CommandSetSchedule struct {
	Name       string
	MeterMacId []byte
	Event      string
	Frequency  uint32
	Enabled    string
}
type XMLCommandSetSchedule struct {
	Name       string
	MeterMacId []byte
	Event      string
	Frequency  string
	Enabled    string
}

// CommandMeterInfo
// get_meter_info
// get_current_period_usage
// get_last_period_usage
// close_current_period
type CommandMeterInfo struct {
	Name       string
	MeterMacId []byte
}

type CommandSetMeterInfo struct {
	Name       string
	MeterMacId []byte
	NickName   string
	Account    string
	Auth       string
	Host       string
	Enabled    string
}

// CommandGet
// get_time
//get_message
//get_current_price
//get_instantaneous_demand
//get_current_summation_delivered
type CommandGet struct {
	Name       string
	MeterMacId []byte
	Refresh    string
}

// CommandPoll
//set_fast_poll
type CommandPoll struct {
	Name       string
	MeterMacId []byte
	Frequency  uint16
	Duration   uint16
}
type XMLCommandPoll struct {
	Name       string
	MeterMacId []byte
	Frequency  string
	Duration   string
}

func convertCommandPoll(xml CommandPoll) (XMLCommandPoll, error) {
	var ret XMLCommandPoll
	ret.Name = xml.Name
	ret.MeterMacId = xml.MeterMacId
	ret.Frequency = "0x" + strconv.FormatUint(uint64(xml.Frequency), 16)
	ret.Duration = "0x" + strconv.FormatUint(uint64(xml.Duration), 16)

	return ret, nil
}

// CommandGetProfileData
// get_profile_data
type CommandGetProfileData struct {
	Name            string
	MeterMacId      []byte
	NumberOfPeriods uint8
	EndTime         uint32
	IntervalChannel string
}
type XMLCommandGetProfileData struct {
	Name            string
	MeterMacId      []byte
	NumberOfPeriods string
	EndTime         string
	IntervalChannel string
}

func convertCommandGetProfileData(xml CommandGetProfileData) (XMLCommandGetProfileData, error) {
	var ret XMLCommandGetProfileData

	ret.Name = xml.Name
	ret.MeterMacId = xml.MeterMacId
	ret.IntervalChannel = xml.IntervalChannel
	ret.NumberOfPeriods = "0x" + strconv.FormatUint(uint64(xml.NumberOfPeriods), 16)
	ret.EndTime = "0x" + strconv.FormatUint(uint64(xml.EndTime), 16)

	return ret, nil
}

type XMLConnectionStatus struct {
	XMLName      xml.Name `xml:"ConnectionStatus"`
	DeviceMacId  []byte
	MeterMacId   []byte
	Status       string
	Description  string
	StatusCode   string
	ExtPanId     string
	Channel      string
	ShortAddr    string
	LinkStrength string
}

// ConnectionStatus
// connection_status
type ConnectionStatus struct {
	XMLName      xml.Name `xml:"ConnectionStatus"`
	DeviceMacId  []byte
	MeterMacId   []byte
	Status       string
	Description  string
	StatusCode   uint8
	ExtPanId     uint32
	Channel      int
	ShortAddr    uint16
	LinkStrength uint8
}

func convertXMLConnectionStatus(xml XMLConnectionStatus) (ConnectionStatus, error) {
	var ret ConnectionStatus
	ret.DeviceMacId = xml.DeviceMacId
	ret.MeterMacId = xml.MeterMacId
	ret.Status = xml.Status
	ret.Description = xml.Description
	if xml.StatusCode == "" {
		ret.StatusCode = 0
	} else {
		statusCode, err := strconv.ParseUint(xml.StatusCode[2:], 16, 8)
		if err != nil {
			return ret, err
		}
		ret.StatusCode = uint8(statusCode)
	}
	if xml.ExtPanId == "" {
		ret.ExtPanId = 0
	} else {
		extPanid, err := strconv.ParseUint(xml.ExtPanId[2:], 16, 32)
		if err != nil {
			return ret, err
		}
		ret.ExtPanId = uint32(extPanid)
	}
	if xml.Channel == "" {
		ret.Channel = 0
	} else {
		channel, err := strconv.ParseInt(xml.Channel, 10, 8)
		if err != nil {
			return ret, err
		}
		ret.Channel = int(channel)
	}
	if xml.ShortAddr == "" {
		ret.ShortAddr = 0
	} else {
		shortAddr, err := strconv.ParseUint(xml.ShortAddr[2:], 16, 16)
		if err != nil {
			return ret, err
		}
		ret.ShortAddr = uint16(shortAddr)
	}
	linkStrength, err := strconv.ParseUint(xml.LinkStrength[2:], 16, 8)
	if err != nil {
		return ret, err
	}
	ret.LinkStrength = uint8(linkStrength)
	return ret, nil
}

type DeviceInfo struct {
	XMLName      xml.Name `xml:"DeviceInfo"`
	DeviceMacId  []byte
	InstallCode  []byte
	LinkKey      []byte
	FWVersion    string
	HWVersion    string
	ImageType    string
	Manufacturer string
	ModelId      string
	DateCode     string
}

type ScheduleInfo struct {
	XMLName     xml.Name `xml:"ScheduleInfo"`
	DeviceMacId []byte
	MeterMacId  []byte
	Event       string
	Frequency   string
	Enabled     string
}

type MeterList struct {
	XMLName     xml.Name `xml:"MeterList"`
	DeviceMacId []byte
	MeterMacId  [][]byte
}

type MeterInfo struct {
	XMLName     xml.Name `xml:"MeterInfo"`
	DeviceMacId []byte
	MeterMacId  []byte
	MeterType   string
	NickName    string
	Account     string
	Auth        string
	Host        string
	Enabled     string
}

type XMLNetworkInfo struct {
	XMLName      xml.Name `xml:"NetworkInfo"`
	DeviceMacId  []byte
	CoordMacId   []byte
	Status       string
	Description  string
	StatusCode   string
	ExtPanId     []byte
	Channel      string
	ShortAddr    string
	LinkStrength string
}
type NetworkInfo struct {
	DeviceMacId  []byte
	CoordMacId   []byte
	Status       string
	Description  string
	StatusCode   uint8
	ExtPanId     []byte
	Channel      uint8
	ShortAddr    uint16
	LinkStrength uint8
}

func convertNetworkInfo(xml XMLNetworkInfo) (NetworkInfo, error) {
	var ret NetworkInfo
	ret.DeviceMacId = xml.DeviceMacId
	ret.CoordMacId = xml.CoordMacId
	ret.Status = xml.Status
	ret.Description = xml.Description
	ret.ExtPanId = xml.ExtPanId
	statusCode, err := strconv.ParseUint(xml.StatusCode[2:], 16, 8)
	if err != nil {
		return ret, err
	}
	ret.StatusCode = uint8(statusCode)

	if xml.Channel == "" {
		ret.Channel = 0
	} else {
		channel, err := strconv.ParseInt(xml.Channel, 10, 8)
		if err != nil {
			return ret, err
		}
		ret.Channel = uint8(channel)
	}
	if xml.ShortAddr == "" {
		ret.ShortAddr = 0
	} else {
		shortAddr, err := strconv.ParseUint(xml.ShortAddr[2:], 16, 16)
		if err != nil {
			return ret, err
		}
		ret.ShortAddr = uint16(shortAddr)
	}
	linkStrength, err := strconv.ParseUint(xml.LinkStrength[2:], 16, 8)
	if err != nil {
		return ret, err
	}
	ret.LinkStrength = uint8(linkStrength)
	return ret, nil
}

type TimeCluster struct {
	XMLName     xml.Name `xml:"TimeCluster"`
	DeviceMacId []byte
	MeterMacId  []byte
	UTCTime     time.Time
	LocalTime   time.Time
}
type XMLTimeCluster struct {
	DeviceMacId []byte
	MeterMacId  []byte
	UTCTime     string
	LocalTime   string
}

func convertXMLTimeCluster(xml XMLTimeCluster) (TimeCluster, error) {
	var ret TimeCluster
	ret.DeviceMacId = xml.DeviceMacId
	ret.MeterMacId = xml.MeterMacId
	UTCtime, err := convertTimeStamp(xml.UTCTime)
	if err != nil {
		return ret, err
	}
	ret.UTCTime = UTCtime

	localTime, err := convertTimeStamp(xml.LocalTime)
	if err != nil {
		return ret, err
	}
	ret.LocalTime = localTime

	return ret, nil
}

func convertTimeStamp(t string) (time.Time, error) {
	epoch := time.Date(2000, time.January, 1, 0, 0, 0, 0, time.UTC)

	utcInt, err := strconv.ParseUint(t[2:], 16, 64)
	if err != nil {
		return time.Time{}, err
	}
	utcStr := strconv.FormatUint(utcInt, 10)
	utc, err := time.ParseDuration(utcStr + "s")
	if err != nil {
		return time.Time{}, err
	}

	return epoch.Add(utc), nil
}

type XMLMessageCluster struct {
	XMLName              xml.Name `xml:"MessageCluster"`
	DeviceMacId          []byte
	MeterMacId           []byte
	TimeStamp            string
	Id                   string
	Text                 string
	ConfirmationRequired string
	Confirmed            string
	Queue                string
}
type MessageCluster struct {
	DeviceMacId          []byte
	MeterMacId           []byte
	TimeStamp            uint32
	Id                   uint32
	Text                 string
	ConfirmationRequired string
	Confirmed            string
	Queue                string
}

type XMLPriceCluster struct {
	XMLName        xml.Name `xml:"PriceCluster"`
	DeviceMacId    []byte
	MeterMacId     []byte
	TimeStamp      string
	Price          string
	Currency       string
	TrailingDigits string
	Tier           string
	TierLabel      string
	RateLabel      string
}
type PriceCluster struct {
	DeviceMacId []byte
	MeterMacId  []byte
	TimeStamp   time.Time
	Price       decimal.Decimal
	Tier        uint8
	TierLabel   string
	RateLabel   string
	Currency    int
}

func convertXMLPriceCluster(xml XMLPriceCluster) (PriceCluster, error) {
	var ret PriceCluster
	var err error = nil
	ret.DeviceMacId = xml.DeviceMacId
	ret.MeterMacId = xml.MeterMacId
	ret.TierLabel = xml.TierLabel
	ret.RateLabel = xml.RateLabel
	ret.TimeStamp, err = convertTimeStamp(xml.TimeStamp)
	if err != nil {
		return ret, err
	}
	cur, err := strconv.ParseInt(xml.Currency[2:], 16, 64)
	if err != nil {
		return ret, err
	}
	ret.Currency = int(cur)

	tier, err := strconv.ParseUint(xml.Tier[2:], 16, 8)
	if err != nil {
		return ret, err
	}
	ret.Tier = uint8(tier)

	price, err := strconv.ParseInt(xml.Price[2:], 16, 32)
	if err != nil {
		return ret, err
	}
	trail, err := strconv.ParseInt(xml.TrailingDigits[2:], 16, 32)
	if err != nil {
		return ret, err
	}
	divisor := math.Pow10(int(trail))
	ret.Price = decimal.NewFromInt(price).Div(decimal.NewFromFloat(divisor))

	return ret, err
}

type XMLInstantaneousDemand struct {
	XMLName             xml.Name `xml:"InstantaneousDemand"`
	DeviceMacId         []byte
	MeterMacId          []byte
	TimeStamp           string
	Demand              string
	Multiplier          string
	Divisor             string
	DigitsRight         string
	DigitsLeft          string
	SuppressLeadingZero string
}
type InstantaneousDemand struct {
	DeviceMacId []byte
	MeterMacId  []byte
	TimeStamp   time.Time
	Demand      decimal.Decimal
}

func convertXMLInstantaneousDemand(xml XMLInstantaneousDemand) (InstantaneousDemand, error) {
	var ret InstantaneousDemand
	var err error = nil
	ret.DeviceMacId = xml.DeviceMacId
	ret.MeterMacId = xml.MeterMacId
	ret.TimeStamp, err = convertTimeStamp(xml.TimeStamp)
	if err != nil {
		return ret, err
	}
	rawDemand, err := strconv.ParseInt(xml.Demand[2:], 16, 32)
	if err != nil {
		return ret, err
	}
	multiplier, err := strconv.ParseInt(xml.Multiplier[2:], 16, 32)
	if err != nil {
		return ret, err
	}
	divisor, err := strconv.ParseInt(xml.Divisor[2:], 16, 32)
	if err != nil {
		return ret, err
	}

	if multiplier == 0 {
		multiplier = 1
	}
	if divisor == 0 {
		divisor = 1
	}

	demand := decimal.NewFromInt(rawDemand).Mul(decimal.NewFromInt(multiplier)).Div(decimal.NewFromInt(divisor))
	ret.Demand = demand
	return ret, err
}

type CurrentSummationDelivered struct {
	XMLName            xml.Name `xml:"CurrentSummationDelivered"`
	DeviceMacId        []byte
	MeterMacId         []byte
	TimeStamp          time.Time
	SummationDelivered decimal.Decimal
	SummationReceived  decimal.Decimal
}
type XMLCurrentSummationDelivered struct {
	XMLName             xml.Name `xml:"CurrentSummationDelivered"`
	DeviceMacId         []byte
	MeterMacId          []byte
	TimeStamp           string
	SummationDelivered  string
	SummationReceived   string
	Multiplier          string
	Divisor             string
	DigitsRight         string
	DigitsLeft          string
	SuppressLeadingZero string
}

func convertXMLCurrentSummationDelivered(xml XMLCurrentSummationDelivered) (CurrentSummationDelivered, error) {
	var ret CurrentSummationDelivered
	var err error = nil
	ret.DeviceMacId = xml.DeviceMacId
	ret.MeterMacId = xml.MeterMacId
	ret.TimeStamp, err = convertTimeStamp(xml.TimeStamp)
	if err != nil {
		return ret, err
	}
	/*
		Demand              string
		Multiplier          string
		Divisor             string
		DigitsRight         string
		DigitsLeft          string
		SuppressLeadingZero string
	*/
	sumDeliv, err := strconv.ParseInt(xml.SummationDelivered[2:], 16, 64)
	if err != nil {
		return ret, err
	}
	sumRec, err := strconv.ParseInt(xml.SummationReceived[2:], 16, 64)
	if err != nil {
		return ret, err
	}
	multiplier, err := strconv.ParseInt(xml.Multiplier[2:], 16, 32)
	if err != nil {
		return ret, err
	}
	divisor, err := strconv.ParseInt(xml.Divisor[2:], 16, 32)
	if err != nil {
		return ret, err
	}

	if multiplier == 0 {
		multiplier = 1
	}
	if divisor == 0 {
		divisor = 1
	}

	ret.SummationDelivered = decimal.NewFromInt(sumDeliv).Mul(decimal.NewFromInt(multiplier)).Div(decimal.NewFromInt(divisor))
	ret.SummationReceived = decimal.NewFromInt(sumRec).Mul(decimal.NewFromInt(multiplier)).Div(decimal.NewFromInt(divisor))

	return ret, err
}

type CurrentPeriodUsage struct {
	XMLName      xml.Name `xml:"CurrentPeriodUsage"`
	DeviceMacId  []byte
	MeterMacId   []byte
	TimeStamp    time.Time
	CurrentUsage decimal.Decimal
	StartDate    time.Time
}
type XMLCurrentPeriodUsage struct {
	XMLName             xml.Name `xml:"CurrentPeriodUsage"`
	DeviceMacId         []byte
	MeterMacId          []byte
	TimeStamp           string
	CurrentUsage        string
	Multiplier          string
	Divisor             string
	DigitsRight         string
	DigitsLeft          string
	SuppressLeadingZero string
	StartDate           string
}

func convertXMLCurrentPeriodUsage(xml XMLCurrentPeriodUsage) (CurrentPeriodUsage, error) {
	var ret CurrentPeriodUsage
	var err error = nil
	ret.DeviceMacId = xml.DeviceMacId
	ret.MeterMacId = xml.MeterMacId
	ret.TimeStamp, err = convertTimeStamp(xml.TimeStamp)
	if err != nil {
		return ret, err
	}
	ret.StartDate, err = convertTimeStamp(xml.StartDate)
	if err != nil {
		return ret, err
	}
	current, err := strconv.ParseInt(xml.CurrentUsage[2:], 16, 32)
	if err != nil {
		return ret, err
	}
	multiplier, err := strconv.ParseInt(xml.Multiplier[2:], 16, 32)
	if err != nil {
		return ret, err
	}
	divisor, err := strconv.ParseInt(xml.Divisor[2:], 16, 32)
	if err != nil {
		return ret, err
	}

	if multiplier == 0 {
		multiplier = 1
	}
	if divisor == 0 {
		divisor = 1
	}

	ret.CurrentUsage = decimal.NewFromInt(current).Mul(decimal.NewFromInt(multiplier)).Div(decimal.NewFromInt(divisor))

	return ret, err
}

type LastPeriodUsage struct {
	XMLName     xml.Name `xml:"LastPeriodUsage"`
	DeviceMacId []byte
	MeterMacId  []byte
	TimeStamp   time.Time
	LastUsage   decimal.Decimal
	StartDate   time.Time
	EndDate     time.Time
}

type XMLLastPeriodUsage struct {
	XMLName             xml.Name `xml:"LastPeriodUsage"`
	DeviceMacId         []byte
	MeterMacId          []byte
	TimeStamp           string
	LastUsage           string
	Multiplier          string
	Divisor             string
	DigitsRight         string
	DigitsLeft          string
	SuppressLeadingZero string
	StartDate           string
	EndDate             string
}

func convertXMLLastPeriodUsage(xml XMLLastPeriodUsage) (LastPeriodUsage, error) {
	var ret LastPeriodUsage
	var err error = nil
	ret.DeviceMacId = xml.DeviceMacId
	ret.MeterMacId = xml.MeterMacId
	ret.TimeStamp, err = convertTimeStamp(xml.TimeStamp)
	if err != nil {
		return ret, err
	}
	ret.StartDate, err = convertTimeStamp(xml.StartDate)
	if err != nil {
		return ret, err
	}
	ret.EndDate, err = convertTimeStamp(xml.EndDate)
	if err != nil {
		return ret, err
	}
	lastUsage, err := strconv.ParseInt(xml.LastUsage[2:], 16, 32)
	if err != nil {
		return ret, err
	}
	multiplier, err := strconv.ParseInt(xml.Multiplier[2:], 16, 32)
	if err != nil {
		return ret, err
	}
	divisor, err := strconv.ParseInt(xml.Divisor[2:], 16, 32)
	if err != nil {
		return ret, err
	}

	if multiplier == 0 {
		multiplier = 1
	}
	if divisor == 0 {
		divisor = 1
	}

	ret.LastUsage = decimal.NewFromInt(lastUsage).Mul(decimal.NewFromInt(multiplier)).Div(decimal.NewFromInt(divisor))

	return ret, err
}

type ProfileData struct {
	XMLName                 xml.Name `xml:"ProfileData"`
	DeviceMacId             []byte
	MeterMacId              []byte
	EndTime                 time.Time
	Status                  uint8
	ProfileIntervalPeriod   uint8
	NumberOfPointsDelivered uint8
	IntervalData            []uint32
}
