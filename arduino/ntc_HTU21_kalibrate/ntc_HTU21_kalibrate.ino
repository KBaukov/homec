#include <Arduino.h>
#include <U8g2lib.h>
#include <math.h>

#ifdef U8X8_HAVE_HW_SPI
#include <SPI.h>
#include <Wire.h>
#include <HTU21D.h>
#endif

#define SDA  D4
#define SCL  D3


//------------------------------
byte NTCPin = A0;
#define SERIESRESISTOR 9100
#define NOMINAL_RESISTANCE 10000
#define NOMINAL_TEMPERATURE 25
#define BCOEFFICIENT 3950

HTU21D myHTU21D(HTU21D_RES_RH12_TEMP14);

U8G2_PCD8544_84X48_F_4W_SW_SPI u8g2(U8G2_R0, D8, D7, RX, D6, D0);  

float tFloor = 0;
float tAir = 0;
float hAir = 0;
float destTf = 25.0;
float Resistance;
float ADCvalue;

//###############################################################################
//### FUNCTIONS ###


float readNTC() {
    
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

void render(void) {
  
  String t = String(tAir);
  //String h  = String(hAir);
  String r = String(Resistance); 
  String a = String(ADCvalue); 
  
  const char * tt = t.c_str(); //const char * dd1 = d1.c_str();
  const char * rr = r.c_str(); //const char * dd3 = d3.c_str();
  const char * aa = a.c_str(); //const char * dd4 = d4.c_str();
  //const char * hh = h.c_str();   //const char * dd2 = d2.c_str();
  
  u8g2.clearBuffer();
  // air temp and humidity 
  u8g2.setFont(u8g2_font_timR10_tr); 
  u8g2.drawStr(0,10, tt);
  //u8g2.drawStr(0,20, hh);
  u8g2.drawStr(0,25, rr);
  u8g2.drawStr(0,40, aa);

  u8g2.sendBuffer();
}


void setup() {
  Serial.begin(115200);
  Serial.println();//Serial.println();Serial.println();

  while (myHTU21D.begin(SDA,SCL) != true)
  {
    Serial.println(F("Si7021 sensor is faild or not connected"));
    delay(5000);
  }
  Serial.println(F("Si7021 sensor is active"));

  tFloor = readNTC(); 
  tAir = myHTU21D.readTemperature(); //random(100, 800) / 10.0; 
  hAir = myHTU21D.readHumidity(); //random(100, 800) / 10.0;

  u8g2.begin();

}

void loop() {
  
  tFloor = readNTC(); 
  tAir = myHTU21D.readTemperature(); //random(100, 800) / 10.0; 
  hAir = myHTU21D.readHumidity(); //random(100, 800) / 10.0;
   
  render();
  delay(1000);
}
