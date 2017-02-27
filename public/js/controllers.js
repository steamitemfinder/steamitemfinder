
var sifControllers = angular.module('sifControllers', []);

sifControllers.controller('HeaderCtrl', ['$scope', '$http', function($scope, $http) {
    var url = '/steam/general/user';

    $http.get(url)
        .success(function(User, code) {
            $scope.user = User;
            if(code == 204) {
                $scope.notlogged = true;
            }
        })
        .error(function(html, code) {
            $scope.notlogged = true;
            if(code == 402) {
                $scope.error = html;
            } else {
                $scope.error = "internal api seems to have a problem";
            }
        });
}]);

sifControllers.controller('SearchCtrl', ['$scope', '$http', '$location', function($scope, $http, $location) {
    $scope.message = "type something in the box to search for an item";

    $scope.search = function() {
        if(typeof $scope.query !== 'undefined' && $scope.query.length > 1) {
            $scope.message = "loading..";
            var url = '/steam/440/items?query=' + $scope.query;
            $http.get(url)
                .success(function(Items, code) {
                    if(code == 204) {
                        $scope.items = false;
                        $scope.message = "Oops. We didn't find any matches!";
                    } else {
                        $scope.items = Items;
                        $scope.message = false;
                    }
                })
                .error(function(html, code) {
                    $scope.items = false;
                    $scope.message = "Api seems to have a problem :(";
                });
        } else {
            $scope.items = false;
            $scope.message = "Please type more than one character to start the search.";
        }
    }

    $scope.find = function(item) {
        var serie = "";
        var itemid = item.defindex;
        if(Array.isArray(item.Attributes)) {
            for (i = 0; i < item.Attributes.length; i++) {
                // FIXME: check if Class exists
                if(item.Attributes[i].Class == "supply_crate_series") {
                    serie = "187";
                    itemid = "5734";
                }
            }
        }
        if(serie != "") {
            var tofind = '/item/' + itemid + "/serie/" + serie;
        } else {
            var tofind = '/item/' + itemid;
        }
        $location.path(tofind);
    }

    var tooltip = angular.element( document.querySelector('#tooltip'));

    $scope.hideTip = function() {
        tooltip.css("display", "none");
    }

    $scope.showTip = function(item) {
        tooltip.css("display", "block");
        var desc = "<h5>";
        if(item.proper_name) {
            desc = desc + "The ";
        }
        desc = desc + item.item_name + "</h5>";
        if(item.used_by_classes) {
            desc = desc + "<p>" + item.used_by_classes.join(", ") +"</p>";
        }
        if(item.item_slot) {
            desc = desc + "<p>" + item.item_slot + " slot <small>" + item.item_type_name + "</small></p>";
        } else {
            desc = desc + "<p>" + item.item_type_name + "</p>";
        }

        if(item.holiday_restriction) {
            desc = desc + "<p class='orange'>Holiday Restriction: Halloween / Full Moon</p>";
        }
        tooltip.html(desc);
    }

    // for tooltip
    $scope.trackMouse = function(e) {
        var cursorX = (window.Event) ? e.pageX : event.clientX + (document.documentElement.scrollLeft ? document.documentElement.scrollLeft : document.body.scrollLeft);
        var cursorY = (window.Event) ? e.pageY : event.clientY + (document.documentElement.scrollTop ? document.documentElement.scrollTop : document.body.scrollTop);
        tooltip.css("top", cursorY + "px");
        tooltip.css("left", cursorX + "px");
    }

}]);

function getScrollingPosition()
{
    var position = [0, 0];
    if (typeof window.pageYOffset != 'undefined')
    {
        position = [
            window.pageXOffset,
            window.pageYOffset
                ];
    }
    else if (typeof document.documentElement.scrollTop
            != 'undefined' && document.documentElement.scrollTop > 0)
    {
        position = [
            document.documentElement.scrollLeft,
            document.documentElement.scrollTop
                ];
    }
    else if (typeof document.body.scrollTop != 'undefined')
    {
        position = [
            document.body.scrollLeft,
            document.body.scrollTop
                ];
    }
    return position;
}

