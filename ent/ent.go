package ent

import (
	"time"
)

type User struct {
	ID          int       `json:"id"`
	LOGIN       string    `json:"login"`
	PASS        string    `json:"pass"`
	ACTIVE_FLAG string    `json:"active_flag"`
	USER_TYPE   string    `json:"user_type"`
	LAST_VISIT  time.Time `json:"last_visit"`
}

type Device struct {
	ID          int    `json:"id"`
	TYPE        string `json:"type"`
	NAME        string `json:"name"`
	IP          string `json:"ip"`
	ACTIVE_FLAG string `json:"active_flag"`
	DESCRIPTION string `json:"description"`
}

type ApiError struct {
	ERROR_CODE   int    `json:"errorCode"`
	EROR_MESSAGE string `json:"errorMEssage"`
}

type ApiResp struct {
	SUCCESS bool        `json:"success"`
	DATA    interface{} `json:"data"`
	MSG     string      `json:"msg"`
}

type Map struct {
	ID          int    `json:"id"`
	TITLE       string `json:"title"`
	PICT        string `json:"pict"`
	W           int    `json:"w"`
	H           int    `json:"h"`
	DESCRIPTION string `json:"description"`
}

type MapSensor struct {
	ID          int     `json:"id"`
	MAP_ID      int     `json:"map_id"`
	DEVICE_ID   int     `json:"device_id"`
	TYPE        string  `json:"type"`
	XK          float32 `json:"xk"`
	YK          float32 `json:"yk"`
	PICT        string  `json:"pict"`
	DESCRIPTION string  `json:"description"`
}

type KotelData struct {
	DEVICE_ID string  `json:"device_id"`
	TO        float64 `json:"to"`
	TP        float64 `json:"tp"`
	KW        int     `json:"kw"`
	PR        float64 `json:"pr"`
	DESTTO    float64 `json:"destTo"`
	DESTTP    float64 `json:"destTp"`
	DESTKW    int     `json:"destKw"`
	DESTPR    float64 `json:"destPr"`
	DESTС     float64 `json:"destTc"`
	STAGE	  string  `json:"stage"`
}

type KotelStatData struct {
	TO        float64 `json:"to"`
	TP        float64 `json:"tp"`
	KW        int     `json:"kw"`
	PR        float64 `json:"pr"`
	DATE	  time.Time `json:"date"`
}

type FloorData struct {
	DEVICE_ID string  `json:"device_id"`
	TF        float64 `json:"tf"`
	TA        float64 `json:"ta"`
	HA        float64 `json:"ha"`
	DESTTF    float64 `json:"destTf"`
	STAGE	  string  `json:"stage"`
}

type SensorsData struct {
	DEVICE_ID   string    `json:"device_id"`
	SENSOR_TYPE string    `json:"sensor_type"`
	T           float64   `json:"t"`
	H           float64   `json:"h"`
	P           float64   `json:"p"`
	DATE        time.Time `json:"date"`
}

type ResendMessage struct {
	ACTION 		string `json"action"`
	RECIPIENT   string `json"recipient"`
	SENDER		string `json"sender"`
	MESSAGE   	string `json:"msg"`
}

type WsSendData struct {
	ACTION string      `json"action"`
	TYPE   string      `json"type"`
	DATA   interface{} `json:"data"`
}

