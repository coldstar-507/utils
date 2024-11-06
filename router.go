package utils

import (
	"encoding/json"
	"log"
	"net/http"
	"sort"
	"strconv"
	"time"
)

// type node_place = uint16
// type media_place = uint16
// type chat_place = uint16

func InitLocalRouter(ip string, st SERVER_TYPE, place uint16) {
	LocalRouter = &server{
		IP:         ip,
		ServerType: st,
		Place:      place,
		RelMedias:  make([]uint16, 0, 10),
		RelNodes:   make([]uint16, 0, 10),
		RelChats:   make([]uint16, 0, 10),
	}
}

var LocalRouter Server

func NodeRouter() Router {
	return routers[NODE_ROUTER]
}

func MediaRouter() Router {
	return routers[MEDIA_ROUTER]
}

func ChatRouter() Router {
	return routers[CHAT_ROUTER]
}

func HandlePing(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("pong\n"))
}

func HandleScoreRequest(w http.ResponseWriter, r *http.Request) {
	if b, err := json.Marshal(LocalRouter.Scores()); err != nil {
		log.Println("HandleScoreRequest error encoding score:", err)
		w.WriteHeader(500)
	} else if _, err := w.Write(b); err != nil {
		log.Println("HandleScoreRequest error writing score:", err)
		w.WriteHeader(501)
	}
}

func HandleAddServer(w http.ResponseWriter, r *http.Request) {
	var srv *server
	if err := json.NewDecoder(r.Body).Decode(srv); err != nil {
		log.Println("HandleAddServer decoding server:", err)
		w.WriteHeader(500)
	} else {
		if routers[srv.ServerType].servers[srv.Place] == nil {
			routers[srv.ServerType].servers[srv.Place] = srv
		}
	}
}

type Router interface {
	Host(place uint16) string
	Port() int
	HostAndPort(place uint16) string
	GetServer(place uint16) Server
	RelativeMedias(place uint16) []uint16
	RelativeNodes(place uint16) []uint16
	RelativeChats(place uint16) []uint16
}

type Server interface {
	Run()
	Scores() *scores
	RelativeMedias() []uint16
	RelativeNodes() []uint16
	RelativeChats() []uint16
}

func (lr *server) Scores() *scores {
	return &scores{
		Medias: lr.RelMedias,
		Nodes:  lr.RelNodes,
		Chats:  lr.RelChats,
	}
}

func (lr *server) Run() {
	tck := time.NewTicker(time.Minute * 1)
	tck2 := time.NewTicker(time.Minute * 2)
	calcChats := func() {
		log.Println("LocalRouter: calculating chat routes")
		lr.RelChats = calculateRoutes(routers[CHAT_ROUTER])
	}
	calcNodes := func() {
		log.Println("LocalRouter: calculating node routes")
		lr.RelNodes = calculateRoutes(routers[NODE_ROUTER])
	}
	calcMedias := func() {
		log.Println("LocalRouter: calculating media routes")
		lr.RelMedias = calculateRoutes(routers[MEDIA_ROUTER])
	}

	calcAll := func() {
		go calcChats()
		go calcNodes()
		go calcMedias()
	}

	updateChatRelativeRoutes := func() {
		log.Println("LocalRouter: Run(): updateChatRelativeRoutes()")
		cs := routers[CHAT_ROUTER]
		futureScores := make(map[uint16]<-chan *scores)
		for p, _ := range cs.servers {
			futureScores[p] = cs.fetchScores(p)
		}
		for p, s := range futureScores {
			if sc := <-s; sc != nil {
				cs.servers[p].RelChats = sc.Chats
			}
		}
	}

	updateNodeRelativeRoutes := func() {
		log.Println("LocalRouter: Run(): updateNodeRelativeRoutes()")
		ns := routers[NODE_ROUTER]
		futureScores := make(map[uint16]<-chan *scores)
		for p, _ := range ns.servers {
			futureScores[p] = ns.fetchScores(p)
		}
		for p, s := range futureScores {
			if sc := <-s; sc != nil {
				ns.servers[p].RelChats = sc.Chats
			}
		}
	}

	updateMediaRelativeRoutes := func() {
		log.Println("LocalRouter: Run(): updateMediaRelativeRoutes()")
		ms := routers[MEDIA_ROUTER]
		futureScores := make(map[uint16]<-chan *scores)
		for p, _ := range ms.servers {
			futureScores[p] = ms.fetchScores(p)
		}
		for p, s := range futureScores {
			if sc := <-s; sc != nil {
				ms.servers[p].RelChats = sc.Chats
			}
		}
	}

	for {
		select {
		case <-tck.C:
			log.Println("LocalRouter: periodic routes calculation")
			calcAll()
		case <-tck2.C:
			log.Println("LocalRouter: periodic relative routes calculation")
			go updateChatRelativeRoutes()
			go updateNodeRelativeRoutes()
			go updateMediaRelativeRoutes()
		}
	}
}

