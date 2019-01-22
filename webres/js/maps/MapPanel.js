Ext.define('MapPanel', {
    extend: 'Ext.panel.Panel',
    initComponent: function() {
        this.border = true;
        this.frame = true;
        this.region = 'center';
        this.margins = '2 2 2 2';
        this.collapsible = true;
        this.collapsed = false;
        this.resizable = false;
        this.autoScroll = true;
        this.bodyPadding = 10;
        this.bodyStyle = 'padding:10px; background: #ffffff'; //#cbddf3;';

        this.wss = [];

        this.initForm();

        MapPanel.superclass.initComponent.apply(this, arguments);
    },
    initForm: function() {
        this.mapData = this.initConfig().mapData;
        this.title = this.mapData.title;
        this.id = 'map'+this.mapData.id;
        this.tpl = [
            '<div style="text-align:center">','<img id="mapImage_{id}" src="/webres/img/maps/{pict}" />','</div>',
            '<tpl for="sensors">',
                '<div class="{type}" id="{parent.id}_sensor_{device_id}"></div>',
            '</tpl>'
        ];
        this.listeners = { scope: this,
            render: function() {
                this.getSensorsData();
                //Ext.TaskManager.start(this.task);
            },
            resize: function() {
                this.resizeImage();
            }
        };
    },
    resizeImage: function() {
        var h = this.getHeight();
        var w = this.getWidth();
        var winRatio = w/h;
        var imgW = w - 40;
        if( winRatio>=1.3 ) {
            imgW = (h-40)*1.3;
        }
        if(this.data)
            var el = Ext.getDom('mapImage_'+this.mapData.id);    
        if(el) {
            el.width=imgW;
            var imgH = imgW / 1.3;
            var dw = (w - imgW) /2;
            var n = this.data.sensors.length;
            for(var i=0; i<n; i++) {
                var ht =  Ext.getDom(this.mapData.id+'_sensor_'+this.mapData.sensors[i].device_id );
                ht.style.top  = (20 + imgH * this.mapData.sensors[i].yk)+'px';
                ht.style.left = (dw + imgW * this.mapData.sensors[i].xk)+'px';
            }
        }
    },
    setContent: function(data) {
        this.data = data;
        this.update(data);
        this.resizeImage();
        this.setListeners();
    },
    getSensorsData: function() {
        Ext.Ajax.request({
            url: '/api/sensors', scope: this, method: 'POST',
            params: { map_id: this.mapData.id },
            success: function(response, opts) {
              //this.unmask();
              var ansv = Ext.decode(response.responseText);
              if(ansv.success) {    
                this.mapData.sensors = ansv.data;
                this.setContent(this.mapData);
                this.getValues();
              } else error_mes('Ошибка', ansv.msg);
            },
            failure: function() { this.unmask(); }
        });
    },
    getValues: function() {
       // var n = this.data.sensors.length;
        var n = this.mapData.sensors.length;
        for(var i=0; i<n; i++) {
            var devId =  this.mapData.sensors[i].device_id;
            var devName =  devices.getName( devId );
            var type = this.mapData.sensors[i].type;
            if(type=='kotelIcon') {
                Ext.Ajax.request({
                    url: '/api/kotel/getvalues', scope: this, method: 'POST',
                    params: { device_id: devName, id: devId },
                    success: function(response, opts) {
                        var ansv = Ext.decode(response.responseText);
                        var sens = Ext.getDom(this.mapData.id+'_sensor_'+opts.params.id);
                        if(ansv.success) {
                            this.dataRender('koteldata', sens, ansv);
                            sens.style.opacity = 0.2;
                        } else {
                            if(ansv.msg=='нет данных') {
                                sens.style.opacity = 0.2;
                            } else error_mes('Ошибка', ansv.msg);
                        }
                    },
                    failure: function() { this.unmask(); }
                });
            } else {
                Ext.Ajax.request({
                    url: '/api/sensors/data', scope: this, method: 'POST',
                    params: { device_id: devName, id: devId },
                    success: function(response, opts) {
                        var ansv = Ext.decode(response.responseText);
                        var sens = Ext.getDom(this.mapData.id+'_sensor_'+opts.params.id);
                        if(ansv.success) {
                            this.dataRender('roomdata', sens, ansv);
                            sens.style.opacity = 0.2;
                        }  else {
                            if(ansv.msg=='нет данных') {
                                //sens.style.opacity = 0.2;
                            } else error_mes('Ошибка', ansv.msg);
                        }
                    },
                    failure: function() { this.unmask(); }
                });
            }

        }
    },
    setListeners: function() {
        var ss = this.data.sensors;
        for(var i=0; i<ss.length; i++) {
            el = Ext.getDom(this.mapData.id+'_sensor_'+ss[i].device_id);
            if (ss[i].type == 'kotelIcon') {
                el.ondblclick = this.kotelControl
                el.ontouchend  = this.kotelControl
            }
            if(ss[i].type == 'tempIcon') {
                el.ondblclick = this.chartData
                el.ontouchend  = this.chartData
            }
            this.wssConnUp(ss[i].type, el)
        }
    },
    kotelControl: function(ev) {
        var cmp = Ext.getCmp('map'+ev.target.id.split('_')[0]);
        cmp.controllWin  = Ext.create('KotelControlWin', {papa: this});
        cmp.controllWin.openWin();
    },
    chartData: function(ev) {
        var dd = ev.target.id.split('_')
        var cmp = Ext.getCmp('map'+dd[0]);
        cmp.chartWin  = Ext.create('RoomDataChartWin', {devId: dd[2], sensName: devices.getName(dd[2])});
        cmp.chartWin.openWin();
    },
    wssConnUp: function(type, el) {
        var wss = new WebSocket("wss://"+window.location.host+"/ws");

        wss.onopen = function(event) {
            var conn = event.target;
            var sens = WssConnections.getElByConn(conn);
            var dd = sens.id.split('_');
            conn.send('{"action":"connect","assign":"'+devices.getName(dd[2])+'"}');
            console.log("WSS Соединение установлено.");
        };
        wss.onerror = function(error) { console.log("Ошибка " + error.message); };
        wss.onclose = function(event) {
            if (event.wasClean) { console.log('Соединение закрыто чисто'); }
            else { console.log('Обрыв соединения'); }
            console.log('Код: ' + event.code + ' причина: ' + event.reason);
            WssConnections.deleteByConn(event.target);
        };
        wss.onmessage = function(event) {
            console.log("Получены данные: " + event.data);
            var data = Ext.decode(event.data);
            var conn = event.target;
            var sens = WssConnections.getElByConn(conn);
            conn.dataRender(data.type, sens, data);
        };

        wss.dataRender = this.dataRender;
        WssConnections.connections.push( {el:el, conn: wss} );
        el.style.opacity = 0.2;
    },
    dataRender: function(type, sens, data) {
        if(type == 'koteldata') {
            sens.innerHTML = parseFloat(data.data.tp) +'°C</br></br>'
                + parseFloat(data.data.to)+'°C</br></br>';
        } else if( type == 'roomdata' ) {
            sens.innerHTML = parseFloat(data.data.t) +'°C</br></br>'
                +( (data.data.h!='0.00') ? parseFloat(data.data.h)+'%' : '');
        }
        sens.style.opacity = 1;
    }
});