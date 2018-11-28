Ext.define('RoomDataChartWin', {
    extend: 'Ext.Window',
    initComponent: function () {

        this.resizable=false;
        this.closable=true;
        this.width=1000;
        this.height=700;
        this.plain=true;
        this.modal= true;
        this.headerPosition='top';
        this.layout='fit';

        var n = new Date();
        var now = new Date((n-0) + 3600000)
        var exDay = new Date(now - 86400000);

        this.shadow='drop';
        this.shadowOffset=8;
        this.buttons = [{ text: 'Закрыть', handler: this.closeWin, scope: this }];
        this.dateFrom = { xtype: 'datefield', width: 80, value: exDay, id: 'chDateFrom'};
        this.dateTo = { xtype: 'datefield', width: 80, value: now, id: 'chDateTo'};
        this.timeFrom = { xtype: 'timefield', format: 'H:i:s', width: 80, value: exDay, id: 'chTimeFrom'};
        this.timeTo = { xtype: 'timefield', format: 'H:i:s', width: 80, value: now, id: 'chTimeFTo'};

        this.tbar = [
            {xtype: 'label', text: '', width: 20},
            {xtype: 'label', text: 'показаны последние '},
            {xtype: 'field', value: 15, id: 'limitData', width: 30 },'-',
            {xtype: 'label', text: 'в период с: '},
            this.dateFrom, this.timeFrom,
            {xtype: 'label', text: ' по: '},
            this.dateTo, this.timeTo,
            '-', '->', '-',
            { text: 'Обновить', handler: this.reloadData, scope: this }, '-',
            { text: 'Обновилять автоматически', handler: this.autoToggle, scope: this }
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
                dateFormat: 'M-d H:i:s',
                groupBy: 'year,month,day',
                //aggregateOp: 'sum',
                constrain: true
            }],
            series: [{
                type: 'line',
                axis: ['left', 'bottom'],
                xField: 'date',
                yField: 't',
                highlight: true,
                //showInLegend: true,
                label: {
                    display: 'over',
                    field: 't',
                    renderer: function(v) { return v + '°C'; },
                    'text-anchor': 'middle'
                },
                markerConfig: {
                    type: 'circle', radius: 2, size: 2
                }
            },{
                type: 'line',
                axis: ['left', 'bottom'],
                xField: 'date',
                yField: 'h',
                highlight: true,
                //showInLegend: true,
                label: {
                    display: 'over',
                    field: 'h',
                    renderer: function(v) { return v  + '%'; },
                    'text-anchor': 'middle'
                },
                markerConfig: {
                    type: 'circle', radius: 2, size: 2
                }
            }]
        });
    },
    closeWin: function(ev) {
        this.close();
    },
    openWin: function() {
        var from = Ext.util.Format.date(this.dateFrom.value, 'Y-m-d') + ' ' + Ext.util.Format.date(this.timeFrom.value, 'H:i:s');
        var to = Ext.util.Format.date(this.dateTo.value, 'Y-m-d') + ' ' + Ext.util.Format.date(this.timeTo.value, 'H:i:s');
        Ext.Ajax.request({
            url: '/api/sensors/stat', scope: this, method: 'POST',
            params: {device_id: this.devId, count: 15, from: from, to: to},
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
            var date = new Date(dd[0], dd[1]-1, dd[2], tt[0], tt[1], tt[2]);
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
        var c = Ext.getCmp('limitData').value || 15;
        var from = Ext.util.Format.date(this.dateFrom.value, 'Y-m-d') + ' ' + Ext.util.Format.date(this.timeFrom.value, 'H:i:s');
        var to = Ext.util.Format.date(this.dateTo.value, 'Y-m-d') + ' ' + Ext.util.Format.date(this.timeTo.value, 'H:i:s');
        Ext.Ajax.request({
            url: '/api/sensors/stat', scope: this, method: 'POST',
            params: {device_id: this.devId, count: c, from: from, to: to},
            success: function(response, opts) {
                var ansv = Ext.decode(response.responseText);
                if(ansv.success) {
                    this.pushData(ansv.data);
                    //this.show();

                } else { error_mes('Ошибка', ansv.msg); }
            },
            failure: function() { }
        });
    },
    autoToggle: function() {
        this.autoReload = ! this.autoReload;

    }
})