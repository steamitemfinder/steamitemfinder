package main

import (
    "fmt"
    "strings"
    "time"
    "io/ioutil"
    "encoding/json"
    //"encoding/hex"
    "net/http"
    "strconv"
    "hash/adler32"
    "crypto/sha512"
    "database/sql"
    "log"

    "github.com/go-martini/martini"
    "github.com/yohcop/openid-go"
    _ "github.com/go-sql-driver/mysql"
)

var apikey string
var apiurl string
var rootUrl string

var nonceStore = &openid.SimpleNonceStore{ Store: make(map[string][]*openid.Nonce) }
var discoveryCache = &openid.SimpleDiscoveryCache{}

type Friend struct {
    SteamId string
}

type User struct {
    SteamId string
    ProfileUrl string
    PersonaName string
    PersonaState int
    Avatar string
    AvatarMedium string
    AvatarFull string
    GameId string `json:",omitempty"`
    GameServerIp string `json:",omitempty"`
}

type BackpackResult struct {
    Status int
    Items []BackpackItem
    Size int64 `json:"num_backpack_slots"`
}

type SifBackpackItem struct {
    Status int
    Item *SifItem
    Player User
}

type SifBackpack struct {
    Status int
    Items []*SifItem
    Size int64
    Player User
}

type SifFullBackpack struct {
    Status int
    PagesOfItems [][]*SifItem
    Size int64
    Player User
    Stats SifBackpackStats
}

type SifBackpackStats struct {
    Total int
    Keys int
    RawMetal string
    Untouched int
}

type SchemaResult struct {
    Status int
    OriginNames []OriginType
    Items []SchemaItem
    Attributes []SchemaAttribute
}

type SchemaAttribute struct {
    Name string
    DefIndex int
    AttributeClass string `json:"attribute_class"`
    DescriptionString string `json:"description_string"`
    DescriptionFormat string `json:"description_format"`
    EffectType string `json:"effect_type"`
    Hidden bool
    IsInteger bool `json:"stored_as_integer"`
}

type OriginType struct {
    Origin int
    Name string
}

type BackpackItem struct {
    Id int64 `json:"id"`
    OriginalId int64 `json:"original_id"`
    DefIndex int64 `json:"defindex"`
    Level int `json:"level"`
    Quality int `json:"quality"`
    Origin int64 `json:"origin"`
    NotTradable bool `json:"flag_cannot_trade"`
    NotCraftable bool `json:"flag_cannot_craft"`
    Attributes []BackpackAttribute `json:",omitempty"`
    Equipped []EquippedInfoBackpackItem `json:",omitempty"`
    Quantity int
    Inventory uint32
}

type EquippedInfoBackpackItem struct {
    Class int `json:",omitempty"`
    Slot int `json:",omitempty"`
}

type SifItem struct {
    Id int64
    originalId int64
    Name string
    ProperName bool `json:",omitempty"`
    Description string `json:",omitempty"`
    TypeName string
    ImageUrl string
    DefIndex int64
    Level int
    Qualities string
    TagsClass string `json:",omitempty"`
    Origin string `json:",omitempty"`
    NotTradable bool `json:",omitempty"`
    NotCraftable bool `json:",omitempty"`
    NewName string `json:",omitempty"`
    NewDescription string `json:",omitempty"`
    Series float64 `json:",omitempty"`
    CraftNumber float64 `json:",omitempty"`
    CraftedById int `json:",omitempty"`
    CraftedByPersonaName string `json:",omitempty"`
    GiftedDate float64 `json:",omitempty"`
    GiftedById int `json:",omitempty"`
    GiftedByPersonaName string `json:",omitempty"`
    PaintCan string `json:",omitempty"`
    PaintedRed string `json:",omitempty"`
    PaintedBlu string `json:",omitempty"`
    EquippedBy map[string]bool `json:",omitempty"`
    HolidayRestriction string `json:",omitempty"`
    Position int64
}

type SchemaItem struct {
    DefIndex int `json:"defindex"`
    Name string `json:"item_name"`
    ProperName bool `json:"proper_name"`
    TypeName string `json:"item_type_name,omitempty"`
    Slot string `json:"item_slot,omitempty"`
    UsedBy []string `json:"used_by_classes,omitempty"`
    filter string
    eTag string
    hidden bool
    Description string `json:"item_description,omitempty"`
    ImageUrl string `json:"image_url"`
    Quality int `json:"item_quality"`
    HolidayRestriction string `json:"holiday_restriction,omitempty"`
    Attributes []SchemaItemAttribute `json:",omitempty"`
}

