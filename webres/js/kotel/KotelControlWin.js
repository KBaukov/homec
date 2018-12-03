Ext.define('KotelControlWin', {
    extend: 'Ext.Window',
    initComponent: function () {
        this.title='Управление отоплением';
        this.resizable=false;
        this.closable=true;
        this.width=435;
        this.height=575;
        this.plain=true;
        this.modal= true;
        this.headerPosition='top';
        this.layout='border';

        this.shadow='drop';
        this.shadowOffset=8;
        this.buttons = [{ text: 'Закрыть', handler: this.closeWin, scope: this }];

        this.initForm();
        this.items=[ this.tabPanel ];

        KotelControlWin.superclass.initComponent.apply(this, arguments);
    },
    initForm: function() {
        this.kotelControlPanel   = Ext.create('KotelControlPanel',  {papa: this} );
        this.ketelChartPanel = Ext.create('Ext.panel.Panel',  {
            title: 'Статистика', disabled: true
        });
        this.tabPanel = Ext.create('Ext.tab.Panel',{
            region: 'center',
            items: [
                this.kotelControlPanel,
                this.ketelChartPanel
            ]
        });


        //this.kotelTabPanel   = Ext.create('KotelTabPanel',  {papa: this} );

    },
    closeWin: function(ev) {
        Ext.Ajax.request({
            url: '/api/kotel/sessionstop', scope: this, method: 'POST',
            //params: {user: user.login},
            success: function(response, opts) {
                var ansv = Ext.decode(response.responseText);

                if(ansv.success) {

                    Ext.TaskManager.stop(this.kotelControlPanel.valuesTask);

                    this.close();

                } else { error_mes('Ошибка', ansv.msg); }
            },
            failure: function() { }
        });

    },
    openWin: function() {
        Ext.Ajax.request({
            url: '/api/kotel/sessionstart', scope: this, method: 'POST',
            params: {user: user.login},
            success: function(response, opts) {
                var ansv = Ext.decode(response.responseText);

                if(ansv.success) {
                    Ext.TaskManager.start(this.kotelControlPanel.valuesTask);
                    this.show();

                } else { error_mes('Ошибка', ansv.msg); }
            },
            failure: function() { }
        });
    }
})