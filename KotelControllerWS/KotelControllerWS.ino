#include <Arduino.h>
#include <Hash.h>
#include <ESP8266WiFi.h>
#include <WebSocketsClient.h>
#include <ArduinoJson.h>
#include <DHT.h>
#include <OneWire.h>
#include <rBase64.h>

#define USE_SERIAL Serial
#define POWER_MODE  0
#define LEFT_BUTT  D5
#define RIGT_BUTT  D6
#define MODE_BUTT  D7
#define EXTC_BUTT  D1
#define RSTPIN D3


#ifdef ESP8266
extern "C" {
  #include "user_interface.h"
}
#endif

// SETTINGS

OneWire  ds1(D4);  // on pin D4 (a 4.7K resistor is necessary)
OneWire  ds2(D3);  // on pin D3 (a 4.7K resistor is necessary)

const char* wlan_ssid             = "WF";
const char* wlan_password         = "k0k0JambA";
//const char* wlan_ssid             = "Home";
//const char* wlan_password         = "q1w2e3r4";

const char* ws_host               = "192.168.0.111";
const int   ws_port               = 443;
//const char* ws_host               = "alabino.ddns.net";
//const int   ws_port               = 8083;
const char* stompUrl              = "/ws"; // don't forget the leading "/" !!!

String deviceId = wifi_station_get_hostname();
String headersString = "Origin: HomeControlApp\r\nDeviceId: "+deviceId;
const char* headers =  (const char *) headersString.c_str();

// VARIABLES

WebSocketsClient webSocket;

byte bufData[9];  // буфер данных

float tp = 0;
float to = 0;
float pr = 0;
int kw = 0;

float destTp = 25.0;
float destTo = 60.0;
float destTc = 70.0;
float destPr = 2.0;
int destKw = 11;
String currentStage = "0_0";

String ctrlComm = "";
//long wait = 994;
long wait = 600000;
long count = 0;
long pressDuration = 100;
int statusId = 0;



// FUNCTIONS

void sendMessage(String & msg) {
    
    webSocket.sendTXT(msg.c_str(), msg.length());
    //lastQuery = msg;
    USE_SERIAL.print("[Wsc] Send message to server: "); USE_SERIAL.println(msg);
}