type BackpackAttribute struct {
    DefIndex int64 `json:",omitempty"`
    FloatValue float64 `json:"float_value,omitempty"`
    Value interface{} `json:",omitempty"`
    AccountInfo BackpackAttributeAccountInfo `json:"account_info,omitempty"`
}

type BackpackAttributeAccountInfo struct {
    SteamId int `json:",omitempty"`
    PersonaName string `json:",omitempty"`
}
type SchemaItemAttribute struct {
    Class string `json:",omitempty"`
}

type Configuration struct {
    Mysql string
    RootUrl string
    MaxOpenConns int
    ListenPort string
    ApiKey string
    ApiUrl string
}

type Schema struct {
    Items map[string]SchemaItem
    Attributes map[string]SchemaAttribute
    OriginTypes map[string]string
}

var db *sql.DB;
var schema Schema

func loadSchema() {
    fmt.Println("loading schema 440")
    var url string = apiurl + "/IEconItems_440/GetSchema/v0001/?key="+apikey+"&format=json&language=en"
    fmt.Println("GET " + url)
    resp, err := http.Get(url)
    if err != nil {
        panic(err)
        // handle error
    }
    defer resp.Body.Close()

    header := resp.Header
    if (resp.StatusCode != 200) {
        fmt.Println(url, resp.Status)
        panic(resp)
    }

    lastModified := header.Get("Last-Modified")
    fmt.Println("last modified:", lastModified)

    hB := []byte(lastModified) // transform to byte array
    hS := fmt.Sprint(adler32.Checksum(hB)) // checksum and cast to string

    body, err := ioutil.ReadAll(resp.Body)

    var f map[string]SchemaResult
    //var f interface{}
    err = json.Unmarshal(body, &f)
    if err != nil {
        panic(err)
        // handle error
    }

    var index string
    schema.Items = make(map[string]SchemaItem)
    schema.Attributes = make(map[string]SchemaAttribute)
    schema.OriginTypes = make(map[string]string)

    for _, i := range f["result"].Items {
        // The first 32 items listed describe the original class weapons
        if i.DefIndex <= 32 {
            i.hidden = true
        }
        if i.ImageUrl == "" {
            i.hidden = true
        }
        if i.Name == "Mann Co. Supply Crate Key" && i.DefIndex != 5021 {
            i.hidden = true
            // FIXME i.altId = 5021 (?) serie (?)
        }
        if i.Name == "Mann Co. Supply Crate" && i.DefIndex != 5022 {
            i.hidden = true
        }
        if i.Name == "Mann Co. Supply Munition" && i.DefIndex != 5734 {
            i.hidden = true
        }
        i.filter = strings.ToLower(i.Name + " " + i.TypeName)
        index = strconv.Itoa(i.DefIndex)
        i.eTag = "SI" + hS + "F" + index
        schema.Items[index] = i

    }

    for _, a := range f["result"].Attributes {
        index = strconv.Itoa(a.DefIndex)
        schema.Attributes[index] = a
    }

    for _, o := range f["result"].OriginNames {
        index = strconv.Itoa(o.Origin)
        schema.OriginTypes[index] = o.Name
    }

    fmt.Println("schema 440 loaded. item count:", len(schema.Items), "attributes count:", len(schema.Attributes) )
}

