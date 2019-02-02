#include <ESP8266WiFi.h>
#include <WebSocketsClient.h>
#include <ArduinoJson.h>
#include <DHT.h>
#include <Wire.h>  // Include Wire if you're using I2C

#define PIN_RESET 255  //
#define DHTPIN D4
#define USE_SERIAL Serial

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

float temp = 0;
float hum  = 0;

long wait = 994; // 81.0 - 1 sec
float k = 99.4;
float period = 10.0;
int statusId = 0;
long count = 900;


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
    Serial.print(".");
  }
  Serial.println(" success.");
  Serial.print("IP: "); Serial.println(WiFi.localIP());

  // connect to websocket
  webSocket.beginSSL(ws_host, ws_port, stompUrl);
  webSocket.setExtraHeaders(headers);
  webSocket.onEvent(webSocketEvent);

  Serial.println("++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++");
  webSocket.loop();
  delay(500);
  webSocket.loop();
  delay(1000);

  Serial.print("Connected to server: ");    
}

void loop() {

  webSocket.loop();
  
  if (statusId == 0 ) {
    Serial.print(".");
    delay(100);
  } else if (statusId == 1 && count >= wait ) {
    temp = dht.readTemperature(); hum = dht.readHumidity();
    //temp = random(100, 800) / 10.0; hum = random(100, 800) / 10.0;
    if (isnan(hum) || isnan(temp)) {    
      Serial.println("DHT Error!");
    } else {
      Serial.print("Temp="); Serial.print(temp);
      Serial.print("C; Humidity="); Serial.print(hum); 
      Serial.print("%; statusId="); Serial.print(statusId);
      Serial.print("; count="); Serial.println(count);
    }
   
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
  count++;
  delay(10);
}

// FUNCTIONS
void sendMessage(String & msg) {    
    webSocket.sendTXT(msg.c_str(), msg.length());
    USE_SERIAL.print("Send message to server: "); USE_SERIAL.println(msg);
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
                String msg = "{ \"action\":\"connect\", \"success\":true, \"deviceId\":\"" + deviceId + "\" }";
                sendMessage(msg);
            }
            break;
        case WStype_TEXT: {

          String text = (char*) payload;
          Serial.printf("[WSc] reciv message: %s\n", payload);
  
          if (text.indexOf("action:connect")>-1) {
              USE_SERIAL.println("[WSc] Connected to server successful");
              delay(100);
              
          } else if (text.indexOf("setDelay") > -1) {
            String d = parseData(text, "delay");
            Serial.println("Change period to: "+d);            
            setPeriod(d);                
          } else if (text.indexOf("reset") > -1) {
            resetDevice();
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

String parseData(String json, String key) {
  StaticJsonBuffer<500> jsonBuffer;
  JsonObject& root = jsonBuffer.parseObject(json);
  if (!root.success()) {
    Serial.println("parseObject() failed");
    return "";
  }
  return root[key];
}

void setPeriod(String val) {
  unsigned char* buf = new unsigned char[10];
  val.getBytes(buf, 10, 0);  
  const char *val2 = (const char*)buf;
  period = atof(val2);
  wait = round(k*period);
}

void resetDevice() {
  Serial.println("Reset..."); 
  const char *rr;
  Serial.println(atof(rr));
}
