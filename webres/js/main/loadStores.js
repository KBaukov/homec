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
    ],
    getName: function(id) {
        var n = devices.data.length;
        for(var i=0; i<n; i++ ) {
            if(devices.getAt(i).data.id == id) {
                return devices.getAt(i).data.name;
            }
        }
        return '';
    },getId: function(name) {
        var n = devices.data.length;
        for(var i=0; i<n; i++ ) {
            if(devices.getAt(i).data.name == name) {
                return devices.getAt(i).data.id;
            }
        }
        return '';
    },
    getType: function(id) {
        var n = devices.data.items.length;
        for(var i=0; i<n; i++ ) {
            if(devices.getAt(i).data.id == id) {
                return devices.getAt(i).data.type;
            }
        }
        return '';
    },
    getKotelId: function(id) {
        var n = devices.data.items.length;
        for(var i=0; i<n; i++ ) {
            if(devices.getAt(i).data.type == 'KotelController') {
                return devices.getAt(i).data.name;
            }
        }
        return '';
    }
});


