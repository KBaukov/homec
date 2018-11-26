Ext.define('RoomDataChartWin', {
    extend: 'Ext.Window',
    initComponent: function () {

        this.resizable=false;
        this.closable=true;
        this.width=900;
        this.height=600;
        this.plain=true;
        this.modal= true;
        this.headerPosition='top';
        this.layout='fit';

        this.shadow='drop';
        this.shadowOffset=8;
        this.buttons = [{ text: 'Закрыть', handler: this.closeWin, scope: this }];
        this.tbar = [
            {xtype: 'label', text: '', width: 20},
            {xtype: 'label', text: 'показаны последние '},
            {xtype: 'field', value: 50, id: 'limitData', width: 30 }, '->', '-',
            { text: 'Обновить', handler: this.reloadData, scope: this }
        ];

        this.initForm();
        this.items=[ this.Chart];
        this.data = [];

        RoomDataChartWin.superclass.initComponent.apply(this, arguments);
    },
    initForm: function() {
        var iData = this.initConfig();
        this.devId = devices.getName( iData.devId );
        this.title='Статистика датчика ' + iData.sensName;

        this.store = Ext.create('Ext.data.JsonStore', {
            fields: ['t', 'h', 'date'],
            data: [
                {t: 57.6, h: 33.5, date: new Date(2018, 11, 23, 15, 20) },
                {t: 86.6, h: 53.4, date: new Date(2018, 11, 23, 15, 21) },
                {t: 45.6, h: 43.3, date: new Date(2018, 11, 23, 15, 22)},
                {t: 27.6, h: 23.2, date: new Date(2018, 11, 23, 15, 23) },
                {t: 11.6, h: 63.5, date: new Date(2018, 11, 23, 15, 24) },
                {t: 66.6, h: 53.4, date: new Date(2018, 11, 23, 15, 25) },
                {t: 28.6, h: 33.3, date: new Date(2018, 11, 23, 15, 26) }
            ]
        });

        this.Chart = Ext.create('Ext.chart.Chart', {
            //xtype: 'chart',
            style: 'background:#fff',
            store: this.store,
            itemId: 'chartCmp', id: 'hcChart',
            axes: [{
                type: 'Numeric',
                minimum: 0,
                maximum: 100,
                position: 'left',
                fields: ['t', 'h'],
                title: 'Температура °C, Влажность %',
                grid: {
                    odd: {
                        fill: '#f9f9f9',
                        stroke: '#eee',
                        'stroke-width': 0.7
                    }
                }
            }, {
                type: 'Time',
                position: 'bottom',
                fields: 'date',
                title: 'Время',
                dateFormat: 'H:i:s',
                groupBy: 'year,month,day',
                aggregateOp: 'sum',
                constrain: true
            }],
            series: [{
                type: 'line',
                axis: ['left', 'bottom'],
                xField: 'date',
                yField: 't',
                label: {
                    //display: 'none',
                    field: 't',
                    renderer: function(v) { return v >> 0; },
                    'text-anchor': 'middle'
                },
                markerConfig: {
                    radius: 2,
                    size: 2
                }
            },{
                type: 'line',
                axis: ['left', 'bottom'],
                xField: 'date',
                yField: 'h',
                label: {
                    display: 'none',
                    field: 'h',
                    renderer: function(v) { return v >> 0; },
                    'text-anchor': 'middle'
                },
                markerConfig: {
                    radius: 2,
                    size: 2
                }
            }]
        });
    },
    closeWin: function(ev) {
        this.close();
    },
    openWin: function() {
        Ext.Ajax.request({
            url: '/api/sensors/stat', scope: this, method: 'POST',
            params: {device_id: this.devId, count: 50},
            success: function(response, opts) {
                var ansv = Ext.decode(response.responseText);
                if(ansv.success) {
                    //Ext.TaskManager.start(this.kotelControlPanel.valuesTask);
                    this.pushData(ansv.data);
                    this.show();

                } else { error_mes('Ошибка', ansv.msg); }
            },
            failure: function() { }
        });
    },
    pushData: function(data) {
        var n = data.length;
        var chData = [];
        var rangeFrom, rangeTo;
        for(var i=0; i<n; i++) {
            var t = data[i].t;
            var h = data[i].h;
            var ddd = data[i].date.substring(0, 19);
            var dd = ddd.split('T')[0].split('-');
            var tt = ddd.split('T')[1].split(':');
            var date = new Date(dd[0], dd[1], dd[2], tt[0], tt[1], tt[2]);
            if(i==0) rangeTo = date;
            if(i==n-1) rangeFrom = date;
            chData.push({date: date, t: t, h: h});
        }
        this.Chart.axes.get(1).fromDate = rangeFrom;
        this.Chart.axes.get(1).toDate = rangeTo;
        this.Chart.markerIndex = 1;
        this.store.loadData(chData);
    },
    reloadData: function() {
        var c = Ext.getCmp('limitData').value;
        Ext.Ajax.request({
            url: '/api/sensors/stat', scope: this, method: 'POST',
            params: {device_id: this.devId, count: c},
            success: function(response, opts) {
                var ansv = Ext.decode(response.responseText);
                if(ansv.success) {
                    this.pushData(ansv.data);
                    //this.show();

                } else { error_mes('Ошибка', ansv.msg); }
            },
            failure: function() { }
        });
    }
})