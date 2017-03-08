package systems

import (
	"engo.io/ecs"
	"engo.io/engo"
	"engo.io/engo/common"
	"fmt"
)

// Button mappings
var (
	GridToggle  = "gridtoggle"
	HorAxis     = "horAxis"
	VertAxis    = "vertAxis"
	SpaceButton = "SpaceButton"
	ShiftKey    = "shiftkey"
	R_remove    = "R_remove"
)

type ActiveSystemsStruct struct {
	RenderSys *common.RenderSystem
	MouseSys  *common.MouseSystem
	CameraSys *common.CameraSystem
}

// File Names

var (
	TownCenterSprite        = "towncentre1.png"
	ETownCenterSprite       = "towncentre2.png"
	MilitaryBlockSprite     = "militarybuilding1.png"
	EMilitaryBlockSprite    = "militarybuilding2.png"
	ResourceBuildingSprite  = "Resourcebuilding1.png"
	EResourceBuildingSprite = "Resourcebuilding2.png"
	HouseSprite             = "house1.png"
	EHouseSprite            = "house2.png"
	BushSprite              = "bush.png"
	TreeSprite              = "tree.png"
	BuildingSprites         = []string{TownCenterSprite, ETownCenterSprite, MilitaryBlockSprite, EMilitaryBlockSprite, ResourceBuildingSprite, EResourceBuildingSprite, HouseSprite, EHouseSprite, BushSprite}
)

// Other Variables
var (
	ActiveSystems ActiveSystemsStruct
	PlayerFood    int
	PlayerWood    int
	PlayerPop     int
	GridSize      int
	ScaleFactor   float32 // Ratio of Game World size with respect to Window Size
	Chunks        [][]StaticEntity
	ChunkSize     int
	Grid          [][]bool

	GridMaxX int
	GridMaxY int

	ShowDebugPathfinding bool
)

// Message Structs
type BuildingMessage struct {
	Action string
	Name   string
	Index  int
	ID     uint64
}

func (BuildingMessage) Type() string {
	return "BuildingMessage"
}

type CreateBuildingMessage struct {
	Name     string
	Position engo.Point
}

func (CreateBuildingMessage) Type() string {
	return "CreateBuildingMessage"
}

type DestroyBuildingMessage struct {
	obj StaticEntity
}

func (DestroyBuildingMessage) Type() string {
	return "DestroyBuildingMessage"
}

type HealthEnquiryMessage struct {
	ID uint64
}

func (HealthEnquiryMessage) Type() string {
	return "HealthEnquiryMessage"
}

type HealthEnquiryResponseStruct struct {
	HealthResult int
	ResourceName string
	set          bool
}

var HealthEnquiryResponse HealthEnquiryResponseStruct

//Other types

type Fillable interface {
	GetPos() (float32, float32)
	GetSize() (float32, float32)
}

type StaticEntity interface {
	GetPos() (float32, float32)
	GetSize() (float32, float32)
	GetStaticComponent() *StaticComponent
}

// Functions

// Get the mouse position adjusted for zoom
func GetAdjustedMousePos(WRTWindow bool) (float32, float32) {
	CamSys := ActiveSystems.CameraSys
	x := engo.Input.Mouse.X * CamSys.Z() * (engo.GameWidth() / engo.CanvasWidth())
	y := engo.Input.Mouse.Y * CamSys.Z() * (engo.GameHeight() / engo.CanvasHeight())

	if !WRTWindow {
		x += CamSys.X() - (engo.GameWidth()/2)*CamSys.Z()
		y += CamSys.Y() - (engo.GameHeight()/2)*CamSys.Z()
	}

	return x, y

}

// Return chunk that the grid point belongs to
func GetChunkFromPos(x, y float32) (*[]StaticEntity, int) {
	rownum := int(engo.WindowWidth()*ScaleFactor) / (GridSize * ChunkSize)
	X := int(x) / (GridSize * ChunkSize)
	Y := int(y) / (GridSize * ChunkSize)

	//fmt.Println("Chunk of index", X, ",", Y)
	//fmt.Println("Row number is", rownum)

	return &Chunks[Y*rownum+X], (Y*rownum + X)
}

// Store Static objects in respective Chunk(s)
func CacheInChunks(se StaticEntity) {
	x, y := se.GetPos()
	X, Y := se.GetSize()
	X = X + x
	Y = Y + y

	chunk1, _ := GetChunkFromPos(x, y)
	*chunk1 = append(*chunk1, se)

	chunk2, _ := GetChunkFromPos(X, y)
	if chunk1 != chunk2 {
		*chunk2 = append(*chunk2, se)
	}

	chunk3, _ := GetChunkFromPos(x, Y)
	if chunk1 != chunk3 {
		*chunk3 = append(*chunk3, se)
	}

	chunk4, _ := GetChunkFromPos(X, Y)
	if chunk4 != chunk2 && chunk4 != chunk3 {
		*chunk4 = append(*chunk4, se)
	}
}

