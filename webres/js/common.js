Ext.QuickTips.init();

CloseWin = function() {
    document.location = '/logout';    
};

Ext.Ajax.on('requestcomplete', function(conn, response, opt) {
    if (response.status == 200) {
        var ansv = Ext.JSON.decode(response.responseText);
        if(!ansv.success && ansv.error ) {
            error_mes('Ошибка', ansv.error );
        }
    } else
        error_mes('Ошибка', "Возникла ошибка во время выполнения запроса к серверу.\nПроверьте подключение к сети." );
});
Ext.Ajax.on('requestexception', function(conn, response, opt) {
    if (response.status == 401) {
        LoginWin.on('authorized', function(){
          LoginWin.removeListener('authorized');
          Ext.Ajax.request(opt);
        }, opt);
        relogin();
    } else
        error_mes( 'Ошибка', "Возникла ошибка во время выполнения запроса к серверу.\nПроверьте подключение к сети." );
});

relogin = function() {
    CloseWin();
};

screenLock = function() {
    clearTimeout(screenLockT);
    if(slCounter) {        
        relogin();
    } else {
        screenLockT=setTimeout(screenLock, 500);
        slCounter = true;
    }
    console.log('--- ' + slCounter);
};

//##############################################################################

reloginRequest = function() {
    LoginWin.items.items[1].getForm().getFields().items[0].enable();
    var data = LoginWin.items.items[1].getForm().getValues();
    LoginWin.items.items[1].getForm().getFields().items[0].disable();
    Ext.Ajax.request({
        url: '/loguot',
        params: data,
        scope: LoginWin,
        callback: function(options, success, response) {
            var ansv = Ext.decode(response.responseText);
            if(ansv.success) {
                LoginWin.fireEvent('authorized');
                LoginWin.hide();
                maskShow(false);
            } else {
                error_mes('Ошибка авторизации', ansv.error);
            }
        }
    });
};

LoginWin = new Ext.Window({
        title: 'Авторизация', resizable: false, closable: false, modal: true, width: 323,
        height: 135, layout: 'border',
        items: [
            new Ext.Panel({
              region: 'west', width: 60, frame: false,  border: false,margins:'3 0 3 3', cmargins:'3 3 3 3',
              bodyStyle: 'background: url(/webres/extjs/img/gpg-icon.png) center no-repeat;'              
            }),
            new Ext.form.FormPanel({
              region: 'center', margins:'3 3 3 3', standardSubmit: true, labelWidth: 50, frame: true, labelPad: 10,
              defaultType: 'textfield',  defaults: {  msgTarget: 'side' },
              keys: [{ key: Ext.EventObject.ENTER,
                fn: function() { LoginWin.buttons[0].handler(); }
              }],
              items: [
                { fieldLabel: 'Логин', width: 200, allowBlank: false, disabled: true, blankText: 'Это поле должно быть заполнено', name: 'user_name',inputType: 'textfield' }, 
                { fieldLabel: 'Пароль', width: 200, id: 'loginFormPass', allowBlank:false, blankText: 'Это поле должно быть заполнено', name: 'user_pass', inputType: 'password',
                 enableKeyEvents: true, listeners: { keydown: function(tf, e, eOpts){ if(e.getKey() == e.ENTER) reloginRequest(); } }
                }
              ]
           })
        ],
        buttons: [{  text: 'Войти', handler: reloginRequest  }]
});

error_mes = function(tit, mes) {
    Ext.Msg.show({
        title:tit,
        msg: mes,
        buttons: Ext.Msg.OK,
        icon: Ext.MessageBox.ERROR,
        fn: null
    });
};

confirmMess = function(mes) {
    var ansv = false;
    Ext.Msg.show({
       title:'Внимание!',
       msg: mes,
       buttons: Ext.Msg.YESNO,
       fn: function(btn){
            if(btn == "yes") {
               ansv = true;
            }
       },
       animEl: 'elId',
       icon: Ext.MessageBox.QUESTION
    });

    return ansv;
};

Ext.grid.feature.Grouping.override({
	/**
	 * Override the default getGroupRows implementation to add the column
	 * title and the rendered group value to the groupHeaderTpl data.
	 */
	getGroupRows: function(g) {
		var me = this,
			view = me.view,
			header = view.getHeaderCt(),
			store = view.store,
			record = g.children[0], // get first record in group
			group = me.callOverridden(arguments),
			grouper = me.getGroupField(),
			column,
			renderer,
			v;
		if (!Ext.isEmpty(grouper) && record) {
			// get column of groupfield
			column = header.down('[dataIndex=' + grouper + ']');
			// render the group value
			renderer = column.renderer;
			if (renderer) {
				v = header.prepareData(record[record.persistenceProperty], store.indexOf(record), record, view, view.ownerCt)[column.id];
			} else {
				v = group.name;
			}
			// apply column title and rendered group value for use in groupHeaderTpl
			Ext.apply(group, {
				header: column.text,
				renderedValue: v
			});
		}
		return group;
	}
});

function getElementPosition(elemId) {
    var elem = document.getElementById(elemId);
	
    var w = elem.offsetWidth;
    var h = elem.offsetHeight;
	
    var l = 0;
    var t = 0;
	
    while (elem)
    {
        l += elem.offsetLeft;
        t += elem.offsetTop;
        elem = elem.offsetParent;
    }

    return {"left":l, "top":t, "width": w, "height":h};
}

WssConnections = {
    connections: [],
    getConnByEl: function(el) {
        var n = this.connections.length;
        for( var i=0; i<n; i++) {
            if( this.connections[i].el == el )
                return this.connections[i].conn;
        }
    },
    getElByConn: function(conn) {
        var n = this.connections.length;
        for( var i=0; i<n; i++) {
            if( this.connections[i].conn == conn )
                return this.connections[i].el;
        }
    },
    deleteByConn: function(conn) {
        var n = this.connections.length;
        for( var i=0; i<n; i++) {
            if( this.connections[i].conn == conn )
                delete this.connections[i];
        }
    }
};

//

AlertBox = function(){
    var msgCt;
    var msgDuration = 2000;

    function createBox(t, s){
        return '<div class="msg"><h3>' + t + '</h3><p>' + s + '</p></div>';
    }
    return {
        msg : function(type, title, format){
            if(!msgCt){
                msgCt = Ext.DomHelper.insertFirst(document.body, {id:'msg-div'}, true);
            }
            var s = Ext.String.format.apply(String, Array.prototype.slice.call(arguments, 2));
            var m = Ext.DomHelper.append(msgCt, createBox(title, s), true);
            var color = (type=='info') ? '#0bca50' : ( (type=='error') ? '#d60d0d' : ( (type=='warn') ? '#f3c000' : '#555555' ) );
            m.setStyle('color', color);
            m.hide();
            m.slideIn('t').ghost("t", { delay: msgDuration, remove: true});
        },
        setDuration: function(d) {
            msgDuration = d;
        },
        init : function(){
            if(!msgCt){
                // It's better to create the msg-div here in order to avoid re-layouts
                // later that could interfere with the HtmlEditor and reset its iFrame.
                msgCt = Ext.DomHelper.insertFirst(document.body, {id:'msg-div'}, true);
            }
        }
    };
}();
