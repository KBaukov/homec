package db

import (
	"errors"
	"homec/ent"
	"log"
	"time"

	"github.com/jmoiron/sqlx"
)


var (
	RoomData     = make(map[string]ent.SensorsData)
)
const (
	authQuery = `SELECT * FROM hc.users WHERE users.login = ? AND users.pass = ?`

	getUsersQuery = `SELECT * FROM hc.users ORDER BY id`
	addUserQuery  = `INSERT INTO hc.users (login, pass, user_type, active_flag, last_visit, id) VALUES (?,?,?,?,?,?)`
	updUserQuery  = `UPDATE hc.users SET login=?, pass=?, user_type=?, active_flag=?, last_visit=? WHERE id=?`
	delUserQuery  = `DELETE FROM hc.users WHERE id=?`
	//lastUsrIdQuery = `SELECT max(id) FROM hc.users`

	getDevicesQuery = `SELECT * FROM hc.devices ORDER BY id`
	addDeviceQuery  = `INSERT INTO hc.devices (type, name, ip, active_flag, description, id) VALUES (?,?,?,?,?,?)`
	updDeviceQuery  = `UPDATE hc.devices SET type = ?, name = ?, ip = ?, active_flag = ?, description = ? WHERE id=?`
	delDeviceQuery  = `DELETE FROM hc.devices WHERE id=?`

	kotelDevIdQuery       = `SELECT id, name FROM hc.devices WHERE type="KotelController"`
	getKotelDataQuery     = "SELECT DEVICE_ID ,`TO`, TP, KW, PR, DEST_TO, DEST_TP, DEST_KW, DEST_PR, DEST_TC, STAGE FROM hc.kotel"
	updKotelDevIdQuery    = `UPDATE hc.kotel SET device_id = ? WHERE 1`
	updKotelDestDataQuery = "UPDATE hc.kotel SET dest_to=?, dest_tp=?, dest_kw=?, dest_pr=?, dest_tc=?, stage=? WHERE 1"
	updKotelMeshDataQuery = "UPDATE hc.kotel SET `to`= ?, tp= ?, kw= ?, pr= ? WHERE 1"

	getMapsQuery = `SELECT * FROM hc.maps ORDER BY id`
	addMapQuery  = `INSERT INTO hc.maps (title, pict, w, h, description, id) VALUES (?,?,?,?,?,?)`
	updMapQuery  = `UPDATE hc.maps SET title=?, pict=?, w=?, h=?, description=?  WHERE id=?`
	delMapQuery  = `DELETE FROM hc.maps WHERE id=?`
	//lastMapIdQuery = `SELECT max(id) as id FROM hc.maps`

	getMapSensorsQuery = `SELECT * FROM hc.map_sensors WHERE map_id= ? ORDER BY id`
	addMapSensorQuery  = `INSERT INTO hc.map_sensors (map_id, device_id, type, xk, yk, pict, description, id) VALUES (?,?,?,?,?,?,?,?)`
	updMapSensorQuery  = `UPDATE hc.map_sensors SET map_id=?, device_id=?, type=?, xk=?, yk=?, pict=?, description=? WHERE id=?`
	delMapSensorQuery  = `DELETE FROM hc.map_sensors WHERE id=?`
	//lastMapSensorIdQuery = `SELECT max(id) as id FROM hc.map_sensors`

	addRoomDataQuery = `INSERT INTO hc.room_data (device_id, sensor_type, t, h, p, date) VALUES (?,?,?,?,?,?)`
	getRoomDataQuery = `SELECT * FROM hc.room_data WHERE device_id= ? ORDER BY date DESC LIMIT 1`
	getRoomDataStat  = `SELECT * FROM hc.room_data WHERE device_id= ? AND year(date) = year(now()) AND week(date, 1) = week(now(), 1) limit 50`
)

// database структура подключения к базе данных
type Database struct {
	Conn *sqlx.DB
}