func UnCacheInChunks(se StaticEntity) {
	x, y := se.GetPos()
	X, Y := se.GetSize()
	X = X + x
	Y = Y + y

	chunk1, _ := GetChunkFromPos(x, y)
	for i, _ := range *chunk1 {
		entity := (*chunk1)[i]
		if entity.GetStaticComponent().ID() == se.GetStaticComponent().ID() {
			(*chunk1)[i] = (*chunk1)[len(*chunk1)-1]
			*chunk1 = (*chunk1)[:len(*chunk1)-1]
		}
	}

	chunk2, _ := GetChunkFromPos(X, y)
	if chunk1 != chunk2 {
		for i, _ := range *chunk2 {
			entity := (*chunk2)[i]
			if entity.GetStaticComponent().ID() == se.GetStaticComponent().ID() {
				(*chunk2)[i] = (*chunk2)[len(*chunk2)-1]
				*chunk2 = (*chunk2)[:len(*chunk2)-1]
			}
		}
	}

	chunk3, _ := GetChunkFromPos(x, Y)
	if chunk1 != chunk3 {
		for i, _ := range *chunk3 {
			entity := (*chunk3)[i]
			if entity.GetStaticComponent().ID() == se.GetStaticComponent().ID() {
				(*chunk3)[i] = (*chunk3)[len(*chunk3)-1]
				*chunk3 = (*chunk3)[:len(*chunk3)-1]
			}
		}
	}

	chunk4, _ := GetChunkFromPos(X, Y)
	if chunk4 != chunk2 && chunk4 != chunk3 {
		for i, _ := range *chunk4 {
			entity := (*chunk4)[i]
			if entity.GetStaticComponent().ID() == se.GetStaticComponent().ID() {
				(*chunk4)[i] = (*chunk4)[len(*chunk4)-1]
				*chunk4 = (*chunk4)[:len(*chunk4)-1]
			}
		}
	}
}

func GetGridAtPos(x, y float32) bool {
	return Grid[int(x)/GridSize][int(y)/GridSize]
}

func WithinGameWindow(x, y float32) bool {
	CamSys := ActiveSystems.CameraSys
	cx, cy := CamSys.X()-engo.WindowWidth()/2, CamSys.Y()-engo.WindowHeight()/2
	ymin := cy + 64
	ymax := cy + engo.WindowHeight() - 160

	return (cx <= x && x <= cx+engo.WindowWidth() && ymin <= y && y <= ymax)
}

// Returns mouse over object?, given button pressed?, given button released?
func StaticMouseCollision(obj StaticEntity, mb engo.MouseButton) (bool, bool, bool) {
	mx, my := GetAdjustedMousePos(false)
	mp := engo.Point{mx, my}
	if WithinGameWindow(mx, my) {
		if obj.GetStaticComponent().Contains(mp) {
			pressed := engo.Input.Mouse.Action == engo.Press
			button_matched := engo.Input.Mouse.Button == mb
			released := engo.Input.Mouse.Action == engo.Release
			return true, (pressed && button_matched), (released && button_matched)
		}
	}
	return false, false, false
}

func GetStaticClicked() StaticEntity {
	mx, my := GetAdjustedMousePos(false)
	mp := engo.Point{mx, my}
	if WithinGameWindow(mx, my) {
		Chunk, _ := GetChunkFromPos(mx, my)
		for i, _ := range *Chunk {
			if (*Chunk)[i].GetStaticComponent().Contains(mp) {
				pressed := engo.Input.Mouse.Action == engo.Press
				button_matched := engo.Input.Mouse.Button == engo.MouseButtonLeft
				if pressed && button_matched {
					return (*Chunk)[i]
				}
			}
		}
	}
	return nil
}

// Mark the solids in the Grid
func FillGrid(f Fillable, val bool) {
	x, y := f.GetPos()
	w, h := f.GetSize()

	for i := int(x) / GridSize; i < int(x+w)/GridSize; i += 1 {
		for j := int(y) / GridSize; j < int(y+h)/GridSize; j += 1 {
			Grid[i][j] = val
		}
	}
}

func RegisterButtons() {
	engo.Input.RegisterButton(GridToggle, engo.Tab)
	engo.Input.RegisterButton(SpaceButton, engo.Space)
	engo.Input.RegisterButton(ShiftKey, engo.LeftShift)
	engo.Input.RegisterAxis(HorAxis, engo.AxisKeyPair{engo.A, engo.D})
	engo.Input.RegisterAxis(VertAxis, engo.AxisKeyPair{engo.W, engo.S})

	fmt.Println("Registered Buttons")
}

func CacheActiveSystems(world *ecs.World) {
	for _, system := range world.Systems() {
		switch sys := system.(type) {
		case *common.RenderSystem:
			ActiveSystems.RenderSys = sys
		case *common.MouseSystem:
			ActiveSystems.MouseSys = sys
		case *common.CameraSystem:
			ActiveSystems.CameraSys = sys
		}
	}

	fmt.Println("Cached Important System References")
}

func InitializeVariables() {
	PlayerFood = 100
	PlayerWood = 50
	PlayerPop = 0

	ShowDebugPathfinding = false

	ScaleFactor = 2

	HealthEnquiryResponse = HealthEnquiryResponseStruct{set: false}

	GridSize = 32

	// Camera bounds is ScaleFactor times window size, also Go defaults to false
	GridMaxX = int(engo.WindowWidth()*ScaleFactor) / GridSize
	GridMaxY = int(engo.WindowHeight()*ScaleFactor) / GridSize
	Grid = make([][]bool, GridMaxX)
	for i, _ := range Grid {
		Grid[i] = make([]bool, GridMaxY)
	}

	// Chunks used to Cache Static Entities
	ChunkSize = 8
	ChunkNum := (int(engo.WindowHeight()*ScaleFactor) / (GridSize * ChunkSize)) * (int(engo.WindowWidth()*ScaleFactor) / (GridSize * ChunkSize))

	Chunks = make([][]StaticEntity, ChunkNum)
	for i, _ := range Chunks {
		Chunks[i] = make([]StaticEntity, 0)
	}
}
