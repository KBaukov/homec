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
                this.wss.send("{\"action\":\"getDestValues\",\"type\":\"koteldata\"}");
                //Ext.TaskManager.start(this.task);

            },
            resize: function() {
                this.resizeImage();
            }
        };
        // this.task = { scope: this,
        //     run: function() { this.getValues() },
        //     interval: 5000
        // };



        Ext.define('KatelData', {
            extend: 'Ext.data.Model',
            fields: [
                {name: 'device_id', type: 'string'},
                {name: 'to',  type: 'float'},
                {name: 'tp',  type: 'float'},
                {name: 'kw',  type: 'int'},
                {name: 'pr',  type: 'float'},
                {name: 'destTo',  type: 'float'},
                {name: 'destTp',  type: 'float'},
                {name: 'destKw',  type: 'int'},
                {name: 'destPr',  type: 'float'},
                {name: 'stage',  type: 'string'}
            ]
        });

        this.store = Ext.create('Ext.data.Store', {
            model: 'KatelData', id: this.id+'_DevStore'
        });

        this.wss = new WebSocket("wss://"+window.location.host+"/ws");

        this.wss.onopen = function() {
            console.log("WSS Соединение установлено.");
        };

        this.wss.onclose = function(event) {
            if (event.wasClean) {
                console.log('Соединение закрыто чисто');
            } else {
                console.log('Обрыв соединения'); // например, "убит" процесс сервера
            }
            console.log('Код: ' + event.code + ' причина: ' + event.reason);
        };

        this.wss.onmessage = function(event) {
            var data = event.data;
            console.log("Получены данные: " + data);

            // var sens = Ext.getDom(this.mapData.id+'_sensor_'+opts.params.id);
            // if(ansv.success) {
            //     sens.style.opacity = 1;
            //     //sens.style.opacity = (ansv.status ==1) ? 1 : 0.3;
            //     sens.innerHTML = parseFloat(ansv.data.tp) +'°C</br></br>'
            //         + parseFloat(ansv.data.to)+'°C</br></br>'
            //     //+ parseFloat(ansv.data.kw)+'kw'
            //     ;
            // } else {
            //     if(ansv.msg=='нет данных') {
            //         sens.style.opacity = 0.2;
            //     } else error_mes('Ошибка', ansv.msg);
            // }
        };

        this.wss.onerror = function(error) {
            console.log("Ошибка " + error.message);
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
              } else error_mes('Ошибка', ansv.msg);
            },
            failure: function() { this.unmask(); }
        });
    },
    getValues: function() {
        var n = this.data.sensors.length;
        for(var i=0; i<n; i++) {
            var devId =  this.data.sensors[i].device_id;
            var devName =  devices.getName( devId );
            var type = this.data.sensors[i].type;
            if(type=='kotelIcon') {
                Ext.Ajax.request({
                    url: '/api/kotel/getvalues', scope: this, method: 'POST',
                    params: { device_id: devName, id: devId },
                    success: function(response, opts) {
                        var ansv = Ext.decode(response.responseText);
                        var sens = Ext.getDom(this.mapData.id+'_sensor_'+opts.params.id);
                        if(ansv.success) {
                            sens.style.opacity = 1;
                            //sens.style.opacity = (ansv.status ==1) ? 1 : 0.3;
                            sens.innerHTML = parseFloat(ansv.data.tp) +'°C</br></br>'
                                + parseFloat(ansv.data.to)+'°C</br></br>'
                                //+ parseFloat(ansv.data.kw)+'kw'
                            ;
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
                            sens.style.opacity = 1;
                            //sens.style.opacity = (ansv.status ==1) ? 1 : 0.3;
                            sens.innerHTML = parseFloat(ansv.data.t) +'°C</br></br>'
                                +( (ansv.h!='0.00') ? parseFloat(ansv.data.h)+'%' : '');

                        }  else {
                            if(ansv.msg=='нет данных') {
                                sens.style.opacity = 0.2;
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
        }
    },
    kotelControl: function(ev) {
        var cmp = Ext.getCmp('map'+ev.target.id.split('_')[0]);
        cmp.controllWin  = Ext.create('KotelControlWin');
        cmp.controllWin.openWin();
    },
    chartData: function(ev) {
        var dd = ev.target.id.split('_')
        var cmp = Ext.getCmp('map'+dd[0]);
        cmp.chartWin  = Ext.create('RoomDataChartWin', {devId: dd[2], sensName: devices.getName(dd[2])});
        cmp.chartWin.openWin();
    }
});