func calculateRoutes(r *router) []uint16 {
	type score struct {
		id    uint16
		score int64
	}

	var futureScores = make(map[uint16]<-chan *int64)
	for p, _ := range r.servers {
		futureScores[p] = r.Ping(p)
	}

	scores := make([]*score, 0, len(r.servers))
	for x, v := range futureScores {
		if s := <-v; s != nil {
			scores = append(scores, &score{id: x, score: *s})
		}
	}

	sort.Slice(scores, func(i, j int) bool {
		return scores[i].score < scores[j].score
	})

	res := make([]uint16, len(scores))
	for i, x := range scores {
		res[i] = x.id
	}

	return res
}

type router struct {
	port    int
	servers map[uint16]*server
}

func (r *router) Host(place uint16) string {
	return r.servers[place].IP
}

func (r *router) Port() int {
	return r.port
}

func (r *router) HostAndPort(place uint16) string {
	return r.servers[place].IP + ":" + strconv.Itoa(r.port)
}

func (r *router) GetServer(place uint16) Server {
	return r.servers[place]
}

func (r *router) RelativeMedias(place uint16) []uint16 {
	return r.GetServer(place).RelativeMedias()

}

func (r *router) RelativeNodes(place uint16) []uint16 {
	return r.GetServer(place).RelativeNodes()
}

func (r *router) RelativeChats(place uint16) []uint16 {
	return r.GetServer(place).RelativeChats()
}

func (s *server) RelativeMedias() []uint16 {
	return s.RelMedias
}

func (s *server) RelativeNodes() []uint16 {
	return s.RelNodes
}

func (s *server) RelativeChats() []uint16 {
	return s.RelChats
}

func (r *router) Ping(place uint16) <-chan *int64 {
	ch := make(chan *int64)
	go func() {
		addr := "http://" + r.HostAndPort(place) + "/ping"
		req, err := http.NewRequest("GET", addr, nil)
		t1 := time.Now()
		if err != nil {
			log.Printf("Ping(%d) error creating req: %v", place, err)
			ch <- nil
		} else if res, err := http.DefaultClient.Do(req); err != nil {
			log.Printf("Ping(%d) error sending req: %v", place, err)
			ch <- nil
		} else if res.StatusCode != 200 {
			log.Printf("Ping(%d) error: code isn't 200: %v", place, err)
			ch <- nil
		} else {
			ms := time.Now().Sub(t1).Milliseconds()
			log.Printf("Ping(%d) took %d ms", place, ms)
			ch <- &ms
		}
		close(ch)
	}()
	return ch
}

type scores struct {
	Medias []uint16 `json:"mediaPlaces"`
	Nodes  []uint16 `json:"nodePlaces"`
	Chats  []uint16 `json:"chatPlaces"`
}

func (r *router) fetchScores(place uint16) <-chan *scores {
	addr := "http://" + r.HostAndPort(place) + "/route-scores"
	ch := make(chan *scores)
	go func() {
		var sc scores
		if req, err := http.NewRequest("GET", addr, nil); err != nil {
			log.Println("fetchScores error making req:", err)
			ch <- nil
		} else if res, err := http.DefaultClient.Do(req); err != nil {
			log.Println("fetchScores error doing req:", err)
			ch <- nil
		} else if res.StatusCode != 200 {
			log.Println("fetchScores status != 200")
			ch <- nil
		} else if err = json.NewDecoder(res.Body).Decode(&sc); err != nil {
			log.Println("fetchScores error decoding res json:", err)
			ch <- nil
		} else {
			ch <- &sc
		}
		close(ch)
	}()
	return ch
}

type SERVER_TYPE = byte

const (
	NODE_ROUTER SERVER_TYPE = iota
	MEDIA_ROUTER
	CHAT_ROUTER
)

type server struct {
	ServerType SERVER_TYPE `json:"serverType"`
	Place      uint16      `json:"place"`
	IP         string      `json:"ip"`
	RelMedias  []uint16    `json:"relMedias"`
	RelNodes   []uint16    `json:"relNodes"`
	RelChats   []uint16    `json:"relChats"`
}

// this should be fetch from a server with this purpose only
var routers = map[SERVER_TYPE]*router{
	NODE_ROUTER: &router{
		port: 8083,
		servers: map[uint16]*server{
			0x0000: {
				ServerType: NODE_ROUTER,
				Place:      0x0000,
				IP:         "localhost",
			},
		},
	},

	MEDIA_ROUTER: &router{
		port: 8081,
		servers: map[uint16]*server{
			0x0100: {
				ServerType: MEDIA_ROUTER,
				Place:      0x0100,
				IP:         "localhost",
			},
		},
	},

	CHAT_ROUTER: &router{
		port: 8082,
		servers: map[uint16]*server{
			0x1000: {
				ServerType: CHAT_ROUTER,
				Place:      0x1000,
				IP:         "localhost",
			},
		},
	},
}