// dbService представляет интерфейс взаимодействия с базой данных
type DbService interface {
	Auth(string, string) ([]*ent.User, error)

	GetLastId(table string) (int, error)

	GetUsers() ([]ent.User, error)
	UpdUser(id int, login string, pass string, userType string, actFlag string, lastV time.Time) (bool, error)
	DelUser(id int) (bool, error)

	GetDevices() ([]ent.Device, error)
	UpdDevice(id int, devType string, devName string, ip string, actFlag string, descr string) (bool, error)
	DelDevice(id int) (bool, error)

	GetKotelID() (int, string, error)
	GetKotelData() (ent.KotelData, error)
	UpdKotelMeshData(to float64, tp float64, kw int, pr float64) error
	UpdKotelDestData(destto float64, desttp float64, destkw int, destpr float64, destc float64, stage string) error

	GetMaps() ([]ent.Map, error)
	UpdMap(id int, title string, pict string, w int, h int, descr string) (bool, error)
	DelMap(id int) (bool, error)

	GetMapSensors(mapId int) ([]ent.MapSensor, error)
	UpdMapSensor(id int, mapId int, devId int, sensorType string, xk float64, yk float64, pict string, descr string) (bool, error)
	DelMapSensor(id int) (bool, error)

	AddRoomData(data ent.SensorsData) (bool, error)
	GetRoomData(devId string) (ent.SensorsData, error)
	GetRoomDataStat(devId string) ([]ent.SensorsData, error)
}

// newDB открывает соединение с базой данных
func NewDB(connectionString string) (Database, error) {
	dbConn, err := sqlx.Open("mysql", connectionString)
	log.Println(connectionString)
	return Database{Conn: dbConn}, err
}

//#################################################################
func (db Database) Auth(login string, password string) ([]*ent.User, error) {

	users := make([]*ent.User, 0)
	//err := db.Conn.Select(&users, authQuery, login, password)

	stmt, err := db.Conn.Prepare(authQuery)
	if err != nil {
		return users, err
	}
	defer stmt.Close()

	rows, err := stmt.Query(login, password)

	for rows.Next() {
		var uid int
		var login string
		var pass string
		var active string
		var userType string
		var lastV time.Time
		err = rows.Scan(&uid, &login, &pass, &active, &userType, &lastV)
		if err != nil {
			return users, err
		}
		u := ent.User{uid, login, pass, active, userType, lastV}
		users = append(users, &u)
	}

	return users, err
}

func (db Database) GetLastId(table string) (int, error) {
	var lastId int
	stmt, err := db.Conn.Prepare("SELECT max(id) as id FROM hc." + table)
	if err != nil {
		return -1, err
	}
	defer stmt.Close()

	rows, err := stmt.Query()
	for rows.Next() {
		err = rows.Scan(&lastId)
		if err != nil {
			return -1, err
		}
	}

	return lastId, err
}

// ############## Users ############################
func (db Database) GetUsers() ([]ent.User, error) {
	users := make([]ent.User, 0)
	stmt, err := db.Conn.Prepare(getUsersQuery)
	if err != nil {
		return users, err
	}
	defer stmt.Close()

	rows, err := stmt.Query()

	for rows.Next() {
		var (
			uid      int
			login    string
			pass     string
			active   string
			userType string
			lastV    time.Time
		)
		err = rows.Scan(&uid, &login, &pass, &active, &userType, &lastV)
		if err != nil {
			return users, err
		}
		u := ent.User{uid, login, pass, active, userType, lastV}
		users = append(users, u)
	}

	return users, err
}

func (db Database) UpdUser(id int, login string, pass string, userType string, actFlag string, lastV time.Time) (bool, error) {

	var lastId int

	execQuery := updUserQuery

	lastId, err := db.GetLastId("users")
	if err != nil {
		return false, err
	}

	if id > lastId {
		execQuery = addUserQuery
	}

	stmt, err := db.Conn.Prepare(execQuery)
	if err != nil {
		return false, err
	}

	_, err = stmt.Exec(login, pass, userType, actFlag, lastV, id)
	if err != nil {
		return false, err
	}

	return true, err
}

func (db Database) DelUser(id int) (bool, error) {

	stmt, err := db.Conn.Prepare(delUserQuery)
	if err != nil {
		return false, err
	}
	defer stmt.Close()

	_, err = stmt.Exec(id)
	if err != nil {
		return false, err
	}

	return true, err
}

//##################### Devices ########################################

