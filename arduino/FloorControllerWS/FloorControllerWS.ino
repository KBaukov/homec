#include <Arduino.h>
#include <U8g2lib.h>

#include <WebSocketsClient.h>
#include <ArduinoJson.h>
#include <rBase64.h>

#include <math.h>

#ifdef U8X8_HAVE_HW_SPI
#include <SPI.h>
#endif

#define LEFT_BUTT D1
#define RIGT_BUTT  D2
#define MODE_BUTT  D3
#define CTRL_PIN   D3
#define LIGHT_PIN  D4


//---------------
byte NTCPin = A0;
#define SERIESRESISTOR 9100
#define NOMINAL_RESISTANCE 10000
#define NOMINAL_TEMPERATURE 25
#define BCOEFFICIENT 3950

#ifdef ESP8266
extern "C" {
  #include "user_interface.h"
}
#endif

WebSocketsClient webSocket;
rBase64generic<250> encoder;

U8G2_PCD8544_84X48_F_4W_SW_SPI u8g2(U8G2_R0, D8, D7, D5, D6, D0);  // Nokia 5110 Display

// Const
const char* wlan_ssid             = "WF";
const char* wlan_password         = "k0k0JambA";
//const char* wlan_ssid           = "Home";
//const char* wlan_password       = "4r3e2w1q";
const char* ws_host               = "baukoff.net";
const int   ws_port               = 443;
//const char* ws_host             = "alabino.ddns.net";
//const int   ws_port             = 443;
const char* stompUrl              = "/ws"; // don't forget the leading "/" !!!

String deviceId = wifi_station_get_hostname();
String headersString = "Origin: HomeControlApp\r\nDeviceId: " + deviceId;
const char* headers =  (const char *) headersString.c_str();

float tFloor = 0;
float tAir = 0;
float hAir = 0;
float destTf = 25.0;
String currentStage = "0_0";

long wait = 994; // 81.0 - 1 sec
float k = 99.4;
//long wait = 6000;
long count = 900;
long pressDuration = 100;
int statusId = 0;
float period = 10.0;

bool isSessionStart = false;
bool isLbPress = false;
bool isRbPress = false;
bool isMbPress = false;

int lcount = 0;
int lwait = 2700;

//###############################################################################
//### FUNCTIONS ###

void setup() {
  Serial.begin(115200);
  Serial.println();//Serial.println();Serial.println();

  u8g2.begin();

  //pinMode(LEFT_BUTT,  INPUT);
  //pinMode(RIGT_BUTT, INPUT);
  //pinMode(MODE_BUTT,  INPUT);

  resetPin(LEFT_BUTT);
  resetPin(RIGT_BUTT);
  
  pinMode(CTRL_PIN,   OUTPUT);
  digitalWrite(CTRL_PIN, LOW);
  pinMode(LIGHT_PIN,  OUTPUT);
  digitalWrite(LIGHT_PIN,LOW);

  // connect to WiFi
  Serial.print("Logging into WiFi: ");
  Serial.print(wlan_ssid);
  Serial.print("..");

  renderConnect(0,12, "Find WiFi...");

  WiFi.mode(WIFI_STA);
  WiFi.begin(wlan_ssid, wlan_password);
  while (WiFi.status() != WL_CONNECTED) {
    delay(500);
    Serial.print(".");
  }
  Serial.println(" success.");
  String ip = IpAddress2String(WiFi.localIP());
  Serial.print("IP: "+ip); //Serial.println(WiFi.localIP());
  
  renderConnect(0,12, ip);

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
  renderConnect(0,12, "Find server...");
}

void loop() {
  webSocket.loop();

  checkButtPressKey();
  //checkButtPressKey(RIGT_BUTT, *isRbPress);
  
  if (statusId == 0 ) {
    Serial.print(".");
    delay(100);
  }

  if (statusId == 1 && count >= wait ) {
    //Serial.println("======================================================================================");
    tFloor = readNTC(); 
    tAir = random(100, 800) / 10.0; 
    hAir = random(100, 800) / 10.0;
    
    String msg = "{\"action\":\"datasend\", \"type\":\"floordata\", \"data\":{ \"deviceId\":\"" + deviceId + "\", "
                 + "\"tFloor\":" + tFloor + ", "
                 + "\"tAir\":" + tAir + ", "
                 + "\"hAir\":" + hAir + ", "
                 + "\"stage\": \"" + currentStage + "\""
                 + " } }";

    sendMessage(msg);
    count = 0;

    if( tFloor >= destTf) {
      digitalWrite(CTRL_PIN, LOW);
    } else {
      digitalWrite(CTRL_PIN, HIGH);
    }

    render();
  } else {
    count++;
  }
  delay(10);
}

