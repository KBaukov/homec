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

        this.assignDevice = '';

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
            },
            resize: function() {
                this.resizeImage();
            }
        };
        this.menu = Ext.create('Ext.menu.Menu', {
            floating: true, scope: this,
            items: [
                { text: 'Reset', handler: this.resetDevice},
                { text: 'Период обновления', handler: this.setDelay}
            ]
        });
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
            //var sens = WssConnections.getElByConn(conn);
            //var dd = sens.id.split('_');
            //conn.send('{"action":"assign","assign":"'+conn.assignDev+'"}');
            console.log("WSS Соединение установлено.");
            Ext.TaskManager.start(conn.task);
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

            if(data.action =='assign' && data.success) {
                Ext.TaskManager.stop(conn.task);
                //sens.style.opacity = 1;
            } else if(data.action == 'unassign') {
                //error_mes('Ошибка', data.msg);
                Ext.TaskManager.start(conn.task)
                sens.style.opacity = 0.2;
            } else if(data.action == 'datasend') {
                conn.dataRender(data.type, sens, data);
            }

        };

        wss.dataRender = this.dataRender;
        wss.assignDev = devices.getName(el.id.split('_')[2]);

        el.style.opacity = 0.2;
        wss.task = { scope: wss,
            run: function() {
                this.send('{"action":"assign","assign":"'+this.assignDev+'"}');
            },
            interval: 500
        };
        el.menu = this.menu;
        el.oncontextmenu = function(e, cmp=this, conn=wss) {

            cmp.menu.init = {el: cmp, conn: conn};

            cmp.menu.showAt(e.pageX, e.pageY, false);

            if(event.preventDefault != undefined)
                event.preventDefault();
            if(event.stopPropagation != undefined)
                event.stopPropagation();

            return null;
        };

        WssConnections.connections.push( {el:el, conn: wss} );
    },
    dataRender: function(type, sens, data) {
        if(type == 'koteldata') {
            sens.innerHTML = parseFloat(data.data.tp).toFixed(2) +'°C</br></br>'
                + parseFloat(data.data.to).toFixed(2)+'°C</br></br>';
        } else if( type == 'roomdata' ) {
            sens.innerHTML = parseFloat(data.data.t).toFixed(2) +'°C</br></br>'
                +( (data.data.h!='0.00') ? parseFloat(data.data.h).toFixed(2)+'%' : '');
        }
        sens.style.opacity = 1;
    },
    resetDevice: function(e,t) {
        var devId =e.parentMenu.init.el.id.split('_')[2];
        var devName = devices.getName(devId);
        e.parentMenu.init.conn.send('{"action":"resend", "recipient":"'+devName+'", "msg":"e3Jlc2V0fQ=="}');
    },
    setDelay: function(e,t) {
        var devId =e.parentMenu.init.el.id.split('_')[2];
        var devName = devices.getName(devId);
        var conn = e.parentMenu.init.conn;
        this.delayWin = Ext.create('KotelDelayWin', {papa: this, devName: devName, wss: conn});
        this.delayWin.openWin();
    }
});