func (db Database) GetDevices() ([]ent.Device, error) {
	devices := make([]ent.Device, 0)
	stmt, err := db.Conn.Prepare(getDevicesQuery)
	if err != nil {
		return devices, err
	}
	defer stmt.Close()

	rows, err := stmt.Query()

	for rows.Next() {
		var (
			did    int
			typ    string
			name   string
			ip     string
			active string
			descr  string
		)

		err = rows.Scan(&did, &typ, &name, &ip, &active, &descr)
		if err != nil {
			return devices, err
		}
		d := ent.Device{did, typ, name, ip, active, descr}
		devices = append(devices, d)
	}

	return devices, err
}

func (db Database) UpdDevice(id int, devType string, devName string, ip string, actFlag string, descr string) (bool, error) {

	var (
		lastId int
		kName  string
	)
	if devType == "KotelController" {
		_, kName, _ = db.GetKotelID()
		if kName == devName {
			return false, errors.New("Устройство с типом KotelController уже существует. Такое устройсто может быть только одно.")
		} else {
			stmt, err := db.Conn.Prepare(updKotelDevIdQuery)
			if err != nil {
				return false, err
			}
			_, err = stmt.Exec(devName)
			if err != nil {
				return false, err
			}
		}

	}

	execQuery := updDeviceQuery

	lastId, err := db.GetLastId("devices")
	if err != nil {
		return false, err
	}

	if id > lastId {
		execQuery = addDeviceQuery
	}

	stmt, err := db.Conn.Prepare(execQuery)
	if err != nil {
		return false, err
	}

	_, err = stmt.Exec(devType, devName, ip, actFlag, descr, id)
	if err != nil {
		return false, err
	}

	return true, err
}

func (db Database) DelDevice(id int) (bool, error) {

	stmt, err := db.Conn.Prepare(delDeviceQuery)
	if err != nil {
		return false, err
	}
	defer stmt.Close()

	_, err = stmt.Exec(id)
	if err != nil {
		return false, err
	}

	return true, err
}

//################# Kotel #####################
func (db Database) GetKotelID() (int, string, error) {
	var (
		kId   int
		kName string
	)
	stmt, err := db.Conn.Prepare(kotelDevIdQuery)
	if err != nil {
		return kId, kName, err
	}
	defer stmt.Close()

	rows, err := stmt.Query()

	for rows.Next() {
		err = rows.Scan(&kId, &kName)
		if err != nil {
			return kId, kName, err
		}
	}

	return kId, kName, err
}

func (db Database) GetKotelData() (ent.KotelData, error) {
	var (
		kData ent.KotelData
	)

	stmt, err := db.Conn.Prepare(getKotelDataQuery)
	if err != nil {
		return kData, err
	}
	defer stmt.Close()

	rows, err := stmt.Query()
	if err != nil {
		return kData, err
	}

	for rows.Next() {
		var (
			deviceId string
			to       float64
			tp       float64
			kw       int
			pr       float64
			destto   float64
			desttp   float64
			destkw   int
			destpr   float64
			desttc   float64
			stage    string
		)
		err = rows.Scan(&deviceId, &to, &tp, &kw, &pr, &destto, &desttp, &destkw, &destpr, &desttc, &stage)
		if err != nil {
			return kData, err
		}

		kData = ent.KotelData{deviceId, to, tp, kw, pr, destto, desttp, destkw, destpr, desttc, stage}
	}

	return kData, err
}

func (db Database) UpdKotelDestData(destto float64, desttp float64, destkw int, destpr float64, destc float64, stage string) error {

	stmt, err := db.Conn.Prepare(updKotelDestDataQuery)
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(destto, desttp, destkw, destpr, destc, stage)
	if err != nil {
		return err
	}

	return err
}

func (db Database) UpdKotelMeshData(to float64, tp float64, kw int, pr float64) error {

	stmt, err := db.Conn.Prepare(updKotelMeshDataQuery)
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(to, tp, kw, pr)
	if err != nil {
		return err
	}

	return err
}

//################## Maps #########################
func (db Database) GetMaps() ([]ent.Map, error) {
	maps := make([]ent.Map, 0)
	stmt, err := db.Conn.Prepare(getMapsQuery)
	if err != nil {
		return maps, err
	}
	defer stmt.Close()

	rows, err := stmt.Query()

	for rows.Next() {
		var (
			id    int
			title string
			pict  string
			w     int
			h     int
			descr string
		)

		err = rows.Scan(&id, &title, &pict, &w, &h, &descr)
		if err != nil {
			return maps, err
		}
		m := ent.Map{id, title, pict, w, h, descr}
		maps = append(maps, m)
	}

	return maps, err
}

