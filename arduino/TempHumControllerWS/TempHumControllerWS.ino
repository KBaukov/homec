#include <Arduino.h>
#include <Hash.h>
#include <ESP8266WiFi.h>
#include <WebSocketsClient.h>
#include <ArduinoJson.h>
#include <DHT.h>

#include <WEMOS_SHT3X.h>
#include <Wire.h>  // Include Wire if you're using I2C
#include <SFE_MicroOLED.h>  // Include the SFE_MicroOLED library

#define POWER_MODE  0 // режим питания, 0 - внешнее, 1 - паразитное
#define PIN_RESET 255  //
#define DC_JUMPER 0  // I2C Addres: 0 - 0x3C, 1 - 0x3D
#define DHTPIN D4
#define USE_SERIAL Serial
#define POWER_MODE  0
#define EXTC_BUTT  D1

#ifdef ESP8266
extern "C" {
  #include "user_interface.h"
}
#endif

// SETTINGS

DHT dht(DHTPIN, DHT21);

const char* wlan_ssid             = "WF";
const char* wlan_password         = "k0k0JambA";

const char* ws_host               = "strobo.ddns.net";
const int   ws_port               = 443;
const char* stompUrl              = "/ws"; // don't forget the leading "/" !!!

String deviceId = wifi_station_get_hostname();
String headersString = "Origin: HomeControlApp\r\nDeviceId: "+deviceId;
const char* headers =  (const char *) headersString.c_str();

// VARIABLES

WebSocketsClient webSocket;

byte bufData[9];  // буфер данных

float temp = 0;
float hum  = 0;

long wait = 10;
long count = 0;
long pressDuration = 100;
int statusId = 0;


void setup() {
  
    USE_SERIAL.begin(115200);
    USE_SERIAL.println();USE_SERIAL.println();USE_SERIAL.println();

    // connect to WiFi
    USE_SERIAL.print("Logging into WLAN: "); 
    Serial.print(wlan_ssid); 
    Serial.print(" ...");
    
    WiFi.mode(WIFI_STA);
    WiFi.begin(wlan_ssid, wlan_password);

    while (WiFi.status() != WL_CONNECTED) {
        delay(500);
        USE_SERIAL.print(".");
    }
    USE_SERIAL.println(" success.");
    USE_SERIAL.print("IP: "); USE_SERIAL.println(WiFi.localIP());


    // connect to websocket
    webSocket.beginSSL(ws_host, ws_port, stompUrl);
    webSocket.setExtraHeaders(headers);
    webSocket.onEvent(webSocketEvent);

    
}

void loop() {

  temp = dht.readTemperature(); hum = dht.readHumidity();
  //temp = 28.35; hum = 54.56;
  if (isnan(hum) || isnan(temp)) {    
    Serial.println("DHT Error!");
  } else {
    Serial.print("Temp="); Serial.print(temp);
    Serial.print("C; Humidity="); Serial.print(hum); 
    Serial.print("%; statusId="); Serial.print(statusId);
    Serial.print("; count="); Serial.println(count);
    
  }
/*
  temp = 25.55; hum = 35.55;
  Serial.print("Temp="); Serial.print(temp);
  Serial.print("C; Humidity="); Serial.print(hum); 
  Serial.print("%; statusId="); Serial.print(statusId);
  Serial.print("; count="); Serial.println(count);
*/
  
  if(statusId == 1 && count >= wait ) {
    String msg = "{\"action\":\"datasend\", \"type\":\"roomdata\", \"data\":{ \"device_id\":\""+deviceId+"\", "
      + "\"sensor_type\":\"tempHym\", "
      + "\"t\":" + temp +", "
      + "\"h\":" + hum +", "
      + "\"p\":0.0, "
      + "\"date\":\"2018-11-22T15:04:05Z\""      
    +" } }";
    
    sendMessage(msg);

    count = 0;
  }
/*  
   if(USE_SERIAL.available()>0) {
    String msg = "";
     while (USE_SERIAL.available() > 0) {
       msg += (char) USE_SERIAL.read();
     }
     sendMessage(msg);
   }
*/    
   count++;
   webSocket.loop();
   Serial.println("====================================================");
   delay(1000);
}

// FUNCTIONS
void sendMessage(String & msg) {
    
    webSocket.sendTXT(msg.c_str(), msg.length());
    //webSocket.sendBIN(msg.c_str(), msg.length() + 1);
    //lastQuery = msg;
    USE_SERIAL.print("Send message to server: "); USE_SERIAL.println(msg);
}

