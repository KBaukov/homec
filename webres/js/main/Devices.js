Ext.define('Devices', {
    extend: 'Ext.grid.Panel',
    initComponent: function() {
        this.title = 'Устройства';
        this.border = true;
        this.frame = false;
        this.region = 'center';
        this.height = 400;
        this.loadMask = true;
        this.emptyText  = 'Нет данных';
        this.margins = '2 2 2 2';
        this.id = 'devicesGrid';
        this.count = 0;
        this.lastId = 0;
        this.collapsible = true;
        this.collapsed = false;
        this.resizable = true;
        
        this.selectedRec = null;
        
        this.initData();
        this.initColumns();

        Devices.superclass.initComponent.apply(this, arguments);
    },
    initColumns: function() {
        
        this.columns = [            
            {header: 'ID', align: 'left', width: 40, dataIndex: 'id' }, 
            {header: 'Тип устройства', align: 'left', width: 120, dataIndex: 'type', editor: new Ext.form.field.ComboBox({
                    typeAhead: true,
                    triggerAction: 'all',
                    store: [ ['TempSensor','TempSensor'], ['TempHumSensor','TempHumSensor'], ['KotelController','KotelController'], ['FloorController','FloorController'] ]
                }) 
            },
            {header: 'Имя', align: 'left', width: 160, dataIndex: 'name', editor: new Ext.form.TextField({ allowBlank: false })  },
            {header: 'IP адрес', align: 'left', width: 160, dataIndex: 'ip', editor: new Ext.form.TextField({ allowBlank: false })  },
            {header: 'Статус', align: 'left', width: 110, dataIndex: 'active_flag', renderer: this.statusRenderer, editor: new Ext.form.field.ComboBox({
                    typeAhead: true,
                    triggerAction: 'all',
                    store: [ ['Y','активно'], ['N','не активно'] ]
                })
            },
            {header: 'Описание', align: 'left', width: 500, dataIndex: 'description', flex: 1, editor: new Ext.form.TextField({ allowBlank: false })  },
            {header: '***', align: 'center',
                xtype: 'actioncolumn', width: 40,
                sortable: false, menuDisabled: true,
                items: [{
                    icon: '/webres/extjs/img/cncl_g.gif',
                    tooltip: 'Удалить устройство', scope: this,
                    handler: this.delDevice
                }]
            }
        ];
        this.cellEditing = Ext.create('Ext.grid.plugin.CellEditing', { clicksToEdit: 2 });
        
        this.plugins = [ this.cellEditing ];
        this.tbar = [
            Ext.create('Ext.Button', {text: 'Добавить устройство', scope: this, disabled: false, id: 'addDeviceButt',
                style: 'background-position: bottom center;', 
                handler: function() { this.addDevice(); }
            }),
            '->', '-',
            Ext.create('Ext.Button', {text: 'Сохранить изменения', scope: this, disabled: false, id: 'editDeviceButt',
                style: 'background-position: bottom center;', 
                handler: function() { this.saveData(); }
            }),            
        ];
        Ext.define('Device', {
            extend: 'Ext.data.Model',
            fields: [
                {name: 'id',  type: 'int'},
                {name: 'type',  type: 'string'},
                {name: 'name',  type: 'string'},
                {name: 'ip',  type: 'string'},
                {name: 'active_flag', type: 'string'},
                {name: 'description', type: 'string'}
            ],
        });
    },
    initData: function() {
      this.papa = this.initConfig().papa;
      this.store =  Ext.create('Ext.data.JsonStore', {
          storeId: 'devicesData', autoLoad: false,  id: 'deviceStore', 
            proxy: {
                type: 'ajax',
                url: '/api/devices',
                reader: {
                    type: 'json',
                    root: 'data',
                    idProperty: 'id'
                }
            },
          fields: [
            {name: 'id'}, {name: 'type'}, {name: 'name'}, {name: 'ip'},
            {name: 'active_flag'}, {name: 'description'}
          ]//,
      });
      this.on('afterlayout', this.loadData, this, {
        delay: 1, single: true
      });
      
    },
    delDevice: function(grid, rowIndex) {
        var id = this.store.getAt(rowIndex).data.id;
        Ext.Msg.show({
            title:'Внимание!',
            msg: 'Вы действительно хотите удалить это устройство ???',
            buttons: Ext.Msg.YESNO, scope: this,
            fn: function(btn){
                if(btn == "yes") {
                    Ext.Ajax.request({
                        url: '/api/device/delete', scope: this, method: 'POST',
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
    addDevice: function() { 
        //this.count++; Ext.data.Model
        var rec =  Ext.create('Device', {
            id: this.lastId+1,
            type: '',
            name: '',
            ip: '',
            active_flag: 'N',
            description: ''
        });
        rec.setDirty();        
        this.store.insert(this.count, rec);
        this.cellEditing.startEditByPosition({
            row: this.count+1, 
            column: 0
        });
    },
    saveData: function() { 
        var modify = this.store.getModifiedRecords();
        var n = modify.length;
        if(n>0) {
            for(var i=0; i<n; i++) {
                var data = {
					id: modify[i].data.id, type: modify[i].data.type, 
					name:modify[i].data.name, ip:modify[i].data.ip, 
					active_flag: modify[i].data.active_flag, 
					description:modify[i].data.description 
				};
                Ext.Ajax.request({
                    url: '/api/device/edit', scope: this, method: 'POST',
                    params: data,
                    success: function(response, opts) {
                      this.unmask();
                      var ansv = Ext.decode(response.responseText);
                      if(ansv.success) {  
                        this.store.reload();            
                        this.count = this.store.count();
                      } else error_mes('Ошибка', ansv.msg);  
                    },
                    failure: function() { this.unmask(); }
                });
            }
        } else {
            error_mes('Ошибка', 'Нет изменений');
        }
    },
    loadData: function(subsId, start, cnt) { 
      this.mask();
      Ext.Ajax.request({
        url: '/api/devices', scope: this, method: 'GET',
        success: function(response, opts) {
          this.unmask();
          var ansv = Ext.decode(response.responseText);
          if(ansv.success) {  
            this.store.loadData(ansv.data);            
            this.count = this.store.count();
            this.lastId = this.store.data.items[this.count-1].data.id;
          } else error_mes('Ошибка', ansv.msg);  
        },
        failure: function() { this.unmask(); }
      });

    },
    statusRenderer: function(val) { 
        if(val=='Y') return 'активно'; 
        else return 'не активно'; 
    }
});

