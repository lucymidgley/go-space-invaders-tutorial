package main

import (
	"embed"
	"game/components"
	"image"
	_ "image/png"
	"io/fs"
	"math"
	"math/rand"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
)

const (
	ScreenWidth  = 800
	ScreenHeight = 600
)

//go:embed assets/*
var assets embed.FS

type Rect struct {
	X      float64
	Y      float64
	Width  float64
	Height float64
}

func NewRect(x, y, width, height float64) Rect {
	return Rect{
		X:      x,
		Y:      y,
		Width:  width,
		Height: height,
	}
}

func (r Rect) MaxX() float64 {
	return r.X + r.Width
}

func (r Rect) MaxY() float64 {
	return r.Y + r.Height
}

func (r Rect) Intersects(other Rect) bool {
	return r.X <= other.MaxX() && other.X <= r.MaxX() && r.Y <= other.MaxY() && other.Y <= r.MaxY()
}

func (p *Player) Collider() Rect {
	bounds := p.sprite.Bounds()

	return NewRect(p.position.X, p.position.Y, float64(bounds.Dx()), float64(bounds.Dy()))
}
func (p *Meteor) Collider() Rect {
	bounds := p.sprite.Bounds()

	return NewRect(p.position.X, p.position.Y, float64(bounds.Dx()), float64(bounds.Dy()))
}
func (p *Bullet) Collider() Rect {
	bounds := p.sprite.Bounds()

	return NewRect(p.position.X, p.position.Y, float64(bounds.Dx()), float64(bounds.Dy()))
}

var MeteorSprites = mustLoadImages("assets/meteors/*.png")

type Meteor struct {
	position      Vector
	rotation      float64
	movement      Vector
	rotationSpeed float64
	sprite        *ebiten.Image
}

func NewMeteor() *Meteor {
	sprite := MeteorSprites[rand.Intn(len(MeteorSprites))]

	target := Vector{
		X: ScreenWidth / 2,
		Y: ScreenHeight / 2,
	}

	r := ScreenWidth / 2.0

	angle := rand.Float64() * 2 * math.Pi

	pos := Vector{
		X: target.X + math.Cos(angle)*r,
		Y: target.Y + math.Sin(angle)*r,
	}

	velocity := 0.25 + rand.Float64()*1.5

	direction := Vector{
		X: target.X - pos.X,
		Y: target.Y - pos.Y,
	}

	normalizedDirection := direction.Normalize()

	movement := Vector{
		X: normalizedDirection.X * velocity,
		Y: normalizedDirection.Y * velocity,
	}

	rotationSpeed := -0.02 + rand.Float64()*0.04

	return &Meteor{
		position:      pos,
		sprite:        sprite,
		movement:      movement,
		rotationSpeed: rotationSpeed,
	}
}

func (m *Meteor) Draw(screen *ebiten.Image) {
	bounds := m.sprite.Bounds()
	halfW := float64(bounds.Dx()) / 2
	halfH := float64(bounds.Dy()) / 2

	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(-halfW, -halfH)
	op.GeoM.Rotate(m.rotation)
	op.GeoM.Translate(halfW, halfH)

	op.GeoM.Translate(m.position.X, m.position.Y)

	screen.DrawImage(m.sprite, op)
}

func (m *Meteor) Update() {
	m.position.X += m.movement.X
	m.position.Y += m.movement.Y
	m.rotation += m.rotationSpeed
}

var PlayerSprite = mustLoadImage("assets/playerShip1_blue.png")

var BulletSprite = mustLoadImage("assets/laserBlue02.png")

type Bullet struct {
	position Vector
	rotation float64
	sprite   *ebiten.Image
}

func NewBullet(position Vector, rotation float64) *Bullet {
	sprite := BulletSprite

	return &Bullet{
		position: position,
		sprite:   sprite,
		rotation: rotation,
	}
}

func (b *Bullet) Draw(screen *ebiten.Image) {
	bounds := b.sprite.Bounds()
	halfW := float64(bounds.Dx()) / 2
	halfH := float64(bounds.Dy()) / 2

	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(-halfW, -halfH)
	op.GeoM.Rotate(b.rotation)
	op.GeoM.Translate(halfW, halfH)

	op.GeoM.Translate(b.position.X, b.position.Y)

	screen.DrawImage(b.sprite, op)
}

func (b *Bullet) Update() {
	speed := 350.0 / float64(ebiten.TPS())

	b.position.X += math.Sin(b.rotation) * speed
	b.position.Y += math.Cos(b.rotation) * -speed
}

type Player struct {
	game          *Game
	position      Vector
	rotation      float64
	sprite        *ebiten.Image
	shootCooldown *components.Timer
}