func main() {
    configurationContent, err := ioutil.ReadFile("/etc/steam-item-finder/config.json")
    if err != nil {
        fmt.Println("error while reading configuration file")
        panic(err.Error())
    }

    // parse configuration
    var c Configuration
    err = json.Unmarshal(configurationContent, &c)
    if err != nil {
        fmt.Println("error while parsing configuration file")
        panic(err.Error())
    }

    apikey = c.ApiKey
    apiurl = c.ApiUrl
    rootUrl = c.RootUrl
    fmt.Println("lauching api on " + rootUrl + " with steam api key " + apikey + ". api url is " + apiurl)
    loadSchema()

    // open mysql
    db, err = sql.Open("mysql", c.Mysql)
    if err != nil {
        panic(err.Error())
    }
    defer db.Close()

    db.SetMaxOpenConns(c.MaxOpenConns)
    // TODO: prepare

    m := martini.New()

    r := martini.NewRouter()
    r.NotFound(func(w http.ResponseWriter, r *http.Request) {
        // Only rewrite paths *not* containing filenames // that's quite tricky
        if !strings.HasPrefix(r.URL.Path, "/steam/") {
            http.ServeFile(w, r, "public/index.html")
        } else {
            w.WriteHeader(http.StatusNotFound)
            w.Write([]byte("404 page not found :("))
        }
    })

    r.Get("/steam/login", loginSteam)
    r.Get("/steam/logout", logoutSteam)
    r.Get("/steam/440/items", getItems)
    r.Get("/steam/440/item/:itemid", getItem)
    r.Get("/steam/general/user", getUserInfo)
    r.Get("/steam/general/user/:steamid", getUserInfo)
    r.Get("/steam/general/user/:steamid/friends", getFriends)
    r.Get("/steam/440/user/:steamid/backpack", getBackpack)

    m.Use(func(c martini.Context, w http.ResponseWriter, r *http.Request) {
        if strings.HasPrefix(r.URL.Path, "/steam/") {
            w.Header().Set("Content-Type", "application/json")
        }
    })

    m.Use(Logger())
    m.Use(martini.Recovery())
    m.Use(martini.Static("public"))
    m.MapTo(r, (*martini.Routes)(nil))

    m.Action(r.Handle)

    var listen string = "0.0.0.0:" + c.ListenPort
    fmt.Println("[martini] http listening on", listen)
    http.ListenAndServe(listen, m)
}

func loginSteam(res http.ResponseWriter, req *http.Request) (int, string) {
    //var referer string = req.Referer() // url after # is not passed
    var openidCheck string = req.FormValue("openid.mode")
    if openidCheck != "id_res" {
        url, err := openid.RedirectURL("http://steamcommunity.com/openid/", rootUrl + "/steam/login", rootUrl)
        if err != nil {
            panic(err)
        }
        res.Header().Set("location", url)
        return 302, url
    } else {
        id, err := openid.Verify(rootUrl + req.RequestURI, discoveryCache, nonceStore)
        if err != nil {
            panic(err)
            return 403, "something went wrong"
        }
        var identity string = strings.Split(id, "/")[5]

        expiration := time.Now().Add(3650 * 24 * time.Hour)

        hash, err := loginUser(identity)
        if err != nil {
            panic(err)
            return 500, ""
        }
        cookie := http.Cookie{Name: "sif_session", Value: hash, Expires: expiration}
        http.SetCookie(res, &cookie)
        res.Header().Set("location", rootUrl)

        return 302, rootUrl
    }
}

func logoutSteam(res http.ResponseWriter, req *http.Request) (int, string) {
    cookie, err := req.Cookie("sif_session")
    if err != nil {
        res.Header().Set("location", rootUrl)
        return 302, "you were not logged!?"
    }

    var hash string = cookie.Value

    // delete session in database
    fmt.Println("[sif] set expire for hash", hash)
    t := time.Now()
    var timestamp int64 = t.Unix()
    var query string = "UPDATE `sif`.`sessions` SET `expired` = ? WHERE `sessions`.`hash` = ?"
    _, err = db.Exec(query, timestamp, hash)
    if err != nil {
        panic(err)
    }

    // delete cookie
    expiration := time.Now().Add(-1)
    cookieHttp := http.Cookie{Name: "sif_session", Value: hash, Expires: expiration, MaxAge: -1}
    http.SetCookie(res, &cookieHttp)
    res.Header().Set("location", rootUrl)
    return 302, rootUrl
}

func loginUser(steamid string) (string, error) {
    var err error

    var query string = "INSERT IGNORE INTO `sif`.`users` (`steamid`) VALUES (?)"
    fmt.Println("[mysql]", query, steamid)
    _, err = db.Exec(query, steamid)
    if err != nil {
        return "", err
    }

    t := time.Now()
    var salt string = "pouuuuuuuuet7" + steamid + t.String()
    s := []byte(salt) // transform to byte array
    h := sha512.Sum512(s)
    hash := fmt.Sprintf("%x", h)
    query = "INSERT INTO `sif`.`sessions` (`hash`, `steamid`, `creation_timestamp`, `ip`, `useragent`) VALUES (?, ?, NULL, NULL, NULL);"
    fmt.Println("[mysql]", query, hash, steamid)
    _, err = db.Exec(query, hash, steamid)
    if err != nil {
        return "", err
    }

    return hash, err
}