void webSocketEvent(WStype_t type, uint8_t * payload, size_t length) {
    //USE_SERIAL.print("EventType:"); USE_SERIAL.println(type);
    
    switch (type) {
        case WStype_DISCONNECTED:
            USE_SERIAL.printf("[WSc] Disconnected!\n");
            statusId = 0;
            break;
        case WStype_CONNECTED: {
                USE_SERIAL.printf("[WSc] Connected to url: %s\n",  payload);
                statusId = 1;    
                String msg = "{\"action\":\"connect\",\"success\":true,\"deviceId\":\"" + deviceId + "\"}";
                sendMessage(msg);
            }
            break;
        case WStype_TEXT: {
                String text = (char*) payload;
                USE_SERIAL.printf("[WSc] reciv message: %s\n", payload);

                if (text.indexOf("action:connect")>-1) {

                    USE_SERIAL.println("[WSc] Connected to server successful");
                    //delay(1000);
                    
                } else if(text.indexOf("\"action\":\"setDestValues\"")>-1) {
                    parseDestData(text);
                    String ansver = "{\"action\":\"setDestValues\", \"success\":true}";
                    sendMessage(ansver);
                } else if(text.indexOf("pessButton")>-1) {
                    char butt = parseButtData(text);
                    presKey(butt);
                    String ansver = "{\"action\":\"pessButton\", \"success\":true}";
                    sendMessage(ansver);
                    
                } else if(text.indexOf("action\":\"update")>-1) {
                    parseUpdateData(text);
                  
                } else if(text.indexOf("reset")>-1) {
                  resetDevice();
                } else if(text.indexOf("setStage")>-1) {
                    parseStageData(text);
                    
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

void setup() {

    pinMode(LEFT_BUTT, OUTPUT);
    pinMode(RIGT_BUTT, OUTPUT);
    pinMode(MODE_BUTT, OUTPUT);
    pinMode(EXTC_BUTT, OUTPUT);
    pinMode(RSTPIN, OUTPUT);
  
    digitalWrite(LEFT_BUTT, HIGH);
    digitalWrite(RIGT_BUTT, HIGH);
    digitalWrite(MODE_BUTT, HIGH);
    digitalWrite(EXTC_BUTT, HIGH);
    digitalWrite(RSTPIN, HIGH);
  
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

  webSocket.loop();

  tp = 50.34; //ttRead(ds1);//+2;
  to = 65.35; //ttRead(ds2);
  kw = 11; //
  pr = 2.43; //

  //USE_SERIAL.print("tp="); USE_SERIAL.println(tp);
  //USE_SERIAL.print("to="); USE_SERIAL.println(to);

  //USE_SERIAL.print("statusId=");
  //USE_SERIAL.print(statusId);
  //USE_SERIAL.print(";  count=");
  //USE_SERIAL.println(count);

  if(statusId == 1 && count >= wait ) {
    //USE_SERIAL.println("==========================================");
    
  
    String msg = "{\"action\":\"datasend\", \"type\":\"koteldata\", \"data\":{ \"deviceId\":\""+deviceId+"\", "
      + "\"to\":" + to +", "
      + "\"tp\":" + tp +", "
      + "\"kw\":" + kw +", "
      + "\"pr\":" + pr //+", "
      //+ "\"stage\":\"" + currentStage +"\""
    +" } }";
    
    sendMessage(msg);

    count = 0;
  } else {
      count++;
  }
  
   

   if(tp<destTc) {
      digitalWrite(EXTC_BUTT, LOW);
      //USE_SERIAL.println("Relay is on");
    } else {
      digitalWrite(EXTC_BUTT, HIGH);
      //USE_SERIAL.println("Relay is off");
    }

   //delay(10);
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

void parseUpdateData(String json) {
  StaticJsonBuffer<512> jsonBuffer;
  JsonObject& root = jsonBuffer.parseObject(json);
  if (!root.success()) {
    Serial.println("parseObject() failed");
    return;
  }
  Serial.println("====== Update file incoming ========");
  String file = root["fname"];
  String tmp = root["cfile"];
  rbase64.decode(tmp);
  String content = rbase64.result();
  Serial.println("!!!: "+content);
  
}

void parseDestData(String json) {
  StaticJsonBuffer<300> jsonBuffer;
  JsonObject& root = jsonBuffer.parseObject(json);
  if (!root.success()) {
    Serial.println("parseObject() failed");
    return;
  }
  //bool succ = root["success"];
  Serial.println("====== Dest Values incoming ========");
  //Serial.println(succ);
  //Serial.println("====================================");
  //if(succ) {
    destTp = atof(root["destTp"]);
    destTo = atof(root["destTo"]);
    destTc = atof(root["destTc"]);
    destPr = atof(root["destPr"]);
    destKw = atoi(root["destKw"]);
    String stage= root["stage"];
    currentStage = stage;
  //}
}

void parseStageData(String json) {
  StaticJsonBuffer<300> jsonBuffer;
  JsonObject& root = jsonBuffer.parseObject(json);
  if (!root.success()) {
    Serial.println("parseObject() failed");
    return;
  }
  //bool succ = root["success"];
  Serial.println("====== Stage changed ========");
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
  String stage = root["satge"];
  currentStage = stage;
  butt.toCharArray(cc, 2);
  return cc[0];
}

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
    USE_SERIAL.println("Error while get temp.");         
  }    

  return temperature;
}

void presKey(char k) {
  int n = (int) k;
  USE_SERIAL.print("Key="); USE_SERIAL.println(k);
  if(n==76) {
    //USE_SERIAL.println("Left");
    digitalWrite(LEFT_BUTT, LOW);
    delay(pressDuration);
    digitalWrite(LEFT_BUTT, HIGH);
  }
  if(n==82) {
    //USE_SERIAL.println("Right");
    digitalWrite(RIGT_BUTT, LOW);
    delay(pressDuration);
    digitalWrite(RIGT_BUTT, HIGH);
  }
  if(n==77) {
    //USE_SERIAL.println("Mode");
    digitalWrite(MODE_BUTT, LOW);
    delay(pressDuration);
    digitalWrite(MODE_BUTT, HIGH);
  }
}

void resetDevice() {
  Serial.println("Reset...");
  digitalWrite(RSTPIN, LOW);
  delay(1000);
}
