#include <Arduino.h>
#include <U8g2lib.h>

#include <WebSocketsClient.h>
#include <ArduinoJson.h>
#include <rBase64.h>

#include <math.h>

#ifdef U8X8_HAVE_HW_SPI
#include <SPI.h>

#endif

#define LEFT_BUTT D3
#define RIGT_BUTT  D4
//#define MODE_BUTT  D3
#define CTRL_PIN   D5
//#define LIGHT_PIN  D4


//------------------------------
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

//HTU21D myHTU21D(HTU21D_RES_RH12_TEMP14);


// Nokia 5110 Display
U8G2_PCD8544_84X48_F_4W_SW_SPI u8g2(U8G2_R0, D8, D7, D5, D6, D0);  

// Const
const char* wlan_ssid             = "WF";
const char* wlan_password         = "k0k0JambA";
//const char* wlan_ssid           = "Home";
//const char* wlan_password       = "4r3e2w1q"
const char* ws_host               = "baukoff.net";
const int   ws_port               = 443;
//const char* ws_host             = "192.168.43.175";
//const int   ws_port             = 8085;
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

bool isOn = false;

//###############################################################################
//### FUNCTIONS ###

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

String IpAddress2String(const IPAddress& ipAddress)
{
  return String(ipAddress[0]) + String(".") +\
  String(ipAddress[1]) + String(".") +\
  String(ipAddress[2]) + String(".") +\
  String(ipAddress[3])  ; 
}

void resetPin(int pin) {
  pinMode(pin, OUTPUT);
  delay(500); 
  pinMode(pin, INPUT);
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

void renderConnect(int x, int y, String txt) {  
  const char * st = txt.c_str();  
  u8g2.clearBuffer();
  u8g2.setFont(u8g2_font_cu12_tr);
  u8g2.drawStr(x,y,st);
  u8g2.sendBuffer();
}

void render(void) {
  
  String t1 = String(tAir).substring(0,2); String d1 = String(tAir).substring(4,3);
  String h  = String(hAir).substring(0,2); String d2 = String(hAir).substring(4,3);
  String t2 = String(tFloor).substring(0,2); String d3 = String(tFloor).substring(4,3);
  String t3 = String(destTf).substring(0,2); String d4 = String(destTf).substring(4,3);
  
  const char * tt1 = t1.c_str(); const char * dd1 = d1.c_str();
  const char * tt2 = t2.c_str(); const char * dd3 = d3.c_str();
  const char * tt3 = t3.c_str(); const char * dd4 = d4.c_str();
  const char * hh = h.c_str();   const char * dd2 = d2.c_str();
  
  u8g2.clearBuffer();
  // air temp and humidity 
  u8g2.setFont(u8g2_font_10x20_tn); 
  u8g2.drawStr(0,13, tt1); u8g2.drawStr(22,13, dd1); 
  u8g2.drawStr(0,30, hh);  u8g2.drawStr(22,30, dd2); 
  u8g2.setFont(u8g2_font_fub11_tr);
  u8g2.drawStr(18,13, "."); u8g2.drawStr(18,30, ".");
  u8g2.setFont(u8g2_font_timR10_tr);
  u8g2.drawStr(35,14,"C");
  u8g2.drawStr(34,30,"%");
  u8g2.setFont(u8g2_font_u8glib_4_tf);
  u8g2.drawStr(33,7,"o");

  // Floor temp
  u8g2.setFont(u8g2_font_t0_17b_tn); 
  u8g2.drawStr(0,48, tt2); u8g2.drawStr(20,48, dd3); 
  u8g2.setFont(u8g2_font_fub11_tr);
  u8g2.drawStr(16,48, ".");
  u8g2.setFont(u8g2_font_timR10_tr);
  u8g2.drawStr(30,48,"C");
  u8g2.setFont(u8g2_font_u8glib_4_tf);
  u8g2.drawStr(29,40,"o");

  // dest Floor temp
  u8g2.setFont(u8g2_font_t0_17b_tn); 
  u8g2.drawStr(55,48, tt3); //u8g2.drawStr(62,48, dd4); 
  //u8g2.setFont(u8g2_font_fub11_tr);
  //u8g2.drawStr(58,48, ".");
  u8g2.setFont(u8g2_font_timR10_tr);
  u8g2.drawStr(74,48,"C");
  u8g2.setFont(u8g2_font_u8glib_4_tf);
  u8g2.drawStr(73,40,"o");

  // Arrow
  u8g2.setFont(u8g2_font_open_iconic_arrow_1x_t);
  u8g2.drawStr(44,47,"B");

  // On/Off icon
  u8g2.setFont(u8g2_font_open_iconic_check_4x_t);
  if (isOn) {
    u8g2.drawStr(51,32,"A");
  } else {
    u8g2.drawStr(51,32,"B");
  }

  // Line
  u8g2.drawLine(0, 33, 84, 33);
  u8g2.drawLine(48, 33, 48, 0);

  u8g2.sendBuffer();
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
    //digitalWrite(LIGHT_PIN,HIGH);
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
    //digitalWrite(LIGHT_PIN,HIGH);
    lcount = 2700;
  }

  if(lcount>=0) lcount--;
  else {
    //digitalWrite(LIGHT_PIN,LOW);
  }
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

  render();
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

void setup() {
  Serial.begin(115200);
  Serial.println();//Serial.println();Serial.println();

/*
  while (myHTU21D.begin(D2,D1) != true)
  {
    Serial.println(F("Si7021 sensor is faild or not connected"));
    delay(5000);
  }
  Serial.println(F("Si7021 sensor is active"));
*/
  u8g2.begin();

  //pinMode(LEFT_BUTT,  INPUT);
  //pinMode(RIGT_BUTT, INPUT);
  //pinMode(MODE_BUTT,  INPUT);

  resetPin(LEFT_BUTT);
  resetPin(RIGT_BUTT);
  
  pinMode(CTRL_PIN,   OUTPUT);
  digitalWrite(CTRL_PIN, LOW);
  //pinMode(LIGHT_PIN,  OUTPUT);
  //digitalWrite(LIGHT_PIN,LOW);

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
  Serial.println("IP: "+ip); //Serial.println(WiFi.localIP());
  
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
                 + "\"tf\":" + tFloor + ", "
                 + "\"ta\":" + tAir + ", "
                 + "\"ha\":" + hAir + ", "
                 + "\"destTf\":" + destTf + ", "
                 + "\"stage\": \"" + currentStage + "\""
                 + " } }";

    sendMessage(msg);
    count = 0;

    if( tFloor >= destTf) {
      isOn = false;
      digitalWrite(CTRL_PIN, LOW);
    } else {
      isOn = true;
      digitalWrite(CTRL_PIN, HIGH);
    }

    render();
  } else {
    count++;
  }
  delay(10);
}
