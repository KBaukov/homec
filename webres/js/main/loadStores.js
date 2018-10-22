devices = Ext.create('Ext.data.JsonStore', {
    storeId: 'devicesSensData', autoLoad: true,   
      proxy: {
          type: 'ajax',
          url: '/api/devices',
          reader: {
              type: 'json',
              root: 'data',
              idProperty: 'id'
          }
      },
    fields: [
      {name: 'id'}, {name: 'type'}, {name: 'name'}, {name: 'ip'},
      {name: 'active_flag'}, {name: 'description'}
    ]//,
});
