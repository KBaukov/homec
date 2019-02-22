Ext.define('FloorControlWin', {
    extend: 'Ext.Window',
    initComponent: function () {
        this.title='Управление контроллером теплого пола';
        this.resizable=false;
        this.closable=false;
        this.width=435;
        this.height=450;
        this.plain=true;
        this.modal= true;
        this.headerPosition='top';
        this.layout='border';

        this.shadow='drop';
        this.shadowOffset=8;
        this.buttons = [{ text: 'Закрыть', handler: this.closeWin, scope: this }];

        this.initForm();
        this.items=[ this.tabPanel ];

        FloorControlWin.superclass.initComponent.apply(this, arguments);
    },
    initForm: function() {
        this.papa = this.initConfig().papa;
        this.floorControlPanel   = Ext.create('FloorControlPanel',  {papa: this} );
        //this.floorChartPanel = Ext.create('FloorDataChartPanel',  {papa: this} );
        this.tabPanel = Ext.create('Ext.tab.Panel',{
            region: 'center',
            items: [
                this.floorControlPanel//,
                //this.floorChartPanel
            ]
        });

        //this.floorId = devices.getKotelId();

        this.wss = new WebSocket("wss://"+window.location.host+"/ws", [user.token]);
        this.wss.onopen = this.wssOnOpen;
        this.wss.onerror = function(error) { console.log("Ошибка " + error.message); };
        this.wss.onclose = function(event) {
            if (event.wasClean) { console.log('Соединение закрыто чисто'); }
            else { console.log('Обрыв соединения'); }
            console.log('Код: ' + event.code + ' причина: ' + event.reason);
            WssConnections.deleteByConn(event.target);
        };
        this.wss.onmessage = this.wssOnMessage;
        this.wss.papa = this;

    },
    wssOnOpen: function(event) {
        console.log("WSS Соединение установлено.");
        event.target.papa.startSession();
    },
    closeWin: function(ev) {
        this.floorControlPanel.setDest();
    },
    openWin: function() {
        this.show();
        //this.floorChartPanel.loadData();
    },
    wssOnMessage: function(event) {
        console.log("Входящее сообщение: " + event.data);
        var data = Ext.decode(event.data);
        var conn = event.target;
        var butt = conn.butt;
        var hash = conn.hash;
        var cmp = conn.papa;
        if(data.success) {
            if (data.hash == hash) {
                if(butt=='M' || butt=='L' || butt=='R') {
                    conn.papa.floorControlPanel.onMessage(butt);
                }

                if(data.action =='sessionStart') {
                    cmp.show();
                }

                if(data.action =='sessionStop') {
                    conn.close();
                    cmp.close();
                }

                //if(data.action =='setDestValues') {

                //}

            }
        } { if(data.success === false) error_mes('Ошибка', data.msg); }
    },
    startSession: function() {
        var hash = md5((new Date()).toLocaleString());
        var rMsg = '{"action":"resend", "recipient":"'+this.kotelId+'", "msg":"'
            +btoa('{"action":"sessionStart","sender":"","hash":"'+hash+'"}')
            +'"}';
        this.wss.hash = hash;
        this.wss.send(rMsg);
    },
    stopSession: function() {
        var hash = md5((new Date()).toLocaleString());
        var rMsg = '{"action":"resend", "recipient":"'+this.kotelId+'", "msg":"'
            +btoa('{"action":"sessionStop","sender":"","hash":"'+hash+'"}')
        +'"}';
        this.wss.hash = hash;
        this.wss.send(rMsg);
    }
})