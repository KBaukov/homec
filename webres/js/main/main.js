Ext.onReady(function(){
  
  Ext.BLANK_IMAGE_URL = '/webres/extjs/img/s.gif';

  var MainToolBar = {
      region: 'south', xtype: 'toolbar', height: 30, border: true,
      items: [
          {xtype: 'tbtext', text: '&nbsp;Имя пользователя: <b>'+user.login+'</b>'}, '->',
          {text: 'Выход', handler: CloseWin, id: 'close-btn',icon: '/webres/extjs/img/exit.png'}
      ]
  };
  
  var ControlTab  = Ext.create('ControlTab');
  var SettingsTab  = Ext.create('SettingsTab');
  var MapPanel  = Ext.create('MapsTab');
  
  var MainPanel  = Ext.create('Ext.tab.Panel',{
    region: 'center',
    items: [ ControlTab, MapPanel, SettingsTab ]
  });
  
  MainPanel.listeners = { scope: MainPanel,
    afterrender: function(){ 
       this.setActiveTab(1);
       this.setActiveTab(0);
    }
  };
  
  var delimHr = { xtype: 'box', html:'', style: 'border-bottom: 1px solid #99BCE8;', height: 5};
  var delim = { xtype: 'box', html:'',height: 5};

  Ext.create('Ext.container.Viewport', {
    layout: 'border', 
    items: [ MainToolBar, MainPanel ]
  });//.render('vp');
 

  //ButtonPanel.updateSate();
//  
  Ext.getDom('boby').onmousemove = function(e) { 
      slCounter = false;
      //console.log('--- ' + slCounter);
  };
  slCounter = false;
  screenLockT=setTimeout(screenLock, 3000000);
});