String IpAddress2String(const IPAddress& ipAddress)
{
  return String(ipAddress[0]) + String(".") +\
  String(ipAddress[1]) + String(".") +\
  String(ipAddress[2]) + String(".") +\
  String(ipAddress[3])  ; 
}

float readNTC() {
  float ADCvalue;
  float Resistance;
  ADCvalue = analogRead(NTCPin);
  Serial.print("Analoge "); Serial.print(ADCvalue); Serial.print(" = ");
  //convert value to resistance
  Resistance = (1023 / ADCvalue) - 1;
  Resistance = SERIESRESISTOR / Resistance;
  Serial.print(Resistance); Serial.println(" Ohm");
  float steinhart;
  steinhart = Resistance / NOMINAL_RESISTANCE; // (R/Ro)
  steinhart = log(steinhart); // ln(R/Ro)
  steinhart /= BCOEFFICIENT; // 1/B * ln(R/Ro)
  steinhart += 1.0 / (NOMINAL_TEMPERATURE + 273.15); // + (1/To)
  steinhart = 1.0 / steinhart; // Invert
  steinhart -= 273.15; // convert to C
  return steinhart+NOMINAL_TEMPERATURE;
}

void checkButtPressKey( ) {
  // read the state of the pushbutton value:
  int lState = digitalRead(LEFT_BUTT);
  if (lState == HIGH && !isLbPress ) {
    isLbPress = true; 
  } else if(lState == HIGH && isLbPress) {
    resetPin(LEFT_BUTT);
    destTf = destTf - 1.0;
    render();
    digitalWrite(LIGHT_PIN,HIGH);
    lcount = 2700;
  }

  int rState = digitalRead(RIGT_BUTT);
  if (rState == HIGH && !isRbPress ) {
    isRbPress = true; 
    //digitalWrite(ledPin, LOW);
  } else if(rState == HIGH && isRbPress) {
    resetPin(RIGT_BUTT);
    destTf = destTf + 1.0;
    render();
    digitalWrite(LIGHT_PIN,HIGH);
    lcount = 2700;
  }

  if(lcount>=0) lcount--;
  else {
    digitalWrite(LIGHT_PIN,LOW);
  }
}

void resetPin(int pin) {
  pinMode(pin, OUTPUT);
  delay(500); 
  pinMode(pin, INPUT);
}