func NewPlayer(game *Game) *Player {
	sprite := PlayerSprite
	shootCooldown := components.NewTimer(time.Millisecond * 1000)

	bounds := sprite.Bounds()
	halfW := float64(bounds.Dx()) / 2
	halfH := float64(bounds.Dy()) / 2

	pos := Vector{
		X: ScreenWidth/2 - halfW,
		Y: ScreenHeight/2 - halfH,
	}
	return &Player{
		game:          game,
		position:      pos,
		sprite:        sprite,
		shootCooldown: shootCooldown,
	}
}

func (p *Player) Update() {
	speed := math.Pi / float64(ebiten.TPS())

	if ebiten.IsKeyPressed(ebiten.KeyLeft) {
		p.rotation -= speed
	}
	if ebiten.IsKeyPressed(ebiten.KeyRight) {
		p.rotation += speed
	}

	p.shootCooldown.Update()
	if p.shootCooldown.IsReady() && ebiten.IsKeyPressed(ebiten.KeySpace) {
		p.shootCooldown.Reset()

		bulletSpawnOffset := 50.0

		bounds := p.sprite.Bounds()
		halfW := float64(bounds.Dx()) / 2
		halfH := float64(bounds.Dy()) / 2

		spawnPos := Vector{
			X: p.position.X + halfW + math.Sin((p.rotation)*bulletSpawnOffset),
			Y: p.position.Y + halfH + math.Cos((p.rotation)*-bulletSpawnOffset),
		}
		bullet := NewBullet(spawnPos, p.rotation)
		p.game.AddBullet(bullet)
	}
}

func (p *Player) Draw(screen *ebiten.Image) {
	bounds := p.sprite.Bounds()
	halfW := float64(bounds.Dx()) / 2
	halfH := float64(bounds.Dy()) / 2

	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(-halfW, -halfH)
	op.GeoM.Rotate(p.rotation)
	op.GeoM.Translate(halfW, halfH)

	op.GeoM.Translate(p.position.X, p.position.Y)

	screen.DrawImage(p.sprite, op)
}

type Vector struct {
	X float64
	Y float64
}

func (v Vector) Normalize() Vector {
	magnitude := math.Sqrt(v.X*v.X + v.Y*v.Y)
	v.X = v.X / magnitude
	v.Y = v.Y / magnitude
	return v
}

type Game struct {
	player           *Player
	meteorSpawnTimer *components.Timer
	meteors          []*Meteor
	bullets          []*Bullet
}

func (g *Game) Update() error {
	g.player.Update()

	g.meteorSpawnTimer.Update()
	if g.meteorSpawnTimer.IsReady() {
		g.meteorSpawnTimer.Reset()

		m := NewMeteor()
		g.meteors = append(g.meteors, m)
	}

	for _, m := range g.meteors {
		m.Update()
	}

	for _, b := range g.bullets {
		b.Update()
	}

	for i, m := range g.meteors {
		for j, b := range g.bullets {
			if m.Collider().Intersects(b.Collider()) {
				g.meteors = append(g.meteors[:i], g.meteors[i+1:]...)
				g.bullets = append(g.bullets[:j], g.bullets[j+1:]...)
			}
		}
	}

	for _, m := range g.meteors {
		if m.Collider().Intersects(g.player.Collider()) {
			g.Reset()
		}
	}

	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	g.player.Draw(screen)

	for _, m := range g.meteors {
		m.Draw(screen)
	}

	for _, b := range g.bullets {
		b.Draw(screen)
	}
}

func (g *Game) AddBullet(b *Bullet) {
	g.bullets = append(g.bullets, b)
}
func (g *Game) Reset() {
	g.player = NewPlayer(g)
	g.meteors = nil
	g.bullets = nil
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return ScreenWidth, ScreenHeight
}

func mustLoadImage(name string) *ebiten.Image {
	f, err := assets.Open(name)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	img, _, err := image.Decode(f)
	if err != nil {
		panic(err)
	}

	return ebiten.NewImageFromImage(img)
}

func mustLoadImages(path string) []*ebiten.Image {
	matches, err := fs.Glob(assets, path)
	if err != nil {
		panic(err)
	}

	images := make([]*ebiten.Image, len(matches))
	for i, match := range matches {
		images[i] = mustLoadImage(match)
	}

	return images
}

func NewGame() *Game {
	g := &Game{
		meteorSpawnTimer: components.NewTimer(time.Second * 1),
	}
	g.player = NewPlayer(g)

	return g
}

func main() {
	g := NewGame()

	err := ebiten.RunGame(g)
	if err != nil {
		panic(err)
	}
}
