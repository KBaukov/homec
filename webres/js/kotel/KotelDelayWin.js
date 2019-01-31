Ext.define('KotelDelayWin', {
    extend: 'Ext.Window',
    initComponent: function () {
        this.title='Период обновления данных';
        this.resizable=false;
        this.closable=true;
        this.width=220;
        this.height=100;
        this.plain=true;
        this.modal= true;
        this.headerPosition='top';
        this.layout='border';

        this.shadow='drop';
        this.shadowOffset=8;
        this.buttons = [{ text: 'Сохранить и Закрыть', handler: this.closeWin, scope: this }];

        this.initForm();
        this.items=[ this.form ];

        KotelDelayWin.superclass.initComponent.apply(this, arguments);
    },
    initForm: function() {
        var init = this.initConfig();
        this.papa = init.papa;
        this.devName = init.devName;
        this.wss = init.wss;

        var spiner = Ext.create('Ext.form.field.Spinner',  {
            name: 'delay',
            fieldLabel: 'Период:',
            value: 10,
            step: 1,
            onSpinUp: function() {
                var me = this;
                if (!me.readOnly) {
                    var val = parseInt(me.getValue().split(' '), 10)||0; // gets rid of " Pack", defaults to zero on parse failure
                    me.setValue((val + me.step) );
                }
            },
            onSpinDown: function() {
                var me = this;
                if (!me.readOnly) {
                    var val = parseInt(me.getValue().split(' '), 10)||0; // gets rid of " Pack", defaults to zero on parse failure
                    if (val <= me.step) {
                        me.setValue('Dry!');
                    } else {
                        me.setValue((val - me.step));
                    }
                }
            }
        });

        this.form = Ext.create('Ext.form.Panel',  {
            region: 'center', margins:'3 3 3 3', //standardSubmit: true,
            frame: true, height: 100, //defaultType: 'textfield',
            defaults: {
                labelPad: 10, allowBlank: false,
                blankText: 'Это поле должно быть заполнено', msgTarget: 'side', width: 185
            },
            items: [ spiner ]
        });

    },
    closeWin: function(ev) {
        var values = this.form.getValues();
        var rMsg = '{"action":"setDelay","delay":"'+values.delay+'"}';

        this.wss.send('{"action":"resend", "recipient":"'+this.devName+'", "msg":"'+btoa(rMsg)+'"}');
        this.close();
    },
    openWin: function() {
        this.show();
    }
})