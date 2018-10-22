Ext.define('KotelTabPanel', {
    extend: 'Ext.tab.Panel',   
    region: 'center',
    title: 'Отопление',   
    //layout: 'border',
    border: false, frame: false,
    initComponent: function() {
      this.initPanel(); 
      this.items =  [ this.kotelViewPanel, this.kotelControlPanel ];
      KotelTabPanel.superclass.initComponent.apply(this, arguments);
    }, 
    initPanel: function() {
      this.kotelViewPanel   = Ext.create('KotelViewPanel',  {papa: this} );
      this.kotelControlPanel   = Ext.create('KotelControlPanel',  {papa: this} );
      //this.devices = Ext.create('Devices',{papa: this});      
      //this.formPanel    = Ext.create('SearchFormPanel',{papa: this});
      this.listeners = { scope: this,
            afterrender: function(){ 
               this.setActiveTab(1);
               this.setActiveTab(0);
            }
        };
    }
});