Ext.define('SettingsTab', {
    extend: 'Ext.tab.Panel',   
    region: 'center',
    title: 'Настройки',   
    //layout: 'border',
    border: false, frame: false,
    initComponent: function() {
      this.initPanel(); 
      this.items =  [ this.users, this.devices, this.maps ];
      SettingsTab.superclass.initComponent.apply(this, arguments);
    }, 
    initPanel: function() {
      this.users   = Ext.create('Users',  {papa: this} );
      this.devices = Ext.create('Devices',{papa: this});      
      this.maps = Ext.create('MapsCtrl',{papa: this});      
    }
});