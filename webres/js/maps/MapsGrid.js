Ext.define('MapsGrid', {
    extend: 'Ext.grid.Panel',
    initComponent: function() {
        //this.title = '';
        this.border = true;
        this.frame = false;
        this.region = 'center';
        this.height = 400;
        this.loadMask = true;
        this.emptyText  = 'Нет данных';
        this.margins = '2 2 2 2';
        this.id = 'mapsGrid';
        this.count = 0;
        this.lastId = 0;
        this.resizable = true;
        
        this.selectedRec = null;
        
        this.initData();
        this.initColumns();

        MapsGrid.superclass.initComponent.apply(this, arguments);
    },
    initColumns: function() {
        
        this.columns = [            
            {header: 'ID', align: 'left', width: 40, dataIndex: 'id' }, 
            {header: 'Заголовок', align: 'left', width: 120, dataIndex: 'title', editor: new Ext.form.TextField({ allowBlank: false }) },
            {header: 'Изображение', align: 'left', width: 160, dataIndex: 'pict', editor: new Ext.form.TextField({ allowBlank: false })  },
            {header: 'Ширина', align: 'left', width: 50, dataIndex: 'w', editor: new Ext.form.TextField({ allowBlank: false })  },
            {header: 'Высота', align: 'left', width: 50, dataIndex: 'h', editor: new Ext.form.TextField({ allowBlank: false }) },
            {header: 'Описание', align: 'left', width: 500, dataIndex: 'description', flex: 1, editor: new Ext.form.TextField({ allowBlank: false })  },
            {header: '***', align: 'center',
                xtype: 'actioncolumn', width: 40,
                sortable: false, menuDisabled: true,
                items: [{
                    icon: '/webres/extjs/img/cncl_g.gif',
                    tooltip: 'Удалить карту', scope: this,
                    handler: this.delMap
                }]
            }
        ];
        this.cellEditing = Ext.create('Ext.grid.plugin.CellEditing', { clicksToEdit: 2 });
        
        this.plugins = [ this.cellEditing ];
        this.tbar = [
            Ext.create('Ext.Button', {text: 'Добавить карту', scope: this, disabled: false, id: 'addMapButt',
                style: 'background-position: bottom center;', 
                handler: function() { this.addMap(); }
            }),
            '->', '-',
            Ext.create('Ext.Button', {text: 'Сохранить изменения', scope: this, disabled: false, id: 'editMapButt',
                style: 'background-position: bottom center;', 
                handler: function() { this.saveData(); }
            }),            
        ];
        Ext.define('MapData', {
            extend: 'Ext.data.Model',
            fields: [
                {name: 'id',  type: 'int'},
                {name: 'title',  type: 'string'},
                {name: 'pict',  type: 'string'},
                {name: 'w',  type: 'string'},
                {name: 'h', type: 'string'},
                {name: 'description', type: 'string'}
            ],
        });
    },
    initData: function() {
      this.papa = this.initConfig().papa;
      this.store =  Ext.create('Ext.data.JsonStore', {
          storeId: 'mapsData', autoLoad: false,   
            proxy: {
                type: 'ajax',
                url: '/api/maps',
                reader: {
                    type: 'json',
                    root: 'data',
                    idProperty: 'id'
                }
            },
          fields: [
            {name: 'id'}, {name: 'title'}, {name: 'pict'}, {name: 'w'},
            {name: 'h'}, {name: 'description'}
          ]//,
      });
      this.on('afterlayout', this.loadData, this, { delay: 1, single: true });
      
      this.listeners = {
        select: {
            scope: this,
            fn: function( gr, record, index, eOpts ) { 
                var data = record.data;
                this.papa.mapSensorGrid.loadData(record.data.id, data);
                 
            }
        }
    };
      
    },
    delMap: function(grid, rowIndex) {
        var id = this.store.getAt(rowIndex).data.id;
        Ext.Msg.show({
            title:'Внимание!',
            msg: 'Вы действительно хотите удалить эту карту ???',
            buttons: Ext.Msg.YESNO, scope: this,
            fn: function(btn){
                if(btn == "yes") {
                    Ext.Ajax.request({
                        url: '/api/maps/delete', scope: this, method: 'POST',
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
    addMap: function() { 
        //this.count++; Ext.data.Model
        var rec =  Ext.create('MapData', {
            id: this.lastId+1,
            title: '',
            pict: '',
            w: '',
            h: '',
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
                var data = {id: modify[i].data.id, title: modify[i].data.title, pict:modify[i].data.pict, w:modify[i].data.w, h: modify[i].data.h, description:modify[i].data.description };
                Ext.Ajax.request({
                    url: '/api/maps/edit', scope: this, method: 'POST',
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
        url: '/api/maps', scope: this, method: 'POST',
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

    }
});

