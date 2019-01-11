#include <Arduino.h>
#include <Hash.h>
#include <ESP8266WiFi.h>
#include <WebSocketsClient.h>
#include <ArduinoJson.h>
#include <OneWire.h>
#include <SD.h>

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
File hcFile;
WebSocketsClient webSocket;

// Const
const char* wlan_ssid             = "WF";
const char* wlan_password         = "k0k0JambA";
//const char* wlan_ssid           = "Home";
//const char* wlan_password       = "q1w2e3r4";
const char* ws_host               = "192.168.43.175";
const int   ws_port               = 8085;
//const char* ws_host             = "alabino.ddns.net";
//const int   ws_port             = 443;
const char* stompUrl              = "/ws"; // don't forget the leading "/" !!!
const char* host = "KotelController";

String deviceId = wifi_station_get_hostname();
String headersString = "Origin: HomeControlApp\r\nDeviceId: " + deviceId;
const char* headers =  (const char *) headersString.c_str();

// VARIABLES
byte bufData[9];  // буфер данных
int led_pin = 2;

float tp = 0;
float to = 0;
float pr = 0;
int kw = 0;
float destTp = 0.0;
float destTo = 0.0;
float destTc = 0.0;
float destPr = 0.0;
int destKw = 0;
String currentStage = "0_0";

String ctrlComm = "";
//long wait = 994;
long wait = 600000;
long count = 0;
long pressDuration = 100;
int statusId = 0;

bool isSessionStart = false;


//###############################################################################
//### FUNCTIONS ###

void setup() {
  Serial.begin(115200);
  Serial.println();//Serial.println();Serial.println();

  if (!SD.begin(4)) {
    Serial.println("initialization failed!");
    return;
  } else {
    parseDestData(
      readKotelData()
    );
  }

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

  // connect to WiFi
  Serial.print("Logging into WLAN: ");
  Serial.print(wlan_ssid);
  Serial.print(" ...");

  WiFi.mode(WIFI_STA);
  WiFi.begin(wlan_ssid, wlan_password);
  while (WiFi.status() != WL_CONNECTED) {
    delay(500);
    Serial.print(".");
  }
  Serial.println(" success.");
  Serial.print("IP: "); Serial.println(WiFi.localIP());

  // connect to websocket
  webSocket.beginSSL(ws_host, ws_port, stompUrl);
  webSocket.setExtraHeaders(headers);
  webSocket.onEvent(webSocketEvent);

  delay(1000);
}

void loop() {

  webSocket.loop();



  //Serial.print("statusId=");
  //Serial.print(statusId);
  //Serial.print(";  count=");
  //Serial.println(count);

  if (statusId == 1 && count >= wait ) {
    //Serial.println("==========================================");
    tp = 36.79; //ttRead(ds1);//+2;
    to = 27.23; //ttRead(ds2);
    kw = 11; //
    pr = 2.34; //
    Serial.print("tp="); Serial.println(tp);
    Serial.print("to="); Serial.println(to);

    String msg = "{\"action\":\"datasend\", \"type\":\"koteldata\", \"data\":{ \"deviceId\":\"" + deviceId + "\", "
                 + "\"to\":" + to + ", "
                 + "\"tp\":" + tp + ", "
                 + "\"kw\":" + kw + ", "
                 + "\"pr\":" + pr
                 + " } }";

    sendMessage(msg);

    count = 0;
  } else {
    count++;
  }



  if (tp < destTc) {
    digitalWrite(EXTC_BUTT, LOW);
    //Serial.println("Relay is on");
  } else {
    digitalWrite(EXTC_BUTT, HIGH);
    //Serial.println("Relay is off");
  }

  //delay(10);
}

void sendMessage(String & msg) {
  webSocket.sendTXT(msg.c_str(), msg.length());
  //lastQuery = msg;
  Serial.print("[Wsc] Send message to server: "); Serial.println(msg);
}