sifControllers.controller('FindCtrl', ['$scope', '$http', '$routeParams', function($scope, $http, $routeParams) {
    var itemid = $routeParams.itemid;
    var serie = $routeParams.serie;
    var filterBy = "item";
    var filter = itemid;
    if(typeof serie != 'undefined') {
        filterBy = "serie";
        filter = serie;
        $scope.serie = true;
    }

    $scope.message = "loading item info.";
    $scope.loading = true;
    var url = '/steam/440/item/' + itemid;
    $http.get(url)
        .success(function(Item, code) {
            $scope.item = Item;

            $scope.message = "loading your account.";
            url = '/steam/general/user';
            var totaldone = 0;
            $http.get(url)
            .success(function(User, code) {
                $scope.user = User;
                if(code == 204) {
                    $scope.notlogged = true;
                } else {
                    $scope.pourc = 1;
                    $scope.message = "loading your friends list.";
                    url = '/steam/general/user/'+ User.SteamId +'/friends';
                    $http.get(url)
                        .success(function(Friends, code) {
                            var friendsnb = Friends.length;
                            $scope.message = "loading backpacks of your " + friendsnb + " friends.";
                            if(Friends == "null") {
                                $scope.error = "Unable to access to your friends list. Your profile may be private.";
                                $scope.message = false;
                                $scope.loading = false;
                                return;
                            } else if(friendsnb == 0) {
                                $scope.error = "You need to have at least 1 friend on Steam.";
                                $scope.message = false;
                                $scope.loading = false;
                                return;
                            }
                            $scope.hasresultok = true;
                            $scope.hasresultko = false;
                            $scope.hasresultno = false;
                            $scope.resultok = {};
                            $scope.resultko = {};
                            $scope.resultno = {};
                            $scope.sum = 0;
                            $scope.pourc = 3;
                            for (var i = 0, len = Friends.length; i < len; i++) {
                                var friend = Friends[i].SteamId;
                                url = '/steam/440/user/'+friend+'/backpack?' + filterBy + '=' + filter;
                                $http.get(url)
                                    .success(function(Backpack, code) {
                                        totaldone++;
                                        $scope.message = "backpacks loaded: " + totaldone + " of " + friendsnb + ".";
                                        $scope.pourc = totaldone * 97 / friendsnb + 3;
                                        if(totaldone == Friends.length) {
                                            $scope.message = "All backpacks loaded.";
                                            $scope.loading = false;
                                        }
                                        if(Backpack.Status == 1 && Backpack.Items) {
                                            console.log(Backpack);
                                            $scope.hasresultok = true;
                                            $scope.sum = $scope.sum + Backpack.Items.length;
                                            $scope.resultok[Backpack.Player.SteamId] = Backpack;
                                        } else if (Backpack.Status == 1) {
                                            $scope.hasresultno = true;
                                            $scope.resultno[Backpack.Player.SteamId] = Backpack;
                                        } else {
                                            $scope.hasresultko = true;
                                            $scope.resultko[Backpack.Player.SteamId] = Backpack;
                                        }

                                });
                                
                            }


                    });
                    
                }
            })
            .error(function(html, code) {
                $scope.notlogged = true;
                if(code == 402) {
                    $scope.error = html;
                } else {
                    $scope.error = "internal api seems to have a problem";
                }
            });

        })
        .error(function(html, code) {
            if(code == 404) {
                $scope.error = "Couldn't find any item with this id..";
            } else {
                $scope.error = "Api seems to have a problem :( ";
            }
        });

    var tooltip = angular.element( document.querySelector('#tooltip'));

    $scope.hideTip = function() {
        tooltip.css("display", "none");
    };

    $scope.showTip = function(sifBp) {
        var desc = "";

        if(sifBp.Player) {

            desc = sifBp.Player.PersonaName;
            if(sifBp.Items && sifBp.Items.constructor === Array) {
                desc = desc + " has got " + sifBp.Items.length;
            }
            tooltip.css("display", "block");
        } else if(sifBp.DefIndex) {
            item = sifBp;
            desc = "<h5>";
            if(item.NewName) {
                desc = "<h5><i class='fa fa-tag'></i> " + item.NewName + "</h5>";
            } else if(item.ProperName && item.Quality == "6") {
                desc = desc + "The ";
            } else if(item.Quality == "1") {
                desc = desc + "Genuine ";
            } else if(item.Quality == "3") {
                desc = desc + "Vintage ";
            } else if(item.Quality == "5") {
                desc = desc + "Unusual ";
            } else if(item.Quality == "7") {
                desc = desc + "Community ";
            } else if(item.Quality == "8") {
                desc = desc + "Valve ";
            } else if(item.Quality == "9") {
                desc = desc + "Self-Made ";
            } else if(item.Quality == "11") {
                desc = desc + "Strange ";
            } else if(item.Quality == "13") {
                desc = desc + "Haunted ";
            } else if(item.Quality == "14") {
                desc = desc + "Collector's ";
            }
            desc = desc + item.Name + "</h5><p>Level " + item.Level + " "+ item.TypeName +"</p>";


            if(item.CraftNumber && !item.CraftedById) {
                desc = desc + "<p>Craft number (restored): "+ item.CraftNumber +"</p>";
            }

            if(item.CraftNumber && item.CraftedById) {
                desc = desc + "<p>Craft number: "+ item.CraftNumber +"</p>";
            }

            if(item.NewDescription) {
                desc = desc + '<p><i '+"class='fa fa-tag'"+'></i> "'+ item.NewDescription +'"</p>';
            }

            if(item.Origin) {
                desc = desc + "<p><i class='fa fa-certificate'></i> Untouched ( "+ item.Origin +" )";
            }

            if(item.NotTradable && item.NotCraftable) {
                desc = desc + "<p class='red'>( Not Tradable or Usable in Crafting )</p>";
            } else if(item.NotTradable) {
                desc = desc + "<p class='red'>( Not Tradable )</p>";
            } else if(item.NotCraftable) {
                desc = desc + "<p class='red'>( Not Usable in Crafting )</p>";
            }

            tooltip.css("display", "block");
        } else {

            tooltip.css("display", "none");
        }
        tooltip.html(desc);
    };

    // for tooltip
    $scope.trackMouse = function(e) {
        var cursorX = (window.Event) ? e.pageX : event.clientX + (document.documentElement.scrollLeft ? document.documentElement.scrollLeft : document.body.scrollLeft);
        var cursorY = (window.Event) ? e.pageY : event.clientY + (document.documentElement.scrollTop ? document.documentElement.scrollTop : document.body.scrollTop);
        tooltip.css("top", cursorY + "px");
        tooltip.css("left", cursorX + "px");
    };


}]);