void sendMessage(String & msg) {
  webSocket.sendTXT(msg.c_str(), msg.length());
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
        String msg = "{\"action\":\"connect\",\"success\":true,\"sender\":\"" + deviceId + "\"}";
        sendMessage(msg);
        delay(500);
        msg = "{\"action\":\"getLastValues\",\"type\":\"floordata\"}";
        sendMessage(msg);
      }
      break;
    case WStype_TEXT: {
        String text = (char*) payload;
        Serial.printf("[WSc] reciv message: %s\n", payload);

        if (text.indexOf("action:connect") > -1) {
          Serial.println("[WSc] Connected to server successful");
        } else if (text.indexOf("\"action\":\"setLastValues\"") > -1) {
          parseDestData(text);
          String ansver = "{\"action\":\"setLastValues\", \"success\":true}";
          sendMessage(ansver);
         } else if (text.indexOf("sessionStart") > -1) {
          String sender = parseData(text, "sender");
          String hash =   parseData(text, "hash"); 
          String user =   parseData(text, "user");  
          String rMsg = "";       
          Serial.println("incoming hash:"+hash);
          if (isSessionStart) {
            rMsg = "{\"action\":\"sessionStart\",\"success\":false,\"msg\":\"session alredy started:device is busy\",\"hash\":\""+hash+"\"}";
          } else {
            isSessionStart = true;
            rMsg = "{\"action\":\"sessionStart\",\"success\":true,\"hash\":\""+hash+"\"}";            
          }
          String msg = "{\"action\":\"resend\",\"recipient\":\""+sender+"\",\"msg\":\""+b64encode(rMsg)+"\"}";
          sendMessage(msg);
          
        } else if (text.indexOf("sessionStop") > -1) {
          String sender = parseData(text, "sender");
          String hash =   parseData(text, "hash");
          isSessionStart = false;
          String rMsg = "{\"action\":\"sessionStop\",\"success\":true,\"hash\":\""+hash+"\"}";
          String msg = "{\"action\":\"resend\",\"recipient\":\""+sender+"\",\"msg\":\""+b64encode(rMsg)+"\"}";
          sendMessage(msg);
        } else if (text.indexOf("pessButton") > -1) {
          String butt =   parseData(text, "butt");
          String sender = parseData(text, "sender");
          String hash =   parseData(text, "hash");
          String stage =  parseData(text, "stage");
          presKey(butt);
          String rMsg = "{\"success\":true,\"butt\":\""+butt+"\",\"hash\":\""+hash+"\"}";
          String msg = "{\"action\":\"resend\",\"recipient\":\""+sender+"\",\"msg\":\""+b64encode(rMsg)+"\"}";
          sendMessage(msg);
          currentStage = stage;
        } else if (text.indexOf("\"action\":\"setDestValues\"") > -1) {
          parseDestData(text);
          String ansver = "{\"action\":\"setDestValues\", \"success\":true}";
          sendMessage(ansver);          
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
      Serial.printf("[WSc] get binary length: %u\n", length);
      hexdump(payload, length);
      webSocket.sendBIN(payload, length);
      break;
  }

}

void renderConnect(int x, int y, String txt) {  
  const char * st = txt.c_str();  
  u8g2.clearBuffer();
  u8g2.setFont(u8g2_font_cu12_tr);
  u8g2.drawStr(x,y,st);
  u8g2.sendBuffer();
}

void render(void) {
  String ttt = String(tFloor).substring(0,4);
  String hhh = String(destTf).substring(0,4);
  //Serial.println(ttt);
  //Serial.println(hhh);
  
  const char * tt = ttt.c_str();
  const char * hh = hhh.c_str();
  
  u8g2.clearBuffer();
  u8g2.setFont(u8g2_font_ncenB18_tr); 
  u8g2.drawStr(0,18, tt);
  u8g2.drawStr(0,40, hh);
  u8g2.drawStr(52,18,"C");
  u8g2.drawStr(52,40,"C");
  u8g2.setFont(u8g2_font_ncenB08_tr);
  u8g2.drawStr(48,5,"o");
  u8g2.drawStr(48,27,"o");
  //u8g2.setFont(u8g2_font_ncenB08_tr);
  //u8g2.drawStr(0,46,"~~~~~~~~~~~~~~~~~~~~~");
  u8g2.sendBuffer();
}

void parseDestData(String json) {

  StaticJsonBuffer<400> jsonBuffer;
  JsonObject& root = jsonBuffer.parseObject(json);
  if (!root.success()) {
    Serial.println("parseObject() failed");
    Serial.println("#### " + json);
    return;
  }

  Serial.println("====== Dest Values incoming ========"); 
  destTf = atof(root["destTf"]);
  String stage = root["stage"]; 
  currentStage = stage;
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

String b64encode(String s) {
  if(encoder.encode(s)==RBASE64_STATUS_OK) {
    String res = encoder.result();
    return res; 
  }
  return "NaN";  
}
String b64decode(String s) {
  encoder.decode(s);  
  return encoder.result();  
}


void presKey(String butt) {
  char cc[2];
  
  butt.toCharArray(cc, 2);
  //return cc[0]
  int n = (int) cc[0];
  Serial.print("Key="); Serial.println(cc[0]);
  if (n == 76) {
    //Serial.println("Left");
    destTf--;
  }
  if (n == 82) {
    //Serial.println("Right");
    destTf++;
  }
  if (n == 77) {
    //Serial.println("Mode")
  }
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