func getItems(_ http.ResponseWriter, req *http.Request) (int, string) {
    // args
    var query string = req.FormValue("query")
    query = strings.ToLower(query)

    list := make(map[string]SchemaItem)
    for k, v := range schema.Items {
        if strings.Contains(v.filter, query) && !v.hidden {
            list[k] = v
        }
    }

    if len(list) == 0 {
        return 204, ""
    }

    b, err := json.Marshal(list)
    if err != nil {
        panic(err)
    }

    return 200, string(b)
}

func getItem(res http.ResponseWriter, req *http.Request, params martini.Params) (int, string) {
    // args
    var itemid string = params["itemid"]

    //var item SchemaItem
    if item, ok := schema.Items[itemid]; ok {
        item = schema.Items[itemid]
        var etag string = item.eTag

        // handle ETag
        if strings.Replace(req.Header.Get("If-None-Match"), "\"", "", -1) == etag {
            return 304, ""
        }

        res.Header().Set("etag", etag)
        b, err := json.Marshal(item)
        if err != nil {
            panic(err)
            // handle error
            return 500, "error"
        }

        return 200, string(b)
    } else {
        return 404, ""
    }
}

func getFriends(_ http.ResponseWriter, req *http.Request, params martini.Params) (int, string) {
    // args
    var steamid string = params["steamid"]
    var url = apiurl + "/ISteamUser/GetFriendList/v1?steamid="+steamid+"&relationship=all&key="+apikey+"&format=json"
    fmt.Println("GET", url)
    resp, err := http.Get(url)
    if err != nil {
        panic(err)
    }
    defer resp.Body.Close()
    body, err := ioutil.ReadAll(resp.Body)

    var f map[string]map[string][]Friend
    err = json.Unmarshal(body, &f)
    if err != nil {
        panic(err)
    }

    friendslist := f["friendslist"]["friends"]
    fmt.Printf("%d friends\n", len(friendslist))

    b, err := json.Marshal(friendslist)

    return 200, string(b)
}

func getUserInfo(_ http.ResponseWriter, req *http.Request, params martini.Params) (int, string) {
    // args
    var steamid string = params["steamid"]
    if steamid == "" {
        cookie, err := req.Cookie("sif_session")
        if err != nil {
            return 204, ""
        }

        // TODO: cache in memory FIXME!
        var hash string = cookie.Value
        var query string = "SELECT steamid FROM sessions WHERE hash = ? AND expired IS NULL LIMIT 1"
        fmt.Println("[sif] looking for this hash in session database", hash)
        fmt.Println("[mysql]", query, hash)
        err = db.QueryRow(query, hash).Scan(&steamid)

        if err == sql.ErrNoRows {
            return 402, "something is wrong with you session id, please sign again in through steam"
        }

        if err != nil {
            panic(err)
        }

        fmt.Println("[sif] updating last activity for", steamid)
        t := time.Now()
        var timestamp int64 = t.Unix()
        query = "UPDATE `sif`.`users` SET `last_activity` = ? WHERE `users`.`steamid` = ?"
        _, err = db.Exec(query, timestamp, steamid)

        if err != nil {
            panic(err)
        }

    }

    var user User = loadPlayerInfo(steamid)
    b, err := json.Marshal(user)
    if err != nil {
        panic(err)
    }

    return 200, string(b)
}

