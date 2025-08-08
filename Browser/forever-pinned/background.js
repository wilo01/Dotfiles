// Manifest V3 Service Worker for Forever Pinned

// Import scripts
importScripts('js/q.js', 'js/lodash.js');

// Main functionality
var buildTabs = {
  list:[],
  createSyncItems:function(){
    var deferItems = Q.defer();
    var deferOptions = Q.defer();

    chrome.storage.sync.get('items', function(data){
      if(typeof(data.items) === "undefined"){
        chrome.storage.sync.set({items:[]}, function (data){
          deferItems.resolve();
        });
      }else{
        deferItems.resolve();
      }
    });

    chrome.storage.sync.get('options', function(data){
      if(typeof(data.options) === "undefined"){
        chrome.storage.sync.set({options:{}}, function(data){
          deferOptions.resolve();
        });
      }else{
        deferOptions.resolve();
      }
    });

    return [deferItems.promise, deferOptions.promise];
  },
  get:function(){
    var self = this;
    var defer = Q.defer();
    defer.resolve('done');
    chrome.storage.sync.get('items', function(data){
      self.removeDups(data.items);
    });
    return defer.promise;
  },
  init:function(){
    var self = this;
    var createSyncItemsPromise = this.createSyncItems();
    Q.all(createSyncItemsPromise).then(function(){
      self.closeDups();
    });
  },
  create:function(){
    for(var i = 0; i < this.list.length; i++){
      chrome.tabs.create({
        pinned:true,
        url:this.list[i].url,
        active:false
      });
    }
  },
  readAllTabs:function(){
    var defer = Q.defer();
    chrome.tabs.query({pinned:true, currentWindow:true}, function(data){
      defer.resolve(data);
    });
    return defer.promise;
  },
  closeDups:function(){
    var self= this;

    this.readAllTabs().then(function(data){
      var arr  = data;
      var dupsId = [];
      var dupsUrl = [];
      var list = [];
      var defer = Q.defer();

      chrome.storage.sync.get('items', function(urlList){
        var urlList = urlList.items;

        for(var i=0; i < urlList.length; i++){
          list.push(urlList[i].url);
        }

        for(var i=0; i<arr.length; i++){
          arr[i].url =self._removeSlash(arr[i].url);
        }
       
       var dupsId = [];
       var unqi = _.uniq(arr, 'url');
       
       arr  = _.filter(arr, function(item){
        return !(_.includes(unqi, item));
       })
       
       _.forEach(arr, function(item){
         dupsId.push(item.id);
       });
        
        chrome.tabs.remove(dupsId, function(){
          self.get();
        });
      });

    });
  },
  removeDups:function(urlList){
    var self = this;
    var list = [];
    var open = [];
    var alreadyOpenedList = [];

    this.list = [];
    this.readAllTabs().then(function(openTabs){
      for(var i=0; i < openTabs.length; i++){
        if (openTabs[i].url.substr(-1) === '/') {
          openTabs[i].url =  openTabs[i].url.substr(0, openTabs[i].url.length - 1);
         }
        open.push(openTabs[i].url);
      }
      for(var i=0; i < urlList.length; i++){
        list.push(urlList[i].url);
      }
      if(applyOptions.options.ignoreParams) {
        notOpenList = list.slice(0);
        for(var i = 0; i < open.length; i++) {
          for(var j = 0; j < list.length; j++) {
            if (open[i].includes(list[j])) {
              alreadyOpenedList.push(list[j]);
            }
          }
        }
        list.forEach(function(listItem){
            if (!(alreadyOpenedList.indexOf(listItem) >= 0)) {
              self.list.push({url:listItem});
            }
        });
      }
      else {
        list.forEach(function(item){
          var localItem = item;
          if(!(open.indexOf(localItem)> -1)){
            self.list.push({url:item});
          }
        });
      }
      self.create();
    });
  },
  _removeSlash:function(str){
    if (str.substr(-1) === '/') {
      str = str.substr(0, str.length - 1);
    }
    return str;
  }
};

var applyOptions = {
  options:{},
  init:function(){
    var self = this;

    this.get()
    .then(function(){
      self.apply();
    });
  },
  get:function(){
    var self = this;
    var defer = Q.defer();
    chrome.storage.sync.get('options', function(data){
      self.options = data.options;
      defer.resolve(data.options);
    });
    return defer.promise;
  },
  apply:function(){
    if(this.options.reopen){
      chrome.windows.onCreated.addListener(function(){
        buildTabs.init();
      });
    }
  }
};

// Handle extension icon click
chrome.action.onClicked.addListener(function() {
  buildTabs.init();
});

// Handle startup
chrome.runtime.onStartup.addListener(function() {
  buildTabs.init();
  applyOptions.init();
});

// Handle installation
chrome.runtime.onInstalled.addListener(function() {
  buildTabs.init();
  applyOptions.init();
});