package game

import (
	"math"
	"math/rand"
	"strconv"
)

type Frame struct {
	Action   string
	InputId  string
	State    bool
	Angle    float64
	Playerid int // needed for cspGame
}

type SendPlayer struct {
	X      float64
	Y      float64
	Number string
	Health int
}

type SendBullet struct {
	X float64
	Y float64
}

type Positions struct {
	Action string
	PList  []SendPlayer
	BList  []SendBullet
}

type MsgToServer struct {
	Action string
	Data   string
}
type GenericResponse struct {
	Action string
	Data   string
}

var XMin = float64(0)
var XMax = float64(490)
var YMin = float64(0)
var YMax = float64(490)

func (g *Game) updateAllPlayers() []SendPlayer {
	g.playerListMutex.Lock()
	defer g.playerListMutex.Unlock()

	pack := make([]SendPlayer, 0)
	for _, p := range g.playerList {
		g.updatePlayer(p)
		pack = append(pack, SendPlayer{p.e.x, p.e.y, p.number, p.health})
	}
	return pack
}

func (g *Game) updateAllBullets() []SendBullet {
	g.playerListMutex.Lock()
	defer g.playerListMutex.Unlock()

	pack := make([]SendBullet, 0)
	for _, b := range g.bulletList {
		g.updateBullet(b)
		if b.toRemove == true {
			delete(g.bulletList, strconv.Itoa(b.id))
		} else {
			pack = append(pack, SendBullet{b.e.x, b.e.y})
		}
	}
	return pack
}

type Entity struct {
	x        float64
	y        float64
	spdX     float64
	spdY     float64
	id       int
	isPlayer bool
}

func (g *Game) newEntity(x, y float64, isPlayer bool) *Entity {
	return &Entity{x: x, y: y, spdX: 0, spdY: 0, id: g.id, isPlayer: isPlayer}
}

func (g *Game) updateEntity(e *Entity) {
	e.x += e.spdX
	e.y += e.spdY

	if e.isPlayer == false {
		return
	}

	if e.x < XMin {
		e.x = XMin
	} else if e.x > XMax {
		e.x = XMax
	}

	if e.y < YMin {
		e.y = YMin
	} else if e.y > YMax {
		e.y = YMax
	}

}

func (g *Game) getDistance(e *Entity, p *Entity) float64 {
	return math.Sqrt(math.Pow(e.x-p.x, 2) + math.Pow(e.y-p.y, 2))
}

type Player struct {
	e              *Entity
	id             int
	number         string
	pressingRight  bool
	pressingLeft   bool
	pressingUp     bool
	pressingDown   bool
	pressingAttack bool
	mouseAngle     float64
	maxSpd         float64
	health         int
}

func (g *Game) newPlayer(id int, number string) *Player {
	p := &Player{e: g.newEntity(250, 250, true)}
	p.id = id
	p.number = number
	p.pressingRight = false
	p.pressingLeft = false
	p.pressingUp = false
	p.pressingDown = false
	p.pressingAttack = false
	p.mouseAngle = 0
	p.maxSpd = 10
	p.health = 25

	g.playerList[p.id] = p
	return p
}

func (g *Game) updatePlayer(p *Player) {
	//update player speed
	if p.pressingRight {
		p.e.spdX = p.maxSpd
	} else if p.pressingLeft {
		p.e.spdX = -p.maxSpd
	} else {
		p.e.spdX = 0
	}

	if p.pressingUp {
		p.e.spdY = -p.maxSpd
	} else if p.pressingDown {
		p.e.spdY = p.maxSpd
	} else {
		p.e.spdY = 0
	}

	//super_update --> entity update
	g.updateEntity(p.e)

	if p.pressingAttack {
		//shootBullet --> adds to bulletList
		g.newBullet(p.id, p.e, p.mouseAngle)
		//b.x = p.x -->handled by newBullet
		//b.y = p.y
	}
	p.pressingAttack = false
	//p.pressingDown = false
	//p.pressingLeft = false
	//p.pressingRight = false
	//p.pressingUp = false

}

type Bullet struct {
	e        *Entity
	id       int
	parent   int
	timer    int
	toRemove bool
}

func (g *Game) newBullet(parent int, e *Entity, angle float64) *Bullet {
	b := &Bullet{e: g.newEntity(e.x, e.y, false)}
	b.id = rand.Intn(1000000)
	b.e.spdX = math.Cos(float64(angle/180*math.Pi)) * 10
	b.e.spdY = math.Sin(float64(angle/180*math.Pi)) * 10
	b.parent = parent
	b.timer = 0
	b.toRemove = false

	g.bulletList[strconv.Itoa(b.id)] = b
	return b
}

func (g *Game) updateBullet(b *Bullet) {
	b.timer++
	if b.timer > 100 {
		b.toRemove = true
	}
	g.updateEntity(b.e) //superUpdate

	for _, p := range g.playerList {
		if g.getDistance(b.e, p.e) < 20 && b.parent != p.id {
			p.health -= 1
			b.toRemove = true
		}
	}
}
