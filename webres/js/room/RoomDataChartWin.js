Ext.define('RoomDataChartWin', {
    extend: 'Ext.Window',
    initComponent: function () {
        this.title='Статистика';
        this.resizable=false;
        this.closable=true;
        this.width=900;
        this.height=600;
        this.plain=true;
        this.modal= true;
        this.headerPosition='top';
        this.layout='border';

        this.shadow='drop';
        this.shadowOffset=8;
        this.buttons = [{ text: 'Закрыть', handler: this.closeWin, scope: this }];

        this.initForm();
        this.items=[this.Chart];
        this.data = [];

        RoomDataChartWin.superclass.initComponent.apply(this, arguments);
    },
    initForm: function() {
        this.devId = devices.getName( this.initConfig().devId );

        this.store = Ext.create('Ext.data.JsonStore', {
            fields: ['t', 'h', 'date'],
            data: [
                {t: 27.6, h: 1, date: '2018-11-23T15:20:00' },
                {t: 26.6, h: 2, date: '2018-11-23T15:21:00' },
                {t: 25.6, h: 3, date: '2018-11-23T15:22:00' },
                {t: 27.6, h: 4, date: '2018-11-23T15:23:00' },
                {t: 21.6, h: 5, date: '2018-11-23T15:24:00' },
                {t: 37.6, h: 6, date: '2018-11-23T15:25:00' },
                {t: 28.6, h: 7, date: '2018-11-23T15:26:00' }
            ]
        });

        this.Chart = Ext.create('Ext.chart.Chart', {
            store: this.store, id: 'hcChart',
            //width: 400,
            //height: 300,
            axes: [
                {
                    title: 'Температура',
                    type: 'Numeric',
                    position: 'left',
                    fields: ['t'],
                    minimum: 0,
                    maximum: 100
                },
                {
                    title: 'Время',
                    type: 'Date',
                    position: 'bottom',
                    fields: ['h'],
                    dateFormat: 'm-d H:i'
                }
            ]
        });
    },
    closeWin: function(ev) {
        // Ext.Ajax.request({
        //     url: '/api/kotel/sessionstop', scope: this, method: 'POST',
        //     //params: {user: user.login},
        //     success: function(response, opts) {
        //         var ansv = Ext.decode(response.responseText);
        //
        //         if(ansv.success) {
        //
        //             Ext.TaskManager.stop(this.kotelControlPanel.valuesTask);

                    this.close();

        //         } else { error_mes('Ошибка', ansv.msg); }
        //     },
        //     failure: function() { }
        // });

    },
    openWin: function() {
        Ext.Ajax.request({
            url: '/api/sensors/stat', scope: this, method: 'POST',
            params: {device_id: this.devId},
            success: function(response, opts) {
                var ansv = Ext.decode(response.responseText);
                if(ansv.success) {
                    //Ext.TaskManager.start(this.kotelControlPanel.valuesTask);
                    this.data = ansv.data;
                    this.show();

                } else { error_mes('Ошибка', ansv.msg); }
            },
            failure: function() { }
        });
    }
})