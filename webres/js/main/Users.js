Ext.define('Users', {
    extend: 'Ext.grid.Panel',
    initComponent: function() {
        this.title = 'Пользователи';
        this.border = true;
        this.frame = false;
        this.region = 'center';
        this.loadMask = true;
        this.emptyText  = 'Нет данных';
        this.margins = '2 2 2 2';
        this.id = 'UsersGrid';
	  this.count = 0;
        this.lastId = 0;
        this.queryParams = {};
        
        this.initData();
        this.initColumns();
        

        Users.superclass.initComponent.apply(this, arguments);
    },
    initColumns: function() {
      this.columns = [
          {header: 'ID', align: 'left', width: 70, dataIndex: 'id' },
          {header: 'Тип пользователя', align: 'left', width: 110, dataIndex: 'user_type', editor: new Ext.form.field.ComboBox({
                    typeAhead: true,
                    triggerAction: 'all',
                    store: [ ['admin','админ'], ['user','пользователь'] ]
                })
	     },
          {header: 'Login', align: 'left', width: 100, dataIndex: 'login', editor: new Ext.form.TextField({ allowBlank: false }) },
          {header: 'Pass', align: 'left', width: 100, dataIndex: 'pass', editor: new Ext.form.TextField({ allowBlank: false }) },
          {header: 'Статус', align: 'left', width: 110, dataIndex: 'active_flag', renderer: this.statusRenderer, editor: new Ext.form.field.ComboBox({
                    typeAhead: true,  triggerAction: 'all',  store: [ ['Y','активен'], ['N','не активен'] ]
                }) 
		},
           {header: 'Дата последнего визита', align: 'left', width: 200, dataIndex: 'last_visit', renderer: Ext.util.Format.dateRenderer('d.m.Y H:i:s') },
           {header: '***', align: 'center',
                xtype: 'actioncolumn', width: 40,
                sortable: false, menuDisabled: true,
                items: [{
                    icon: '/webres/extjs/img/cncl_g.gif',
                    tooltip: 'Удалить пользователя', scope: this,
                    handler: this.delUser
                }]
           }
      ];
	this.cellEditing = Ext.create('Ext.grid.plugin.CellEditing', { clicksToEdit: 2 });
      this.plugins = [ this.cellEditing ];
	
	this.tbar = [
            Ext.create('Ext.Button', {text: 'Добавить пользователя', scope: this, disabled: false, id: 'addUserButt',
                style: 'background-position: bottom center;', 
                handler: function() { this.addUser(); }
            }),
            '->', '-',
            Ext.create('Ext.Button', {text: 'Сохранить изменения', scope: this, disabled: false, id: 'editUserButt',
                style: 'background-position: bottom center;', 
                handler: function() { this.saveData(); }
            }),            
        ];
        Ext.define('User', {
            extend: 'Ext.data.Model',
            fields: [
                {name: 'id',  type: 'int'},
                {name: 'user_type',  type: 'string'},
                {name: 'login',  type: 'string'},
                {name: 'pass',  type: 'string'},
                {name: 'active_flag', type: 'string'},
                {name: 'last_visit', type: 'string'}
            ],
        });
    },  
    initData: function() {
      
        this.papa = this.initConfig().papa;
        this.store =  Ext.create('Ext.data.JsonStore', {
            autoLoad: false,
            proxy: {
                type: 'ajax',
                url: '/api/users',
                reader: {
                    type: 'json',
                    root: 'data',
                    idProperty: 'id'
                }
            },
            fields: [
              {name: 'id', type: 'int'}, 
              {name: 'user_type'},
              {name: 'login'}, 
              {name: 'pass'}, 
              {name: 'active_flag'},
              {name: 'last_visit'}
            ]
        });
        this.listeners = { scope: this,
              afterrender: function() {
                  this.loadData();
              }
        };
 
    },
    loadData: function() {     
      this.mask();
      Ext.Ajax.request({
        url: '/api/users', scope: this, method: 'GET',
        params: { },
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
        if(val=='Y') return 'активный'; 
        else return 'не активный'; 
    },
    dateRenderer: function(val) {
      return Ext.util.Format.date(val, 'd.m.Y H:i:s');
    },
    delUser: function(grid, rowIndex) {
        var id = this.store.getAt(rowIndex).data.id;
        Ext.Msg.show({
            title:'Внимание!',
            msg: 'Вы действительно хотите удалить этого пользователя ???',
            buttons: Ext.Msg.YESNO, scope: this,
            fn: function(btn){
                if(btn == "yes") {
                    Ext.Ajax.request({
                        url: '/api/user/delete', scope: this, method: 'POST',
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
    addUser: function() { 
        //this.count++; Ext.data.Model
        var rec =  Ext.create('User', {
            id: this.lastId+1,
            user_type: '',
            login: '',
            pass: '',
            active_flag: 'N',
            last_visit: ''
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
					id: modify[i].data.id, user_type: modify[i].data.user_type, 
					login:modify[i].data.login, pass:modify[i].data.pass, 
					active_flag: modify[i].data.active_flag, 
					last_visit:modify[i].data.last_visit 
				};
                Ext.Ajax.request({
                    url: '/api/user/edit', scope: this, method: 'POST',
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
    }
});