func getBackpack(_ http.ResponseWriter, req *http.Request, params martini.Params) (int, string) {
    // args
    var itemFilter string = req.FormValue("item")
    var serieFilter string = req.FormValue("serie")
    var idFilter string = req.FormValue("id")
    var filter int64

    if itemFilter != "" {
        filter, _ = strconv.ParseInt(itemFilter, 10, 64)
    } else if serieFilter != "" {
        filter, _ = strconv.ParseInt(serieFilter, 10, 64)
    } else if idFilter != "" {
        filter, _ = strconv.ParseInt(idFilter, 10, 64)
    }

    var steamid string = params["steamid"]

    var url string = apiurl + "/IEconItems_440/GetPlayerItems/v0001/?steamid="+steamid+"&key="+apikey+"&format=json"
    fmt.Println("GET", url)
    resp, err := http.Get(url)
    if err != nil {
        panic(err)
        // handle error
    }
    defer resp.Body.Close()
    body, err := ioutil.ReadAll(resp.Body)

    var f map[string]BackpackResult
    err = json.Unmarshal(body, &f)
    if err != nil {
        panic(err)
        // handle error
    }

    bp := f["result"]
    var list []BackpackItem
    items := bp.Items

    if filter != 0 {
        if itemFilter != "" {
            for _, v := range items {
                if v.DefIndex == filter {
                    list = append(list, v)
                }
            }
        } else if serieFilter != "" {
            for _, v := range items {
                for _, a := range v.Attributes {
                    if a.DefIndex  == filter {
                        list = append(list, v)
                    }
                }
            }
        } else if idFilter != "" {
            for _, v := range items {
                if v.Id == filter {
                    list = append(list, v)
                }
            }
        }
    } else {
        // full backpack
        list = bp.Items
    }

    var sifBp SifBackpack

    for _, playerItem := range list {
        var item SifItem
        var schemaItem SchemaItem
        var index string = strconv.FormatInt(playerItem.DefIndex, 10)
        schemaItem = schema.Items[index]

        item.Id = playerItem.Id
        item.originalId = playerItem.OriginalId
        item.Name = schemaItem.Name
        item.ProperName = schemaItem.ProperName
        item.Description = schemaItem.Description
        item.ImageUrl = schemaItem.ImageUrl
        item.TypeName = schemaItem.TypeName
        item.DefIndex = playerItem.DefIndex
        item.Level = playerItem.Level
        item.Qualities = fmt.Sprintf("quality%d", playerItem.Quality)
        if playerItem.NotTradable {
            item.Qualities = item.Qualities + " untradable"
        }
        if playerItem.NotCraftable {
            item.Qualities = item.Qualities + " uncraftable"
        }

        if playerItem.Id == playerItem.OriginalId {
            var originIndex string = strconv.FormatInt(playerItem.Origin, 10)
            item.Origin = schema.OriginTypes[originIndex]
            if item.Origin == "Traded" {
                item.Origin = ""
            }
        }
        item.NotTradable = playerItem.NotTradable
        item.NotCraftable = playerItem.NotCraftable

        for _, attr := range playerItem.Attributes {
            var attrIndex string = strconv.FormatInt(attr.DefIndex, 10)
            schemaAttr := schema.Attributes[attrIndex]

            // crafted by
            if attr.DefIndex == 228 {
                if attr.AccountInfo.SteamId != 0 {
                    item.CraftedById = attr.AccountInfo.SteamId
                    item.CraftedByPersonaName = attr.AccountInfo.PersonaName
                }
            }

            // craft number
            if attr.DefIndex == 229 {
                item.CraftNumber = attr.Value.(float64)
            }

            // gifted by
            if attr.DefIndex == 186 {
                if attr.AccountInfo.SteamId != 0 {
                    item.GiftedById = attr.AccountInfo.SteamId
                    item.GiftedByPersonaName = attr.AccountInfo.PersonaName
                }
            }

            // gift date
            if attr.DefIndex == 185 {
                item.GiftedDate = attr.Value.(float64)
            }

            // crate series
            if attr.DefIndex == 187 {
                item.Series = attr.FloatValue
            }

            // new name
            if attr.DefIndex == 500 {
                item.NewName = attr.Value.(string)
            }

            // new description
            if attr.DefIndex == 501 {
                item.NewDescription = attr.Value.(string)
            }

            // set_item_tint_rgb
            if attr.DefIndex == 142 {
                item.PaintCan = "true"
                //b := []byte(attr.FloatValue)
                //fmt.Println(hex.Encode(b))
                // TODO color
            }

            // XXX
            fmt.Printf("\tattr: %+v\n", attr)
            fmt.Printf("\tschema attr: %+v\n\n\n", schemaAttr)
        }

        if item.Origin != "" {
            item.TagsClass = "fa-certificate"
        } else if item.NewName != "" && item.NewDescription != "" {
            item.TagsClass = "fa-tags"
        } else if item.NewName != "" || item.NewDescription != "" {
            item.TagsClass = "fa-tag"
        }
        item.HolidayRestriction = schemaItem.HolidayRestriction

        // bp position
        item.Position = parseInventory(playerItem.Inventory)

        // DOING
        fmt.Printf("playerItem %+v\n", playerItem)
        sifBp.Items = append(sifBp.Items, &item)
    }

    if(bp.Status == 8) {
        return 404, "invalid steamid"
    }

    if(bp.Status == 18) {
        return 404, "this steamid does not exist"
    }

    if(bp.Status == 18) {
        return 410, "private backpack"
    }

    fmt.Printf("%d items\n", len(sifBp.Items))
    fmt.Printf("%d status\n", bp.Status)

    sifBp.Player = loadPlayerInfo(steamid)
    sifBp.Status = bp.Status
    sifBp.Size = bp.Size

    var b []byte
    if filter != 0 {
        if idFilter != "" {
            var sifBpItem SifBackpackItem
            sifBpItem.Status = bp.Status
            sifBpItem.Player = sifBp.Player
            sifBpItem.Item = sifBp.Items[0]
            b, err = json.Marshal(sifBpItem)
        } else {
            b, err = json.Marshal(sifBp)
        }
    } else {
        // a map of new items (position 0)
        var newItems []*SifItem
        // a map for the item in the inventory
        var invItems map[int64]SifItem
        invItems = make(map[int64]SifItem)

        var stats SifBackpackStats
        var metal float64

        for _, item := range sifBp.Items {
            if item.originalId == item.Id {
                stats.Untouched++
            }

            if item.DefIndex == 5021 {
                stats.Keys++
            }

            if item.DefIndex == 5000 {
                metal = metal + 0.11
            }

            if item.DefIndex == 5001 {
                metal = metal + 0.33
            }

            if item.DefIndex == 5002 {
                metal = metal + 1
            }

            if item.Position == 0 {
                newItems = append(newItems, item)
            } else {
                invItems[item.Position] = *item
            }
        }

        stats.Total = len(sifBp.Items)

        stats.RawMetal = fmt.Sprintf("%.2f", metal)
        var sifFullBp SifFullBackpack
        sifFullBp.Player = sifBp.Player
        sifFullBp.Status = bp.Status
        sifFullBp.Size = bp.Size
        sifFullBp.Stats = stats

        var page []*SifItem
        // each new item is in page 0
        for _, item := range newItems {
            page = append(page, item)
        }
        sifFullBp.PagesOfItems = append(sifFullBp.PagesOfItems, page)

        var p int64
        page = nil
        // fulling the full backpack starting from page 1
        for p = 1; p <= bp.Size; p++ {
            if item, ok := invItems[p]; ok {
                page = append(page, &item)
            } else {
                page = append(page, nil)
            }

            // 50 items per page
            if(p%50 == 0) {
                sifFullBp.PagesOfItems = append(sifFullBp.PagesOfItems, page)
                page = nil
            }
        }

        b, err = json.Marshal(sifFullBp)
    }

    if err != nil {
        panic(err)
    }

    return 200, string(b)
}

