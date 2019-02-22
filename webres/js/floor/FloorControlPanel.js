Ext.define('FloorControlPanel', {
    extend: 'Ext.panel.Panel',
    initComponent: function() {
        this.title = 'Управление';
        this.border = true;
        this.frame = true;
        //this.region = 'center';
        //this.height = 100;
        this.margins = '2 2 2 2';
        this.id = 'floorControlPanel';
        //this.collapsible = true;
        this.maskOnDisable = true;
        this.collapsed = false;
        this.resizable = false;
        this.bodyPadding = 10;
        this.bodyStyle = 'padding:10px;';

        this.destf = 0;
        this.mode = 0;

        this.initForm();

        FloorControlPanel.superclass.initComponent.apply(this, arguments);
    },
    initForm: function() {
        this.papa = this.initConfig().papa;

        this.task = { scope: this,
            run: function() { this.blink(); },
            interval: 500
        };

        this.valuesTask = {
            scope: this,
            run: function() {  this.getValues(); },
            interval: 2000
        };

        this.html = '<div id="flBase">'
            //+'<div class="led" id="led1" on="on"></div><div class="led" id="led2"></div>'
            //+'<div id="dig1"></div><div id="dig2"></div><div id="dot"></div>'
            +'<div id="flleftButt"></div><div id="flrightButt"></div>'
            +'</div>';

        //this.kotelId = devices.getKotelId();


        // this.tbar = [
        //     Ext.create('Ext.Button', {text: 'Отменить', scope: this, disabled: false, id: 'execButt',
        //         style: 'background-position: bottom center;',
        //          handler: function() { this.command = ''; }
        //     }),
        //     '->', '-',
        //     Ext.create('Ext.Button', {text: 'Применить', scope: this, disabled: false, id: 'execButt',
        //         style: 'background-position: bottom center;',
        //         handler: function() { this.sendCommands(); }
        //     })
        // ];

        this.listeners = { scope: this,
            afterrender: function(){
                this.led1 = document.getElementById("led1");
                this.d1 =   document.getElementById("dig1");
                this.d2 =   document.getElementById("dig2");
                this.dot =  document.getElementById("dot");
                this.bl =   document.getElementById("flleftButt");
                this.br =   document.getElementById("flrightButt");


                this.bl.onclick = this.buttClick;
                this.br.onclick = this.buttClick;
                //this.bo.onclick = this.buttClick;

                this.bl.ontouch = this.buttClick;
                this.br.ontouch = this.buttClick;
                //this.bo.ontouch = this.buttClick;

                this.getValues();

                //this.dispCurrentView();
                //this.getValues();
                //Ext.TaskManager.start(this.valuesTask);
            }
        };


    },
    onMessage: function(butt) {
        if(butt=='L')
            this.leftClick();
        if(butt=='R')
            this.rightClick();
        if(butt=='M')
            this.okClick();

        //this.setDest();
        this.setDisabled(false);
    },
    leftClick: function(ev) {
        //var cmp =Ext.getCmp('kotelControlPanel');
        //cmp.pressButt("L");
        // if(this.mode==0) {
        //     if(this.curSatgeInx==0) {
        //         this.curSatgeInx = 3;
        //     } else
        //         this.curSatgeInx--;
        //
        //     this.setLedPosition(this.curSatgeInx);
        // }
        // if(this.mode==1 && this.stage[this.curSatgeInx]=='ot') {
        //     var val = this.destVal[this.curSatgeInx] == 'mm' ? 25 : parseInt(this.destVal[this.curSatgeInx]);
        //     val--;
        //     if(val<25) { val = 'mm'; }
        //
        //     this.destVal[this.curSatgeInx] = val+'';
        // }
        // if(this.mode==1 && this.stage[this.curSatgeInx]=='vs') {
        //     var val = this.destVal[this.curSatgeInx] == 'mm' ? 35 : parseInt(this.destVal[this.curSatgeInx]);
        //     val--;
        //     if(val<35) { val = 'mm'; }
        //
        //     this.destVal[this.curSatgeInx] = val+'';
        // }
        // if(this.mode==1 && this.stage[this.curSatgeInx]=='va') {
        //     this.curVaInx = this.curVaInx == 0 ? 0 : (--this.curVaInx);
        //     var val = this.va[this.curVaInx];
        //     this.destVal[this.curSatgeInx] = val+'';
        // }

        this.dispCurrentView();

        //this.command += "L";
    },
    rightClick: function(ev) {
        //var cmp =Ext.getCmp('kotelControlPanel');
        //cmp.pressButt("R");
        // if(this.mode==0) {
        //     if(this.curSatgeInx==3) {
        //         this.curSatgeInx = 0;
        //     } else
        //         this.curSatgeInx++;
        //     this.setLedPosition(this.curSatgeInx);
        // }
        // if(this.mode==1 && this.stage[this.curSatgeInx]=='ot') {
        //     var val = this.destVal[this.curSatgeInx] == 'mm' ? 24 : (this.destVal[this.curSatgeInx] == '80' ? 79 : parseInt(this.destVal[this.curSatgeInx])) ;
        //     val++;
        //     //if(val>99) { val = "mm"; }
        //
        //     this.destVal[this.curSatgeInx] = val+'';
        // }
        // if(this.mode==1 && this.stage[this.curSatgeInx]=='vs') {
        //     var val = this.destVal[this.curSatgeInx] == 'mm' ? 34 : (this.destVal[this.curSatgeInx] == '70' ? 69 : parseInt(this.destVal[this.curSatgeInx])) ;
        //     val++;
        //     //if(val>99) { val = "mm"; }
        //
        //     this.destVal[this.curSatgeInx] = val+'';
        // }
        // if(this.mode==1 && this.stage[this.curSatgeInx]=='va') {
        //     this.curVaInx = this.curVaInx == 5 ? 5 : (++this.curVaInx);
        //     var val = this.va[this.curVaInx];
        //     this.destVal[this.curSatgeInx] = val+'';
        // }

        this.dispCurrentView();

        //this.command += "R";
    },
    okClick: function(ev) {
        //var cmp =Ext.getCmp('kotelControlPanel');
        //cmp.pressButt("M");
        // if(this.stage[this.curSatgeInx]=='pr') return;
        // if(this.mode==0) {
        //     this.mode=1;
        //     this.startBlink();
        //     this.display(this.destVal[this.curSatgeInx]);
        // } else {
        //     this.mode=0;
        //     this.stopBlink();
        //     //this.setDest();
        //     this.display(this.currVal[this.curSatgeInx]);
        // }

        //this.dispCurrentView();

        //this.command += "M";
    },
    dispCurrentView: function() {
        if(this.mode==0)
            this.display(this.currVal[this.curSatgeInx]);
        else
            this.display(this.destVal[this.curSatgeInx]);
    },
    display: function(val) {
        // var d1 = '';
        // var d2 = '';
        // if(val.length>1) {
        //     d1 = val.substr(0,1);
        //     if(val.indexOf('.')>-1) {
        //         d2 = val.substr(2,1);
        //         this.dot.style.opacity=1;
        //     } else {
        //         d2 = val.substr(1,1);
        //         this.dot.style.opacity=0;
        //     }
        //
        // } else {
        //     d2 = val.substr(0,1);
        // }
        //
        // this.d1.style.backgroundImage ='url("../webres/img/kotel/dr'+d1+'.png")';
        // this.d2.style.backgroundImage ='url("../webres/img/kotel/dr'+d2+'.png")';
    },
    buttClick: function(ev) {
        var id = ev.target.id;
        var butt = id.substring(0,1).toUpperCase();
        var cmp =Ext.getCmp('kotelControlPanel');
        if(cmp.stage[cmp.curSatgeInx]=='pr' && butt=='M'){
            cmp.setDisabled(true);
            setTimeout( function(cmp) { var cmp =Ext.getCmp('kotelControlPanel'); cmp.setDisabled(false) }, 200);
            return;
        }
        var stage = cmp.nextStage(butt);
        var hash = md5((new Date()).toLocaleString());
        var rMsg = '{"action":"resend", "recipient":"'+cmp.kotelId+'", "msg":"'
            +btoa('{"action":"pessButton","butt":"'+butt+'","sender":"","hash":"'+hash+'","stage":"'+stage+'"}') //
            +'"}';
        cmp.setDisabled(true);
        cmp.papa.wss.butt = butt;
        cmp.papa.wss.hash = hash;
        cmp.papa.wss.send(rMsg);
    },
    getValues: function() {
        Ext.Ajax.request({
            url: '/api/floor/getvalues', scope: this, method: 'GET',
            success: function(response, opts) {
                var ansv = Ext.decode(response.responseText);
                if(ansv.success) {
                    this.tf = parseFloat(ansv.data.tf);
                    this.ta = parseFloat(ansv.data.ta);
                    this.ha = parseFloat(ansv.data.ha);
                    this.destTf = parseFloat(ansv.data.destTf);
                } else error_mes('Ошибка', ansv.msg);
            },
            failure: function() { }
        });
    },
    setDest: function() {
        var suf = '.0';
        var currStage = this.mode+'_'+this.curSatgeInx;
        var data = {
            destTf: this.destTf,
            stage: currStage
        };
        Ext.Ajax.request({
            url: '/api/floor/setdest', scope: this, method: 'POST',
            params: data,
            success: function(response, opts) {
                var ansv = Ext.decode(response.responseText);
                if(ansv.success) {
                    this.papa.stopSession();
                } else error_mes('Ошибка', ansv.msg);
            },
            failure: function() { }
        });
    }

});