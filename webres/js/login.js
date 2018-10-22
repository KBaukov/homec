Ext.require([
    'Ext.window.Window',
    'Ext.layout.container.Border',
    'Ext.util'
]);

Ext.onReady(function() {
	
    Ext.BLANK_IMAGE_URL = '/webres/extjs/img/s.gif';

    Ext.QuickTips.init();

    var authForm = [
        { xtype: 'textfield', fieldLabel: 'Логин', name: 'username' },
        { fieldLabel: 'Пароль', name: 'password', inputType: 'password', enableKeyEvents: true,
            listeners: { keydown: function(tf, e, eOpts){ if(e.getKey() == e.ENTER) loginHandler(); } }
        }
    ];

    var auth = Ext.create('Ext.form.Panel',  {
        region: 'center', margins:'3 3 3 3', standardSubmit: true,
        frame: true, defaultType: 'textfield', height: 100,
        defaults: {
            labelWidth: 50, labelPad: 10, allowBlank: false,
            blankText: 'Это поле должно быть заполнено', msgTarget: 'side', width: 185
        }, 
        items: authForm
    });

    var logo = Ext.create('Ext.Panel', {
        region: 'west', width: 60, margins:'3 0 3 3', frame: false, border: false,
        bodyStyle: 'background: url(/webres/extjs/img/gpg-icon.png) center no-repeat;'		
    });

    var loginHandler = function() {
        if (auth.getForm().isValid())
            auth.getForm().submit();
    };

    var win = Ext.create('Ext.Window', {
        title: 'Авторизация',
            resizable: false, closable: false,
            width: 280, height: 127,
        plain: true,
        headerPosition: 'top',
        layout: 'border',
		items: [logo, auth],
        shadow: 'drop', shadowOffset: 8,
            buttons: [{
                    text: 'Войти',
                    handler: loginHandler
            }]
    });
	
    //if (error != '') {		
    if (document.URL.lastIndexOf('auth=false')>-1) {		
        Ext.Msg.show({
            title:'Ошибка авторизации',
            msg: 'В доступе отказано',
            buttons: Ext.Msg.OK,
            icon: Ext.MessageBox.ERROR,
            fn: function() {
                    win.show();
            }
        });		
    } else {
        win.show();
    }
	
});