func parseInventory(inv uint32) (int64) {
    b := fmt.Sprintf("%b", inv)
    a := strings.Split(b, "")

    var position string

    for k, v := range a {
        switch k {
            case 0:
                if v != "1" {
                    panic(inv)
                }
            case 1:
                if v == "1" {
                    return 0
                }
            case 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31:
                position = fmt.Sprintf("%s%s", position, v)
        }
    }

    d, err := strconv.ParseInt(position, 2, 64)
    if err != nil {
        panic(err)
    }
    return d
}

func loadPlayerInfo(steamid string) User {
    // TODO: cache in database
    // if cache present, take it. if not, create it
    var url string = apiurl + "/ISteamUser/GetPlayerSummaries/v0002?steamids="+steamid+"&key="+apikey+"&format=json"
    fmt.Println("GET", url)
    resp, err := http.Get(url)
    if err != nil {
        panic(err)
        // handle error
    }
    defer resp.Body.Close()
    body, err := ioutil.ReadAll(resp.Body)

    var f map[string]map[string][]User
    err = json.Unmarshal(body, &f)
    if err != nil {
        panic(err)
        // handle error
    }

    user := f["response"]["players"][0]
    return user
}


// Logger returns a middleware handler that logs the request as it goes in and the response as it goes out.
func Logger() martini.Handler {
    return func(res http.ResponseWriter, req *http.Request, c martini.Context, log *log.Logger) {
        start := time.Now()

        addr := req.Header.Get("X-Real-IP")
        if addr == "" {
            addr = req.Header.Get("X-Forwarded-For")
            if addr == "" {
                addr = req.RemoteAddr
            }
        }

        log.Printf("%s [%d] Started %s %s for %s", start.Format(time.RFC3339), start.Nanosecond(), req.Method, req.URL.Path, addr)

        rw := res.(martini.ResponseWriter)
        c.Next()

        stop := time.Now()
        log.Printf("%s [%d] Completed %v %s in %v\n", stop.Format(time.RFC3339), start.Nanosecond(), rw.Status(), http.StatusText(rw.Status()), time.Since(start))
    }
}

