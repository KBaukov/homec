Ext.define('MapTemplate', {
    extend: 'Ext.panel.Panel',
    initComponent: function() {
        this.border = true;
        this.frame = true;
        this.region = 'south';
        this.height = 700;
        this.id = 'tMapPanel';
        this.margins = '0 0 0 0';
        this.collapsible = true;
        this.collapsed = true;
        this.resizable = false;
        this.autoScroll = true;
        this.bodyPadding = 10;
        this.bodyStyle = 'padding:10px; background: #ffffff'; //#cbddf3;';
        this.data = null;
        
        this.initForm();

        MapTemplate.superclass.initComponent.apply(this, arguments);
    },
    initForm: function() {
        this.html = '';               
        this.tpl = [
            '<div style="text-align:center">','<img id="tmapImage_{id}" src="/webres/img/maps/{pict}" />','</div>',
            '<tpl for="sensors">',
                '<div class="{type}" id="tht{id}"></div>',
            '</tpl>'
        ];
        
        this.menu = Ext.create('Ext.menu.Menu', {
            floating: true, scope: this,
            items: [
                { text: 'Создать сенсор', handler: function() {
                        var cmp = Ext.getCmp('tMapPanel');
                        var cmpPos = cmp.getXY();
                        cmp.addSensorAt( { x: cmp.newSensorPosition.x - cmpPos[0]-4 , y: cmp.newSensorPosition.y - cmpPos[1]-32 });
                        //alert('x=' + (cmp.newSensorPosition.x - cmpPos[0]-4)+'; y='+( cmp.newSensorPosition.y - cmpPos[1]-32) );
                    } 
                } 
            ]
        });
        
        this.listeners = { scope: this,
            render: function() {
                this.body.on('contextmenu', function(e) {
                    var cmp = Ext.getCmp('tMapPanel');
                    cmp.menu.showAt(e.getXY());
                    cmp.newSensorPosition = e.getPoint();
                     e.stopEvent();
                    return null;
                });
                this.resizeImage();
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
            var el = Ext.getDom('tmapImage_'+this.data.id);
        if(el) {
            el.width=imgW;
            var imgH = imgW / 1.3;
            var dw = (w - imgW) /2;
            var n = this.data.sensors.length;
            for(var i=0; i<n; i++) {
                var ht =  Ext.getDom('tht'+this.data.sensors[i].id );
                ht.style.top  = (20 + imgH * this.data.sensors[i].yk)+'px';
                ht.style.left = (dw + imgW * this.data.sensors[i].xk)+'px';
            }
        }
    },
    setContent: function(data) {
        this.data = data;
        this.update(data);
        this.resizeImage();
    }, 
    addSensorAt: function(pos) {
        var position = {};  
        if(this.data)
            var el = Ext.getDom('tmapImage_'+this.data.id);
         if(el) {
            var w = this.getWidth(); 
            var h = this.getHeight(); 
            var imgW = el.width;
            var imgH = imgW / 1.3;
            var dw = (w - imgW) / 2;
            position.x = (pos.x - dw) / imgW;
            position.y = (pos.y - 20) / imgH;
        }
        
        this.papa.mapSensorGrid.addSensor(position);
    }
});