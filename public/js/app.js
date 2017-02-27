var sifApp = angular.module('sifApp', [
    'ngRoute',
    'sifControllers'
]);

sifApp.config(['$routeProvider', '$locationProvider',
    function($routeProvider, $locationProvider) {
        $routeProvider.
            when('/', {
                templateUrl: 'partials/search.html',
                controller: 'SearchCtrl'
            }).
            when('/item/:itemid', {
                templateUrl: 'partials/find.html',
                controller: 'FindCtrl'
            }).
            when('/item/:itemid/serie/:serie', {
                templateUrl: 'partials/find.html',
                controller: 'FindCtrl'
            }).
            when('/profile/:steamid/inventory/440/item/:itemid', {
                templateUrl: 'partials/inventory440_item.html',
                controller: 'InventoryItemCtrl'
            }).
            when('/profile/:steamid/inventory/440', {
                templateUrl: 'partials/inventory440.html',
                controller: 'InventoryCtrl'
            }).
            otherwise({
                redirectTo: '/'
            });
        $locationProvider.html5Mode(true);
    }
]);
