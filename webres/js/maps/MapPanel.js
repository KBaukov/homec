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
                //Ext.TaskManager.start(this.task);

            },
            resize: function() {
                this.resizeImage();
            }
        };
        this.task = { scope: this,
            run: function() { this.getValues() },
            interval: 5000
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
            var devId = this.data.sensors[i].device_id;
            Ext.Ajax.request({
                url: '/api/rooms/getvalues', scope: this, method: 'POST',
                params: { device_id: devId },
                success: function(response, opts) {
                  var ansv = Ext.decode(response.responseText);
                  if(ansv.success) {    
                    var sens = Ext.getDom(this.mapData.id+'_sensor_'+ansv.device_id);
                        sens.style.opacity = (ansv.status ==1) ? 1 : 0.3;
                        sens.innerHTML = parseFloat(ansv.t) +'°C</br></br>'
                        +( (ansv.h!='0.00') ? parseFloat(ansv.h)+'%' : '');
                    
                  } else error_mes('Ошибка', ansv.msg);  
                },
                failure: function() { this.unmask(); }
            });
        }
    },
    setListeners: function() {
        var ss = this.data.sensors;
        for(var i=0; i<ss.length; i++) {
            if (ss[i].type == 'kotelIcon') {
                el = Ext.getDom(this.mapData.id+'_sensor_'+ss[i].device_id);
                el.ondblclick = this.kotelControl
                el.ontouchend  = this.kotelControl
            }
        }
    },
    kotelControl: function(ev) {
        var cmp = Ext.getCmp('map'+ev.target.id.split('_')[0]);
        cmp.controllWin  = Ext.create('KotelControlWin');
        cmp.controllWin.show();
    }
});