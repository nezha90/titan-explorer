package statistics

import (
	"context"
	"fmt"
	"github.com/gnasnik/titan-explorer/core/dao"
	"github.com/gnasnik/titan-explorer/core/generated/model"
	"github.com/golang-module/carbon/v2"
	"strconv"
	"time"
)

const (
	TimeFormatYMD = "2006-01-02"
)

var (
	DeviceIDAndUserId map[string]string
)

func addDeviceInfoHours(ctx context.Context, deviceInfo []*model.DeviceInfo) error {
	log.Info("start fetch device info hours")
	start := time.Now()
	defer func() {
		log.Infof("fetch device info hours done, cost: %v", time.Since(start))
	}()

	var upsertDevice []*model.DeviceInfoHour
	for _, device := range deviceInfo {
		var deviceInfoHour model.DeviceInfoHour
		deviceInfoHour.Time = start
		deviceInfoHour.DiskUsage = device.DiskUsage
		deviceInfoHour.DeviceID = device.DeviceID
		deviceInfoHour.PkgLossRatio = device.PkgLossRatio
		deviceInfoHour.HourIncome = device.CumulativeProfit
		deviceInfoHour.OnlineTime = device.OnlineTime
		deviceInfoHour.Latency = device.Latency
		deviceInfoHour.DiskUsage = device.DiskUsage
		_, ok := DeviceIDAndUserId[deviceInfoHour.DeviceID]
		if ok {
			deviceInfoHour.UserID = DeviceIDAndUserId[deviceInfoHour.DeviceID]
		}

		upsertDevice = append(upsertDevice, &deviceInfoHour)
	}

	err := dao.BulkUpsertDeviceInfoHours(context.Background(), upsertDevice)
	if err != nil {
		log.Errorf("bulk upsert device info: %v", err)
	}
	return nil
}

func Str2Float64(s string) float64 {
	ret, err := strconv.ParseFloat(s, 64)
	if err != nil {
		log.Error(err.Error())
		return 0.00
	}
	return ret
}

func QueryDataByDate(DateFrom, DateTo string) []map[string]string {

	sqlClause := fmt.Sprintf("select device_id, sum(income) as income,online_time from device_info_daily "+
		"where  time>='%s' and time<='%s' group by device_id;", DateFrom, DateTo)
	if DateFrom == "" {
		sqlClause = fmt.Sprintf("select device_id, sum(income) as income,online_time from device_info_daily " +
			" group by device_id;")
	}
	data, err := dao.GetQueryDataList(sqlClause)
	if err != nil {
		log.Error(err.Error())
		return nil
	}

	return data
}

func (s *Statistic) SumDeviceInfoDaily() error {
	log.Info("start sum device info daily")
	start := time.Now()
	defer func() {
		log.Infof("sum device info daily done, cost: %v", time.Since(start))
	}()

	startOfTodayTime := carbon.Now().StartOfDay().String()
	endOfTodayTime := carbon.Now().EndOfDay().String()
	sqlClause := fmt.Sprintf("select user_id, device_id, date_format(time, '%%Y-%%m-%%d') as date, avg(nat_ratio) as nat_ratio, avg(disk_usage) as disk_usage, avg(latency) as latency, avg(pkg_loss_ratio) as pkg_loss_ratio, max(hour_income) as hour_income_max, min(hour_income) as hour_income_min ,max(online_time) as online_time_max,min(online_time) as online_time_min from device_info_hour "+
		"where time>='%s' and time<='%s' group by date, device_id", startOfTodayTime, endOfTodayTime)
	datas, err := dao.GetQueryDataList(sqlClause)
	if err != nil {
		log.Error(err.Error())
		return err
	}

	var dailyInfos []*model.DeviceInfoDaily
	for _, data := range datas {
		var daily model.DeviceInfoDaily
		daily.Time, _ = time.Parse(TimeFormatYMD, data["date"])
		daily.DiskUsage = Str2Float64(data["disk_usage"])
		daily.NatRatio = Str2Float64(data["nat_ratio"])
		daily.Income = Str2Float64(data["hour_income_max"]) - Str2Float64(data["hour_income_min"])
		daily.OnlineTime = Str2Float64(data["online_time_max"]) - Str2Float64(data["online_time_min"])
		daily.PkgLossRatio = Str2Float64(data["pkg_loss_ratio"])
		daily.Latency = Str2Float64(data["latency"])
		daily.DeviceID = data["device_id"]
		daily.UserID = data["user_id"]
		daily.CreatedAt = time.Now()
		daily.UpdatedAt = time.Now()
		dailyInfos = append(dailyInfos, &daily)
	}

	err = dao.BulkUpsertDeviceInfoDaily(context.Background(), dailyInfos)
	if err != nil {
		log.Errorf("upsert device info daily: %v", err)
		return err
	}

	return nil
}

func (s *Statistic) SumDeviceInfoProfit() error {
	log.Info("start sum device info profit")
	start := time.Now()
	defer func() {
		log.Infof("sum device info profit done, cost: %v", time.Since(start))
	}()

	updatedDevices := make(map[string]*model.DeviceInfo)

	startOfTodayTime := carbon.Now().StartOfDay().String()
	endOfTodayTime := carbon.Now().EndOfDay().String()
	startOfYesterday := carbon.Yesterday().StartOfDay().String()
	endOfYesterday := carbon.Yesterday().EndOfDay().String()
	dataY := QueryDataByDate(startOfYesterday, endOfYesterday)
	for _, data := range dataY {
		_, ok := updatedDevices[data["device_id"]]
		if !ok {
			updatedDevices[data["device_id"]] = &model.DeviceInfo{
				DeviceID: data["device_id"],
			}
		}
		updatedDevices[data["device_id"]].YesterdayProfit = Str2Float64(data["income"])
	}

	startOfWeekTime := carbon.Now().SubDays(6).StartOfDay().String()
	dataS := QueryDataByDate(startOfWeekTime, endOfTodayTime)

	for _, data := range dataS {
		_, ok := updatedDevices[data["device_id"]]
		if !ok {
			updatedDevices[data["device_id"]] = &model.DeviceInfo{
				DeviceID: data["device_id"],
			}
		}
		updatedDevices[data["device_id"]].SevenDaysProfit = Str2Float64(data["income"])
	}

	startOfMonthTime := carbon.Now().SubDays(29).StartOfDay().String()
	dataM := QueryDataByDate(startOfMonthTime, endOfTodayTime)

	for _, data := range dataM {
		_, ok := updatedDevices[data["device_id"]]
		if !ok {
			updatedDevices[data["device_id"]] = &model.DeviceInfo{
				DeviceID: data["device_id"],
			}
		}
		updatedDevices[data["device_id"]].MonthProfit = Str2Float64(data["income"])
	}

	dataT := QueryDataByDate(startOfTodayTime, endOfTodayTime)
	for _, data := range dataT {
		_, ok := updatedDevices[data["device_id"]]
		if !ok {
			updatedDevices[data["device_id"]] = &model.DeviceInfo{
				DeviceID: data["device_id"],
			}
		}
		updatedDevices[data["device_id"]].TodayProfit = Str2Float64(data["income"])
		updatedDevices[data["device_id"]].TodayOnlineTime = Str2Float64(data["online_time"])
	}

	var deviceInfos []*model.DeviceInfo
	for _, deviceInfo := range updatedDevices {
		deviceInfos = append(deviceInfos, deviceInfo)
	}

	if err := dao.BulkUpdateDeviceInfo(context.Background(), deviceInfos); err != nil {
		log.Errorf("bulk update device: %v", err)
	}

	return nil
}