sifControllers.controller('InventoryCtrl', ['$scope', '$http', '$location', '$routeParams', function($scope, $http, $location, $routeParams) {
    var steamid = $routeParams.steamid;
    var url = '/steam/440/user/' + steamid + '/backpack';
    $http.get(url)
        .success(function(sifBp, code) {
            $scope.sifBp = sifBp;
            if(sifBp.Status == 1) {
                $scope.error = false;
            } else {
                $scope.error = "backpack unavailable";
            }
        })
        .error(function(html, code) {
            if(code == 404) {
                $scope.error = "invalid steam id";

            } else if(code == 410) {
                $scope.error = "private backpack";

            } else {
                $scope.error = "api not available";
            }
        }
    );

    var tooltip = angular.element( document.querySelector('#tooltip'));

    $scope.hideTip = function() {
        tooltip.css("display", "none");
    };

    $scope.showTip = function(sifBp) {
        var desc = "";

        if(!sifBp) {
            return
        }

        if(sifBp.Player) {

            desc = sifBp.Player.PersonaName;
            if(sifBp.Items && sifBp.Items.constructor === Array) {
                desc = desc + " has got " + sifBp.Items.length;
            }
            tooltip.css("display", "block");
        } else if(sifBp.DefIndex) {
            item = sifBp;
            desc = "<h5>";
            if(item.NewName) {
                desc = "<h5><i class='fa fa-tag'></i> " + item.NewName + "</h5>";
            } else if(item.ProperName && item.Quality == "6") {
                desc = desc + "The ";
            } else if(item.Quality == "1") {
                desc = desc + "Genuine ";
            } else if(item.Quality == "3") {
                desc = desc + "Vintage ";
            } else if(item.Quality == "5") {
                desc = desc + "Unusual ";
            } else if(item.Quality == "7") {
                desc = desc + "Community ";
            } else if(item.Quality == "8") {
                desc = desc + "Valve ";
            } else if(item.Quality == "9") {
                desc = desc + "Self-Made ";
            } else if(item.Quality == "11") {
                desc = desc + "Strange ";
            } else if(item.Quality == "13") {
                desc = desc + "Haunted ";
            } else if(item.Quality == "14") {
                desc = desc + "Collector's ";
            }
            desc = desc + item.Name + "</h5><p>Level " + item.Level + " "+ item.TypeName +"</p>";


            if(item.CraftNumber && !item.CraftedById) {
                desc = desc + "<p>Craft number (restored): "+ item.CraftNumber +"</p>";
            }

            if(item.CraftNumber && item.CraftedById) {
                desc = desc + "<p>Craft number: "+ item.CraftNumber +"</p>";
            }

            if(item.NewDescription) {
                desc = desc + '<p><i '+"class='fa fa-tag'"+'></i> "'+ item.NewDescription +'"</p>';
            }

            if(item.Origin) {
                desc = desc + "<p><i class='fa fa-certificate'></i> Untouched ( "+ item.Origin +" )";
            }

            if(item.NotTradable && item.NotCraftable) {
                desc = desc + "<p class='red'>( Not Tradable or Usable in Crafting )</p>";
            } else if(item.NotTradable) {
                desc = desc + "<p class='red'>( Not Tradable )</p>";
            } else if(item.NotCraftable) {
                desc = desc + "<p class='red'>( Not Usable in Crafting )</p>";
            }

            tooltip.css("display", "block");
        } else {
            tooltip.css("display", "none");
        }
        tooltip.html(desc);
    };

    // for tooltip
    $scope.trackMouse = function(e) {
        var cursorX = (window.Event) ? e.pageX : event.clientX + (document.documentElement.scrollLeft ? document.documentElement.scrollLeft : document.body.scrollLeft);
        var cursorY = (window.Event) ? e.pageY : event.clientY + (document.documentElement.scrollTop ? document.documentElement.scrollTop : document.body.scrollTop);
        tooltip.css("top", cursorY + "px");
        tooltip.css("left", cursorX + "px");
    };
}]);

sifControllers.controller('InventoryItemCtrl', ['$scope', '$http', '$routeParams', function($scope, $http, $routeParams) {
    var steamid = $routeParams.steamid;
    var itemid = $routeParams.itemid;
    var url = '/steam/440/user/' + steamid + '/backpack?id=' + itemid;
    $http.get(url)
        .success(function(sifBp, code) {
            $scope.sifBp = sifBp;
        }
    );
}]);

sifControllers.controller('ProfileCtrl', ['$location', '$routeParams', function($location, $routeParams) {
    var steamid = $routeParams.steamid;
    // DOING
    //$location.path("/profile/"+steamid+"/inventory/440");
    console.log(steamid);
}]);


function sleep(milliseconds) {
  var start = new Date().getTime();
  for (var i = 0; i < 1e7; i++) {
    if ((new Date().getTime() - start) > milliseconds){
      break;
    }
  }
}
