Ext.define('MapSensorsGrid', {
    extend: 'Ext.grid.Panel',
    initComponent: function() {
        //this.title = '';
        this.border = true;
        this.frame = false;
        this.region = 'east';
        this.width = 1100;
        this.loadMask = true;
        this.emptyText  = 'Нет данных';
        this.margins = '2 2 2 2';
        this.id = 'sensorsGrid';
        this.count = 0;
        this.lastId = 0;
        this.mapId = 0;
        
        this.isInit = false;
        
        this.selectedRec = null;
        
        this.initData();
        this.initColumns();

        MapSensorsGrid.superclass.initComponent.apply(this, arguments);
    },
    initColumns: function() {
	  this.typeStore =  [
	      ['tempIcon','Темп. Сенсор'], ['pirIcon','Сенсор Движения'],
          ['pressIcon','Сенсор Давления'], ['kotelIcon','Контроллер Котла'],
          ['floorIcon','Контролер теплого пола']
      ];
        
        this.columns = [            
            {header: 'ID', align: 'left', width: 40, dataIndex: 'id' }, 
            {header: 'Карта', align: 'left', width: 120, dataIndex: 'map_id', editor: new Ext.form.TextField({ allowBlank: false }), renderer: this.mapNameRenderer },
            {header: 'Устройство', align: 'left', width: 120, dataIndex: 'device_id', editor: new Ext.form.field.ComboBox({
                    typeAhead: true, displayField: 'name', valueField: 'id',
                    triggerAction: 'all',
                    store: devices
                }), renderer: this.devNameRenderer
            },
            {header: 'Тип', align: 'left', width: 160, dataIndex: 'type', editor: new Ext.form.field.ComboBox({
                    typeAhead: true, displayField: 'name', valueField: 'id',
                    triggerAction: 'all',
                    store:this.typeStore
                }), renderer: this.typeRenderer
            },
            {header: 'Изображение', align: 'left', width: 160, dataIndex: 'pict', editor: new Ext.form.TextField({ allowBlank: false })  },
            {header: 'Позиция X', align: 'left', width: 160, dataIndex: 'xk', editor: new Ext.form.TextField({ allowBlank: false })  },
            {header: 'Позиция Y', align: 'left', width: 160, dataIndex: 'yk', editor: new Ext.form.TextField({ allowBlank: false })  },
            {header: 'Описание', align: 'left', width: 500, dataIndex: 'description', flex: 1, editor: new Ext.form.TextField({ allowBlank: false })  },
            {header: '***', align: 'center',
                xtype: 'actioncolumn', width: 40,
                sortable: false, menuDisabled: true,
                items: [{
                    icon: '/webres/extjs/img/cncl_g.gif',
                    tooltip: 'Удалить сенсор', scope: this,
                    handler: this.delSensor
                }]
            }
        ];
        this.cellEditing = Ext.create('Ext.grid.plugin.CellEditing', { clicksToEdit: 2 });
        
        this.plugins = [ this.cellEditing ];
        this.tbar = [
            Ext.create('Ext.Button', {text: 'Добавить сенсор', scope: this, disabled: false, id: 'addSensorButt',
                style: 'background-position: bottom center;', 
                handler: function() { this.getLastId(); }
            }),
            '->', '-',
            Ext.create('Ext.Button', {text: 'Сохранить изменения', scope: this, disabled: false, id: 'editSensorButt',
                style: 'background-position: bottom center;', 
                handler: function() { this.saveData(); }
            }),            
        ];
        Ext.define('SensorData', {
            extend: 'Ext.data.Model',
            fields: [
                {name: 'id',  type: 'int'},
                {name: 'map_id',  type: 'int'},
                {name: 'device_id',  type: 'int'},
                {name: 'type',  type: 'string'},
                {name: 'pict',  type: 'string'},
                {name: 'xk',  type: 'float'},
                {name: 'yk', type: 'float'},
                {name: 'description', type: 'string'}
            ],
        });
    },
    initData: function() {
      this.getLastId();
      this.papa = this.initConfig().papa;
      this.store =  Ext.create('Ext.data.JsonStore', {
          storeId: 'sensorsData', autoLoad: false,   
            proxy: {
                type: 'ajax',
                url: '/api/sensors',
                reader: {
                    type: 'json',
                    root: 'data',
                    idProperty: 'id'
                }
            },
          fields: [
            {name: 'id'}, {name: 'map_id'}, {name: 'device_id'}, {name: 'type'}, {name: 'pict'}, 
            {name: 'xk'}, {name: 'yk'}, {name: 'description'}
          ]//,
      });
      
      this.mapStore = this.papa.mapsGrid.store;
	  //this.devStore = devices; //this.papa.papa.devices.store;
      
    },
    delSensor: function(grid, rowIndex) {
        var id = this.store.getAt(rowIndex).data.id;
        Ext.Msg.show({
            title:'Внимание!',
            msg: 'Вы действительно хотите удалить этот сенсор ???',
            buttons: Ext.Msg.YESNO, scope: this,
            fn: function(btn){
                if(btn == "yes") {
                    Ext.Ajax.request({
                        url: '/api/sensors/delete', scope: this, method: 'POST',
                        params: {id: id},
                        success: function(response, opts) {
                          this.unmask();
                          var ansv = Ext.decode(response.responseText);
                          if(ansv.success) {  
                            this.getStore().removeAt(rowIndex);           
                            this.count = this.store.count();
                          } else error_mes('Ошибка', ansv.msg);  
                        },
                        failure: function() { this.unmask(); }
                    });
                    this.getStore().removeAt(rowIndex);
                }
            },
            animEl: 'elId',
            icon: Ext.MessageBox.QUESTION
        });
    },
    addSensor: function(pos) { 
        var rec =  Ext.create('SensorData', {
            id: this.lastId+1,
            map_id: this.mapId,
            device_id: '',
            type: '',
            pict: '',
            xk: (pos) ? pos.x : 0,
            yk: (pos) ? pos.y : 0,
            description: ''
        });
        rec.setDirty();        
        this.store.insert(this.count, rec);
        this.cellEditing.startEditByPosition({ row: this.count+1, column: 0 });
    },
    saveData: function() { 
        var modify = this.store.getModifiedRecords();
        var n = modify.length;
        if(n>0) {
            for(var i=0; i<n; i++) {
                Ext.Ajax.request({
                    url: '/api/sensors/edit', scope: this, method: 'POST',
                    params: {id: modify[i].data.id, map_id: modify[i].data.map_id, device_id: modify[i].data.device_id, type:modify[i].data.type, pict:modify[i].data.pict, xk:modify[i].data.xk, yk: modify[i].data.yk, description:modify[i].data.description },
                    success: function(response, opts) {
                      this.unmask();
                      var ansv = Ext.decode(response.responseText);
                      if(ansv.success) {  
                        this.loadData(this.mapId);            
                        //this.count = this.store.count();
                      } else error_mes('Ошибка', ansv.msg);  
                    },
                    failure: function() { this.unmask(); }
                });
            }
        } else {
            error_mes('Ошибка', 'Нет изменений');
        }
    },
    loadData: function(mapId, upData) { 
      this.mapId = mapId;
      this.mask();
      Ext.Ajax.request({
        url: '/api/sensors', scope: this, method: 'POST',
        params: { map_id: this.mapId },
        success: function(response, opts) {
          this.unmask();
          var ansv = Ext.decode(response.responseText);
          if(ansv.success) {  
            this.store.loadData(ansv.data);
            this.count = this.store.count();
            if(upData) {
                upData.sensors = ansv.data;
                this.papa.mapTemplate.setContent(upData);
            }
          } else error_mes('Ошибка', ansv.msg);  
        },
        failure: function() { this.unmask(); }
      });

    },
    getLastId: function() {
        Ext.Ajax.request({
        url: '/api/sensors/lastid', scope: this, method: 'POST',
        success: function(response, opts) {
          var ansv = Ext.decode(response.responseText);
          if(ansv.success) {  
            this.lastId = ansv.data;
            if(this.isInit) this.addSensor();
            else this.isInit = true;
          } else error_mes('Ошибка', ansv.msg);  
        },
        failure: function() { this.unmask(); }
      });
    },
    devNameRenderer: function(val) { 
        for(var i=0; i<devices.getCount(); i++ ) {
            if( devices.getAt(i).data.id==val )
                return devices.getAt(i).data.name;
        }
        return '';
    },
	typeRenderer: function(val) { 
		for(var i=0; i<this.typeStore.length; i++) {
			if(this.typeStore[i][0]==val)
				return this.typeStore[i][1];	
		}
	},
	mapNameRenderer: function(val) {
		for(var i=0; i<this.mapStore.getCount(); i++ ) {
	            if( this.mapStore.getAt(i).data.id==val )
	                return this.mapStore.getAt(i).data.title;
        	}
       	 return '';
	}
});