void webSocketEvent(WStype_t type, uint8_t * payload, size_t length) {
  //Serial.print("EventType:"); Serial.println(type);
  switch (type) {
    case WStype_DISCONNECTED:
      Serial.printf("[WSc] Disconnected!\n");
      statusId = 0;
      break;
    case WStype_CONNECTED: {
        Serial.printf("[WSc] Connected to url: %s\n",  payload);
        statusId = 1;
        String msg = "{\"action\":\"connect\",\"success\":true,\"deviceId\":\"" + deviceId + "\"}";
        sendMessage(msg);
        delay(500);
        msg = "{\"action\":\"getDestValues\",\"type\":\"koteldata\"}";
        sendMessage(msg);
      }
      break;
    case WStype_TEXT: {
        String text = (char*) payload;
        Serial.printf("[WSc] reciv message: %s\n", payload);

        if (text.indexOf("action:connect") > -1) {

          Serial.println("[WSc] Connected to server successful");
          //delay(1000);
        } else if (text.indexOf("\"action\":\"setDestValues\"") > -1) {
          parseDestData(text);
          String ansver = "{\"action\":\"setDestValues\", \"success\":true}";
          sendMessage(ansver);
         } else if (text.indexOf("sessoinStart") > -1) {
          if (isSessionStart) {
            String msg = "{\"action\":\"sessionStart\",\"success\":false,\"msg\":\"кто-то уже управляет котлом\"}";
            sendMessage(msg);
          } else {
            isSessionStart = true;
            String msg = "{\"action\":\"sessionStart\",\"success\":true,\"msg\":\"\"}";
            sendMessage(msg);
          }
        } else if (text.indexOf("sessoinStop") > -1) {
          isSessionStart = false;
          String msg = "{\"action\":\"sessionStop\",\"success\":true,\"msg\":\"\"}";
          sendMessage(msg);
        } else if (text.indexOf("pessButton") > -1) {
          char butt = parseButtData(text);
          presKey(butt);

        } else if (text.indexOf("setDelay") > -1) {

        } else if (text.indexOf("reset") > -1) {
          resetDevice();
        }

        break;
      }
    case WStype_BIN:
      Serial.printf("[WSc] get binary length: %u\n", length);
      hexdump(payload, length);

      // send data to server
      // webSocket.sendBIN(payload, length);
      break;
  }

}



String readKotelData() {
  String content = ""; int i = 0;
  hcFile = SD.open("kotel.dat");
  int sz = hcFile.size();
  if (hcFile) {
    byte b[sz];
    Serial.println("kotel.dat:");
    while (hcFile.available()) {
      b[i++] =  hcFile.read();
    }
    content = String((char*) b).substring(0, sz);
  } else {
    Serial.println("Error while reading kotel.data file");
  }
  hcFile.close();
  Serial.print("Size: ");
  Serial.println(i);
  Serial.println("Content: " + content);
  return content;
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

  StaticJsonBuffer<500> jsonBuffer;
  JsonObject& root = jsonBuffer.parseObject(json);
  if (!root.success()) {
    Serial.println("parseObject() failed");
    Serial.println("#### " + json);
    return;
  }

  Serial.println("====== Dest Values incoming ========");
  //{"action":"setDestValues", "destTo":"0.0",  "destTp":"0.0",  "destTc":"25.0",  "destPr":"1.8",  "destKw":"2",  "stage":"0_0", }
  //{"action":"setDestValues", "destTo":"0.0",  "destTp":"0.0",  "destTc":"25.0",  "destPr":"1.8",  "destKw":"2",  "stage":"0_0", }
  String dTp = root["destTp"]; destTp = atof(root["destTp"]);
  String dTo = root["destTo"]; destTo = atof(root["destTo"]);
  String dTc = root["destTc"]; destTc = atof(root["destTc"]);
  String dPr = root["destPr"]; destPr = atof(root["destPr"]);
  String dKw = root["destKw"]; destKw = atoi(root["destKw"]);
  String stage = root["stage"]; currentStage = stage;

  if (SD.exists("kotel.dat")) {
    //Serial.println("kotel.dat exists.");
    SD.remove("kotel.dat");
    hcFile = SD.open("kotel.dat", FILE_WRITE);
    if (hcFile) {
      String cont = "{destTo:" + dTo + ",destTp:" + dTp + ",destKw:" + dKw + ",destPr:" + dPr + ",destTc:" + dTc + ",stage:\"" + stage + "\"}";
      Serial.print("Writing to kotel.dat: " + cont);
      hcFile.println(cont);
      // close the file:
      hcFile.close();
      Serial.println("done.");
    } else {
      // if the file didn't open, print an error:
      Serial.println("error opening kotel.dat");
    }
  } else {
    Serial.println("kotel.dat doesn't exist.");
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
    temperature =  (float)((int)bufData[0] | (((int)bufData[1]) << 8)) * 0.0625 + 0.03125;

    //return String(temperature)+" C";
  } else {
    // ошибка CRC, отображается ----
    Serial.println("Error while get temp.");
  }

  return temperature;
}

void presKey(char k) {
  int n = (int) k;
  Serial.print("Key="); Serial.println(k);
  if (n == 76) {
    //Serial.println("Left");
    digitalWrite(LEFT_BUTT, LOW);
    delay(pressDuration);
    digitalWrite(LEFT_BUTT, HIGH);
  }
  if (n == 82) {
    //Serial.println("Right");
    digitalWrite(RIGT_BUTT, LOW);
    delay(pressDuration);
    digitalWrite(RIGT_BUTT, HIGH);
  }
  if (n == 77) {
    //Serial.println("Mode");
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