void webSocketEvent(WStype_t type, uint8_t * payload, size_t length) {
    USE_SERIAL.print("EventType:"); USE_SERIAL.println(type);
    
    switch (type) {
        case WStype_DISCONNECTED:
            USE_SERIAL.printf("[WSc] Disconnected!\n");
            statusId = 0;
            break;
        case WStype_CONNECTED: {
                USE_SERIAL.printf("[WSc] Connected to url: %s\n",  payload);
                statusId = 1;    
                String msg = "{ \"action\":\"connect\", \"success\":true, \"deviceId\":\"" + deviceId + "\" }";
                sendMessage(msg);
            }
            break;
        case WStype_TEXT: {

                String text = (char*) payload;
                USE_SERIAL.printf("[WSc] get text: %s\n", payload);

                if (text.indexOf("action:connect")>-1) {

                    USE_SERIAL.println("[WSc] Connected to server successful");
                    delay(1000);
                    
                } else if(text.indexOf("setDestination")>-1) {
                  
                } else if(text.indexOf("pessButton")>-1) {
                    
                } else if(text.indexOf("setDelay")>-1) {
                  
                } else if(text.indexOf("reset")>-1) {
                  
                }

                break;
            }
        case WStype_BIN:
            USE_SERIAL.printf("[WSc] get binary length: %u\n", length);
            hexdump(payload, length);

            // send data to server
            // webSocket.sendBIN(payload, length);
            break;
    }

}

void parseData(String json) {
  StaticJsonBuffer<100> jsonBuffer;
  JsonObject& root = jsonBuffer.parseObject(json);
  if (!root.success()) {
    Serial.println("parseObject() failed");
    return;
  }
  bool succ = root["success"];
  Serial.println("====================================");
  Serial.println(succ);
  Serial.println("====================================");
}

void parseDestData(String json) {
  StaticJsonBuffer<300> jsonBuffer;
  JsonObject& root = jsonBuffer.parseObject(json);
  if (!root.success()) {
    Serial.println("parseObject() failed");
    return;
  }
  bool succ = root["success"];
  Serial.println("====================================");
  Serial.println(succ);
  Serial.println("====================================");
}

void parseDeviceData(String json) {
  Serial.println("Parse Device Data:");
  StaticJsonBuffer<600> jsonBuffer;
  JsonObject& root = jsonBuffer.parseObject(json);
  if (!root.success()) {
    Serial.println("parseObject() failed");
    return;
  }
  bool succ = root["success"];
  int n = root["data"].size();
  Serial.print("Count devices: "); Serial.println(n);
  for(int i=0; i<n; i++) {
    String ip = root["data"][i]["ip"];
    String type = root["data"][i]["type"];
    Serial.print("Type: "); Serial.print(type);
    Serial.print("; Ip: "); Serial.println(ip);
    //devIp[i] = ip; devType[i] = type;    
  }

}

char parseButtData(String json) {
  char cc[2];
  //Serial.println("Parse Button Data:");
  StaticJsonBuffer<100> jsonBuffer;
  JsonObject& root = jsonBuffer.parseObject(json);
  if (!root.success()) {
    Serial.println("parseObject() failed");
    return 0;
  }
  String butt = root["butt"];
  butt.toCharArray(cc, 2);
  return cc[0];
}

/*
float ttRead(OneWire ds) {
  float temperature;
  ds.reset();  // сброс шины
  ds.write(0xCC, POWER_MODE); // пропуск ROM
  ds.write(0x44, POWER_MODE); // инициализация измерения
  delay(900);  // пауза 0,9 сек
  ds.reset();  // сброс шины
  ds.write(0xCC, POWER_MODE); // пропуск ROM  
  ds.write(0xBE, POWER_MODE); // команда чтения памяти датчика  
  ds.read_bytes(bufData, 9);  // чтение памяти датчика, 9 байтов

  if ( OneWire::crc8(bufData, 8) == bufData[8] ) {  // проверка CRC
    // данные правильные
    temperature=  (float)((int)bufData[0] | (((int)bufData[1]) << 8)) * 0.0625 + 0.03125; 
  
    //return String(temperature)+" C";    
  } else {  
    // ошибка CRC, отображается ----
    USE_SERIAL.println("Error");         
  }    

  return temperature;
}
*/
