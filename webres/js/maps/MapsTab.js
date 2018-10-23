Ext.define('MapsTab', {
    extend: 'Ext.tab.Panel',
    initComponent: function() {
        this.title = 'Катры';
        this.border = false;
        this.frame = false;
        this.region = 'center';
        this.mapsData = [];
        this.loadMaps();

        this.items = [ Ext.create('Ext.panel.Panel', null) ];

        this.superclass.initComponent.apply(this, arguments);
    },
    loadMaps: function() {
        Ext.Ajax.request({
            url: '/api/maps', scope: this, method: 'POST',
            success: function(response, opts) {
                var ansv = Ext.decode(response.responseText);
                if(ansv.success) {
                    this.mapsData = ansv.data;
                    this.createMapsTabs();
                }
            },
            failure: function() { this.unmask(); }
        });
    },
    createMapsTabs: function() {
        var n = this.mapsData.length;
        var tabs = new Array(n);
        for(var i=0; i<n; i++) {
            tabs[i] = Ext.create('MapPanel',  {mapData: this.mapsData[i] } );
        }
        this.items  = tabs;
        this.superclass.initComponent.apply(this, arguments);
    }

});