func (db Database) UpdMap(id int, title string, pict string, w int, h int, descr string) (bool, error) {

	var lastId int

	execQuery := updMapQuery

	lastId, err := db.GetLastId("maps")
	if err != nil {
		return false, err
	}

	if id > lastId {
		execQuery = addMapQuery
	}

	stmt, err := db.Conn.Prepare(execQuery)
	if err != nil {
		return false, err
	}

	_, err = stmt.Exec(title, pict, w, h, descr, id)
	if err != nil {
		return false, err
	}

	return true, err
}

func (db Database) DelMap(id int) (bool, error) {

	stmt, err := db.Conn.Prepare(delMapQuery)
	if err != nil {
		return false, err
	}
	defer stmt.Close()

	_, err = stmt.Exec(id)
	if err != nil {
		return false, err
	}

	return true, err
}

//###################### MapSensors #####################
func (db Database) GetMapSensors(mapId int) ([]ent.MapSensor, error) {
	mapSens := make([]ent.MapSensor, 0)
	stmt, err := db.Conn.Prepare(getMapSensorsQuery)
	if err != nil {
		return mapSens, err
	}
	defer stmt.Close()

	rows, err := stmt.Query(mapId)

	for rows.Next() {
		var (
			id       int
			mapId    int
			devId    int
			sensType string
			xk       float32
			yk       float32
			pict     string
			descr    string
		)

		err = rows.Scan(&id, &mapId, &devId, &sensType, &xk, &yk, &pict, &descr)
		if err != nil {
			return mapSens, err
		}
		s := ent.MapSensor{id, mapId, devId, sensType, xk, yk, pict, descr}
		mapSens = append(mapSens, s)
	}

	return mapSens, err
}

func (db Database) UpdMapSensor(id int, mapId int, devId int, sensorType string, xk float64, yk float64, pict string, descr string) (bool, error) {

	var lastId int

	execQuery := updMapSensorQuery

	lastId, err := db.GetLastId("map_sensors")
	if err != nil {
		return false, err
	}

	if id > lastId {
		execQuery = addMapSensorQuery
	}

	stmt, err := db.Conn.Prepare(execQuery)
	if err != nil {
		return false, err
	}

	_, err = stmt.Exec(mapId, devId, sensorType, xk, yk, pict, descr, id)
	if err != nil {
		return false, err
	}

	return true, err
}

func (db Database) DelMapSensor(id int) (bool, error) {

	stmt, err := db.Conn.Prepare(delMapSensorQuery)
	if err != nil {
		return false, err
	}
	defer stmt.Close()

	_, err = stmt.Exec(id)
	if err != nil {
		return false, err
	}

	return true, err
}

// ########################### ROOM DATA ######################################

func (db Database) AddRoomData(data ent.SensorsData) (bool, error) {
	stmt, err := db.Conn.Prepare(addRoomDataQuery)
	if err != nil {
		return false, err
	}
	defer stmt.Close()

	_, err = stmt.Exec(data.DEVICE_ID, data.SENSOR_TYPE, data.T, data.H, data.P, data.DATE)
	if err != nil {
		return false, err
	}

	RoomData[data.DEVICE_ID] = data

	return true, err
}

func (db Database) GetRoomData(devId string) (ent.SensorsData, error) {
	var err error
	data, ok := RoomData[devId]
	if !ok  {
		err = errors.New("нет данных")
	}
	return data, err
}

func (db Database) GetRoomDataStat(devId string) ([]ent.SensorsData, error) {
	var data = make([]ent.SensorsData, 0)

	stmt, err := db.Conn.Prepare(getRoomDataStat)
	if err != nil {
		return data, err
	}
	defer stmt.Close()

	rows, err := stmt.Query(devId)

	for rows.Next() {
		var (
			dId    string
			sType string
			t       float64
			h       float64
			p       float64
			d 		time.Time
		)

		err = rows.Scan(&dId, &sType, &t, &h, &p, &d)
		if err != nil {
			return data, err
		}
		sData := ent.SensorsData{dId, sType, t, h, p, d}
		data = append(data, sData)
	}

	return data, err
}

