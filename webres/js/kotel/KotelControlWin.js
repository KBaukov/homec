Ext.define('KotelControlWin', {
    extend: 'Ext.Window',
    initComponent: function () {
        this.title='Управление отоплением';
        this.resizable=false;
        this.closable=true;
        this.width=435;
        this.height=550;
        this.plain=true;
        this.modal= true;
        this.headerPosition='top';
        this.layout='border';

        this.shadow='drop';
        this.shadowOffset=8;
        this.buttons = [{ text: 'Закрыть', handler: this.closeWin, scope: this }];

        this.initForm();
        this.items=[this.kotelControlPanel];
        //this.html = "rtrgrtrt";

        KotelControlWin.superclass.initComponent.apply(this, arguments);
    },
    initForm: function() {
        //this.kotelTabPanel   = Ext.create('KotelTabPanel',  {papa: this} );
        this.kotelControlPanel   = Ext.create('KotelControlPanel',  {papa: this} );
    },
    closeWin: function(ev) {
        this.close();
    }
})